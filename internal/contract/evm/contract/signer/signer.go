package signer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
)

type Signer interface {
	// SignTransaction signs a transaction and returns the signed transaction
	SignTransaction(tx *types.Transaction) (signedTx *types.Transaction, err error)
	// SignMessage signs an arbitrary message and returns the signature
	SignMessageString(message string) (signature string, err error)
	// VerifyMessageString verifies a signature against a message for a given address
	VerifyMessageString(address common.Address, message string, signature string) (isValid bool, recoveredAddress common.Address, err error)
}

type SignerWithTransport interface {
	Signer
	// CallContractMethod calls a contract method and returns the result
	// For read-only methods (view/pure), returns decoded result values
	// For write methods, returns transaction status and hash
	CallContractMethod(contractAddress common.Address, contractABI abi.ABI, methodName string, value *big.Int, gasLimit uint64, gasPrice *big.Int, args ...any) (result []any, err error)

	// EstimateGas estimates the gas required for a transaction
	EstimateGas(tx *types.Transaction) (gas uint64, err error)

	// GetTransactionCount gets the nonce for an address
	GetTransactionCount(address common.Address) (nonce uint64, err error)

	// GetBalance gets the balance of an address
	GetBalance(address common.Address) (balance *big.Int, err error)

	// SendTransaction sends a transaction and returns the transaction hash
	SendTransaction(tx *types.Transaction) (txHash common.Hash, err error)

	// WaitForTransactionReceipt waits for a transaction receipt and returns it
	WaitForTransactionReceipt(txHash common.Hash) (receipt *types.Receipt, err error)

	// GetAddress gets the address of the signer
	GetAddress() (address common.Address, err error)
}
