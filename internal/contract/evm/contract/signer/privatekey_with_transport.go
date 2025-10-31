package signer

import (
	"fmt"
	"math/big"
	"strings"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/contract/transport"
	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
)

type PrivateKeySignerWithTransport struct {
	*PrivateKeySigner
	transport transport.Transport
}

// WithTransport creates a new PrivateKeySignerWithTransport with the given transport.
func (p *PrivateKeySigner) WithTransport(transport transport.Transport) SignerWithTransport {
	return &PrivateKeySignerWithTransport{
		PrivateKeySigner: p,
		transport:        transport,
	}
}

// Helper function to find method in ABI.
func findMethodInABI(customABI abi.ABI, methodName string) *abi.ABIElement {
	elements := customABI.Elements()
	for i := range elements {
		elem := &elements[i]
		if elem.Type == "function" && elem.Name == methodName {
			return elem
		}
	}
	return nil
}

// Helper function to convert custom ABI to go-ethereum ABI.
func convertToEthereumABI(customABI abi.ABI) (ethabi.ABI, error) {
	abiJSON, err := customABI.MarshalJSON()
	if err != nil {
		return ethabi.ABI{}, errors.WrapABIError(err, errors.ErrCodeABIMarshalFailed, "failed to marshal custom ABI")
	}

	ethABI, err := ethabi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return ethabi.ABI{}, errors.WrapABIError(err, errors.ErrCodeABIParseFailed, "failed to parse ABI")
	}

	return ethABI, nil
}

// executeReadOnlyCall handles read-only contract method calls.
func (p *PrivateKeySignerWithTransport) executeReadOnlyCall(contractAddress common.Address, contractABI abi.ABI, method *abi.ABIElement, methodName string, args ...any) ([]any, error) {
	// Call the contract using transport
	rawResult, err := p.transport.CallContract(contractAddress, contractABI, methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract method %s: %w", methodName, err)
	}

	// If no outputs, return nil
	if len(method.Outputs) == 0 {
		return nil, nil
	}

	// Decode the result
	return p.decodeCallResult(contractABI, methodName, rawResult)
}

// decodeCallResult unpacks raw contract call result into typed values.
func (p *PrivateKeySignerWithTransport) decodeCallResult(contractABI abi.ABI, methodName string, rawResult []byte) ([]any, error) {
	ethABI, err := convertToEthereumABI(contractABI)
	if err != nil {
		return nil, err
	}

	// Use Unpack which returns []interface{} directly
	results, err := ethABI.Unpack(methodName, rawResult)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIUnpackFailed, fmt.Sprintf("failed to unpack result for method %s", methodName))
	}

	return results, nil
}

// packFunctionData encodes function call data with arguments.
func (p *PrivateKeySignerWithTransport) packFunctionData(contractABI abi.ABI, methodName string, args ...any) ([]byte, error) {
	ethABI, err := convertToEthereumABI(contractABI)
	if err != nil {
		return nil, err
	}

	data, err := ethABI.Pack(methodName, args...)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIPackFailed, fmt.Sprintf("failed to pack function %s", methodName))
	}

	return data, nil
}

// setDefaultTransactionParams sets default values for value and gasPrice if not provided.
func setDefaultTransactionParams(value, gasPrice **big.Int) {
	if *value == nil {
		*value = big.NewInt(0)
	}

	if *gasPrice == nil {
		*gasPrice = big.NewInt(1000000000) // 1 gwei default
	}
}

// buildTransaction creates a transaction with gas estimation if needed.
func (p *PrivateKeySignerWithTransport) buildTransaction(contractAddress common.Address, nonce uint64, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (*types.Transaction, error) {
	// Get chain ID from transport
	chainID, err := p.transport.GetChainID()
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Use EIP-1559 transaction for better compatibility
	gasTipCap := gasPrice
	gasFeeCap := new(big.Int).Mul(gasPrice, big.NewInt(2)) // 2x gasPrice for max fee

	if gasLimit == 0 {
		// Estimate gas using a signed transaction so it has a valid 'from' address
		// Use a reasonable default gas limit for estimation (not too high to avoid balance issues)
		tempTx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       100000, // reasonable gas limit for estimation
			To:        &contractAddress,
			Value:     value,
			Data:      data,
		})

		// Sign the temp transaction so it has a valid 'from' address
		signedTempTx, err := p.PrivateKeySigner.SignTransaction(tempTx)
		if err != nil {
			return nil, err
		}

		estimatedGas, err := p.transport.EstimateGas(signedTempTx)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
		// Add 50% buffer to gas estimate to avoid out-of-gas errors
		// Gas estimation can be inaccurate, especially for complex contracts
		gasLimit = estimatedGas + (estimatedGas / 2)
	}

	transaction := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &contractAddress,
		Value:     value,
		Data:      data,
	})
	return transaction, nil
}

// executeWriteTransaction signs and sends a transaction, then waits for receipt.
func (p *PrivateKeySignerWithTransport) executeWriteTransaction(tx *types.Transaction) ([]any, error) {
	// Sign the transaction
	signedTx, err := p.PrivateKeySigner.SignTransaction(tx)
	if err != nil {
		return nil, err
	}

	// Send the transaction
	txHash, err := p.transport.SendTransaction(signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Wait for transaction receipt
	receipt, err := p.transport.WaitForTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}

	// Return status and transaction hash
	return []any{receipt.Status, txHash.Hex()}, nil
}

// CallContractMethod implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) CallContractMethod(contractAddress common.Address, contractABI abi.ABI, methodName string, value *big.Int, gasLimit uint64, gasPrice *big.Int, args ...any) (result []any, err error) {
	// Find method in ABI
	method := findMethodInABI(contractABI, methodName)
	if method == nil {
		return nil, errors.NewABIError(errors.ErrCodeMethodNotFound, fmt.Sprintf("method %s not found in ABI", methodName))
	}

	// Check if it's a read-only operation
	if method.IsReadOnly() {
		return p.executeReadOnlyCall(contractAddress, contractABI, method, methodName, args...)
	}

	// Write operation - pack function data
	data, err := p.packFunctionData(contractABI, methodName, args...)
	if err != nil {
		return nil, err
	}

	// Get nonce for transaction
	signerAddress := p.PrivateKeySigner.GetAddress()
	nonce, err := p.transport.GetTransactionCount(signerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction count: %w", err)
	}

	// Set default parameters
	setDefaultTransactionParams(&value, &gasPrice)

	// Build transaction with gas estimation
	transaction, err := p.buildTransaction(contractAddress, nonce, value, gasLimit, gasPrice, data)
	if err != nil {
		return nil, err
	}

	// Execute the transaction
	return p.executeWriteTransaction(transaction)
}

// EstimateGas implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) EstimateGas(tx *types.Transaction) (gas uint64, err error) {
	gas, err = p.transport.EstimateGas(tx)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}
	return gas, nil
}

// GetAddress implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) GetAddress() (address common.Address, err error) {
	return p.PrivateKeySigner.GetAddress(), nil
}

// GetBalance implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) GetBalance(address common.Address) (balance *big.Int, err error) {
	balance, err = p.transport.GetBalance(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// GetTransactionCount implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) GetTransactionCount(address common.Address) (nonce uint64, err error) {
	nonce, err = p.transport.GetTransactionCount(address)
	if err != nil {
		return 0, fmt.Errorf("failed to get transaction count: %w", err)
	}
	return nonce, nil
}

// SendTransaction implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) SendTransaction(tx *types.Transaction) (txHash common.Hash, err error) {
	// Sign the transaction first
	signedTx, err := p.PrivateKeySigner.SignTransaction(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// Send the signed transaction
	txHash, err = p.transport.SendTransaction(signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}
	return txHash, nil
}

// SignMessageString implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) SignMessageString(message string) (signature string, err error) {
	return p.PrivateKeySigner.SignMessageString(message)
}

// SignTransaction implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) SignTransaction(tx *types.Transaction) (signedTx *types.Transaction, err error) {
	return p.PrivateKeySigner.SignTransaction(tx)
}

// VerifyMessageString implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) VerifyMessageString(address common.Address, message string, signature string) (isValid bool, recoveredAddress common.Address, err error) {
	return p.PrivateKeySigner.VerifyMessageString(address, message, signature)
}

// WaitForTransactionReceipt implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) WaitForTransactionReceipt(txHash common.Hash) (receipt *types.Receipt, err error) {
	receipt, err = p.transport.WaitForTransactionReceipt(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction receipt: %w", err)
	}
	return receipt, nil
}
