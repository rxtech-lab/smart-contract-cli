package transport

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
)

type Transport interface {
	// SendTransaction sends a transaction and returns the transaction hash
	SendTransaction(tx *types.Transaction) (txHash common.Hash, err error)

	// WaitForTransactionReceipt waits for a transaction receipt and returns it
	WaitForTransactionReceipt(txHash common.Hash) (receipt *types.Receipt, err error)

	// CallContract calls a contract function and returns the result
	CallContract(contractAddress common.Address, abi abi.ABI, functionName string, args ...any) (result []byte, err error)

	// EstimateGas estimates the gas required for a transaction
	EstimateGas(tx *types.Transaction) (gas uint64, err error)

	// GetTransactionCount gets the nonce for an address
	GetTransactionCount(address common.Address) (nonce uint64, err error)

	// GetBalance gets the balance of an address
	GetBalance(address common.Address) (balance *big.Int, err error)

	// GetChainID gets the chain ID from the blockchain
	GetChainID() (chainID *big.Int, err error)
}
