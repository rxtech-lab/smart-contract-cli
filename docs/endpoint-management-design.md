# Endpoint Management - Terminal UI Design

This document shows the mock terminal design for the endpoint management page.

## 1. Main Endpoint List View

```
Endpoint Management

Manage your network endpoints

  > Ethereum Mainnet
    URL: https://mainnet.infura.io/v3/abc123...
    Chain ID: 1
    Status: ✓ Active
    Created: 2024-10-15 10:30 AM

    Polygon Mainnet
    URL: https://polygon-rpc.com
    Chain ID: 137
    Status: ✓ Active
    Created: 2024-10-20 02:15 PM

  ★ Ethereum Sepolia (Default)
    URL: https://sepolia.infura.io/v3/abc123...
    Chain ID: 11155111
    Status: ✓ Active
    Created: 2024-10-25 09:45 AM

    Local Anvil
    URL: http://localhost:8545
    Chain ID: 31337
    Status: ✗ Unreachable
    Created: 2024-10-28 11:00 AM

    Arbitrum One
    URL: https://arb1.arbitrum.io/rpc
    Chain ID: 42161
    Status: ✓ Active
    Created: 2024-10-30 03:15 PM


Page 1 of 2 • Showing 5 of 8 endpoints

Legend:
> = Selected • ★ = Default endpoint

↑/k: up • ↓/j: down • enter: view details • a: add new • d: delete • s: set default
n: next page • p: previous page • esc/q: back
```

## 2. Add Endpoint - URL Input

User presses 'a' to add a new endpoint.

```
Add New Endpoint

Enter the RPC endpoint URL:

URL: https://mainnet.infura.io/v3/YOUR_API_KEY_


Examples:
• Infura: https://mainnet.infura.io/v3/YOUR_API_KEY
• Alchemy: https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY
• Public: https://cloudflare-eth.com
• Local: http://localhost:8545

enter: verify connection • esc: cancel
```

## 3. Add Endpoint - Verifying Connection

After entering URL and pressing enter.

```
Add New Endpoint

Verifying connection to endpoint...

URL: https://mainnet.infura.io/v3/abc123...

⠋ Connecting to network...
⠋ Detecting chain ID...
⠋ Fetching network information...
```

## 4. Add Endpoint - Connection Confirmation

After successful connection, show detected information for user to confirm.

```
Add New Endpoint - Connection Verified

✓ Connection successful!

Detected Network Information:

Basic Information
• Chain ID: 1
• Network Name: Ethereum Mainnet
• Currency: ETH

Current Status
• Latest Block: 18,234,567
• Block Time: ~12 seconds
• Gas Price: 25 gwei
• Network Latency: 245ms

Endpoint Details
• URL: https://mainnet.infura.io/v3/abc123...
• Protocol: HTTPS
• Connection: Stable

Enter a name for this endpoint (press Enter to use detected name):
Name: Ethereum Mainnet_


enter: save • esc: cancel
```

## 5. Add Endpoint - Custom Name

User can override the detected network name.

```
Add New Endpoint - Connection Verified

✓ Connection successful!

Detected Network Information:

Basic Information
• Chain ID: 1
• Network Name: Ethereum Mainnet
• Currency: ETH

Current Status
• Latest Block: 18,234,567
• Block Time: ~12 seconds
• Gas Price: 25 gwei
• Network Latency: 245ms

Endpoint Details
• URL: https://mainnet.infura.io/v3/abc123...
• Protocol: HTTPS
• Connection: Stable

Enter a name for this endpoint (press Enter to use detected name):
Name: My Infura Mainnet_


enter: save • esc: cancel
```

## 6. Add Endpoint - Success

After saving the endpoint.

```
Add New Endpoint - Success

✓ Endpoint saved successfully!

Name: My Infura Mainnet
Chain ID: 1
Network: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...

This endpoint is now available for use with your contracts.


Press any key to return to endpoint list...
```

## 7. Add Endpoint - Connection Failed

If connection verification fails.

```
Add New Endpoint - Connection Failed

✗ Failed to connect to endpoint

URL: https://mainnet.infura.io/v3/invalid-key

Error: Authentication required

Details:
• HTTP Status: 401 Unauthorized
• Response: Invalid project ID

Possible reasons:
• Invalid or expired API key
• Incorrect endpoint URL
• Network connectivity issues
• Endpoint service is down

What would you like to do?

  > Try different URL
    Edit the endpoint URL

    Save anyway (not recommended)
    Save endpoint without verification

    Cancel
    Return to endpoint list

↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 8. Add Endpoint - Network Error

Different error scenario - network timeout.

```
Add New Endpoint - Connection Failed

✗ Failed to connect to endpoint

URL: http://localhost:8545

Error: Connection timeout after 30 seconds

Details:
• Connection refused
• No response from server

Possible reasons:
• Local node is not running
• Firewall blocking the connection
• Wrong port number
• Network interface not accessible

What would you like to do?

  > Try different URL
    Edit the endpoint URL

    Save anyway (not recommended)
    Save endpoint without verification

    Cancel
    Return to endpoint list

↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 9. Add Endpoint - Unexpected Chain ID

When detected chain ID doesn't match common networks.

```
Add New Endpoint - Connection Verified

⚠ Unknown network detected

Detected Network Information:

Basic Information
• Chain ID: 31337
• Network Name: Unknown
• Currency: Unknown

Current Status
• Latest Block: 1,234
• Block Time: Unknown
• Gas Price: 0 gwei
• Network Latency: 12ms

Endpoint Details
• URL: http://localhost:8545
• Protocol: HTTP
• Connection: Stable

⚠ This appears to be a custom or local network (e.g., Anvil, Hardhat, Ganache)

Enter a name for this endpoint:
Name: Local Anvil_


enter: save • esc: cancel
```

## 10. View Endpoint Details

User presses 'enter' on an endpoint in the main list.

```
Endpoint Details - Ethereum Mainnet

Name: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...
Chain ID: 1
Network: Ethereum Mainnet
Status: ✓ Active
Default: No

Network Information
• Currency Symbol: ETH
• Block Explorer: https://etherscan.io
• Latest Block: 18,234,567
• Gas Price: 25 gwei

Connection Stats
• Latency: 245ms
• Success Rate: 99.8%
• Last Checked: 2024-11-01 10:45:30 AM

Timestamps
• Created: 2024-10-15 10:30:15 AM
• Last Modified: 2024-10-15 10:30:15 AM
• Last Used: 2024-11-01 09:15:22 AM

What would you like to do?

  > Test connection
    Verify endpoint is still reachable

    Set as default
    Make this the default endpoint

    Edit name
    Change the endpoint name

    Back to list

↑/k: up • ↓/j: down • enter: select • esc/q: back
```

## 11. Test Endpoint Connection

User selects "Test connection" from endpoint details.

```
Test Endpoint Connection

Testing connection to endpoint...

Endpoint: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...

⠋ Pinging endpoint...
⠋ Fetching chain ID...
⠋ Getting latest block...
⠋ Checking gas price...
```

## 12. Test Connection - Success

```
Test Endpoint Connection - Results

✓ Connection successful!

Network Information
• Chain ID: 1 (Ethereum Mainnet)
• Network Version: 1
• Latest Block: 18,234,589
• Block Timestamp: 2024-11-01 10:47:12 AM

Connection Quality
• Response Time: 235ms (Good)
• Ping: 120ms
• Connection: Stable

Gas Information
• Current Gas Price: 26 gwei
• Suggested Base Fee: 24 gwei
• Priority Fee: 2 gwei

Endpoint is healthy and ready to use!


Press any key to return...
```

## 13. Test Connection - Degraded Performance

```
Test Endpoint Connection - Results

⚠ Connection successful with issues

Network Information
• Chain ID: 1 (Ethereum Mainnet)
• Network Version: 1
• Latest Block: 18,234,589
• Block Timestamp: 2024-11-01 10:47:12 AM

Connection Quality
• Response Time: 4,523ms (Slow)
• Ping: 2,150ms
• Connection: Unstable

Gas Information
• Current Gas Price: Unable to fetch
• Error: Request timeout

⚠ Warning: Endpoint is responding slowly. Consider using a different endpoint
or checking your network connection.


Press any key to return...
```

## 14. Test Connection - Failed

```
Test Endpoint Connection - Results

✗ Connection failed

Error: Failed to connect to endpoint after 30 seconds

Details:
• Endpoint: https://mainnet.infura.io/v3/abc123...
• Error Type: Network timeout
• Attempts: 3

Possible reasons:
• Endpoint is down or unavailable
• API key expired or invalid
• Network connectivity issues
• Firewall or proxy blocking the connection

Suggestions:
• Check your internet connection
• Verify API key is still valid
• Try a different endpoint
• Contact the endpoint provider


Press any key to return...
```

## 15. Set Default Endpoint

User selects "Set as default" or presses 's' on endpoint.

```
Set Default Endpoint

Current default: Ethereum Sepolia

Are you sure you want to set this as the default endpoint?

New default: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...
Chain ID: 1

The default endpoint will be used when creating new contracts and for
general network operations.

  > Yes, set as default
    No, cancel

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 16. Set Default Endpoint - Success

```
Set Default Endpoint - Success

✓ Default endpoint updated!

New default endpoint: Ethereum Mainnet
Chain ID: 1
URL: https://mainnet.infura.io/v3/abc123...

This endpoint will now be used by default for new contracts and operations.


Press any key to return to endpoint list...
```

## 17. Edit Endpoint Name

User selects "Edit name" from endpoint details.

```
Edit Endpoint Name

Current name: Ethereum Mainnet

New name: My Infura Mainnet Node_


Note: Only the name can be edited. To change the URL or other settings,
delete this endpoint and create a new one.

enter: save • esc: cancel
```

## 18. Edit Endpoint Name - Success

```
Edit Endpoint Name - Success

✓ Endpoint name updated!

Old name: Ethereum Mainnet
New name: My Infura Mainnet Node


Press any key to return...
```

## 19. Delete Endpoint Confirmation

User presses 'd' to delete an endpoint.

```
Delete Endpoint

Are you sure you want to delete this endpoint?

Name: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...
Chain ID: 1

This action cannot be undone.

  > No, cancel
    Yes, delete

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 20. Delete Endpoint - Cannot Delete Default

If user tries to delete the default endpoint.

```
Delete Endpoint - Error

✗ Cannot delete default endpoint

The endpoint "Ethereum Sepolia" is currently set as the default endpoint
and cannot be deleted.

To delete this endpoint:
1. Set a different endpoint as default
2. Then delete this endpoint

Current default: Ethereum Sepolia ★


Press any key to return...
```

## 21. Delete Endpoint - Has Active Contracts

If contracts are using this endpoint.

```
Delete Endpoint

Are you sure you want to delete this endpoint?

Name: Ethereum Mainnet
URL: https://mainnet.infura.io/v3/abc123...
Chain ID: 1

⚠ Warning: 5 contracts are using this endpoint:
  • USDC Token Contract (0x1234...5678)
  • DAI Token Contract (0xabcd...efgh)
  • Uniswap Router (0x9876...4321)
  • My NFT Collection (0x5678...9012)
  • Custom Contract (0x3456...7890)

Deleting this endpoint will make these contracts unable to connect to the
network until you assign them a different endpoint.

  > No, cancel
    Yes, delete anyway

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 22. Delete Endpoint - Success

```
Delete Endpoint - Success

✓ Endpoint deleted successfully!

The endpoint "Ethereum Mainnet" has been removed.

⚠ Note: 5 contracts were using this endpoint. You may need to update
their network configuration.


Press any key to return to endpoint list...
```

## 23. Empty State

When no endpoints are configured yet.

```
Endpoint Management

No endpoints found

You haven't added any network endpoints yet. Endpoints are required to
connect to blockchain networks and interact with smart contracts.

Press 'a' to add your first endpoint


a: add new • esc/q: back
```

## 24. Add Endpoint - Testnet Detection

When a testnet is detected during verification.

```
Add New Endpoint - Connection Verified

✓ Connection successful!

Detected Network Information:

Basic Information
• Chain ID: 11155111
• Network Name: Ethereum Sepolia Testnet
• Currency: ETH (Testnet)

🧪 Testnet Network Detected

Current Status
• Latest Block: 5,234,567
• Block Time: ~12 seconds
• Gas Price: 2 gwei
• Network Latency: 189ms
• Faucet: https://sepoliafaucet.com

Endpoint Details
• URL: https://sepolia.infura.io/v3/abc123...
• Protocol: HTTPS
• Connection: Stable

⚠ This is a test network. Do not use real funds.

Enter a name for this endpoint (press Enter to use detected name):
Name: Ethereum Sepolia_


enter: save • esc: cancel
```

## 25. Add Endpoint - Multiple Same Chain ID

When adding an endpoint with a chain ID that already exists.

```
Add New Endpoint - Duplicate Chain ID Warning

✓ Connection successful!

Detected Network Information:

Basic Information
• Chain ID: 1
• Network Name: Ethereum Mainnet
• Currency: ETH

⚠ You already have an endpoint for Chain ID 1:
  • Name: My Infura Mainnet
  • URL: https://mainnet.infura.io/v3/old-key...

You can have multiple endpoints for the same network (e.g., as backups or
different providers).

Do you want to continue adding this endpoint?

  > Yes, add anyway
    No, cancel

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## Summary of Key Features

### CRUD Operations
- **Create**: Add endpoint with automatic network detection
- **Read**: View endpoint list and detailed information
- **Update**: Edit endpoint name only (URL is immutable)
- **Delete**: Remove endpoint with warnings for default/in-use endpoints

### Network Detection & Verification
- Automatic chain ID detection
- Network name resolution (Mainnet, testnets, etc.)
- Latest block number and timestamp
- Current gas price information
- Network latency measurement
- Connection quality assessment

### Confirmation Screen Details
- Chain ID and network name
- Latest block number
- Block time estimate
- Current gas price
- Network latency/response time
- Protocol (HTTP/HTTPS)
- Connection stability
- Testnet detection with warning
- Block explorer URL (for known networks)

### Connection Testing
- Test network connectivity on demand
- Verify endpoint is still reachable
- Check chain ID matches
- Measure response time and latency
- Get current gas prices
- Display helpful error messages

### Default Endpoint
- Mark one endpoint as default
- Prevent deletion of default endpoint
- Easy switching between endpoints
- Visual indicator (★) in list

### Error Handling
- Network timeout errors
- Authentication/API key errors
- Invalid URL format
- Connection refused (local node not running)
- Unexpected/unknown chain IDs
- Rate limiting errors
- General network errors

### User Experience
- Real-time connection verification during setup
- Loading states for async operations
- Color-coded status indicators (✓ Active, ✗ Unreachable)
- Performance warnings (slow response times)
- Helpful suggestions for fixing issues
- Testnet detection and warnings
- Duplicate chain ID warnings
- Contract dependency warnings before deletion
- Block explorer links for known networks
