package transport

import (
	"context"
	goerrors "errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	customabi "github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
)

type HTTPTransport struct {
	Endpoint string
	client   *ethclient.Client
	timeout  time.Duration
}

func NewHTTPTransport(endpoint string, timeout time.Duration) (Transport, error) {
	if endpoint == "" {
		return nil, errors.NewTransportError(errors.ErrCodeEndpointRequired, "endpoint is required")
	}

	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, errors.WrapTransportError(err, errors.ErrCodeConnectionFailed, "failed to dial endpoint")
	}

	// Verify connectivity by attempting to get the chain ID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.ChainID(ctx)
	if err != nil {
		return nil, errors.WrapTransportError(err, errors.ErrCodeConnectionFailed, "failed to connect to endpoint")
	}

	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPTransport{
		Endpoint: endpoint,
		client:   client,
		timeout:  timeout,
	}, nil
}

// convertToEthereumABI converts custom ABI to go-ethereum's ABI.
func convertToEthereumABI(customABI customabi.ABI) (abi.ABI, error) {
	// Marshal the custom ABI back to JSON
	abiJSON, err := customABI.MarshalJSON()
	if err != nil {
		return abi.ABI{}, errors.WrapABIError(err, errors.ErrCodeABIMarshalFailed, "failed to marshal custom ABI")
	}

	// Parse it using go-ethereum's ABI parser
	ethABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return abi.ABI{}, errors.WrapABIError(err, errors.ErrCodeABIParseFailed, "failed to parse ABI")
	}

	return ethABI, nil
}

// CallContract implements Transport.
func (h *HTTPTransport) CallContract(contractAddress common.Address, customABI customabi.ABI, functionName string, args ...any) (result []byte, err error) {
	// Convert custom ABI to go-ethereum ABI
	ethABI, err := convertToEthereumABI(customABI)
	if err != nil {
		return nil, err
	}

	// Pack the function call data
	data, err := ethABI.Pack(functionName, args...)
	if err != nil {
		return nil, errors.WrapABIError(err, errors.ErrCodeABIPackFailed, fmt.Sprintf("failed to pack function %s", functionName))
	}

	// Create call message
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	// Call the contract
	ctx := context.Background()
	result, err = h.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, errors.WrapTransportError(err, errors.ErrCodeRPCCallFailed, fmt.Sprintf("failed to call contract function %s", functionName))
	}

	return result, nil
}

// EstimateGas implements Transport.
func (h *HTTPTransport) EstimateGas(transaction *types.Transaction) (gas uint64, err error) {
	ctx := context.Background()

	// Estimate gas for the transaction
	msg := ethereum.CallMsg{
		To:         transaction.To(),
		Gas:        transaction.Gas(),
		GasPrice:   transaction.GasPrice(),
		GasFeeCap:  transaction.GasFeeCap(),
		GasTipCap:  transaction.GasTipCap(),
		Value:      transaction.Value(),
		Data:       transaction.Data(),
		AccessList: transaction.AccessList(),
	}

	gas, err = h.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, errors.WrapTransportError(err, errors.ErrCodeGasEstimateFailed, "failed to estimate gas")
	}

	return gas, nil
}

// GetBalance implements Transport.
func (h *HTTPTransport) GetBalance(address common.Address) (balance *big.Int, err error) {
	ctx := context.Background()

	balance, err = h.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, errors.WrapTransportError(err, errors.ErrCodeBalanceQueryFailed, "failed to query balance")
	}

	return balance, nil
}

// GetTransactionCount implements Transport.
func (h *HTTPTransport) GetTransactionCount(address common.Address) (nonce uint64, err error) {
	ctx := context.Background()

	nonce, err = h.client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, errors.WrapTransportError(err, errors.ErrCodeNonceQueryFailed, "failed to query nonce")
	}

	return nonce, nil
}

// SendTransaction implements Transport.
func (h *HTTPTransport) SendTransaction(transaction *types.Transaction) (txHash common.Hash, err error) {
	ctx := context.Background()

	err = h.client.SendTransaction(ctx, transaction)
	if err != nil {
		return common.Hash{}, errors.WrapTransportError(err, errors.ErrCodeTransactionSendFailed, "failed to send transaction")
	}

	return transaction.Hash(), nil
}

// WaitForTransactionReceipt implements Transport.
func (h *HTTPTransport) WaitForTransactionReceipt(txHash common.Hash) (receipt *types.Receipt, err error) {
	ctx := context.Background()

	// Poll for the transaction receipt
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(h.timeout)

	for {
		select {
		case <-timeout:
			return nil, errors.NewTransportError(errors.ErrCodeTransactionTimeout, "timeout waiting for transaction receipt")
		case <-ticker.C:
			receipt, err = h.client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// If error is "not found", continue polling
			if goerrors.Is(err, ethereum.NotFound) {
				continue
			}
			// For any other error, return it
			return nil, errors.WrapTransportError(err, errors.ErrCodeReceiptQueryFailed, "failed to query transaction receipt")
		}
	}
}

// GetChainID implements Transport.
func (h *HTTPTransport) GetChainID() (chainID *big.Int, err error) {
	ctx := context.Background()

	chainID, err = h.client.ChainID(ctx)
	if err != nil {
		return nil, errors.WrapTransportError(err, errors.ErrCodeChainIDQueryFailed, "failed to query chain ID")
	}

	return chainID, nil
}
