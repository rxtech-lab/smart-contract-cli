package deletewallet

import (
	"fmt"
	"math/big"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/wallet"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRouter implements view.Router interface.
type MockRouter struct {
	mock.Mock
}

func (m *MockRouter) AddRoute(route view.Route) {
	m.Called(route)
}

func (m *MockRouter) SetRoutes(routes []view.Route) {
	m.Called(routes)
}

func (m *MockRouter) RemoveRoute(path string) {
	m.Called(path)
}

func (m *MockRouter) GetRoutes() []view.Route {
	args := m.Called()
	return args.Get(0).([]view.Route)
}

func (m *MockRouter) GetCurrentRoute() view.Route {
	args := m.Called()
	return args.Get(0).(view.Route)
}

func (m *MockRouter) NavigateTo(path string, queryParams map[string]string) error {
	args := m.Called(path, queryParams)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

func (m *MockRouter) ReplaceRoute(path string) error {
	args := m.Called(path)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

func (m *MockRouter) Back() {
	m.Called()
}

func (m *MockRouter) CanGoBack() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRouter) GetParam(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockRouter) GetQueryParam(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockRouter) GetPath() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRouter) Refresh() {
	m.Called()
}

func (m *MockRouter) Init() tea.Cmd {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(tea.Cmd)
}

func (m *MockRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	args := m.Called(msg)
	return args.Get(0).(tea.Model), nil
}

func (m *MockRouter) View() string {
	args := m.Called()
	return args.String(0)
}

// MockWalletService implements wallet.WalletService interface.
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) ImportPrivateKey(alias, privateKey string) (*models.EVMWallet, error) {
	args := m.Called(alias, privateKey)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) ImportMnemonic(alias, mnemonic, derivationPath string) (*models.EVMWallet, error) {
	args := m.Called(alias, mnemonic, derivationPath)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) GenerateWallet(alias string) (*models.EVMWallet, string, string, error) {
	args := m.Called(alias)
	if args.Get(0) == nil {
		return nil, "", "", args.Error(3) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.String(1), args.String(2), args.Error(3) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) GetWalletWithBalance(walletID uint, rpcEndpoint string) (*wallet.WalletWithBalance, error) {
	args := m.Called(walletID, rpcEndpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*wallet.WalletWithBalance), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) ListWalletsWithBalances(page int64, pageSize int64, rpcEndpoint string) ([]wallet.WalletWithBalance, int64, error) {
	args := m.Called(page, pageSize, rpcEndpoint)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).([]wallet.WalletWithBalance), args.Get(1).(int64), args.Error(2) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) GetWallet(walletID uint) (*models.EVMWallet, error) {
	args := m.Called(walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) DeleteWallet(walletID uint) error {
	args := m.Called(walletID)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) WalletExistsByAlias(alias string) (bool, error) {
	args := m.Called(alias)
	return args.Bool(0), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) WalletExistsByAddress(address string) (bool, error) {
	args := m.Called(address)
	return args.Bool(0), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) ValidatePrivateKey(privateKey string) error {
	args := m.Called(privateKey)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) ValidateMnemonic(mnemonic string) error {
	args := m.Called(mnemonic)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) GetPrivateKey(walletID uint) (string, error) {
	args := m.Called(walletID)
	return args.String(0), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) GetMnemonic(walletID uint) (string, error) {
	args := m.Called(walletID)
	return args.String(0), args.Error(1) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) UpdateWalletAlias(walletID uint, newAlias string) error {
	args := m.Called(walletID, newAlias)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

func (m *MockWalletService) UpdateWalletPrivateKey(walletID uint, newPrivateKeyHex string) error {
	args := m.Called(walletID, newPrivateKeyHex)
	return args.Error(0) //nolint:wrapcheck // Mock method
}

// WalletDeletePageTestSuite is the test suite for the wallet delete page.
type WalletDeletePageTestSuite struct {
	suite.Suite
	router        *MockRouter
	sharedMemory  storage.SharedMemory
	walletService *MockWalletService
	model         Model
}

func TestWalletDeletePageTestSuite(t *testing.T) {
	suite.Run(t, new(WalletDeletePageTestSuite))
}

func (suite *WalletDeletePageTestSuite) SetupTest() {
	suite.router = new(MockRouter)
	suite.sharedMemory = storage.NewSharedMemory()
	suite.walletService = new(MockWalletService)

	// Create model with mocked wallet service
	page := NewPageWithService(suite.router, suite.sharedMemory, suite.walletService)
	suite.model = page.(Model)
}

func (suite *WalletDeletePageTestSuite) TearDownTest() {
	suite.router.AssertExpectations(suite.T())
	suite.walletService.AssertExpectations(suite.T())
}

// TestDeleteNonSelectedWallet tests deleting a wallet that is not currently selected.
func (suite *WalletDeletePageTestSuite) TestDeleteNonSelectedWallet() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("1")

	// Mock wallet service
	balance, _ := new(big.Int).SetString("5000000000000000000", 10) // 5 ETH
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil)

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify state
	suite.False(suite.model.loading)
	suite.Equal(uint(1), suite.model.walletID)
	suite.NotNil(suite.model.wallet)
	suite.Equal("test-wallet", suite.model.wallet.Wallet.Alias)
	suite.Equal(modeConfirmation, suite.model.mode)

	// Select "Yes, delete permanently" (index 1)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	// Mock deletion
	suite.walletService.On("DeleteWallet", uint(1)).Return(nil)

	// Confirm deletion
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Wait for deletion to complete
	deleteMsg := suite.model.deleteWallet()
	updatedModel, _ = suite.model.Update(deleteMsg)
	suite.model = updatedModel.(Model)

	// Verify deletion was successful (navigation command is returned)
}

// TestCannotDeleteSelectedWallet tests that the currently selected wallet cannot be deleted.
func (suite *WalletDeletePageTestSuite) TestCannotDeleteSelectedWallet() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("1")

	// Mock wallet service
	balance, _ := new(big.Int).SetString("10000000000000000000", 10) // 10 ETH
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "selected-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil)

	// Set this wallet as selected in shared memory
	_ = suite.sharedMemory.Set(config.SelectedWalletIDKey, uint(1))

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify state - should be in "cannot delete" mode
	suite.False(suite.model.loading)
	suite.Equal(uint(1), suite.model.walletID)
	suite.Equal(uint(1), suite.model.selectedWalletID)
	suite.Equal(modeCannotDelete, suite.model.mode)

	// Any key should navigate back
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)
}

// TestCancelDeletion tests canceling the deletion process.
func (suite *WalletDeletePageTestSuite) TestCancelDeletion() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("2")

	// Mock wallet service
	balance := big.NewInt(0)
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      2,
			Alias:   "wallet-to-cancel",
			Address: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(2), "http://localhost:8545").Return(testWallet, nil)

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify we're in confirmation mode
	suite.Equal(modeConfirmation, suite.model.mode)

	// Select "No, cancel" (index 0 - default)
	suite.Equal(0, suite.model.selectedIndex)

	// Press enter to cancel
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Should navigate back (returns a command)
}

// TestInvalidWalletID tests error handling when wallet ID is invalid.
func (suite *WalletDeletePageTestSuite) TestInvalidWalletID() {
	// Mock router to return invalid wallet ID
	suite.router.On("GetQueryParam", "id").Return("invalid")

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.False(suite.model.loading)
	suite.NotEmpty(suite.model.errorMsg)
	suite.Contains(suite.model.errorMsg, "invalid wallet ID")
}

// TestMissingWalletID tests error handling when wallet ID is not provided.
func (suite *WalletDeletePageTestSuite) TestMissingWalletID() {
	// Mock router to return empty wallet ID
	suite.router.On("GetQueryParam", "id").Return("")

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.False(suite.model.loading)
	suite.NotEmpty(suite.model.errorMsg)
	suite.Contains(suite.model.errorMsg, "wallet ID not provided")
}

// TestWalletNotFound tests error handling when wallet doesn't exist.
func (suite *WalletDeletePageTestSuite) TestWalletNotFound() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("999")

	// Mock wallet service to return error
	suite.walletService.On("GetWalletWithBalance", uint(999), "http://localhost:8545").
		Return(nil, fmt.Errorf("wallet not found"))

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.False(suite.model.loading)
	suite.NotEmpty(suite.model.errorMsg)
	suite.Contains(suite.model.errorMsg, "failed to load wallet")
}

// TestDeletionError tests error handling when deletion fails.
func (suite *WalletDeletePageTestSuite) TestDeletionError() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("3")

	// Mock wallet service
	balance := big.NewInt(0)
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      3,
			Alias:   "test-wallet",
			Address: "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(3), "http://localhost:8545").Return(testWallet, nil)

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Select "Yes, delete permanently"
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)

	// Mock deletion to fail
	suite.walletService.On("DeleteWallet", uint(3)).Return(fmt.Errorf("database error"))

	// Confirm deletion
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Wait for deletion to complete
	deleteMsg := suite.model.deleteWallet()
	updatedModel, _ = suite.model.Update(deleteMsg)
	suite.model = updatedModel.(Model)

	// Should show error
	suite.NotEmpty(suite.model.errorMsg)
	suite.Contains(suite.model.errorMsg, "database error")
}

// TestNavigationKeys tests up/down navigation.
func (suite *WalletDeletePageTestSuite) TestNavigationKeys() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("1")

	// Mock wallet service
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: big.NewInt(0),
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil)

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Test navigation
	suite.Equal(0, suite.model.selectedIndex) // Starts at "No, cancel"

	// Navigate down to "Yes, delete"
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	// Can't go down past last option
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyDown})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	// Navigate up
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyUp})
	suite.model = updatedModel.(Model)
	suite.Equal(0, suite.model.selectedIndex)

	// Can't go up past first option
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyUp})
	suite.model = updatedModel.(Model)
	suite.Equal(0, suite.model.selectedIndex)

	// Test vim keys (j/k)
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	suite.model = updatedModel.(Model)
	suite.Equal(1, suite.model.selectedIndex)

	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	suite.model = updatedModel.(Model)
	suite.Equal(0, suite.model.selectedIndex)
}

// TestViewRendering tests that all view modes render correctly.
func (suite *WalletDeletePageTestSuite) TestViewRendering() {
	// Test loading state
	suite.model.loading = true
	view := suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Loading")

	// Test error state
	suite.model.loading = false
	suite.model.errorMsg = "Test error"
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Error")
	suite.Contains(view, "Test error")

	// Test confirmation mode
	suite.model.errorMsg = ""
	suite.model.wallet = &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: big.NewInt(1000000000000000000),
	}
	suite.model.mode = modeConfirmation
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Delete Wallet")
	suite.Contains(view, "test-wallet")
	suite.Contains(view, "Are you sure")

	// Test cannot delete mode
	suite.model.mode = modeCannotDelete
	view = suite.model.View()
	suite.NotEmpty(view)
	suite.Contains(view, "Cannot delete currently selected wallet")
	suite.Contains(view, "Select another wallet")
}

// TestHelpText tests that help text is provided for all modes.
func (suite *WalletDeletePageTestSuite) TestHelpText() {
	// Test loading state
	suite.model.loading = true
	help, _ := suite.model.Help()
	suite.NotEmpty(help)
	suite.Contains(help, "Loading")

	// Test cannot delete mode
	suite.model.loading = false
	suite.model.mode = modeCannotDelete
	help, _ = suite.model.Help()
	suite.NotEmpty(help)
	suite.Contains(help, "Press any key")

	// Test confirmation mode
	suite.model.mode = modeConfirmation
	help, _ = suite.model.Help()
	suite.NotEmpty(help)
	suite.Contains(help, "enter")
}

// TestBalanceFormatting tests that balance is formatted correctly in views.
func (suite *WalletDeletePageTestSuite) TestBalanceFormatting() {
	suite.model.loading = false
	suite.model.mode = modeConfirmation

	// Test with balance
	balance, _ := new(big.Int).SetString("1234567890123456789", 10)
	suite.model.wallet = &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}

	view := suite.model.View()
	suite.Contains(view, "1.2346 ETH")

	// Test with balance error
	suite.model.wallet.Error = fmt.Errorf("RPC error")
	suite.model.wallet.Balance = nil
	view = suite.model.View()
	suite.Contains(view, "unavailable")
}

// TestInitialState tests that the model is initialized correctly.
func (suite *WalletDeletePageTestSuite) TestInitialState() {
	page := NewPageWithService(suite.router, suite.sharedMemory, suite.walletService)
	model := page.(Model)

	suite.True(model.loading)
	suite.Equal(modeConfirmation, model.mode)
	suite.Equal(0, model.selectedIndex)
	suite.Len(model.options, 2)
	suite.Equal("No, cancel", model.options[0].label)
	suite.Equal("Yes, delete permanently", model.options[1].label)
	suite.False(model.options[0].value)
	suite.True(model.options[1].value)
}
