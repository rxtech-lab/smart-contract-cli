package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/evm/storage/sql"
	"github.com/rxtech-lab/smart-contract-cli/internal/contract/types"
	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
	"github.com/rxtech-lab/smart-contract-cli/internal/log"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

var logger, _ = log.NewFileLogger("./logs/home/page.log")

type Option struct {
	Label string
	Route string
}

var options = []Option{
	{Label: "EVM", Route: "/evm"},
}

type Model struct {
	router       view.Router
	sharedMemory storage.SharedMemory

	isUnlocked     bool
	selectedOption Option
	selectedIndex  int

	// Password prompt state
	passwordInput *component.TextInput
	errorMessage  string
	secureStorage storage.SecureStorage
	isCreatingNew bool
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	model := Model{
		router:         router,
		sharedMemory:   sharedMemory,
		selectedOption: options[0],
		selectedIndex:  0,
		passwordInput: component.TextInputC().
			Prompt("Password: ").
			Placeholder("Enter password").
			EchoMode(textinput.EchoPassword).
			Focused(true),
	}

	// Check if already unlocked
	if password, err := sharedMemory.Get(config.SecureStoragePasswordKey); err == nil && password != nil {
		if pwd, ok := password.(string); ok && pwd != "" {
			model.isUnlocked = true
			// Initialize secure storage with the password from shared memory
			model.secureStorage, _ = storage.NewSecureStorageWithEncryption(pwd, "")
		}
	}

	// If not unlocked, check if storage file exists to show appropriate message
	if !model.isUnlocked {
		// Create a temporary secure storage just to check if file exists
		// We can't actually use it without the password
		tempStorage, err := storage.NewSecureStorageWithEncryption("temp", "")
		if err == nil && !tempStorage.Exists() {
			model.isCreatingNew = true
		}
	}

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	return "Use arrow keys to navigate and enter to select", view.HelpDisplayOptionAppend
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If not unlocked, handle password input
	if !m.isUnlocked {
		return m.handlePasswordInput(msg)
	}

	// Normal navigation when unlocked
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		m.selectedIndex = m.moveUp(m.selectedIndex)
		m.selectedOption = options[m.selectedIndex]
	case "down", "j":
		m.selectedIndex = m.moveDown(m.selectedIndex)
		m.selectedOption = options[m.selectedIndex]
	case "enter", " ":
		err := m.router.NavigateTo(m.selectedOption.Route, nil)
		if err != nil {
			return m, tea.Quit
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// handlePasswordInput handles password input and unlock logic.
func (m Model) handlePasswordInput(msg tea.Msg) (Model, tea.Cmd) {
	if m.passwordInput == nil {
		return m, nil
	}

	inputModel := m.passwordInput.GetModel()
	var cmd tea.Cmd
	inputModel, cmd = inputModel.Update(msg)

	// Update the password input with new value
	m.passwordInput = component.TextInputC().
		Prompt("Password: ").
		Placeholder("Enter password").
		EchoMode(textinput.EchoPassword).
		Focused(true).
		Value(inputModel.Value())

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, cmd
	}

	switch keyMsg.String() {
	case "enter":
		return m.handlePasswordSubmit(inputModel.Value())
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, cmd
}

// handlePasswordSubmit processes the password submission and unlocks storage.
func (m Model) handlePasswordSubmit(password string) (Model, tea.Cmd) {
	if password == "" {
		m.errorMessage = "Password cannot be empty"
		return m, nil
	}

	// Initialize secure storage with the user's password as the encryption key
	if err := m.ensureSecureStorageInitialized(password); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to initialize storage: %v", err)
		return m, nil
	}

	logger.Info("Creating storage if needed")
	if err := m.createStorageIfNeeded(password); err != nil {
		logger.Error("Failed to create storage: %v", err)
		m.errorMessage = fmt.Sprintf("Failed to create storage: %v", err)
		return m, nil
	}

	if err := m.unlockAndStorePassword(password); err != nil {
		logger.Error("Failed to unlock: %v", err)
		m.errorMessage = fmt.Sprintf("Failed to unlock: %v", err)
		return m, nil
	}

	m.isUnlocked = true
	m.errorMessage = ""
	return m, nil
}

// ensureSecureStorageInitialized initializes secure storage with the provided password.
// The password is used as the encryption key for all data.
func (m *Model) ensureSecureStorageInitialized(password string) error {
	// Always reinitialize with the correct password as encryption key
	var err error
	m.secureStorage, err = storage.NewSecureStorageWithEncryption(password, "")
	if err != nil {
		return fmt.Errorf("failed to create secure storage: %w", err)
	}
	return nil
}

func (m Model) tryLoadStorageClient() (sql.Storage, error) {
	// load storage client from secure storage
	logger.Info("Loading storage client from secure storage")
	storageClientTypeString, err := m.secureStorage.Get(config.SecureStorageClientTypeKey)
	if err != nil {
		logger.Error("Failed to get storage client type from secure storage: %v", err)
		return nil, errors.NewStorageClientNotInitializedError("storage client not initialized")
	}
	logger.Info("Got Storage client type: %v", storageClientTypeString)

	storageClientType := types.StorageClient(storageClientTypeString)
	switch storageClientType {
	case types.StorageClientSQLite:
		logger.Info("Loading SQLite storage client from secure storage")
		sqlitePath, err := m.secureStorage.Get(config.SecureStorageKeySqlitePathKey)
		if err != nil {
			return nil, errors.NewStorageClientNotInitializedError("sqlite path not initialized")
		}
		client, err := sql.GetStorage(types.StorageClientSQLite, sqlitePath)
		if err != nil {
			logger.Error("Failed to get SQLite storage: %v", err)
			return nil, fmt.Errorf("failed to get SQLite storage: %w", err)
		}
		return client, nil
	case types.StorageClientPostgres:
		logger.Info("Loading Postgres storage client from secure storage")
		postgresURL, err := m.secureStorage.Get(config.SecureStorageKeyPostgresURLKey)
		if err != nil {
			return nil, errors.NewStorageClientNotInitializedError("postgres url not initialized")
		}
		client, err := sql.GetStorage(types.StorageClientPostgres, postgresURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get Postgres storage: %w", err)
		}
		return client, nil
	default:
		return nil, errors.NewStorageClientNotInitializedError(fmt.Sprintf("invalid storage client type: %s", storageClientType))
	}
}

// createStorageIfNeeded creates storage if it doesn't exist.
func (m Model) createStorageIfNeeded(password string) error {
	if m.secureStorage.Exists() {
		return nil
	}
	if err := m.secureStorage.Create(password); err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	return nil
}

// unlockAndStorePassword unlocks storage and stores password in shared memory.
func (m Model) unlockAndStorePassword(password string) error {
	if err := m.secureStorage.TestPassword(password); err != nil {
		return fmt.Errorf("failed to unlock storage: %w", err)
	}
	if err := m.sharedMemory.Set(config.SecureStoragePasswordKey, password); err != nil {
		return fmt.Errorf("failed to store password in shared memory: %w", err)
	}

	// initialize storage client in shared memory
	// skip adding storage client to shared memory if it returns not initialized error
	// show error message to user if it returns other error
	storageClient, err := m.tryLoadStorageClient()
	logger.Info("Loading storage client: %v", storageClient)
	if err != nil {
		logger.Error("Failed to load storage client: %v", err)
		if !errors.HasCode(err, errors.ErrCodeStorageClientNotInitialized) {
			return err
		}
	}
	if err := m.sharedMemory.Set(config.StorageClientKey, storageClient); err != nil {
		logger.Error("Failed to store storage client in shared memory: %v", err)
		return fmt.Errorf("failed to store storage client in shared memory: %w", err)
	}
	return nil
}

func (m Model) moveUp(currentIndex int) int {
	if currentIndex > 0 {
		return currentIndex - 1
	}
	return currentIndex
}

func (m Model) moveDown(currentIndex int) int {
	if currentIndex < len(options)-1 {
		return currentIndex + 1
	}
	return currentIndex
}

func (m Model) View() string {
	if !m.isUnlocked {
		return m.renderPasswordPrompt()
	}
	return m.renderMainMenu()
}

// renderPasswordPrompt renders the password input screen.
func (m Model) renderPasswordPrompt() string {
	promptText := m.getPasswordPromptText()

	components := []component.Component{
		component.T(promptText).Bold(true).Primary(),
		component.SpacerV(1),
	}

	if m.passwordInput != nil {
		components = append(components, m.passwordInput)
	}

	if m.errorMessage != "" {
		components = append(components,
			component.SpacerV(1),
			component.T(m.errorMessage).Foreground(lipgloss.Color("1")), // Red color
		)
	}

	components = append(components,
		component.SpacerV(1),
		component.T("Press Enter to submit, Ctrl+C to quit").Muted(),
	)

	return component.VStackC(components...).Render()
}

// getPasswordPromptText returns the appropriate prompt text based on state.
func (m Model) getPasswordPromptText() string {
	if m.isCreatingNew {
		return "Create a password for secure storage:"
	}
	return "Enter password to unlock secure storage:"
}

// renderMainMenu renders the main blockchain selection menu.
func (m Model) renderMainMenu() string {
	items := make([]component.ListItem, len(options))
	for i, opt := range options {
		items[i] = component.Item(opt.Label, opt.Route, "")
	}

	return component.VStackC(
		component.T("Select a blockchain:").Bold(true).Primary(),
		component.SpacerV(1),
		component.NewList(items).
			Selected(m.selectedOption.Route).
			SelectedPrefix("> ").
			UnselectedPrefix("  "),
		component.SpacerV(1),
		component.T("Selected: "+m.selectedOption.Label).Muted(),
	).Render()
}
