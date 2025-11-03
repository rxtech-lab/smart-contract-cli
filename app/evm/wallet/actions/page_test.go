package actions

import (
	"fmt"
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

// ActionsPageTestSuite tests the wallet actions page using teatest.
type ActionsPageTestSuite struct {
	suite.Suite
	testStoragePath string
	sharedMemory    storage.SharedMemory
	router          view.Router
	mockCtrl        *gomock.Controller
}

func TestActionsPageTestSuite(t *testing.T) {
	suite.Run(t, new(ActionsPageTestSuite))
}

func (s *ActionsPageTestSuite) SetupTest() {
	// Create a temporary directory for test storage
	tmpDir, err := os.MkdirTemp("", "smart-contract-cli-actions-test-*")
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

func (s *ActionsPageTestSuite) TearDownTest() {
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

func (s *ActionsPageTestSuite) getOutput(tm *teatest.TestModel) string {
	output, err := io.ReadAll(tm.Output())
	s.NoError(err, "Should be able to read output")
	return string(output)
}

// setupMockWalletService creates a mock wallet service for testing.
func (s *ActionsPageTestSuite) setupMockWalletService() *walletsvc.MockWalletService {
	return walletsvc.NewMockWalletService(s.mockCtrl)
}

// setupMockStorage creates a mock storage for testing.
func (s *ActionsPageTestSuite) setupMockStorage() *sql.MockStorage {
	return sql.NewMockStorage(s.mockCtrl)
}

// createTestWallet creates a test wallet with balance.
func (s *ActionsPageTestSuite) createTestWallet(alias string, balanceEth string) *walletsvc.WalletWithBalance {
	balance := new(big.Int)
	if balanceEth != "" {
		balance.SetString(balanceEth, 10)
	}

	return &walletsvc.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   alias,
			Address: "0x1111111111111111111111111111111111111111",
		},
		Balance: balance,
	}
}

// TestNoConfigFound_ShowError tests error when GetCurrentConfig returns error.
func (s *ActionsPageTestSuite) TestNoConfigFound_ShowError() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig to return error
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(models.EVMConfig{}, fmt.Errorf("no current config found"))

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Error:", "Should show error message")
	s.Contains(output, "failed to get current config", "Should show config error")
}

// TestNoEndpointConfigured_ShowError tests error when config has nil endpoint.
func (s *ActionsPageTestSuite) TestNoEndpointConfigured_ShowError() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig to return config with nil endpoint
	configNoEndpoint := models.EVMConfig{
		ID:               1,
		Endpoint:         nil,
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(configNoEndpoint, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Error:", "Should show error message")
}

// TestNoSelectedWalletInConfig_ShowError tests error when config has nil SelectedWalletID.
func (s *ActionsPageTestSuite) TestNoSelectedWalletInConfig_ShowError() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig to return config with nil SelectedWalletID
	configNoSelected := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: nil,
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(configNoSelected, nil)

	// Mock wallet service - it will be called before the SelectedWalletID check
	testWallet := s.createTestWallet("Test Wallet", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Error:", "Should show error message")
	s.Contains(output, "no selected wallet ID found in config", "Should show selected wallet error")
}

// TestSelectActiveWallet tests displaying "Select as active wallet" option when wallet is not selected.
func (s *ActionsPageTestSuite) TestSelectActiveWallet() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig - wallet 1 is being viewed, but wallet 2 is selected
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(2), // Different wallet is selected
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service to return wallet data
	testWallet := s.createTestWallet("Test Wallet", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/select",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return &mockComponent{}
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Test Wallet", "Should show wallet name")
	s.Contains(output, "Select as active wallet", "Should show select option")
	s.NotContains(output, "★", "Should not show star indicator for non-selected wallet")

	// Press enter to select the wallet
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Verify navigation occurred
	currentRoute := s.router.GetCurrentRoute()
	s.Equal("/evm/wallet/select", currentRoute.Path, "Should navigate to select page")
	s.Equal("1", s.router.GetQueryParam("id"), "Should pass wallet ID")
}

// TestWalletIsSelected_NoSelectOption tests that "Select as active wallet" is hidden when wallet is selected.
func (s *ActionsPageTestSuite) TestWalletIsSelected_NoSelectOption() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig - wallet 1 is both being viewed AND selected
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1), // Same wallet is selected
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service to return wallet data
	testWallet := s.createTestWallet("Selected Wallet", "2000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Selected Wallet", "Should show wallet name")
	s.Contains(output, "★", "Should show star indicator for selected wallet")
	s.NotContains(output, "Select as active wallet", "Should NOT show select option")
	s.Contains(output, "This is your currently selected wallet", "Should show selected wallet note")
	s.Contains(output, "You cannot deselect the active wallet", "Should show deselect warning")

	// First option should be "View details" (not "Select as active wallet")
	s.Contains(output, "> View details", "First option should be View details")
}

// TestNavigationUpDown tests keyboard navigation through action options.
func (s *ActionsPageTestSuite) TestNavigationUpDown() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(2),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service
	testWallet := s.createTestWallet("Nav Test", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	// Initially cursor should be on first option
	output := s.getOutput(testModel)
	s.Contains(output, "> Select as active wallet", "Cursor should be on first option")

	// Press down arrow to move to second option
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> View details", "Cursor should move to View details")

	// Press down again to move to third option
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> Update wallet", "Cursor should move to Update wallet")

	// Press up to go back to second option
	testModel.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> View details", "Cursor should move back to View details")
}

// TestNavigationWithVimKeys tests navigation using vim-style keys (j/k).
func (s *ActionsPageTestSuite) TestNavigationWithVimKeys() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(2),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service
	testWallet := s.createTestWallet("Vim Test", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	// Press 'j' to move down
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'j'},
	})
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "> View details", "Should move down with 'j' key")

	// Press 'k' to move up
	testModel.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'k'},
	})
	time.Sleep(100 * time.Millisecond)

	output = s.getOutput(testModel)
	s.Contains(output, "> Select as active wallet", "Should move up with 'k' key")
}

// TestViewDetailsNavigation tests navigating to view details page.
func (s *ActionsPageTestSuite) TestViewDetailsNavigation() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service
	testWallet := s.createTestWallet("Details Test", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add routes
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/details",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return &mockComponent{}
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	// First option should be "View details" (wallet is selected, so no "Select" option)
	output := s.getOutput(testModel)
	s.Contains(output, "> View details", "First option should be View details")

	// Press enter to navigate
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Verify navigation
	currentRoute := s.router.GetCurrentRoute()
	s.Equal("/evm/wallet/details", currentRoute.Path, "Should navigate to details page")
	s.Equal("1", s.router.GetQueryParam("id"), "Should pass wallet ID")
}

// TestUpdateWalletNavigation tests navigating to update wallet page.
func (s *ActionsPageTestSuite) TestUpdateWalletNavigation() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service
	testWallet := s.createTestWallet("Update Test", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add routes
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/update",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return &mockComponent{}
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	// Navigate to "Update wallet" option (second option when wallet is selected)
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "> Update wallet", "Should be on Update wallet option")

	// Press enter to navigate
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Verify navigation
	currentRoute := s.router.GetCurrentRoute()
	s.Equal("/evm/wallet/update", currentRoute.Path, "Should navigate to update page")
	s.Equal("1", s.router.GetQueryParam("id"), "Should pass wallet ID")
}

// TestDeleteWalletNavigation tests navigating to delete wallet page.
func (s *ActionsPageTestSuite) TestDeleteWalletNavigation() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service
	testWallet := s.createTestWallet("Delete Test", "1000000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add routes
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/delete",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return &mockComponent{}
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	// Navigate to "Delete wallet" option (third option when wallet is selected)
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "> Delete wallet", "Should be on Delete wallet option")

	// Press enter to navigate
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Verify navigation
	currentRoute := s.router.GetCurrentRoute()
	s.Equal("/evm/wallet/delete", currentRoute.Path, "Should navigate to delete page")
	s.Equal("1", s.router.GetQueryParam("id"), "Should pass wallet ID")
}

// TestWalletLoadError tests displaying error when wallet service fails.
func (s *ActionsPageTestSuite) TestWalletLoadError() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service to return error
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(nil, fmt.Errorf("wallet not found"))

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Error:", "Should show error message")
	s.Contains(output, "failed to load wallet", "Should show wallet load error")
	s.NotContains(output, "View details", "Should not show action options")
}

// TestBalanceDisplay tests balance formatting and display.
func (s *ActionsPageTestSuite) TestBalanceDisplay() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service with specific balance (1.5 ETH)
	testWallet := s.createTestWallet("Balance Test", "1500000000000000000")
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Balance: 1.5000 ETH", "Should show formatted balance")
	s.Contains(output, "Address: 0x1111111111111111111111111111111111111111", "Should show address")
}

// TestBalanceUnavailable tests display when balance is nil.
func (s *ActionsPageTestSuite) TestBalanceUnavailable() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Mock GetCurrentConfig
	validConfig := models.EVMConfig{
		ID: 1,
		Endpoint: &models.EVMEndpoint{
			Url: "http://localhost:8545",
		},
		SelectedWalletID: uintPtr(1),
	}
	mockStorage.EXPECT().
		GetCurrentConfig().
		Return(validConfig, nil)

	// Mock wallet service with nil balance
	testWallet := &walletsvc.WalletWithBalance{
		Wallet: models.EVMWallet{
			ID:      1,
			Alias:   "Offline Wallet",
			Address: "0x1111111111111111111111111111111111111111",
		},
		Balance: nil, // Nil balance
	}
	mockWalletSvc.EXPECT().
		GetWalletWithBalance(uint(1), "http://localhost:8545").
		Return(testWallet, nil)

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add wallet ID to router query params
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "1"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Offline Wallet", "Should show wallet name")
	s.Contains(output, "unavailable ⚠", "Should show unavailable for nil balance")
}

// TestMissingWalletID tests error when wallet ID is not in query params.
func (s *ActionsPageTestSuite) TestMissingWalletID() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add route without wallet ID
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", nil) // No query params
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Error:", "Should show error message")
	s.Contains(output, "wallet ID not provided", "Should show missing ID error")
}

// TestInvalidWalletID tests error when wallet ID is not a valid number.
func (s *ActionsPageTestSuite) TestInvalidWalletID() {
	mockStorage := s.setupMockStorage()
	mockWalletSvc := s.setupMockWalletService()

	// Set up shared memory
	err := s.sharedMemory.Set(config.StorageClientKey, mockStorage)
	s.NoError(err, "Should set storage client in shared memory")

	// Add route with invalid wallet ID
	s.router.AddRoute(view.Route{
		Path: "/evm/wallet/actions",
		Component: func(router view.Router, sharedMemory storage.SharedMemory) view.View {
			return NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
		},
	})
	err = s.router.NavigateTo("/evm/wallet/actions", map[string]string{"id": "invalid"})
	s.NoError(err)

	model := NewPageWithService(s.router, s.sharedMemory, mockWalletSvc)
	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for wallet loading
	time.Sleep(300 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Error:", "Should show error message")
	s.Contains(output, "invalid wallet ID", "Should show invalid ID error")
}

// mockComponent is a simple mock component for router testing.
type mockComponent struct{}

func (m *mockComponent) Init() tea.Cmd                           { return nil }
func (m *mockComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *mockComponent) View() string                            { return "mock" }
func (m *mockComponent) Help() (string, view.HelpDisplayOption) {
	return "", view.HelpDisplayOptionAppend
}

// uintPtr is a helper function to create a pointer to a uint.
func uintPtr(u uint) *uint {
	return &u
}
