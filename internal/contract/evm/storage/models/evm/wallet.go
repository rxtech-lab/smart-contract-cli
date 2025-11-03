package models

import (
	"fmt"
	"time"
)

// EVMWallet represents a wallet entity in the database.
// Private keys and mnemonics are stored separately in secure storage.
type EVMWallet struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Alias   string `json:"alias" gorm:"not null;uniqueIndex"`
	Address string `json:"address" gorm:"not null;uniqueIndex"`

	// DerivationPath is the HD wallet derivation path (e.g., "m/44'/60'/0'/0/0")
	// Only set when wallet is derived from a mnemonic
	DerivationPath *string `json:"derivation_path" gorm:"type:varchar(100)"`

	// IsFromMnemonic indicates if this wallet was created from a mnemonic phrase
	// If true, both mnemonic and private key are stored in secure storage
	// If false, only private key is stored
	IsFromMnemonic bool `json:"is_from_mnemonic" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for EVMWallet.
func (EVMWallet) TableName() string {
	return "evm_wallets"
}

// GetSecureStorageKeyPrivateKey returns the secure storage key for the wallet's private key.
func (w *EVMWallet) GetSecureStorageKeyPrivateKey() string {
	return GetWalletPrivateKeyStorageKey(w.ID)
}

// GetSecureStorageKeyMnemonic returns the secure storage key for the wallet's mnemonic.
func (w *EVMWallet) GetSecureStorageKeyMnemonic() string {
	return GetWalletMnemonicStorageKey(w.ID)
}

// GetWalletPrivateKeyStorageKey returns the secure storage key format for a wallet's private key.
func GetWalletPrivateKeyStorageKey(walletID uint) string {
	return fmt.Sprintf("wallet:%d:privatekey", walletID)
}

// GetWalletMnemonicStorageKey returns the secure storage key format for a wallet's mnemonic.
func GetWalletMnemonicStorageKey(walletID uint) string {
	return fmt.Sprintf("wallet:%d:mnemonic", walletID)
}
