package deletewallet

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

var logger, _ = log.NewFileLogger("./logs/evm/wallet/delete.log")

type confirmationOption struct {
	label string
	value bool
}

type viewMode int

const (
	modeConfirmation viewMode = iota
	modeCannotDelete
)

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	walletID         uint
	wallet           *wallet.WalletWithBalance
	selectedWalletID uint
	selectedIndex    int
	options          []confirmationOption
	mode             viewMode

	loading  bool
	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new delete wallet page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	return Model{
		router:        router,
		sharedMemory:  sharedMemory,
		walletService: walletService,
		loading:       true,
		mode:          modeConfirmation,
		selectedIndex: 0,
		options: []confirmationOption{
			{label: "No, cancel", value: false},
			{label: "Yes, delete permanently", value: true},
		},
	}
}

func (m Model) Init() tea.Cmd {
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
	if walletIDStr == "" {
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

	rpcEndpoint := "http://localhost:8545"

	// Get wallet with balance
	walletData, err := walletService.GetWalletWithBalance(uint(walletID), rpcEndpoint)
	if err != nil {
		return walletLoadedMsg{err: fmt.Errorf("failed to load wallet: %w", err)}
	}

	// Get selected wallet ID from shared memory
	selectedWalletIDVal, _ := m.sharedMemory.Get(config.SelectedWalletIDKey)
	var selectedWalletID uint
	if selectedWalletIDVal != nil {
		if id, ok := selectedWalletIDVal.(uint); ok {
			selectedWalletID = id
		}
	}

	return walletLoadedMsg{
		walletID:         uint(walletID),
		wallet:           walletData,
		walletService:    walletService,
		selectedWalletID: selectedWalletID,
	}
}

type walletLoadedMsg struct {
	walletID         uint
	wallet           *wallet.WalletWithBalance
	walletService    wallet.WalletService
	selectedWalletID uint
	err              error
}

type walletDeletedMsg struct {
	success bool
	err     error
}

func (m Model) deleteWallet() tea.Msg {
	err := m.walletService.DeleteWallet(m.walletID)
	if err != nil {
		return walletDeletedMsg{success: false, err: err}
	}
	return walletDeletedMsg{success: true}
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
		m.walletID = msg.walletID
		m.wallet = msg.wallet
		m.walletService = msg.walletService
		m.selectedWalletID = msg.selectedWalletID

		// Check if trying to delete currently selected wallet
		if m.walletID == m.selectedWalletID {
			m.mode = modeCannotDelete
		} else {
			m.mode = modeConfirmation
		}

		return m, nil

	case walletDeletedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		if msg.success {
			logger.Info("Wallet deleted successfully: %d", m.walletID)
			// Navigate back to wallet list
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch m.mode {
		case modeCannotDelete:
			// Any key goes back
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
					// User confirmed, delete wallet
					return m, m.deleteWallet
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
	case modeCannotDelete:
		return "Press any key to go back...", view.HelpDisplayOptionOverride
	default:
		return "↑/k: up • ↓/j: down • enter: confirm • esc: cancel", view.HelpDisplayOptionAppend
	}
}

func (m Model) View() string {
	if m.loading {
		return component.VStackC(
			component.T("Delete Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Loading wallet...").Muted(),
		).Render()
	}

	if m.errorMsg != "" {
		return component.VStackC(
			component.T("Delete Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
			component.SpacerV(1),
			component.T("Press 'esc' to go back").Muted(),
		).Render()
	}

	if m.wallet == nil {
		return component.VStackC(
			component.T("Delete Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Wallet not found").Error(),
		).Render()
	}

	switch m.mode {
	case modeCannotDelete:
		return m.renderCannotDelete()
	default:
		return m.renderConfirmation()
	}
}

func (m Model) renderConfirmation() string {
	// Format balance
	balanceStr := "unavailable"
	if m.wallet.Error == nil && m.wallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.wallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		balanceStr = fmt.Sprintf("%.4f ETH", ethValue)
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
			description = "Remove this wallet from the system"
		} else {
			description = "Keep this wallet"
		}

		optionComponents = append(optionComponents, component.VStackC(
			labelStyle,
			component.T("  "+description).Muted(),
			component.SpacerV(1),
		))
	}

	return component.VStackC(
		component.T("Delete Wallet").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Are you sure you want to delete this wallet?").Bold(true),
		component.SpacerV(1),
		component.T("Wallet Information:").Bold(true),
		component.T("• Alias: "+m.wallet.Wallet.Alias).Muted(),
		component.T("• Address: "+m.wallet.Wallet.Address).Muted(),
		component.T("• Balance: "+balanceStr).Muted(),
		component.SpacerV(1),
		component.T("⚠ Warning: This action cannot be undone!").Warning(),
		component.SpacerV(1),
		component.T("Make sure you have:"),
		component.T("• Backed up your private key or mnemonic").Muted(),
		component.T("• Transferred funds to another wallet").Muted(),
		component.T("• No pending transactions").Muted(),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderCannotDelete() string {
	// Format balance
	balanceStr := "unavailable"
	if m.wallet.Error == nil && m.wallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.wallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		balanceStr = fmt.Sprintf("%.4f ETH", ethValue)
	}

	return component.VStackC(
		component.T("Delete Wallet").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Cannot delete currently selected wallet!").Error(),
		component.SpacerV(1),
		component.T("Wallet Information:").Bold(true),
		component.T("• Alias: "+m.wallet.Wallet.Alias+" ★").Muted(),
		component.T("• Address: "+m.wallet.Wallet.Address).Muted(),
		component.T("• Balance: "+balanceStr).Muted(),
		component.T("• Status: Currently Selected").Muted(),
		component.SpacerV(1),
		component.T("To delete this wallet:").Bold(true),
		component.T("1. Select another wallet as active"),
		component.T("2. Return to this wallet"),
		component.T("3. Then delete it"),
		component.SpacerV(1),
		component.T("This prevents accidentally losing access to your active wallet.").Muted(),
	).Render()
}
