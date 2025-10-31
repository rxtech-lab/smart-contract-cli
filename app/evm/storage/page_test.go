package storage

import (
	"io"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
	"github.com/stretchr/testify/suite"
)

// StoragePageTestSuite tests storage client page functionality using teatest.
type StoragePageTestSuite struct {
	suite.Suite
	testStoragePath string
	sharedMemory    storage.SharedMemory
	secureStorage   storage.SecureStorage
	router          view.Router
	password        string
}

func TestStoragePageTestSuite(t *testing.T) {
	suite.Run(t, new(StoragePageTestSuite))
}

func (s *StoragePageTestSuite) SetupTest() {
	// Create a temporary directory for test storage
	tmpDir, err := os.MkdirTemp("", "storage-page-test-*")
	s.NoError(err, "Should create temp directory")
	s.testStoragePath = tmpDir

	// Override the storage path for tests
	err = os.Setenv("HOME", tmpDir)
	s.NoError(err, "Should set HOME environment variable")

	// Set up password and storage
	s.password = "testpassword123"
	s.sharedMemory = storage.NewSharedMemory()

	// Store password in shared memory with the correct key that NewPage expects
	err = s.sharedMemory.Set(config.SecureStoragePasswordKey, s.password)
	s.NoError(err, "Should store password in shared memory")

	// Create secure storage for pre-configuration tests
	s.secureStorage, err = storage.NewSecureStorageWithEncryption(s.password, "")
	s.NoError(err, "Should create secure storage")

	// Initialize the secure storage (create it if it doesn't exist)
	if !s.secureStorage.Exists() {
		err = s.secureStorage.Create(s.password)
		s.NoError(err, "Should create secure storage")
	}

	// Create router
	s.router = view.NewRouter()
}

func (s *StoragePageTestSuite) TearDownTest() {
	// Clean up test storage
	if s.testStoragePath != "" {
		err := os.RemoveAll(s.testStoragePath)
		s.NoError(err, "Should clean up test storage directory")
	}
}

func (s *StoragePageTestSuite) getOutput(tm *teatest.TestModel) string {
	output, err := io.ReadAll(tm.Output())
	s.NoError(err, "Should be able to read output")
	return string(output)
}

// TestInitialState tests the initial rendering of the storage page.
func (s *StoragePageTestSuite) TestInitialState() {
	// Verify password is in shared memory
	pwd, err := s.sharedMemory.Get(config.SecureStoragePasswordKey)
	s.NoError(err, "Should get password from shared memory")
	s.NotNil(pwd, "Password should not be nil")
	s.Equal(s.password, pwd, "Password should match")

	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Storage Client Configuration", "Should show page title")
	s.Contains(output, "SQLite", "Should show SQLite option")
	s.Contains(output, "Postgres", "Should show Postgres option")
	s.Contains(output, "Legend:", "Should show legend")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestNavigationUpDown tests keyboard navigation between storage options.
func (s *StoragePageTestSuite) TestNavigationUpDown() {
	model := NewPage(s.router, s.sharedMemory)
	pageModel := model.(Model)

	// Initially should be at index 0 (SQLite)
	s.Equal(0, pageModel.selectedIndex, "Should start at first option")

	// Simulate down arrow
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	pageModel = updatedModel.(Model)
	s.Equal(1, pageModel.selectedIndex, "Should move to second option")

	// Simulate up arrow
	updatedModel, _ = pageModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	pageModel = updatedModel.(Model)
	s.Equal(0, pageModel.selectedIndex, "Should move back to first option")

	// Try to go up past first option (should stay at 0)
	updatedModel, _ = pageModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	pageModel = updatedModel.(Model)
	s.Equal(0, pageModel.selectedIndex, "Should stay at first option")
}

// TestSelectSQLiteFirstTime tests selecting SQLite for the first time.
func (s *StoragePageTestSuite) TestSelectSQLiteFirstTime() {
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Press Enter to select SQLite (first option)
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Should show input mode for SQLite path
	output := s.getOutput(testModel)
	s.Contains(output, "Configure SQLite", "Should show SQLite configuration")
	s.Contains(output, "Enter the path", "Should prompt for path")

	// Type a path
	testPath := "/tmp/test.db"
	for _, char := range testPath {
		testModel.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit the path
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond)

	// Should return to normal view
	output = s.getOutput(testModel)
	s.Contains(output, "Storage Client Configuration", "Should return to main view")
	s.Contains(output, testPath, "Should display the configured path")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestSelectPostgresFirstTime tests selecting Postgres for the first time.
func (s *StoragePageTestSuite) TestSelectPostgresFirstTime() {
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Navigate to Postgres (second option)
	testModel.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(50 * time.Millisecond)

	// Press Enter to select Postgres
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Should show input mode for Postgres URL
	output := s.getOutput(testModel)
	s.Contains(output, "Configure Postgres", "Should show Postgres configuration")
	s.Contains(output, "PostgreSQL connection URL", "Should prompt for URL")

	// Type a URL
	testURL := "postgres://user:pass@localhost:5432/db"
	for _, char := range testURL {
		testModel.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit the URL
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond)

	// Should return to normal view with masked password
	output = s.getOutput(testModel)
	s.Contains(output, "Storage Client Configuration", "Should return to main view")
	s.Contains(output, "postgres://user:****@localhost:5432/db", "Should display masked URL")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestCancelInput tests canceling input with Escape key.
func (s *StoragePageTestSuite) TestCancelInput() {
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Select SQLite
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Should be in input mode
	output := s.getOutput(testModel)
	s.Contains(output, "Configure SQLite", "Should be in SQLite config mode")

	// Press Escape to cancel
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(100 * time.Millisecond)

	// Should return to normal view
	output = s.getOutput(testModel)
	s.Contains(output, "Storage Client Configuration", "Should return to main view")
	s.NotContains(output, "Configure SQLite", "Should exit config mode")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestEmptyPathValidation tests that empty path/URL is rejected.
func (s *StoragePageTestSuite) TestEmptyPathValidation() {
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Select SQLite
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Submit without entering anything
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Should show error
	output := s.getOutput(testModel)
	s.Contains(output, "Path/URL cannot be empty", "Should show validation error")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestActiveClientHighlighting tests that active client is highlighted.
func (s *StoragePageTestSuite) TestActiveClientHighlighting() {
	// Pre-configure SQLite as active client
	err := s.secureStorage.Set("storage_client_type", "sqlite")
	s.NoError(err, "Should set active client")
	err = s.secureStorage.Set("storage_client_sqlite_path", "/tmp/test.db")
	s.NoError(err, "Should set SQLite path")

	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Should show active client marker
	output := s.getOutput(testModel)
	s.Contains(output, "★", "Should show star marker for active client")
	s.Contains(output, "/tmp/test.db", "Should show configured path")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestConfirmationDialog tests the confirmation dialog for existing config.
func (s *StoragePageTestSuite) TestConfirmationDialog() {
	// Pre-configure SQLite
	err := s.secureStorage.Set("storage_client_type", "sqlite")
	s.NoError(err, "Should set active client")
	err = s.secureStorage.Set("storage_client_sqlite_path", "/tmp/existing.db")
	s.NoError(err, "Should set SQLite path")

	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Select SQLite (already configured)
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(100 * time.Millisecond)

	// Should show confirmation dialog
	output := s.getOutput(testModel)
	s.Contains(output, "SQLite Configuration", "Should show configuration dialog")
	s.Contains(output, "Use existing configuration", "Should show use existing option")
	s.Contains(output, "Change configuration", "Should show change option")
	s.Contains(output, "Remove configuration", "Should show remove option")
	s.Contains(output, "/tmp/existing.db", "Should show current path")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestPasswordMasking tests that Postgres password is properly masked.
func (s *StoragePageTestSuite) TestPasswordMasking() {
	testURL := "postgres://user:secretpass@localhost:5432/mydb"
	masked := maskPostgresURL(testURL)

	s.Contains(masked, "postgres://user:****@localhost:5432/mydb", "Should mask password")
	s.NotContains(masked, "secretpass", "Should not contain actual password")
}

// TestVimStyleNavigation tests vim-style navigation (j/k keys).
func (s *StoragePageTestSuite) TestVimStyleNavigation() {
	model := NewPage(s.router, s.sharedMemory)
	pageModel := model.(Model)

	// Initially at index 0
	s.Equal(0, pageModel.selectedIndex)

	// Simulate 'j' key (down)
	updatedModel, _ := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'j'},
	})
	pageModel = updatedModel.(Model)
	s.Equal(1, pageModel.selectedIndex, "Should move down with 'j'")

	// Simulate 'k' key (up)
	updatedModel, _ = pageModel.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'k'},
	})
	pageModel = updatedModel.(Model)
	s.Equal(0, pageModel.selectedIndex, "Should move up with 'k'")
}

// TestBackNavigation tests that Esc/q triggers back navigation.
func (s *StoragePageTestSuite) TestBackNavigation() {
	backCalled := false
	mockRouter := &mockRouter{
		backFunc: func() {
			backCalled = true
		},
	}

	// Store password for initialization
	err := s.sharedMemory.Set(config.SecureStoragePasswordKey, s.password)
	s.NoError(err, "Should store password in shared memory")

	model := NewPage(mockRouter, s.sharedMemory)

	// Press 'q' to go back
	model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'q'},
	})

	s.True(backCalled, "Should call router.Back() when 'q' is pressed")
}

// mockRouter is a mock implementation of view.Router for testing.
type mockRouter struct {
	backFunc func()
}

func (m *mockRouter) AddRoute(route view.Route)                                   {}
func (m *mockRouter) SetRoutes(routes []view.Route)                               {}
func (m *mockRouter) RemoveRoute(path string)                                     {}
func (m *mockRouter) GetRoutes() []view.Route                                     { return nil }
func (m *mockRouter) GetCurrentRoute() view.Route                                 { return view.Route{} }
func (m *mockRouter) NavigateTo(path string, queryParams map[string]string) error { return nil }
func (m *mockRouter) ReplaceRoute(path string) error {
	return nil
}
func (m *mockRouter) Back() {
	if m.backFunc != nil {
		m.backFunc()
	}
}
func (m *mockRouter) CanGoBack() bool                         { return true }
func (m *mockRouter) GetParam(key string) string              { return "" }
func (m *mockRouter) GetQueryParam(key string) string         { return "" }
func (m *mockRouter) GetPath() string                         { return "" }
func (m *mockRouter) Refresh()                                {}
func (m *mockRouter) Init() tea.Cmd                           { return nil }
func (m *mockRouter) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *mockRouter) View() string                            { return "" }
