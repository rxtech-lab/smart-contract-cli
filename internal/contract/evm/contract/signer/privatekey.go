package signer

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
)

type PrivateKeySigner struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewPrivateKeySigner(privateKey string) (Signer, error) {
	// Remove 0x prefix if present
	if len(privateKey) > 2 && privateKey[:2] == "0x" {
		privateKey = privateKey[2:]
	}

	privKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, errors.WrapSignerError(err, errors.ErrCodeInvalidPrivateKey, "invalid private key format")
	}

	return &PrivateKeySigner{
		PrivateKey: privKey,
	}, nil
}

// SignMessageString implements Signer.
func (p *PrivateKeySigner) SignMessageString(message string) (signature string, err error) {
	// Hash the message using Keccak256 with Ethereum's message prefix
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Sign the hash
	sig, err := crypto.Sign(hash.Bytes(), p.PrivateKey)
	if err != nil {
		return "", errors.WrapSignerError(err, errors.ErrCodeSigningFailed, "failed to sign message")
	}

	// Return the signature as a hex string
	return "0x" + hex.EncodeToString(sig), nil
}

// SignTransaction implements Signer.
func (p *PrivateKeySigner) SignTransaction(transaction *types.Transaction) (signedTx *types.Transaction, err error) {
	// Get the chain ID from the transaction
	chainID := transaction.ChainId()
	if chainID == nil {
		return nil, errors.NewSignerError(errors.ErrCodeInvalidChainID, "transaction has no chain ID")
	}

	// Create a signer for the chain
	signer := types.NewLondonSigner(chainID)

	// Sign the transaction
	signedTx, err = types.SignTx(transaction, signer, p.PrivateKey)
	if err != nil {
		return nil, errors.WrapSignerError(err, errors.ErrCodeTransactionSignFailed, "failed to sign transaction")
	}

	return signedTx, nil
}

// VerifyMessageString implements Signer.
func (p *PrivateKeySigner) VerifyMessageString(address common.Address, message string, signature string) (isValid bool, recoveredAddress common.Address, err error) {
	// Remove 0x prefix if present
	if len(signature) > 2 && signature[:2] == "0x" {
		signature = signature[2:]
	}

	// Decode the signature from hex
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false, common.Address{}, errors.WrapSignerError(err, errors.ErrCodeSignatureDecode, "failed to decode signature")
	}

	// Ensure signature is 65 bytes (r, s, v)
	if len(sig) != 65 {
		return false, common.Address{}, errors.NewSignerErrorWithDetails(
			errors.ErrCodeInvalidSignatureLength,
			"invalid signature length",
			fmt.Sprintf("expected 65 bytes, got %d", len(sig)),
		)
	}

	// Hash the message with Ethereum's message prefix (same as signing)
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))

	// Adjust recovery id if needed (go-ethereum expects 0 or 1, but metamask sends 27 or 28)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	// Recover the public key from the signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false, common.Address{}, errors.WrapSignerError(err, errors.ErrCodePublicKeyRecovery, "failed to recover public key")
	}

	// Get the address from the recovered public key
	recoveredAddress = crypto.PubkeyToAddress(*pubKey)

	// Check if the recovered address matches the provided address
	isValid = recoveredAddress == address

	return isValid, recoveredAddress, nil
}

func (p *PrivateKeySigner) GetAddress() common.Address {
	return crypto.PubkeyToAddress(p.PrivateKey.PublicKey)
}
