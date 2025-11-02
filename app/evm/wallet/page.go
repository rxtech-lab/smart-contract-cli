package wallet

import (
	"fmt"
	"math/big"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/wallet"
	"github.com/rxtech-lab/smart-contract-cli/internal/log"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/utils"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

var logger, _ = log.NewFileLogger("./logs/evm/wallet/page.log")

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	wallets          []wallet.WalletWithBalance
	selectedIndex    int
	selectedWalletID uint
	rpcEndpoint      string

	loading  bool
	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new wallet page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	return Model{
		router:        router,
		sharedMemory:  sharedMemory,
		walletService: walletService,
		selectedIndex: 0,
		loading:       true,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadWallets
}

func (m Model) createWalletService() (wallet.WalletService, error) {
	// Get storage client from shared memory using utils
	sqlStorage, err := utils.GetStorageClientFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return nil, fmt.Errorf("failed to get storage client from shared memory: %w", err)
	}

	// Get secure storage
	secureStorage, _, err := utils.GetSecureStorageFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get secure storage from shared memory: %v", err)
		return nil, fmt.Errorf("failed to get secure storage from shared memory: %w", err)
	}
	// Create wallet service
	return wallet.NewWalletService(sqlStorage, secureStorage), nil
}

func (m Model) loadWallets() tea.Msg {
	// Use injected wallet service if available (for testing)
	walletService := m.walletService
	if walletService == nil {
		svc, err := m.createWalletService()
		if err != nil {
			return walletLoadedMsg{err: err}
		}
		walletService = svc
	}

	// Get RPC endpoint from database
	sqlStorage, err := utils.GetStorageClientFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return walletLoadedMsg{err: fmt.Errorf("failed to get storage client from shared memory: %w", err)}
	}

	// Get the current config
	config, err := sqlStorage.GetCurrentConfig()
	if err != nil {
		logger.Error("Failed to get current config: %v", err)
		return walletLoadedMsg{err: fmt.Errorf("failed to get current config: %w", err)}
	}
	if config.Endpoint == nil {
		logger.Error("No RPC endpoint configured")
		return walletLoadedMsg{err: fmt.Errorf("no RPC endpoint configured. Please configure an endpoint first")}
	}
	rpcEndpoint := config.Endpoint.Url
	logger.Info("Using RPC endpoint: %s", rpcEndpoint)

	// Get selected wallet ID from config
	var selectedWalletID uint
	if config.SelectedWalletID != nil {
		selectedWalletID = *config.SelectedWalletID
	}

	// List wallets with balances
	wallets, totalCount, err := walletService.ListWalletsWithBalances(1, 100, rpcEndpoint)
	if err != nil {
		return walletLoadedMsg{err: err}
	}

	return walletLoadedMsg{
		wallets:          wallets,
		totalCount:       totalCount,
		walletService:    walletService,
		rpcEndpoint:      rpcEndpoint,
		selectedWalletID: selectedWalletID,
	}
}

type walletLoadedMsg struct {
	wallets          []wallet.WalletWithBalance
	totalCount       int64
	walletService    wallet.WalletService
	rpcEndpoint      string
	selectedWalletID uint
	err              error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case walletLoadedMsg:
		if msg.err != nil {
			m.loading = false
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		m.loading = false
		m.wallets = msg.wallets
		m.walletService = msg.walletService
		m.rpcEndpoint = msg.rpcEndpoint
		m.selectedWalletID = msg.selectedWalletID
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.wallets)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Navigate to wallet actions
			if len(m.wallets) > 0 {
				walletID := m.wallets[m.selectedIndex].Wallet.ID
				logger.Info("Enter pressed, navigating to wallet actions for wallet ID: %d", walletID)
				err := m.router.NavigateTo("/evm/wallet/actions", map[string]string{
					"id": strconv.FormatUint(uint64(walletID), 10),
				})
				if err != nil {
					logger.Error("Navigation error: %v", err)
				}
				return m, nil
			}

		case "a":
			// Navigate to add wallet page
			logger.Info("Navigating to add wallet page")
			if err := m.router.NavigateTo("/evm/wallet/add", nil); err != nil {
				logger.Error("Failed to navigate to add wallet page: %v", err)
			}
			return m, nil

		case "r":
			// Refresh wallets
			m.loading = true
			return m, m.loadWallets
		}
	}

	return m, nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	if m.loading {
		return "Loading...", view.HelpDisplayOptionOverride
	}

	return "↑/k: up • ↓/j: down • enter: actions • a: add wallet • r: refresh • esc/q: back", view.HelpDisplayOptionAppend
}

func (m Model) View() string {
	if m.loading {
		return component.VStackC(
			component.T("Wallet Management").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Loading wallets...").Muted(),
		).Render()
	}

	if m.errorMsg != "" {
		return component.VStackC(
			component.T("Wallet Management").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
			component.SpacerV(1),
			component.T("Press 'r' to retry or 'esc' to go back").Muted(),
		).Render()
	}

	// Empty state
	if len(m.wallets) == 0 {
		return component.VStackC(
			component.T("Wallet Management").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("No wallets found").Bold(true),
			component.SpacerV(1),
			component.T("You haven't added any wallets yet. Wallets are required to sign"),
			component.T("transactions and interact with smart contracts."),
			component.SpacerV(1),
			component.T("Get started by:"),
			component.T("• Importing an existing wallet with private key or mnemonic"),
			component.T("• Generating a new wallet"),
			component.SpacerV(1),
			component.T("Press 'a' to add your first wallet").Muted(),
		).Render()
	}

	// Wallet list
	walletItems := make([]component.Component, 0)

	for index, walletItem := range m.wallets {
		isSelected := walletItem.Wallet.ID == m.selectedWalletID
		isCursor := index == m.selectedIndex

		// Format balance
		balanceStr := "unavailable ⚠"
		if walletItem.Error == nil && walletItem.Balance != nil {
			// Convert wei to ETH
			ethValue := new(big.Float).Quo(
				new(big.Float).SetInt(walletItem.Balance),
				new(big.Float).SetInt(big.NewInt(1e18)),
			)
			balanceStr = fmt.Sprintf("%.4f ETH", ethValue)
		}

		// Build wallet item
		prefix := "  "
		if isSelected {
			prefix = "★ "
		}
		if isCursor {
			prefix = "> "
			if isSelected {
				prefix = ">★"
			}
		}

		aliasStyle := component.T(prefix + walletItem.Wallet.Alias)
		if isSelected {
			aliasStyle = aliasStyle.Foreground(lipgloss.Color("42")) // Green for selected
		}
		if isCursor {
			aliasStyle = aliasStyle.Bold(true)
		}

		walletComponent := component.VStackC(
			aliasStyle,
			component.T("  Address: "+walletItem.Wallet.Address).Muted(),
			component.T("  Balance: "+balanceStr).Muted(),
			component.IfC(isSelected,
				component.T("  Status: Selected").Muted(),
				component.T("  Status: Available").Muted(),
			),
			component.SpacerV(1),
		)

		walletItems = append(walletItems, walletComponent)
	}

	return component.VStackC(
		component.T("Wallet Management").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Manage your wallets").Muted(),
		component.SpacerV(1),
		component.VStackC(walletItems...),
		component.SpacerV(1),
		component.T("Endpoint: "+m.rpcEndpoint+" (Anvil)").Muted(),
		component.SpacerV(1),
		component.T("Legend:").Muted(),
		component.T("★ = Currently selected wallet").Muted(),
		component.T("> = Cursor position").Muted(),
	).Render()
}
