package storage

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/log"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

var logger, err = log.NewFileLogger("./logs/evm/storage/page.log")

// InputMode represents the current input mode of the page.
type InputMode int

const (
	InputModeNone InputMode = iota
	InputModeSqlitePath
	InputModePostgresURL
	InputModeConfirmation
)

// StorageOption represents a storage client option.
type StorageOption struct {
	Label       string
	Value       types.StorageClient
	Description string
}

var storageOptions = []StorageOption{
	{
		Label:       "SQLite",
		Value:       types.StorageClientSQLite,
		Description: "Local file-based database",
	},
	{
		Label:       "Postgres",
		Value:       types.StorageClientPostgres,
		Description: "PostgreSQL database server",
	},
}

// Model represents the storage client page model.
type Model struct {
	router        view.Router
	sharedMemory  storage.SharedMemory
	secureStorage storage.SecureStorage

	// UI state
	selectedIndex int
	options       []StorageOption

	// Input mode
	inputMode      InputMode
	textInput      textinput.Model
	confirmOptions []string
	confirmIndex   int

	// Storage state
	activeClient types.StorageClient // "sqlite" or "postgres" (currently in use)
	sqlitePath   string              // Loaded from secure storage
	postgresURL  string              // Loaded from secure storage

	errorMessage string
}

// Storage keys for secure storage.

// NewPage creates a new storage client page.
func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	// Get password from shared memory
	passwordRaw, err := sharedMemory.Get(config.SecureStoragePasswordKey)
	if err != nil {
		return Model{
			router:       router,
			sharedMemory: sharedMemory,
			options:      storageOptions,
			errorMessage: "Failed to get password from shared memory",
		}
	}

	password, ok := passwordRaw.(string)
	if !ok {
		return Model{
			router:       router,
			sharedMemory: sharedMemory,
			options:      storageOptions,
			errorMessage: "Password in shared memory is not a string",
		}
	}

	// Initialize secure storage
	secureStorage, err := storage.NewSecureStorageWithEncryption(password, "")
	if err != nil {
		return Model{
			router:       router,
			sharedMemory: sharedMemory,
			options:      storageOptions,
			errorMessage: fmt.Sprintf("Failed to initialize secure storage: %v", err),
		}
	}

	// Create text input for later use
	textInput := textinput.New()
	textInput.Focus()
	textInput.CharLimit = 256
	textInput.Width = 50

	model := Model{
		router:        router,
		sharedMemory:  sharedMemory,
		secureStorage: secureStorage,
		options:       storageOptions,
		selectedIndex: 0,
		inputMode:     InputModeNone,
		textInput:     textInput,
		confirmOptions: []string{
			"Use existing configuration",
			"Change configuration",
			"Remove configuration",
		},
		confirmIndex: 0,
	}

	// Load existing configuration from secure storage
	model.loadFromSecureStorage()

	return model
}

// loadFromSecureStorage loads existing configuration from secure storage.
func (m *Model) loadFromSecureStorage() {
	if m.secureStorage == nil {
		return
	}

	// Load active client type
	if clientType, err := m.secureStorage.Get(config.SecureStorageClientTypeKey); err == nil {
		m.activeClient = types.StorageClient(clientType)
	}

	// Load SQLite path
	if sqlitePath, err := m.secureStorage.Get(config.SecureStorageKeySqlitePathKey); err == nil {
		m.sqlitePath = sqlitePath
	}

	// Load Postgres URL
	if postgresURL, err := m.secureStorage.Get(config.SecureStorageKeyPostgresURLKey); err == nil {
		m.postgresURL = postgresURL
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		// Update text input if in input mode
		if m.inputMode == InputModeSqlitePath || m.inputMode == InputModePostgresURL {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Handle quit (Ctrl+C only, 'q' navigates back)
	if keyMsg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// Handle input based on mode
	switch m.inputMode {
	case InputModeNone:
		return m.handleNormalMode(keyMsg)
	case InputModeSqlitePath, InputModePostgresURL:
		return m.handleInputMode(keyMsg)
	case InputModeConfirmation:
		return m.handleConfirmationMode(keyMsg)
	}

	return m, nil
}

// handleNormalMode handles key events in normal mode (list navigation).
func (m Model) handleNormalMode(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case "down", "j":
		if m.selectedIndex < len(m.options)-1 {
			m.selectedIndex++
		}
	case "enter":
		m.handleClientSelection()
	case "esc", "q":
		m.router.Back()
	}
	return m, nil
}

// handleInputMode handles key events when inputting path/URL.
func (m Model) handleInputMode(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.String() {
	case "enter":
		m = m.handleInputSubmit()
	case "esc":
		m = m.cancelInput()
	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(keyMsg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleInputSubmit() Model {
	value := m.textInput.Value()
	if value == "" {
		m.errorMessage = "Path/URL cannot be empty"
		return m
	}

	switch m.inputMode {
	case InputModeSqlitePath:
		return m.saveSQLiteConfiguration(value)
	case InputModePostgresURL:
		return m.savePostgresConfiguration(value)
	}
	return m
}

func (m Model) saveSQLiteConfiguration(value string) Model {
	if err := m.saveStorageClient(types.StorageClientSQLite, value); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to save: %v", err)
		return m
	}
	m.sqlitePath = value
	m.activeClient = types.StorageClientSQLite
	m.inputMode = InputModeNone
	m.textInput.SetValue("")
	m.errorMessage = ""
	return m
}

func (m Model) savePostgresConfiguration(value string) Model {
	if err := m.saveStorageClient(types.StorageClientPostgres, value); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to save: %v", err)
		return m
	}
	m.postgresURL = value
	m.activeClient = types.StorageClientPostgres
	m.inputMode = InputModeNone
	m.textInput.SetValue("")
	m.errorMessage = ""
	return m
}

func (m Model) cancelInput() Model {
	m.inputMode = InputModeNone
	m.textInput.SetValue("")
	m.errorMessage = ""
	return m
}

// handleConfirmationMode handles key events in confirmation mode.
func (m Model) handleConfirmationMode(keyMsg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch keyMsg.String() {
	case "up", "k":
		if m.confirmIndex > 0 {
			m.confirmIndex--
		}
	case "down", "j":
		if m.confirmIndex < len(m.confirmOptions)-1 {
			m.confirmIndex++
		}
	case "enter":
		m = m.handleConfirmationSelection()
	case "esc":
		m = m.cancelConfirmation()
	}
	return m, nil
}

func (m Model) handleConfirmationSelection() Model {
	selectedOption := m.options[m.selectedIndex]
	switch m.confirmIndex {
	case 0: // Use existing
		m = m.useExistingConfiguration(types.StorageClient(selectedOption.Value))
	case 1: // Change configuration
		m = m.changeConfiguration(types.StorageClient(selectedOption.Value))
	case 2: // Remove configuration
		m = m.removeConfiguration(types.StorageClient(selectedOption.Value))
	}
	return m
}

func (m Model) useExistingConfiguration(clientType types.StorageClient) Model {
	if err := m.switchActiveClient(clientType); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to switch: %v", err)
	} else {
		m.activeClient = clientType
	}
	m.inputMode = InputModeNone
	m.confirmIndex = 0
	return m
}

func (m Model) changeConfiguration(clientType types.StorageClient) Model {
	switch clientType {
	case types.StorageClientSQLite:
		m.inputMode = InputModeSqlitePath
		m.textInput.SetValue(m.sqlitePath)
		m.textInput.Placeholder = "Enter SQLite file path"
	case types.StorageClientPostgres:
		m.inputMode = InputModePostgresURL
		m.textInput.SetValue(m.postgresURL)
		m.textInput.Placeholder = "Enter PostgreSQL connection URL"
	}
	m.confirmIndex = 0
	return m
}

func (m Model) removeConfiguration(clientType types.StorageClient) Model {
	if err := m.removeStorageClient(clientType); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to remove: %v", err)
		m.inputMode = InputModeNone
		m.confirmIndex = 0
		return m
	}

	switch clientType {
	case types.StorageClientSQLite:
		m.sqlitePath = ""
	case types.StorageClientPostgres:
		m.postgresURL = ""
	}

	// If removing active client, clear it
	if m.activeClient == clientType {
		m.activeClient = ""
		if m.secureStorage != nil {
			if err := m.secureStorage.Delete(config.StorageKeyTypeKey); err != nil {
				m.errorMessage = fmt.Sprintf("Failed to clear active client: %v", err)
			}
		}
	}

	m.inputMode = InputModeNone
	m.confirmIndex = 0
	return m
}

func (m Model) cancelConfirmation() Model {
	m.inputMode = InputModeNone
	m.confirmIndex = 0
	return m
}

// handleClientSelection handles when user selects a storage client.
func (m *Model) handleClientSelection() {
	selectedOption := m.options[m.selectedIndex]

	// Check if configuration exists
	var hasConfig bool
	switch selectedOption.Value {
	case types.StorageClientSQLite:
		hasConfig = m.sqlitePath != ""
	case types.StorageClientPostgres:
		hasConfig = m.postgresURL != ""
	}

	if hasConfig {
		// Show confirmation dialog
		m.inputMode = InputModeConfirmation
		m.confirmIndex = 0
	} else {
		// Show input dialog
		switch selectedOption.Value {
		case types.StorageClientSQLite:
			m.inputMode = InputModeSqlitePath
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Enter SQLite file path (e.g., ~/.smart-contract-cli/data.db)"
		case types.StorageClientPostgres:
			m.inputMode = InputModePostgresURL
			m.textInput.SetValue("")
			m.textInput.Placeholder = "Enter PostgreSQL URL (e.g., postgres://user:pass@localhost:5432/db)"
		}
	}
}

// saveStorageClient saves the storage client configuration to secure storage.
func (m *Model) saveStorageClient(clientType types.StorageClient, value string) error {
	logger.Info("Saving storage client configuration: %v, %v", clientType, value)
	if m.secureStorage == nil {
		return fmt.Errorf("secure storage not initialized")
	}

	// Save the value
	var key string
	switch clientType {
	case types.StorageClientSQLite:
		key = config.SecureStorageKeySqlitePathKey
	case types.StorageClientPostgres:
		key = config.SecureStorageKeyPostgresURLKey
	default:
		return fmt.Errorf("invalid client type: %s", clientType)
	}

	if err := m.secureStorage.Set(key, value); err != nil {
		return fmt.Errorf("failed to save storage client configuration: %w", err)
	}

	// Save as active client
	if err := m.secureStorage.Set(config.SecureStorageClientTypeKey, string(clientType)); err != nil {
		return fmt.Errorf("failed to set active storage client: %w", err)
	}
	return nil
}

// switchActiveClient switches the active storage client.
func (m *Model) switchActiveClient(clientType types.StorageClient) error {
	if m.secureStorage == nil {
		return fmt.Errorf("secure storage not initialized")
	}
	if err := m.secureStorage.Set(config.SecureStorageClientTypeKey, string(clientType)); err != nil {
		return fmt.Errorf("failed to switch active storage client: %w", err)
	}
	return nil
}

// removeStorageClient removes a storage client configuration.
func (m *Model) removeStorageClient(clientType types.StorageClient) error {
	if m.secureStorage == nil {
		return fmt.Errorf("secure storage not initialized")
	}

	// Determine which key to delete
	var key string
	switch clientType {
	case types.StorageClientSQLite:
		key = config.SecureStorageKeySqlitePathKey
	case types.StorageClientPostgres:
		key = config.SecureStorageKeyPostgresURLKey
	default:
		return fmt.Errorf("invalid client type: %s", clientType)
	}

	if err := m.secureStorage.Delete(key); err != nil {
		return fmt.Errorf("failed to delete storage client configuration: %w", err)
	}
	return nil
}

// maskPostgresURL masks the password in a Postgres URL for display.
func maskPostgresURL(url string) string {
	// Find password in URL (between : and @)
	parts := strings.Split(url, "@")
	if len(parts) < 2 {
		return url
	}

	beforeAt := parts[0]
	afterAt := strings.Join(parts[1:], "@")

	// Find the last colon in the part before @
	colonIdx := strings.LastIndex(beforeAt, ":")
	if colonIdx == -1 {
		return url
	}

	// Replace password with ****
	masked := beforeAt[:colonIdx+1] + "****@" + afterAt
	return masked
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	switch m.inputMode {
	case InputModeNone:
		return "↑/k: up • ↓/j: down • enter: select • esc/q: back", view.HelpDisplayOptionAppend
	case InputModeSqlitePath, InputModePostgresURL:
		return "enter: save • esc: cancel", view.HelpDisplayOptionAppend
	case InputModeConfirmation:
		return "↑/k: up • ↓/j: down • enter: confirm • esc: cancel", view.HelpDisplayOptionAppend
	}
	return "", view.HelpDisplayOptionAppend
}

func (m Model) View() string {
	// Show error if any
	if m.errorMessage != "" {
		return component.VStackC(
			component.T("Storage Client Configuration").Bold(true).Primary(),
			component.SpacerV(1),
			component.T("✗ "+m.errorMessage).Error(),
			component.SpacerV(1),
			component.T("Press any key to continue...").Muted(),
		).Render()
	}

	// Input mode - show text input
	switch m.inputMode {
	case InputModeSqlitePath:
		return m.renderInputView("Configure SQLite", "Enter the path for your SQLite database file:")
	case InputModePostgresURL:
		return m.renderInputView("Configure Postgres", "Enter the PostgreSQL connection URL:")
	}

	// Confirmation mode - show options
	if m.inputMode == InputModeConfirmation {
		return m.renderConfirmationView()
	}

	// Normal mode - show list of storage clients
	return m.renderNormalView()
}

// renderInputView renders the input view for entering path/URL.
func (m Model) renderInputView(title string, prompt string) string {
	return component.VStackC(
		component.T(title).Bold(true).Primary(),
		component.SpacerV(1),
		component.T(prompt).Muted(),
		component.SpacerV(1),
		component.Raw(m.textInput.View()),
		component.SpacerV(1),
		component.T("Press Enter to save, Esc to cancel").Muted(),
	).Render()
}

// renderConfirmationView renders the confirmation dialog.
func (m Model) renderConfirmationView() string {
	selectedOption := m.options[m.selectedIndex]
	currentValue := ""
	switch selectedOption.Value {
	case types.StorageClientSQLite:
		currentValue = m.sqlitePath
	case types.StorageClientPostgres:
		currentValue = maskPostgresURL(m.postgresURL)
	}

	// Create list items
	items := make([]component.ListItem, len(m.confirmOptions))
	for i, opt := range m.confirmOptions {
		items[i] = component.Item(opt, fmt.Sprintf("%d", i), "")
	}

	selectedValue := fmt.Sprintf("%d", m.confirmIndex)

	return component.VStackC(
		component.T(selectedOption.Label+" Configuration").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Current: "+currentValue).Success(),
		component.SpacerV(1),
		component.T("What would you like to do?").Muted(),
		component.SpacerV(1),
		component.NewList(items).
			Selected(selectedValue).
			Spacing(0),
	).Render()
}

// renderNormalView renders the main storage client list view.
func (m Model) renderNormalView() string {
	// Build list items with descriptions
	items := make([]component.ListItem, len(m.options))
	for idx, opt := range m.options {
		desc := opt.Description

		// Add stored path/URL to description if available
		if opt.Value == types.StorageClientSQLite && m.sqlitePath != "" {
			desc = desc + "\nPath: " + m.sqlitePath
		} else if opt.Value == types.StorageClientPostgres && m.postgresURL != "" {
			desc = desc + "\nURL: " + maskPostgresURL(m.postgresURL)
		}

		items[idx] = component.Item(opt.Label, string(opt.Value), desc)
	}

	// Highlight the active client
	highlightedValues := []string{}
	if m.activeClient != "" {
		highlightedValues = append(highlightedValues, string(m.activeClient))
	}

	selectedValue := m.options[m.selectedIndex].Value

	return component.VStackC(
		component.T("Storage Client Configuration").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Select your preferred storage client:").Muted(),
		component.SpacerV(1),
		component.NewList(items).
			Selected(string(selectedValue)).
			Highlighted(highlightedValues...).
			ShowDescription(true).
			DescriptionSpacing(0).
			Spacing(1),
		component.SpacerV(1),
		component.T("Legend: ").Bold(true),
		component.T("> = Cursor position | ★ = Active/configured").Muted(),
	).Render()
}
