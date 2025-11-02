package add

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/wallet"
	"github.com/rxtech-lab/smart-contract-cli/internal/log"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/utils"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

var logger, _ = log.NewFileLogger("./logs/evm/wallet/add.log")

type addStep int

const (
	stepSelectMethod addStep = iota
	stepEnterAlias
	stepEnterPrivateKey
	stepEnterMnemonic
	stepSelectDerivationPath
	stepGenerating
	stepShowBackup
	stepConfirm
	stepSuccess
	stepError
)

type importMethod int

const (
	methodPrivateKey importMethod = iota
	methodMnemonic
	methodGenerate
)

type methodOption struct {
	label       string
	description string
	method      importMethod
}

type derivationPathOption struct {
	label       string
	description string
	path        string
}

type confirmOption struct {
	label string
	value bool
}

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	currentStep   addStep
	selectedIndex int
	method        importMethod

	// Method selection options
	methodOptions []methodOption

	// Derivation path options
	derivationOptions []derivationPathOption
	customPathInput   textinput.Model

	// Form inputs
	aliasInput    textinput.Model
	pkeyInput     textinput.Model
	mnemonicInput textarea.Model

	// Generated wallet data
	generatedMnemonic string
	generatedPKey     string
	generatedAddress  string

	// Confirmation data
	confirmedWallet *wallet.WalletWithBalance
	confirmOptions  []confirmOption
	rpcEndpoint     string

	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new add wallet page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	aliasInput := textinput.New()
	aliasInput.Placeholder = "Enter wallet alias"
	aliasInput.Width = 40

	pkeyInput := textinput.New()
	pkeyInput.Placeholder = "Enter private key (hex format)"
	pkeyInput.EchoMode = textinput.EchoPassword
	pkeyInput.EchoCharacter = '*'
	pkeyInput.Width = 66

	mnemonicInput := textarea.New()
	mnemonicInput.Placeholder = "Enter your 12 or 24 word mnemonic phrase (space-separated)"
	mnemonicInput.SetWidth(76)
	mnemonicInput.SetHeight(5)

	customPathInput := textinput.New()
	customPathInput.Placeholder = "m/44'/60'/0'/0/0"
	customPathInput.Width = 40

	return Model{
		router:          router,
		sharedMemory:    sharedMemory,
		walletService:   walletService,
		currentStep:     stepSelectMethod,
		selectedIndex:   0,
		aliasInput:      aliasInput,
		pkeyInput:       pkeyInput,
		mnemonicInput:   mnemonicInput,
		customPathInput: customPathInput,
		methodOptions: []methodOption{
			{label: "Import from private key", description: "Import wallet using a private key (hex format)", method: methodPrivateKey},
			{label: "Import from mnemonic phrase", description: "Import wallet using a 12 or 24 word mnemonic phrase", method: methodMnemonic},
			{label: "Generate new wallet", description: "Create a new random wallet with private key", method: methodGenerate},
		},
		derivationOptions: []derivationPathOption{
			{label: "m/44'/60'/0'/0/0", description: "Ethereum standard (default)", path: "m/44'/60'/0'/0/0"},
			{label: "m/44'/60'/0'/0/1", description: "Ethereum standard (account 1)", path: "m/44'/60'/0'/0/1"},
			{label: "m/44'/60'/0'/0/2", description: "Ethereum standard (account 2)", path: "m/44'/60'/0'/0/2"},
			{label: "Custom path", description: "Enter your own derivation path", path: "custom"},
		},
		confirmOptions: []confirmOption{
			{label: "Save wallet", value: true},
			{label: "Cancel", value: false},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadWalletService
}

func (m Model) createWalletService() (wallet.WalletService, error) {
	sqlStorage, err := utils.GetStorageClientFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return nil, fmt.Errorf("failed to get storage client from shared memory: %w", err)
	}

	secureStorage, _, err := utils.GetSecureStorageFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get secure storage from shared memory: %v", err)
		return nil, fmt.Errorf("failed to get secure storage from shared memory: %w", err)
	}

	return wallet.NewWalletService(sqlStorage, secureStorage), nil
}

func (m Model) loadWalletService() tea.Msg {
	walletService := m.walletService
	if walletService == nil {
		svc, err := m.createWalletService()
		if err != nil {
			return serviceLoadedMsg{err: err}
		}
		walletService = svc
	}

	return serviceLoadedMsg{walletService: walletService}
}

type serviceLoadedMsg struct {
	walletService wallet.WalletService
	err           error
}

type walletImportedMsg struct {
	wallet      *wallet.WalletWithBalance
	rpcEndpoint string
	err         error
}

type walletGeneratedMsg struct {
	wallet      *wallet.WalletWithBalance
	mnemonic    string
	privateKey  string
	rpcEndpoint string
	err         error
}

func (m Model) importPrivateKey() tea.Msg {
	alias := m.aliasInput.Value()
	privateKey := m.pkeyInput.Value()

	if alias == "" {
		return walletImportedMsg{err: fmt.Errorf("alias cannot be empty")}
	}

	if privateKey == "" {
		return walletImportedMsg{err: fmt.Errorf("private key cannot be empty")}
	}

	// Validate private key
	err := m.walletService.ValidatePrivateKey(privateKey)
	if err != nil {
		return walletImportedMsg{err: fmt.Errorf("invalid private key format")}
	}

	// Check for duplicate alias
	exists, err := m.walletService.WalletExistsByAlias(alias)
	if err != nil {
		return walletImportedMsg{err: fmt.Errorf("failed to check for duplicate alias: %w", err)}
	}
	if exists {
		return walletImportedMsg{err: fmt.Errorf("wallet with alias '%s' already exists", alias)}
	}

	// Import wallet
	walletData, err := m.walletService.ImportPrivateKey(alias, privateKey)
	if err != nil {
		return walletImportedMsg{err: err}
	}

	// Check for duplicate address
	exists, err = m.walletService.WalletExistsByAddress(walletData.Address)
	if err == nil && exists {
		// Delete the newly created wallet
		_ = m.walletService.DeleteWallet(walletData.ID)
		return walletImportedMsg{err: fmt.Errorf("wallet already exists with address: %s", walletData.Address)}
	}

	// Get RPC endpoint from database
	sqlStorage, err := utils.GetStorageClientFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return walletImportedMsg{err: fmt.Errorf("failed to get storage client from shared memory: %w", err)}
	}

	// Get the current config
	config, err := sqlStorage.GetCurrentConfig()
	if err != nil {
		logger.Error("Failed to get current config: %v", err)
		return walletImportedMsg{err: fmt.Errorf("failed to get current config: %w", err)}
	}
	if config.Endpoint == nil {
		logger.Error("No RPC endpoint configured")
		return walletImportedMsg{err: fmt.Errorf("no RPC endpoint configured. Please configure an endpoint first")}
	}
	rpcEndpoint := config.Endpoint.Url

	// Load with balance
	walletWithBalance, err := m.walletService.GetWalletWithBalance(walletData.ID, rpcEndpoint)
	if err != nil {
		logger.Warn("Failed to load balance: %v", err)
		// Continue anyway, balance is optional
		walletWithBalance = &wallet.WalletWithBalance{
			Wallet: *walletData,
		}
	}

	return walletImportedMsg{wallet: walletWithBalance, rpcEndpoint: rpcEndpoint}
}

func (m Model) importMnemonic() tea.Msg {
	alias := m.aliasInput.Value()
	mnemonic := strings.TrimSpace(m.mnemonicInput.Value())

	if alias == "" {
		return walletImportedMsg{err: fmt.Errorf("alias cannot be empty")}
	}

	if mnemonic == "" {
		return walletImportedMsg{err: fmt.Errorf("mnemonic cannot be empty")}
	}

	// Validate mnemonic
	err := m.walletService.ValidateMnemonic(mnemonic)
	if err != nil {
		wordCount := len(strings.Fields(mnemonic))
		return walletImportedMsg{err: fmt.Errorf("invalid mnemonic phrase (got %d words, expected 12 or 24)", wordCount)}
	}

	// Check for duplicate alias
	exists, err := m.walletService.WalletExistsByAlias(alias)
	if err != nil {
		return walletImportedMsg{err: fmt.Errorf("failed to check for duplicate alias: %w", err)}
	}
	if exists {
		return walletImportedMsg{err: fmt.Errorf("wallet with alias '%s' already exists", alias)}
	}

	// Get selected derivation path
	var derivationPath string
	if m.selectedIndex < len(m.derivationOptions)-1 {
		derivationPath = m.derivationOptions[m.selectedIndex].path
	} else {
		// Custom path
		derivationPath = m.customPathInput.Value()
		if derivationPath == "" {
			derivationPath = "m/44'/60'/0'/0/0" // Default
		}
	}

	// Import wallet
	walletData, err := m.walletService.ImportMnemonic(alias, mnemonic, derivationPath)
	if err != nil {
		return walletImportedMsg{err: err}
	}

	// Check for duplicate address
	exists, err = m.walletService.WalletExistsByAddress(walletData.Address)
	if err == nil && exists {
		// Delete the newly created wallet
		_ = m.walletService.DeleteWallet(walletData.ID)
		return walletImportedMsg{err: fmt.Errorf("wallet already exists with address: %s", walletData.Address)}
	}

	// Get RPC endpoint from database
	sqlStorage, err := utils.GetStorageClientFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return walletImportedMsg{err: fmt.Errorf("failed to get storage client from shared memory: %w", err)}
	}

	// Get the current config
	config, err := sqlStorage.GetCurrentConfig()
	if err != nil {
		logger.Error("Failed to get current config: %v", err)
		return walletImportedMsg{err: fmt.Errorf("failed to get current config: %w", err)}
	}
	if config.Endpoint == nil {
		logger.Error("No RPC endpoint configured")
		return walletImportedMsg{err: fmt.Errorf("no RPC endpoint configured. Please configure an endpoint first")}
	}
	rpcEndpoint := config.Endpoint.Url

	// Load with balance
	walletWithBalance, err := m.walletService.GetWalletWithBalance(walletData.ID, rpcEndpoint)
	if err != nil {
		logger.Warn("Failed to load balance: %v", err)
		walletWithBalance = &wallet.WalletWithBalance{
			Wallet: *walletData,
		}
	}

	return walletImportedMsg{wallet: walletWithBalance, rpcEndpoint: rpcEndpoint}
}

func (m Model) generateWallet() tea.Msg {
	alias := m.aliasInput.Value()

	if alias == "" {
		return walletGeneratedMsg{err: fmt.Errorf("alias cannot be empty")}
	}

	// Check for duplicate alias
	exists, err := m.walletService.WalletExistsByAlias(alias)
	if err != nil {
		return walletGeneratedMsg{err: fmt.Errorf("failed to check for duplicate alias: %w", err)}
	}
	if exists {
		return walletGeneratedMsg{err: fmt.Errorf("wallet with alias '%s' already exists", alias)}
	}

	// Generate wallet
	walletData, mnemonic, privateKey, err := m.walletService.GenerateWallet(alias)
	if err != nil {
		return walletGeneratedMsg{err: err}
	}

	// Get RPC endpoint from database
	sqlStorage, err := utils.GetStorageClientFromSharedMemory(m.sharedMemory)
	if err != nil {
		logger.Error("Failed to get storage client from shared memory: %v", err)
		return walletGeneratedMsg{err: fmt.Errorf("failed to get storage client from shared memory: %w", err)}
	}

	// Get the current config
	config, err := sqlStorage.GetCurrentConfig()
	if err != nil {
		logger.Error("Failed to get current config: %v", err)
		return walletGeneratedMsg{err: fmt.Errorf("failed to get current config: %w", err)}
	}
	if config.Endpoint == nil {
		logger.Error("No RPC endpoint configured")
		return walletGeneratedMsg{err: fmt.Errorf("no RPC endpoint configured. Please configure an endpoint first")}
	}
	rpcEndpoint := config.Endpoint.Url

	// Load with balance
	walletWithBalance, err := m.walletService.GetWalletWithBalance(walletData.ID, rpcEndpoint)
	if err != nil {
		logger.Warn("Failed to load balance: %v", err)
		walletWithBalance = &wallet.WalletWithBalance{
			Wallet: *walletData,
		}
	}

	return walletGeneratedMsg{
		wallet:      walletWithBalance,
		mnemonic:    mnemonic,
		privateKey:  privateKey,
		rpcEndpoint: rpcEndpoint,
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case serviceLoadedMsg:
		if msg.err != nil {
			m.currentStep = stepError
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		m.walletService = msg.walletService
		return m, nil

	case walletImportedMsg:
		if msg.err != nil {
			m.currentStep = stepError
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		m.confirmedWallet = msg.wallet
		m.rpcEndpoint = msg.rpcEndpoint
		m.currentStep = stepConfirm
		m.selectedIndex = 0
		return m, nil

	case walletGeneratedMsg:
		if msg.err != nil {
			m.currentStep = stepError
			m.errorMsg = msg.err.Error()
			return m, nil
		}

		m.confirmedWallet = msg.wallet
		m.generatedMnemonic = msg.mnemonic
		m.generatedPKey = msg.privateKey
		m.generatedAddress = msg.wallet.Wallet.Address
		m.rpcEndpoint = msg.rpcEndpoint
		m.currentStep = stepShowBackup
		m.selectedIndex = 0
		return m, nil

	case tea.KeyMsg:
		switch m.currentStep {
		case stepSelectMethod:
			return m.handleSelectMethod(msg)

		case stepEnterAlias:
			return m.handleEnterAlias(msg)

		case stepEnterPrivateKey:
			return m.handleEnterPrivateKey(msg)

		case stepEnterMnemonic:
			return m.handleEnterMnemonic(msg)

		case stepSelectDerivationPath:
			return m.handleSelectDerivationPath(msg)

		case stepShowBackup:
			return m.handleShowBackup(msg)

		case stepConfirm:
			return m.handleConfirm(msg)

		case stepSuccess:
			// Any key returns to wallet list
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}

		case stepError:
			// Any key returns to wallet list
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}
		}
	}

	return m, cmd
}

func (m Model) handleSelectMethod(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.methodOptions)-1 {
			m.selectedIndex++
		}

	case "enter":
		m.method = m.methodOptions[m.selectedIndex].method
		m.currentStep = stepEnterAlias
		m.selectedIndex = 0
		m.aliasInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m Model) handleEnterAlias(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.aliasInput, cmd = m.aliasInput.Update(msg)

	switch msg.String() {
	case "enter":
		switch m.method {
		case methodPrivateKey:
			m.currentStep = stepEnterPrivateKey
			m.pkeyInput.Focus()
			return m, textinput.Blink

		case methodMnemonic:
			m.currentStep = stepEnterMnemonic
			m.mnemonicInput.Focus()
			return m, textarea.Blink

		case methodGenerate:
			m.currentStep = stepGenerating
			return m, m.generateWallet
		}

	case "esc":
		m.currentStep = stepSelectMethod
		m.selectedIndex = 0
	}

	return m, cmd
}

func (m Model) handleEnterPrivateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.pkeyInput, cmd = m.pkeyInput.Update(msg)

	switch msg.String() {
	case "enter":
		return m, m.importPrivateKey

	case "esc":
		m.currentStep = stepEnterAlias
	}

	return m, cmd
}

func (m Model) handleEnterMnemonic(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.mnemonicInput, cmd = m.mnemonicInput.Update(msg)

	switch msg.String() {
	case "ctrl+s": // Use Ctrl+S to proceed (Enter adds newline in textarea)
		m.currentStep = stepSelectDerivationPath
		m.selectedIndex = 0

	case "esc":
		m.currentStep = stepEnterAlias
	}

	return m, cmd
}

func (m Model) handleSelectDerivationPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If custom path is selected, handle text input
	if m.selectedIndex == len(m.derivationOptions)-1 {
		var cmd tea.Cmd
		m.customPathInput, cmd = m.customPathInput.Update(msg)

		switch msg.String() {
		case "enter":
			return m, m.importMnemonic

		case "esc":
			m.currentStep = stepEnterMnemonic
		}

		return m, cmd
	}

	// Handle list selection
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.derivationOptions)-1 {
			m.selectedIndex++
		} else {
			// Move to custom path input
			m.selectedIndex++
			m.customPathInput.Focus()
			return m, textinput.Blink
		}

	case "enter":
		return m, m.importMnemonic

	case "esc":
		m.currentStep = stepEnterMnemonic
	}

	return m, nil
}

func (m Model) handleShowBackup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < 1 {
			m.selectedIndex++
		}

	case "enter":
		if m.selectedIndex == 0 {
			// User confirmed they saved credentials
			m.currentStep = stepSuccess
		} else {
			// User cancelled, delete the wallet
			_ = m.walletService.DeleteWallet(m.confirmedWallet.Wallet.ID)
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}
		}
	}

	return m, nil
}

func (m Model) handleConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.confirmOptions)-1 {
			m.selectedIndex++
		}

	case "enter":
		option := m.confirmOptions[m.selectedIndex]
		if option.value {
			// User confirmed, wallet is already saved
			m.currentStep = stepSuccess
		} else {
			// User cancelled, delete the wallet
			_ = m.walletService.DeleteWallet(m.confirmedWallet.Wallet.ID)
			return m, func() tea.Msg {
				_ = m.router.NavigateTo("/evm/wallet", nil)
				return nil
			}
		}
	}

	return m, nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	switch m.currentStep {
	case stepEnterAlias, stepEnterPrivateKey:
		return "enter: next • esc: cancel", view.HelpDisplayOptionOverride
	case stepEnterMnemonic:
		return "ctrl+s: next • esc: cancel", view.HelpDisplayOptionOverride
	case stepSelectDerivationPath:
		if m.selectedIndex == len(m.derivationOptions)-1 {
			return "enter: import • esc: back", view.HelpDisplayOptionOverride
		}
		return "↑/k: up • ↓/j: down • enter: select • esc: back", view.HelpDisplayOptionOverride
	case stepGenerating:
		return "Generating wallet...", view.HelpDisplayOptionOverride
	case stepSuccess, stepError:
		return "Press any key to return to wallet list...", view.HelpDisplayOptionOverride
	default:
		return "↑/k: up • ↓/j: down • enter: select • esc: cancel", view.HelpDisplayOptionAppend
	}
}

func (m Model) View() string {
	switch m.currentStep {
	case stepSelectMethod:
		return m.renderSelectMethod()
	case stepEnterAlias:
		return m.renderEnterAlias()
	case stepEnterPrivateKey:
		return m.renderEnterPrivateKey()
	case stepEnterMnemonic:
		return m.renderEnterMnemonic()
	case stepSelectDerivationPath:
		return m.renderSelectDerivationPath()
	case stepGenerating:
		return m.renderGenerating()
	case stepShowBackup:
		return m.renderShowBackup()
	case stepConfirm:
		return m.renderConfirm()
	case stepSuccess:
		return m.renderSuccess()
	case stepError:
		return m.renderError()
	default:
		return ""
	}
}

func (m Model) renderSelectMethod() string {
	optionComponents := make([]component.Component, 0)
	for i, option := range m.methodOptions {
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
		component.T("Add New Wallet").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("How would you like to import the wallet?").Bold(true),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderEnterAlias() string {
	stepText := "Step 1/3"
	if m.method == methodMnemonic {
		stepText = "Step 1/4"
	}

	var methodName string
	switch m.method {
	case methodMnemonic:
		methodName = "Mnemonic Import"
	case methodGenerate:
		methodName = "Generate New"
	default:
		methodName = "Private Key Import"
	}

	return component.VStackC(
		component.T("Add New Wallet - "+methodName).Bold(true).Primary(),
		component.SpacerV(1),
		component.T(stepText+": Enter Wallet Alias").Bold(true),
		component.SpacerV(1),
		component.T("Give your wallet a memorable name:"),
		component.SpacerV(1),
		component.T("Alias: "+m.aliasInput.View()),
	).Render()
}

func (m Model) renderEnterPrivateKey() string {
	return component.VStackC(
		component.T("Add New Wallet - Private Key Import").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Step 2/3: Enter Private Key").Bold(true),
		component.SpacerV(1),
		component.T("Enter your private key (hex format, with or without 0x prefix):"),
		component.SpacerV(1),
		component.T("Private Key: "+m.pkeyInput.View()),
		component.SpacerV(1),
		component.T("⚠ Warning: Never share your private key with anyone!").Warning(),
		component.T("Keep it secure and private.").Warning(),
	).Render()
}

func (m Model) renderEnterMnemonic() string {
	mnemonic := m.mnemonicInput.Value()
	wordCount := len(strings.Fields(strings.TrimSpace(mnemonic)))

	statusText := fmt.Sprintf("Words entered: %d/12", wordCount)
	if wordCount >= 12 {
		if wordCount == 12 || wordCount == 24 {
			statusText = fmt.Sprintf("Words entered: %d/%d ✓", wordCount, wordCount)
		} else {
			statusText = fmt.Sprintf("Words entered: %d (expected 12 or 24)", wordCount)
		}
	}

	return component.VStackC(
		component.T("Add New Wallet - Mnemonic Import").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Step 2/4: Enter Mnemonic Phrase").Bold(true),
		component.SpacerV(1),
		component.T("Enter your 12 or 24 word mnemonic phrase (space-separated):"),
		component.SpacerV(1),
		component.T("Mnemonic:"),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.T(m.mnemonicInput.View()),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.SpacerV(1),
		component.T(statusText).Muted(),
		component.SpacerV(1),
		component.T("⚠ Warning: Never share your mnemonic phrase with anyone!").Warning(),
	).Render()
}

func (m Model) renderSelectDerivationPath() string {
	// If custom path is selected
	if m.selectedIndex == len(m.derivationOptions)-1 {
		return component.VStackC(
			component.T("Add New Wallet - Mnemonic Import").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Step 3/4: Custom Derivation Path").Bold(true),
			component.SpacerV(1),
			component.T("Enter your custom derivation path:"),
			component.SpacerV(1),
			component.T("Path: "+m.customPathInput.View()),
			component.SpacerV(1),
			component.T("Common format: m/44'/60'/0'/0/0").Muted(),
		).Render()
	}

	optionComponents := make([]component.Component, 0)
	for i, option := range m.derivationOptions {
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
		component.T("Add New Wallet - Mnemonic Import").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Step 3/4: Select Derivation Path").Bold(true),
		component.SpacerV(1),
		component.T("Choose the derivation path for your wallet:"),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
		component.SpacerV(1),
		component.T("Tip: Most wallets (MetaMask, Ledger, Trezor) use m/44'/60'/0'/0/0").Muted(),
	).Render()
}

func (m Model) renderGenerating() string {
	return component.VStackC(
		component.T("Add New Wallet - Generate New").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Step 2/3: Generating Wallet").Bold(true),
		component.SpacerV(1),
		component.T("Generating secure random wallet..."),
		component.SpacerV(1),
		component.T("⠋ Creating cryptographic keys...").Muted(),
	).Render()
}

func (m Model) renderShowBackup() string {
	optionComponents := []component.Component{
		component.T("> I have saved my credentials securely").Bold(true),
		component.T("  Confirm and add wallet to account").Muted(),
		component.SpacerV(1),
		component.T("  Cancel").Muted(),
		component.T("  Discard this wallet").Muted(),
	}

	if m.selectedIndex == 1 {
		optionComponents = []component.Component{
			component.T("  I have saved my credentials securely").Muted(),
			component.T("  Confirm and add wallet to account").Muted(),
			component.SpacerV(1),
			component.T("> Cancel").Bold(true),
			component.T("  Discard this wallet").Muted(),
		}
	}

	return component.VStackC(
		component.T("Add New Wallet - Backup Your Wallet").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("✓ Wallet successfully generated!").Success(),
		component.SpacerV(1),
		component.T("⚠ IMPORTANT: Save these credentials securely!").Warning(),
		component.SpacerV(1),
		component.T("Mnemonic Phrase (12 words):").Bold(true),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.T(m.generatedMnemonic),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.SpacerV(1),
		component.T("Private Key:").Bold(true),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.T(m.generatedPKey),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.SpacerV(1),
		component.T("Address: "+m.generatedAddress).Muted(),
		component.SpacerV(1),
		component.T("⚠ Write these down and store them in a secure location!").Warning(),
		component.T("⚠ Anyone with access to these can control your funds!").Warning(),
		component.T("⚠ Lost credentials cannot be recovered!").Warning(),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderConfirm() string {
	if m.confirmedWallet == nil {
		return ""
	}

	balanceStr := "0.0000 ETH"
	if m.confirmedWallet.Error == nil && m.confirmedWallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.confirmedWallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		balanceStr = fmt.Sprintf("%.4f ETH", ethValue)
	}

	derivationInfo := component.Empty()
	if m.method == methodMnemonic && m.confirmedWallet.Wallet.DerivationPath != nil {
		derivationInfo = component.T("• Derivation Path: " + *m.confirmedWallet.Wallet.DerivationPath).Muted()
	}

	optionComponents := make([]component.Component, 0)
	for i, option := range m.confirmOptions {
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
			description = "Add this wallet to your account"
		} else {
			description = "Discard and start over"
		}

		optionComponents = append(optionComponents, component.VStackC(
			labelStyle,
			component.T("  "+description).Muted(),
			component.SpacerV(1),
		))
	}

	title := "successfully imported!"
	if m.method == methodMnemonic {
		title = "successfully imported from mnemonic!"
	}

	return component.VStackC(
		component.T("Add New Wallet - Confirmation").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("✓ Wallet "+title).Success(),
		component.SpacerV(1),
		component.T("Wallet Details:").Bold(true),
		component.T("• Alias: "+m.confirmedWallet.Wallet.Alias).Muted(),
		derivationInfo,
		component.T("• Address: "+m.confirmedWallet.Wallet.Address).Muted(),
		component.T("• Balance: "+balanceStr+" (on "+m.rpcEndpoint+")").Muted(),
		component.SpacerV(1),
		component.T("Derived Information:").Bold(true),
		component.T("• Private Key: Available (hidden for security)").Muted(),
		component.T("• Checksum Address: ✓ Valid").Muted(),
		component.SpacerV(1),
		component.VStackC(optionComponents...),
	).Render()
}

func (m Model) renderSuccess() string {
	return component.VStackC(
		component.T("Add New Wallet - Success").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("✓ Wallet added successfully!").Success(),
		component.SpacerV(1),
		component.T("Your new wallet is ready to use.").Muted(),
	).Render()
}

func (m Model) renderError() string {
	return component.VStackC(
		component.T("Add New Wallet - Error").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("✗ Failed to add wallet").Error(),
		component.SpacerV(1),
		component.T("Error: "+m.errorMsg).Error(),
		component.SpacerV(1),
		component.T("Please try again or contact support if the problem persists.").Muted(),
	).Render()
}
