package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

type Option struct {
	Label string
	Value string
}

var options = []Option{
	{Label: "EVM", Value: "evm"},
	{Label: "Solana", Value: "solana"},
	{Label: "Bitcoin", Value: "bitcoin"},
}

type Model struct {
	router         view.Router
	selectedOption Option
}

func NewPage(router view.Router) view.View {
	return Model{
		router:         router,
		selectedOption: options[0], // Default to first option
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Help() string {
	return "Use arrow keys to navigate and enter to select"
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			// Find current index and move up
			for i, opt := range options {
				if opt.Value == m.selectedOption.Value {
					if i > 0 {
						m.selectedOption = options[i-1]
					}
					break
				}
			}
		case "down", "j":
			// Find current index and move down
			for i, opt := range options {
				if opt.Value == m.selectedOption.Value {
					if i < len(options)-1 {
						m.selectedOption = options[i+1]
					}
					break
				}
			}
		case "enter", " ":
			// Confirm selection - could navigate or perform action
			// Example: m.router.NavigateTo("/"+m.selectedOption.Value, nil)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString("Select a blockchain:\n\n")

	for _, option := range options {
		marker := "  "
		if m.selectedOption.Value == option.Value {
			marker = "> "
		}

		b.WriteString(fmt.Sprintf("%s%s\n", marker, option.Label))
	}

	b.WriteString(fmt.Sprintf("\nSelected: %s\n", m.selectedOption.Label))

	return b.String()
}
