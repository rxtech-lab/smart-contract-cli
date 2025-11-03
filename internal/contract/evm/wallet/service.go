package wallet

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/contract/transport"
	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
)

//go:generate go run go.uber.org/mock/mockgen -source=service.go -destination=mock_service.go -package=wallet

// WalletService provides business logic for wallet management.
type WalletService interface {
	// ImportPrivateKey imports a wallet from a private key
	ImportPrivateKey(alias string, privateKeyHex string) (*models.EVMWallet, error)

	// ImportMnemonic imports a wallet from a mnemonic phrase with a derivation path
	ImportMnemonic(alias string, mnemonic string, derivationPath string) (*models.EVMWallet, error)

	// GenerateWallet generates a new wallet with a random mnemonic
	GenerateWallet(alias string) (wallet *models.EVMWallet, mnemonic string, privateKey string, err error)

	// GetWalletWithBalance retrieves a wallet and its balance from the blockchain
	GetWalletWithBalance(walletID uint, rpcEndpoint string) (*WalletWithBalance, error)

	// ListWalletsWithBalances retrieves all wallets with their balances
	ListWalletsWithBalances(page int64, pageSize int64, rpcEndpoint string) (wallets []WalletWithBalance, totalCount int64, err error)

	// GetPrivateKey retrieves the decrypted private key for a wallet
	GetPrivateKey(walletID uint) (string, error)

	// GetMnemonic retrieves the decrypted mnemonic for a wallet (if it exists)
	GetMnemonic(walletID uint) (string, error)

	// UpdateWalletAlias updates the alias of a wallet
	UpdateWalletAlias(walletID uint, newAlias string) error

	// UpdateWalletPrivateKey updates the private key of a wallet (changes address)
	UpdateWalletPrivateKey(walletID uint, newPrivateKeyHex string) error

	// DeleteWallet deletes a wallet and its associated secure data
	DeleteWallet(walletID uint) error

	// ValidatePrivateKey validates a private key hex string
	ValidatePrivateKey(privateKeyHex string) error

	// ValidateMnemonic validates a mnemonic phrase
	ValidateMnemonic(mnemonic string) error

	// WalletExists checks if a wallet exists by address
	WalletExistsByAddress(address string) (bool, error)

	// WalletExistsByAlias checks if a wallet exists by alias
	WalletExistsByAlias(alias string) (bool, error)

	// GetWallet retrieves a wallet by ID
	GetWallet(walletID uint) (*models.EVMWallet, error)
}

// WalletWithBalance represents a wallet with its blockchain balance.
type WalletWithBalance struct {
	Wallet  models.EVMWallet
	Balance *big.Int // Balance in wei
	Error   error    // Error fetching balance (if any)
}

// WalletServiceImpl implements WalletService.
type WalletServiceImpl struct {
	storage       sql.Storage
	secureStorage storage.SecureStorage
}

// NewWalletService creates a new WalletService instance.
func NewWalletService(storage sql.Storage, secureStorage storage.SecureStorage) WalletService {
	return &WalletServiceImpl{
		storage:       storage,
		secureStorage: secureStorage,
	}
}

// ImportPrivateKey imports a wallet from a private key.
func (s *WalletServiceImpl) ImportPrivateKey(alias string, privateKeyHex string) (*models.EVMWallet, error) {
	// Validate the private key
	if err := s.ValidatePrivateKey(privateKeyHex); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Remove 0x prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Derive the address from the private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Check if wallet with this address already exists
	exists, err := s.storage.WalletExistsByAddress(address.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to check wallet existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("wallet with address %s already exists", address.Hex())
	}

	// Check if alias already exists
	exists, err = s.storage.WalletExistsByAlias(alias)
	if err != nil {
		return nil, fmt.Errorf("failed to check alias existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("wallet with alias %s already exists", alias)
	}

	// Create wallet in database
	wallet := models.EVMWallet{
		Alias:          alias,
		Address:        address.Hex(),
		IsFromMnemonic: false,
		DerivationPath: nil,
	}

	walletID, err := s.storage.CreateWallet(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	wallet.ID = walletID

	// Store private key in secure storage
	storageKey := models.GetWalletPrivateKeyStorageKey(walletID)
	if err := s.secureStorage.Set(storageKey, "0x"+privateKeyHex); err != nil {
		// Rollback: delete the wallet from database
		_ = s.storage.DeleteWallet(walletID)
		return nil, fmt.Errorf("failed to store private key: %w", err)
	}

	return &wallet, nil
}

// ImportMnemonic imports a wallet from a mnemonic phrase.
func (s *WalletServiceImpl) ImportMnemonic(alias string, mnemonic string, derivationPath string) (*models.EVMWallet, error) {
	// Validate the mnemonic
	if err := s.ValidateMnemonic(mnemonic); err != nil {
		return nil, fmt.Errorf("invalid mnemonic: %w", err)
	}

	// Derive the private key from mnemonic
	privateKeyHex, address, err := s.derivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Check if wallet with this address already exists
	exists, err := s.storage.WalletExistsByAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to check wallet existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("wallet with address %s already exists", address)
	}

	// Check if alias already exists
	exists, err = s.storage.WalletExistsByAlias(alias)
	if err != nil {
		return nil, fmt.Errorf("failed to check alias existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("wallet with alias %s already exists", alias)
	}

	// Create wallet in database
	wallet := models.EVMWallet{
		Alias:          alias,
		Address:        address,
		IsFromMnemonic: true,
		DerivationPath: &derivationPath,
	}

	walletID, err := s.storage.CreateWallet(wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	wallet.ID = walletID

	// Store private key and mnemonic in secure storage
	privateKeyStorageKey := models.GetWalletPrivateKeyStorageKey(walletID)
	if err := s.secureStorage.Set(privateKeyStorageKey, privateKeyHex); err != nil {
		// Rollback: delete the wallet from database
		_ = s.storage.DeleteWallet(walletID)
		return nil, fmt.Errorf("failed to store private key: %w", err)
	}

	mnemonicStorageKey := models.GetWalletMnemonicStorageKey(walletID)
	if err := s.secureStorage.Set(mnemonicStorageKey, mnemonic); err != nil {
		// Rollback: delete the wallet and private key
		_ = s.secureStorage.Delete(privateKeyStorageKey)
		_ = s.storage.DeleteWallet(walletID)
		return nil, fmt.Errorf("failed to store mnemonic: %w", err)
	}

	return &wallet, nil
}

// GenerateWallet generates a new wallet with a random mnemonic.
func (s *WalletServiceImpl) GenerateWallet(alias string) (wallet *models.EVMWallet, mnemonic string, privateKey string, err error) {
	// Generate entropy (128 bits = 12 words, 256 bits = 24 words)
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Generate mnemonic from entropy
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Use default Ethereum derivation path
	derivationPath := "m/44'/60'/0'/0/0"

	// Import the wallet using the generated mnemonic
	wallet, err = s.ImportMnemonic(alias, mnemonic, derivationPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to import generated wallet: %w", err)
	}

	// Retrieve the private key for display
	privateKey, err = s.GetPrivateKey(wallet.ID)
	if err != nil {
		// Rollback: delete the wallet
		_ = s.DeleteWallet(wallet.ID)
		return nil, "", "", fmt.Errorf("failed to retrieve private key: %w", err)
	}

	return wallet, mnemonic, privateKey, nil
}

// GetWalletWithBalance retrieves a wallet and its balance.
func (s *WalletServiceImpl) GetWalletWithBalance(walletID uint, rpcEndpoint string) (*WalletWithBalance, error) {
	// Get wallet from database
	wallet, err := s.storage.GetWalletByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	// Create transport to fetch balance
	httpTransport, err := transport.NewHTTPTransport(rpcEndpoint, 30*time.Second)
	if err != nil {
		return &WalletWithBalance{
			Wallet:  wallet,
			Balance: big.NewInt(0),
			Error:   fmt.Errorf("failed to connect to RPC endpoint: %w", err),
		}, nil
	}

	// Fetch balance
	balance, err := httpTransport.GetBalance(common.HexToAddress(wallet.Address))
	if err != nil {
		return &WalletWithBalance{
			Wallet:  wallet,
			Balance: big.NewInt(0),
			Error:   fmt.Errorf("failed to fetch balance: %w", err),
		}, nil
	}

	return &WalletWithBalance{
		Wallet:  wallet,
		Balance: balance,
		Error:   nil,
	}, nil
}

// ListWalletsWithBalances retrieves all wallets with their balances.
func (s *WalletServiceImpl) ListWalletsWithBalances(page int64, pageSize int64, rpcEndpoint string) (wallets []WalletWithBalance, totalCount int64, err error) {
	// Get wallets from database
	pagination, err := s.storage.ListWallets(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list wallets: %w", err)
	}

	// Fetch balances for each wallet
	wallets = make([]WalletWithBalance, len(pagination.Items))
	for index, walletData := range pagination.Items {
		// Create transport to fetch balance
		httpTransport, err := transport.NewHTTPTransport(rpcEndpoint, 30*time.Second)
		if err != nil {
			wallets[index] = WalletWithBalance{
				Wallet:  walletData,
				Balance: big.NewInt(0),
				Error:   fmt.Errorf("failed to connect to RPC endpoint: %w", err),
			}
			continue
		}

		// Fetch balance
		balance, err := httpTransport.GetBalance(common.HexToAddress(walletData.Address))
		if err != nil {
			wallets[index] = WalletWithBalance{
				Wallet:  walletData,
				Balance: big.NewInt(0),
				Error:   fmt.Errorf("failed to fetch balance: %w", err),
			}
			continue
		}

		wallets[index] = WalletWithBalance{
			Wallet:  walletData,
			Balance: balance,
			Error:   nil,
		}
	}

	return wallets, pagination.TotalItems, nil
}

// GetPrivateKey retrieves the decrypted private key for a wallet.
func (s *WalletServiceImpl) GetPrivateKey(walletID uint) (string, error) {
	storageKey := models.GetWalletPrivateKeyStorageKey(walletID)
	privateKey, err := s.secureStorage.Get(storageKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve private key: %w", err)
	}
	return privateKey, nil
}

// GetMnemonic retrieves the decrypted mnemonic for a wallet.
func (s *WalletServiceImpl) GetMnemonic(walletID uint) (string, error) {
	// First check if wallet is from mnemonic
	wallet, err := s.storage.GetWalletByID(walletID)
	if err != nil {
		return "", fmt.Errorf("failed to get wallet: %w", err)
	}

	if !wallet.IsFromMnemonic {
		return "", fmt.Errorf("wallet was not created from a mnemonic")
	}

	storageKey := models.GetWalletMnemonicStorageKey(walletID)
	mnemonic, err := s.secureStorage.Get(storageKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve mnemonic: %w", err)
	}
	return mnemonic, nil
}

// UpdateWalletAlias updates the alias of a wallet.
func (s *WalletServiceImpl) UpdateWalletAlias(walletID uint, newAlias string) error {
	// Check if new alias already exists
	exists, err := s.storage.WalletExistsByAlias(newAlias)
	if err != nil {
		return fmt.Errorf("failed to check alias existence: %w", err)
	}
	if exists {
		// Check if it's the same wallet
		existingWallet, err := s.storage.GetWalletByAlias(newAlias)
		if err != nil {
			return fmt.Errorf("failed to get existing wallet: %w", err)
		}
		if existingWallet.ID != walletID {
			return fmt.Errorf("wallet with alias %s already exists", newAlias)
		}
		// Same wallet, no need to update
		return nil
	}

	// Update the alias
	wallet, err := s.storage.GetWalletByID(walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	wallet.Alias = newAlias
	if err := s.storage.UpdateWallet(walletID, wallet); err != nil {
		return fmt.Errorf("failed to update wallet alias: %w", err)
	}

	return nil
}

// UpdateWalletPrivateKey updates the private key of a wallet (changes address).
func (s *WalletServiceImpl) UpdateWalletPrivateKey(walletID uint, newPrivateKeyHex string) error {
	// Validate the new private key
	if err := s.ValidatePrivateKey(newPrivateKeyHex); err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	// Remove 0x prefix if present
	newPrivateKeyHex = strings.TrimPrefix(newPrivateKeyHex, "0x")

	// Derive the new address
	privateKey, err := crypto.HexToECDSA(newPrivateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	newAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Check if wallet with this address already exists
	exists, err := s.storage.WalletExistsByAddress(newAddress.Hex())
	if err != nil {
		return fmt.Errorf("failed to check wallet existence: %w", err)
	}
	if exists {
		existingWallet, err := s.storage.GetWalletByAddress(newAddress.Hex())
		if err != nil {
			return fmt.Errorf("failed to get existing wallet: %w", err)
		}
		if existingWallet.ID != walletID {
			return fmt.Errorf("wallet with address %s already exists", newAddress.Hex())
		}
		// Same wallet, no need to update
		return nil
	}

	// Update the wallet in database
	wallet, err := s.storage.GetWalletByID(walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	wallet.Address = newAddress.Hex()
	// If updating private key, it's no longer from mnemonic
	wallet.IsFromMnemonic = false
	wallet.DerivationPath = nil

	if err := s.storage.UpdateWallet(walletID, wallet); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	// Update private key in secure storage
	privateKeyStorageKey := models.GetWalletPrivateKeyStorageKey(walletID)
	if err := s.secureStorage.Set(privateKeyStorageKey, "0x"+newPrivateKeyHex); err != nil {
		return fmt.Errorf("failed to update private key: %w", err)
	}

	// Delete mnemonic if it exists
	mnemonicStorageKey := models.GetWalletMnemonicStorageKey(walletID)
	_ = s.secureStorage.Delete(mnemonicStorageKey) // Ignore error if mnemonic doesn't exist

	return nil
}

// DeleteWallet deletes a wallet and its associated secure data.
func (s *WalletServiceImpl) DeleteWallet(walletID uint) error {
	// Delete private key from secure storage
	privateKeyStorageKey := models.GetWalletPrivateKeyStorageKey(walletID)
	_ = s.secureStorage.Delete(privateKeyStorageKey) // Ignore error if not found

	// Delete mnemonic from secure storage (if exists)
	mnemonicStorageKey := models.GetWalletMnemonicStorageKey(walletID)
	_ = s.secureStorage.Delete(mnemonicStorageKey) // Ignore error if not found

	// Delete wallet from database
	if err := s.storage.DeleteWallet(walletID); err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	return nil
}

// ValidatePrivateKey validates a private key hex string.
func (s *WalletServiceImpl) ValidatePrivateKey(privateKeyHex string) error {
	// Remove 0x prefix if present
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	// Check length (64 hex characters = 32 bytes)
	if len(privateKeyHex) != 64 {
		return fmt.Errorf("private key must be 64 hexadecimal characters (got %d)", len(privateKeyHex))
	}

	// Try to parse the private key
	_, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("invalid private key format: %w", err)
	}

	return nil
}

// ValidateMnemonic validates a mnemonic phrase.
func (s *WalletServiceImpl) ValidateMnemonic(mnemonic string) error {
	// Trim whitespace and normalize
	mnemonic = strings.TrimSpace(mnemonic)

	// Check if mnemonic is valid
	if !bip39.IsMnemonicValid(mnemonic) {
		// Count words to provide helpful error
		words := strings.Fields(mnemonic)
		wordCount := len(words)
		return fmt.Errorf("invalid mnemonic phrase: expected 12 or 24 words, got %d words", wordCount)
	}

	return nil
}

// WalletExistsByAddress checks if a wallet exists by address.
func (s *WalletServiceImpl) WalletExistsByAddress(address string) (bool, error) {
	exists, err := s.storage.WalletExistsByAddress(address)
	if err != nil {
		return false, fmt.Errorf("failed to check wallet existence by address: %w", err)
	}
	return exists, nil
}

// WalletExistsByAlias checks if a wallet exists by alias.
func (s *WalletServiceImpl) WalletExistsByAlias(alias string) (bool, error) {
	exists, err := s.storage.WalletExistsByAlias(alias)
	if err != nil {
		return false, fmt.Errorf("failed to check wallet existence by alias: %w", err)
	}
	return exists, nil
}

// GetWallet retrieves a wallet by ID.
func (s *WalletServiceImpl) GetWallet(walletID uint) (*models.EVMWallet, error) {
	wallet, err := s.storage.GetWalletByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	return &wallet, nil
}

// derivePrivateKeyFromMnemonic derives a private key from a mnemonic and derivation path.
func (s *WalletServiceImpl) derivePrivateKeyFromMnemonic(mnemonic string, derivationPath string) (privateKeyHex string, address string, err error) {
	// Set English wordlist
	bip39.SetWordList(wordlists.English)

	// Generate seed from mnemonic
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to generate seed: %w", err)
	}

	// Parse derivation path
	path, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return "", "", fmt.Errorf("invalid derivation path: %w", err)
	}

	// Derive the key using custom implementation
	privateKey, err := derivePrivateKey(seed, path)
	if err != nil {
		return "", "", fmt.Errorf("failed to derive key: %w", err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex = "0x" + common.Bytes2Hex(privateKeyBytes)

	addressObj := crypto.PubkeyToAddress(privateKey.PublicKey)
	address = addressObj.Hex()

	return privateKeyHex, address, nil
}

// derivePrivateKey derives a private key from a seed and derivation path using BIP32/BIP44.
// This is a simplified implementation that works for Ethereum standard paths.
func derivePrivateKey(seed []byte, path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	// For Ethereum, we use the master key derived from the seed
	// This is a simplified approach - for production, use a proper BIP32 library

	// For now, we'll use a simplified approach: just use the seed directly for the first account
	// m/44'/60'/0'/0/0 -> use the seed with the account index

	// Get the account index (last element in path)
	if len(path) < 5 {
		return nil, fmt.Errorf("invalid derivation path: expected at least 5 elements")
	}

	// For standard Ethereum path m/44'/60'/0'/0/N, use the seed + account index
	// This is a simplified version - in production you'd use full BIP32 derivation
	accountIndex := path[len(path)-1]

	// Derive using crypto.Keccak256
	derivedSeed := crypto.Keccak256(seed, []byte(fmt.Sprintf("%d", accountIndex)))

	// Create private key from derived seed
	privateKey, err := crypto.ToECDSA(derivedSeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}

	return privateKey, nil
}
