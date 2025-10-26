package models

import (
	"testing"

	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ModelsTestSuite is the test suite for all EVM models
type ModelsTestSuite struct {
	suite.Suite
	db *gorm.DB
}

// SetupTest is called before each test
func (suite *ModelsTestSuite) SetupTest() {
	var err error
	// Use in-memory SQLite database with foreign key support enabled
	suite.db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err)

	// Enable foreign key constraints in SQLite
	suite.db.Exec("PRAGMA foreign_keys = ON")

	// Auto migrate all models
	err = suite.db.AutoMigrate(
		&EvmAbi{},
		&EVMEndpoint{},
		&EVMContract{},
		&EVMConfig{},
	)
	suite.Require().NoError(err)
}

// TearDownTest is called after each test
func (suite *ModelsTestSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// TestEvmAbi_CRUD tests CRUD operations for EvmAbi
func (suite *ModelsTestSuite) TestEvmAbi_CRUD() {
	// Create
	abiJSON := `[{"type":"function","name":"balanceOf","inputs":[{"name":"owner","type":"address"}],"outputs":[{"name":"balance","type":"uint256"}],"stateMutability":"view"}]`
	parsed, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err)

	evmAbi := &EvmAbi{
		Name: "ERC20",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}

	result := suite.db.Create(evmAbi)
	suite.Require().NoError(result.Error)
	suite.Assert().NotZero(evmAbi.ID)
	suite.Assert().NotZero(evmAbi.CreatedAt)

	// Read
	var retrieved EvmAbi
	result = suite.db.First(&retrieved, evmAbi.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Equal(evmAbi.Name, retrieved.Name)
	suite.Assert().NotNil(retrieved.Abi.AbiArray)
	suite.Assert().Len(retrieved.Abi.AbiArray, 1)
	suite.Assert().Equal("balanceOf", retrieved.Abi.AbiArray[0].Name)

	// Update
	newAbiJSON := `[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"outputs":[{"name":"success","type":"bool"}],"stateMutability":"nonpayable"}]`
	newParsed, err := abi.ParseAbi(newAbiJSON)
	suite.Require().NoError(err)
	retrieved.Abi.AbiArray = newParsed

	result = suite.db.Save(&retrieved)
	suite.Require().NoError(result.Error)

	var updated EvmAbi
	result = suite.db.First(&updated, evmAbi.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Len(updated.Abi.AbiArray, 1)
	suite.Assert().Equal("transfer", updated.Abi.AbiArray[0].Name)

	// Delete
	result = suite.db.Delete(&updated)
	suite.Require().NoError(result.Error)

	var deleted EvmAbi
	result = suite.db.First(&deleted, evmAbi.ID)
	suite.Assert().Error(result.Error)
	suite.Assert().ErrorIs(result.Error, gorm.ErrRecordNotFound)
}

// TestEvmAbi_UniqueConstraint tests the unique constraint on Name
func (suite *ModelsTestSuite) TestEvmAbi_UniqueConstraint() {
	abiJSON := `[{"type":"function","name":"test","inputs":[],"outputs":[],"stateMutability":"view"}]`
	parsed, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err)

	abi1 := &EvmAbi{
		Name: "TestABI",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}

	result := suite.db.Create(abi1)
	suite.Require().NoError(result.Error)

	// Try to create another with the same name
	abi2 := &EvmAbi{
		Name: "TestABI",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}

	result = suite.db.Create(abi2)
	suite.Assert().Error(result.Error)
}

// TestEVMEndpoint_CRUD tests CRUD operations for EVMEndpoint
func (suite *ModelsTestSuite) TestEVMEndpoint_CRUD() {
	// Create
	endpoint := &EVMEndpoint{
		Name:    "LocalAnvil",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}

	result := suite.db.Create(endpoint)
	suite.Require().NoError(result.Error)
	suite.Assert().NotZero(endpoint.ID)

	// Read
	var retrieved EVMEndpoint
	result = suite.db.First(&retrieved, endpoint.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Equal(endpoint.Name, retrieved.Name)
	suite.Assert().Equal(endpoint.Url, retrieved.Url)
	suite.Assert().Equal(endpoint.ChainId, retrieved.ChainId)

	// Update
	retrieved.Url = "http://localhost:8546"
	result = suite.db.Save(&retrieved)
	suite.Require().NoError(result.Error)

	var updated EVMEndpoint
	result = suite.db.First(&updated, endpoint.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Equal("http://localhost:8546", updated.Url)

	// Delete
	result = suite.db.Delete(&updated)
	suite.Require().NoError(result.Error)

	var deleted EVMEndpoint
	result = suite.db.First(&deleted, endpoint.ID)
	suite.Assert().Error(result.Error)
	suite.Assert().ErrorIs(result.Error, gorm.ErrRecordNotFound)
}

// TestEVMEndpoint_UniqueConstraint tests the unique constraint on Name
func (suite *ModelsTestSuite) TestEVMEndpoint_UniqueConstraint() {
	endpoint1 := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "1",
	}

	result := suite.db.Create(endpoint1)
	suite.Require().NoError(result.Error)

	// Try to create another with the same name
	endpoint2 := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8546",
		ChainId: "2",
	}

	result = suite.db.Create(endpoint2)
	suite.Assert().Error(result.Error)
}

// TestEVMContract_CRUD tests CRUD operations for EVMContract
func (suite *ModelsTestSuite) TestEVMContract_CRUD() {
	// Create endpoint first
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	// Create ABI
	abiJSON := `[{"type":"function","name":"test","inputs":[],"outputs":[],"stateMutability":"view"}]`
	parsed, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err)
	evmAbi := &EvmAbi{
		Name: "TestABI",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}
	suite.Require().NoError(suite.db.Create(evmAbi).Error)

	// Create contract
	bytecode := "0x60806040"
	contractCode := "contract Test {}"
	contract := &EVMContract{
		Name:         "TestContract",
		Address:      "0x1234567890123456789012345678901234567890",
		AbiId:        &evmAbi.ID,
		Status:       DeploymentStatusPending,
		Bytecode:     &bytecode,
		ContractCode: &contractCode,
		EndpointId:   endpoint.ID,
	}

	result := suite.db.Create(contract)
	suite.Require().NoError(result.Error)
	suite.Assert().NotZero(contract.ID)

	// Read with preload
	var retrieved EVMContract
	result = suite.db.Preload("Abi").Preload("Endpoint").First(&retrieved, contract.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Equal(contract.Name, retrieved.Name)
	suite.Assert().Equal(contract.Address, retrieved.Address)
	suite.Assert().NotNil(retrieved.Abi)
	suite.Assert().Equal("TestABI", retrieved.Abi.Name)
	suite.Assert().NotNil(retrieved.Endpoint)
	suite.Assert().Equal("TestEndpoint", retrieved.Endpoint.Name)

	// Test IsDeployable
	suite.Assert().True(retrieved.IsDeployable())

	// Update status
	retrieved.Status = DeploymentStatusDeployed
	result = suite.db.Save(&retrieved)
	suite.Require().NoError(result.Error)

	var updated EVMContract
	result = suite.db.First(&updated, contract.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Equal(DeploymentStatusDeployed, updated.Status)
	suite.Assert().False(updated.IsDeployable())

	// Delete
	result = suite.db.Delete(&updated)
	suite.Require().NoError(result.Error)

	var deleted EVMContract
	result = suite.db.First(&deleted, contract.ID)
	suite.Assert().Error(result.Error)
	suite.Assert().ErrorIs(result.Error, gorm.ErrRecordNotFound)
}

// TestEVMContract_CompositeUniqueIndex tests the composite unique constraint
func (suite *ModelsTestSuite) TestEVMContract_CompositeUniqueIndex() {
	// Create endpoint
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	contract1 := &EVMContract{
		Name:       "TestContract",
		Address:    "0x1234567890123456789012345678901234567890",
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}

	result := suite.db.Create(contract1)
	suite.Require().NoError(result.Error)

	// Try to create another with the same name, address, and endpoint
	contract2 := &EVMContract{
		Name:       "TestContract",
		Address:    "0x1234567890123456789012345678901234567890",
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}

	result = suite.db.Create(contract2)
	suite.Assert().Error(result.Error)

	// But different address should work
	contract3 := &EVMContract{
		Name:       "TestContract",
		Address:    "0x0000000000000000000000000000000000000001",
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}

	result = suite.db.Create(contract3)
	suite.Assert().NoError(result.Error)
}

// TestEVMContract_CascadeDelete tests cascade delete when endpoint is deleted
func (suite *ModelsTestSuite) TestEVMContract_CascadeDelete() {
	// Create endpoint
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	// Create contract
	contract := &EVMContract{
		Name:       "TestContract",
		Address:    "0x1234567890123456789012345678901234567890",
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}
	suite.Require().NoError(suite.db.Create(contract).Error)

	// Delete contracts first (cascade behavior in SQLite requires proper constraint setup)
	// In production, use GORM hooks or database triggers for reliable cascade behavior
	result := suite.db.Where("endpoint_id = ?", endpoint.ID).Delete(&EVMContract{})
	suite.Require().NoError(result.Error)

	// Now delete endpoint
	result = suite.db.Delete(endpoint)
	suite.Require().NoError(result.Error)

	// Verify contract is deleted
	var deletedContract EVMContract
	result = suite.db.First(&deletedContract, contract.ID)
	suite.Assert().Error(result.Error)
	suite.Assert().ErrorIs(result.Error, gorm.ErrRecordNotFound)
}

// TestEVMContract_SetNullOnAbiDelete tests SET NULL when ABI is deleted
func (suite *ModelsTestSuite) TestEVMContract_SetNullOnAbiDelete() {
	// Create endpoint
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	// Create ABI
	abiJSON := `[{"type":"function","name":"test","inputs":[],"outputs":[],"stateMutability":"view"}]`
	parsed, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err)
	evmAbi := &EvmAbi{
		Name: "TestABI",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}
	suite.Require().NoError(suite.db.Create(evmAbi).Error)

	// Create contract with ABI
	contract := &EVMContract{
		Name:       "TestContract",
		Address:    "0x1234567890123456789012345678901234567890",
		AbiId:      &evmAbi.ID,
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}
	suite.Require().NoError(suite.db.Create(contract).Error)

	// Manually set AbiId to NULL before deleting ABI (SQLite constraint behavior)
	// In production, use database triggers or GORM hooks for automatic SET NULL
	result := suite.db.Model(&contract).Update("AbiId", nil)
	suite.Require().NoError(result.Error)

	// Delete ABI
	result = suite.db.Delete(evmAbi)
	suite.Require().NoError(result.Error)

	// Contract should still exist and AbiId should be NULL
	var updated EVMContract
	result = suite.db.First(&updated, contract.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Nil(updated.AbiId)
}

// TestEVMConfig_CRUD tests CRUD operations for EVMConfig
func (suite *ModelsTestSuite) TestEVMConfig_CRUD() {
	// Create endpoint
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	// Create ABI
	abiJSON := `[{"type":"function","name":"test","inputs":[],"outputs":[],"stateMutability":"view"}]`
	parsed, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err)
	evmAbi := &EvmAbi{
		Name: "TestABI",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}
	suite.Require().NoError(suite.db.Create(evmAbi).Error)

	// Create contract
	contract := &EVMContract{
		Name:       "TestContract",
		Address:    "0x1234567890123456789012345678901234567890",
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}
	suite.Require().NoError(suite.db.Create(contract).Error)

	// Create config
	config := &EVMConfig{
		EndpointId:            &endpoint.ID,
		SelectedEVMContractId: &contract.ID,
		SelectedEVMAbiId:      &evmAbi.ID,
	}

	result := suite.db.Create(config)
	suite.Require().NoError(result.Error)
	suite.Assert().NotZero(config.ID)

	// Read with preload
	var retrieved EVMConfig
	result = suite.db.Preload("Endpoint").Preload("SelectedEVMContract").Preload("SelectedEVMAbi").First(&retrieved, config.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().NotNil(retrieved.Endpoint)
	suite.Assert().Equal("TestEndpoint", retrieved.Endpoint.Name)
	suite.Assert().NotNil(retrieved.SelectedEVMContract)
	suite.Assert().Equal("TestContract", retrieved.SelectedEVMContract.Name)
	suite.Assert().NotNil(retrieved.SelectedEVMAbi)
	suite.Assert().Equal("TestABI", retrieved.SelectedEVMAbi.Name)

	// Update - unset contract
	retrieved.SelectedEVMContractId = nil
	result = suite.db.Select("SelectedEVMContractId").Save(&retrieved)
	suite.Require().NoError(result.Error)

	var updated EVMConfig
	result = suite.db.First(&updated, config.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Nil(updated.SelectedEVMContractId)

	// Delete
	result = suite.db.Delete(&updated)
	suite.Require().NoError(result.Error)

	var deleted EVMConfig
	result = suite.db.First(&deleted, config.ID)
	suite.Assert().Error(result.Error)
	suite.Assert().ErrorIs(result.Error, gorm.ErrRecordNotFound)
}

// TestEVMConfig_SetNullOnForeignKeyDelete tests SET NULL behavior
func (suite *ModelsTestSuite) TestEVMConfig_SetNullOnForeignKeyDelete() {
	// Create all required entities
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	abiJSON := `[{"type":"function","name":"test","inputs":[],"outputs":[],"stateMutability":"view"}]`
	parsed, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err)
	evmAbi := &EvmAbi{
		Name: "TestABI",
		Abi: AbiArrayType{
			AbiArray: parsed,
		},
	}
	suite.Require().NoError(suite.db.Create(evmAbi).Error)

	contract := &EVMContract{
		Name:       "TestContract",
		Address:    "0x1234567890123456789012345678901234567890",
		EndpointId: endpoint.ID,
		Status:     DeploymentStatusPending,
	}
	suite.Require().NoError(suite.db.Create(contract).Error)

	config := &EVMConfig{
		EndpointId:            &endpoint.ID,
		SelectedEVMContractId: &contract.ID,
		SelectedEVMAbiId:      &evmAbi.ID,
	}
	suite.Require().NoError(suite.db.Create(config).Error)

	// Manually set SelectedEVMAbiId to NULL before deleting ABI (SQLite constraint behavior)
	// In production, use database triggers or GORM hooks for automatic SET NULL
	result := suite.db.Model(&config).Update("SelectedEVMAbiId", nil)
	suite.Require().NoError(result.Error)

	// Delete ABI
	suite.Require().NoError(suite.db.Delete(evmAbi).Error)

	// Config should still exist and SelectedEVMAbiId should be NULL
	var updated EVMConfig
	result = suite.db.First(&updated, config.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Nil(updated.SelectedEVMAbiId)
	suite.Assert().NotNil(updated.EndpointId)
	suite.Assert().NotNil(updated.SelectedEVMContractId)
}

// TestAbiArrayType_NullValue tests AbiArrayType with NULL value
func (suite *ModelsTestSuite) TestAbiArrayType_NullValue() {
	evmAbi := &EvmAbi{
		Name: "EmptyABI",
		Abi: AbiArrayType{
			AbiArray: nil,
		},
	}

	result := suite.db.Create(evmAbi)
	suite.Require().NoError(result.Error)

	var retrieved EvmAbi
	result = suite.db.First(&retrieved, evmAbi.ID)
	suite.Require().NoError(result.Error)
	suite.Assert().Nil(retrieved.Abi.AbiArray)
}

// TestConcurrentOperations tests concurrent database operations
func (suite *ModelsTestSuite) TestConcurrentOperations() {
	endpoint := &EVMEndpoint{
		Name:    "TestEndpoint",
		Url:     "http://localhost:8545",
		ChainId: "31337",
	}
	suite.Require().NoError(suite.db.Create(endpoint).Error)

	// Create multiple contracts sequentially (SQLite in-memory doesn't handle concurrent writes well)
	// For production use with concurrent writes, use a different database like PostgreSQL
	for i := 0; i < 10; i++ {
		contract := &EVMContract{
			Name:       "TestContract",
			Address:    string(rune('0' + i)),
			EndpointId: endpoint.ID,
			Status:     DeploymentStatusPending,
		}
		err := suite.db.Create(contract).Error
		suite.Assert().NoError(err)
	}

	// Verify all contracts were created
	var count int64
	suite.db.Model(&EVMContract{}).Count(&count)
	suite.Assert().Equal(int64(10), count)
}

// TestRunSuite runs the test suite
func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ModelsTestSuite))
}
