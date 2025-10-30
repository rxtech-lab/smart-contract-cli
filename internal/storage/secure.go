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
	"sync"
)

// SecureStorage is an interface for encrypted key-value storage.
type SecureStorage interface {
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
	data          map[string]string // stores encrypted values
	filePath      string
	mu            sync.RWMutex
}

// encryptedData represents the structure stored in the file.
type encryptedData struct {
	Data map[string]string `json:"data"`
}

// NewSecureStorageWithEncryption creates a new encrypted storage instance.
// encryptionKey: the key used for encryption (will be hashed to 32 bytes for AES-256).
// filePath: optional file path for persistence (empty string for in-memory only).
func NewSecureStorageWithEncryption(encryptionKey string, filePath string) (SecureStorage, error) {
	// Derive a 32-byte key from the provided encryption key using SHA-256
	hash := sha256.Sum256([]byte(encryptionKey))

	storage := &SecureStorageWithEncryption{
		encryptionKey: hash[:],
		data:          make(map[string]string),
		filePath:      filePath,
	}

	// Load existing data if file path is provided
	if filePath != "" {
		if err := storage.load(); err != nil {
			// If file doesn't exist, that's okay - we'll create it on first save
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load storage: %w", err)
			}
		}
	}

	return storage, nil
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

	// Marshal the data
	ed := encryptedData{Data: s.data}
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

	var ed encryptedData
	if err := json.Unmarshal(data, &ed); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	s.mu.Lock()
	s.data = ed.Data
	s.mu.Unlock()

	return nil
}
