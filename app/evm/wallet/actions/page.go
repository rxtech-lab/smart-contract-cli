package actions

import (
	"fmt"
	"math/big"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/wallet"
	"github.com/rxtech-lab/smart-contract-cli/internal/log"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/utils"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

var logger, _ = log.NewFileLogger("./logs/evm/wallet/actions.log")

type actionOption struct {
	label       string
	description string
	route       string
	action      func(m *Model) tea.Cmd
}

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	walletID         uint
	wallet           *wallet.WalletWithBalance
	selectedWalletID uint
	selectedIndex    int
	options          []actionOption

	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new actions page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	return Model{
		router:        router,
		sharedMemory:  sharedMemory,
		walletService: walletService,
		selectedIndex: 0,
	}
}

func (m Model) Init() tea.Cmd {
	logger.Info("Actions page Init() called")
	return m.loadWallet
}

func (m Model) createWalletService() (wallet.WalletService, error) {
	storageClient, err := m.sharedMemory.Get(config.StorageClientKey)
	if err != nil || storageClient == nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return nil, fmt.Errorf("storage client not initialized")
	}

	sqlStorage, isValidStorage := storageClient.(sql.Storage)
	if !isValidStorage {
		logger.Error("Invalid storage client type")
		return nil, fmt.Errorf("invalid storage client type")
	}

	secureStorage, _, err := utils.GetSecureStorageFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get secure storage from shared memory: %v", err)
		return nil, fmt.Errorf("failed to get secure storage from shared memory: %w", err)
	}

	return wallet.NewWalletService(sqlStorage, secureStorage), nil
}

func (m Model) loadWallet() tea.Msg {
	// Get wallet ID from query params
	walletIDStr := m.router.GetQueryParam("id")
	logger.Info("Loading wallet with ID: %s", walletIDStr)
	if walletIDStr == "" {
		logger.Error("Wallet ID not provided in query params")
		return walletLoadedMsg{err: fmt.Errorf("wallet ID not provided")}
	}

	walletID, err := strconv.ParseUint(walletIDStr, 10, 32)
	if err != nil {
		return walletLoadedMsg{err: fmt.Errorf("invalid wallet ID: %w", err)}
	}

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

	// Get wallet with balance
	logger.Info("Fetching wallet %d with balance from %s", walletID, rpcEndpoint)
	walletData, err := walletService.GetWalletWithBalance(uint(walletID), rpcEndpoint)
	if err != nil {
		logger.Error("Failed to get wallet with balance: %v", err)
		return walletLoadedMsg{err: fmt.Errorf("failed to load wallet: %w", err)}
	}
	logger.Info("Successfully loaded wallet: %s", walletData.Wallet.Alias)

	// Get selected wallet ID from shared memory
	selectedWalletID := config.SelectedWalletID
	if selectedWalletID == nil {
		return walletLoadedMsg{err: fmt.Errorf("no selected wallet ID found in config")}
	}

	return walletLoadedMsg{
		walletID:         uint(walletID),
		wallet:           walletData,
		walletService:    walletService,
		selectedWalletID: *config.SelectedWalletID,
	}
}

type walletLoadedMsg struct {
	walletID         uint
	wallet           *wallet.WalletWithBalance
	walletService    wallet.WalletService
	selectedWalletID uint
	err              error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case walletLoadedMsg:
		logger.Info("Received walletLoadedMsg")
		if msg.err != nil {
			logger.Error("walletLoadedMsg contains error: %v", msg.err)
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		logger.Info("Setting wallet data, loading=false")
		m.walletID = msg.walletID
		m.wallet = msg.wallet
		m.walletService = msg.walletService
		m.selectedWalletID = msg.selectedWalletID

		// Build action options based on whether this is the selected wallet
		isSelected := m.walletID == m.selectedWalletID
		m.options = m.buildActionOptions(isSelected)
		logger.Info("Wallet loaded successfully, options count: %d", len(m.options))

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.options)-1 {
				m.selectedIndex++
			}

		case "enter":
			if len(m.options) > 0 {
				option := m.options[m.selectedIndex]
				if option.action != nil {
					return m, option.action(&m)
				} else if option.route != "" {
					return m, func() tea.Msg {
						_ = m.router.NavigateTo(option.route, map[string]string{
							"id": strconv.FormatUint(uint64(m.walletID), 10),
						})
						return nil
					}
				}
			}

		case "ctrl+c":
			// Handle quit for testing purposes
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) buildActionOptions(isSelected bool) []actionOption {
	options := []actionOption{}

	// If not selected, add "Select as active wallet" option first
	if !isSelected {
		options = append(options, actionOption{
			label:       "Select as active wallet",
			description: "Make this the current active wallet for transactions",
			route:       "/evm/wallet/select",
		})
	}

	// Common options for all wallets
	options = append(options, []actionOption{
		{
			label:       "View details",
			description: "View full wallet information and transaction history",
			route:       "/evm/wallet/details",
		},
		{
			label:       "Update wallet",
			description: "Update wallet alias or private key",
			route:       "/evm/wallet/update",
		},
		{
			label:       "Delete wallet",
			description: "Remove this wallet from the system",
			route:       "/evm/wallet/delete",
		},
	}...)

	return options
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	return "↑/k: up • ↓/j: down • enter: confirm • esc: cancel", view.HelpDisplayOptionAppend
}

func (m Model) View() string {
	if m.errorMsg != "" {
		return component.VStackC(
			component.T("Wallet Actions").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
			component.SpacerV(1),
			component.T("Press 'esc' to go back").Muted(),
		).Render()
	}

	if m.wallet == nil {
		return component.VStackC(
			component.T("Wallet Actions").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Wallet not found").Error(),
		).Render()
	}

	// Format balance
	balanceStr := "unavailable ⚠"
	if m.wallet.Error == nil && m.wallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.wallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		balanceStr = fmt.Sprintf("%.4f ETH", ethValue)
	}

	// Title with selected indicator
	title := "Wallet Actions - " + m.wallet.Wallet.Alias
	if m.walletID == m.selectedWalletID {
		title = title + " ★"
	}

	// Build header
	header := component.VStackC(
		component.T(title).Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Address: "+m.wallet.Wallet.Address).Muted(),
		component.T("Balance: "+balanceStr).Muted(),
		component.SpacerV(1),
	)

	// Add note for selected wallet
	if m.walletID == m.selectedWalletID {
		header = component.VStackC(
			header,
			component.T("This is your currently selected wallet.").Muted(),
			component.SpacerV(1),
		)
	}

	// Build action menu
	header = component.VStackC(
		header,
		component.T("What would you like to do?").Bold(true),
		component.SpacerV(1),
	)

	// Build options list
	optionComponents := make([]component.Component, 0)
	for i, option := range m.options {
		isCursor := i == m.selectedIndex

		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		labelStyle := component.T(prefix + option.label)
		if isCursor {
			labelStyle = labelStyle.Bold(true)
		}

		optionComponents = append(optionComponents, component.VStackC(
			labelStyle,
			component.T("  "+option.description).Muted(),
			component.SpacerV(1),
		))
	}

	// Add note for selected wallet (cannot deselect)
	footer := component.Empty()
	if m.walletID == m.selectedWalletID {
		footer = component.VStackC(
			component.SpacerV(1),
			component.T("Note: You cannot deselect the active wallet. Select another wallet first.").Muted(),
		)
	}

	return component.VStackC(
		header,
		component.VStackC(optionComponents...),
		footer,
	).Render()
}
