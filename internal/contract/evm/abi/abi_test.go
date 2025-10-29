package abi

import (
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
