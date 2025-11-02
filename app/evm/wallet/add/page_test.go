package add

import (
	"fmt"
	"math/big"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/wallet"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// WalletAddPageTestSuite is the test suite for the wallet add page.
type WalletAddPageTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	router        *view.MockRouter
	sharedMemory  storage.SharedMemory
	walletService *wallet.MockWalletService
	mockStorage   *sql.MockStorage
	model         Model
}

func TestWalletAddPageTestSuite(t *testing.T) {
	suite.Run(t, new(WalletAddPageTestSuite))
}

func (suite *WalletAddPageTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.router = view.NewMockRouter(suite.ctrl)
	suite.sharedMemory = storage.NewSharedMemory()
	suite.walletService = wallet.NewMockWalletService(suite.ctrl)
	suite.mockStorage = sql.NewMockStorage(suite.ctrl)

	// Set up mock storage in shared memory with endpoint configured
	testEndpoint := &models.EVMEndpoint{
		ID:  1,
		Url: "http://localhost:8545",
	}
	testConfig := models.EVMConfig{
		ID:         1,
		EndpointId: &testEndpoint.ID,
		Endpoint:   testEndpoint,
	}
	// Allow any number of calls to GetCurrentConfig - use a large number
	suite.mockStorage.EXPECT().GetCurrentConfig().Return(testConfig, nil).AnyTimes()

	// Add mock storage to shared memory
	err := suite.sharedMemory.Set(config.StorageClientKey, suite.mockStorage)
	suite.NoError(err, "Should set storage client in shared memory")

	// Create model with mocked wallet service
	page := NewPageWithService(suite.router, suite.sharedMemory, suite.walletService)
	suite.model = page.(Model)
}

func (suite *WalletAddPageTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateWalletWithPrivateKey tests the full flow of creating a wallet with a private key.
func (suite *WalletAddPageTestSuite) TestCreateWalletWithPrivateKey() {
	// Step 1: Start at method selection
	suite.Equal(stepSelectMethod, suite.model.currentStep)
	suite.Equal(0, suite.model.selectedIndex)

	// Step 2: Select "Import from private key" (first option, index 0)
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepEnterAlias, suite.model.currentStep)
	suite.Equal(methodPrivateKey, suite.model.method)

	// Step 3: Enter alias
	suite.model.aliasInput.SetValue("test-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepEnterPrivateKey, suite.model.currentStep)

	// Step 4: Enter private key and mock the service
	testPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.pkeyInput.SetValue(testPrivateKey)

	// Mock wallet service expectations
	suite.walletService.EXPECT().ValidatePrivateKey(testPrivateKey).Return(nil)
	suite.walletService.EXPECT().WalletExistsByAlias("test-wallet").Return(false, nil)

	expectedWallet := &models.EVMWallet{
		ID:      1,
		Alias:   "test-wallet",
		Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
	}
	suite.walletService.EXPECT().ImportPrivateKey("test-wallet", testPrivateKey).Return(expectedWallet, nil)
	suite.walletService.EXPECT().WalletExistsByAddress(expectedWallet.Address).Return(false, nil)

	balance, _ := new(big.Int).SetString("10000000000000000000", 10) // 10 ETH
	walletWithBalance := &wallet.WalletWithBalance{
		Wallet:  *expectedWallet,
		Balance: balance,
	}
	suite.walletService.EXPECT().GetWalletWithBalance(uint(1), "http://localhost:8545").Return(walletWithBalance, nil)

	// Trigger import
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Wait for import to complete (simulate async message)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	// Should be at confirmation step
	suite.Equal(stepConfirm, suite.model.currentStep)
	suite.NotNil(suite.model.confirmedWallet)
	suite.Equal("test-wallet", suite.model.confirmedWallet.Wallet.Alias)
	suite.Equal("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", suite.model.confirmedWallet.Wallet.Address)

	// Step 5: Confirm save (select first option "Save wallet")
	suite.Equal(0, suite.model.selectedIndex)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepSuccess, suite.model.currentStep)
}

// TestCreateWalletWithMnemonic tests the full flow of creating a wallet with a mnemonic phrase.
func (suite *WalletAddPageTestSuite) TestCreateWalletWithMnemonic() {
	// Step 1: Select "Import from mnemonic phrase" (second option, index 1)
	suite.model.selectedIndex = 1
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepEnterAlias, suite.model.currentStep)
	suite.Equal(methodMnemonic, suite.model.method)

	// Step 2: Enter alias
	suite.model.aliasInput.SetValue("mnemonic-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepEnterMnemonic, suite.model.currentStep)

	// Step 3: Enter mnemonic
	testMnemonic := "test test test test test test test test test test test junk"
	suite.model.mnemonicInput.SetValue(testMnemonic)

	// Proceed to derivation path selection (Ctrl+S)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	suite.model = updatedModel.(Model)
	suite.Equal(stepSelectDerivationPath, suite.model.currentStep)

	// Step 4: Select default derivation path (first option)
	suite.Equal(0, suite.model.selectedIndex)

	// Mock wallet service expectations
	suite.walletService.EXPECT().ValidateMnemonic(testMnemonic).Return(nil)
	suite.walletService.EXPECT().WalletExistsByAlias("mnemonic-wallet").Return(false, nil)

	derivationPath := "m/44'/60'/0'/0/0"
	expectedWallet := &models.EVMWallet{
		ID:             2,
		Alias:          "mnemonic-wallet",
		Address:        "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
		DerivationPath: &derivationPath,
	}
	suite.walletService.EXPECT().ImportMnemonic("mnemonic-wallet", testMnemonic, derivationPath).Return(expectedWallet, nil)
	suite.walletService.EXPECT().WalletExistsByAddress(expectedWallet.Address).Return(false, nil)

	balance, _ := new(big.Int).SetString("5000000000000000000", 10) // 5 ETH
	walletWithBalance := &wallet.WalletWithBalance{
		Wallet:  *expectedWallet,
		Balance: balance,
	}
	suite.walletService.EXPECT().GetWalletWithBalance(uint(2), "http://localhost:8545").Return(walletWithBalance, nil)

	// Trigger import
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Wait for import to complete
	importMsg := suite.model.importMnemonic()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	// Should be at confirmation step
	suite.Equal(stepConfirm, suite.model.currentStep)
	suite.NotNil(suite.model.confirmedWallet)
	suite.Equal("mnemonic-wallet", suite.model.confirmedWallet.Wallet.Alias)
	suite.Equal("0x70997970C51812dc3A010C7d01b50e0d17dc79C8", suite.model.confirmedWallet.Wallet.Address)
	suite.NotNil(suite.model.confirmedWallet.Wallet.DerivationPath)
	suite.Equal(derivationPath, *suite.model.confirmedWallet.Wallet.DerivationPath)

	// Step 5: Confirm save
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepSuccess, suite.model.currentStep)
}

// TestGenerateNewWallet tests the full flow of generating a new random wallet.
func (suite *WalletAddPageTestSuite) TestGenerateNewWallet() {
	// Step 1: Select "Generate new wallet" (third option, index 2)
	suite.model.selectedIndex = 2
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepEnterAlias, suite.model.currentStep)
	suite.Equal(methodGenerate, suite.model.method)

	// Step 2: Enter alias
	suite.model.aliasInput.SetValue("generated-wallet")

	// Mock wallet service expectations
	suite.walletService.EXPECT().WalletExistsByAlias("generated-wallet").Return(false, nil)

	generatedMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	generatedPrivateKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	expectedWallet := &models.EVMWallet{
		ID:      3,
		Alias:   "generated-wallet",
		Address: "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
	}
	suite.walletService.EXPECT().GenerateWallet("generated-wallet").Return(
		expectedWallet,
		generatedMnemonic,
		generatedPrivateKey,
		nil,
	)

	balance := big.NewInt(0) // New wallet, no balance
	walletWithBalance := &wallet.WalletWithBalance{
		Wallet:  *expectedWallet,
		Balance: balance,
	}
	suite.walletService.EXPECT().GetWalletWithBalance(uint(3), "http://localhost:8545").Return(walletWithBalance, nil)

	// Trigger generation (pressing Enter will call generateWallet)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepGenerating, suite.model.currentStep)

	// Wait for generation to complete
	genMsg := suite.model.generateWallet()
	updatedModel, _ = suite.model.Update(genMsg)
	suite.model = updatedModel.(Model)

	// Should be at backup screen
	suite.Equal(stepShowBackup, suite.model.currentStep)
	suite.NotNil(suite.model.confirmedWallet)
	suite.Equal(generatedMnemonic, suite.model.generatedMnemonic)
	suite.Equal(generatedPrivateKey, suite.model.generatedPKey)
	suite.Equal("0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC", suite.model.generatedAddress)

	// Step 3: Confirm backup saved (select first option)
	suite.Equal(0, suite.model.selectedIndex)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepSuccess, suite.model.currentStep)
}

// TestPrivateKeyValidationError tests error handling when private key is invalid.
func (suite *WalletAddPageTestSuite) TestPrivateKeyValidationError() {
	// Navigate to private key entry
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("test-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Enter invalid private key
	invalidKey := "invalid-key"
	suite.model.pkeyInput.SetValue(invalidKey)

	suite.walletService.EXPECT().ValidatePrivateKey(invalidKey).Return(fmt.Errorf("invalid private key"))

	// Trigger import
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.Equal(stepError, suite.model.currentStep)
	suite.Contains(suite.model.errorMsg, "invalid private key format")
}

// TestMnemonicValidationError tests error handling when mnemonic is invalid.
func (suite *WalletAddPageTestSuite) TestMnemonicValidationError() {
	// Navigate to mnemonic entry
	suite.model.selectedIndex = 1
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("test-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Enter invalid mnemonic (only 5 words)
	invalidMnemonic := "word1 word2 word3 word4 word5"
	suite.model.mnemonicInput.SetValue(invalidMnemonic)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	suite.model = updatedModel.(Model)

	suite.walletService.EXPECT().ValidateMnemonic(invalidMnemonic).Return(fmt.Errorf("invalid mnemonic"))

	// Trigger import
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importMnemonic()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.Equal(stepError, suite.model.currentStep)
	suite.Contains(suite.model.errorMsg, "invalid mnemonic phrase")
	suite.Contains(suite.model.errorMsg, "got 5 words")
}

// TestDuplicateAliasError tests error handling when alias already exists.
func (suite *WalletAddPageTestSuite) TestDuplicateAliasError() {
	// Navigate to private key entry
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("existing-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	testPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.pkeyInput.SetValue(testPrivateKey)

	suite.walletService.EXPECT().ValidatePrivateKey(testPrivateKey).Return(nil)
	suite.walletService.EXPECT().WalletExistsByAlias("existing-wallet").Return(true, nil)

	// Trigger import
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.Equal(stepError, suite.model.currentStep)
	suite.Contains(suite.model.errorMsg, "wallet with alias 'existing-wallet' already exists")
}

// TestDuplicateAddressError tests error handling when wallet address already exists.
func (suite *WalletAddPageTestSuite) TestDuplicateAddressError() {
	// Navigate to private key entry
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("test-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	testPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.pkeyInput.SetValue(testPrivateKey)

	suite.walletService.EXPECT().ValidatePrivateKey(testPrivateKey).Return(nil)
	suite.walletService.EXPECT().WalletExistsByAlias("test-wallet").Return(false, nil)

	expectedWallet := &models.EVMWallet{
		ID:      1,
		Alias:   "test-wallet",
		Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
	}
	suite.walletService.EXPECT().ImportPrivateKey("test-wallet", testPrivateKey).Return(expectedWallet, nil)
	suite.walletService.EXPECT().WalletExistsByAddress(expectedWallet.Address).Return(true, nil)
	suite.walletService.EXPECT().DeleteWallet(uint(1)).Return(nil)

	// Trigger import
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.Equal(stepError, suite.model.currentStep)
	suite.Contains(suite.model.errorMsg, "wallet already exists with address")
}

// TestCancelAtConfirmation tests canceling wallet creation at confirmation step.
func (suite *WalletAddPageTestSuite) TestCancelAtConfirmation() {
	// Setup: create a wallet and get to confirmation step
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("test-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	testPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.pkeyInput.SetValue(testPrivateKey)

	suite.walletService.EXPECT().ValidatePrivateKey(testPrivateKey).Return(nil)
	suite.walletService.EXPECT().WalletExistsByAlias("test-wallet").Return(false, nil)

	expectedWallet := &models.EVMWallet{
		ID:      1,
		Alias:   "test-wallet",
		Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
	}
	suite.walletService.EXPECT().ImportPrivateKey("test-wallet", testPrivateKey).Return(expectedWallet, nil)
	suite.walletService.EXPECT().WalletExistsByAddress(expectedWallet.Address).Return(false, nil)

	balance, _ := new(big.Int).SetString("10000000000000000000", 10)
	walletWithBalance := &wallet.WalletWithBalance{
		Wallet:  *expectedWallet,
		Balance: balance,
	}
	suite.walletService.EXPECT().GetWalletWithBalance(uint(1), "http://localhost:8545").Return(walletWithBalance, nil)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	suite.Equal(stepConfirm, suite.model.currentStep)

	// Select "Cancel" option (index 1)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	// Mock deletion
	suite.walletService.EXPECT().DeleteWallet(uint(1)).Return(nil)

	// Confirm cancellation - this will trigger navigation via command
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Note: Navigation happens via tea.Cmd, not direct call
}

// TestCancelGeneratedWalletAtBackup tests canceling a generated wallet at backup step.
func (suite *WalletAddPageTestSuite) TestCancelGeneratedWalletAtBackup() {
	// Setup: generate a wallet and get to backup step
	suite.model.selectedIndex = 2
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("generated-wallet")

	suite.walletService.EXPECT().WalletExistsByAlias("generated-wallet").Return(false, nil)

	expectedWallet := &models.EVMWallet{
		ID:      3,
		Alias:   "generated-wallet",
		Address: "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
	}
	suite.walletService.EXPECT().GenerateWallet("generated-wallet").Return(
		expectedWallet,
		"test mnemonic phrase",
		"0xprivatekey",
		nil,
	)

	balance := big.NewInt(0)
	walletWithBalance := &wallet.WalletWithBalance{
		Wallet:  *expectedWallet,
		Balance: balance,
	}
	suite.walletService.EXPECT().GetWalletWithBalance(uint(3), "http://localhost:8545").Return(walletWithBalance, nil)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	genMsg := suite.model.generateWallet()
	updatedModel, _ = suite.model.Update(genMsg)
	suite.model = updatedModel.(Model)

	suite.Equal(stepShowBackup, suite.model.currentStep)

	// Select "Cancel" option (index 1)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	// Mock deletion
	suite.walletService.EXPECT().DeleteWallet(uint(3)).Return(nil)

	// Confirm cancellation - this will trigger navigation via command
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Note: Navigation happens via tea.Cmd, not direct call
}

// TestCustomDerivationPath tests using a custom derivation path with mnemonic.
func (suite *WalletAddPageTestSuite) TestCustomDerivationPath() {
	// Navigate to mnemonic import
	suite.model.selectedIndex = 1
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("custom-path-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	testMnemonic := "test test test test test test test test test test test junk"
	suite.model.mnemonicInput.SetValue(testMnemonic)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	suite.model = updatedModel.(Model)

	// Navigate to custom path option (last option)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(3, suite.model.selectedIndex) // Custom path is at index 3

	// Enter custom path
	customPath := "m/44'/60'/1'/0/0"
	suite.model.customPathInput.SetValue(customPath)

	suite.walletService.EXPECT().ValidateMnemonic(testMnemonic).Return(nil)
	suite.walletService.EXPECT().WalletExistsByAlias("custom-path-wallet").Return(false, nil)

	expectedWallet := &models.EVMWallet{
		ID:             4,
		Alias:          "custom-path-wallet",
		Address:        "0x90F79bf6EB2c4f870365E785982E1f101E93b906",
		DerivationPath: &customPath,
	}
	suite.walletService.EXPECT().ImportMnemonic("custom-path-wallet", testMnemonic, customPath).Return(expectedWallet, nil)
	suite.walletService.EXPECT().WalletExistsByAddress(expectedWallet.Address).Return(false, nil)

	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	walletWithBalance := &wallet.WalletWithBalance{
		Wallet:  *expectedWallet,
		Balance: balance,
	}
	suite.walletService.EXPECT().GetWalletWithBalance(uint(4), "http://localhost:8545").Return(walletWithBalance, nil)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importMnemonic()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	suite.Equal(stepConfirm, suite.model.currentStep)
	suite.NotNil(suite.model.confirmedWallet.Wallet.DerivationPath)
	suite.Equal(customPath, *suite.model.confirmedWallet.Wallet.DerivationPath)
}

// TestEmptyAliasError tests error when alias is empty.
func (suite *WalletAddPageTestSuite) TestEmptyAliasError() {
	// Navigate to private key entry without entering alias
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	// Don't set alias, leave it empty
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	testPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.pkeyInput.SetValue(testPrivateKey)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	suite.Equal(stepError, suite.model.currentStep)
	suite.Contains(suite.model.errorMsg, "alias cannot be empty")
}

// TestEmptyPrivateKeyError tests error when private key is empty.
func (suite *WalletAddPageTestSuite) TestEmptyPrivateKeyError() {
	// Navigate to private key entry
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.model.aliasInput.SetValue("test-wallet")
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Don't set private key, leave it empty
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	importMsg := suite.model.importPrivateKey()
	updatedModel, _ = suite.model.Update(importMsg)
	suite.model = updatedModel.(Model)

	suite.Equal(stepError, suite.model.currentStep)
	suite.Contains(suite.model.errorMsg, "private key cannot be empty")
}

// TestNavigationKeys tests up/down/esc navigation.
func (suite *WalletAddPageTestSuite) TestNavigationKeys() {
	// Test up/down navigation in method selection
	suite.Equal(0, suite.model.selectedIndex)

	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(2, suite.model.selectedIndex)

	// Can't go down past last option
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(2, suite.model.selectedIndex)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyUp})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyUp})
	suite.model = updatedModel.(Model)
	suite.Equal(0, suite.model.selectedIndex)

	// Can't go up past first option
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyUp})
	suite.model = updatedModel.(Model)
	suite.Equal(0, suite.model.selectedIndex)
}

// TestEscNavigation tests ESC key navigation to go back.
func (suite *WalletAddPageTestSuite) TestEscNavigation() {
	// Go to alias entry
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
	suite.Equal(stepEnterAlias, suite.model.currentStep)

	// Press ESC to go back
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	suite.model = updatedModel.(Model)
	suite.Equal(stepSelectMethod, suite.model.currentStep)
}

// TestViewRendering tests that all view rendering methods return non-empty strings.
func (suite *WalletAddPageTestSuite) TestViewRendering() {
	// Test each step's view rendering
	suite.model.currentStep = stepSelectMethod
	view := suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Add New Wallet")

	suite.model.currentStep = stepEnterAlias
	view = suite.model.View()
	suite.NotEmpty(view)

	suite.model.currentStep = stepEnterPrivateKey
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Private Key")

	suite.model.currentStep = stepEnterMnemonic
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Mnemonic")

	suite.model.currentStep = stepSelectDerivationPath
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Derivation Path")

	suite.model.currentStep = stepGenerating
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Generating")

	suite.model.currentStep = stepSuccess
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Success")

	suite.model.currentStep = stepError
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Error")
}

// TestHelpText tests that help text is provided for all steps.
func (suite *WalletAddPageTestSuite) TestHelpText() {
	suite.model.currentStep = stepSelectMethod
	help, _ := suite.model.Help()
	suite.NotEmpty(help)

	suite.model.currentStep = stepEnterAlias
	help, _ = suite.model.Help()
	suite.NotEmpty(help)
	suite.Contains(help, "enter")
	suite.Contains(help, "esc")

	suite.model.currentStep = stepEnterPrivateKey
	help, _ = suite.model.Help()
	suite.NotEmpty(help)

	suite.model.currentStep = stepEnterMnemonic
	help, _ = suite.model.Help()
	suite.NotEmpty(help)
	suite.Contains(help, "ctrl+s")

	suite.model.currentStep = stepSuccess
	help, _ = suite.model.Help()
	suite.NotEmpty(help)
}
