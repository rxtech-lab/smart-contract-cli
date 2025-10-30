package view

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HomeState represents the state of the home view.
type HomeState struct {
	selectedOption int
	options        []string
	hoveredOption  int
}

// HomeModel is the Bubble Tea model for the home view.
type HomeModel struct {
	state HomeState
}

// NewHomeModel creates a new home view model with two options.
func NewHomeModel() View {
	return HomeModel{
		state: HomeState{
			selectedOption: 0,
			hoveredOption:  0,
			options: []string{
				"Option 1: Manage Contracts",
				"Option 2: Manage Wallet",
			},
		},
	}
}

// Init initializes the home view.
func (m HomeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.state.hoveredOption > 0 {
				m.state.hoveredOption--
			}

		case "down", "j":
			if m.state.hoveredOption < len(m.state.options)-1 {
				m.state.hoveredOption++
			}

		case "enter":
			m.state.selectedOption = m.state.hoveredOption
			// You can add logic here to transition to other views
			return m, nil
		}
	}

	return m, nil
}

// View renders the home view.
func (m HomeModel) View() string {
	// Define the border style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(50)

	// Build the options content
	var optionsBuilder strings.Builder
	optionsBuilder.WriteString("Please select an option:\n\n")

	for i, option := range m.state.options {
		cursor := " "
		if m.state.hoveredOption == i {
			cursor = ">"
		}

		checked := " "
		if m.state.selectedOption == i {
			checked = "✓"
		}

		optionsBuilder.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, option))
	}

	// Render the boxed content
	boxedContent := boxStyle.Render(optionsBuilder.String())

	// Build the final view
	result := "Welcome to Smart Contract CLI\n\n"
	result += boxedContent
	result += "\n\nUse arrow keys (↑/↓) or j/k to move, Enter to select, q to quit\n"

	return result
}

// GetSelectedOption returns the currently selected option index.
func (m HomeModel) GetSelectedOption() int {
	return m.state.selectedOption
}

// GetState returns the current state.
func (m HomeModel) GetState() HomeState {
	return m.state
}
