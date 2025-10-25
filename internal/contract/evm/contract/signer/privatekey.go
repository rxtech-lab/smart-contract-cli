package signer

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type PrivateKeySigner struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewPrivateKeySigner(privateKey string) (Signer, error) {
	privKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &PrivateKeySigner{
		PrivateKey: privKey,
	}, nil
}

// SignMessageString implements Signer.
func (p *PrivateKeySigner) SignMessageString(message string) (signature string, err error) {
	panic("unimplemented")
}

// SignTransaction implements Signer.
func (p *PrivateKeySigner) SignTransaction(tx *types.Transaction) (signedTx *types.Transaction, err error) {
	panic("unimplemented")
}

// VerifyMessageString implements Signer.
func (p *PrivateKeySigner) VerifyMessageString(address common.Address, message string, signature string) (isValid bool, recoveredAddress common.Address, err error) {
	panic("unimplemented")
}
