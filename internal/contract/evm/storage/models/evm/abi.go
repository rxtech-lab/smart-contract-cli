package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
)

type EvmAbi struct {
	ID        uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string       `json:"name" gorm:"uniqueIndex;not null"`
	Abi       AbiArrayType `json:"abi" gorm:"type:text"`
	CreatedAt time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for EvmAbi
func (EvmAbi) TableName() string {
	return "evm_abis"
}

// AbiArrayType wraps abi.AbiArray for database serialization
type AbiArrayType struct {
	abi.AbiArray
}

// Scan implements sql.Scanner interface for reading from database
func (a *AbiArrayType) Scan(value any) error {
	if value == nil {
		a.AbiArray = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	parsed, err := abi.ParseAbi(string(bytes))
	if err != nil {
		return err
	}

	a.AbiArray = parsed
	return nil
}

// Value implements driver.Valuer interface for writing to database
func (a AbiArrayType) Value() (driver.Value, error) {
	if a.AbiArray == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(a.AbiArray)
	if err != nil {
		return nil, err
	}

	return string(bytes), nil
}
