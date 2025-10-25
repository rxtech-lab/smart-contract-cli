package transport

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
	"github.com/stretchr/testify/suite"
)

const (
	// Anvil default endpoint
	testEndpoint = "http://localhost:8545"
	// Anvil test account with pre-funded balance
	testAddress = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

// Simple ERC20-like ABI for testing
var testABI = `[
	{
		"type": "function",
		"name": "balanceOf",
		"inputs": [
			{
				"name": "account",
				"type": "address"
			}
		],
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view"
	},
	{
		"type": "function",
		"name": "totalSupply",
		"inputs": [],
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"stateMutability": "view"
	}
]`

// HttpTransportTestSuite defines the test suite for HTTP transport
type HttpTransportTestSuite struct {
	suite.Suite
	transport       Transport
	testABIObj      abi.ABI
	testAddr        common.Address
	contractAddress common.Address
}

// SetupSuite runs once before all tests in the suite
func (suite *HttpTransportTestSuite) SetupSuite() {
	// Parse test ABI
	err := json.Unmarshal([]byte(testABI), &suite.testABIObj)
	suite.Require().NoError(err, "failed to parse test ABI")

	// Set up test addresses
	suite.testAddr = common.HexToAddress(testAddress)
	suite.contractAddress = common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Initialize transport
	transport, err := NewHttpTransport(testEndpoint)
	if err != nil {
		suite.T().Skipf("Anvil network not running: %v (run 'make e2e-network' first)", err)
	}
	suite.transport = transport
}

// TearDownSuite runs once after all tests in the suite
func (suite *HttpTransportTestSuite) TearDownSuite() {
	// Cleanup if needed
}

// SetupTest runs before each test
func (suite *HttpTransportTestSuite) SetupTest() {
	// Per-test setup if needed
}

// TearDownTest runs after each test
func (suite *HttpTransportTestSuite) TearDownTest() {
	// Per-test cleanup if needed
}

// TestNewHttpTransport tests transport initialization
func (suite *HttpTransportTestSuite) TestNewHttpTransport() {
	tests := []struct {
		name        string
		endpoint    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid endpoint",
			endpoint: testEndpoint,
			wantErr:  false,
		},
		{
			name:        "empty endpoint",
			endpoint:    "",
			wantErr:     true,
			errContains: "endpoint is required",
		},
		{
			name:     "invalid endpoint returns error",
			endpoint: "http://invalid:9999",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			transport, err := NewHttpTransport(tt.endpoint)

			if tt.wantErr {
				suite.Error(err, "expected error but got none")
				if tt.errContains != "" && err != nil {
					suite.Contains(err.Error(), tt.errContains, "error should contain expected text")
				}
				return
			}

			if tt.endpoint == testEndpoint && err != nil {
				suite.T().Skipf("Anvil network not running: %v", err)
			}

			suite.NoError(err, "unexpected error")
			suite.NotNil(transport, "expected transport but got nil")
		})
	}
}

// TestGetBalance tests balance retrieval
func (suite *HttpTransportTestSuite) TestGetBalance() {
	tests := []struct {
		name          string
		address       common.Address
		wantErr       bool
		checkMinimum  bool
		minimumAmount *big.Int
	}{
		{
			name:          "get balance of test account",
			address:       suite.testAddr,
			wantErr:       false,
			checkMinimum:  true,
			minimumAmount: new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18)),
		},
		{
			name:         "get balance of zero address",
			address:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			wantErr:      false,
			checkMinimum: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			balance, err := suite.transport.GetBalance(tt.address)

			if tt.wantErr {
				suite.Error(err, "expected error but got none")
				return
			}

			suite.NoError(err, "GetBalance should not return error")
			suite.NotNil(balance, "expected balance but got nil")

			// Check minimum balance for test account
			if tt.checkMinimum {
				suite.GreaterOrEqual(
					balance.Cmp(tt.minimumAmount),
					0,
					"balance should be >= %v, got %v", tt.minimumAmount, balance,
				)
				suite.T().Logf("Account %s has balance: %v wei", tt.address.Hex(), balance)
			}
		})
	}
}

// TestGetTransactionCount tests nonce retrieval
func (suite *HttpTransportTestSuite) TestGetTransactionCount() {
	tests := []struct {
		name    string
		address common.Address
		wantErr bool
	}{
		{
			name:    "get nonce of test account",
			address: suite.testAddr,
			wantErr: false,
		},
		{
			name:    "get nonce of zero address",
			address: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			nonce, err := suite.transport.GetTransactionCount(tt.address)

			if tt.wantErr {
				suite.Error(err, "expected error but got none")
				return
			}

			suite.NoError(err, "GetTransactionCount should not return error")
			suite.T().Logf("Account %s has nonce: %d", tt.address.Hex(), nonce)
		})
	}
}

// TestCallContract tests contract call functionality
func (suite *HttpTransportTestSuite) TestCallContract() {
	suite.Run("call non-existent contract", func() {
		// Calling a non-existent contract should either return empty data or an error
		// depending on the RPC implementation
		result, err := suite.transport.CallContract(
			suite.contractAddress,
			suite.testABIObj,
			"totalSupply",
		)

		// Both outcomes are acceptable for a non-existent contract
		if err != nil {
			suite.T().Logf("Expected error calling non-existent contract: %v", err)
		} else {
			suite.T().Logf("Call returned result (may be empty): %x", result)
		}
	})
}

// TestEstimateGas tests gas estimation
func (suite *HttpTransportTestSuite) TestEstimateGas() {
	suite.T().Skip("Gas estimation requires a properly signed transaction - skipping in basic e2e test")
}

// TestSequentialOperations tests multiple operations in sequence
func (suite *HttpTransportTestSuite) TestSequentialOperations() {
	// Get balance
	balance, err := suite.transport.GetBalance(suite.testAddr)
	suite.NoError(err, "GetBalance should not return error")
	suite.T().Logf("Balance: %v wei", balance)

	// Get nonce
	nonce, err := suite.transport.GetTransactionCount(suite.testAddr)
	suite.NoError(err, "GetTransactionCount should not return error")
	suite.T().Logf("Nonce: %d", nonce)

	// Verify balance is substantial (Anvil pre-funds accounts)
	expectedMin := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))
	suite.GreaterOrEqual(
		balance.Cmp(expectedMin),
		0,
		"balance should be >= %v, got %v", expectedMin, balance,
	)
}

// TestConcurrentOperations tests multiple operations running concurrently
func (suite *HttpTransportTestSuite) TestConcurrentOperations() {
	done := make(chan error, 2)

	// Get balance concurrently
	go func() {
		_, err := suite.transport.GetBalance(suite.testAddr)
		done <- err
	}()

	// Get nonce concurrently
	go func() {
		_, err := suite.transport.GetTransactionCount(suite.testAddr)
		done <- err
	}()

	// Wait for both operations with timeout
	timeout := time.After(10 * time.Second)
	for range 2 {
		select {
		case err := <-done:
			suite.NoError(err, "concurrent operation should not return error")
		case <-timeout:
			suite.Fail("timeout waiting for concurrent operations")
			return
		}
	}
}

// TestHttpTransportTestSuite runs the test suite
func TestHttpTransportTestSuite(t *testing.T) {
	suite.Run(t, new(HttpTransportTestSuite))
}
