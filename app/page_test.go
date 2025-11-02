package app

import (
	"io"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
	"github.com/stretchr/testify/suite"
)

// PagePasswordTestSuite tests password unlock functionality using teatest.
type PagePasswordTestSuite struct {
	suite.Suite
	testStoragePath string
	sharedMemory    storage.SharedMemory
	router          view.Router
}

func TestPagePasswordTestSuite(t *testing.T) {
	suite.Run(t, new(PagePasswordTestSuite))
}

func (s *PagePasswordTestSuite) SetupTest() {
	// Create a temporary directory for test storage
	tmpDir, err := os.MkdirTemp("", "smart-contract-cli-test-*")
	s.NoError(err, "Should create temp directory")
	s.testStoragePath = tmpDir

	// Override the storage path for tests
	err = os.Setenv("HOME", tmpDir)
	s.NoError(err, "Should set HOME environment variable")

	// Create shared memory and router for each test
	s.sharedMemory = storage.NewSharedMemory()
	s.router = view.NewRouter()
}

func (s *PagePasswordTestSuite) TearDownTest() {
	// Clean up test storage
	if s.testStoragePath != "" {
		err := os.RemoveAll(s.testStoragePath)
		s.NoError(err, "Should clean up test storage directory")
	}
}

func (s *PagePasswordTestSuite) getOutput(tm *teatest.TestModel) string {
	output, err := io.ReadAll(tm.Output())
	s.NoError(err, "Should be able to read output")
	return string(output)
}

// TestInitialStateNewStorage tests that a new storage creation prompt is shown.
func (s *PagePasswordTestSuite) TestInitialStateNewStorage() {
	model := NewPage(s.router, s.sharedMemory)
	pageModel := model.(Model)

	// Verify initial state
	s.False(pageModel.isUnlocked, "Should not be unlocked initially")
	s.True(pageModel.isCreatingNew, "Should be in creating new storage mode")

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Create a password", "Should show create password prompt")
	s.Contains(output, "Password:", "Should show password input field")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestSuccessfulPasswordCreation tests creating a new storage with a password.
func (s *PagePasswordTestSuite) TestSuccessfulPasswordCreation() {
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Type password
	password := "testpass123"
	for _, char := range password {
		testModel.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit password with Enter
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for unlock to process
	time.Sleep(200 * time.Millisecond)

	// Verify password stored in shared memory
	storedPassword, err := s.sharedMemory.Get("secure_storage_password")
	s.NoError(err, "Should retrieve password from shared memory")
	s.Equal(password, storedPassword, "Password should match")

	// Verify main menu is shown
	output := s.getOutput(testModel)
	s.Contains(output, "Select a blockchain", "Should show main menu after unlock")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestExistingStorageUnlock tests unlocking existing storage with correct password.
func (s *PagePasswordTestSuite) TestExistingStorageUnlock() {
	// Pre-create storage with a known password
	password := "mypassword"
	secureStorage, err := storage.NewSecureStorageWithEncryption("smart-contract-cli-key", "")
	s.NoError(err, "Should create secure storage")

	err = secureStorage.Create(password)
	s.NoError(err, "Should create storage file")

	// Now create the model - it should detect existing storage
	model := NewPage(s.router, s.sharedMemory)
	pageModel := model.(Model)

	s.False(pageModel.isUnlocked, "Should not be unlocked initially")
	s.False(pageModel.isCreatingNew, "Should not be in create mode for existing storage")

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	output := s.getOutput(testModel)
	s.Contains(output, "Enter password to unlock", "Should show unlock prompt")

	// Type the correct password
	for _, char := range password {
		testModel.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit password
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for unlock
	time.Sleep(200 * time.Millisecond)

	// Verify unlocked
	output = s.getOutput(testModel)
	s.Contains(output, "Select a blockchain", "Should show main menu after successful unlock")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestInvalidPasswordError tests that wrong password shows error.
func (s *PagePasswordTestSuite) TestInvalidPasswordError() {
	// Pre-create storage with a known password
	correctPassword := "correctpass"
	secureStorage, err := storage.NewSecureStorageWithEncryption("smart-contract-cli-key", "")
	s.NoError(err, "Should create secure storage")

	err = secureStorage.Create(correctPassword)
	s.NoError(err, "Should create storage file")

	// Create model
	model := NewPage(s.router, s.sharedMemory)

	testModel2 := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Type wrong password
	wrongPassword := "wrongpass"
	for _, char := range wrongPassword {
		testModel2.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit wrong password
	testModel2.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for error processing
	time.Sleep(200 * time.Millisecond)

	// Verify error message shown
	output := s.getOutput(testModel2)
	s.Contains(output, "Failed to unlock", "Should show unlock failure message")
	s.NotContains(output, "Select a blockchain", "Should not show main menu")

	// Verify password not stored in shared memory
	storedPassword, err := s.sharedMemory.Get("secure_storage_password")
	s.NoError(err, "Get should not return error")
	s.Nil(storedPassword, "Should not have password in shared memory after failed unlock")

	// Quit
	testModel2.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel2.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestEmptyPasswordValidation tests that empty password is rejected.
func (s *PagePasswordTestSuite) TestEmptyPasswordValidation() {
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Submit without typing anything
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for validation
	time.Sleep(200 * time.Millisecond)

	// Verify error message
	output := s.getOutput(testModel)
	s.Contains(output, "Password cannot be empty", "Should show empty password error")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestSharedMemoryIntegration tests that pre-existing password in shared memory skips unlock.
func (s *PagePasswordTestSuite) TestSharedMemoryIntegration() {
	// Pre-create storage and set password in shared memory
	password := "presetpass"
	secureStorage, err := storage.NewSecureStorageWithEncryption("smart-contract-cli-key", "")
	s.NoError(err, "Should create secure storage")

	err = secureStorage.Create(password)
	s.NoError(err, "Should create storage file")

	// Store password in shared memory before creating model
	err = s.sharedMemory.Set("secure_storage_password", password)
	s.NoError(err, "Should store password in shared memory")

	// Create model - should automatically unlock
	model := NewPage(s.router, s.sharedMemory)
	pageModel := model.(Model)

	s.True(pageModel.isUnlocked, "Should be unlocked immediately with password in shared memory")

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Verify main menu shown immediately
	output := s.getOutput(testModel)
	s.Contains(output, "Select a blockchain", "Should show main menu immediately")
	s.NotContains(output, "Enter password", "Should not show password prompt")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestQuitDuringPasswordEntry tests that Ctrl+C works during password entry.
func (s *PagePasswordTestSuite) TestQuitDuringPasswordEntry() {
	model := NewPage(s.router, s.sharedMemory)

	testModel3 := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Type some password
	testModel3.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'t', 'e', 's', 't'},
	})

	// Quit with Ctrl+C
	testModel3.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel3.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))

	// If we get here without hanging, the test passes
	s.True(true, "Should quit cleanly during password entry")
}

// TestLoadStorageClientFromSharedMemory tests that a previously configured storage client
// can be loaded from shared memory when unlocking with a password.
func (s *PagePasswordTestSuite) TestLoadStorageClientFromSharedMemory() {
	// Pre-create storage with a known password (using password as encryption key)
	password := "testpassword"
	secureStorage, err := storage.NewSecureStorageWithEncryption(password, "")
	s.NoError(err, "Should create secure storage")

	err = secureStorage.Create(password)
	s.NoError(err, "Should create storage file")

	// Unlock the secure storage so we can store SQLite configuration
	err = secureStorage.TestPassword(password)
	s.NoError(err, "Should unlock secure storage")

	// Set up a SQLite storage client configuration in secure storage
	sqlitePath := s.testStoragePath + "/test.db"
	err = secureStorage.Set(config.SecureStorageClientTypeKey, string(types.StorageClientSQLite))
	s.NoError(err, "Should set storage client type")

	err = secureStorage.Set(config.SecureStorageKeySqlitePathKey, sqlitePath)
	s.NoError(err, "Should set sqlite path")

	// Create the actual SQLite database file so it can be loaded
	// This simulates a previously configured storage client
	_ = os.MkdirAll(s.testStoragePath, 0750)
	// Ignore error if directory already exists

	// Create model and unlock with password
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Type the correct password
	for _, char := range password {
		testModel.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit password
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for unlock and storage client initialization
	time.Sleep(300 * time.Millisecond)

	// Verify storage client was loaded into shared memory
	storageClient, err := s.sharedMemory.Get(config.StorageClientKey)
	s.NoError(err, "Should retrieve storage client from shared memory")
	s.NotNil(storageClient, "Storage client should be loaded into shared memory")

	// Verify the storage client is of the correct type (sql.Storage interface)
	_, ok := storageClient.(sql.Storage)
	s.True(ok, "Storage client should implement sql.Storage interface")

	// Verify main menu is shown after successful unlock
	output := s.getOutput(testModel)
	s.Contains(output, "Select a blockchain", "Should show main menu after successful unlock")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestLoadStorageClientWhenNotConfigured tests that when no storage client is configured,
// the page still unlocks successfully but shared memory doesn't contain a storage client.
func (s *PagePasswordTestSuite) TestLoadStorageClientWhenNotConfigured() {
	// Pre-create storage with a known password but DON'T configure storage client (using password as encryption key)
	password := "testpassword"
	secureStorage, err := storage.NewSecureStorageWithEncryption(password, "")
	s.NoError(err, "Should create secure storage")

	err = secureStorage.Create(password)
	s.NoError(err, "Should create storage file")

	// Don't set any storage client configuration - simulating first-time use

	// Create model and unlock with password
	model := NewPage(s.router, s.sharedMemory)

	testModel := teatest.NewTestModel(
		s.T(),
		model,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for initial render
	time.Sleep(100 * time.Millisecond)

	// Type the correct password
	for _, char := range password {
		testModel.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		time.Sleep(10 * time.Millisecond)
	}

	// Submit password
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for unlock
	time.Sleep(300 * time.Millisecond)

	// Verify storage client is nil in shared memory (not configured)
	storageClient, err := s.sharedMemory.Get(config.StorageClientKey)
	s.NoError(err, "Should not error when getting storage client")
	s.Nil(storageClient, "Storage client should be nil when not configured")

	// Verify main menu is still shown (unlock succeeded despite no storage client)
	output := s.getOutput(testModel)
	s.Contains(output, "Select a blockchain", "Should show main menu even without storage client")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}
