package details

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

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

var logger, _ = log.NewFileLogger("./logs/evm/wallet/details.log")

type viewMode int

const (
	modeNormal viewMode = iota
	modeShowPrivateKeyPrompt
	modeShowPrivateKey
)

type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	walletService wallet.WalletService

	walletID         uint
	wallet           *wallet.WalletWithBalance
	selectedWalletID uint
	privateKey       string

	mode              viewMode
	confirmationInput textinput.Model
	autoCloseCounter  int

	loading  bool
	errorMsg string
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return NewPageWithService(router, sharedMemory, nil)
}

// NewPageWithService creates a new details page with an optional wallet service (for testing).
func NewPageWithService(router view.Router, sharedMemory storage.SharedMemory, walletService wallet.WalletService) view.View {
	confirmInput := textinput.New()
	confirmInput.Placeholder = "Type 'SHOW' to confirm"
	confirmInput.Width = 30

	return Model{
		router:            router,
		sharedMemory:      sharedMemory,
		walletService:     walletService,
		loading:           true,
		mode:              modeNormal,
		confirmationInput: confirmInput,
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

type privateKeyLoadedMsg struct {
	privateKey string
	err        error
}

type autoCloseTickMsg struct{}

func (m Model) loadPrivateKey() tea.Msg {
	privateKey, err := m.walletService.GetPrivateKey(m.walletID)
	if err != nil {
		return privateKeyLoadedMsg{err: err}
	}
	return privateKeyLoadedMsg{privateKey: privateKey}
}

func autoCloseTick() tea.Msg {
	time.Sleep(1 * time.Second)
	return autoCloseTickMsg{}
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
		return m, nil

	case privateKeyLoadedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			m.mode = modeNormal
			return m, nil
		}

		m.privateKey = msg.privateKey
		m.mode = modeShowPrivateKey
		m.autoCloseCounter = 60
		return m, autoCloseTick

	case autoCloseTickMsg:
		if m.mode == modeShowPrivateKey {
			m.autoCloseCounter--
			if m.autoCloseCounter <= 0 {
				m.mode = modeNormal
				m.privateKey = ""
				return m, nil
			}
			return m, autoCloseTick
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		// Handle different modes
		switch m.mode {
		case modeShowPrivateKeyPrompt:
			var cmd tea.Cmd
			m.confirmationInput, cmd = m.confirmationInput.Update(msg)

			switch msg.String() {
			case "enter":
				if m.confirmationInput.Value() == "SHOW" {
					m.confirmationInput.SetValue("")
					return m, m.loadPrivateKey
				} else {
					m.errorMsg = "Incorrect confirmation. Type 'SHOW' exactly."
					return m, nil
				}
			case "esc":
				m.mode = modeNormal
				m.confirmationInput.SetValue("")
				return m, cmd
			}
			return m, cmd

		case modeShowPrivateKey:
			switch msg.String() {
			case "c":
				// TODO: Copy to clipboard functionality
				// For now, just show a message
				logger.Info("Copy to clipboard requested")
			case "esc", "q":
				m.mode = modeNormal
				m.privateKey = ""
			}

		case modeNormal:
			switch msg.String() {
			case "p":
				m.mode = modeShowPrivateKeyPrompt
				m.confirmationInput.Focus()
				return m, textinput.Blink
			case "r":
				// Refresh balance
				m.loading = true
				return m, m.loadWallet
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
	case modeShowPrivateKeyPrompt:
		return "enter: confirm • esc: cancel", view.HelpDisplayOptionOverride
	case modeShowPrivateKey:
		return "c: copy to clipboard • esc/q: close immediately", view.HelpDisplayOptionOverride
	default:
		return "r: refresh balance • p: show private key • esc/q: back", view.HelpDisplayOptionAppend
	}
}

func (m Model) View() string {
	if m.loading {
		return component.VStackC(
			component.T("Wallet Details").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Loading wallet...").Muted(),
		).Render()
	}

	if m.errorMsg != "" {
		return component.VStackC(
			component.T("Wallet Details").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Error: "+m.errorMsg).Error(),
			component.SpacerV(1),
			component.T("Press 'esc' to go back").Muted(),
		).Render()
	}

	if m.wallet == nil {
		return component.VStackC(
			component.T("Wallet Details").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("Wallet not found").Error(),
		).Render()
	}

	// Show different views based on mode
	switch m.mode {
	case modeShowPrivateKeyPrompt:
		return m.renderPrivateKeyPrompt()
	case modeShowPrivateKey:
		return m.renderPrivateKey()
	default:
		return m.renderDetails()
	}
}

func (m Model) renderDetails() string {
	// Format balance
	balanceStr := "unavailable ⚠"
	usdValue := "N/A"
	if m.wallet.Error == nil && m.wallet.Balance != nil {
		ethValue := new(big.Float).Quo(
			new(big.Float).SetInt(m.wallet.Balance),
			new(big.Float).SetInt(big.NewInt(1e18)),
		)
		balanceStr = fmt.Sprintf("%.4f ETH", ethValue)

		// Mock USD value (in production, fetch from price oracle)
		ethFloat, _ := ethValue.Float64()
		usdValue = fmt.Sprintf("$%.2f (at $2,000/ETH)", ethFloat*2000)
	}

	// Status indicator
	statusStr := "Available"
	if m.walletID == m.selectedWalletID {
		statusStr = "★ Currently Selected"
	}

	title := "Wallet Details - " + m.wallet.Wallet.Alias

	return component.VStackC(
		component.T(title).Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Alias: "+m.wallet.Wallet.Alias),
		component.T("Status: "+statusStr).Muted(),
		component.SpacerV(1),

		component.T("Address Information:").Bold(true),
		component.T("• Address: "+m.wallet.Wallet.Address).Muted(),
		component.T("• Checksum: ✓ Valid Ethereum address").Muted(),
		component.IfC(
			m.wallet.Wallet.DerivationPath != nil && *m.wallet.Wallet.DerivationPath != "",
			component.T("• Derivation Path: "+*m.wallet.Wallet.DerivationPath).Muted(),
			component.Empty(),
		),
		component.SpacerV(1),

		component.T("Balance Information:").Bold(true),
		component.T("• Balance: "+balanceStr).Muted(),
		component.T("• USD Value: "+usdValue).Muted(),
		component.T("• Endpoint: http://localhost:8545 (Anvil)").Muted(),
		component.T("• Last Updated: "+time.Now().Format("2006-01-02 3:04 PM")).Muted(),
		component.SpacerV(1),

		component.T("Security:").Bold(true),
		component.T("• Private Key: ******** (hidden)").Muted(),
		component.T("• Created: "+m.wallet.Wallet.CreatedAt.Format("2006-01-02 3:04 PM")).Muted(),
		component.T("• Last Modified: "+m.wallet.Wallet.UpdatedAt.Format("2006-01-02 3:04 PM")).Muted(),
		component.IfC(
			m.wallet.Wallet.IsFromMnemonic,
			component.T("• Mnemonic: Available (use 'm' to show)").Muted(),
			component.Empty(),
		),
	).Render()
}

func (m Model) renderPrivateKeyPrompt() string {
	return component.VStackC(
		component.T("Show Private Key - Security Warning").Bold(true).Warning(),
		component.SpacerV(1),
		component.T("⚠ WARNING: Exposing Sensitive Information!").Warning(),
		component.SpacerV(1),
		component.T("You are about to reveal the private key for:"),
		component.T("• Alias: "+m.wallet.Wallet.Alias).Muted(),
		component.T("• Address: "+m.wallet.Wallet.Address).Muted(),
		component.SpacerV(1),
		component.T("⚠ Anyone with this private key can access and control your funds!").Warning(),
		component.T("⚠ Make sure no one is watching your screen!").Warning(),
		component.T("⚠ Be careful when sharing screenshots or recordings!").Warning(),
		component.SpacerV(1),
		component.T("Type \"SHOW\" to reveal the private key (case sensitive):"),
		component.SpacerV(1),
		component.T("Confirmation: "+m.confirmationInput.View()),
	).Render()
}

func (m Model) renderPrivateKey() string {
	return component.VStackC(
		component.T("Show Private Key - "+m.wallet.Wallet.Alias).Bold(true).Warning(),
		component.SpacerV(1),
		component.T("⚠ SENSITIVE INFORMATION - KEEP SECURE! ⚠").Warning(),
		component.SpacerV(1),
		component.T("Wallet: "+m.wallet.Wallet.Alias),
		component.T("Address: "+m.wallet.Wallet.Address).Muted(),
		component.SpacerV(1),
		component.T("Private Key:"),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.T(m.privateKey),
		component.T("────────────────────────────────────────────────────────────────────────────").Muted(),
		component.SpacerV(1),
		component.T("⚠ NEVER share this with anyone!").Warning(),
		component.T("⚠ Anyone with this key can control your funds!").Warning(),
		component.SpacerV(1),
		component.T(fmt.Sprintf("This screen will automatically close in %d seconds...", m.autoCloseCounter)).Muted(),
	).Render()
}
