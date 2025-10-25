package transport

import (
	"context"
	"errors"
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
)

type HttpTransport struct {
	Endpoint string
	client   *ethclient.Client
}

func NewHttpTransport(endpoint string) (Transport, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint is required")
	}

	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	return &HttpTransport{
		Endpoint: endpoint,
		client:   client,
	}, nil
}

// convertToEthereumABI converts custom ABI to go-ethereum's ABI
func convertToEthereumABI(customABI customabi.ABI) (abi.ABI, error) {
	// Marshal the custom ABI back to JSON
	abiJSON, err := customABI.MarshalJSON()
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to marshal custom ABI: %w", err)
	}

	// Parse it using go-ethereum's ABI parser
	ethABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return ethABI, nil
}

// CallContract implements Transport.
func (h *HttpTransport) CallContract(contractAddress common.Address, customABI customabi.ABI, functionName string, args ...any) (result []byte, err error) {
	// Convert custom ABI to go-ethereum ABI
	ethABI, err := convertToEthereumABI(customABI)
	if err != nil {
		return nil, err
	}

	// Pack the function call data
	data, err := ethABI.Pack(functionName, args...)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return result, nil
}

// EstimateGas implements Transport.
func (h *HttpTransport) EstimateGas(tx *types.Transaction) (gas uint64, err error) {
	ctx := context.Background()

	// Estimate gas for the transaction
	msg := ethereum.CallMsg{
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		GasFeeCap:  tx.GasFeeCap(),
		GasTipCap:  tx.GasTipCap(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: tx.AccessList(),
	}

	gas, err = h.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}

	return gas, nil
}

// GetBalance implements Transport.
func (h *HttpTransport) GetBalance(address common.Address) (balance *big.Int, err error) {
	ctx := context.Background()

	balance, err = h.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// GetTransactionCount implements Transport.
func (h *HttpTransport) GetTransactionCount(address common.Address) (nonce uint64, err error) {
	ctx := context.Background()

	nonce, err = h.client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

// SendTransaction implements Transport.
func (h *HttpTransport) SendTransaction(tx *types.Transaction) (txHash common.Hash, err error) {
	ctx := context.Background()

	err = h.client.SendTransaction(ctx, tx)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.Hash(), nil
}

// WaitForTransactionReceipt implements Transport.
func (h *HttpTransport) WaitForTransactionReceipt(txHash common.Hash) (receipt *types.Receipt, err error) {
	ctx := context.Background()

	// Poll for the transaction receipt
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction receipt")
		case <-ticker.C:
			receipt, err = h.client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// If error is not "not found", return it
			if err != ethereum.NotFound {
				return nil, err
			}
			// Otherwise, continue polling
		}
	}
}
