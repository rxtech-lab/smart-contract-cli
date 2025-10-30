# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Call subagent to update the claude.md if needed

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
# Build the CLI (ALWAYS use this command)
make build

# Install dependencies
go mod download
go mod tidy

# Generate routes from app folder structure
make generate-routes
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
└── view/                # TUI (Terminal User Interface) layer
    ├── types.go         # Router interface and Route definitions
    └── router.go        # Router implementation for Bubble Tea navigation
```

### Key Design Patterns

**1. Interface-Based Abstractions**

- `Signer` interface: Supports multiple key management strategies (currently: private key)
- `Transport` interface: Allows different communication methods (currently: HTTP/JSON-RPC)
- `Router` interface: Enables different navigation strategies for TUI applications

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

## View Layer (TUI Router)

### Router Architecture

The view layer provides a navigation system for Bubble Tea TUI applications, centered around the Router pattern. The Router manages navigation between different views/screens while maintaining navigation history and supporting both path and query parameters.

**Key Components:**

- **Router Interface** (`internal/view/types.go`): Defines the contract for navigation management
- **RouterImplementation** (`internal/view/router.go`): Concrete implementation with full Bubble Tea integration
- **Route**: Defines a path pattern and its associated Bubble Tea component

### Router Features

**1. Route Management**

```go
AddRoute(route Route)                    // Add a new route to the router
SetRoutes(routes []Route)                // Replace all routes
RemoveRoute(path string)                 // Remove a route by path
GetRoutes() []Route                      // Get all registered routes
GetCurrentRoute() *Route                 // Get the currently active route
```

**2. Navigation Methods**

```go
NavigateTo(path string, queryParams map[string]string) error  // Navigate to a path with optional query params
ReplaceRoute(path string, queryParams map[string]string) error // Replace current route without adding to history
Back() error                                                   // Navigate to previous route
CanGoBack() bool                                              // Check if back navigation is possible
```

**3. Parameter Support**

- **Path Parameters**: Extract dynamic segments from URLs
  - Pattern: `/users/:id` matches `/users/123`
  - Pattern: `/posts/:postId/comments/:commentId` matches `/posts/42/comments/7`
  - Access via `GetParam(key string) string`

- **Query Parameters**: Pass key-value pairs in navigation
  - Example: `NavigateTo("/users", map[string]string{"filter": "active", "page": "2"})`
  - Access via `GetQueryParam(key string) string`

**4. Navigation Stack**

- Maintains history of visited routes for back navigation
- Each entry stores: route, query params, path params, and full path
- `Back()` pops the stack and restores previous route state
- `ReplaceRoute()` updates current route without growing the stack

**5. Bubble Tea Integration**

The Router implements the Bubble Tea `Model` interface:

```go
Init() tea.Cmd           // Initialize the router and current component
Update(tea.Msg) (tea.Model, tea.Cmd)  // Delegates to current component's Update
View() string            // Delegates to current component's View
```

### Route Matching Algorithm

1. **Exact Match**: First attempts exact path matching
2. **Pattern Match**: If no exact match, checks dynamic patterns
   - Converts patterns like `/users/:id` to regex: `^/users/([^/]+)$`
   - Extracts named parameters from matched segments
   - Supports multiple parameters in nested paths

### Usage Examples

**Basic Setup**

```go
// Create router and add routes
router := NewRouter()
router.AddRoute(Route{Path: "/", Component: NewHomeModel()})
router.AddRoute(Route{Path: "/users", Component: NewUserListModel()})
router.AddRoute(Route{Path: "/users/:id", Component: NewUserDetailModel()})

// Start with home route
router.NavigateTo("/", nil)
```

**Navigation with Path Parameters**

```go
// Navigate to user detail page
router.NavigateTo("/users/123", nil)

// Inside the component's Update method:
userId := router.GetParam("id")  // Returns "123"
```

**Navigation with Query Parameters**

```go
// Navigate with filters
queryParams := map[string]string{
    "filter": "active",
    "sort": "name",
    "page": "1",
}
router.NavigateTo("/users", queryParams)

// Inside component:
filter := router.GetQueryParam("filter")  // Returns "active"
page := router.GetQueryParam("page")      // Returns "1"
```

**Back Navigation**

```go
// In a component's Update method handling a "back" key press
if router.CanGoBack() {
    router.Back()
}
```

**Replace Route (No History)**

```go
// Replace login screen with dashboard after successful auth
router.ReplaceRoute("/dashboard", nil)
// User cannot navigate back to login screen
```

### Testing

- **Location**: `internal/view/router_test.go`
- **Framework**: Uses `testify/suite` pattern (project standard)
- **Coverage**: 27 comprehensive test cases including:
  - Route registration and retrieval
  - Exact path matching
  - Dynamic path parameter extraction (single and multiple params)
  - Complex nested path patterns
  - Query parameter handling
  - Navigation stack management
  - Back navigation scenarios
  - Edge cases (empty paths, missing routes, invalid patterns)
  - Bubble Tea lifecycle methods (Init, Update, View)

### Implementation Notes

- **RouterImplementation** is the concrete type; use `NewRouter()` factory function
- Navigation stack uses `routeEntry` structs internally to store complete route state
- Current component is cached and updated during navigation
- Pattern matching uses Go's `regexp` package for path parameter extraction
- All navigation methods return errors for invalid paths/states
- Router is safe for use in Bubble Tea's message-passing model

### Design Rationale

**Why a Router Pattern?**

- Centralizes navigation logic instead of scattering it across components
- Provides type-safe route definitions
- Enables deep linking and URL-based state management
- Simplifies testing by making navigation observable and controllable

**Interface-Based Design**

- Allows for alternative router implementations (e.g., declarative routing, middleware support)
- Facilitates testing through dependency injection
- Follows the project's pattern of interface-based abstractions (see Signer, Transport)

## Declarative UI Component System

### Overview

The component system (`internal/ui/component/`) provides a React/SwiftUI-like declarative approach to building TUI views. Instead of imperatively building strings with `strings.Builder`, you compose UI using nested components with chainable modifiers.

**Location:** `internal/ui/component/`

### Core Philosophy

**Before (Imperative):**
```go
func (m Model) View() string {
    var b strings.Builder
    b.WriteString("Select a blockchain:\n\n")
    for _, option := range options {
        marker := "  "
        if m.selectedOption.Value == option.Value {
            marker = "> "
        }
        b.WriteString(fmt.Sprintf("%s%s\n", marker, option.Label))
    }
    return b.String()
}
```

**After (Declarative):**
```go
func (m Model) View() string {
    items := make([]ListItem, len(options))
    for i, opt := range options {
        items[i] = Item(opt.Label, opt.Value)
    }

    return VStackC(
        T("Select a blockchain:").Bold(true).Primary(),
        SpacerV(1),
        Card(
            NewList(items).
                Selected(m.selectedOption.Value).
                SelectedPrefix("> ").
                UnselectedPrefix("  "),
        ),
        SpacerV(1),
        T("Selected: "+m.selectedOption.Label).Muted(),
    ).Render()
}
```

### Architecture

**Component Interface** (`component.go`):
```go
type Component interface {
    Render() string
}
```

All components implement `Render()` which returns the final string representation. Components are composable, immutable (return new instances on modification), and integrate seamlessly with `lipgloss` for styling.

### Core Components

#### 1. Text Component (`text.go`)

Rich text rendering with chainable modifiers:

```go
T("Hello").Bold(true).Primary()
T("Error message").Error()
T("Muted text").Muted()
T("Styled").
    Bold(true).
    Foreground(lipgloss.Color("205")).
    Padding(1, 2).
    Border(lipgloss.RoundedBorder())
```

**Preset Styles:**
- `.Primary()` - Bold, primary color
- `.Success()` - Green text
- `.Error()` - Bold red text
- `.Warning()` - Yellow text
- `.Info()` - Blue text
- `.Muted()` - Faint gray text
- `.Secondary()` - Faint text

**Modifiers:**
- Style: `Bold()`, `Italic()`, `Underline()`, `Strikethrough()`, `Faint()`, `Blink()`
- Color: `Foreground()`, `Background()`, `Color()`, `BgColor()`
- Layout: `Width()`, `Height()`, `MaxWidth()`, `MaxHeight()`, `Align()`
- Spacing: `Padding()`, `Margin()`, `PaddingAll()`, `MarginAll()`
- Border: `Border()`, `BorderStyle()`, `BorderTop()`, `BorderForeground()`

#### 2. Layout Components (`layout.go`)

**VStack** - Vertical stack (like SwiftUI):
```go
VStackC(
    T("Line 1"),
    T("Line 2"),
    T("Line 3"),
).Spacing(1).Render()

// Or with methods
NewVStack(
    T("Line 1"),
    T("Line 2"),
).Spacing(2).Align(lipgloss.Center)
```

**HStack** - Horizontal stack:
```go
HStackC(
    T("Left"),
    SpacerH(5),
    T("Right"),
).Spacing(2).Render()
```

**ZStack** - Overlay stack (layers components):
```go
ZStackC(
    T("Background"),
    T("Foreground"),
).Align(lipgloss.Center).Render()
```

**Spacer** - Flexible/fixed spacing:
```go
SpacerV(2)  // Vertical spacer (2 empty lines)
SpacerH(5)  // Horizontal spacer (5 spaces)
```

**Divider** - Horizontal line:
```go
DividerLine(40)  // 40-character divider
NewDivider("─", 20).WithStyle(style)
```

#### 3. List Components (`list.go`)

**Selectable List:**
```go
items := []ListItem{
    Item("Option 1", "opt1"),
    Item("Option 2", "opt2"),
}

NewList(items).
    Selected("opt1").
    SelectedPrefix("> ").
    UnselectedPrefix("  ").
    Spacing(1)
```

**Numbered List:**
```go
NewNumberedList(
    T("First item"),
    T("Second item"),
    T("Third item"),
).Start(1).Spacing(1)
```

**Bullet List:**
```go
NewBulletList(
    T("Item 1"),
    T("Item 2"),
).Bullet("• ").Spacing(1)
```

**String List** (convenience):
```go
StringList([]string{"A", "B", "C"}).Selected("B")
```

**Custom Item Rendering:**
```go
NewList(items).RenderItem(func(item ListItem, isSelected bool) Component {
    return HStackC(
        IfC(isSelected, T("→ "), T("  ")),
        T(item.GetLabel()).Bold(isSelected),
    )
})
```

#### 4. Container Components (`container.go`)

**Box** - Container with borders and styling:
```go
NewBox(T("Content")).
    RoundedBorder().
    BorderColor(lipgloss.Color("205")).
    PaddingAll(1).
    Width(40)
```

**Border Presets:**
- `.RoundedBorder()` - Rounded corners
- `.NormalBorder()` - Standard box border
- `.ThickBorder()` - Thick border
- `.DoubleBorder()` - Double-line border
- `.HiddenBorder()` - Invisible border (for spacing)

**Convenience Containers:**
```go
Card(content)   // Rounded border + padding
Panel(content)  // Normal border + padding
```

**Padding Container:**
```go
NewPadding(content).
    All(2).           // Equal padding all sides
    Vertical(1).      // Top and bottom
    Horizontal(2).    // Left and right
    Top(1).           // Individual sides
    UseMargins()      // Use margins instead of padding
```

**Center Container:**
```go
NewCenter(content, 80, 24)  // Center in 80x24 space
```

#### 5. Conditional Rendering (`conditional.go`)

**If/Else:**
```go
IfC(condition, T("True"), T("False"))
IfElse(isLoggedIn, T("Welcome!"), T("Login"))
```

**If/Then (no else):**
```go
IfThenC(showMessage, T("Message"))
When(isAdmin, T("Admin Panel"))
```

**Unless (inverse of IfThen):**
```go
UnlessC(isHidden, T("Visible content"))
```

**Switch Statement:**
```go
SwitchC(
    Case(Match(status, "success"), T("✓ Success").Success()),
    Case(Match(status, "error"), T("✗ Error").Error()),
    Case(Match(status, "pending"), T("⏳ Pending").Warning()),
).Default(T("Unknown"))
```

**Match Helpers:**
```go
Match(value, target)              // Equality check
MatchAny(value, "a", "b", "c")   // Multiple values
MatchRange(value, 1, 10)         // Range check (int/float)
```

**Simple Helpers:**
```go
Show(condition, component)       // Show if true
Hide(condition, component)       // Hide if true
Toggle(state, onComp, offComp)   // Toggle between two
```

### Usage Patterns

#### Import Pattern

Use dot import for cleaner syntax:
```go
import (
    . "github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
    "github.com/charmbracelet/lipgloss"
)
```

#### Component Naming Conventions

- **Struct types**: `VStack`, `HStack`, `Text`, `List`, etc.
- **Constructor functions**: `NewVStack()`, `NewText()`, `NewList()`
- **Convenience functions** (return Component interface): `VStackC()`, `T()`, `ListC()`
- Use convenience functions for inline composition
- Use constructors when you need to call methods before rendering

#### Complex Layouts

```go
// Dashboard example
VStackC(
    T("Dashboard").Bold(true).Primary(),
    SpacerV(1),

    // Stats row
    HStackC(
        Card(VStackC(
            T("Users").Muted(),
            T("1,234").Bold(true).Primary(),
        )),
        SpacerH(2),
        Card(VStackC(
            T("Revenue").Muted(),
            T("$45,678").Bold(true).Success(),
        )),
    ),

    SpacerV(1),

    // Recent activity
    Card(
        VStackC(
            T("Recent Activity").Bold(true),
            DividerLine(40),
            SpacerV(1),
            NewBulletList(
                T("User registered"),
                T("Payment processed"),
                T("Backup completed"),
            ),
        ),
    ),
).Render()
```

#### Form-like UIs

```go
VStackC(
    T("Settings").Bold(true).Primary(),
    SpacerV(1),

    VStackC(
        T("Username: " + username),
        When(len(username) >= 3, T("✓ Valid").Success()),
        Unless(len(username) >= 3, T("✗ Too short").Error()),
    ),

    SpacerV(1),

    IfC(
        isFormValid,
        BoxC(T("Save")).Background(lipgloss.Color("42")),
        BoxC(T("Save")).Background(lipgloss.Color("240")),
    ),
).Render()
```

#### Table-like Layouts

```go
VStackC(
    HStackC(
        T("ID").Bold(true).Width(10),
        T("Name").Bold(true).Width(20),
        T("Status").Bold(true).Width(15),
    ),
    DividerLine(45),
    HStackC(
        T("1").Width(10),
        T("Alice").Width(20),
        T("Active").Success().Width(15),
    ),
    HStackC(
        T("2").Width(10),
        T("Bob").Width(20),
        T("Inactive").Muted().Width(15),
    ),
).Render()
```

### Testing

**Location:** `internal/ui/component/component_test.go`

**Framework:** `testify/suite` (project standard)

**Coverage:** 50+ comprehensive test cases covering:
- All component types and their methods
- Chainable modifiers
- Layout composition
- Conditional rendering
- Complex nested scenarios
- Style application

**Running Tests:**
```bash
go test ./internal/ui/component/ -v
```

### Examples

**Location:** `internal/ui/component/examples_test.go`

Contains 10+ real-world examples demonstrating:
- Basic text styling
- Layout patterns (VStack, HStack, Card)
- Interactive lists
- Conditional rendering
- Complex dashboards
- Form validation UIs
- Status pages
- Responsive layouts
- Table-like structures

### Design Principles

1. **Declarative over Imperative**: Describe what you want, not how to build it
2. **Composability**: Components nest infinitely and predictably
3. **Immutability**: Methods return new instances, never mutate
4. **Type Safety**: Go compiler catches errors at compile time
5. **Integration**: Seamless `lipgloss` integration for all styling
6. **Familiarity**: API inspired by React/SwiftUI for easy adoption
7. **Performance**: Efficient rendering with minimal allocations

### Integration with Bubble Tea

Components work seamlessly with Bubble Tea's Model-View-Update pattern:

```go
type Model struct {
    selectedIndex int
    items         []string
}

func (m Model) View() string {
    return VStackC(
        T("My App").Bold(true).Primary(),
        SpacerV(1),
        StringList(m.items).Selected(m.items[m.selectedIndex]),
    ).Render()
}
```

The component system handles rendering; Bubble Tea handles state management and user input.

### Best Practices

1. **Use convenience functions** (`VStackC`, `T`) for inline composition
2. **Use constructors** (`NewVStack`, `NewText`) when calling methods
3. **Leverage presets** (`.Primary()`, `.Success()`) for consistent styling
4. **Compose small components** into larger ones for reusability
5. **Extract complex layouts** into helper functions
6. **Use conditionals** (`IfC`, `When`) for dynamic UIs
7. **Prefer semantic names** over generic styling when possible
8. **Test component rendering** in isolation before integration

### Common Gotcalls

- Components are immutable - methods return new instances
- Don't forget to call `.Render()` at the end of your composition
- Use `Component` interface for function parameters, not concrete types
- Dot imports (`.`) are recommended for cleaner syntax
- `lipgloss.Color()` requires string color codes (e.g., "205", "42")

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

### Router Tests

- **Location:** `internal/view/router_test.go`
- Uses `testify/suite` pattern (project standard)
- 27 comprehensive test cases covering all navigation scenarios
- No external dependencies required (pure unit tests)
- Tests: route matching, parameter extraction, navigation stack, Bubble Tea integration

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

## Router Generator (routegen)

### Overview

The router generator is a code generation tool that automatically creates route definitions from an app folder structure, similar to Next.js's file-based routing system. It scans the `app/` directory for `page.go` files and generates a `routes_gen.go` file with all route configurations.

**Location:** `tools/routergen/`

### App Folder Structure Conventions

The generator follows these conventions:

```
app/
├── page.go                    → Route: "/"          (root route)
├── users/
│   ├── page.go               → Route: "/users"     (users list)
│   └── _id/
│       └── page.go           → Route: "/users/:id" (user detail - dynamic segment)
└── posts/
    ├── page.go               → Route: "/posts"
    └── _postId/
        ├── page.go           → Route: "/posts/:postId"
        └── comments/
            └── _commentId/
                └── page.go   → Route: "/posts/:postId/comments/:commentId"
```

**Key Conventions:**

1. **page.go files**: Define route endpoints. Each `page.go` must export a `NewPage()` function that returns a `view.View`
2. **Dynamic segments**: Folders prefixed with `_` (e.g., `_id`, `_userId`) become dynamic route parameters (`:id`, `:userId`)
3. **Package naming**: All page.go files should be in the `app` package (or appropriate subpackage)
4. **Nested routes**: Subdirectories create nested route paths

### Page File Structure

Each `page.go` file must implement:

```go
package app  // or appropriate subpackage

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/rxtech-lab/smart-contract-cli/internal/view"
)

type Model struct {
    // Your model state
}

// Required: Must export NewPage() function
func NewPage() view.View {
    return Model{}
}

// Implement view.View interface
func (m Model) Init() tea.Cmd { return nil }
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m Model) View() string { return "Page content" }
```

### Generated Output

The generator creates `app/routes_gen.go`:

```go
package app

import (
    "github.com/rxtech-lab/smart-contract-cli/internal/view"
    users__id_page "github.com/rxtech-lab/smart-contract-cli/app/users/_id"
    users_page "github.com/rxtech-lab/smart-contract-cli/app/users"
)

// GetRoutes returns all routes generated from the app folder structure.
func GetRoutes() []view.Route {
    return []view.Route{
        {Path: "/", Component: NewPage()},
        {Path: "/users", Component: users_page.NewPage()},
        {Path: "/users/:id", Component: users__id_page.NewPage()},
    }
}
```

**Note:** The root `app/page.go` is called directly as `NewPage()` to avoid circular imports since `routes_gen.go` is also in the `app` package.

### Usage

**Generate routes:**

```bash
make generate-routes
```

This command:
1. Builds the `routegen` tool to `bin/routegen`
2. Scans the `app/` directory
3. Generates `app/routes_gen.go` with route definitions

**Manual invocation:**

```bash
# Build the tool
go build -o bin/routegen ./tools/routergen/*.go

# Generate routes
./bin/routegen -dir ./app -module-root .

# Custom options
./bin/routegen \
  -dir /path/to/app \
  -output /path/to/output.go \
  -module github.com/your/module \
  -module-root /path/to/module/root
```

**CLI Flags:**

- `-dir`: Path to app folder (default: `./app`)
- `-output`: Output file path (default: `<app-dir>/routes_gen.go`)
- `-module`: Go module name (auto-detected from go.mod if not provided)
- `-module-root`: Path to module root directory (default: `.`)

### Workflow

1. **Create page files** in the `app/` folder structure
2. **Run generator**: `make generate-routes`
3. **Use routes** in your application:
   ```go
   router := view.NewRouter()
   router.SetRoutes(app.GetRoutes())
   router.NavigateTo("/", nil)
   ```

### Implementation Details

**Core Components:**

- `tools/routergen/types.go`: Route definition types
- `tools/routergen/generator.go`: Business logic for scanning and code generation
- `tools/routergen/generator_test.go`: Comprehensive test suite (testify/suite pattern)
- `tools/routergen/main.go`: CLI entry point

**Key Functions:**

- `ScanAppFolder(rootDir string)`: Walks directory tree, finds `page.go` files
- `convertToRoutePath(fsPath string)`: Converts `users/_id` → `/users/:id`
- `generatePackageAlias(fsPath string)`: Creates unique import aliases
- `GenerateRoutesFile(routes, moduleName)`: Generates Go source code
- `ConvertAbsoluteToModulePath()`: Converts filesystem paths to Go import paths

**Testing:**

```bash
# Run generator tests
go test ./tools/routergen/

# Test with verbose output
go test ./tools/routergen/ -v
```

The test suite includes 27+ test cases covering:
- Route path conversion (static and dynamic)
- Package alias generation
- Directory scanning with various structures
- Generated code format validation
- Edge cases and error handling

### Git Ignore

Generated files are ignored in version control:

```gitignore
routes_gen.go   # Generated route file
bin/            # Compiled tools
```

### Why This Approach?

**Benefits:**

1. **Convention over configuration**: File structure directly maps to routes
2. **Type safety**: Generated code is checked by Go compiler
3. **No manual route registration**: Routes auto-discovered from filesystem
4. **Familiar pattern**: Developers comfortable with Next.js will recognize the structure
5. **Testable**: Generator logic is fully tested and separated from CLI
6. **Flexible**: Supports nested routes and dynamic segments

**Trade-offs:**

- Requires code generation step (added to Makefile)
- Dynamic segments use `_param` convention instead of `[param]` (Go package naming limitation)
- Generated file should not be manually edited (regenerate instead)
