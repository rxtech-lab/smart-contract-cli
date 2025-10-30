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
}

func (s *SecureStorageTestSuite) TearDownTest() {
	if s.storage != nil {
		s.storage.Close()
	}

	// Clean up temporary directory
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func TestSecureStorageTestSuite(t *testing.T) {
	suite.Run(t, new(SecureStorageTestSuite))
}

// Test basic Set and Get operations
func (s *SecureStorageTestSuite) TestSetAndGet() {
	err := s.storage.Set("key1", "value1")
	s.Require().NoError(err)

	value, err := s.storage.Get("key1")
	s.Require().NoError(err)
	s.Equal("value1", value)
}

// Test setting multiple values
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

// Test getting non-existent key
func (s *SecureStorageTestSuite) TestGetNonExistentKey() {
	value, err := s.storage.Get("non-existent-key")
	s.Error(err)
	s.Empty(value)
	s.Contains(err.Error(), "key not found")
}

// Test Delete operation
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

// Test List operation
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

// Test Clear operation
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

// Test persistence to file
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

	// Create new storage instance with same file
	newStorage, err := NewSecureStorageWithEncryption("test-encryption-key", s.tempFile)
	s.Require().NoError(err)
	defer newStorage.Close()

	// Verify data was loaded
	for key, expectedValue := range testData {
		value, err := newStorage.Get(key)
		s.Require().NoError(err)
		s.Equal(expectedValue, value)
	}
}

// Test encryption - verify data is actually encrypted on disk
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

// Test wrong encryption key
func (s *SecureStorageTestSuite) TestWrongEncryptionKey() {
	// Set data with first key
	err := s.storage.Set("secret", "value123")
	s.Require().NoError(err)
	s.storage.Close()

	// Try to load with different key
	wrongStorage, err := NewSecureStorageWithEncryption("wrong-key", s.tempFile)
	s.Require().NoError(err)
	defer wrongStorage.Close()

	// Should fail to decrypt
	_, err = wrongStorage.Get("secret")
	s.Error(err)
	s.Contains(err.Error(), "failed to decrypt")
}

// Test in-memory storage (no file path)
func (s *SecureStorageTestSuite) TestInMemoryStorage() {
	inMemStorage, err := NewSecureStorageWithEncryption("memory-key", "")
	s.Require().NoError(err)
	defer inMemStorage.Close()

	// Set and get data
	err = inMemStorage.Set("key1", "value1")
	s.Require().NoError(err)

	value, err := inMemStorage.Get("key1")
	s.Require().NoError(err)
	s.Equal("value1", value)

	// Close should not error even without file path
	err = inMemStorage.Close()
	s.NoError(err)
}

// Test concurrent operations
func (s *SecureStorageTestSuite) TestConcurrentOperations() {
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			key := "concurrent-key-" + string(rune('0'+index))
			value := "concurrent-value-" + string(rune('0'+index))
			err := s.storage.Set(key, value)
			s.NoError(err)
			done <- true
		}(i)
	}

	// Wait for all writes to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			key := "concurrent-key-" + string(rune('0'+index))
			expectedValue := "concurrent-value-" + string(rune('0'+index))
			value, err := s.storage.Get(key)
			s.NoError(err)
			s.Equal(expectedValue, value)
			done <- true
		}(i)
	}

	// Wait for all reads to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Test updating existing value
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

// Test special characters and unicode
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

// Test empty values
func (s *SecureStorageTestSuite) TestEmptyValues() {
	err := s.storage.Set("empty-key", "")
	s.Require().NoError(err)

	value, err := s.storage.Get("empty-key")
	s.Require().NoError(err)
	s.Equal("", value)
}

// Test large values
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

// Test file permissions
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

// Test directory creation
func (s *SecureStorageTestSuite) TestDirectoryCreation() {
	nestedPath := filepath.Join(s.tempDir, "nested", "path", "storage.json")

	storage, err := NewSecureStorageWithEncryption("test-key", nestedPath)
	s.Require().NoError(err)
	defer storage.Close()

	err = storage.Set("key", "value")
	s.Require().NoError(err)

	// Verify directory was created
	_, err = os.Stat(filepath.Dir(nestedPath))
	s.NoError(err, "Nested directory should be created")

	// Verify file was created
	_, err = os.Stat(nestedPath)
	s.NoError(err, "Storage file should be created in nested directory")
}
