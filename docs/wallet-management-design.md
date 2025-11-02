# Wallet Management - Terminal UI Design

This document shows the mock terminal design for the wallet management page.

## 1. Main Wallet List View

```
Wallet Management

Manage your wallets

  ★ Main Wallet
    Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
    Balance: 1,234.5678 ETH
    Status: Selected

    Dev Wallet
    Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
    Balance: 456.7890 ETH
    Status: Available

    Testing Account
    Address: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
    Balance: 89.1234 ETH
    Status: Available

    Cold Storage
    Address: 0x90F79bf6EB2c4f870365E785982E1f101E93b906
    Balance: 10,000.0000 ETH
    Status: Available

    Trading Bot
    Address: 0x15d34AAf54267DB7D7c367839AAf71A00a2C6A65
    Balance: 0.0000 ETH
    Status: Available


Endpoint: http://localhost:8545 (Anvil)

Legend:
★ = Currently selected wallet
> = Cursor position

↑/k: up • ↓/j: down • enter: actions • a: add new wallet • esc/q: back
```

## 2. Wallet Actions Menu

User presses 'enter' on a wallet (not the currently selected one).

```
Wallet Actions - Dev Wallet

Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
Balance: 456.7890 ETH

What would you like to do?

  > Select as active wallet
    Make this the current active wallet for transactions

    View details
    View full wallet information and transaction history

    Update wallet
    Update wallet alias or private key

    Delete wallet
    Remove this wallet from the system


↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 3. Wallet Actions Menu (Currently Selected Wallet)

User presses 'enter' on the currently selected wallet.

```
Wallet Actions - Main Wallet ★

Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
Balance: 1,234.5678 ETH

This is your currently selected wallet.

What would you like to do?

  > View details
    View full wallet information and transaction history

    Update wallet
    Update wallet alias or private key

    Delete wallet
    Remove this wallet from the system


Note: You cannot deselect the active wallet. Select another wallet first.

↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 4. Add New Wallet - Import Method Selection

User presses 'a' to add a new wallet.

```
Add New Wallet

How would you like to import the wallet?

  > Import from private key
    Import wallet using a private key (hex format)

    Import from mnemonic phrase
    Import wallet using a 12 or 24 word mnemonic phrase

    Generate new wallet
    Create a new random wallet with private key


↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 5a. Add Wallet - Private Key Import (Step 1: Alias)

User selects "Import from private key".

```
Add New Wallet - Private Key Import

Step 1/3: Enter Wallet Alias

Give your wallet a memorable name:

Alias: My New Wallet_


enter: next • esc: cancel
```

## 5b. Add Wallet - Private Key Import (Step 2: Private Key)

After entering alias.

```
Add New Wallet - Private Key Import

Step 2/3: Enter Private Key

Enter your private key (hex format, with or without 0x prefix):

Private Key: ****************************************************************_


⚠ Warning: Never share your private key with anyone!
Keep it secure and private.

enter: next • esc: cancel
```

## 5c. Add Wallet - Private Key Import (Step 3: Confirmation)

After entering private key, fetching balance from selected endpoint.

```
Add New Wallet - Confirmation

✓ Wallet successfully imported!

Wallet Details:
• Alias: My New Wallet
• Address: 0x1234567890123456789012345678901234567890
• Balance: 0.0000 ETH (on http://localhost:8545)

Derived Information:
• Public Key: 0x04a8f2c... (compressed)
• Checksum Address: ✓ Valid


  > Save wallet
    Add this wallet to your account

    Cancel
    Discard and start over


↑/k: up • ↓/j: down • enter: confirm
```

## 6a. Add Wallet - Mnemonic Import (Step 1: Alias)

User selects "Import from mnemonic phrase".

```
Add New Wallet - Mnemonic Import

Step 1/4: Enter Wallet Alias

Give your wallet a memorable name:

Alias: Recovery Wallet_


enter: next • esc: cancel
```

## 6b. Add Wallet - Mnemonic Import (Step 2: Mnemonic)

After entering alias.

```
Add New Wallet - Mnemonic Import

Step 2/4: Enter Mnemonic Phrase

Enter your 12 or 24 word mnemonic phrase (space-separated):

Mnemonic:
────────────────────────────────────────────────────────────────────────────
witch collapse practice feed shame open despair creek road again ice least_




────────────────────────────────────────────────────────────────────────────

Words entered: 12/12 ✓

⚠ Warning: Never share your mnemonic phrase with anyone!

enter: next • esc: cancel
```

## 6c. Add Wallet - Mnemonic Import (Step 3: Derivation Path)

After entering mnemonic.

```
Add New Wallet - Mnemonic Import

Step 3/4: Select Derivation Path

Choose the derivation path for your wallet:

  > m/44'/60'/0'/0/0
    Ethereum standard (default)

    m/44'/60'/0'/0/1
    Ethereum standard (account 1)

    m/44'/60'/0'/0/2
    Ethereum standard (account 2)

    Custom path
    Enter your own derivation path


Tip: Most wallets (MetaMask, Ledger, Trezor) use m/44'/60'/0'/0/0

↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 6d. Add Wallet - Mnemonic Import (Step 4: Confirmation)

After selecting derivation path.

```
Add New Wallet - Confirmation

✓ Wallet successfully imported from mnemonic!

Wallet Details:
• Alias: Recovery Wallet
• Derivation Path: m/44'/60'/0'/0/0
• Address: 0xabcdefabcdefabcdefabcdefabcdefabcdefabcd
• Balance: 25.7890 ETH (on http://localhost:8545)

Derived Information:
• Private Key: Available (hidden for security)
• Public Key: 0x04b7f3d... (compressed)
• Checksum Address: ✓ Valid


  > Save wallet
    Add this wallet to your account

    Cancel
    Discard and start over


↑/k: up • ↓/j: down • enter: confirm
```

## 7a. Add Wallet - Generate New (Step 1: Alias)

User selects "Generate new wallet".

```
Add New Wallet - Generate New

Step 1/3: Enter Wallet Alias

Give your wallet a memorable name:

Alias: Fresh Wallet_


enter: next • esc: cancel
```

## 7b. Add Wallet - Generate New (Step 2: Generating)

After entering alias.

```
Add New Wallet - Generate New

Step 2/3: Generating Wallet

Generating secure random wallet...

⠋ Creating cryptographic keys...
```

## 7c. Add Wallet - Generate New (Step 3: Backup)

After generation completes.

```
Add New Wallet - Backup Your Wallet

✓ Wallet successfully generated!

⚠ IMPORTANT: Save these credentials securely!

Mnemonic Phrase (12 words):
────────────────────────────────────────────────────────────────────────────
abandon ability able about above absent absorb abstract absurd abuse access
accident
────────────────────────────────────────────────────────────────────────────

Private Key:
──────────────��─────────────────────────────────────────────────────────────
0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef
────────────────────────────────────────────────────────────────────────────

Address: 0x9876543210987654321098765432109876543210

⚠ Write these down and store them in a secure location!
⚠ Anyone with access to these can control your funds!
⚠ Lost credentials cannot be recovered!


  > I have saved my credentials securely
    Confirm and add wallet to account

    Cancel
    Discard this wallet


↑/k: up • ↓/j: down • enter: confirm
```

## 8. View Wallet Details

User selects "View details" from actions menu.

```
Wallet Details - Main Wallet

Alias: Main Wallet
Status: ★ Currently Selected

Address Information:
• Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
• Checksum: ✓ Valid Ethereum address
• ENS Name: mainwallet.eth (if available)

Balance Information:
• Balance: 1,234.5678 ETH
• USD Value: $2,469,135.60 (at $2,000/ETH)
• Endpoint: http://localhost:8545 (Anvil)
• Last Updated: 2024-11-10 3:45 PM

Security:
• Private Key: ******** (hidden)
• Created: 2024-10-01 9:00 AM
• Last Modified: Never

Transaction Statistics:
• Total Transactions: 156
• Total Sent: 450.1234 ETH
• Total Received: 1,684.6912 ETH
• Last Activity: 2024-11-10 2:30 PM


r: refresh balance • p: show private key • esc/q: back
```

## 9. Update Wallet - Update Selection

User selects "Update wallet" from actions menu.

```
Update Wallet - Dev Wallet

Current Information:
• Alias: Dev Wallet
• Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8

What would you like to update?

  > Update alias
    Change the wallet name

    Update private key
    Replace the private key (will change address)

    Cancel
    Go back without changes


↑/k: up • ↓/j: down • enter: select • esc: cancel
```

## 10a. Update Wallet - Update Alias

User selects "Update alias".

```
Update Wallet - Update Alias

Current alias: Dev Wallet

Enter new alias:

New Alias: Development Wallet_


enter: save • esc: cancel
```

## 10b. Update Wallet - Update Private Key (Warning)

User selects "Update private key".

```
Update Wallet - Update Private Key

⚠ Warning: Changing Private Key

Updating the private key will:
• Change the wallet address
• This wallet will control a different Ethereum account
• You will lose access to the old address funds

Current Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8

Are you sure you want to continue?

  > No, cancel
    Keep the current private key

    Yes, update private key
    I understand the consequences


↑/k: up • ↓/j: down • enter: confirm
```

## 10c. Update Wallet - Update Private Key (Input)

User confirms the warning.

```
Update Wallet - Update Private Key

Alias: Dev Wallet

Enter new private key (hex format, with or without 0x prefix):

Private Key: ****************************************************************_


⚠ This will replace the existing private key and change the address!

enter: save • esc: cancel
```

## 10d. Update Wallet - Update Confirmation

After updating private key.

```
Update Wallet - Success

✓ Wallet updated successfully!

Changes:
• Alias: Dev Wallet (unchanged)
• Old Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
• New Address: 0x1111111111111111111111111111111111111111
• New Balance: 0.0000 ETH (on http://localhost:8545)

⚠ Important: The old address is no longer accessible with this wallet!


Press any key to return to wallet list...
```

## 11. Delete Wallet Confirmation (Non-Selected Wallet)

User selects "Delete wallet" from actions menu (for a non-selected wallet).

```
Delete Wallet

Are you sure you want to delete this wallet?

Wallet Information:
• Alias: Testing Account
• Address: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
• Balance: 89.1234 ETH

⚠ Warning: This action cannot be undone!

Make sure you have:
• Backed up your private key or mnemonic
• Transferred funds to another wallet
• No pending transactions

  > No, cancel
    Keep this wallet

    Yes, delete permanently
    Remove this wallet from the system


↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 12. Delete Wallet Confirmation (Currently Selected Wallet)

User tries to delete the currently selected wallet.

```
Delete Wallet

Cannot delete currently selected wallet!

Wallet Information:
• Alias: Main Wallet ★
• Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
• Balance: 1,234.5678 ETH
• Status: Currently Selected

To delete this wallet:
1. Select another wallet as active
2. Return to this wallet
3. Then delete it

This prevents accidentally losing access to your active wallet.


Press any key to go back...
```

## 13. Select Wallet Confirmation

User selects "Select as active wallet" from actions menu.

```
Select Active Wallet

Switch active wallet?

Current Active Wallet:
• Alias: Main Wallet ★
• Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
• Balance: 1,234.5678 ETH

New Active Wallet:
• Alias: Dev Wallet
• Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
• Balance: 456.7890 ETH

All future transactions will use the new active wallet.

  > Yes, switch wallet
    Make Dev Wallet the active wallet

    No, cancel
    Keep Main Wallet as active


↑/k: up • ↓/j: down • enter: confirm • esc: cancel
```

## 14. Select Wallet Success

After confirming wallet selection.

```
Select Active Wallet

✓ Active wallet changed successfully!

Your active wallet is now:
• Alias: Dev Wallet ★
• Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
• Balance: 456.7890 ETH

All transactions will now use this wallet.


Press any key to return to wallet list...
```

## 15. Empty State

When no wallets are stored yet.

```
Wallet Management

No wallets found

You haven't added any wallets yet. Wallets are required to sign
transactions and interact with smart contracts.

Get started by:
• Importing an existing wallet with private key or mnemonic
• Generating a new wallet

Press 'a' to add your first wallet


a: add new • esc/q: back
```

## 16. Show Private Key (Security Prompt)

User presses 'p' in wallet details to show private key.

```
Show Private Key - Security Warning

⚠ WARNING: Exposing Sensitive Information!

You are about to reveal the private key for:
• Alias: Main Wallet
• Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

⚠ Anyone with this private key can access and control your funds!
⚠ Make sure no one is watching your screen!
⚠ Be careful when sharing screenshots or recordings!

Type "SHOW" to reveal the private key (case sensitive):

Confirmation: _


enter: confirm • esc: cancel
```

## 17. Show Private Key (Revealed)

After typing "SHOW" correctly.

```
Show Private Key - Main Wallet

⚠ SENSITIVE INFORMATION - KEEP SECURE! ⚠

Wallet: Main Wallet
Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

Private Key:
────────────────────────────────────────────────────────────────────────────
0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
────────────────────────────────────────────────────────────────────────────

⚠ NEVER share this with anyone!
⚠ Anyone with this key can control your funds!

This screen will automatically close in 60 seconds...


c: copy to clipboard • esc/q: close immediately
```

## 18. Import Error - Invalid Private Key

When user enters an invalid private key.

```
Add New Wallet - Error

✗ Failed to import wallet

Error: Invalid private key format

Details:
• Private key must be 64 hexadecimal characters
• Can optionally start with "0x" prefix
• Contains only characters 0-9 and a-f

Examples of valid private keys:
• 0x1234...abcd (with 0x prefix)
• 1234...abcd (without prefix)


Press any key to go back and try again...
```

## 19. Import Error - Invalid Mnemonic

When user enters an invalid mnemonic phrase.

```
Add New Wallet - Error

✗ Failed to import wallet

Error: Invalid mnemonic phrase

Details:
• Mnemonic must be 12 or 24 words
• Words must be from the BIP39 word list
• Check for typos or extra spaces

You entered: 11 words (expected 12 or 24)

Tip: Common issues:
• Extra spaces between words
• Typos in words
• Missing words at the end


Press any key to go back and try again...
```

## 20. Import Error - Duplicate Wallet

When user tries to import a wallet that already exists.

```
Add New Wallet - Error

✗ Wallet already exists

Error: This wallet is already in your account

Existing Wallet:
• Alias: Main Wallet
• Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

The private key you entered derives to the same address as an
existing wallet.

Options:
• Use the existing wallet
• Import with a different private key


Press any key to go back...
```

## 21. Balance Fetch Error

When balance cannot be fetched from endpoint.

```
Wallet Management

Manage your wallets

  ★ Main Wallet
    Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
    Balance: unavailable ⚠
    Status: Selected

    Dev Wallet
    Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
    Balance: unavailable ⚠
    Status: Available


Endpoint: http://localhost:8545 (Anvil)
⚠ Connection Error: Failed to fetch balances from endpoint

Legend:
★ = Currently selected wallet
> = Cursor position

↑/k: up • ↓/j: down • enter: actions • a: add new wallet • r: retry • esc/q: back
```

## 22. Wallet List with Different Colors

Visual representation showing the selected wallet highlighted differently.

```
Wallet Management

Manage your wallets

  ★ Main Wallet                                           [HIGHLIGHTED IN GREEN]
    Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
    Balance: 1,234.5678 ETH
    Status: Selected

    Dev Wallet                                            [NORMAL WHITE]
    Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
    Balance: 456.7890 ETH
    Status: Available

  > Testing Account                                       [CURSOR HIGHLIGHTED]
    Address: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
    Balance: 89.1234 ETH
    Status: Available


Endpoint: http://localhost:8545 (Anvil)

Legend:
★ = Currently selected wallet (shown in green/primary color)
> = Cursor position (shown with selection highlight)

↑/k: up • ↓/j: down • enter: actions • a: add new wallet • esc/q: back
```

## Summary of Key Features

### CRUD Operations
- **Create**: Import wallet via private key, mnemonic phrase, or generate new
- **Read**: View wallet list, details, balance, and transaction statistics
- **Update**: Edit wallet alias or replace private key
- **Delete**: Remove wallet (with protection for currently selected wallet)

### Import Methods
1. **Private Key Import**: Import using hex-encoded private key
2. **Mnemonic Import**: Import using 12 or 24 word BIP39 mnemonic with derivation path selection
3. **Generate New**: Create random wallet with automatic mnemonic generation

### Wallet Selection
- **Active Wallet**: One wallet marked as currently selected (shown with ★)
- **Visual Distinction**: Selected wallet displayed in different color (green/primary)
- **Switch Active**: Select another wallet to use for transactions
- **Protection**: Cannot delete currently selected wallet

### Security Features
- **Private Key Masking**: Private keys hidden by default with asterisks
- **Reveal Protection**: Requires typing "SHOW" to reveal private key
- **Auto-hide Timer**: Revealed private key auto-closes after 60 seconds
- **Backup Warnings**: Multiple warnings when generating new wallets
- **Deletion Warnings**: Confirmation required with balance information

### Balance Display
- **Real-time Balance**: Fetches balance from selected RPC endpoint
- **USD Conversion**: Shows estimated USD value based on current ETH price
- **Connection Status**: Shows endpoint status and handles errors gracefully
- **Refresh**: Manual refresh option when viewing wallet details

### User Experience
- **Color Coding**: Selected wallet shown in distinct color
- **Status Indicators**: Visual indicators for selected (★) and cursor (>) positions
- **Comprehensive Errors**: Detailed error messages with troubleshooting tips
- **Validation**: Input validation for private keys, mnemonics, and derivation paths
- **Empty State**: Helpful guidance when no wallets exist
- **Step-by-Step Flows**: Multi-step processes for import/generation with progress indication
