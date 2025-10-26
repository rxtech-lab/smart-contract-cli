# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool for interacting with EVM (Ethereum Virtual Machine) smart contracts. It provides abstractions for ABI parsing, transaction signing, and blockchain communication via JSON-RPC.

**Module:** `github.com/rxtech-lab/smart-contract-cli`
**Go Version:** 1.25.0

## Build and Test Commands

### Running Tests

```bash
# Run all tests (starts Anvil, runs tests, stops Anvil automatically)
make test

# Run tests manually
go test ./...

# Run specific package tests
go test ./internal/contract/evm/abi/
go test ./internal/contract/evm/contract/transport/

# Run single test
go test ./internal/contract/evm/abi/ -run TestParseAbi
```

### E2E Network Management

```bash
# Start Anvil local blockchain (required for E2E tests)
make e2e-network

# Stop Anvil
make e2e-test-stop
```

**Note:** E2E tests in `transport/http_test.go` require Anvil running on http://localhost:8545

### Building

```bash
# Build the CLI
go build -o smart-contract-cli

# Install dependencies
go mod download
go mod tidy
```

## Architecture Overview

### Core Package Structure

```
internal/
├── contract/evm/
│   ├── abi/              # ABI parsing and type definitions
│   │   ├── abi.go        # ParseAbi() - handles both array and object JSON formats
│   │   └── types.go      # ABIElement, ABIParam, ABI, ABIArray, ABIObject
│   └── contract/
│       ├── signer/       # Transaction and message signing
│       │   ├── signer.go        # Signer interface
│       │   └── privatekey.go    # PrivateKeySigner implementation
│       └── transport/    # Blockchain communication
│           ├── transport.go     # Transport interface
│           └── http.go          # HttpTransport (JSON-RPC via go-ethereum)
└── view/abi/            # (Placeholder for CLI output formatting)
```

### Key Design Patterns

**1. Interface-Based Abstractions**

- `Signer` interface: Supports multiple key management strategies (currently: private key)
- `Transport` interface: Allows different communication methods (currently: HTTP/JSON-RPC)

**2. Adapter Pattern**

- `HttpTransport` wraps `go-ethereum/ethclient.Client`
- `convertToEthereumABI()` bridges custom ABI types to go-ethereum's ABI format

**3. Flexible ABI Parsing**

- `ParseAbi()` handles two formats:
  - Solidity compiler array output: `[{type: "function", ...}, ...]`
  - Hardhat/Foundry object output: `{abi: [...], bytecode: "0x..."}`

### Component Interactions

```
┌─────────────────────┐
│    CLI (main.go)    │  ← Currently empty, needs implementation
└──────────┬──────────┘
           │
    ┌──────┴──────┐
    │             │
┌───▼────┐   ┌───▼─────┐
│ Signer │   │Transport│
└───┬────┘   └───┬─────┘
    │            │
    │            ├─ SendTransaction()
    │            ├─ WaitForTransactionReceipt()
    │            ├─ CallContract()  ← Uses ABI to pack function args
    │            ├─ EstimateGas()
    │            ├─ GetBalance()
    │            └─ GetTransactionCount()
    │
    └─ SignTransaction()  ← Uses EIP-1559 (London signer)
       SignMessageString()  ← Uses Ethereum message prefix
       VerifyMessageString()
       GetAddress()
```

### Transaction Signing Flow

1. `PrivateKeySigner.SignTransaction()` uses `types.NewLondonSigner(chainID)` for EIP-1559 compatibility
2. Message signing follows Ethereum standard: `\x19Ethereum Signed Message:\n{length}{message}`
3. Signature verification handles both MetaMask format (v=27/28) and standard format (v=0/1)

### Contract Calling Flow

1. Parse ABI with `abi.ParseAbi(jsonString)`
2. `HttpTransport.CallContract()` internally:
   - Converts custom ABI to go-ethereum ABI format
   - Packs function arguments using go-ethereum's ABI encoder
   - Creates `ethereum.CallMsg` with packed data
   - Executes via `ethclient.CallContract()`
   - Returns raw bytes (caller must decode)

### Transaction Receipt Polling

- `WaitForTransactionReceipt()` polls every 1 second with 5-minute timeout
- Uses simple polling (not WebSocket subscriptions) for broader RPC compatibility

## Key Implementation Details

### ABI Parsing

- **Location:** `internal/contract/evm/abi/abi.go:24-51`
- Detects JSON format automatically (array vs object)
- Custom `UnmarshalJSON()` on `ABI` type handles dual format support
- Returns `ABIArray` for easier iteration over functions/events

### Private Key Signer

- **Location:** `internal/contract/evm/contract/signer/privatekey.go`
- `NewPrivateKeySigner(hexKey)` accepts hex-encoded private key (without 0x prefix)
- `GetAddress()` derives Ethereum address from private key

### HTTP Transport

- **Location:** `internal/contract/evm/contract/transport/http.go`
- `NewHttpTransport(endpoint)` validates connectivity on initialization
- RPC endpoint typically: `http://localhost:8545` (Anvil) or Infura/Alchemy URLs
- Thread-safe for concurrent operations (see `http_test.go` concurrent test)

## Testing Strategy

### ABI Tests

- **Location:** `internal/contract/evm/abi/abi_test.go`
- 9 comprehensive test cases including real ERC20 ABIs
- Tests edge cases: empty ABIs, complex tuples, payable functions

### E2E Transport Tests

- **Location:** `internal/contract/evm/contract/transport/http_test.go`
- Uses `testify/suite` for organized test lifecycle
- Requires Anvil running (pre-funded accounts with ~1000 ETH)
- Tests: balance queries, nonce retrieval, gas estimation, concurrent operations

## Development Notes

** Always use test suite testing structure to write tests**

### Adding New Transport Implementation

1. Implement `transport.Transport` interface
2. Add factory function (e.g., `NewWebSocketTransport()`)
3. Consider thread-safety for concurrent usage

### Adding New Signer Implementation

1. Implement `signer.Signer` interface
2. Ensure EIP-1559 compatibility in `SignTransaction()`
3. Follow Ethereum message prefix standard in `SignMessageString()`

### Working with ABIs

- Use `abi.ParseAbi()` - it auto-detects format
- Access elements via `ABIArray` for iteration
- `ABIElement.Type` values: "function", "event", "constructor", "fallback", "receive"

### Test Account (Anvil Default)

- Private Key: `0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80`
- Address: `0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266`
- Balance: ~10000 ETH (Anvil default funding)
