package selectwallet

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

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

var logger, _ = log.NewFileLogger("./logs/evm/wallet/select.log")

type confirmationOption struct {
	label string
	value bool
}

type viewMode int

const (
	modeConfirmation viewMode = iota
	modeSuccess
)

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	newWalletID     uint
	currentWalletID uint
	newWallet       *wallet.WalletWithBalance
	currentWallet   *wallet.WalletWithBalance
	selectedIndex   int
	options         []confirmationOption
	mode            viewMode
	successMessage  string

	loading  bool
	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new select wallet page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	return Model{
		router:        router,
		sharedMemory:  sharedMemory,
		walletService: walletService,
		loading:       true,
		mode:          modeConfirmation,
		selectedIndex: 0,
		options: []confirmationOption{
			{label: "Yes, switch wallet", value: true},
			{label: "No, cancel", value: false},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadWallets
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

func (m Model) loadWallets() tea.Msg {
	// Get new wallet ID from query params
	newWalletIDStr := m.router.GetQueryParam("id")
	if newWalletIDStr == "" {
		return walletsLoadedMsg{err: fmt.Errorf("wallet ID not provided")}
	}

	newWalletID, err := strconv.ParseUint(newWalletIDStr, 10, 32)
	if err != nil {
		return walletsLoadedMsg{err: fmt.Errorf("invalid wallet ID: %w", err)}
	}

	// Use injected wallet service if available (for testing)
	walletService := m.walletService
	if walletService == nil {
		svc, err := m.createWalletService()
		if err != nil {
			return walletsLoadedMsg{err: err}
		}
		walletService = svc
	}

	// Get current selected wallet ID from shared memory
	currentWalletIDVal, _ := m.sharedMemory.Get(config.SelectedWalletIDKey)
	var currentWalletID uint
	if currentWalletIDVal != nil {
		if id, ok := currentWalletIDVal.(uint); ok {
			currentWalletID = id
		}
	}

	// If trying to select the same wallet, show error
	if uint(newWalletID) == currentWalletID {
		return walletsLoadedMsg{err: fmt.Errorf("wallet is already selected")}
	}

	rpcEndpoint := "http://localhost:8545"

	// Load both wallets
	newWallet, err := walletService.GetWalletWithBalance(uint(newWalletID), rpcEndpoint)
	if err != nil {
		return walletsLoadedMsg{err: fmt.Errorf("failed to load new wallet: %w", err)}
	}

	var currentWallet *wallet.WalletWithBalance
	if currentWalletID != 0 {
		currentWallet, err = walletService.GetWalletWithBalance(currentWalletID, rpcEndpoint)
		if err != nil {
			logger.Warn("Failed to load current wallet: %v", err)
			// Don't fail if current wallet can't be loaded, just continue
		}
	}

	return walletsLoadedMsg{
		newWalletID:     uint(newWalletID),
		currentWalletID: currentWalletID,
		newWallet:       newWallet,
		currentWallet:   currentWallet,
		walletService:   walletService,
	}
}

type walletsLoadedMsg struct {
	newWalletID     uint
	currentWalletID uint
	newWallet       *wallet.WalletWithBalance
	currentWallet   *wallet.WalletWithBalance
	walletService   wallet.WalletService
	err             error
}

type walletSelectedMsg struct {
	success bool
	err     error
}

func (m Model) selectWallet() tea.Msg {
	// Update selected wallet in shared memory
	err := m.sharedMemory.Set(config.SelectedWalletIDKey, m.newWalletID)
	if err != nil {
		return walletSelectedMsg{success: false, err: err}
	}

	return walletSelectedMsg{success: true}
}

type successDismissMsg struct{}

func waitForDismiss() tea.Msg {
	time.Sleep(2 * time.Second)
	return successDismissMsg{}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case walletsLoadedMsg:
		if msg.err != nil {
			m.loading = false
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		m.loading = false
		m.newWalletID = msg.newWalletID
		m.currentWalletID = msg.currentWalletID
		m.newWallet = msg.newWallet
		m.currentWallet = msg.currentWallet
		m.walletService = msg.walletService
		return m, nil

	case walletSelectedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		if msg.success {
			m.mode = modeSuccess
			m.successMessage = "✓ Active wallet changed successfully!"
			return m, waitForDismiss
		}

	case successDismissMsg:
		// Navigate back to wallet list
		return m, func() tea.Msg {
			_ = m.router.NavigateTo("/evm/wallet", nil)
			return nil
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch m.mode {
		case modeSuccess:
			// Any key dismisses success screen
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}

		case modeConfirmation:
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
				option := m.options[m.selectedIndex]
				if option.value {
					// User confirmed, switch wallet
					return m, m.selectWallet
				} else {
					// User cancelled, go back
					return m, func() tea.Msg {
						_ = m.router.NavigateTo("/evm/wallet", nil)
						return nil
					}
				}
			}
		}
	}

	return m, nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	if m.loading {
		return "Loading...", view.HelpDisplayOptionOverride
	}

	switch m.mode {
	case modeSuccess:
		return "Press any key to return to wallet list...", view.HelpDisplayOptionOverride
	default:
		return "↑/k: up • ↓/j: down • enter: confirm • esc: cancel", view.HelpDisplayOptionAppend
	}
}

func (m Model) View() string {
	if m.loading {
		return component.VStackC(
			component.T("Select Active Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Loading wallets...").Muted(),
		).Render()
	}

	if m.errorMsg != "" {
		return component.VStackC(
			component.T("Select Active Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
			component.SpacerV(1),
			component.T("Press 'esc' to go back").Muted(),
		).Render()
	}

	switch m.mode {
	case modeSuccess:
		return m.renderSuccess()
	default:
		return m.renderConfirmation()
	}
}

func (m Model) renderConfirmation() string {
	// Format balances
	currentBalance := "N/A"
	if m.currentWallet != nil && m.currentWallet.Error == nil && m.currentWallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.currentWallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		currentBalance = fmt.Sprintf("%.4f ETH", ethValue)
	}

	newBalance := "unavailable"
	if m.newWallet.Error == nil && m.newWallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.newWallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		newBalance = fmt.Sprintf("%.4f ETH", ethValue)
	}

	// Build current wallet display
	var currentWalletDisplay component.Component
	if m.currentWallet != nil {
		currentWalletDisplay = component.VStackC(
			component.T("Current Active Wallet:").Bold(true),
			component.T("• Alias: "+m.currentWallet.Wallet.Alias+" ★").Muted(),
			component.T("• Address: "+m.currentWallet.Wallet.Address).Muted(),
			component.T("• Balance: "+currentBalance).Muted(),
			component.SpacerV(1),
		)
	} else {
		currentWalletDisplay = component.VStackC(
			component.T("Current Active Wallet: None").Bold(true),
			component.SpacerV(1),
		)
	}

	// Build options
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

		var description string
		if option.value {
			description = "Make " + m.newWallet.Wallet.Alias + " the active wallet"
		} else {
			if m.currentWallet != nil {
				description = "Keep " + m.currentWallet.Wallet.Alias + " as active"
			} else {
				description = "Keep no wallet selected"
			}
		}

		optionComponents = append(optionComponents, component.VStackC(
			labelStyle,
			component.T("  "+description).Muted(),
			component.SpacerV(1),
		))
	}

	return component.VStackC(
		component.T("Select Active Wallet").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Switch active wallet?"),
		component.SpacerV(1),
		currentWalletDisplay,
		component.T("New Active Wallet:").Bold(true),
		component.T("• Alias: "+m.newWallet.Wallet.Alias).Muted(),
		component.T("• Address: "+m.newWallet.Wallet.Address).Muted(),
		component.T("• Balance: "+newBalance).Muted(),
		component.SpacerV(1),
		component.T("All future transactions will use the new active wallet."),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderSuccess() string {
	// Format balance
	newBalance := "unavailable"
	if m.newWallet.Error == nil && m.newWallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.newWallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		newBalance = fmt.Sprintf("%.4f ETH", ethValue)
	}

	return component.VStackC(
		component.T("Select Active Wallet").Bold(true).Primary(),
		component.SpacerV(1),
		component.T(m.successMessage).Success(),
		component.SpacerV(1),
		component.T("Your active wallet is now:").Bold(true),
		component.T("• Alias: "+m.newWallet.Wallet.Alias+" ★").Muted(),
		component.T("• Address: "+m.newWallet.Wallet.Address).Muted(),
		component.T("• Balance: "+newBalance).Muted(),
		component.SpacerV(1),
		component.T("All transactions will now use this wallet."),
	).Render()
}
