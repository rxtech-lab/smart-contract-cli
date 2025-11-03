# Contract Management - Terminal UI Design

This document shows the mock terminal design for the contract management page.

## 1. Main Contract List View

```
Contract Management

Manage your smart contracts

  > USDC Token Contract
    Address: 0x1234...5678
    ABI: ERC20 Token
    Network: Ethereum Mainnet
    Created: 2024-10-15 10:30 AM

    Uniswap V2 Router
    Address: 0xabcd...efgh
    ABI: Uniswap V2 Router
    Network: Ethereum Mainnet
    Created: 2024-10-20 02:15 PM

    My NFT Collection
    Address: 0x9876...4321
    ABI: NFT Collection
    Network: Polygon Mainnet
    Created: 2024-10-25 09:45 AM

    Staking Contract
    Address: 0x5555...6666
    ABI: Staking Contract
    Network: Ethereum Mainnet
    Created: 2024-10-28 11:15 AM

    DEX Router
    Address: 0x7777...8888
    ABI: Custom DEX Router
    Network: Arbitrum One
    Created: 2024-10-30 02:45 PM


Page 1 of 2 • Showing 5 of 10 contracts

Legend:
> = Selected

↑/k: up • ↓/j: down • enter: view/update • a: add new • d: delete
n: next page • p: previous page • esc/q: back
```

## 2. Add Contract - Name Input

User presses 'a' to add a new contract.

```
Add New Contract

Step 1/5: Enter contract name

Name: My Token Contract_


enter: next • esc: cancel
```

## 3. Add Contract - Address Input

```
Add New Contract

Step 2/5: Enter contract address

Address: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48_


Note: Must be a valid Ethereum address (42 characters starting with 0x)

enter: next • esc: cancel
```

## 4. Add Contract - Endpoint Selection

```
Add New Contract

Step 3/5: Select network endpoint

Search: _

  > Ethereum Mainnet
    URL: https://mainnet.infura.io/v3/abc123...
    Chain ID: 1
    Status: ✓ Active

    Polygon Mainnet
    URL: https://polygon-rpc.com
    Chain ID: 137
    Status: ✓ Active

  ★ Ethereum Sepolia (Default)
    URL: https://sepolia.infura.io/v3/abc123...
    Chain ID: 11155111
    Status: ✓ Active

    Arbitrum One
    URL: https://arb1.arbitrum.io/rpc
    Chain ID: 42161
    Status: ✓ Active


Type to search • ↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 4b. Add Contract - Endpoint Selection (No Endpoints Available)

If no endpoints exist in the system yet.

```
Add New Contract

Step 3/5: Select network endpoint

No endpoints found

You haven't added any network endpoints yet. Network endpoints are required
to connect to the blockchain.

What would you like to do?

  > Go to Endpoint Management
    Navigate to endpoint management to add an endpoint

    Cancel
    Return to contract list

↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 4c. Add Contract - Redirected to Endpoint Management

User selects "Go to Endpoint Management" and is redirected.

```
Endpoint Management

[Navigation context: Adding contract "My Token Contract" - Step 3/5]

Manage your network endpoints

  > Ethereum Mainnet
    URL: https://mainnet.infura.io/v3/abc123...
    Chain ID: 1
    Status: ✓ Active
    Created: 2024-10-15 10:30 AM

...

ℹ After adding an endpoint, you'll return to contract creation


↑/k: up • ↓/j: down • enter: view details • a: add new • esc/q: return to contract creation
```

## 5. Add Contract - ABI Selection (List View)

```
Add New Contract

Step 4/5: Select ABI to link

Search: _

  > ERC20 Token
    Functions: 9 • Events: 2

    Uniswap V2 Router
    Functions: 24 • Events: 0

    NFT Collection
    Functions: 12 • Events: 3

    Custom Contract ABI
    Functions: 8 • Events: 2


Type to search • ↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 6. Add Contract - ABI Selection (Search Active)

User types in the search bar to filter ABIs.

```
Add New Contract

Step 4/5: Select ABI to link

Search: erc20_

  > ERC20 Token
    Functions: 9 • Events: 2


Type to search • ↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 7. Add Contract - ABI Selection (No ABIs Available)

If no ABIs exist in the system yet.

```
Add New Contract

Step 4/5: Select ABI to link

No ABIs found

You haven't added any ABIs yet. ABIs are required to interact with smart
contracts.

What would you like to do?

  > Go to ABI Management
    Navigate to ABI management to add an ABI

    Import ABI inline
    Import an ABI now and continue

    Cancel
    Return to contract list

↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 7b. Add Contract - Redirected to ABI Management

User selects "Go to ABI Management" and is redirected.

```
ABI Management

[Navigation context: Adding contract "My Token Contract" - Step 4/5]

Manage your contract ABIs

  > ERC20 Token
    Name: ERC20 Token
    Functions: 9 • Events: 2
    Created: 2024-10-15 10:30 AM

...

ℹ After adding an ABI, you'll return to contract creation


↑/k: up • ↓/j: down • enter: view details • a: add new • esc/q: return to contract creation
```

## 8. Add Contract - Import ABI Inline (Method Selection)

User selects "Import ABI inline" from the no ABIs screen.

```
Import ABI

How would you like to import the ABI?

  > Enter manually
    Type or paste the ABI JSON directly

    Import from URL
    Fetch ABI from a remote URL (e.g., Etherscan)

    Import from local file
    Load ABI from a file on your system


↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 9. Add Contract - Import from URL

User selects "Import from URL".

```
Import ABI from URL

Enter the URL to fetch the ABI from:

URL: https://api.etherscan.io/api?module=contract&action=getabi&address=0xA0b_


Examples:
• Etherscan API: https://api.etherscan.io/api?module=contract&action=...
• Direct JSON: https://example.com/contracts/MyContract.json

enter: fetch • esc: cancel
```

## 10. Add Contract - Import from URL (Fetching)

```
Import ABI from URL

Fetching ABI from remote URL...

URL: https://api.etherscan.io/api?module=contract&action=getabi&address=...

⠋ Please wait...
```

## 11. Add Contract - Import Confirmation

After successful fetch, show summary and ask to confirm import.

```
Import ABI - Confirmation

✓ ABI successfully fetched!

Detected Information:
• Functions: 9
  - name() → string
  - symbol() → string
  - decimals() → uint8
  - totalSupply() → uint256
  - balanceOf(address) → uint256
  - transfer(address, uint256) → bool
  - allowance(address, address) → uint256
  - approve(address, uint256) → bool
  - transferFrom(address, address, uint256) → bool

• Events: 2
  - Transfer(address indexed, address indexed, uint256)
  - Approval(address indexed, address indexed, uint256)

• Constructor: Yes
• Fallback/Receive: No

Enter a name for this ABI:
Name: USDC Token ABI_


enter: import and use • esc: cancel
```

## 12. Add Contract - Import Success

After successfully importing the ABI.

```
Import ABI - Success

✓ ABI imported successfully!

Name: USDC Token ABI
Functions: 9
Events: 2

This ABI has been added to your collection and linked to your contract.


Press any key to continue...
```

## 13. Add Contract - Import Error

If import fails.

```
Import ABI - Error

✗ Failed to import ABI

Error: Failed to fetch from URL: HTTP 404 Not Found

Possible reasons:
• Invalid URL or endpoint
• Network connection issues
• API rate limiting
• Invalid contract address


Press any key to go back and try again...
```

## 14. Add Contract - Final Confirmation

After selecting both endpoint and ABI, show a final confirmation screen.

```
Add New Contract

Step 5/5: Review and confirm

Contract Details:
• Name: My Token Contract
• Address: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

Endpoint:
• Ethereum Mainnet
• URL: https://mainnet.infura.io/v3/abc123...
• Chain ID: 1

ABI:
• USDC Token ABI
• Functions: 9
• Events: 2

Everything looks correct?

  > Yes, create contract
    No, go back

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 15. Add Contract - Success

After successfully creating a contract.

```
Add New Contract - Success

✓ Contract created successfully!

Name: My Token Contract
Address: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
Endpoint: Ethereum Mainnet
ABI: USDC Token ABI


Press any key to return to contract list...
```

## 16. View/Update Contract Details

User presses 'enter' on a contract in the main list.

```
Contract Details - USDC Token Contract

Name: USDC Token Contract
Address: 0x1234...5678
Full Address: 0x1234567890123456789012345678901234567890
Endpoint: Ethereum Mainnet
  URL: https://mainnet.infura.io/v3/abc123...
  Chain ID: 1
ABI: ERC20 Token
  Functions: 9 (6 view, 3 non-payable)
  Events: 2
Created: 2024-10-15 10:30 AM
Last Modified: 2024-10-15 10:30 AM

What would you like to do?

  > Interact with contract
    Call contract methods and functions

    Update ABI
    Change the linked ABI

    Update Endpoint
    Change the network endpoint

    View contract details
    See full contract information

    Test contract connection
    Verify contract is accessible

    Delete contract
    Remove this contract

    Back to list


↑/k: up • ↓/j: down • enter: select • d: delete • esc/q: back
```

## 17. Interact with Contract - Method List

User selects "Interact with contract" from contract details.

```
Interact with Contract - USDC Token Contract

Contract: 0x1234567890123456789012345678901234567890
Network: Ethereum Mainnet

Search methods: _

View Functions (6)
  > name() → string
    symbol() → string
    decimals() → uint8
    totalSupply() → uint256
    balanceOf(address account) → uint256
    allowance(address owner, address spender) → uint256

Write Functions (3)
    transfer(address to, uint256 amount) → bool
    approve(address spender, uint256 amount) → bool
    transferFrom(address from, address to, uint256 amount) → bool


Type to search • ↑/k: up • ↓/j: down • enter: call method • esc/q: back
```

## 18. Interact with Contract - Method List (With Search)

User types to filter methods.

```
Interact with Contract - USDC Token Contract

Contract: 0x1234567890123456789012345678901234567890
Network: Ethereum Mainnet

Search methods: transfer_

Write Functions (2)
  > transfer(address to, uint256 amount) → bool
    transferFrom(address from, address to, uint256 amount) → bool


Type to search • ↑/k: up • ↓/j: down • enter: call method • esc/q: back
```

## 19. Call View Function - No Parameters

User selects `name()` method.

```
Call Method - name()

Contract: USDC Token Contract
Method: name() → string
Type: View (Read-only)

This is a read-only function. No transaction will be sent.


  > Call function

↑/k: up • ↓/j: down • enter: execute • esc: cancel
```

## 20. Call View Function - Result

After calling the function.

```
Call Method - name()

Contract: USDC Token Contract
Method: name() → string
Type: View (Read-only)
Return value:
> "USD Coin"

-> Retry
-> Back
```

## 21. Call View Function - With Parameters (Step 1)

User selects `balanceOf(address account)` method.

```
Call Method - balanceOf(address)

Contract: USDC Token Contract
Method: balanceOf(address account) → uint256
Type: View (Read-only)

Enter function parameters:
Validation: ✓ Valid Ethereum address

Parameter 1 of 1
account (address):
0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb_

enter: call function • esc: cancel
```

## 22. Call View Function - With Parameters (Validation Error)

User enters invalid address.

```
Call Method - balanceOf(address)

Contract: USDC Token CONTRACT
Method: balanceOf(address account) → uint256
Type: View (Read-only)

Enter function parameters:
Validation: ✗ Invalid Ethereum address format
Expected: 42 characters starting with 0x (e.g., 0x1234...5678)

Parameter 1 of 1
account (address):
invalid_address_

enter: call function (disabled) • esc: cancel
```

## 23. Call View Function - With Parameters (Result)

After calling with valid parameters.

```
Call Method - balanceOf(address)

Contract: USDC Token Contract
Method: balanceOf(address account) → uint256
Type: View (Read-only)

Parameters:
• account: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb

✓ Function called successfully

Return value:
> 1000000000 (1,000 USDC - 6 decimals)


Press any key to go back...
```

## 24. Call Write Function - With Parameters (Step 1)

User selects `transfer(address to, uint256 amount)` method.

```
Call Method - transfer(address, uint256)

Contract: USDC Token Contract
Method: transfer(address to, uint256 amount) → bool
Type: Write (Sends transaction)

Enter function parameters:

> Parameter 1 of 2
  to (address):
  _

  Parameter 2 of 2
  amount (uint256):


↑/↓: navigate fields • enter: next/call • esc: cancel
```

## 25. Call Write Function - With Parameters (Step 2)

User fills in first parameter and moves to second.

```
Call Method - transfer(address, uint256)

Contract: USDC Token Contract
Method: transfer(address to, uint256 amount) → bool
Type: Write (Sends transaction)

Enter function parameters:

  Parameter 1 of 2
  to (address):
  0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
  Validation: ✓ Valid

> Parameter 2 of 2
  amount (uint256):
  1000000_

  Validation: ✓ Valid


↑/↓: navigate fields • enter: next/call • esc: cancel
```

## 26. Call Payable Function - With Value Field

User selects a payable function (hypothetical example).

```
Call Method - deposit()

Contract: Staking Contract
Method: deposit() → bool
Type: Payable (Requires ETH)

Enter function parameters:

No parameters required

> Value to send (ETH):
  0.1_

  Validation: ✓ Valid amount


↑/↓: navigate fields • enter: send transaction • esc: cancel
```

## 27. Call Write Function - Transaction Confirmation

After filling all parameters.

```
Send Transaction - Confirmation

Contract: USDC Token Contract
Method: transfer(address to, uint256 amount) → bool

Transaction Details:
• To: 0x1234567890123456789012345678901234567890 (Contract)
• From: 0xYourWalletAddress...
• Network: Ethereum Mainnet

Function Parameters:
• to: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
• amount: 1000000 (1 USDC)

Value: 0 ETH

Gas Estimate:
• Gas Limit: 65,000
• Gas Price: 25 gwei
• Max Fee: ~0.001625 ETH (~$3.25)

Total Cost: ~0.001625 ETH + 0 ETH = ~0.001625 ETH

⚠ This will send a transaction to the blockchain and cannot be undone.

  > Confirm and sign
    Cancel

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 28. Call Payable Function - Transaction Confirmation

For payable functions with value.

```
Send Transaction - Confirmation

Contract: Staking Contract
Method: deposit()

Transaction Details:
• To: 0x5555666677778888999900001111222233334444 (Contract)
• From: 0xYourWalletAddress...
• Network: Ethereum Mainnet

Function Parameters:
None

Value: 0.1 ETH (~$200)

Gas Estimate:
• Gas Limit: 45,000
• Gas Price: 25 gwei
• Max Fee: ~0.001125 ETH (~$2.25)

Total Cost: 0.1 ETH + ~0.001125 ETH = ~0.101125 ETH (~$202.25)

⚠ This will send 0.1 ETH to the contract and cannot be undone.

  > Confirm and sign
    Cancel

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 29. Send Transaction - Processing

After user confirms.

```
Send Transaction - Processing

Sending transaction to network...

Contract: USDC Token Contract
Method: transfer(address, uint256)

⠋ Signing transaction...
⠋ Broadcasting to network...
⠋ Waiting for confirmation...

Transaction Hash: 0xabc123def456...
```

## 30. Send Transaction - Success

After transaction is confirmed.

```
Send Transaction - Success

✓ Transaction confirmed!

Contract: USDC Token Contract
Method: transfer(address, uint256)

Transaction Details:
• Hash: 0xabc123def456789...
• Block: 18,234,567
• Gas Used: 52,341
• Actual Fee: 0.001309 ETH

Return Value:
> true

View on Etherscan: https://etherscan.io/tx/0xabc123def456789...


Press any key to go back...
```

## 31. Send Transaction - Failed

If transaction fails.

```
Send Transaction - Failed

✗ Transaction failed

Contract: USDC Token Contract
Method: transfer(address, uint256)

Error: Execution reverted: ERC20: transfer amount exceeds balance

Transaction Details:
• Hash: 0xabc123def456789...
• Block: 18,234,567
• Gas Used: 23,456 (all gas consumed)
• Actual Fee: 0.000587 ETH

Possible reasons:
• Insufficient token balance
• Contract execution reverted
• Invalid parameters

View on Etherscan: https://etherscan.io/tx/0xabc123def456789...


Press any key to go back...
```

## 32. Call Function - Multiple Parameters

User selects `transferFrom(address from, address to, uint256 amount)` method.

```
Call Method - transferFrom(address, address, uint256)

Contract: USDC Token Contract
Method: transferFrom(address from, address to, uint256 amount) → bool
Type: Write (Sends transaction)

Enter function parameters:

> Parameter 1 of 3
  from (address):
  0x1111222233334444555566667777888899990000_

  Validation: ✓ Valid

  Parameter 2 of 3
  to (address):
  0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
  Validation: ✓ Valid

  Parameter 3 of 3
  amount (uint256):
  500000
  Validation: ✓ Valid


↑/↓: navigate fields • enter: send transaction • esc: cancel
```

## 33. Update Contract ABI - Selection

User selects "Update ABI" from contract details.

```
Update ABI - USDC Token Contract

Current ABI: ERC20 Token
Functions: 9 • Events: 2

Select a new ABI:

Search: _

  > ERC20 Token (Current)
    Functions: 9 • Events: 2

    Uniswap V2 Router
    Functions: 24 • Events: 0

    NFT Collection
    Functions: 12 • Events: 3

    Custom Contract ABI
    Functions: 8 • Events: 2


Type to search • ↑/k: up • ↓/j: down • enter: select • i: import new • esc: cancel
```

## 34. Update Contract ABI - With Search

User types to filter ABIs.

```
Update ABI - USDC Token Contract

Current ABI: ERC20 Token
Functions: 9 • Events: 2

Select a new ABI:

Search: uniswap_

  > Uniswap V2 Router
    Functions: 24 • Events: 0


Type to search • ↑/k: up • ↓/j: down • enter: select • i: import new • esc: cancel
```

## 35. Update Contract Endpoint - Selection

User selects "Update Endpoint" from contract details.

```
Update Endpoint - USDC Token Contract

Current Endpoint: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...
Chain ID: 1

Select a new endpoint:

Search: _

  > Ethereum Mainnet (Current)
    URL: https://mainnet.infura.io/v3/abc123...
    Chain ID: 1
    Status: ✓ Active

    Polygon Mainnet
    URL: https://polygon-rpc.com
    Chain ID: 137
    Status: ✓ Active

  ★ Ethereum Sepolia (Default)
    URL: https://sepolia.infura.io/v3/abc123...
    Chain ID: 11155111
    Status: ✓ Active

    Arbitrum One
    URL: https://arb1.arbitrum.io/rpc
    Chain ID: 42161
    Status: ✓ Active


Type to search • ↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 36. Update Contract Endpoint - Confirmation

```
Update Endpoint - Confirmation

Are you sure you want to update the endpoint?

Contract: USDC Token Contract
Address: 0x1234...5678

Current Endpoint: Ethereum Mainnet
• URL: https://mainnet.infura.io/v3/abc123...
• Chain ID: 1

New Endpoint: Polygon Mainnet
• URL: https://polygon-rpc.com
• Chain ID: 137

⚠ Warning: Make sure the contract exists on the new network at this address.
Using an endpoint for a different network may result in connection errors.

  > No, cancel
    Yes, update

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 37. Update Contract Endpoint - Success

```
Update Endpoint - Success

✓ Endpoint updated successfully!

Contract: USDC Token Contract
New Endpoint: Polygon Mainnet
URL: https://polygon-rpc.com
Chain ID: 137


Press any key to return to contract details...
```

## 38. Update Contract ABI - Confirmation

```
Update ABI - Confirmation

Are you sure you want to update the ABI?

Contract: USDC Token Contract
Address: 0x1234...5678

Current ABI: ERC20 Token
• Functions: 9
• Events: 2

New ABI: Uniswap V2 Router
• Functions: 24
• Events: 0

⚠ Warning: Make sure the new ABI matches the contract at this address.
Using an incorrect ABI may result in errors when calling contract functions.

  > No, cancel
    Yes, update

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 39. Update Contract ABI - Success

```
Update ABI - Success

✓ ABI updated successfully!

Contract: USDC Token Contract
New ABI: Uniswap V2 Router
Functions: 24
Events: 0


Press any key to return to contract details...
```

## 40. View Contract Details (Full Info)

User selects "View contract details" from the contract menu.

```
Contract Information - USDC Token Contract

Basic Information
• Name: USDC Token Contract
• Address: 0x1234567890123456789012345678901234567890
• Endpoint: Ethereum Mainnet
  - URL: https://mainnet.infura.io/v3/...
  - Chain ID: 1

ABI Information
• Name: ERC20 Token
• Functions: 9 (6 view, 3 non-payable)
• Events: 2

Functions Available
  name() → string
  symbol() → string
  decimals() → uint8
  totalSupply() → uint256
  balanceOf(address account) → uint256
  transfer(address to, uint256 amount) → bool
  allowance(address owner, address spender) → uint256
  approve(address spender, uint256 amount) → bool
  transferFrom(address from, address to, uint256 amount) → bool

Timestamps
• Created: 2024-10-15 10:30:15 AM
• Last Modified: 2024-10-15 10:30:15 AM


esc/q: back
```

## 41. Delete Contract Confirmation

User presses 'd' to delete a contract.

```
Delete Contract

Are you sure you want to delete this contract?

Name: USDC Token Contract
Address: 0x1234...5678
Network: Ethereum Mainnet
ABI: ERC20 Token

⚠ This will remove the contract from your saved list. The contract on the
blockchain will not be affected.

This action cannot be undone.

  > No, cancel
    Yes, delete

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 42. Delete Contract Success

```
Delete Contract - Success

✓ Contract deleted successfully!

The contract "USDC Token Contract" has been removed from your list.


Press any key to return to contract list...
```

## 43. Empty State

When no contracts are stored yet.

```
Contract Management

No contracts found

You haven't added any contracts yet. Contracts represent deployed smart
contracts that you want to interact with.

Press 'a' to add your first contract


a: add new • esc/q: back
```

## 44. Test Contract Connection

User selects "Test contract connection" from contract menu.

```
Test Contract Connection

Testing connection to contract...

Contract: USDC Token Contract
Address: 0x1234567890123456789012345678901234567890
Endpoint: Ethereum Mainnet (https://mainnet.infura.io/v3/...)

⠋ Checking network connectivity...
```

## 45. Test Contract Connection - Success

```
Test Contract Connection - Results

✓ Connection successful!

Network Status
• Network: Ethereum Mainnet (Chain ID: 1)
• Block Number: 18234567
• Network Latency: 245ms

Contract Status
• Contract exists at address: Yes
• Code size: 5,432 bytes
• Is contract: Yes

Test Calls (using ABI)
• name(): "USD Coin"
• symbol(): "USDC"
• decimals(): 6


Press any key to return...
```

## 46. Test Contract Connection - Failure

```
Test Contract Connection - Results

✗ Connection failed

Error: Failed to connect to network endpoint

Details:
• Endpoint: Ethereum Mainnet
• URL: https://mainnet.infura.io/v3/...
• Error: Request timeout after 30 seconds

Possible reasons:
• Network endpoint is down or unavailable
• Network connectivity issues
• Invalid or expired API key
• Rate limiting

Suggestion: Check your endpoint configuration in Endpoint Management


Press any key to return...
```

## 47. Test Contract Connection - Contract Not Found

```
Test Contract Connection - Results

⚠ Contract not found at address

Network Status
• Network: Ethereum Mainnet (Chain ID: 1)
• Block Number: 18234567
• Network Latency: 245ms

Contract Status
• Contract exists at address: No
• Code size: 0 bytes
• Is contract: No

The address 0x1234567890123456789012345678901234567890 does not contain
a contract on Ethereum Mainnet.

Possible reasons:
• Wrong network selected
• Contract has been destroyed (selfdestruct)
• Invalid contract address
• Contract not yet deployed


Press any key to return...
```

## Summary of Key Features

### CRUD Operations

- **Create**: Add contract with name, address, endpoint, and ABI (5-step wizard)
- **Read**: View contract list and detailed information
- **Update**: Change linked ABI and endpoint with search functionality
- **Delete**: Remove contract from saved list

### Endpoint Selection

- Select from existing endpoints during contract creation
- Search/filter endpoints by name
- Show endpoint details (URL, chain ID, status)
- Navigate to Endpoint Management if no endpoints exist
- Update endpoint after creation with confirmation
- Warning when changing to different network

### ABI Linking

- Select from existing ABIs with search/filter
- Navigate to ABI Management or import inline if no ABIs exist
- Update ABI after creation with confirmation
- Show ABI details and function signatures

### Search/Filter Functionality

- Search bar for both ABI and endpoint selection
- Type to filter by name
- Up/down cursor to navigate filtered results
- Works seamlessly with pagination for large lists

### Navigation & Context Preservation

- Redirect to ABI/Endpoint Management when needed
- Show navigation context banner when redirected
- Return to contract creation after adding ABI/endpoint
- Preserve entered data (name, address) during navigation

### Import Flow

- Detect when user needs to import ABI (no ABIs available)
- Options: Navigate to ABI Management, import inline, or cancel
- Inline import during contract creation
- Show comprehensive summary of imported ABI
- Display all methods with names, parameters, and return types
- Support manual entry, URL import, and local file import

### Final Confirmation

- Review all details before creating contract
- Show contract details, endpoint info, and ABI summary
- Option to go back and make changes
- Clear success message after creation

### Contract Interaction

- **Method List**: View all contract functions grouped by type (View/Write)
- **Search/Filter**: Search methods by name for quick access
- **Syntax Highlighting**: Color-coded display of function signatures
  - Function name, parameters (with types), and return types
- **Call View Functions**:
  - Read-only methods with no transaction
  - Immediate results displayed
  - No gas fees
- **Call Write Functions**:
  - Parameter input with live validation
  - Navigate between fields with up/down arrows
  - Input validation with helpful error messages
  - Gas estimation before sending
- **Payable Functions**:
  - Value field to enter ETH amount
  - Clear indication of total cost (value + gas)
- **Transaction Confirmation**:
  - Review all details: parameters, value, gas estimate, total cost
  - Warning about irreversible action
  - Confirm and sign button
- **Transaction Processing**:
  - Real-time status updates (signing, broadcasting, confirming)
  - Transaction hash displayed
  - Link to block explorer
- **Result Display**:
  - Return values shown for all function calls
  - Success/failure status
  - Transaction details (hash, block, gas used)
  - Error messages with helpful suggestions

### Contract Testing

- Test network connectivity via configured endpoint
- Verify contract exists at address
- Make test function calls using ABI
- Show helpful error messages and suggestions

### User Experience

- Step-by-step 5-step contract creation wizard
- Clear validation and error messages
- Confirmation dialogs for destructive actions
- Search functionality for large ABI and endpoint lists
- Pagination for long lists
- Loading states for async operations
- Helpful suggestions when errors occur
- Context-aware navigation with preserved state
- Live parameter validation with type checking
- Intuitive navigation between input fields
- Real-time feedback for all operations
