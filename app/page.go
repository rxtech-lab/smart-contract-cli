package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rxtech-lab/smart-contract-cli/internal/config"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

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
	if password, err := sharedMemory.Get("secure_storage_password"); err == nil && password != nil {
		if pwd, ok := password.(string); ok && pwd != "" {
			model.isUnlocked = true
		}
	}

	// Initialize secure storage if not unlocked
	if !model.isUnlocked {
		var err error
		model.secureStorage, err = storage.NewSecureStorageWithEncryption("smart-contract-cli-key", "")
		if err == nil {
			// Check if storage exists
			if !model.secureStorage.Exists() {
				model.isCreatingNew = true
			}
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

	if err := m.ensureSecureStorageInitialized(); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to initialize storage: %v", err)
		return m, nil
	}

	if err := m.createStorageIfNeeded(password); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to create storage: %v", err)
		return m, nil
	}

	if err := m.unlockAndStorePassword(password); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to unlock: %v", err)
		return m, nil
	}

	m.isUnlocked = true
	m.errorMessage = ""
	return m, nil
}

// ensureSecureStorageInitialized initializes secure storage if not already done.
func (m *Model) ensureSecureStorageInitialized() error {
	if m.secureStorage != nil {
		return nil
	}

	var err error
	m.secureStorage, err = storage.NewSecureStorageWithEncryption("smart-contract-cli-key", "")
	if err != nil {
		return fmt.Errorf("failed to create secure storage: %w", err)
	}
	return nil
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
	if err := m.secureStorage.Unlock(password); err != nil {
		return fmt.Errorf("failed to unlock storage: %w", err)
	}
	if err := m.sharedMemory.Set(config.SecureStoragePasswordKey, password); err != nil {
		return fmt.Errorf("failed to store password in shared memory: %w", err)
	}

	// Store secure storage in shared memory
	if err := m.sharedMemory.Set("secure_storage", m.secureStorage); err != nil {
		return fmt.Errorf("failed to store secure storage in shared memory: %w", err)
	}

	// Initialize and store SQL storage client
	if err := m.initializeStorageClient(); err != nil {
		return fmt.Errorf("failed to initialize storage client: %w", err)
	}

	return nil
}

// initializeStorageClient creates and stores the SQL storage client.
func (m Model) initializeStorageClient() error {
	sqlStorage, err := storage.NewSQLiteStorage("")
	if err != nil {
		return fmt.Errorf("failed to create SQLite storage: %w", err)
	}

	if err := m.sharedMemory.Set(config.StorageClientKey, sqlStorage); err != nil {
		return fmt.Errorf("failed to store storage client: %w", err)
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
