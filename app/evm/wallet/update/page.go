package update

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
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

var logger, _ = log.NewFileLogger("./logs/evm/wallet/update.log")

type updateStep int

const (
	stepSelectUpdate updateStep = iota
	stepUpdateAlias
	stepPrivateKeyWarning
	stepUpdatePrivateKey
	stepSuccess
)

type updateOption struct {
	label       string
	description string
	step        updateStep
}

type confirmationOption struct {
	label string
	value bool
}

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	walletID      uint
	wallet        *wallet.WalletWithBalance
	currentStep   updateStep
	selectedIndex int
	options       []updateOption
	confirmOpts   []confirmationOption

	aliasInput textinput.Model
	pkeyInput  textinput.Model

	oldAddress     string
	newAddress     string
	updatedBalance string

	loading  bool
	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new update wallet page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	aliasInput := textinput.New()
	aliasInput.Placeholder = "Enter new alias"
	aliasInput.Width = 40

	pkeyInput := textinput.New()
	pkeyInput.Placeholder = "Enter private key (hex format)"
	pkeyInput.EchoMode = textinput.EchoPassword
	pkeyInput.EchoCharacter = '*'
	pkeyInput.Width = 66

	return Model{
		router:        router,
		sharedMemory:  sharedMemory,
		walletService: walletService,
		loading:       true,
		currentStep:   stepSelectUpdate,
		selectedIndex: 0,
		aliasInput:    aliasInput,
		pkeyInput:     pkeyInput,
		options: []updateOption{
			{label: "Update alias", description: "Change the wallet name", step: stepUpdateAlias},
			{label: "Update private key", description: "Replace the private key (will change address)", step: stepPrivateKeyWarning},
			{label: "Cancel", description: "Go back without changes", step: -1},
		},
		confirmOpts: []confirmationOption{
			{label: "No, cancel", value: false},
			{label: "Yes, update private key", value: true},
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
	walletIDStr := m.router.GetQueryParam("id")
	if walletIDStr == "" {
		return walletLoadedMsg{err: fmt.Errorf("wallet ID not provided")}
	}

	walletID, err := strconv.ParseUint(walletIDStr, 10, 32)
	if err != nil {
		return walletLoadedMsg{err: fmt.Errorf("invalid wallet ID: %w", err)}
	}

	walletService := m.walletService
	if walletService == nil {
		svc, err := m.createWalletService()
		if err != nil {
			return walletLoadedMsg{err: err}
		}
		walletService = svc
	}

	rpcEndpoint := "http://localhost:8545"

	walletData, err := walletService.GetWalletWithBalance(uint(walletID), rpcEndpoint)
	if err != nil {
		return walletLoadedMsg{err: fmt.Errorf("failed to load wallet: %w", err)}
	}

	return walletLoadedMsg{
		walletID:      uint(walletID),
		wallet:        walletData,
		walletService: walletService,
	}
}

type walletLoadedMsg struct {
	walletID      uint
	wallet        *wallet.WalletWithBalance
	walletService wallet.WalletService
	err           error
}

type walletUpdatedMsg struct {
	success    bool
	updateType string
	oldAddress string
	newAddress string
	newBalance *big.Int
	err        error
}

func (m Model) updateAlias() tea.Msg {
	newAlias := m.aliasInput.Value()
	if newAlias == "" {
		return walletUpdatedMsg{success: false, err: fmt.Errorf("alias cannot be empty")}
	}

	err := m.walletService.UpdateWalletAlias(m.walletID, newAlias)
	if err != nil {
		return walletUpdatedMsg{success: false, err: err}
	}

	return walletUpdatedMsg{success: true, updateType: "alias"}
}

func (m Model) updatePrivateKey() tea.Msg {
	newPrivateKey := m.pkeyInput.Value()
	if newPrivateKey == "" {
		return walletUpdatedMsg{success: false, err: fmt.Errorf("private key cannot be empty")}
	}

	// Validate private key format
	err := m.walletService.ValidatePrivateKey(newPrivateKey)
	if err != nil {
		return walletUpdatedMsg{success: false, err: fmt.Errorf("invalid private key: %w", err)}
	}

	// Save old address
	oldAddress := m.wallet.Wallet.Address

	// Update private key
	err = m.walletService.UpdateWalletPrivateKey(m.walletID, newPrivateKey)
	if err != nil {
		return walletUpdatedMsg{success: false, err: err}
	}

	// Reload wallet to get new address and balance
	rpcEndpoint := "http://localhost:8545"
	updatedWallet, err := m.walletService.GetWalletWithBalance(m.walletID, rpcEndpoint)
	if err != nil {
		logger.Warn("Failed to reload wallet after update: %v", err)
		return walletUpdatedMsg{
			success:    true,
			updateType: "private_key",
			oldAddress: oldAddress,
		}
	}

	return walletUpdatedMsg{
		success:    true,
		updateType: "private_key",
		oldAddress: oldAddress,
		newAddress: updatedWallet.Wallet.Address,
		newBalance: updatedWallet.Balance,
	}
}

//nolint:gocognit,gocyclo // Update method complexity is acceptable for state machine
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

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

		// Set initial alias value
		m.aliasInput.SetValue(m.wallet.Wallet.Alias)

		return m, nil

	case walletUpdatedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		if msg.success {
			logger.Info("Wallet updated successfully: %s", msg.updateType)
			m.currentStep = stepSuccess
			m.oldAddress = msg.oldAddress
			m.newAddress = msg.newAddress

			if msg.newBalance != nil {
				ethValue := new(big.Float).Quo(
					new(big.Float).SetInt(msg.newBalance),
					new(big.Float).SetInt(big.NewInt(1e18)),
				)
				m.updatedBalance = fmt.Sprintf("%.4f ETH", ethValue)
			}
		}

		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch m.currentStep {
		case stepSelectUpdate:
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
				if option.step == -1 {
					// Cancel - go back
					return m, func() tea.Msg {
						_ = m.router.NavigateTo("/evm/wallet", nil)
						return nil
					}
				}
				m.currentStep = option.step
				m.selectedIndex = 0

				// Focus appropriate input
				if option.step == stepUpdateAlias {
					m.aliasInput.Focus()
					return m, textinput.Blink
				}
			}

		case stepUpdateAlias:
			m.aliasInput, cmd = m.aliasInput.Update(msg)

			switch msg.String() {
			case "enter":
				return m, m.updateAlias
			case "esc":
				m.currentStep = stepSelectUpdate
				m.errorMsg = ""
			}
			return m, cmd

		case stepPrivateKeyWarning:
			switch msg.String() {
			case "up", "k":
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}

			case "down", "j":
				if m.selectedIndex < len(m.confirmOpts)-1 {
					m.selectedIndex++
				}

			case "enter":
				option := m.confirmOpts[m.selectedIndex]
				if option.value {
					// User confirmed, proceed to enter new private key
					m.currentStep = stepUpdatePrivateKey
					m.pkeyInput.Focus()
					return m, textinput.Blink
				} else {
					// User cancelled, go back to selection
					m.currentStep = stepSelectUpdate
					m.selectedIndex = 0
				}

			case "esc":
				m.currentStep = stepSelectUpdate
				m.selectedIndex = 0
			}

		case stepUpdatePrivateKey:
			m.pkeyInput, cmd = m.pkeyInput.Update(msg)

			switch msg.String() {
			case "enter":
				return m, m.updatePrivateKey
			case "esc":
				m.currentStep = stepPrivateKeyWarning
				m.selectedIndex = 0
				m.errorMsg = ""
			}
			return m, cmd

		case stepSuccess:
			// Any key goes back to wallet list
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}
		}
	}

	return m, nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	if m.loading {
		return "Loading...", view.HelpDisplayOptionOverride
	}

	switch m.currentStep {
	case stepUpdateAlias, stepUpdatePrivateKey:
		return "enter: save • esc: cancel", view.HelpDisplayOptionOverride
	case stepSuccess:
		return "Press any key to return to wallet list...", view.HelpDisplayOptionOverride
	default:
		return "↑/k: up • ↓/j: down • enter: select • esc: cancel", view.HelpDisplayOptionAppend
	}
}

func (m Model) View() string {
	if m.loading {
		return component.VStackC(
			component.T("Update Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Loading wallet...").Muted(),
		).Render()
	}

	if m.errorMsg != "" && m.currentStep != stepUpdateAlias && m.currentStep != stepUpdatePrivateKey {
		return component.VStackC(
			component.T("Update Wallet").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
			component.SpacerV(1),
			component.T("Press 'esc' to go back").Muted(),
		).Render()
	}

	switch m.currentStep {
	case stepSelectUpdate:
		return m.renderSelectUpdate()
	case stepUpdateAlias:
		return m.renderUpdateAlias()
	case stepPrivateKeyWarning:
		return m.renderPrivateKeyWarning()
	case stepUpdatePrivateKey:
		return m.renderUpdatePrivateKey()
	case stepSuccess:
		return m.renderSuccess()
	default:
		return ""
	}
}

func (m Model) renderSelectUpdate() string {
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

		optionComponents = append(optionComponents, component.VStackC(
			labelStyle,
			component.T("  "+option.description).Muted(),
			component.SpacerV(1),
		))
	}

	return component.VStackC(
		component.T("Update Wallet - "+m.wallet.Wallet.Alias).Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Current Information:").Bold(true),
		component.T("• Alias: "+m.wallet.Wallet.Alias).Muted(),
		component.T("• Address: "+m.wallet.Wallet.Address).Muted(),
		component.SpacerV(1),
		component.T("What would you like to update?").Bold(true),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderUpdateAlias() string {
	errorDisplay := component.Empty()
	if m.errorMsg != "" {
		errorDisplay = component.VStackC(
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
		)
	}

	return component.VStackC(
		component.T("Update Wallet - Update Alias").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Current alias: "+m.wallet.Wallet.Alias).Muted(),
		component.SpacerV(1),
		component.T("Enter new alias:"),
		component.SpacerV(1),
		component.T("New Alias: "+m.aliasInput.View()),
		errorDisplay,
	).Render()
}

func (m Model) renderPrivateKeyWarning() string {
	// Build options
	optionComponents := make([]component.Component, 0)
	for i, option := range m.confirmOpts {
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
			description = "I understand the consequences"
		} else {
			description = "Keep the current private key"
		}

		optionComponents = append(optionComponents, component.VStackC(
			labelStyle,
			component.T("  "+description).Muted(),
			component.SpacerV(1),
		))
	}

	return component.VStackC(
		component.T("Update Wallet - Update Private Key").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("⚠ Warning: Changing Private Key").Warning(),
		component.SpacerV(1),
		component.T("Updating the private key will:"),
		component.T("• Change the wallet address").Muted(),
		component.T("• This wallet will control a different Ethereum account").Muted(),
		component.T("• You will lose access to the old address funds").Muted(),
		component.SpacerV(1),
		component.T("Current Address: "+m.wallet.Wallet.Address).Muted(),
		component.SpacerV(1),
		component.T("Are you sure you want to continue?").Bold(true),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderUpdatePrivateKey() string {
	errorDisplay := component.Empty()
	if m.errorMsg != "" {
		errorDisplay = component.VStackC(
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
		)
	}

	return component.VStackC(
		component.T("Update Wallet - Update Private Key").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Alias: "+m.wallet.Wallet.Alias).Muted(),
		component.SpacerV(1),
		component.T("Enter new private key (hex format, with or without 0x prefix):"),
		component.SpacerV(1),
		component.T("Private Key: "+m.pkeyInput.View()),
		component.SpacerV(1),
		component.T("⚠ This will replace the existing private key and change the address!").Warning(),
		errorDisplay,
	).Render()
}

func (m Model) renderSuccess() string {
	if m.oldAddress != "" && m.newAddress != "" {
		// Private key was updated
		balanceStr := "0.0000 ETH"
		if m.updatedBalance != "" {
			balanceStr = m.updatedBalance
		}

		return component.VStackC(
			component.T("Update Wallet - Success").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("✓ Wallet updated successfully!").Success(),
			component.SpacerV(1),
			component.T("Changes:").Bold(true),
			component.T("• Alias: "+m.wallet.Wallet.Alias+" (unchanged)").Muted(),
			component.T("• Old Address: "+m.oldAddress).Muted(),
			component.T("• New Address: "+m.newAddress).Muted(),
			component.T("• New Balance: "+balanceStr+" (on http://localhost:8545)").Muted(),
			component.SpacerV(1),
			component.T("⚠ Important: The old address is no longer accessible with this wallet!").Warning(),
		).Render()
	} else {
		// Alias was updated
		return component.VStackC(
			component.T("Update Wallet - Success").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("✓ Wallet alias updated successfully!").Success(),
			component.SpacerV(1),
			component.T("New alias: "+m.aliasInput.Value()).Bold(true),
		).Render()
	}
}
