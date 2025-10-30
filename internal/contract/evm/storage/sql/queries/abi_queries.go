package queries

import (
	"errors"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	customerrors "github.com/rxtech-lab/smart-contract-cli/internal/errors"
	"gorm.io/gorm"
)

// ABIQueries provides database operations for EvmAbi model.
type ABIQueries struct {
	db *gorm.DB
}

// NewABIQueries creates a new ABIQueries instance.
func NewABIQueries(db *gorm.DB) *ABIQueries {
	return &ABIQueries{db: db}
}

// List retrieves a paginated list of ABIs.
func (q *ABIQueries) List(page int64, pageSize int64) (*types.Pagination[models.EvmAbi], error) {
	if page < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageNumber, "page number must be greater than 0")
	}
	if pageSize < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageSize, "page size must be greater than 0")
	}

	var items []models.EvmAbi
	var totalItems int64

	// Count total items
	if err := q.db.Model(&models.EvmAbi{}).Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count ABIs")
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Retrieve paginated items
	offset := (page - 1) * pageSize
	if err := q.db.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to list ABIs")
	}

	return &types.Pagination[models.EvmAbi]{
		Items:       items,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
	}, nil
}

// GetByID retrieves an ABI by its ID.
func (q *ABIQueries) GetByID(id uint) (*models.EvmAbi, error) {
	var abi models.EvmAbi
	if err := q.db.First(&abi, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "ABI not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get ABI by ID")
	}
	return &abi, nil
}

// Create creates a new ABI.
func (q *ABIQueries) Create(abi *models.EvmAbi) error {
	if err := q.db.Create(abi).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to create ABI")
	}
	return nil
}

// Update updates an ABI by ID with the provided updates.
func (q *ABIQueries) Update(id uint, updates map[string]interface{}) error {
	result := q.db.Model(&models.EvmAbi{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to update ABI")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "ABI not found")
	}
	return nil
}

// Delete deletes an ABI by ID.
func (q *ABIQueries) Delete(id uint) error {
	result := q.db.Delete(&models.EvmAbi{}, id)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to delete ABI")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "ABI not found")
	}
	return nil
}

// Exists checks if an ABI with the given ID exists.
func (q *ABIQueries) Exists(id uint) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EvmAbi{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check ABI existence")
	}
	return count > 0, nil
}

// Count returns the total number of ABIs.
func (q *ABIQueries) Count() (int64, error) {
	var count int64
	if err := q.db.Model(&models.EvmAbi{}).Count(&count).Error; err != nil {
		return 0, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count ABIs")
	}
	return count, nil
}

// Search searches for ABIs by name.
func (q *ABIQueries) Search(query string) (*types.Pagination[models.EvmAbi], error) {
	var items []models.EvmAbi
	var totalItems int64

	searchPattern := "%" + query + "%"

	// Count total matching items
	if err := q.db.Model(&models.EvmAbi{}).Where("name LIKE ?", searchPattern).Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count ABIs")
	}

	// Retrieve all matching items
	if err := q.db.Where("name LIKE ?", searchPattern).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to search ABIs")
	}

	return &types.Pagination[models.EvmAbi]{
		Items:       items,
		TotalPages:  1,
		CurrentPage: 1,
		PageSize:    totalItems,
		TotalItems:  totalItems,
	}, nil
}
