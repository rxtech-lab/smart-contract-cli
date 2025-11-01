package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rxtech-lab/smart-contract-cli/internal/config"
)

// SecureStorage is an interface for encrypted key-value storage.
type SecureStorage interface {
	// Exists checks if the storage exists.
	Exists() bool
	// Create creates the storage.
	Create(password string) error
	// Unlock unlocks the storage.
	Unlock(password string) error
	// Get retrieves and decrypts the value for a given key.
	Get(key string) (value string, err error)
	// Set encrypts and stores the value for a given key.
	Set(key string, value string) (err error)
	// Delete removes the value for a given key.
	Delete(key string) (err error)
	// List returns all keys (not decrypted values).
	List() (keys []string, err error)
	// Clear removes all stored data.
	Clear() (err error)
	// Close saves the data and cleans up resources.
	Close() error
}

// SecureStorageWithEncryption implements SecureStorage using AES-GCM encryption.
type SecureStorageWithEncryption struct {
	encryptionKey []byte
	passwordHash  string            // SHA-256 hash of the password
	data          map[string]string // Stores encrypted values
	filePath      string
	mu            sync.RWMutex
}

// encryptedData represents the structure stored in the file.
type encryptedData struct {
	PasswordHash string            `json:"password_hash"`
	Data         map[string]string `json:"data"`
}

// NewSecureStorageWithEncryption creates a new encrypted storage instance.
// EncryptionKey: the key used for encryption (will be hashed to 32 bytes for AES-256).
// FilePath: optional file path for persistence (empty string uses default path from config).
func NewSecureStorageWithEncryption(encryptionKey string, filePath string) (SecureStorage, error) {
	// Use default path if not provided
	if filePath == "" {
		filePath = expandPath(config.DefaultSecureStoragePath)
	} else {
		filePath = expandPath(filePath)
	}

	// Derive a 32-byte key from the provided encryption key using SHA-256
	hash := sha256.Sum256([]byte(encryptionKey))

	storage := &SecureStorageWithEncryption{
		encryptionKey: hash[:],
		data:          make(map[string]string),
		filePath:      filePath,
	}

	// Load existing data if file exists
	if _, err := os.Stat(filePath); err == nil {
		if err := storage.load(); err != nil {
			return nil, fmt.Errorf("failed to load storage: %w", err)
		}
	}

	return storage, nil
}

// expandPath expands the tilde (~) in file paths to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// Exists checks if the storage file exists.
func (s *SecureStorageWithEncryption) Exists() bool {
	if s.filePath == "" {
		return false
	}
	_, err := os.Stat(s.filePath)
	return err == nil
}

// Create creates a new storage with the given password.
func (s *SecureStorageWithEncryption) Create(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Check if storage already exists
	if s.Exists() {
		return fmt.Errorf("storage already exists at %s", s.filePath)
	}

	// Generate password hash
	hash := sha256.Sum256([]byte(password))
	s.passwordHash = fmt.Sprintf("%x", hash)

	// Initialize empty data
	s.mu.Lock()
	s.data = make(map[string]string)
	s.mu.Unlock()

	// Save to disk
	if s.filePath != "" {
		if err := s.save(); err != nil {
			return fmt.Errorf("failed to create storage: %w", err)
		}
	}

	return nil
}

// Unlock verifies the password against the stored password hash.
func (s *SecureStorageWithEncryption) Unlock(password string) error {
	// Load data if not already loaded
	if s.filePath != "" && s.passwordHash == "" {
		if err := s.load(); err != nil {
			return fmt.Errorf("failed to load storage: %w", err)
		}
	}

	// Hash the provided password
	hash := sha256.Sum256([]byte(password))
	providedHash := fmt.Sprintf("%x", hash)

	// Compare with stored hash
	if s.passwordHash != providedHash {
		return fmt.Errorf("incorrect password")
	}

	return nil
}

// encrypt encrypts plaintext using AES-GCM.
func (s *SecureStorageWithEncryption) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-GCM.
func (s *SecureStorageWithEncryption) decrypt(ciphertext string) (string, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Get retrieves and decrypts the value for a given key.
func (s *SecureStorageWithEncryption) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encryptedValue, exists := s.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return s.decrypt(encryptedValue)
}

// Set encrypts and stores the value for a given key.
func (s *SecureStorageWithEncryption) Set(key string, value string) error {
	encryptedValue, err := s.encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	s.mu.Lock()
	s.data[key] = encryptedValue
	s.mu.Unlock()

	// Persist to file if path is set
	if s.filePath != "" {
		if err := s.save(); err != nil {
			return fmt.Errorf("failed to save storage: %w", err)
		}
	}

	return nil
}

// Delete removes the value for a given key.
func (s *SecureStorageWithEncryption) Delete(key string) error {
	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()

	// Persist to file if path is set
	if s.filePath != "" {
		if err := s.save(); err != nil {
			return fmt.Errorf("failed to save storage: %w", err)
		}
	}

	return nil
}

// List returns all keys (not decrypted values).
func (s *SecureStorageWithEncryption) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}

	return keys, nil
}

// Clear removes all stored data.
func (s *SecureStorageWithEncryption) Clear() error {
	s.mu.Lock()
	s.data = make(map[string]string)
	s.mu.Unlock()

	// Persist to file if path is set
	if s.filePath != "" {
		if err := s.save(); err != nil {
			return fmt.Errorf("failed to save storage: %w", err)
		}
	}

	return nil
}

// Close saves the data and cleans up resources.
func (s *SecureStorageWithEncryption) Close() error {
	if s.filePath != "" {
		return s.save()
	}
	return nil
}

// save persists the encrypted data to disk.
func (s *SecureStorageWithEncryption) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal the data with password hash
	ed := encryptedData{
		PasswordHash: s.passwordHash,
		Data:         s.data,
	}
	jsonData, err := json.MarshalIndent(ed, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(s.filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// load reads the encrypted data from disk.
func (s *SecureStorageWithEncryption) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// Don't wrap this error so that os.IsNotExist() works correctly
		return err //nolint:wrapcheck // Need unwrapped error for os.IsNotExist() check
	}

	var encData encryptedData
	if err := json.Unmarshal(data, &encData); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	s.mu.Lock()
	s.passwordHash = encData.PasswordHash
	s.data = encData.Data
	s.mu.Unlock()

	return nil
}
