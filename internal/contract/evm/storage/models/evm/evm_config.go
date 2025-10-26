package models

import "time"

type EVMConfig struct {
	ID                    uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	EndpointId            *uint        `json:"endpoint_id" gorm:"index;constraint:OnDelete:SET NULL"`
	Endpoint              *EVMEndpoint `json:"endpoint,omitempty" gorm:"foreignKey:EndpointId;references:ID"`
	SelectedEVMContractId *uint        `json:"selected_evm_contract_id" gorm:"index;constraint:OnDelete:SET NULL"`
	SelectedEVMContract   *EVMContract `json:"selected_evm_contract,omitempty" gorm:"foreignKey:SelectedEVMContractId;references:ID"`
	SelectedEVMAbiId      *uint        `json:"selected_evm_abi_id" gorm:"index;constraint:OnDelete:SET NULL"`
	SelectedEVMAbi        *EvmAbi      `json:"selected_evm_abi,omitempty" gorm:"foreignKey:SelectedEVMAbiId;references:ID"`
	CreatedAt             time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for EVMConfig
func (EVMConfig) TableName() string {
	return "evm_configs"
}
