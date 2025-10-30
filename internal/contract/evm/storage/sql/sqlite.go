package sql

import (
	"os"
	"path/filepath"

	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql/queries"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteStorage struct {
	db              *gorm.DB
	abiQueries      *queries.AbiQueries
	endpointQueries *queries.EndpointQueries
	contractQueries *queries.ContractQueries
	configQueries   *queries.ConfigQueries
}

// CountAbis implements Storage.
func (s *SQLiteStorage) CountAbis() (count int64, err error) {
	panic("unimplemented")
}

// CountConfigs implements Storage.
func (s *SQLiteStorage) CountConfigs() (count int64, err error) {
	panic("unimplemented")
}

// CountContracts implements Storage.
func (s *SQLiteStorage) CountContracts() (count int64, err error) {
	panic("unimplemented")
}

// CountEndpoints implements Storage.
func (s *SQLiteStorage) CountEndpoints() (count int64, err error) {
	panic("unimplemented")
}

// CreateAbi implements Storage.
func (s *SQLiteStorage) CreateAbi(abi models.EvmAbi) (id uint, err error) {
	panic("unimplemented")
}

// CreateConfig implements Storage.
func (s *SQLiteStorage) CreateConfig(config models.EVMConfig) (id uint, err error) {
	panic("unimplemented")
}

// CreateContract implements Storage.
func (s *SQLiteStorage) CreateContract(contract models.EVMContract) (id uint, err error) {
	panic("unimplemented")
}

// CreateEndpoint implements Storage.
func (s *SQLiteStorage) CreateEndpoint(endpoint models.EVMEndpoint) (id uint, err error) {
	panic("unimplemented")
}

// DeleteAbi implements Storage.
func (s *SQLiteStorage) DeleteAbi(id uint) (err error) {
	panic("unimplemented")
}

// DeleteConfig implements Storage.
func (s *SQLiteStorage) DeleteConfig(id uint) (err error) {
	panic("unimplemented")
}

// DeleteContract implements Storage.
func (s *SQLiteStorage) DeleteContract(id uint) (err error) {
	panic("unimplemented")
}

// DeleteEndpoint implements Storage.
func (s *SQLiteStorage) DeleteEndpoint(id uint) (err error) {
	panic("unimplemented")
}

// GetAbiById implements Storage.
func (s *SQLiteStorage) GetAbiById(id uint) (abi models.EvmAbi, err error) {
	panic("unimplemented")
}

// GetConfigById implements Storage.
func (s *SQLiteStorage) GetConfigById(id uint) (config models.EVMConfig, err error) {
	panic("unimplemented")
}

// GetContractById implements Storage.
func (s *SQLiteStorage) GetContractById(id uint) (contract models.EVMContract, err error) {
	panic("unimplemented")
}

// GetEndpointById implements Storage.
func (s *SQLiteStorage) GetEndpointById(id uint) (endpoint models.EVMEndpoint, err error) {
	panic("unimplemented")
}

// ListAbis implements Storage.
func (s *SQLiteStorage) ListAbis(page int64, pageSize int64) (abis types.Pagination[models.EvmAbi], err error) {
	panic("unimplemented")
}

// ListConfigs implements Storage.
func (s *SQLiteStorage) ListConfigs(page int64, pageSize int64) (configs types.Pagination[models.EVMConfig], err error) {
	panic("unimplemented")
}

// ListContracts implements Storage.
func (s *SQLiteStorage) ListContracts(page int64, pageSize int64) (contracts types.Pagination[models.EVMContract], err error) {
	panic("unimplemented")
}

// ListEndpoints implements Storage.
func (s *SQLiteStorage) ListEndpoints(page int64, pageSize int64) (endpoints types.Pagination[models.EVMEndpoint], err error) {
	panic("unimplemented")
}

// SearchAbis implements Storage.
func (s *SQLiteStorage) SearchAbis(query string) (abis types.Pagination[models.EvmAbi], err error) {
	panic("unimplemented")
}

// SearchConfigs implements Storage.
func (s *SQLiteStorage) SearchConfigs(query string) (configs types.Pagination[models.EVMConfig], err error) {
	panic("unimplemented")
}

// SearchContracts implements Storage.
func (s *SQLiteStorage) SearchContracts(query string) (contracts types.Pagination[models.EVMContract], err error) {
	panic("unimplemented")
}

// SearchEndpoints implements Storage.
func (s *SQLiteStorage) SearchEndpoints(query string) (endpoints types.Pagination[models.EVMEndpoint], err error) {
	panic("unimplemented")
}

// UpdateAbi implements Storage.
func (s *SQLiteStorage) UpdateAbi(id uint, abi models.EvmAbi) (err error) {
	panic("unimplemented")
}

// UpdateConfig implements Storage.
func (s *SQLiteStorage) UpdateConfig(id uint, config models.EVMConfig) (err error) {
	panic("unimplemented")
}

// UpdateContract implements Storage.
func (s *SQLiteStorage) UpdateContract(id uint, contract models.EVMContract) (err error) {
	panic("unimplemented")
}

// UpdateEndpoint implements Storage.
func (s *SQLiteStorage) UpdateEndpoint(id uint, endpoint models.EVMEndpoint) (err error) {
	panic("unimplemented")
}

// NewSQLiteDB creates a new SQLite database connection.
// If dbPath is empty, it defaults to $HOME/smart-contract-cli.db
func NewSQLiteDB(dbPath string) (Storage, error) {
	// Use default path if none provided
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(homeDir, "smart-contract-cli.db")
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}
