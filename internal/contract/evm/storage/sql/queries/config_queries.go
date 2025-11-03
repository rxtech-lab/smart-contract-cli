package queries

import (
	"errors"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	customerrors "github.com/rxtech-lab/smart-contract-cli/internal/errors"
	"gorm.io/gorm"
)

// ConfigQueries provides database operations for EVMConfig model.
type ConfigQueries struct {
	db *gorm.DB
}

// NewConfigQueries creates a new ConfigQueries instance.
func NewConfigQueries(db *gorm.DB) *ConfigQueries {
	return &ConfigQueries{db: db}
}

func (q *ConfigQueries) GetCurrent() (*models.EVMConfig, error) {
	var config models.EVMConfig
	if err := q.db.Preload("Endpoint").
		Preload("SelectedEVMContract").
		Preload("SelectedEVMAbi").
		Preload("SelectedWallet").
		First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "no config found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get current config")
	}
	return &config, nil
}

// GetByID retrieves a config by its ID with preloaded relationships.
func (q *ConfigQueries) GetByID(id uint) (*models.EVMConfig, error) {
	var config models.EVMConfig
	if err := q.db.Preload("Endpoint").
		Preload("SelectedEVMContract").
		Preload("SelectedEVMAbi").
		First(&config, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "config not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get config by ID")
	}
	return &config, nil
}

// Create creates a new empty config if one doesn't exist, or skips if it does.
func (q *ConfigQueries) Create() error {
	// Check if config already exists
	var count int64
	if err := q.db.Model(&models.EVMConfig{}).Count(&count).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check config existence")
	}

	// If config exists, skip creation
	if count > 0 {
		return nil
	}

	// Create empty config
	config := &models.EVMConfig{}
	if err := q.db.Create(config).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to create config")
	}
	return nil
}

// Update updates the current config with the provided config object.
func (q *ConfigQueries) Update(config *models.EVMConfig) error {
	// Get current config first
	currentConfig, err := q.GetCurrent()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "no config found to update")
		}
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get current config")
	}

	// Update current config with provided values
	updates := map[string]any{
		"endpoint_id":              config.EndpointId,
		"selected_evm_contract_id": config.SelectedEVMContractId,
		"selected_evm_abi_id":      config.SelectedEVMAbiId,
		"selected_wallet_id":       config.SelectedWalletID,
	}

	result := q.db.Model(&models.EVMConfig{}).Where("id = ?", currentConfig.ID).Updates(updates)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to update config")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "config not found")
	}
	return nil
}

// Delete deletes the current config.
func (q *ConfigQueries) Delete() error {
	// Get current config first
	currentConfig, err := q.GetCurrent()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "no config found to delete")
		}
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get current config")
	}

	result := q.db.Delete(&models.EVMConfig{}, currentConfig.ID)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to delete config")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "config not found")
	}
	return nil
}

// Exists checks if a config with the given ID exists.
func (q *ConfigQueries) Exists(id uint) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EVMConfig{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check config existence")
	}
	return count > 0, nil
}
