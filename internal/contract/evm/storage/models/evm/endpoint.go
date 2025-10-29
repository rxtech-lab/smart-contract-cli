package models

import "time"

type EVMEndpoint struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null"`
	Url       string    `json:"url" gorm:"not null"`
	ChainId   string    `json:"chain_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for EVMEndpoint.
func (EVMEndpoint) TableName() string {
	return "evm_endpoints"
}
