package models

import (
	"time"

	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
)

type DeploymentStatus string

const (
	DeploymentStatusPending  DeploymentStatus = "pending"
	DeploymentStatusDeployed DeploymentStatus = "deployed"
	DeploymentStatusFailed   DeploymentStatus = "failed"
)

type EVMContract struct {
	ID      uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	Name    string           `json:"name" gorm:"not null;uniqueIndex:idx_contract_name_address_endpoint"`
	Address string           `json:"address" gorm:"not null;uniqueIndex:idx_contract_name_address_endpoint"`
	AbiId   *uint            `json:"abi_id" gorm:"index;constraint:OnDelete:SET NULL"`
	Abi     *EvmAbi          `json:"abi,omitempty" gorm:"foreignKey:AbiId;references:ID"`
	Status  DeploymentStatus `json:"status" gorm:"default:pending"`

	ContractCode *string   `json:"contract_code" gorm:"type:text"`
	Bytecode     *string   `json:"bytecode" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	EndpointId uint         `json:"endpoint_id" gorm:"not null;index;uniqueIndex:idx_contract_name_address_endpoint;constraint:OnDelete:CASCADE"`
	Endpoint   *EVMEndpoint `json:"endpoint,omitempty" gorm:"foreignKey:EndpointId;references:ID"`
}

// TableName specifies the table name for EVMContract.
func (EVMContract) TableName() string {
	return "evm_contracts"
}

// IsDeployable returns true if the contract is deployable.
func (c *EVMContract) IsDeployable() bool {
	return c.Status == DeploymentStatusPending && c.Bytecode != nil
}

func (c *EVMContract) Compile() error {
	if c.ContractCode == nil {
		return errors.NewContractError(errors.ErrCodeContractCodeRequired, "contract code is required")
	}

	return nil
}
