package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SecureStorageTestSuite struct {
	suite.Suite
	storage  SecureStorage
	tempDir  string
	tempFile string
}

func (s *SecureStorageTestSuite) SetupTest() {
	// Create temporary directory for test files
	var err error
	s.tempDir, err = os.MkdirTemp("", "secure-storage-test-*")
	s.Require().NoError(err)

	s.tempFile = filepath.Join(s.tempDir, "test-storage.json")

	// Create storage with encryption
	s.storage, err = NewSecureStorageWithEncryption("test-encryption-key", s.tempFile)
	s.Require().NoError(err)

	// Create the storage with a password
	err = s.storage.Create("test-password")
	s.Require().NoError(err)
}

func (s *SecureStorageTestSuite) TearDownTest() {
	if s.storage != nil {
		err := s.storage.Close()
		s.NoError(err, "Should close storage")
	}

	// Clean up temporary directory
	if s.tempDir != "" {
		err := os.RemoveAll(s.tempDir)
		s.NoError(err, "Should clean up temp directory")
	}
}

func TestSecureStorageTestSuite(t *testing.T) {
	suite.Run(t, new(SecureStorageTestSuite))
}

// Test basic Set and Get operations.
func (s *SecureStorageTestSuite) TestSetAndGet() {
	err := s.storage.Set("key1", "value1")
	s.Require().NoError(err)

	value, err := s.storage.Get("key1")
	s.Require().NoError(err)
	s.Equal("value1", value)
}

// Test setting multiple values.
func (s *SecureStorageTestSuite) TestSetMultipleValues() {
	testData := map[string]string{
		"username": "john_doe",
		"password": "super_secret_password",
		"api_key":  "sk-1234567890abcdef",
		"token":    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
	}

	for key, value := range testData {
		err := s.storage.Set(key, value)
		s.Require().NoError(err)
	}

	for key, expectedValue := range testData {
		value, err := s.storage.Get(key)
		s.Require().NoError(err)
		s.Equal(expectedValue, value, "Value mismatch for key: %s", key)
	}
}

// Test getting non-existent key.
func (s *SecureStorageTestSuite) TestGetNonExistentKey() {
	value, err := s.storage.Get("non-existent-key")
	s.Error(err)
	s.Empty(value)
	s.Contains(err.Error(), "key not found")
}

// Test Delete operation.
func (s *SecureStorageTestSuite) TestDelete() {
	// Set a value
	err := s.storage.Set("key-to-delete", "value-to-delete")
	s.Require().NoError(err)

	// Verify it exists
	value, err := s.storage.Get("key-to-delete")
	s.Require().NoError(err)
	s.Equal("value-to-delete", value)

	// Delete it
	err = s.storage.Delete("key-to-delete")
	s.Require().NoError(err)

	// Verify it's gone
	_, err = s.storage.Get("key-to-delete")
	s.Error(err)
	s.Contains(err.Error(), "key not found")
}

// Test List operation.
func (s *SecureStorageTestSuite) TestList() {
	// Set multiple values
	testKeys := []string{"key1", "key2", "key3"}
	for _, key := range testKeys {
		err := s.storage.Set(key, "value-"+key)
		s.Require().NoError(err)
	}

	// List keys
	keys, err := s.storage.List()
	s.Require().NoError(err)
	s.Len(keys, len(testKeys))

	// Verify all keys are present
	for _, expectedKey := range testKeys {
		s.Contains(keys, expectedKey)
	}
}

// Test Clear operation.
func (s *SecureStorageTestSuite) TestClear() {
	// Set multiple values
	for i := 0; i < 5; i++ {
		key := "key" + string(rune('0'+i))
		err := s.storage.Set(key, "value"+string(rune('0'+i)))
		s.Require().NoError(err)
	}

	// Verify data exists
	keys, err := s.storage.List()
	s.Require().NoError(err)
	s.Len(keys, 5)

	// Clear all data
	err = s.storage.Clear()
	s.Require().NoError(err)

	// Verify data is cleared
	keys, err = s.storage.List()
	s.Require().NoError(err)
	s.Len(keys, 0)
}

// Test persistence to file.
func (s *SecureStorageTestSuite) TestPersistence() {
	// Set some data
	testData := map[string]string{
		"persistent_key1": "persistent_value1",
		"persistent_key2": "persistent_value2",
	}

	for key, value := range testData {
		err := s.storage.Set(key, value)
		s.Require().NoError(err)
	}

	// Close the storage (saves to file)
	err := s.storage.Close()
	s.Require().NoError(err)

	// Verify file was created
	_, err = os.Stat(s.tempFile)
	s.Require().NoError(err, "Storage file should exist")

	// Create new storage instance with same file (it will auto-load existing data)
	newStorage, err := NewSecureStorageWithEncryption("test-encryption-key", s.tempFile)
	s.Require().NoError(err)
	defer func() {
		if closeErr := newStorage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Verify data was loaded
	for key, expectedValue := range testData {
		value, err := newStorage.Get(key)
		s.Require().NoError(err)
		s.Equal(expectedValue, value)
	}
}

// Test encryption - verify data is actually encrypted on disk.
func (s *SecureStorageTestSuite) TestEncryptionOnDisk() {
	sensitiveData := "super_secret_password_12345"
	err := s.storage.Set("password", sensitiveData)
	s.Require().NoError(err)

	// Close to ensure data is written to disk
	err = s.storage.Close()
	s.Require().NoError(err)

	// Read the raw file content
	fileContent, err := os.ReadFile(s.tempFile)
	s.Require().NoError(err)

	// Verify the sensitive data is NOT in plaintext in the file
	s.NotContains(string(fileContent), sensitiveData, "Sensitive data should be encrypted")

	// Verify file contains encrypted data structure
	s.Contains(string(fileContent), "data", "File should contain JSON structure")
}

// Test wrong encryption key.
func (s *SecureStorageTestSuite) TestWrongEncryptionKey() {
	// Set data with first key
	err := s.storage.Set("secret", "value123")
	s.Require().NoError(err)
	err = s.storage.Close()
	s.Require().NoError(err)

	// Try to load with different encryption key (but it will still load the file)
	wrongStorage, err := NewSecureStorageWithEncryption("wrong-key", s.tempFile)
	s.Require().NoError(err)
	defer func() {
		if closeErr := wrongStorage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Should fail to decrypt because encryption key is different
	_, err = wrongStorage.Get("secret")
	s.Error(err)
	s.Contains(err.Error(), "failed to decrypt")
}

// Test in-memory storage (uses default path).
func (s *SecureStorageTestSuite) TestDefaultPath() {
	// When no path is provided, it should use the default path
	// For testing, we'll provide an explicit path instead
	testPath := filepath.Join(s.tempDir, "default-path-test.json")
	storage, err := NewSecureStorageWithEncryption("memory-key", testPath)
	s.Require().NoError(err)
	defer func() {
		if closeErr := storage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Create the storage
	err = storage.Create("test-password")
	s.Require().NoError(err)

	// Set and get data
	err = storage.Set("key1", "value1")
	s.Require().NoError(err)

	value, err := storage.Get("key1")
	s.Require().NoError(err)
	s.Equal("value1", value)

	// Close should not error
	err = storage.Close()
	s.NoError(err)
}

// Test concurrent operations.
func (s *SecureStorageTestSuite) TestConcurrentOperations() {
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	// Concurrent writes
	for index := 0; index < numGoroutines; index++ {
		go func(index int) {
			key := "concurrent-key-" + string(rune('0'+index))
			value := "concurrent-value-" + string(rune('0'+index))
			err := s.storage.Set(key, value)
			s.NoError(err)
			done <- true
		}(index)
	}

	// Wait for all writes to complete
	for index := 0; index < numGoroutines; index++ {
		<-done
	}

	// Concurrent reads
	for index := 0; index < numGoroutines; index++ {
		go func(index int) {
			key := "concurrent-key-" + string(rune('0'+index))
			expectedValue := "concurrent-value-" + string(rune('0'+index))
			value, err := s.storage.Get(key)
			s.NoError(err)
			s.Equal(expectedValue, value)
			done <- true
		}(index)
	}

	// Wait for all reads to complete
	for index := 0; index < numGoroutines; index++ {
		<-done
	}
}

// Test updating existing value.
func (s *SecureStorageTestSuite) TestUpdateValue() {
	key := "update-key"

	// Set initial value
	err := s.storage.Set(key, "initial-value")
	s.Require().NoError(err)

	// Verify initial value
	value, err := s.storage.Get(key)
	s.Require().NoError(err)
	s.Equal("initial-value", value)

	// Update value
	err = s.storage.Set(key, "updated-value")
	s.Require().NoError(err)

	// Verify updated value
	value, err = s.storage.Get(key)
	s.Require().NoError(err)
	s.Equal("updated-value", value)
}

// Test special characters and unicode.
func (s *SecureStorageTestSuite) TestSpecialCharacters() {
	testCases := map[string]string{
		"unicode":       "Hello 世界 🌍",
		"special-chars": "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		"newlines":      "line1\nline2\nline3",
		"tabs":          "col1\tcol2\tcol3",
		"json":          `{"key": "value", "nested": {"array": [1, 2, 3]}}`,
	}

	for key, value := range testCases {
		err := s.storage.Set(key, value)
		s.Require().NoError(err)

		retrieved, err := s.storage.Get(key)
		s.Require().NoError(err)
		s.Equal(value, retrieved, "Special character handling failed for key: %s", key)
	}
}

// Test empty values.
func (s *SecureStorageTestSuite) TestEmptyValues() {
	err := s.storage.Set("empty-key", "")
	s.Require().NoError(err)

	value, err := s.storage.Get("empty-key")
	s.Require().NoError(err)
	s.Equal("", value)
}

// Test large values.
func (s *SecureStorageTestSuite) TestLargeValues() {
	// Create a large value (1 MB)
	largeValue := make([]byte, 1024*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	err := s.storage.Set("large-key", string(largeValue))
	s.Require().NoError(err)

	retrieved, err := s.storage.Get("large-key")
	s.Require().NoError(err)
	s.Equal(string(largeValue), retrieved)
}

// Test file permissions.
func (s *SecureStorageTestSuite) TestFilePermissions() {
	err := s.storage.Set("key", "value")
	s.Require().NoError(err)

	// Close to ensure file is written
	err = s.storage.Close()
	s.Require().NoError(err)

	// Check file permissions
	fileInfo, err := os.Stat(s.tempFile)
	s.Require().NoError(err)

	// File should be readable and writable by owner only (0600)
	expectedPerms := os.FileMode(0600)
	s.Equal(expectedPerms, fileInfo.Mode().Perm())
}

// Test directory creation.
func (s *SecureStorageTestSuite) TestDirectoryCreation() {
	nestedPath := filepath.Join(s.tempDir, "nested", "path", "storage.json")

	storage, err := NewSecureStorageWithEncryption("test-key", nestedPath)
	s.Require().NoError(err)
	defer func() {
		if closeErr := storage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	err = storage.Create("test-password")
	s.Require().NoError(err)

	err = storage.Set("key", "value")
	s.Require().NoError(err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(nestedPath))
	s.NoError(err, "Nested directory should be created")

	// Verify file was created
	_, err = os.Stat(nestedPath)
	s.NoError(err, "Storage file should be created in nested directory")

	// Close storage
	err = storage.Close()
	s.NoError(err, "Should close storage")
}

// Test Exists method.
func (s *SecureStorageTestSuite) TestExists() {
	// Test file should exist after Create() in SetupTest
	s.True(s.storage.Exists(), "Storage should exist after creation")

	// Test non-existent file
	newPath := filepath.Join(s.tempDir, "non-existent.json")
	newStorage, err := NewSecureStorageWithEncryption("test-key", newPath)
	s.Require().NoError(err)
	defer func() {
		if closeErr := newStorage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	s.False(newStorage.Exists(), "Storage should not exist before creation")
}

// Test Create method.
func (s *SecureStorageTestSuite) TestCreate() {
	newPath := filepath.Join(s.tempDir, "new-storage.json")
	storage, err := NewSecureStorageWithEncryption("test-key", newPath)
	s.Require().NoError(err)
	defer func() {
		if closeErr := storage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Create should succeed
	err = storage.Create("my-password")
	s.NoError(err)

	// File should exist
	s.True(storage.Exists())

	// Should be able to set and get values
	err = storage.Set("key", "value")
	s.NoError(err)

	value, err := storage.Get("key")
	s.NoError(err)
	s.Equal("value", value)
}

// Test Create with empty password.
func (s *SecureStorageTestSuite) TestCreateEmptyPassword() {
	newPath := filepath.Join(s.tempDir, "empty-password-storage.json")
	storage, err := NewSecureStorageWithEncryption("test-key", newPath)
	s.Require().NoError(err)
	defer func() {
		if closeErr := storage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Create should fail with empty password
	err = storage.Create("")
	s.Error(err)
	s.Contains(err.Error(), "password cannot be empty")
}

// Test Create when storage already exists.
func (s *SecureStorageTestSuite) TestCreateAlreadyExists() {
	// Storage was already created in SetupTest
	err := s.storage.Create("another-password")
	s.Error(err)
	s.Contains(err.Error(), "storage already exists")
}

// Test Unlock with correct password.
func (s *SecureStorageTestSuite) TestUnlockSuccess() {
	// Close and reload storage
	err := s.storage.Close()
	s.Require().NoError(err)

	// Reload storage
	newStorage, err := NewSecureStorageWithEncryption("test-encryption-key", s.tempFile)
	s.Require().NoError(err)
	defer func() {
		if closeErr := newStorage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Unlock with correct password should succeed
	err = newStorage.TestPassword("test-password")
	s.NoError(err)
}

// Test Unlock with wrong password.
func (s *SecureStorageTestSuite) TestUnlockWrongPassword() {
	// Close and reload storage
	err := s.storage.Close()
	s.Require().NoError(err)

	// Reload storage
	newStorage, err := NewSecureStorageWithEncryption("test-encryption-key", s.tempFile)
	s.Require().NoError(err)
	defer func() {
		if closeErr := newStorage.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Unlock with wrong password should fail
	err = newStorage.TestPassword("wrong-password")
	s.Error(err)
	s.Contains(err.Error(), "incorrect password")
}

// Test full Create-Unlock workflow.
func (s *SecureStorageTestSuite) TestCreateUnlockWorkflow() {
	newPath := filepath.Join(s.tempDir, "workflow-storage.json")
	password := "secure-password-123"

	// Step 1: Create new storage
	storage1, err := NewSecureStorageWithEncryption("encryption-key", newPath)
	s.Require().NoError(err)

	err = storage1.Create(password)
	s.Require().NoError(err)

	// Step 2: Add some data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range testData {
		err = storage1.Set(key, value)
		s.Require().NoError(err)
	}

	// Step 3: Close storage
	err = storage1.Close()
	s.Require().NoError(err)

	// Step 4: Load storage again
	storage2, err := NewSecureStorageWithEncryption("encryption-key", newPath)
	s.Require().NoError(err)
	defer func() {
		if closeErr := storage2.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Step 5: Verify password with Unlock
	err = storage2.TestPassword(password)
	s.NoError(err, "Should unlock with correct password")

	// Step 6: Verify wrong password fails
	err = storage2.TestPassword("wrong-password")
	s.Error(err, "Should fail with wrong password")

	// Step 7: Verify data is accessible (regardless of unlock)
	for key, expectedValue := range testData {
		value, err := storage2.Get(key)
		s.NoError(err)
		s.Equal(expectedValue, value)
	}
}

// Test that storage operations work without calling Unlock.
func (s *SecureStorageTestSuite) TestOperationsWithoutUnlock() {
	// Storage was created but Unlock was never called
	// Operations should still work

	err := s.storage.Set("test-key", "test-value")
	s.NoError(err, "Set should work without calling Unlock")

	value, err := s.storage.Get("test-key")
	s.NoError(err, "Get should work without calling Unlock")
	s.Equal("test-value", value)
}
