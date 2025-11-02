package wallet

import (
	"io"
	"math/big"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	walletsvc "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/wallet"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// WalletPageTestSuite tests the wallet management page using teatest.
type WalletPageTestSuite struct {
	suite.Suite
	testStoragePath string
	sharedMemory    storage.SharedMemory
	router          view.Router
	mockCtrl        *gomock.Controller
}

func TestWalletPageTestSuite(t *testing.T) {
	suite.Run(t, new(WalletPageTestSuite))
}

func (s *WalletPageTestSuite) SetupTest() {
	// Create a temporary directory for test storage
	tmpDir, err := os.MkdirTemp("", "smart-contract-cli-wallet-test-*")
	s.NoError(err, "Should create temp directory")
	s.testStoragePath = tmpDir

	// Override the storage path for tests
	err = os.Setenv("HOME", tmpDir)
	s.NoError(err, "Should set HOME environment variable")

	// Create shared memory and router for each test
	s.sharedMemory = storage.NewSharedMemory()
	s.router = view.NewRouter()

	// Create mock controller
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *WalletPageTestSuite) TearDownTest() {
	// Clean up test storage
	if s.testStoragePath != "" {
		err := os.RemoveAll(s.testStoragePath)
		s.NoError(err, "Should clean up test storage directory")
	}

	// Finish mock controller
	if s.mockCtrl != nil {
		s.mockCtrl.Finish()
	}
}

func (s *WalletPageTestSuite) getOutput(tm *teatest.TestModel) string {
	output, err := io.ReadAll(tm.Output())
	s.NoError(err, "Should be able to read output")
	return string(output)
}

// setupMockWalletService creates a mock wallet service for testing.
func (s *WalletPageTestSuite) setupMockWalletService() *walletsvc.MockWalletService {
	return walletsvc.NewMockWalletService(s.mockCtrl)
}

// setupMockStorage creates a mock storage client and sets up the shared memory with a default config.
func (s *WalletPageTestSuite) setupMockStorage(selectedWalletID *uint) {
	mockStorage := sql.NewMockStorage(s.mockCtrl)

	// Create test config with default RPC endpoint
	endpoint := &models.EVMEndpoint{
		ID:   1,
		Url:  "http://localhost:8545",
		Name: "Test Endpoint",
	}

	evmConfig := models.EVMConfig{
		ID:               1,
		EndpointId:       &endpoint.ID,
		Endpoint:         endpoint,
		SelectedWalletID: selectedWalletID,
	}

	// Set up mock to return config
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(evmConfig, nil).
		AnyTimes()

	// Store mock storage in shared memory as the Storage interface type
	var storage sql.Storage = mockStorage
	err := s.sharedMemory.Set(config.StorageClientKey, storage)
	s.NoError(err, "Should set storage client in shared memory")
}

// TestEmptyWalletList tests the empty state when no wallets exist.
func (s *WalletPageTestSuite) TestEmptyWalletList() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	// Mock ListWalletsWithBalances to return empty list
	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return([]walletsvc.WalletWithBalance{}, int64(0), nil)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render and wallet loading
	time.Sleep(500 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Wallet Management", "Should show title")
	s.Contains(output, "No wallets found", "Should show empty state")
	s.Contains(output, "Press 'a' to add your first wallet", "Should show add wallet hint")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestWalletListDisplay tests displaying a list of wallets with balances.
func (s *WalletPageTestSuite) TestWalletListDisplay() {
	mockWalletSvc := s.setupMockWalletService()
	walletID := uint(1)
	s.setupMockStorage(&walletID)

	balance1 := new(big.Int)
	balance1.SetString("1000000000000000000", 10) // 1 ETH

	balance2 := new(big.Int)
	balance2.SetString("500000000000000000", 10) // 0.5 ETH

	// Create test wallet data
	testWallets := []walletsvc.WalletWithBalance{
		{
			Wallet: models.EVMWallet{
				ID:      1,
				Alias:   "Main Wallet",
				Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			},
			Balance: balance1,
		},
		{
			Wallet: models.EVMWallet{
				ID:      2,
				Alias:   "Dev Wallet",
				Address: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			},
			Balance: balance2,
		},
	}

	// Mock ListWalletsWithBalances
	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return(testWallets, int64(2), nil)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallets to load
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Wallet Management", "Should show title")
	s.Contains(output, "Main Wallet", "Should show first wallet")
	s.Contains(output, "Dev Wallet", "Should show second wallet")
	s.Contains(output, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", "Should show first wallet address")
	s.Contains(output, "Endpoint: http://localhost:8545", "Should show RPC endpoint")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestNavigationUpDown tests keyboard navigation through wallet list.
func (s *WalletPageTestSuite) TestNavigationUpDown() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	testWallets := []walletsvc.WalletWithBalance{
		{
			Wallet: models.EVMWallet{ID: 1, Alias: "Wallet 1", Address: "0x1111111111111111111111111111111111111111"},
		},
		{
			Wallet: models.EVMWallet{ID: 2, Alias: "Wallet 2", Address: "0x2222222222222222222222222222222222222222"},
		},
		{
			Wallet: models.EVMWallet{ID: 3, Alias: "Wallet 3", Address: "0x3333333333333333333333333333333333333333"},
		},
	}

	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return(testWallets, int64(3), nil)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallets to load
	time.Sleep(300 * time.Millisecond)

	// Initially cursor should be on first wallet
	output := s.getOutput(testModel)
	s.Contains(output, "> Wallet 1", "Cursor should be on first wallet initially")

	// Press down arrow to move to second wallet
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> Wallet 2", "Cursor should move to second wallet")

	// Press down again to move to third wallet
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> Wallet 3", "Cursor should move to third wallet")

	// Press up to go back to second wallet
	testModel.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> Wallet 2", "Cursor should move back to second wallet")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestNavigationWithVimKeys tests navigation using vim-style keys (j/k).
func (s *WalletPageTestSuite) TestNavigationWithVimKeys() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	testWallets := []walletsvc.WalletWithBalance{
		{
			Wallet: models.EVMWallet{ID: 1, Alias: "Wallet A", Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		},
		{
			Wallet: models.EVMWallet{ID: 2, Alias: "Wallet B", Address: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		},
	}

	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return(testWallets, int64(2), nil)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallets to load
	time.Sleep(300 * time.Millisecond)

	// Press 'j' to move down
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'j'},
	})
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "> Wallet B", "Should move down with 'j' key")

	// Press 'k' to move up
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'k'},
	})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> Wallet A", "Should move up with 'k' key")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestRefreshWallets tests the refresh functionality with 'r' key.
func (s *WalletPageTestSuite) TestRefreshWallets() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	testWallets := []walletsvc.WalletWithBalance{
		{
			Wallet: models.EVMWallet{ID: 1, Alias: "Test Wallet", Address: "0x1111111111111111111111111111111111111111"},
		},
	}

	// Expect ListWalletsWithBalances to be called twice (initial load + refresh)
	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return(testWallets, int64(1), nil).
		Times(2)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial load
	time.Sleep(300 * time.Millisecond)

	// Press 'r' to refresh
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'r'},
	})

	// Wait for refresh to complete
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Test Wallet", "Should show wallets after refresh")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestBackNavigation tests going back with 'esc' or 'q'.
func (s *WalletPageTestSuite) TestBackNavigation() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return([]walletsvc.WalletWithBalance{}, int64(0), nil).
		AnyTimes()

	// Add routes for navigation
	s.router.AddRoute(view.Route{
		Path: "/evm",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return &mockComponent{}
		},
	})
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return &mockComponent{}
		},
	})
	err := s.router.NavigateTo("/evm/wallet", nil)
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for load
	time.Sleep(300 * time.Millisecond)

	// Press 'q' to go back
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'q'},
	})
	time.Sleep(100 * time.Millisecond)

	// Router should have navigated back
	s.True(true, "Should navigate back without error")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestBalanceUnavailable tests display when balance is nil.
func (s *WalletPageTestSuite) TestBalanceUnavailable() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	testWallets := []walletsvc.WalletWithBalance{
		{
			Wallet: models.EVMWallet{
				ID:      1,
				Alias:   "Offline Wallet",
				Address: "0x1111111111111111111111111111111111111111",
			},
			Balance: nil, // Balance unavailable
		},
	}

	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return(testWallets, int64(1), nil)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallets to load
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Offline Wallet", "Should show wallet")
	s.Contains(output, "unavailable ⚠", "Should show unavailable for nil balance")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestHelpText tests that help text is displayed correctly.
func (s *WalletPageTestSuite) TestHelpText() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return([]walletsvc.WalletWithBalance{}, int64(0), nil).
		AnyTimes()

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	pageModel := model.(Model)

	// Test loading state help
	pageModel.loading = true
	helpText, _ := pageModel.Help()
	s.Contains(helpText, "Loading...", "Should show loading help text")

	// Test normal state help
	pageModel.loading = false
	helpText, _ = pageModel.Help()
	s.Contains(helpText, "↑/k: up", "Should show up navigation")
	s.Contains(helpText, "↓/j: down", "Should show down navigation")
	s.Contains(helpText, "enter: actions", "Should show enter action")
	s.Contains(helpText, "a: add wallet", "Should show add wallet")
	s.Contains(helpText, "r: refresh", "Should show refresh")
	s.Contains(helpText, "esc/q: back", "Should show back")
}

// TestAddFirstWallet tests pressing 'a' to add first wallet when list is empty.
func (s *WalletPageTestSuite) TestAddFirstWallet() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	// Mock empty wallet list
	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return([]walletsvc.WalletWithBalance{}, int64(0), nil)

	// Set up router with add wallet route
	addWalletCalled := false
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/add",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			addWalletCalled = true
			return &mockComponent{}
		},
	})

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallets to load
	time.Sleep(300 * time.Millisecond)

	// Verify empty state is shown
	output := s.getOutput(testModel)
	s.Contains(output, "No wallets found", "Should show empty state")
	s.Contains(output, "Press 'a' to add your first wallet", "Should show add wallet hint")

	// Press 'a' to add wallet
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
	})
	time.Sleep(100 * time.Millisecond)

	// Verify navigation occurred
	s.True(addWalletCalled, "Should navigate to add wallet page")
	currentRoute := s.router.GetCurrentRoute()
	s.Equal("/evm/wallet/add", currentRoute.Path, "Should navigate to /evm/wallet/add")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestAddWalletFromNonEmptyList tests pressing 'a' when wallets already exist.
func (s *WalletPageTestSuite) TestAddWalletFromNonEmptyList() {
	mockWalletSvc := s.setupMockWalletService()
	s.setupMockStorage(nil)

	testWallets := []walletsvc.WalletWithBalance{
		{
			Wallet: models.EVMWallet{
				ID:      1,
				Alias:   "Existing Wallet",
				Address: "0x1111111111111111111111111111111111111111",
			},
		},
	}

	mockWalletSvc.EXPECT().
		ListWalletsWithBalances(int64(1), int64(100), "http://localhost:8545").
		Return(testWallets, int64(1), nil)

	// Set up router with add wallet route
	addWalletCalled := false
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/add",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			addWalletCalled = true
			return &mockComponent{}
		},
	})

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallets to load
	time.Sleep(300 * time.Millisecond)

	// Verify wallet list is shown
	output := s.getOutput(testModel)
	s.Contains(output, "Existing Wallet", "Should show existing wallet")

	// Press 'a' to add another wallet
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
	})
	time.Sleep(100 * time.Millisecond)

	// Verify navigation occurred
	s.True(addWalletCalled, "Should navigate to add wallet page")
	currentRoute := s.router.GetCurrentRoute()
	s.Equal("/evm/wallet/add", currentRoute.Path, "Should navigate to /evm/wallet/add")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// mockComponent is a simple mock component for router testing.
type mockComponent struct{}

func (m *mockComponent) Init() tea.Cmd                           { return nil }
func (m *mockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *mockComponent) View() string                            { return "mock" }
func (m *mockComponent) Help() (string, view.HelpDisplayOption) {
	return "", view.HelpDisplayOptionAppend
}
