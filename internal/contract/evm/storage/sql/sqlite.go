package sql

import (
	"fmt"
	"os"
	"path/filepath"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql/queries"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStorage struct {
	abiQueries      *queries.ABIQueries
	endpointQueries *queries.EndpointQueries
	contractQueries *queries.ContractQueries
	configQueries   *queries.ConfigQueries
	walletQueries   *queries.WalletQueries
}

// ABI Methods

// CountABIs implements Storage.
func (s *SQLiteStorage) CountABIs() (count int64, err error) {
	count, err = s.abiQueries.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count ABIs: %w", err)
	}
	return count, nil
}

// CreateABI implements Storage.
func (s *SQLiteStorage) CreateABI(abi models.EvmAbi) (id uint, err error) {
	if err := s.abiQueries.Create(&abi); err != nil {
		return 0, fmt.Errorf("failed to create ABI: %w", err)
	}
	return abi.ID, nil
}

// DeleteABI implements Storage.
func (s *SQLiteStorage) DeleteABI(id uint) (err error) {
	if err := s.abiQueries.Delete(id); err != nil {
		return fmt.Errorf("failed to delete ABI: %w", err)
	}
	return nil
}

// GetABIByID implements Storage.
func (s *SQLiteStorage) GetABIByID(id uint) (abi models.EvmAbi, err error) {
	result, err := s.abiQueries.GetByID(id)
	if err != nil {
		return models.EvmAbi{}, fmt.Errorf("failed to get ABI by ID: %w", err)
	}
	return *result, nil
}

// ListABIs implements Storage.
func (s *SQLiteStorage) ListABIs(page int64, pageSize int64) (abis types.Pagination[models.EvmAbi], err error) {
	result, err := s.abiQueries.List(page, pageSize)
	if err != nil {
		return types.Pagination[models.EvmAbi]{}, fmt.Errorf("failed to list ABIs: %w", err)
	}
	return *result, nil
}

// SearchABIs implements Storage.
func (s *SQLiteStorage) SearchABIs(query string) (abis types.Pagination[models.EvmAbi], err error) {
	result, err := s.abiQueries.Search(query)
	if err != nil {
		return types.Pagination[models.EvmAbi]{}, fmt.Errorf("failed to search ABIs: %w", err)
	}
	return *result, nil
}

// UpdateABI implements Storage.
func (s *SQLiteStorage) UpdateABI(id uint, abi models.EvmAbi) (err error) {
	updates := map[string]any{
		"name": abi.Name,
		"abi":  abi.Abi,
	}
	if err := s.abiQueries.Update(id, updates); err != nil {
		return fmt.Errorf("failed to update ABI: %w", err)
	}
	return nil
}

// Endpoint Methods

// CountEndpoints implements Storage.
func (s *SQLiteStorage) CountEndpoints() (count int64, err error) {
	count, err = s.endpointQueries.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count endpoints: %w", err)
	}
	return count, nil
}

// CreateEndpoint implements Storage.
func (s *SQLiteStorage) CreateEndpoint(endpoint models.EVMEndpoint) (id uint, err error) {
	if err := s.endpointQueries.Create(&endpoint); err != nil {
		return 0, fmt.Errorf("failed to create endpoint: %w", err)
	}
	return endpoint.ID, nil
}

// DeleteEndpoint implements Storage.
func (s *SQLiteStorage) DeleteEndpoint(id uint) (err error) {
	if err := s.endpointQueries.Delete(id); err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}
	return nil
}

// GetEndpointByID implements Storage.
func (s *SQLiteStorage) GetEndpointByID(id uint) (endpoint models.EVMEndpoint, err error) {
	result, err := s.endpointQueries.GetByID(id)
	if err != nil {
		return models.EVMEndpoint{}, fmt.Errorf("failed to get endpoint by ID: %w", err)
	}
	return *result, nil
}

// ListEndpoints implements Storage.
func (s *SQLiteStorage) ListEndpoints(page int64, pageSize int64) (endpoints types.Pagination[models.EVMEndpoint], err error) {
	result, err := s.endpointQueries.List(page, pageSize)
	if err != nil {
		return types.Pagination[models.EVMEndpoint]{}, fmt.Errorf("failed to list endpoints: %w", err)
	}
	return *result, nil
}

// SearchEndpoints implements Storage.
func (s *SQLiteStorage) SearchEndpoints(query string) (endpoints types.Pagination[models.EVMEndpoint], err error) {
	result, err := s.endpointQueries.Search(query)
	if err != nil {
		return types.Pagination[models.EVMEndpoint]{}, fmt.Errorf("failed to search endpoints: %w", err)
	}
	return *result, nil
}

// UpdateEndpoint implements Storage.
func (s *SQLiteStorage) UpdateEndpoint(endpointID uint, endpoint models.EVMEndpoint) (err error) {
	updates := map[string]any{
		"name":     endpoint.Name,
		"url":      endpoint.Url,
		"chain_id": endpoint.ChainId,
	}
	if err := s.endpointQueries.Update(endpointID, updates); err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}
	return nil
}

// Contract Methods

// CountContracts implements Storage.
func (s *SQLiteStorage) CountContracts() (count int64, err error) {
	count, err = s.contractQueries.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count contracts: %w", err)
	}
	return count, nil
}

// CreateContract implements Storage.
func (s *SQLiteStorage) CreateContract(contract models.EVMContract) (id uint, err error) {
	if err := s.contractQueries.Create(&contract); err != nil {
		return 0, fmt.Errorf("failed to create contract: %w", err)
	}
	return contract.ID, nil
}

// DeleteContract implements Storage.
func (s *SQLiteStorage) DeleteContract(id uint) (err error) {
	if err := s.contractQueries.Delete(id); err != nil {
		return fmt.Errorf("failed to delete contract: %w", err)
	}
	return nil
}

// GetContractByID implements Storage.
func (s *SQLiteStorage) GetContractByID(id uint) (contract models.EVMContract, err error) {
	result, err := s.contractQueries.GetByID(id)
	if err != nil {
		return models.EVMContract{}, fmt.Errorf("failed to get contract by ID: %w", err)
	}
	return *result, nil
}

// ListContracts implements Storage.
func (s *SQLiteStorage) ListContracts(page int64, pageSize int64) (contracts types.Pagination[models.EVMContract], err error) {
	result, err := s.contractQueries.List(page, pageSize)
	if err != nil {
		return types.Pagination[models.EVMContract]{}, fmt.Errorf("failed to list contracts: %w", err)
	}
	return *result, nil
}

// SearchContracts implements Storage.
func (s *SQLiteStorage) SearchContracts(query string) (contracts types.Pagination[models.EVMContract], err error) {
	result, err := s.contractQueries.Search(query)
	if err != nil {
		return types.Pagination[models.EVMContract]{}, fmt.Errorf("failed to search contracts: %w", err)
	}
	return *result, nil
}

// UpdateContract implements Storage.
func (s *SQLiteStorage) UpdateContract(contractID uint, contract models.EVMContract) (err error) {
	updates := map[string]any{
		"name":          contract.Name,
		"address":       contract.Address,
		"abi_id":        contract.AbiId,
		"status":        contract.Status,
		"contract_code": contract.ContractCode,
		"bytecode":      contract.Bytecode,
		"endpoint_id":   contract.EndpointId,
	}
	if err := s.contractQueries.Update(contractID, updates); err != nil {
		return fmt.Errorf("failed to update contract: %w", err)
	}
	return nil
}

// Config Methods
// CreateConfig implements Storage.
func (s *SQLiteStorage) CreateConfig() (err error) {
	if err := s.configQueries.Create(); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	return nil
}

// DeleteConfig implements Storage.
func (s *SQLiteStorage) DeleteConfig() (err error) {
	if err := s.configQueries.Delete(); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	return nil
}

// GetConfigByID implements Storage.
func (s *SQLiteStorage) GetConfigByID(id uint) (config models.EVMConfig, err error) {
	result, err := s.configQueries.GetByID(id)
	if err != nil {
		return models.EVMConfig{}, fmt.Errorf("failed to get config by ID: %w", err)
	}
	return *result, nil
}

// UpdateConfig implements Storage.
func (s *SQLiteStorage) UpdateConfig(config models.EVMConfig) (err error) {
	if err := s.configQueries.Update(&config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	return nil
}

// Wallet Methods

// CountWallets implements Storage.
func (s *SQLiteStorage) CountWallets() (count int64, err error) {
	count, err = s.walletQueries.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count wallets: %w", err)
	}
	return count, nil
}

// CreateWallet implements Storage.
func (s *SQLiteStorage) CreateWallet(wallet models.EVMWallet) (id uint, err error) {
	if err := s.walletQueries.Create(&wallet); err != nil {
		return 0, fmt.Errorf("failed to create wallet: %w", err)
	}
	return wallet.ID, nil
}

// DeleteWallet implements Storage.
func (s *SQLiteStorage) DeleteWallet(id uint) (err error) {
	if err := s.walletQueries.Delete(id); err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}
	return nil
}

// GetWalletByID implements Storage.
func (s *SQLiteStorage) GetWalletByID(id uint) (wallet models.EVMWallet, err error) {
	result, err := s.walletQueries.GetByID(id)
	if err != nil {
		return models.EVMWallet{}, fmt.Errorf("failed to get wallet by ID: %w", err)
	}
	return *result, nil
}

// GetWalletByAddress implements Storage.
func (s *SQLiteStorage) GetWalletByAddress(address string) (wallet models.EVMWallet, err error) {
	result, err := s.walletQueries.GetByAddress(address)
	if err != nil {
		return models.EVMWallet{}, fmt.Errorf("failed to get wallet by address: %w", err)
	}
	return *result, nil
}

// GetWalletByAlias implements Storage.
func (s *SQLiteStorage) GetWalletByAlias(alias string) (wallet models.EVMWallet, err error) {
	result, err := s.walletQueries.GetByAlias(alias)
	if err != nil {
		return models.EVMWallet{}, fmt.Errorf("failed to get wallet by alias: %w", err)
	}
	return *result, nil
}

// ListWallets implements Storage.
func (s *SQLiteStorage) ListWallets(page int64, pageSize int64) (wallets types.Pagination[models.EVMWallet], err error) {
	result, err := s.walletQueries.List(page, pageSize)
	if err != nil {
		return types.Pagination[models.EVMWallet]{}, fmt.Errorf("failed to list wallets: %w", err)
	}
	return *result, nil
}

// SearchWallets implements Storage.
func (s *SQLiteStorage) SearchWallets(query string) (wallets types.Pagination[models.EVMWallet], err error) {
	result, err := s.walletQueries.Search(query)
	if err != nil {
		return types.Pagination[models.EVMWallet]{}, fmt.Errorf("failed to search wallets: %w", err)
	}
	return *result, nil
}

// UpdateWallet implements Storage.
func (s *SQLiteStorage) UpdateWallet(walletID uint, wallet models.EVMWallet) (err error) {
	updates := map[string]any{
		"alias":            wallet.Alias,
		"address":          wallet.Address,
		"derivation_path":  wallet.DerivationPath,
		"is_from_mnemonic": wallet.IsFromMnemonic,
	}
	if err := s.walletQueries.Update(walletID, updates); err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	return nil
}

// WalletExistsByAddress implements Storage.
func (s *SQLiteStorage) WalletExistsByAddress(address string) (exists bool, err error) {
	exists, err = s.walletQueries.ExistsByAddress(address)
	if err != nil {
		return false, fmt.Errorf("failed to check wallet existence by address: %w", err)
	}
	return exists, nil
}

// WalletExistsByAlias implements Storage.
func (s *SQLiteStorage) WalletExistsByAlias(alias string) (exists bool, err error) {
	exists, err = s.walletQueries.ExistsByAlias(alias)
	if err != nil {
		return false, fmt.Errorf("failed to check wallet existence by alias: %w", err)
	}
	return exists, nil
}

// GetCurrentConfig implements Storage.
func (s *SQLiteStorage) GetCurrentConfig() (config models.EVMConfig, err error) {
	result, err := s.configQueries.GetCurrent()
	if err != nil {
		return models.EVMConfig{}, fmt.Errorf("failed to get current config: %w", err)
	}
	return *result, nil
}

// NewSQLiteDB creates a new SQLite database connection.
// If dbPath is empty, it defaults to $HOME/smart-contract-cli.db.
func NewSQLiteDB(dbPath string) (Storage, error) {
	// Use default path if none provided
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		dbPath = filepath.Join(homeDir, "smart-contract-cli.db")
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with GORM
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate the schema
	if err := database.AutoMigrate(
		&models.EvmAbi{},
		&models.EVMEndpoint{},
		&models.EVMContract{},
		&models.EVMConfig{},
		&models.EVMWallet{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	// Initialize query helpers
	return &SQLiteStorage{
		abiQueries:      queries.NewABIQueries(database),
		endpointQueries: queries.NewEndpointQueries(database),
		contractQueries: queries.NewContractQueries(database),
		configQueries:   queries.NewConfigQueries(database),
		walletQueries:   queries.NewWalletQueries(database),
	}, nil
}
