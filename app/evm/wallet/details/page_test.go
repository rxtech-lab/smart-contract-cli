package details

import (
	"fmt"
	"math/big"
	"testing"
	"time"

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

func (m *MockRouter) GetCurrentRoute() view.Route {
	args := m.Called()
	return args.Get(0).(view.Route) //nolint:forcetypeassert // Mock method
}

func (m *MockRouter) GetRoutes() []view.Route {
	args := m.Called()
	return args.Get(0).([]view.Route) //nolint:forcetypeassert // Mock method
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

func (m *MockRouter) GetQueryParam(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockRouter) GetParam(key string) string {
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

func (m *MockRouter) View() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRouter) Init() tea.Cmd {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(tea.Cmd) //nolint:forcetypeassert // Mock method
}

func (m *MockRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	args := m.Called(msg)
	return args.Get(0).(tea.Model), nil //nolint:forcetypeassert // Mock method
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
	return args.Get(0).(*models.EVMWallet), args.Error(1) //nolint:wrapcheck,forcetypeassert // Mock method
}

func (m *MockWalletService) ImportMnemonic(alias, mnemonic, derivationPath string) (*models.EVMWallet, error) {
	args := m.Called(alias, mnemonic, derivationPath)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.Error(1) //nolint:wrapcheck,forcetypeassert // Mock method
}

func (m *MockWalletService) GenerateWallet(alias string) (*models.EVMWallet, string, string, error) {
	args := m.Called(alias)
	if args.Get(0) == nil {
		return nil, "", "", args.Error(3) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.String(1), args.String(2), args.Error(3) //nolint:wrapcheck,forcetypeassert // Mock method
}

func (m *MockWalletService) GetWalletWithBalance(walletID uint, rpcEndpoint string) (*wallet.WalletWithBalance, error) {
	args := m.Called(walletID, rpcEndpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*wallet.WalletWithBalance), args.Error(1) //nolint:wrapcheck,forcetypeassert // Mock method
}

func (m *MockWalletService) ListWalletsWithBalances(page int64, pageSize int64, rpcEndpoint string) ([]wallet.WalletWithBalance, int64, error) {
	args := m.Called(page, pageSize, rpcEndpoint)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).([]wallet.WalletWithBalance), args.Get(1).(int64), args.Error(2) //nolint:wrapcheck,forcetypeassert // Mock method
}

func (m *MockWalletService) GetWallet(walletID uint) (*models.EVMWallet, error) {
	args := m.Called(walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1) //nolint:wrapcheck // Mock method
	}
	return args.Get(0).(*models.EVMWallet), args.Error(1) //nolint:wrapcheck,forcetypeassert // Mock method
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

// WalletDetailsPageTestSuite is the test suite for wallet details page.
type WalletDetailsPageTestSuite struct {
	suite.Suite
	router        *MockRouter
	sharedMemory  storage.SharedMemory
	walletService *MockWalletService
	model         Model
}

func TestWalletDetailsPageTestSuite(t *testing.T) {
	suite.Run(t, new(WalletDetailsPageTestSuite))
}

func (suite *WalletDetailsPageTestSuite) SetupTest() {
	suite.router = new(MockRouter)
	suite.sharedMemory = storage.NewSharedMemory()
	suite.walletService = new(MockWalletService)

	// Set up shared memory with config
	_ = suite.sharedMemory.Set(config.SelectedWalletIDKey, uint(2))

	// Create model with mocked dependencies
	page := NewPageWithService(suite.router, suite.sharedMemory, suite.walletService)
	suite.model = page.(Model)
}

func (suite *WalletDetailsPageTestSuite) TearDownTest() {
	suite.router.AssertExpectations(suite.T())
	suite.walletService.AssertExpectations(suite.T())
}

// TestLoadWalletDetails tests loading wallet details successfully.
func (suite *WalletDetailsPageTestSuite) TestLoadWalletDetails() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("1")

	// Mock wallet service
	balance, _ := new(big.Int).SetString("5000000000000000000", 10) // 5 ETH
	derivationPath := "m/44'/60'/0'/0/0"
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:             1,
			Alias:          "my-wallet",
			Address:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			IsFromMnemonic: true,
			DerivationPath: &derivationPath,
			CreatedAt:      time.Now().Add(-24 * time.Hour),
			UpdatedAt:      time.Now(),
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
	suite.Equal(uint(2), suite.model.selectedWalletID)
	suite.Equal("my-wallet", suite.model.wallet.Wallet.Alias)
	suite.Equal(modeNormal, suite.model.mode)

	// Verify view contains expected information
	view := suite.model.View()
	suite.Contains(view, "my-wallet")
	suite.Contains(view, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	suite.Contains(view, "5.0000 ETH")
	suite.Contains(view, "m/44'/60'/0'/0/0")
}

// TestMissingWalletID tests error handling when wallet ID is not provided.
func (suite *WalletDetailsPageTestSuite) TestMissingWalletID() {
	// Mock router to return empty wallet ID
	suite.router.On("GetQueryParam", "id").Return("")

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify error state
	suite.False(suite.model.loading)
	suite.Contains(suite.model.errorMsg, "wallet ID not provided")
}

// TestInvalidWalletID tests error handling when wallet ID is invalid.
func (suite *WalletDetailsPageTestSuite) TestInvalidWalletID() {
	// Mock router to return invalid wallet ID
	suite.router.On("GetQueryParam", "id").Return("invalid")

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify error state
	suite.False(suite.model.loading)
	suite.Contains(suite.model.errorMsg, "invalid wallet ID")
}

// TestWalletNotFound tests error handling when wallet is not found.
func (suite *WalletDetailsPageTestSuite) TestWalletNotFound() {
	// Mock router to return wallet ID
	suite.router.On("GetQueryParam", "id").Return("999")

	// Mock wallet service to return error
	suite.walletService.On("GetWalletWithBalance", uint(999), "http://localhost:8545").
		Return(nil, fmt.Errorf("wallet not found"))

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify error state
	suite.False(suite.model.loading)
	suite.Contains(suite.model.errorMsg, "failed to load wallet")
}

// TestShowPrivateKeyPrompt tests entering private key prompt mode.
func (suite *WalletDetailsPageTestSuite) TestShowPrivateKeyPrompt() {
	// Setup wallet first
	suite.router.On("GetQueryParam", "id").Return("1")
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
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

	// Press 'p' to show private key prompt
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	suite.model = updatedModel.(Model)

	// Verify mode changed
	suite.Equal(modeShowPrivateKeyPrompt, suite.model.mode)

	// Verify view shows warning
	view := suite.model.View()
	suite.Contains(view, "Security Warning")
	suite.Contains(view, "WARNING")
	suite.Contains(view, "Type \"SHOW\" to reveal")
}

// TestCancelPrivateKeyPrompt tests canceling private key prompt.
func (suite *WalletDetailsPageTestSuite) TestCancelPrivateKeyPrompt() {
	// Setup wallet and enter prompt mode
	suite.router.On("GetQueryParam", "id").Return("1")
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil)

	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Enter prompt mode
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	suite.model = updatedModel.(Model)

	// Press 'esc' to cancel
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	suite.model = updatedModel.(Model)

	// Verify mode changed back to normal
	suite.Equal(modeNormal, suite.model.mode)
}

// TestShowPrivateKeyWithCorrectConfirmation tests showing private key with correct confirmation.
func (suite *WalletDetailsPageTestSuite) TestShowPrivateKeyWithCorrectConfirmation() {
	// Setup wallet and enter prompt mode
	suite.router.On("GetQueryParam", "id").Return("1")
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil)

	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Enter prompt mode
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	suite.model = updatedModel.(Model)

	// Type "SHOW"
	suite.model.confirmationInput.SetValue("SHOW")

	// Mock private key loading
	suite.walletService.On("GetPrivateKey", uint(1)).
		Return("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", nil)

	// Press enter
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Handle private key loaded message
	privateKeyMsg := suite.model.loadPrivateKey()
	updatedModel, _ = suite.model.Update(privateKeyMsg)
	suite.model = updatedModel.(Model)

	// Verify mode changed to show private key
	suite.Equal(modeShowPrivateKey, suite.model.mode)
	suite.Equal("0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", suite.model.privateKey)
	suite.Equal(60, suite.model.autoCloseCounter)

	// Verify view shows private key
	view := suite.model.View()
	suite.Contains(view, "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	suite.Contains(view, "SENSITIVE INFORMATION")
	suite.Contains(view, "automatically close in 60 seconds")
}

// TestShowPrivateKeyWithIncorrectConfirmation tests showing error with incorrect confirmation.
func (suite *WalletDetailsPageTestSuite) TestShowPrivateKeyWithIncorrectConfirmation() {
	// Setup wallet and enter prompt mode
	suite.router.On("GetQueryParam", "id").Return("1")
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil)

	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Enter prompt mode
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	suite.model = updatedModel.(Model)

	// Type wrong confirmation
	suite.model.confirmationInput.SetValue("show")

	// Press enter
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	suite.model = updatedModel.(Model)

	// Verify error message
	suite.Contains(suite.model.errorMsg, "Incorrect confirmation")
	suite.Equal(modeShowPrivateKeyPrompt, suite.model.mode)
}

// TestClosePrivateKeyView tests closing private key view manually.
func (suite *WalletDetailsPageTestSuite) TestClosePrivateKeyView() {
	// Setup model in showPrivateKey mode
	suite.model.loading = false
	suite.model.mode = modeShowPrivateKey
	suite.model.privateKey = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.autoCloseCounter = 60
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	suite.model.wallet = &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}

	// Press 'q' to close
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	suite.model = updatedModel.(Model)

	// Verify mode changed back to normal and private key cleared
	suite.Equal(modeNormal, suite.model.mode)
	suite.Equal("", suite.model.privateKey)
}

// TestAutoClosePrivateKeyView tests auto-closing private key view after countdown.
func (suite *WalletDetailsPageTestSuite) TestAutoClosePrivateKeyView() {
	// Setup model in showPrivateKey mode
	suite.model.loading = false
	suite.model.mode = modeShowPrivateKey
	suite.model.privateKey = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.model.autoCloseCounter = 1
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	suite.model.wallet = &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}

	// Simulate auto-close tick
	updatedModel, _ := suite.model.Update(autoCloseTickMsg{})
	suite.model = updatedModel.(Model)

	// Verify counter decremented and mode changed back to normal
	suite.Equal(modeNormal, suite.model.mode)
	suite.Equal("", suite.model.privateKey)
}

// TestRefreshBalance tests refreshing wallet balance.
func (suite *WalletDetailsPageTestSuite) TestRefreshBalance() {
	// Setup wallet first
	suite.router.On("GetQueryParam", "id").Return("1").Times(2)
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(1), "http://localhost:8545").Return(testWallet, nil).Times(2)

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Press 'r' to refresh
	updatedModel, _ = suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	suite.model = updatedModel.(Model)

	// Verify loading state
	suite.True(suite.model.loading)

	// Handle refresh message
	loadMsg = suite.model.loadWallet()
	updatedModel, _ = suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify wallet reloaded
	suite.False(suite.model.loading)
	suite.Equal(uint(1), suite.model.walletID)
}

// TestViewRenderingWithSelectedWallet tests view rendering when wallet is selected.
func (suite *WalletDetailsPageTestSuite) TestViewRenderingWithSelectedWallet() {
	// Setup wallet as selected
	suite.router.On("GetQueryParam", "id").Return("2")
	balance, _ := new(big.Int).SetString("3000000000000000000", 10) // 3 ETH
	emptyPath := ""
	now := time.Now()
	testWallet := &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:             2,
			Alias:          "selected-wallet",
			Address:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			DerivationPath: &emptyPath,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Balance: balance,
	}
	suite.walletService.On("GetWalletWithBalance", uint(2), "http://localhost:8545").Return(testWallet, nil)

	// Load wallet
	loadMsg := suite.model.loadWallet()
	updatedModel, _ := suite.model.Update(loadMsg)
	suite.model = updatedModel.(Model)

	// Verify view shows selected status
	view := suite.model.View()
	suite.Contains(view, "★ Currently Selected")
}

// TestViewRenderingWithBalanceError tests view rendering when balance fetch fails.
func (suite *WalletDetailsPageTestSuite) TestViewRenderingWithBalanceError() {
	// Setup wallet with balance error
	suite.model.loading = false
	suite.model.mode = modeNormal
	suite.model.walletID = 1
	suite.model.selectedWalletID = 2
	emptyPath := ""
	now := time.Now()
	suite.model.wallet = &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:             1,
			Alias:          "test-wallet",
			Address:        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			DerivationPath: &emptyPath,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Balance: nil,
		Error:   fmt.Errorf("RPC connection failed"),
	}

	// Verify view shows balance unavailable
	view := suite.model.View()
	suite.Contains(view, "unavailable")
}

// TestHelpText tests help text in different modes.
func (suite *WalletDetailsPageTestSuite) TestHelpText() {
	// Test normal mode
	suite.model.loading = false
	suite.model.mode = modeNormal
	helpText, _ := suite.model.Help()
	suite.Contains(helpText, "refresh balance")
	suite.Contains(helpText, "show private key")

	// Test private key prompt mode
	suite.model.mode = modeShowPrivateKeyPrompt
	helpText, _ = suite.model.Help()
	suite.Contains(helpText, "confirm")
	suite.Contains(helpText, "cancel")

	// Test show private key mode
	suite.model.mode = modeShowPrivateKey
	helpText, _ = suite.model.Help()
	suite.Contains(helpText, "copy to clipboard")
	suite.Contains(helpText, "close immediately")

	// Test loading mode
	suite.model.loading = true
	helpText, _ = suite.model.Help()
	suite.Contains(helpText, "Loading")
}

// TestInitialState tests that the model is initialized correctly.
func (suite *WalletDetailsPageTestSuite) TestInitialState() {
	page := NewPageWithService(suite.router, suite.sharedMemory, suite.walletService)
	model := page.(Model)

	suite.True(model.loading)
	suite.Equal(modeNormal, model.mode)
	suite.Equal("Type 'SHOW' to confirm", model.confirmationInput.Placeholder)
}

// TestPrivateKeyLoadError tests error handling when loading private key fails.
func (suite *WalletDetailsPageTestSuite) TestPrivateKeyLoadError() {
	// Setup model
	suite.model.loading = false
	suite.model.mode = modeShowPrivateKeyPrompt
	suite.model.walletID = 1
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	suite.model.wallet = &wallet.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "test-wallet",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
		Balance: balance,
	}

	// Mock private key loading error
	suite.walletService.On("GetPrivateKey", uint(1)).Return("", fmt.Errorf("decryption failed"))

	// Load private key
	privateKeyMsg := suite.model.loadPrivateKey()
	updatedModel, _ := suite.model.Update(privateKeyMsg)
	suite.model = updatedModel.(Model)

	// Verify error state
	suite.Contains(suite.model.errorMsg, "decryption failed")
	suite.Equal(modeNormal, suite.model.mode)
	suite.Equal("", suite.model.privateKey)
}

// TestIgnoreKeysWhileLoading tests that keys are ignored while loading.
func (suite *WalletDetailsPageTestSuite) TestIgnoreKeysWhileLoading() {
	suite.model.loading = true

	// Try pressing keys
	updatedModel, _ := suite.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	suite.model = updatedModel.(Model)

	// Verify nothing changed
	suite.Equal(modeNormal, suite.model.mode)
}
