package queries

import (
	"errors"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	customerrors "github.com/rxtech-lab/smart-contract-cli/internal/errors"
	"gorm.io/gorm"
)

// WalletQueries provides database operations for EVMWallet model.
type WalletQueries struct {
	db *gorm.DB
}

// NewWalletQueries creates a new WalletQueries instance.
func NewWalletQueries(db *gorm.DB) *WalletQueries {
	return &WalletQueries{db: db}
}

// List retrieves a paginated list of wallets.
func (q *WalletQueries) List(page int64, pageSize int64) (*types.Pagination[models.EVMWallet], error) {
	if page < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageNumber, "page number must be greater than 0")
	}
	if pageSize < 1 {
		return nil, customerrors.NewDatabaseError(customerrors.ErrCodeInvalidPageSize, "page size must be greater than 0")
	}

	var items []models.EVMWallet
	var totalItems int64

	// Count total items
	if err := q.db.Model(&models.EVMWallet{}).Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count wallets")
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	// Retrieve paginated items
	offset := (page - 1) * pageSize
	if err := q.db.
		Offset(int(offset)).Limit(int(pageSize)).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to list wallets")
	}

	return &types.Pagination[models.EVMWallet]{
		Items:       items,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
	}, nil
}

// GetByID retrieves a wallet by its ID.
func (q *WalletQueries) GetByID(id uint) (*models.EVMWallet, error) {
	var wallet models.EVMWallet
	if err := q.db.First(&wallet, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "wallet not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get wallet by ID")
	}
	return &wallet, nil
}

// GetByAddress retrieves a wallet by its address.
func (q *WalletQueries) GetByAddress(address string) (*models.EVMWallet, error) {
	var wallet models.EVMWallet
	if err := q.db.Where("address = ?", address).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "wallet not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get wallet by address")
	}
	return &wallet, nil
}

// GetByAlias retrieves a wallet by its alias.
func (q *WalletQueries) GetByAlias(alias string) (*models.EVMWallet, error) {
	var wallet models.EVMWallet
	if err := q.db.Where("alias = ?", alias).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeRecordNotFound, "wallet not found")
		}
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to get wallet by alias")
	}
	return &wallet, nil
}

// Create creates a new wallet.
func (q *WalletQueries) Create(wallet *models.EVMWallet) error {
	if err := q.db.Create(wallet).Error; err != nil {
		return customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to create wallet")
	}
	return nil
}

// Update updates a wallet by ID with the provided updates.
func (q *WalletQueries) Update(id uint, updates map[string]interface{}) error {
	result := q.db.Model(&models.EVMWallet{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to update wallet")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "wallet not found")
	}
	return nil
}

// Delete deletes a wallet by ID.
func (q *WalletQueries) Delete(id uint) error {
	result := q.db.Delete(&models.EVMWallet{}, id)
	if result.Error != nil {
		return customerrors.WrapDatabaseError(result.Error, customerrors.ErrCodeDatabaseOperationFailed, "failed to delete wallet")
	}
	if result.RowsAffected == 0 {
		return customerrors.NewDatabaseError(customerrors.ErrCodeRecordNotFound, "wallet not found")
	}
	return nil
}

// Exists checks if a wallet with the given ID exists.
func (q *WalletQueries) Exists(id uint) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EVMWallet{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check wallet existence")
	}
	return count > 0, nil
}

// ExistsByAddress checks if a wallet with the given address exists.
func (q *WalletQueries) ExistsByAddress(address string) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EVMWallet{}).Where("address = ?", address).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check wallet existence by address")
	}
	return count > 0, nil
}

// ExistsByAlias checks if a wallet with the given alias exists.
func (q *WalletQueries) ExistsByAlias(alias string) (bool, error) {
	var count int64
	if err := q.db.Model(&models.EVMWallet{}).Where("alias = ?", alias).Count(&count).Error; err != nil {
		return false, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to check wallet existence by alias")
	}
	return count > 0, nil
}

// Count returns the total number of wallets.
func (q *WalletQueries) Count() (int64, error) {
	var count int64
	if err := q.db.Model(&models.EVMWallet{}).Count(&count).Error; err != nil {
		return 0, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count wallets")
	}
	return count, nil
}

// Search searches for wallets by alias or address.
func (q *WalletQueries) Search(query string) (*types.Pagination[models.EVMWallet], error) {
	var items []models.EVMWallet
	var totalItems int64

	searchPattern := "%" + query + "%"

	// Count total matching items
	if err := q.db.Model(&models.EVMWallet{}).
		Where("alias LIKE ? OR address LIKE ?", searchPattern, searchPattern).
		Count(&totalItems).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to count wallets")
	}

	// Retrieve all matching items
	if err := q.db.
		Where("alias LIKE ? OR address LIKE ?", searchPattern, searchPattern).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, customerrors.WrapDatabaseError(err, customerrors.ErrCodeDatabaseOperationFailed, "failed to search wallets")
	}

	return &types.Pagination[models.EVMWallet]{
		Items:       items,
		TotalPages:  1,
		CurrentPage: 1,
		PageSize:    totalItems,
		TotalItems:  totalItems,
	}, nil
}
