package signer

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Signer interface {
	// SignTransaction signs a transaction and returns the signed transaction
	SignTransaction(tx *types.Transaction) (signedTx *types.Transaction, err error)
	// SignMessage signs an arbitrary message and returns the signature
	SignMessageString(message string) (signature string, err error)
	// VerifyMessageString verifies a signature against a message for a given address
	VerifyMessageString(address common.Address, message string, signature string) (isValid bool, recoveredAddress common.Address, err error)
}
