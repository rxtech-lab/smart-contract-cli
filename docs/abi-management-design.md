# ABI Management - Terminal UI Design

This document shows the mock terminal design for the ABI management page.

## 1. Main ABI List View

```
ABI Management

Manage your contract ABIs

  > ERC20 Token
    Name: ERC20 Token
    Functions: 9 • Events: 2
    Created: 2024-10-15 10:30 AM

    Uniswap V2 Router
    Name: Uniswap V2 Router
    Functions: 24 • Events: 0
    Created: 2024-10-20 02:15 PM

    NFT Collection
    Name: NFT Collection
    Functions: 12 • Events: 3
    Created: 2024-10-25 09:45 AM

    Custom DEX Router
    Name: Custom DEX Router
    Functions: 18 • Events: 5
    Created: 2024-10-28 03:20 PM

    Staking Contract
    Name: Staking Contract
    Functions: 14 • Events: 4
    Created: 2024-10-30 11:15 AM


Page 1 of 3 • Showing 5 of 15 ABIs

Legend:
> = Selected

↑/k: up • ↓/j: down • enter: view details • a: add new • d: delete • e: edit
n: next page • p: previous page • esc/q: back
```

## 2. Add ABI - Import Method Selection

User presses 'a' to add a new ABI.

```
Add New ABI

How would you like to import the ABI?

  > Enter manually
    Type or paste the ABI JSON directly

    Import from URL
    Fetch ABI from a remote URL (e.g., Etherscan)

    Import from local file
    Load ABI from a file on your system


↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 3a. Add ABI - Manual Entry (Name Input)

User selects "Enter manually".

```
Add New ABI - Manual Entry

Step 1/2: Enter ABI Name

Name: My Custom Contract_


enter: next • esc: cancel
```

## 3b. Add ABI - Manual Entry (JSON Input)

After entering name, user proceeds to ABI JSON input.

```
Add New ABI - Manual Entry

Step 2/2: Paste or type the ABI JSON

ABI JSON (supports array or object format):
────────────────────────────────────────────────────────────────────────────
[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"
},{"name":"amount","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],
"stateMutability":"nonpayable"},{"type":"function","name":"balanceOf","inputs
":[{"name":"account","type":"address"}],"outputs":[{"name":"","type":"uint25
6"}],"stateMutability":"view"}]_


────────────────────────────────────────────────────────────────────────────

Tip: Supports both array format and Hardhat/Foundry object format

enter: save • esc: cancel
```

## 4a. Add ABI - Import from URL

User selects "Import from URL".

```
Add New ABI - Import from URL

Enter the URL to fetch the ABI from:

URL: https://api.etherscan.io/api?module=contract&action=getabi&address=0x_


Examples:
• Etherscan API: https://api.etherscan.io/api?module=contract&action=...
• Direct JSON: https://example.com/contracts/MyContract.json
• GitHub raw: https://raw.githubusercontent.com/user/repo/path/abi.json

enter: fetch • esc: cancel
```

## 4b. Add ABI - Import from URL (Fetching)

After entering URL and pressing enter.

```
Add New ABI - Import from URL

Fetching ABI from remote URL...

URL: https://api.etherscan.io/api?module=contract&action=getabi&address=...

⠋ Please wait...
```

## 4c. Add ABI - Import from URL (Confirmation)

After successful fetch, show confirmation with detected information.

```
Add New ABI - Import Confirmation

✓ ABI successfully fetched!

Detected Information:
• Functions: 15
• Events: 4
• Constructor: Yes
• Fallback/Receive: No

Enter a name for this ABI:
Name: USDC Token_


enter: save • esc: cancel
```

## 5a. Add ABI - Import from Local File

User selects "Import from local file".

```
Add New ABI - Import from Local File

Enter the path to the ABI JSON file:

File path: ~/contracts/artifacts/MyContract.json_


Examples:
• Absolute: /Users/user/project/artifacts/MyContract.json
• Relative: ./contracts/MyContract.json
• Home: ~/contracts/MyContract.json

Supported formats: JSON (array or Hardhat/Foundry object)

enter: load • esc: cancel
```

## 5b. Add ABI - Import from Local File (Success)

After successfully loading the file.

```
Add New ABI - Import Confirmation

✓ ABI successfully loaded!

File: ~/contracts/artifacts/MyContract.json

Detected Information:
• Functions: 8
• Events: 2
• Constructor: Yes
• Fallback/Receive: No

Enter a name for this ABI:
Name: My Custom Contract_


enter: save • esc: cancel
```

## 6. Add ABI - Import Error (URL)

If URL import fails.

```
Add New ABI - Import Error

✗ Failed to import ABI

Error: Failed to fetch from URL: HTTP 404 Not Found

Possible reasons:
• Invalid URL or endpoint
• Network connection issues
• API rate limiting
• Invalid API key or permissions


Press any key to go back and try again...
```

## 6b. Add ABI - Import Error (File Not Found)

If file import fails.

```
Add New ABI - Import Error

✗ Failed to import ABI

Error: File not found: ~/contracts/artifacts/MyContract.json

Possible reasons:
• File does not exist at the specified path
• Incorrect file path or typo
• Insufficient file permissions


Press any key to go back and try again...
```

## 6c. Add ABI - Import Error (Invalid JSON)

If JSON parsing fails.

```
Add New ABI - Import Error

✗ Failed to parse ABI

Error: Invalid JSON format at line 15, column 23

Details:
• Expected ',' or ']' after object in array
• Make sure the JSON is properly formatted
• Check for missing commas, brackets, or quotes

Tip: Validate your JSON at https://jsonlint.com

Press any key to go back and try again...
```

## 7. View ABI Details

User presses 'enter' on an ABI in the main list to view details.

```
ABI Details - ERC20 Token

Name: ERC20 Token
Created: 2024-10-15 10:30 AM
Last Modified: 2024-10-15 10:30 AM

Summary
• Functions: 9 (6 view, 3 non-payable)
• Events: 2
• Constructor: Yes

Functions
  name() → string
  symbol() → string
  decimals() → uint8
  totalSupply() → uint256
  balanceOf(address account) → uint256
  transfer(address to, uint256 amount) → bool
  allowance(address owner, address spender) → uint256
  approve(address spender, uint256 amount) → bool
  transferFrom(address from, address to, uint256 amount) → bool

Events
  Transfer(address indexed from, address indexed to, uint256 value)
  Approval(address indexed owner, address indexed spender, uint256 value)


m: view methods • e: edit • d: delete • esc/q: back
```

## 8. View ABI Methods (Detailed)

User presses 'm' to view all methods with full details.

```
ABI Methods - ERC20 Token

Showing 9 functions

> 1. name
   Type: function
   State: view
   Inputs: none
   Outputs: string

  2. symbol
   Type: function
   State: view
   Inputs: none
   Outputs: string

  3. decimals
   Type: function
   State: view
   Inputs: none
   Outputs: uint8

  4. totalSupply
   Type: function
   State: view
   Inputs: none
   Outputs: uint256

↑/k: up • ↓/j: down • esc/q: back
```

Scrolling down to show more complex method:

```
ABI Methods - ERC20 Token

Showing 9 functions

  5. balanceOf
   Type: function
   State: view
   Inputs:
     • account (address)
   Outputs:
     • uint256

> 6. transfer
   Type: function
   State: nonpayable
   Inputs:
     • to (address)
     • amount (uint256)
   Outputs:
     • bool

  7. allowance
   Type: function
   State: view
   Inputs:
     • owner (address)
     • spender (address)
   Outputs:
     • uint256

↑/k: up • ↓/j: down • esc/q: back
```

## 9. Edit ABI

User presses 'e' to edit an ABI (edit name).

```
Edit ABI

Current name: ERC20 Token

New name: USDC Token_


Note: You can only edit the name. To update the ABI JSON, delete and
create a new one.

enter: save • esc: cancel
```

## 10. Delete ABI Confirmation (With Contract References)

User presses 'd' to delete an ABI that has contracts using it.

```
Delete ABI

Are you sure you want to delete this ABI?

Name: ERC20 Token
Functions: 9
Events: 2

⚠ Warning: 3 contracts are using this ABI:
  • USDC Contract (0x1234...5678)
  • DAI Contract (0xabcd...efgh)
  • Custom Token (0x9876...4321)

Deleting this ABI will unlink it from these contracts.

  > No, cancel
    Yes, delete

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 10b. Delete ABI Confirmation (No References)

If no contracts are using the ABI:

```
Delete ABI

Are you sure you want to delete this ABI?

Name: ERC20 Token
Functions: 9
Events: 2

This action cannot be undone.

  > No, cancel
    Yes, delete

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 11. Empty State

When no ABIs are stored yet.

```
ABI Management

No ABIs found

You haven't added any ABIs yet. ABIs (Application Binary Interfaces) are
required to interact with smart contracts.

Press 'a' to add your first ABI


a: add new • esc/q: back
```

## 12. Pagination - Page 2

User presses 'n' to go to next page.

```
ABI Management

Manage your contract ABIs

  > Governance Contract
    Name: Governance Contract
    Functions: 22 • Events: 6
    Created: 2024-11-01 09:00 AM

    Multisig Wallet
    Name: Multisig Wallet
    Functions: 16 • Events: 3
    Created: 2024-11-02 02:30 PM

    Token Vesting
    Name: Token Vesting
    Functions: 11 • Events: 2
    Created: 2024-11-03 10:45 AM

    Liquidity Pool
    Name: Liquidity Pool
    Functions: 20 • Events: 4
    Created: 2024-11-04 01:15 PM

    Oracle Aggregator
    Name: Oracle Aggregator
    Functions: 13 • Events: 3
    Created: 2024-11-05 04:00 PM


Page 2 of 3 • Showing 6-10 of 15 ABIs

Legend:
> = Selected

↑/k: up • ↓/j: down • enter: view details • a: add new • d: delete • e: edit
n: next page • p: previous page • esc/q: back
```

## 13. Pagination - Last Page

User presses 'n' again to go to last page.

```
ABI Management

Manage your contract ABIs

  > Bridge Contract
    Name: Bridge Contract
    Functions: 19 • Events: 5
    Created: 2024-11-06 08:30 AM

    Timelock Controller
    Name: Timelock Controller
    Functions: 15 • Events: 4
    Created: 2024-11-07 11:20 AM

    Auction House
    Name: Auction House
    Functions: 17 • Events: 6
    Created: 2024-11-08 03:45 PM

    Rewards Distributor
    Name: Rewards Distributor
    Functions: 12 • Events: 3
    Created: 2024-11-09 10:10 AM

    Price Oracle
    Name: Price Oracle
    Functions: 10 • Events: 2
    Created: 2024-11-10 02:55 PM


Page 3 of 3 • Showing 11-15 of 15 ABIs

Legend:
> = Selected

↑/k: up • ↓/j: down • enter: view details • a: add new • d: delete • e: edit
p: previous page • esc/q: back
```

## Summary of Key Features

### CRUD Operations

- **Create**: Add ABI via manual entry, URL import, or local file import
- **Read**: View ABI list, details, and individual methods with parameters
- **Update**: Edit ABI name (JSON is immutable, requires delete + recreate)
- **Delete**: Delete ABI with warning if contracts are using it

### Import Methods

1. **Manual Entry**: Type or paste ABI JSON directly
2. **Import from URL**: Fetch from Etherscan API, GitHub, or any JSON URL
3. **Import from Local File**: Load from filesystem (Hardhat/Foundry artifacts)

### ABI Format Support

- Solidity compiler array format: `[{type: "function", ...}, ...]`
- Hardhat/Foundry object format: `{abi: [...], bytecode: "0x..."}`

### User Experience

- Clear visual hierarchy with list navigation
- Detailed summaries showing function count, event count, etc.
- Method viewer with full parameter and return type information
- Comprehensive error handling with helpful messages
- Warning system when deleting ABIs linked to contracts
- Loading states for async operations (URL fetch)
