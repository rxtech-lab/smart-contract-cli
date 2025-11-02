# Wallet Management Implementation Plan

This document describes the implementation of the wallet management system for the smart-contract-cli project.

## Overview

The wallet management system provides a comprehensive TUI-based interface for managing Ethereum wallets with the following features:

- Import wallets from private keys or mnemonic phrases
- Generate new wallets with BIP39 mnemonics
- Secure encrypted storage of private keys and mnemonics
- HD wallet derivation path support
- Balance fetching from RPC endpoints
- CRUD operations on wallet metadata

## Architecture

### 1. Database Layer

**Location:** `internal/contract/evm/storage/`

#### Wallet Model

**File:** `internal/contract/evm/storage/models/evm/wallet.go`

```go
type EVMWallet struct {
    ID             uint
    Alias          string
    Address        string
    DerivationPath *string  // For HD wallets (e.g., "m/44'/60'/0'/0/0")
    IsFromMnemonic bool     // Indicates if wallet was created from mnemonic
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

**Key Features:**

- `Alias`: User-friendly name for the wallet (unique)
- `Address`: Ethereum address derived from private key (unique)
- `DerivationPath`: BIP44 derivation path (only for mnemonic-based wallets)
- `IsFromMnemonic`: Flag to track wallet creation method
- Private keys/mnemonics are NOT stored in the database (stored in secure storage)

#### Wallet Queries

**File:** `internal/contract/evm/storage/sql/queries/wallet_queries.go`

Provides CRUD operations:

- `Create(wallet)` - Create new wallet
- `GetByID(id)` - Retrieve by ID
- `GetByAddress(address)` - Retrieve by Ethereum address
- `GetByAlias(alias)` - Retrieve by user-defined alias
- `List(page, pageSize)` - Paginated list
- `Search(query)` - Search by alias or address
- `Update(id, updates)` - Update wallet metadata
- `Delete(id)` - Delete wallet
- `ExistsByAddress(address)` - Check existence by address
- `ExistsByAlias(alias)` - Check existence by alias
- `Count()` - Total wallet count

#### Storage Interface Extension

**File:** `internal/contract/evm/storage/sql/storage.go`

Added wallet methods to the `Storage` interface:

```go
type Storage interface {
    // ... existing methods ...

    // Wallet methods
    CreateWallet(wallet models.EVMWallet) (id uint, err error)
    ListWallets(page int64, pageSize int64) (wallets types.Pagination[models.EVMWallet], err error)
    SearchWallets(query string) (wallets types.Pagination[models.EVMWallet], err error)
    GetWalletByID(id uint) (wallet models.EVMWallet, err error)
    GetWalletByAddress(address string) (wallet models.EVMWallet, err error)
    GetWalletByAlias(alias string) (wallet models.EVMWallet, err error)
    CountWallets() (count int64, err error)
    UpdateWallet(id uint, wallet models.EVMWallet) (err error)
    DeleteWallet(id uint) (err error)
    WalletExistsByAddress(address string) (exists bool, err error)
    WalletExistsByAlias(alias string) (exists bool, err error)
}
```

#### SQLite Implementation

**File:** `internal/contract/evm/storage/sql/sqlite.go`

- Added `walletQueries *queries.WalletQueries` field to `SQLiteStorage` struct
- Implemented all wallet interface methods
- Added `&models.EVMWallet{}` to `AutoMigrate()` for automatic table creation
- Initialized `walletQueries` in `NewSQLiteDB()` constructor

### 2. Service Layer

**Location:** `internal/contract/evm/wallet/`

#### Wallet Service Interface

**File:** `internal/contract/evm/wallet/service.go`

```go
type WalletService interface {
    // Import/Generate wallets
    ImportPrivateKey(alias string, privateKeyHex string) (*models.EVMWallet, error)
    ImportMnemonic(alias string, mnemonic string, derivationPath string) (*models.EVMWallet, error)
    GenerateWallet(alias string) (wallet *models.EVMWallet, mnemonic string, privateKey string, err error)

    // Retrieve wallets with balance
    GetWalletWithBalance(walletID uint, rpcEndpoint string) (*WalletWithBalance, error)
    ListWalletsWithBalances(page int64, pageSize int64, rpcEndpoint string) (wallets []WalletWithBalance, totalCount int64, err error)

    // Secure data access
    GetPrivateKey(walletID uint) (string, error)
    GetMnemonic(walletID uint) (string, error)

    // Update operations
    UpdateWalletAlias(walletID uint, newAlias string) error
    UpdateWalletPrivateKey(walletID uint, newPrivateKeyHex string) error

    // Delete operations
    DeleteWallet(walletID uint) error

    // Validation
    ValidatePrivateKey(privateKeyHex string) error
    ValidateMnemonic(mnemonic string) error

    // Existence checks
    WalletExistsByAddress(address string) (bool, error)
    WalletExistsByAlias(alias string) (bool, error)

    // Retrieval
    GetWallet(walletID uint) (*models.EVMWallet, error)
}
```

#### Implementation Details

**Dependencies:**

- `sql.Storage` - For wallet metadata CRUD operations
- `storage.SecureStorage` - For encrypted private key/mnemonic storage

**Secure Storage Keys:**

- Private keys: `wallet:{id}:privatekey`
- Mnemonics: `wallet:{id}:mnemonic`

**Key Features:**

1. **ImportPrivateKey:**

   - Validates private key format (64 hex characters)
   - Derives Ethereum address from private key
   - Checks for duplicate addresses and aliases
   - Stores wallet metadata in database
   - Encrypts and stores private key in secure storage
   - Rollback on failure

2. **ImportMnemonic:**

   - Validates mnemonic phrase (12 or 24 words, BIP39 compliant)
   - Derives private key using BIP44 derivation path
   - Default path: `m/44'/60'/0'/0/0` (Ethereum standard)
   - Stores both mnemonic and derived private key in secure storage
   - Marks wallet as `IsFromMnemonic = true`

3. **GenerateWallet:**

   - Generates 128-bit entropy (12-word mnemonic)
   - Uses BIP39 standard for mnemonic generation
   - Automatically imports using default derivation path
   - Returns wallet, mnemonic, and private key for user backup

4. **GetWalletWithBalance:**

   - Retrieves wallet metadata from database
   - Connects to RPC endpoint using `HttpTransport`
   - Fetches balance via `GetBalance()`
   - Returns `WalletWithBalance` struct with balance or error

5. **HD Wallet Derivation:**
   - Simplified BIP32/BIP44 implementation
   - Supports standard Ethereum paths
   - Uses `accounts.ParseDerivationPath()` for path parsing
   - Derives private key from seed + account index

**Validation:**

- Private key: Must be 64 hex characters (with or without 0x prefix)
- Mnemonic: Must be valid BIP39 phrase (12 or 24 words)
- Checks performed before any database operations

### 3. Dependencies

Added to `go.mod`:

```
github.com/tyler-smith/go-bip39 v1.1.0
go.uber.org/mock v0.6.0
```

**BIP39:** For mnemonic generation and validation
**MockGen:** For generating mock implementations for testing

### 4. Mock Generation

**Configuration:** `tools/tools.go`

```go
//go:generate go run go.uber.org/mock/mockgen -source=../internal/contract/evm/wallet/service.go -destination=../internal/contract/evm/wallet/mock_service.go -package=wallet
//go:generate go run go.uber.org/mock/mockgen -source=../internal/contract/evm/storage/sql/storage.go -destination=../internal/contract/evm/storage/sql/mock_storage.go -package=sql
//go:generate go run go.uber.org/mock/mockgen -source=../internal/storage/secure.go -destination=../internal/storage/mock_secure.go -package=storage
```

**Usage:** Run `go generate ./tools/...` to generate mocks

### 5. TUI Pages

#### Main Wallet List Page

**Location:** `app/evm/wallet/page.go`

**Features:**

- Displays all wallets with balances
- Shows selected wallet with ★ indicator
- Color-coded: selected wallet in green
- Cursor navigation (up/down/k/j)
- Actions:
  - `enter` - Open wallet actions menu
  - `a` - Add new wallet
  - `r` - Refresh balances
  - `esc/q` - Go back

**States:**

- Loading: Shows loading spinner while fetching wallets
- Empty: Helpful message when no wallets exist
- Error: Displays error message with retry option
- Loaded: Shows wallet list with balances

**Data Flow:**

1. `Init()` triggers `loadWallets()` command
2. Retrieves `Storage` and `SecureStorage` from shared memory
3. Creates `WalletService` instance
4. Fetches wallets with balances from RPC endpoint
5. Returns `walletLoadedMsg` with data or error
6. `Update()` handles message and updates model state

**Balance Fetching:**

- Uses `HttpTransport` to connect to RPC endpoint
- Gracefully handles connection errors
- Shows "unavailable ⚠" if balance fetch fails
- Converts Wei to ETH for display

#### Router Integration

**Updated:** `app/evm/page.go`

Added wallet management option to EVM menu:

```go
{
    Label: "Wallet Management",
    Value: "wallet-management",
    Route: "/evm/wallet",
    Description: "Manage your wallets and private keys"
}
```

### 6. Shared Memory Keys

**Location:** `internal/config/shared_memory_keys.go`

Added:

```go
SelectedWalletIDKey = "selected_wallet_id"
```

Stores the currently selected wallet ID for use across the application.

### 7. Security Considerations

**Secure Storage:**

- All private keys and mnemonics encrypted using AES-GCM
- Encryption key derived from user password using SHA-256
- Secure storage file permissions: `0600` (read/write owner only)
- Keys stored separately from wallet metadata

**Key Formats:**

- Private keys: Stored with `0x` prefix
- Mnemonics: Stored as plain space-separated words
- Addresses: Stored as checksummed Ethereum addresses

**Validation:**

- Duplicate address detection before wallet creation
- Duplicate alias detection before wallet creation
- Private key format validation (64 hex chars)
- Mnemonic BIP39 validation

### 8. Error Handling

**Database Errors:**

- Wrapped with custom error types from `internal/errors`
- Error codes: `ErrCodeDatabaseOperationFailed`, `ErrCodeRecordNotFound`
- Rollback on failure (delete wallet if secure storage fails)

**Network Errors:**

- Balance fetch failures don't prevent wallet display
- Errors stored in `WalletWithBalance.Error` field
- User can retry with 'r' key

**Validation Errors:**

- Clear error messages with helpful hints
- Example: "Invalid mnemonic phrase: expected 12 or 24 words, got 11 words"

## Future Enhancements (Not Implemented)

The following features are planned but not yet implemented:

### Additional TUI Pages

1. **Wallet Actions Page** (`app/evm/wallet/actions/page.go`)

   - Select as active wallet
   - View details
   - Update wallet
   - Delete wallet

2. **Add Wallet Pages**

   - Method selection (`app/evm/wallet/add/page.go`)
   - Private key import (`app/evm/wallet/add/privatekey/page.go`)
   - Mnemonic import (`app/evm/wallet/add/mnemonic/page.go`)
   - Generate new (`app/evm/wallet/add/generate/page.go`)

3. **Wallet Details Page** (`app/evm/wallet/details/page.go`)

   - Full wallet information
   - Balance history
   - Transaction statistics
   - Show private key (with security prompt)

4. **Update Pages**

   - Update alias (`app/evm/wallet/update/page.go`)
   - Update private key (with warnings)

5. **Delete Confirmation** (`app/evm/wallet/delete/page.go`)

   - Confirmation dialog
   - Protection for selected wallet

6. **Select Wallet** (`app/evm/wallet/select/page.go`)
   - Switch active wallet confirmation

### Testing

**Unit Tests:**

- `internal/contract/evm/wallet/service_test.go`
- `internal/contract/evm/storage/sql/queries/wallet_queries_test.go`
- Test wallet CRUD operations
- Test mnemonic generation/validation
- Test HD wallet derivation
- Test duplicate detection

**E2E Tests:**

- `app/evm/wallet/page_test.go`
- Use teatest framework (like `app/page_test.go`)
- Mock WalletService and Storage
- Test user interactions (navigation, adding wallet, etc.)

### Enhanced Features

1. **ENS Name Resolution**

   - Display ENS names if available
   - Reverse resolution for addresses

2. **USD Value Display**

   - Fetch ETH price from API
   - Display balance in USD

3. **Transaction History**

   - Fetch from blockchain explorer API
   - Display recent transactions

4. **Multiple RPC Endpoints**

   - Support for different networks (mainnet, testnets)
   - Network selection in UI

5. **Backup/Export**

   - Export wallet as JSON
   - Backup mnemonics to file

6. **Hardware Wallet Support**
   - Ledger integration
   - Trezor integration

## File Structure

```
smart-contract-cli/
├── internal/
│   ├── config/
│   │   └── shared_memory_keys.go          # Added SelectedWalletIDKey
│   ├── contract/evm/
│   │   ├── storage/
│   │   │   ├── models/evm/
│   │   │   │   └── wallet.go              # NEW: Wallet model
│   │   │   └── sql/
│   │   │       ├── queries/
│   │   │       │   └── wallet_queries.go  # NEW: Wallet CRUD operations
│   │   │       ├── storage.go             # UPDATED: Added wallet methods
│   │   │       └── sqlite.go              # UPDATED: Implemented wallet methods
│   │   └── wallet/
│   │       └── service.go                 # NEW: Wallet business logic
│   └── storage/
│       ├── secure.go                      # Used for encrypted storage
│       └── shared_memory.go               # Used for sharing data
├── app/
│   └── evm/
│       ├── page.go                        # UPDATED: Added wallet management option
│       └── wallet/
│           └── page.go                    # NEW: Main wallet list page
├── tools/
│   └── tools.go                           # UPDATED: Added mockgen directives
└── docs/
    ├── wallet-management-design.md        # UI/UX design specification
    └── wallet-management-implementation.md # This file
```

## Build and Run

```bash
# Install dependencies
go mod tidy

# Generate mocks (if needed)
go generate ./tools/...

# Build project
make build

# Run tests
make test
```

## Usage Flow

1. User navigates to EVM blockchain menu
2. Selects "Wallet Management"
3. Main wallet list page displays:
   - Empty state if no wallets
   - List of wallets with balances if wallets exist
4. User can:
   - Press 'a' to add new wallet (not implemented yet)
   - Press 'enter' on a wallet for actions (not implemented yet)
   - Press 'r' to refresh balances
   - Press 'esc' to go back

## Implementation Notes

### HD Wallet Derivation

The current implementation uses a simplified BIP32/BIP44 derivation:

- Seed is generated from mnemonic using BIP39
- Derivation path is parsed using go-ethereum's `accounts.ParseDerivationPath()`
- Private key is derived using a simplified approach (seed + account index hashing)
- **For production:** Should use a full BIP32 library for proper hierarchical derivation

### Balance Fetching Performance

Current implementation fetches balances sequentially for each wallet. For better performance:

- Could implement concurrent balance fetching with worker pool
- Could cache balances with TTL
- Could use WebSocket subscriptions for real-time updates

### Security Best Practices

- Never log private keys or mnemonics
- Clear sensitive data from memory after use
- Use secure random number generation for entropy
- Validate all user inputs before processing
- Implement rate limiting for RPC calls

## References

- [BIP39 - Mnemonic Code](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki)
- [BIP32 - Hierarchical Deterministic Wallets](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki)
- [BIP44 - Multi-Account Hierarchy](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
- [go-ethereum Documentation](https://geth.ethereum.org/docs)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
