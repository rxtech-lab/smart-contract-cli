package signer

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/contract/transport"
	"github.com/stretchr/testify/suite"
)

const (
	// Anvil default test account
	testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	testAddress    = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	// Anvil endpoint for E2E tests
	testEndpoint = "http://localhost:8545"
	// Anvil chain ID
	testChainID = 31337
)

// PrivateKeySignerTestSuite defines the test suite for PrivateKeySigner
type PrivateKeySignerTestSuite struct {
	suite.Suite
	signer      Signer
	testAddress common.Address
}

// SetupSuite runs once before all tests in the suite
func (suite *PrivateKeySignerTestSuite) SetupSuite() {
	// Initialize signer with test private key
	signer, err := NewPrivateKeySigner(testPrivateKey)
	suite.Require().NoError(err, "failed to create signer")
	suite.signer = signer

	// Set expected address
	suite.testAddress = common.HexToAddress(testAddress)
}

// TestNewPrivateKeySigner tests signer creation
func (suite *PrivateKeySignerTestSuite) TestNewPrivateKeySigner() {
	tests := []struct {
		name        string
		privateKey  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid private key",
			privateKey: testPrivateKey,
			wantErr:    false,
		},
		{
			name:       "valid private key with 0x prefix",
			privateKey: "0x" + testPrivateKey,
			wantErr:    false,
		},
		{
			name:        "invalid private key - too short",
			privateKey:  "123",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "invalid private key - not hex",
			privateKey:  "not-a-valid-hex-key",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "empty private key",
			privateKey:  "",
			wantErr:     true,
			errContains: "invalid",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			signer, err := NewPrivateKeySigner(tt.privateKey)

			if tt.wantErr {
				suite.Error(err, "expected error but got none")
				if tt.errContains != "" && err != nil {
					suite.Contains(err.Error(), tt.errContains, "error should contain expected text")
				}
				suite.Nil(signer, "expected nil signer on error")
				return
			}

			suite.NoError(err, "unexpected error")
			suite.NotNil(signer, "expected signer but got nil")
		})
	}
}

// TestGetAddress tests address derivation
func (suite *PrivateKeySignerTestSuite) TestGetAddress() {
	// Get address from signer
	address := suite.signer.(*PrivateKeySigner).GetAddress()

	// Verify address matches expected
	suite.Equal(suite.testAddress, address, "address should match expected test address")
	suite.T().Logf("Derived address: %s", address.Hex())
}

// TestSignMessageString tests message signing
func (suite *PrivateKeySignerTestSuite) TestSignMessageString() {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "simple message",
			message: "Hello, Ethereum!",
			wantErr: false,
		},
		{
			name:    "empty message",
			message: "",
			wantErr: false,
		},
		{
			name:    "long message",
			message: "This is a very long message that contains many characters and should still be signed correctly by the signer implementation.",
			wantErr: false,
		},
		{
			name:    "message with special characters",
			message: "Special chars: \n\t!@#$%^&*()",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			signature, err := suite.signer.SignMessageString(tt.message)

			if tt.wantErr {
				suite.Error(err, "expected error but got none")
				return
			}

			suite.NoError(err, "SignMessageString should not return error")
			suite.NotEmpty(signature, "signature should not be empty")
			suite.True(len(signature) > 2, "signature should be longer than 0x prefix")
			suite.Equal("0x", signature[:2], "signature should start with 0x")
			// Signature should be 65 bytes (130 hex chars) + 2 for 0x prefix = 132 total
			suite.Equal(132, len(signature), "signature should be 132 characters (0x + 130 hex)")
			suite.T().Logf("Message: %q", tt.message)
			suite.T().Logf("Signature: %s", signature)
		})
	}
}

// TestVerifyMessageString tests message verification with invalid cases
func (suite *PrivateKeySignerTestSuite) TestVerifyMessageString() {
	message := "Test message for verification"

	// First, sign a message to get a valid signature
	signature, err := suite.signer.SignMessageString(message)
	suite.Require().NoError(err, "failed to sign message for verification test")

	tests := []struct {
		name               string
		address            common.Address
		message            string
		signature          string
		wantValid          bool
		wantErr            bool
		checkRecoveredAddr bool
	}{
		{
			name:               "valid signature",
			address:            suite.testAddress,
			message:            message,
			signature:          signature,
			wantValid:          true,
			wantErr:            false,
			checkRecoveredAddr: true,
		},
		{
			name:      "wrong message",
			address:   suite.testAddress,
			message:   "Different message",
			signature: signature,
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "wrong address",
			address:   common.HexToAddress("0x0000000000000000000000000000000000000001"),
			message:   message,
			signature: signature,
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "invalid signature - too short",
			address:   suite.testAddress,
			message:   message,
			signature: "0x123",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "invalid signature - not hex",
			address:   suite.testAddress,
			message:   message,
			signature: "0xnothex",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "signature without 0x prefix",
			address:   suite.testAddress,
			message:   message,
			signature: signature[2:], // Remove 0x prefix
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			isValid, recoveredAddr, err := suite.signer.VerifyMessageString(tt.address, tt.message, tt.signature)

			if tt.wantErr {
				suite.Error(err, "expected error but got none")
				return
			}

			suite.NoError(err, "VerifyMessageString should not return error")
			suite.Equal(tt.wantValid, isValid, "signature validity should match expected")

			if tt.checkRecoveredAddr && tt.wantValid {
				suite.Equal(suite.testAddress, recoveredAddr, "recovered address should match signer address")
				suite.T().Logf("Recovered address: %s", recoveredAddr.Hex())
			}
		})
	}
}

// TestSignAndVerifyMessageRoundtrip tests the full sign and verify flow
func (suite *PrivateKeySignerTestSuite) TestSignAndVerifyMessageRoundtrip() {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "simple message roundtrip",
			message: "Hello, World!",
		},
		{
			name:    "complex message roundtrip",
			message: "Sign this transaction: 0x1234567890abcdef",
		},
		{
			name:    "empty message roundtrip",
			message: "",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Step 1: Sign the message
			signature, err := suite.signer.SignMessageString(tt.message)
			suite.Require().NoError(err, "failed to sign message")
			suite.T().Logf("Signed message %q with signature: %s", tt.message, signature)

			// Step 2: Verify the signature with correct address
			isValid, recoveredAddr, err := suite.signer.VerifyMessageString(suite.testAddress, tt.message, signature)
			suite.Require().NoError(err, "failed to verify message")
			suite.True(isValid, "signature should be valid for correct address")
			suite.Equal(suite.testAddress, recoveredAddr, "recovered address should match signer address")
			suite.T().Logf("Verified signature successfully, recovered address: %s", recoveredAddr.Hex())

			// Step 3: Verify the signature with wrong address (should fail)
			wrongAddress := common.HexToAddress("0x0000000000000000000000000000000000000001")
			isValid, _, err = suite.signer.VerifyMessageString(wrongAddress, tt.message, signature)
			suite.NoError(err, "verify should not error even with wrong address")
			suite.False(isValid, "signature should be invalid for wrong address")
		})
	}
}

// TestVerifyMessageWithMetaMaskFormat tests v=27/28 signature format
func (suite *PrivateKeySignerTestSuite) TestVerifyMessageWithMetaMaskFormat() {
	message := "Test MetaMask format"

	// Sign the message
	signature, err := suite.signer.SignMessageString(message)
	suite.Require().NoError(err, "failed to sign message")

	// Manually convert v from 0/1 to 27/28 (MetaMask format)
	sigBytes := common.Hex2Bytes(signature[2:])
	suite.Require().Equal(65, len(sigBytes), "signature should be 65 bytes")

	// Store original v value
	originalV := sigBytes[64]
	suite.T().Logf("Original v value: %d", originalV)

	// Convert to MetaMask format
	if sigBytes[64] < 27 {
		sigBytes[64] += 27
	}
	metaMaskSignature := "0x" + common.Bytes2Hex(sigBytes)
	suite.T().Logf("MetaMask format v value: %d", sigBytes[64])
	suite.T().Logf("MetaMask signature: %s", metaMaskSignature)

	// Verify that the signature with v=27/28 works correctly
	isValid, recoveredAddr, err := suite.signer.VerifyMessageString(suite.testAddress, message, metaMaskSignature)
	suite.NoError(err, "verification should not error")
	suite.True(isValid, "MetaMask format signature should be valid")
	suite.Equal(suite.testAddress, recoveredAddr, "recovered address should match")
}

// TestSignTransaction tests transaction signing
func (suite *PrivateKeySignerTestSuite) TestSignTransaction() {
	// Create a simple transaction
	nonce := uint64(0)
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasFeeCap := big.NewInt(30000000000) // 30 gwei
	gasTipCap := big.NewInt(2000000000)  // 2 gwei
	chainID := big.NewInt(testChainID)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     amount,
		Data:      nil,
	})

	// Sign the transaction
	signedTx, err := suite.signer.SignTransaction(tx)
	suite.NoError(err, "SignTransaction should not return error")
	suite.NotNil(signedTx, "signed transaction should not be nil")

	// Verify the transaction has a signature
	v, r, s := signedTx.RawSignatureValues()
	suite.NotNil(v, "v value should not be nil")
	suite.NotNil(r, "r value should not be nil")
	suite.NotNil(s, "s value should not be nil")
	suite.True(r.Sign() > 0, "r should be positive")
	suite.True(s.Sign() > 0, "s should be positive")

	suite.T().Logf("Transaction signed successfully")
	suite.T().Logf("Transaction hash: %s", signedTx.Hash().Hex())
	suite.T().Logf("V: %s, R: %s, S: %s", v.String(), r.String(), s.String())
}

// TestSignTransactionAndSendToE2E is an E2E integration test
func (suite *PrivateKeySignerTestSuite) TestSignTransactionAndSendToE2E() {
	// Create transport
	transport, err := transport.NewHttpTransport(testEndpoint)
	if err != nil {
		suite.T().Skipf("Anvil network not running: %v (run 'make e2e-network' first)", err)
		return
	}

	// Get current nonce
	nonce, err := transport.GetTransactionCount(suite.testAddress)
	suite.Require().NoError(err, "failed to get nonce")
	suite.T().Logf("Current nonce: %d", nonce)

	// Get current balance
	balanceBefore, err := transport.GetBalance(suite.testAddress)
	suite.Require().NoError(err, "failed to get balance")
	suite.T().Logf("Balance before: %s wei", balanceBefore.String())

	// Create a transaction to send ETH to another address
	// Use Anvil's second default account as recipient
	to := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	amount := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasFeeCap := big.NewInt(30000000000) // 30 gwei
	gasTipCap := big.NewInt(2000000000)  // 2 gwei
	chainID := big.NewInt(testChainID)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     amount,
		Data:      nil,
	})

	suite.T().Logf("Created transaction to send %s wei to %s", amount.String(), to.Hex())

	// Sign the transaction
	signedTx, err := suite.signer.SignTransaction(tx)
	suite.Require().NoError(err, "failed to sign transaction")
	suite.T().Logf("Transaction signed, hash: %s", signedTx.Hash().Hex())

	// Send the transaction
	txHash, err := transport.SendTransaction(signedTx)
	suite.Require().NoError(err, "failed to send transaction")
	suite.Equal(signedTx.Hash(), txHash, "transaction hash should match")
	suite.T().Logf("Transaction sent: %s", txHash.Hex())

	// Wait for transaction receipt
	receipt, err := transport.WaitForTransactionReceipt(txHash)
	suite.Require().NoError(err, "failed to get transaction receipt")
	suite.NotNil(receipt, "receipt should not be nil")
	suite.T().Logf("Transaction mined in block: %d", receipt.BlockNumber.Uint64())
	suite.T().Logf("Gas used: %d", receipt.GasUsed)
	suite.T().Logf("Status: %d (1=success, 0=failure)", receipt.Status)

	// Verify transaction succeeded
	suite.Equal(uint64(1), receipt.Status, "transaction should succeed")

	// Verify balance decreased
	balanceAfter, err := transport.GetBalance(suite.testAddress)
	suite.Require().NoError(err, "failed to get balance after transaction")
	suite.T().Logf("Balance after: %s wei", balanceAfter.String())

	// Balance should decrease by amount + gas costs
	expectedMaxDecrease := new(big.Int).Add(amount, new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), gasFeeCap))
	actualDecrease := new(big.Int).Sub(balanceBefore, balanceAfter)
	suite.T().Logf("Actual balance decrease: %s wei", actualDecrease.String())
	suite.T().Logf("Expected max decrease: %s wei", expectedMaxDecrease.String())

	// Actual decrease should be at least the amount sent
	suite.GreaterOrEqual(actualDecrease.Cmp(amount), 0, "balance should decrease by at least the sent amount")
	// And should not exceed amount + max gas cost
	suite.LessOrEqual(actualDecrease.Cmp(expectedMaxDecrease), 0, "balance should not decrease more than amount + gas")
}

// TestPrivateKeySignerTestSuite runs the test suite
func TestPrivateKeySignerTestSuite(t *testing.T) {
	suite.Run(t, new(PrivateKeySignerTestSuite))
}
