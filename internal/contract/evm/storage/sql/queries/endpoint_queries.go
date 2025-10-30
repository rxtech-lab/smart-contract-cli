package queries

import (
	"errors"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	customerrors "github.com/rxtech-lab/smart-contract-cli/internal/errors"
	"gorm.io/gorm"
)

// EndpointQueries provides database operations for EVMEndpoint model.
type EndpointQueries struct {
	db *gorm.DB
}

// NewEndpointQueries creates a new EndpointQueries instance.
func NewEndpointQueries(db *gorm.DB) *EndpointQueries {
	return &EndpointQueries{db: db}
}

// List retrieves a paginated list of endpoints.
func (q *EndpointQueries) List(page int64, pageSize int64) (*types.Pagination[models.EVMEndpoint], error) {
	if page < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageNumber, "page number must be greater than 0")
	}
	if pageSize < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageSize, "page size must be greater than 0")
	}

	var items []models.EVMEndpoint
	var totalItems int64

	// Count total items
	if err := q.db.Model(&models.EVMEndpoint{}).Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count endpoints")
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Retrieve paginated items
	offset := (page - 1) * pageSize
	if err := q.db.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to list endpoints")
	}

	return &types.Pagination[models.EVMEndpoint]{
		Items:       items,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
	}, nil
}

// GetById retrieves an endpoint by its ID.
func (q *EndpointQueries) GetById(id uint) (*models.EVMEndpoint, error) {
	var endpoint models.EVMEndpoint
	if err := q.db.First(&endpoint, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "endpoint not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get endpoint by ID")
	}
	return &endpoint, nil
}

// Create creates a new endpoint.
func (q *EndpointQueries) Create(endpoint *models.EVMEndpoint) error {
	if err := q.db.Create(endpoint).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to create endpoint")
	}
	return nil
}

// Update updates an endpoint by ID with the provided updates.
func (q *EndpointQueries) Update(id uint, updates map[string]interface{}) error {
	result := q.db.Model(&models.EVMEndpoint{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to update endpoint")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "endpoint not found")
	}
	return nil
}

// Delete deletes an endpoint by ID.
func (q *EndpointQueries) Delete(id uint) error {
	result := q.db.Delete(&models.EVMEndpoint{}, id)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to delete endpoint")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "endpoint not found")
	}
	return nil
}

// Exists checks if an endpoint with the given ID exists.
func (q *EndpointQueries) Exists(id uint) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EVMEndpoint{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check endpoint existence")
	}
	return count > 0, nil
}

// Count returns the total number of endpoints.
func (q *EndpointQueries) Count() (int64, error) {
	var count int64
	if err := q.db.Model(&models.EVMEndpoint{}).Count(&count).Error; err != nil {
		return 0, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count endpoints")
	}
	return count, nil
}
