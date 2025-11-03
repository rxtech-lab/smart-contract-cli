package utils

import (
	"fmt"

	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
)

// GetSecureStorageFromSharedMemory gets the secure storage from the shared memory.
func GetSecureStorageFromSharedMemory(sharedMemory storage.SharedMemory) (storage.SecureStorage, string, error) {
	passwordRaw, err := sharedMemory.Get(config.SecureStoragePasswordKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get password from shared memory: %w", err)
	}
	password, ok := passwordRaw.(string)
	if !ok {
		return nil, "", fmt.Errorf("password in shared memory is not a string")
	}

	// Test password
	secureStorage, err := storage.NewSecureStorageWithEncryption(password, "")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create secure storage: %w", err)
	}

	if err := secureStorage.TestPassword(password); err != nil {
		return nil, "", fmt.Errorf("failed to test password: %w", err)
	}

	return secureStorage, password, nil
}

// GetStorageClientFromSharedMemory gets the storage client from the shared memory.
func GetStorageClientFromSharedMemory(sharedMemory storage.SharedMemory) (sql.Storage, error) {
	storageClient, err := sharedMemory.Get(config.StorageClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage client from shared memory: %w", err)
	}
	sqlStorage, isValidStorage := storageClient.(sql.Storage)
	if !isValidStorage {
		return nil, fmt.Errorf("invalid storage client type")
	}
	return sqlStorage, nil
}
