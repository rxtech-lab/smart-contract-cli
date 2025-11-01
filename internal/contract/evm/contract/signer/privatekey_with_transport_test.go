package signer

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/abi"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/contract/transport"
	solc "github.com/rxtech-lab/solc-go"
	"github.com/stretchr/testify/suite"
)

const testContractSource = `
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TestContract {
    uint256 public value;
    address public lastSender;

    event Deposited(address indexed sender, uint256 amount, uint256 newValue);
    event ValueChanged(address indexed sender, uint256 oldValue, uint256 newValue);

    // Payable function - accepts ETH and updates state
    function deposit(uint256 amount) public payable returns (uint256) {
        value += amount;
        lastSender = msg.sender;
        emit Deposited(msg.sender, amount, value);
        return value;
    }

    // View function - multiple returns
    function getInfo() public view returns (uint256, address) {
        return (value, lastSender);
    }

    // Pure function
    function add(uint256 a, uint256 b) public pure returns (uint256) {
        return a + b;
    }

    // Non-payable write function
    function setValue(uint256 newValue) public returns (uint256) {
        uint256 oldValue = value;
        value = newValue;
        lastSender = msg.sender;
        emit ValueChanged(msg.sender, oldValue, newValue);
        return oldValue;
    }

    // Getter for value
    function getValue() public view returns (uint256) {
        return value;
    }
}
`

// PrivateKeySignerWithTransportTestSuite is the test suite.
type PrivateKeySignerWithTransportTestSuite struct {
	suite.Suite
	signer          SignerWithTransport
	transport       transport.Transport
	contractAddress common.Address
	contractABI     abi.ABI
	testPrivateKey  string
	testAddress     common.Address
	chainID         *big.Int
}

// SetupSuite runs once before all tests.
func (suite *PrivateKeySignerWithTransportTestSuite) SetupSuite() {
	// Anvil test account private key
	suite.testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	suite.testAddress = common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")

	// Create transport
	tr, err := transport.NewHTTPTransport("http://localhost:8545", 5*time.Second)
	suite.Require().NoError(err, "Failed to create transport")
	suite.transport = tr

	// Get chain ID from blockchain
	suite.chainID, err = suite.transport.GetChainID()
	suite.Require().NoError(err, "Failed to get chain ID")
	suite.T().Logf("Using chain ID: %s", suite.chainID.String())

	// Create signer
	baseSigner, err := NewPrivateKeySigner(suite.testPrivateKey)
	suite.Require().NoError(err, "Failed to create signer")

	pkSigner, isValid := baseSigner.(*PrivateKeySigner)
	suite.Require().True(isValid, "Failed to cast to PrivateKeySigner")

	suite.signer = pkSigner.WithTransport(suite.transport)

	// Compile contract using solc-go
	compiler, err := solc.NewWithVersion("0.8.20")
	suite.Require().NoError(err, "Failed to create compiler")
	defer func() {
		if closeErr := compiler.Close(); closeErr != nil {
			// Log error but don't fail the test
			_ = closeErr
		}
	}()

	// Create input for compilation
	input := &solc.Input{
		Language: "Solidity",
		Sources: map[string]solc.SourceIn{
			"TestContract.sol": {
				Content: testContractSource,
			},
		},
		Settings: solc.Settings{
			Optimizer: solc.Optimizer{
				Enabled: false,
			},
			OutputSelection: map[string]map[string][]string{
				"*": {
					"*": []string{"abi", "evm.bytecode"},
				},
			},
		},
	}

	result, err := compiler.CompileWithOptions(input, nil)
	suite.Require().NoError(err, "Failed to compile contract")
	suite.Require().NotNil(result, "Compilation result is nil")
	suite.Require().Empty(result.Errors, "Compilation has errors")

	// Get contract bytecode and ABI
	contracts := result.Contracts
	suite.Require().NotEmpty(contracts, "No contracts found in compilation result")

	// Get the TestContract
	sourceContracts, ok := contracts["TestContract.sol"]
	suite.Require().True(ok, "TestContract.sol not found in contracts")

	contract, ok := sourceContracts["TestContract"]
	suite.Require().True(ok, "TestContract not found")

	bytecode := contract.EVM.Bytecode.Object

	// contract.ABI is []json.RawMessage, we need to marshal it to string
	abiBytes, err := json.Marshal(contract.ABI)
	suite.Require().NoError(err, "Failed to marshal ABI")
	abiJSON := string(abiBytes)

	suite.Require().NotEmpty(bytecode, "Bytecode is empty")
	suite.Require().NotEmpty(abiJSON, "ABI is empty")

	// Parse ABI
	parsedABI, err := abi.ParseAbi(abiJSON)
	suite.Require().NoError(err, "Failed to parse ABI")

	// Create ABI wrapper
	customABI := &abi.ABI{}
	customABI.SetElements(abi.ABIArray(parsedABI))
	suite.contractABI = *customABI

	// Deploy contract
	suite.deployContract(bytecode)
}

// deployContract deploys the test contract.
func (suite *PrivateKeySignerWithTransportTestSuite) deployContract(bytecode string) {
	// Get nonce
	nonce, err := suite.transport.GetTransactionCount(suite.testAddress)
	suite.Require().NoError(err, "Failed to get nonce")

	// Create deployment transaction
	deployData := common.FromHex(bytecode)

	// Use EIP-1559 transaction
	transaction := types.NewTx(&types.DynamicFeeTx{
		ChainID:   suite.chainID,
		Nonce:     nonce,
		GasTipCap: big.NewInt(1000000000), // 1 gwei
		GasFeeCap: big.NewInt(2000000000), // 2 gwei
		Gas:       3000000,                // gas limit
		To:        nil,                    // contract creation
		Value:     big.NewInt(0),
		Data:      deployData,
	})

	// Send transaction
	txHash, err := suite.signer.SendTransaction(transaction)
	suite.Require().NoError(err, "Failed to send deployment transaction")
	suite.T().Logf("Deployment transaction sent: %s", txHash.Hex())

	// Wait for receipt
	receipt, err := suite.transport.WaitForTransactionReceipt(txHash)
	suite.Require().NoError(err, "Failed to get deployment receipt")
	suite.Require().Equal(uint64(1), receipt.Status, "Deployment transaction failed")

	suite.contractAddress = receipt.ContractAddress
	suite.Require().NotEqual(common.Address{}, suite.contractAddress, "Contract address is empty")
}

// TestGetAddress tests the GetAddress method.
func (suite *PrivateKeySignerWithTransportTestSuite) TestGetAddress() {
	address, err := suite.signer.GetAddress()
	suite.Require().NoError(err)
	suite.Assert().Equal(suite.testAddress, address)
}

// TestGetBalance tests the GetBalance method.
func (suite *PrivateKeySignerWithTransportTestSuite) TestGetBalance() {
	balance, err := suite.signer.GetBalance(suite.testAddress)
	suite.Require().NoError(err)
	suite.Assert().NotNil(balance)
	// Anvil starts with ~10000 ETH
	suite.Assert().True(balance.Cmp(big.NewInt(0)) > 0)
}

// TestGetTransactionCount tests nonce retrieval.
func (suite *PrivateKeySignerWithTransportTestSuite) TestGetTransactionCount() {
	nonce, err := suite.signer.GetTransactionCount(suite.testAddress)
	suite.Require().NoError(err)
	suite.Assert().True(nonce > 0, "Nonce should be > 0 after deployment")
}

// TestCallContractMethod_PureFunction tests calling a pure function.
func (suite *PrivateKeySignerWithTransportTestSuite) TestCallContractMethod_PureFunction() {
	// Call add(10, 20)
	result, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"add",
		nil, // no value
		0,   // auto gas
		nil, // default gas price
		big.NewInt(10),
		big.NewInt(20),
	)

	suite.Require().NoError(err, "Failed to call pure function")
	suite.Require().Len(result, 1, "Expected 1 return value")

	// Convert result to big.Int
	resultBigInt, ok := result[0].(*big.Int)
	suite.Require().True(ok, "Result should be *big.Int")
	suite.Assert().Equal(int64(30), resultBigInt.Int64())

	// Verify the method is pure using enum
	method := findMethodInABI(suite.contractABI, "add")
	suite.Require().NotNil(method)
	suite.Assert().Equal(abi.StateMutabilityPure, method.GetStateMutability())
	suite.Assert().True(method.IsReadOnly())
}

// TestCallContractMethod_ViewFunction tests calling a view function.
func (suite *PrivateKeySignerWithTransportTestSuite) TestCallContractMethod_ViewFunction() {
	// First set a value so we have something to read
	_, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"setValue",
		nil,
		0,
		nil,
		big.NewInt(42),
	)
	suite.Require().NoError(err, "Failed to set value")

	// Now call getInfo()
	result, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"getInfo",
		nil,
		0,
		nil,
	)

	suite.Require().NoError(err, "Failed to call view function")
	suite.Require().Len(result, 2, "Expected 2 return values")

	// Check value
	valueBigInt, isOk := result[0].(*big.Int)
	suite.Require().True(isOk, "First result should be *big.Int")
	suite.Assert().Equal(int64(42), valueBigInt.Int64())

	// Check address
	addr, ok := result[1].(common.Address)
	suite.Require().True(ok, "Second result should be common.Address")
	suite.Assert().Equal(suite.testAddress, addr)

	// Verify the method is view using enum
	method := findMethodInABI(suite.contractABI, "getInfo")
	suite.Require().NotNil(method)
	suite.Assert().Equal(abi.StateMutabilityView, method.GetStateMutability())
	suite.Assert().True(method.IsReadOnly())
}

// TestCallContractMethod_NonPayableWrite tests a non-payable write function.
func (suite *PrivateKeySignerWithTransportTestSuite) TestCallContractMethod_NonPayableWrite() {
	// Call setValue(123)
	result, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"setValue",
		nil, // no ETH value
		0,   // auto gas estimation
		nil, // default gas price
		big.NewInt(123),
	)

	suite.Require().NoError(err, "Failed to call write function")
	suite.Require().Len(result, 2, "Expected 2 return values (status, txHash)")

	// Check transaction succeeded
	status, statusOk := result[0].(uint64)
	suite.Require().True(statusOk, "First result should be uint64 status")

	// Check tx hash
	txHash, hashOk := result[1].(string)
	suite.Require().True(hashOk, "Second result should be string tx hash")
	suite.Assert().NotEmpty(txHash)
	suite.T().Logf("setValue transaction hash: %s, status: %d", txHash, status)

	suite.Assert().Equal(uint64(1), status, "Transaction should succeed")

	// Verify state changed by calling getValue()
	readResult, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"getValue",
		nil,
		0,
		nil,
	)

	suite.Require().NoError(err)
	suite.Require().Len(readResult, 1)

	value, isOk := readResult[0].(*big.Int)
	suite.Require().True(isOk)
	suite.Assert().Equal(int64(123), value.Int64(), "Value should be updated to 123")

	// Verify method is nonpayable using enum
	method := findMethodInABI(suite.contractABI, "setValue")
	suite.Require().NotNil(method)
	suite.Assert().Equal(abi.StateMutabilityNonPayable, method.GetStateMutability())
	suite.Assert().True(method.IsWriteOperation())
	suite.Assert().False(method.IsPayable())
}

// TestCallContractMethod_PayableFunction tests a payable function.
func (suite *PrivateKeySignerWithTransportTestSuite) TestCallContractMethod_PayableFunction() {
	// Get contract balance before
	balanceBefore, err := suite.transport.GetBalance(suite.contractAddress)
	suite.Require().NoError(err)

	// Call deposit(456) with 0.0001 ETH (reduced to avoid gas estimation issues)
	depositAmount := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e14)) // 0.0001 ETH
	result, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"deposit",
		depositAmount,
		0,
		nil,
		big.NewInt(456),
	)

	suite.Require().NoError(err, "Failed to call payable function")
	suite.Require().Len(result, 2)

	// Check status
	status, statusOk := result[0].(uint64)
	suite.Require().True(statusOk)
	suite.Assert().Equal(uint64(1), status)

	// Verify contract balance increased
	balanceAfter, err := suite.transport.GetBalance(suite.contractAddress)
	suite.Require().NoError(err)

	expectedIncrease := depositAmount
	actualIncrease := new(big.Int).Sub(balanceAfter, balanceBefore)
	suite.Assert().Equal(expectedIncrease.String(), actualIncrease.String(), "Contract balance should increase by deposit amount")

	// Verify method is payable using enum
	method := findMethodInABI(suite.contractABI, "deposit")
	suite.Require().NotNil(method)
	suite.Assert().Equal(abi.StateMutabilityPayable, method.GetStateMutability())
	suite.Assert().True(method.IsPayable())
	suite.Assert().True(method.IsWriteOperation())
}

// TestEstimateGas tests gas estimation.
func (suite *PrivateKeySignerWithTransportTestSuite) TestEstimateGas() {
	// Create a transaction with actual contract call data
	nonce, err := suite.transport.GetTransactionCount(suite.testAddress)
	suite.Require().NoError(err)

	// Simple ETH transfer (not a contract call) for gas estimation
	recipient := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	transaction2 := types.NewTx(&types.DynamicFeeTx{
		ChainID:   suite.chainID,
		Nonce:     nonce,
		GasTipCap: big.NewInt(1000000000), // 1 gwei
		GasFeeCap: big.NewInt(2000000000), // 2 gwei
		Gas:       100000,
		To:        &recipient,
		Value:     big.NewInt(1000),
		Data:      []byte{},
	})

	// Estimate gas
	gas, err := suite.signer.EstimateGas(transaction2)
	suite.Require().NoError(err)
	suite.Assert().True(gas > 0, "Gas estimate should be positive")
	suite.Assert().True(gas >= 21000, "Gas estimate should be at least 21000 for a simple transfer")
}

// TestSignAndVerifyMessage tests message signing and verification.
func (suite *PrivateKeySignerWithTransportTestSuite) TestSignAndVerifyMessage() {
	message := "Hello, Ethereum!"

	// Sign message
	signature, err := suite.signer.SignMessageString(message)
	suite.Require().NoError(err)
	suite.Assert().NotEmpty(signature)

	// Verify with correct address
	isValid, recoveredAddr, err := suite.signer.VerifyMessageString(suite.testAddress, message, signature)
	suite.Require().NoError(err)
	suite.Assert().True(isValid, "Signature should be valid")
	suite.Assert().Equal(suite.testAddress, recoveredAddr)

	// Verify with wrong address
	wrongAddress := common.HexToAddress("0x0000000000000000000000000000000000000001")
	isValid, _, err = suite.signer.VerifyMessageString(wrongAddress, message, signature)
	suite.Require().NoError(err)
	suite.Assert().False(isValid, "Signature should be invalid for wrong address")
}

// TestSendTransaction_Manual tests manual transaction sending.
func (suite *PrivateKeySignerWithTransportTestSuite) TestSendTransaction_Manual() {
	// Get nonce
	nonce, err := suite.transport.GetTransactionCount(suite.testAddress)
	suite.Require().NoError(err)

	// Create a simple value transfer
	recipient := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	value := big.NewInt(1000000000000000) // 0.001 ETH

	// Use EIP-1559 transaction
	transaction3 := types.NewTx(&types.DynamicFeeTx{
		ChainID:   suite.chainID,
		Nonce:     nonce,
		GasTipCap: big.NewInt(1000000000), // 1 gwei
		GasFeeCap: big.NewInt(2000000000), // 2 gwei
		Gas:       21000,
		To:        &recipient,
		Value:     value,
		Data:      nil,
	})

	// Send transaction
	txHash, err := suite.signer.SendTransaction(transaction3)
	suite.Require().NoError(err)
	suite.Assert().NotEqual(common.Hash{}, txHash)

	// Wait for receipt
	receipt, err := suite.signer.WaitForTransactionReceipt(txHash)
	suite.Require().NoError(err)
	suite.Assert().Equal(uint64(1), receipt.Status)
}

// TestStateMutabilityHelpers tests enum helper methods.
func (suite *PrivateKeySignerWithTransportTestSuite) TestStateMutabilityHelpers() {
	// Test pure function
	addMethod := findMethodInABI(suite.contractABI, "add")
	suite.Require().NotNil(addMethod)
	suite.Assert().True(addMethod.IsReadOnly())
	suite.Assert().False(addMethod.IsWriteOperation())
	suite.Assert().False(addMethod.IsPayable())
	suite.Assert().True(addMethod.IsReadable())
	suite.Assert().False(addMethod.IsWritable())

	// Test view function
	getInfoMethod := findMethodInABI(suite.contractABI, "getInfo")
	suite.Require().NotNil(getInfoMethod)
	suite.Assert().True(getInfoMethod.IsReadOnly())
	suite.Assert().False(getInfoMethod.IsWriteOperation())

	// Test nonpayable function
	setValueMethod := findMethodInABI(suite.contractABI, "setValue")
	suite.Require().NotNil(setValueMethod)
	suite.Assert().False(setValueMethod.IsReadOnly())
	suite.Assert().True(setValueMethod.IsWriteOperation())
	suite.Assert().False(setValueMethod.IsPayable())
	suite.Assert().True(setValueMethod.IsWritable())

	// Test payable function
	depositMethod := findMethodInABI(suite.contractABI, "deposit")
	suite.Require().NotNil(depositMethod)
	suite.Assert().False(depositMethod.IsReadOnly())
	suite.Assert().True(depositMethod.IsWriteOperation())
	suite.Assert().True(depositMethod.IsPayable())
	suite.Assert().True(depositMethod.IsWritable())
}

// TestErrorHandling tests various error cases.
func (suite *PrivateKeySignerWithTransportTestSuite) TestErrorHandling() {
	// Test non-existent method
	_, err := suite.signer.CallContractMethod(
		suite.contractAddress,
		suite.contractABI,
		"nonExistentMethod",
		nil,
		0,
		nil,
	)
	suite.Assert().Error(err, "Should error for non-existent method")

	// Test invalid contract address
	invalidAddress := common.HexToAddress("0x0000000000000000000000000000000000000001")
	_, err = suite.signer.CallContractMethod(
		invalidAddress,
		suite.contractABI,
		"getValue",
		nil,
		0,
		nil,
	)
	suite.Assert().Error(err, "Should error for invalid contract address")
}

// TestRunSuite runs the test suite.
func TestRunSuite(t *testing.T) {
	suite.Run(t, new(PrivateKeySignerWithTransportTestSuite))
}
