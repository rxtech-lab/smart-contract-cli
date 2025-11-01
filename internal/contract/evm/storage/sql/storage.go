package sql

import (
	"fmt"

	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	models "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/models/evm"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
)

type Storage interface {
	// ABI methods
	CreateABI(abi models.EvmAbi) (id uint, err error)
	ListABIs(page int64, pageSize int64) (abis types.Pagination[models.EvmAbi], err error)
	SearchABIs(query string) (abis types.Pagination[models.EvmAbi], err error)
	GetABIByID(id uint) (abi models.EvmAbi, err error)
	CountABIs() (count int64, err error)
	UpdateABI(id uint, abi models.EvmAbi) (err error)
	DeleteABI(id uint) (err error)

	// Endpoint methods
	CreateEndpoint(endpoint models.EVMEndpoint) (id uint, err error)
	ListEndpoints(page int64, pageSize int64) (endpoints types.Pagination[models.EVMEndpoint], err error)
	UpdateEndpoint(id uint, endpoint models.EVMEndpoint) (err error)
	SearchEndpoints(query string) (endpoints types.Pagination[models.EVMEndpoint], err error)
	GetEndpointByID(id uint) (endpoint models.EVMEndpoint, err error)
	CountEndpoints() (count int64, err error)
	DeleteEndpoint(id uint) (err error)

	// Contract methods
	CreateContract(contract models.EVMContract) (id uint, err error)
	ListContracts(page int64, pageSize int64) (contracts types.Pagination[models.EVMContract], err error)
	SearchContracts(query string) (contracts types.Pagination[models.EVMContract], err error)
	GetContractByID(id uint) (contract models.EVMContract, err error)
	CountContracts() (count int64, err error)
	UpdateContract(id uint, contract models.EVMContract) (err error)
	DeleteContract(id uint) (err error)

	// Config methods
	CreateConfig(config models.EVMConfig) (id uint, err error)
	ListConfigs(page int64, pageSize int64) (configs types.Pagination[models.EVMConfig], err error)
	SearchConfigs(query string) (configs types.Pagination[models.EVMConfig], err error)
	GetConfigByID(id uint) (config models.EVMConfig, err error)
	CountConfigs() (count int64, err error)
	UpdateConfig(id uint, config models.EVMConfig) (err error)
	DeleteConfig(id uint) (err error)
}

func GetStorage(storageType string, params ...any) (Storage, error) {
	switch storageType {
	case config.StorageClientTypeSQLite:
		if len(params) == 0 || params[0] == nil {
			return nil, fmt.Errorf("sqlite path is required")
		}
		sqlitePath, ok := params[0].(string)
		if !ok {
			return nil, fmt.Errorf("sqlite path must be a string")
		}
		return NewSQLiteDB(sqlitePath)
	default:
		return nil, fmt.Errorf("invalid storage type: %s", storageType)
	}
}
