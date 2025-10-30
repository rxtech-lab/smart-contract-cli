package queries

import (
	"errors"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	customerrors "github.com/rxtech-lab/smart-contract-cli/internal/errors"
	"gorm.io/gorm"
)

// ContractQueries provides database operations for EVMContract model.
type ContractQueries struct {
	db *gorm.DB
}

// NewContractQueries creates a new ContractQueries instance.
func NewContractQueries(db *gorm.DB) *ContractQueries {
	return &ContractQueries{db: db}
}

// List retrieves a paginated list of contracts with preloaded relationships.
func (q *ContractQueries) List(page int64, pageSize int64) (*types.Pagination[models.EVMContract], error) {
	if page < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageNumber, "page number must be greater than 0")
	}
	if pageSize < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageSize, "page size must be greater than 0")
	}

	var items []models.EVMContract
	var totalItems int64

	// Count total items
	if err := q.db.Model(&models.EVMContract{}).Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count contracts")
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Retrieve paginated items with preloaded relationships
	offset := (page - 1) * pageSize
	if err := q.db.Preload("Abi").Preload("Endpoint").
		Offset(int(offset)).Limit(int(pageSize)).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to list contracts")
	}

	return &types.Pagination[models.EVMContract]{
		Items:       items,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
	}, nil
}

// GetByID retrieves a contract by its ID with preloaded relationships.
func (q *ContractQueries) GetByID(id uint) (*models.EVMContract, error) {
	var contract models.EVMContract
	if err := q.db.Preload("Abi").Preload("Endpoint").First(&contract, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "contract not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get contract by ID")
	}
	return &contract, nil
}

// Create creates a new contract.
func (q *ContractQueries) Create(contract *models.EVMContract) error {
	if err := q.db.Create(contract).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to create contract")
	}
	return nil
}

// Update updates a contract by ID with the provided updates.
func (q *ContractQueries) Update(id uint, updates map[string]interface{}) error {
	result := q.db.Model(&models.EVMContract{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to update contract")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "contract not found")
	}
	return nil
}

// Delete deletes a contract by ID.
func (q *ContractQueries) Delete(id uint) error {
	result := q.db.Delete(&models.EVMContract{}, id)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to delete contract")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "contract not found")
	}
	return nil
}

// Exists checks if a contract with the given ID exists.
func (q *ContractQueries) Exists(id uint) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EVMContract{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check contract existence")
	}
	return count > 0, nil
}

// Count returns the total number of contracts.
func (q *ContractQueries) Count() (int64, error) {
	var count int64
	if err := q.db.Model(&models.EVMContract{}).Count(&count).Error; err != nil {
		return 0, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count contracts")
	}
	return count, nil
}

// Search searches for contracts by name or address.
func (q *ContractQueries) Search(query string) (*types.Pagination[models.EVMContract], error) {
	var items []models.EVMContract
	var totalItems int64

	searchPattern := "%" + query + "%"

	// Count total matching items
	if err := q.db.Model(&models.EVMContract{}).
		Where("name LIKE ? OR address LIKE ?", searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count contracts")
	}

	// Retrieve all matching items with preloaded relationships
	if err := q.db.Preload("Abi").Preload("Endpoint").
		Where("name LIKE ? OR address LIKE ?", searchPattern, searchPattern).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to search contracts")
	}

	return &types.Pagination[models.EVMContract]{
		Items:       items,
		TotalPages:  1,
		CurrentPage: 1,
		PageSize:    totalItems,
		TotalItems:  totalItems,
	}, nil
}
