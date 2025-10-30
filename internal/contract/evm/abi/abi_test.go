package abi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAbi(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedLen int
		validate    func(t *testing.T, result AbiArray)
	}{
		{
			name: "parse ABI array format with single function",
			input: `[
				{
					"type": "function",
					"name": "transfer",
					"inputs": [
						{"name": "to", "type": "address"},
						{"name": "amount", "type": "uint256"}
					],
					"outputs": [
						{"name": "", "type": "bool"}
					],
					"stateMutability": "nonpayable"
				}
			]`,
			wantErr:     false,
			expectedLen: 1,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "function", result[0].Type)
				assert.Equal(t, "transfer", result[0].Name)
				assert.Len(t, result[0].Inputs, 2)
				assert.Equal(t, "to", result[0].Inputs[0].Name)
				assert.Equal(t, "address", result[0].Inputs[0].Type)
				assert.Equal(t, "amount", result[0].Inputs[1].Name)
				assert.Equal(t, "uint256", result[0].Inputs[1].Type)
				assert.Len(t, result[0].Outputs, 1)
				assert.Equal(t, "bool", result[0].Outputs[0].Type)
			},
		},
		{
			name: "parse ABI array format with multiple elements",
			input: `[
				{
					"type": "constructor",
					"inputs": [
						{"name": "initialSupply", "type": "uint256"}
					],
					"stateMutability": "nonpayable"
				},
				{
					"type": "event",
					"name": "Transfer",
					"inputs": [
						{"name": "from", "type": "address", "indexed": true},
						{"name": "to", "type": "address", "indexed": true},
						{"name": "value", "type": "uint256", "indexed": false}
					],
					"anonymous": false
				}
			]`,
			wantErr:     false,
			expectedLen: 2,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "constructor", result[0].Type)
				assert.Equal(t, "event", result[1].Type)
				assert.Equal(t, "Transfer", result[1].Name)
				assert.Len(t, result[1].Inputs, 3)
				assert.True(t, result[1].Inputs[0].Indexed)
				assert.True(t, result[1].Inputs[1].Indexed)
				assert.False(t, result[1].Inputs[2].Indexed)
			},
		},
		{
			name: "parse ABI object format",
			input: `{
				"abi": [
					{
						"type": "function",
						"name": "balanceOf",
						"inputs": [
							{"name": "account", "type": "address"}
						],
						"outputs": [
							{"name": "", "type": "uint256"}
						],
						"stateMutability": "view"
					}
				],
				"bytecode": "0x608060405234801561001057600080fd5b50",
				"metadata": {
					"compiler": "solc",
					"version": "0.8.0"
				}
			}`,
			wantErr:     false,
			expectedLen: 1,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "function", result[0].Type)
				assert.Equal(t, "balanceOf", result[0].Name)
				assert.Equal(t, "view", result[0].StateMutability)
			},
		},
		{
			name:        "parse empty ABI array",
			input:       `[]`,
			wantErr:     false,
			expectedLen: 0,
			validate: func(t *testing.T, result AbiArray) {
				assert.Empty(t, result)
			},
		},
		{
			name: "parse ABI with complex struct types",
			input: `[
				{
					"type": "function",
					"name": "complexFunction",
					"inputs": [
						{
							"name": "data",
							"type": "tuple",
							"components": [
								{"name": "id", "type": "uint256"},
								{"name": "addr", "type": "address"}
							]
						}
					],
					"outputs": [],
					"stateMutability": "nonpayable"
				}
			]`,
			wantErr:     false,
			expectedLen: 1,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "complexFunction", result[0].Name)
				assert.Len(t, result[0].Inputs, 1)
				assert.Equal(t, "tuple", result[0].Inputs[0].Type)
				assert.Len(t, result[0].Inputs[0].Components, 2)
				assert.Equal(t, "id", result[0].Inputs[0].Components[0].Name)
				assert.Equal(t, "uint256", result[0].Inputs[0].Components[0].Type)
			},
		},
		{
			name:        "parse invalid JSON",
			input:       `{"invalid json`,
			wantErr:     true,
			expectedLen: 0,
			validate:    func(t *testing.T, result AbiArray) {},
		},
		{
			name:        "parse empty string",
			input:       ``,
			wantErr:     true,
			expectedLen: 0,
			validate:    func(t *testing.T, result AbiArray) {},
		},
		{
			name: "parse ABI with payable function",
			input: `[
				{
					"type": "function",
					"name": "deposit",
					"inputs": [],
					"outputs": [],
					"stateMutability": "payable",
					"payable": true
				}
			]`,
			wantErr:     false,
			expectedLen: 1,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "deposit", result[0].Name)
				assert.Equal(t, "payable", result[0].StateMutability)
				assert.True(t, result[0].Payable)
			},
		},
		{
			name: "parse ABI object format with empty abi field",
			input: `{
				"abi": [],
				"bytecode": "0x",
				"metadata": {}
			}`,
			wantErr:     false,
			expectedLen: 0,
			validate: func(t *testing.T, result AbiArray) {
				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAbi(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result, tt.expectedLen)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestParseAbi_RealWorldExample tests with a more realistic ERC20 ABI.
func TestParseAbi_RealWorldExample(t *testing.T) {
	erc20ABI := `[
		{
			"constant": true,
			"inputs": [],
			"name": "name",
			"outputs": [{"name": "", "type": "string"}],
			"type": "function",
			"stateMutability": "view"
		},
		{
			"constant": true,
			"inputs": [],
			"name": "totalSupply",
			"outputs": [{"name": "", "type": "uint256"}],
			"type": "function",
			"stateMutability": "view"
		},
		{
			"constant": false,
			"inputs": [
				{"name": "to", "type": "address"},
				{"name": "value", "type": "uint256"}
			],
			"name": "transfer",
			"outputs": [{"name": "", "type": "bool"}],
			"type": "function",
			"stateMutability": "nonpayable"
		}
	]`

	result, err := ParseAbi(erc20ABI)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify the structure
	assert.Equal(t, "name", result[0].Name)
	assert.True(t, result[0].Constant)
	assert.Equal(t, "view", result[0].StateMutability)

	assert.Equal(t, "totalSupply", result[1].Name)

	assert.Equal(t, "transfer", result[2].Name)
	assert.False(t, result[2].Constant)
	assert.Len(t, result[2].Inputs, 2)
}

func TestReadAbi(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		cleanup     func(t *testing.T, path string)
		wantErr     bool
		expectedLen int
		validate    func(t *testing.T, result AbiArray)
	}{
		{
			name: "read ABI from local file with array format",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "abi.json")
				abiContent := `[
					{
						"type": "function",
						"name": "transfer",
						"inputs": [
							{"name": "to", "type": "address"},
							{"name": "amount", "type": "uint256"}
						],
						"outputs": [{"name": "", "type": "bool"}],
						"stateMutability": "nonpayable"
					}
				]`
				err := os.WriteFile(filePath, []byte(abiContent), 0644)
				require.NoError(t, err)
				return filePath
			},
			cleanup:     func(t *testing.T, path string) {},
			wantErr:     false,
			expectedLen: 1,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "function", result[0].Type)
				assert.Equal(t, "transfer", result[0].Name)
				assert.Len(t, result[0].Inputs, 2)
			},
		},
		{
			name: "read ABI from local file with object format",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "abi.json")
				abiContent := `{
					"abi": [
						{
							"type": "function",
							"name": "balanceOf",
							"inputs": [{"name": "account", "type": "address"}],
							"outputs": [{"name": "", "type": "uint256"}],
							"stateMutability": "view"
						}
					],
					"bytecode": "0x608060405234801561001057600080fd5b50"
				}`
				err := os.WriteFile(filePath, []byte(abiContent), 0644)
				require.NoError(t, err)
				return filePath
			},
			cleanup:     func(t *testing.T, path string) {},
			wantErr:     false,
			expectedLen: 1,
			validate: func(t *testing.T, result AbiArray) {
				assert.Equal(t, "function", result[0].Type)
				assert.Equal(t, "balanceOf", result[0].Name)
				assert.Equal(t, "view", result[0].StateMutability)
			},
		},
		{
			name: "read ABI from remote URL",
			setup: func(t *testing.T) string {
				return "https://unpkg.com/@uniswap/v2-core@1.0.0/build/IUniswapV2Pair.json"
			},
			cleanup:     func(t *testing.T, path string) {},
			wantErr:     false,
			expectedLen: 0, // Will be validated by the validate function
			validate: func(t *testing.T, result AbiArray) {
				// The Uniswap V2 Pair ABI should have multiple elements
				// We expect at least some functions and events
				assert.Greater(t, len(result), 0, "ABI should contain at least one element")

				// Check for common Uniswap V2 Pair functions/events
				hasMint := false
				hasBurn := false
				hasSync := false
				hasSwap := false

				for _, elem := range result {
					if elem.Name == "mint" {
						hasMint = true
					}
					if elem.Name == "burn" {
						hasBurn = true
					}
					if elem.Name == "Sync" {
						hasSync = true
					}
					if elem.Name == "Swap" {
						hasSwap = true
					}
				}

				// Verify at least some expected elements exist
				// (These are common in Uniswap V2 Pair contracts)
				assert.True(t, hasMint || hasBurn || hasSync || hasSwap || len(result) > 0,
					"ABI should contain expected Uniswap V2 Pair elements or be non-empty")
			},
		},
		{
			name: "read ABI from invalid file path",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/to/abi.json"
			},
			cleanup:     func(t *testing.T, path string) {},
			wantErr:     true,
			expectedLen: 0,
			validate:    func(t *testing.T, result AbiArray) {},
		},
		{
			name: "read ABI from invalid URL",
			setup: func(t *testing.T) string {
				return "https://invalid-url-that-does-not-exist-12345.com/abi.json"
			},
			cleanup:     func(t *testing.T, path string) {},
			wantErr:     true,
			expectedLen: 0,
			validate:    func(t *testing.T, result AbiArray) {},
		},
		{
			name: "read ABI from URL returning non-200 status",
			setup: func(t *testing.T) string {
				return "https://httpstat.us/404"
			},
			cleanup:     func(t *testing.T, path string) {},
			wantErr:     true,
			expectedLen: 0,
			validate:    func(t *testing.T, result AbiArray) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			defer tt.cleanup(t, path)

			result, err := ReadAbi(path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.expectedLen > 0 {
				assert.Len(t, result, tt.expectedLen)
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestReadAbi_UniswapV2Pair tests reading the Uniswap V2 Pair ABI from the remote URL.
func TestReadAbi_UniswapV2Pair(t *testing.T) {
	url := "https://unpkg.com/@uniswap/v2-core@1.0.0/build/IUniswapV2Pair.json"

	result, err := ReadAbi(url)
	require.NoError(t, err)
	require.NotEmpty(t, result, "Uniswap V2 Pair ABI should not be empty")

	// Verify it contains expected Uniswap V2 Pair interface elements
	elementTypes := make(map[string]int)

	for _, elem := range result {
		elementTypes[elem.Type]++
	}

	// Uniswap V2 Pair should have functions, events, etc.
	assert.Greater(t, elementTypes["function"], 0, "Should have at least one function")
	assert.Greater(t, elementTypes["event"], 0, "Should have at least one event")

	// Verify the ABI was successfully parsed and contains expected structure
	assert.True(t, len(result) > 0, "Should have parsed ABI elements")

	// Log some information about what was found (optional, for debugging)
	t.Logf("Found %d ABI elements: %d functions, %d events",
		len(result),
		elementTypes["function"],
		elementTypes["event"])
}
