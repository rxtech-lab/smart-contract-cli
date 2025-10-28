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

// WithTransport creates a new PrivateKeySignerWithTransport with the given transport
func (p *PrivateKeySigner) WithTransport(transport transport.Transport) SignerWithTransport {
	return &PrivateKeySignerWithTransport{
		PrivateKeySigner: p,
		transport:        transport,
	}
}

// Helper function to find method in ABI
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

// Helper function to convert custom ABI to go-ethereum ABI
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

// CallContractMethod implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) CallContractMethod(contractAddress common.Address, contractABI abi.ABI, methodName string, value *big.Int, gasLimit uint64, gasPrice *big.Int, args ...any) (result []any, err error) {
	// Find method in ABI
	method := findMethodInABI(contractABI, methodName)
	if method == nil {
		return nil, errors.NewABIError(errors.ErrCodeMethodNotFound, fmt.Sprintf("method %s not found in ABI", methodName))
	}

	// Check if it's a read-only operation using enum
	if method.IsReadOnly() {
		// Read operation - use transport.CallContract
		rawResult, err := p.transport.CallContract(contractAddress, contractABI, methodName, args...)
		if err != nil {
			return nil, err
		}

		// Decode the result
		ethABI, err := convertToEthereumABI(contractABI)
		if err != nil {
			return nil, err
		}

		// If no outputs, return nil
		if len(method.Outputs) == 0 {
			return nil, nil
		}

		// Unpack the result
		results := make([]any, len(method.Outputs))
		err = ethABI.UnpackIntoInterface(&results, methodName, rawResult)
		if err != nil {
			return nil, errors.WrapABIError(err, errors.ErrCodeABIUnpackFailed, fmt.Sprintf("failed to unpack result for method %s", methodName))
		}

		return results, nil
	}

	// Write operation - create, sign, and send transaction
	ethABI, err := convertToEthereumABI(contractABI)
	if err != nil {
		return nil, err
	}

	// Pack the function data
	data, err := ethABI.Pack(methodName, args...)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIPackFailed, fmt.Sprintf("failed to pack function %s", methodName))
	}

	// Get nonce
	signerAddress := p.PrivateKeySigner.GetAddress()
	nonce, err := p.transport.GetTransactionCount(signerAddress)
	if err != nil {
		return nil, err
	}

	// Set default value if nil
	if value == nil {
		value = big.NewInt(0)
	}

	// Set default gas price if nil
	if gasPrice == nil {
		gasPrice = big.NewInt(1000000000) // 1 gwei default
	}

	// Create transaction
	var tx *types.Transaction
	if gasLimit == 0 {
		// Need to estimate gas
		tempTx := types.NewTransaction(nonce, contractAddress, value, 1000000, gasPrice, data)
		estimatedGas, err := p.transport.EstimateGas(tempTx)
		if err != nil {
			return nil, err
		}
		gasLimit = estimatedGas
	}

	tx = types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, data)

	// Sign the transaction
	signedTx, err := p.PrivateKeySigner.SignTransaction(tx)
	if err != nil {
		return nil, err
	}

	// Send the transaction
	txHash, err := p.transport.SendTransaction(signedTx)
	if err != nil {
		return nil, err
	}

	// Wait for transaction receipt
	receipt, err := p.transport.WaitForTransactionReceipt(txHash)
	if err != nil {
		return nil, err
	}

	// Return status and transaction hash
	return []any{receipt.Status, txHash.Hex()}, nil
}

// EstimateGas implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) EstimateGas(tx *types.Transaction) (gas uint64, err error) {
	return p.transport.EstimateGas(tx)
}

// GetAddress implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) GetAddress() (address common.Address, err error) {
	return p.PrivateKeySigner.GetAddress(), nil
}

// GetBalance implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) GetBalance(address common.Address) (balance *big.Int, err error) {
	return p.transport.GetBalance(address)
}

// GetTransactionCount implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) GetTransactionCount(address common.Address) (nonce uint64, err error) {
	return p.transport.GetTransactionCount(address)
}

// SendTransaction implements SignerWithTransport.
func (p *PrivateKeySignerWithTransport) SendTransaction(tx *types.Transaction) (txHash common.Hash, err error) {
	// Sign the transaction first
	signedTx, err := p.PrivateKeySigner.SignTransaction(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// Send the signed transaction
	return p.transport.SendTransaction(signedTx)
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
	return p.transport.WaitForTransactionReceipt(txHash)
}
