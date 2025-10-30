package queries

import (
	"errors"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
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

// List retrieves a paginated list of configs with preloaded relationships.
func (q *ConfigQueries) List(page int64, pageSize int64) (*types.Pagination[models.EVMConfig], error) {
	if page < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageNumber, "page number must be greater than 0")
	}
	if pageSize < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageSize, "page size must be greater than 0")
	}

	var items []models.EVMConfig
	var totalItems int64

	// Count total items
	if err := q.db.Model(&models.EVMConfig{}).Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count configs")
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Retrieve paginated items with preloaded relationships
	offset := (page - 1) * pageSize
	if err := q.db.Preload("Endpoint").
		Preload("SelectedEVMContract").
		Preload("SelectedEVMAbi").
		Offset(int(offset)).Limit(int(pageSize)).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to list configs")
	}

	return &types.Pagination[models.EVMConfig]{
		Items:       items,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
	}, nil
}

// GetById retrieves a config by its ID with preloaded relationships.
func (q *ConfigQueries) GetById(id uint) (*models.EVMConfig, error) {
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

// Create creates a new config.
func (q *ConfigQueries) Create(config *models.EVMConfig) error {
	if err := q.db.Create(config).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to create config")
	}
	return nil
}

// Update updates a config by ID with the provided updates.
func (q *ConfigQueries) Update(id uint, updates map[string]interface{}) error {
	result := q.db.Model(&models.EVMConfig{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to update config")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "config not found")
	}
	return nil
}

// Delete deletes a config by ID.
func (q *ConfigQueries) Delete(id uint) error {
	result := q.db.Delete(&models.EVMConfig{}, id)
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

// Count returns the total number of configs.
func (q *ConfigQueries) Count() (int64, error) {
	var count int64
	if err := q.db.Model(&models.EVMConfig{}).Count(&count).Error; err != nil {
		return 0, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count configs")
	}
	return count, nil
}
