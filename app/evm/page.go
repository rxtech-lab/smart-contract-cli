package evm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

type Model struct {
	router view.Router

	selectedOption Option
}

type Option struct {
	Label       string
	Value       string
	Description string
	Route       string
}

var options = []Option{
	{Label: "Storage Client", Value: "storage-client", Route: "/evm/storage-client", Description: "Manage the storage of the contract"},
	{Label: "Abi Management", Value: "abi-management", Route: "/evm/abi-management", Description: "Manage the ABI of the contract"},
	{Label: "Contract Management", Value: "contract-management", Route: "/evm/contract-management", Description: "Manage the contract of the contract"},
	{Label: "Endpoint Management", Value: "endpoint-management", Route: "/evm/endpoint-management", Description: "Manage the endpoint of the contract"},
}

func NewPage(router view.Router) view.View {
	return Model{
		router:         router,
		selectedOption: options[0], // Initialize with first option
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		// Move selection up
		currentIndex := m.getCurrentIndex()
		if currentIndex > 0 {
			m.selectedOption = options[currentIndex-1]
		}
	case "down", "j":
		// Move selection down
		currentIndex := m.getCurrentIndex()
		if currentIndex < len(options)-1 {
			m.selectedOption = options[currentIndex+1]
		}
	case "enter":
		// Navigate to selected option's route
		if m.selectedOption.Route != "" {
			if err := m.router.NavigateTo(m.selectedOption.Route, nil); err != nil {
				// Navigation error - stay on current page
				return m, nil
			}
		}
	case "esc", "q":
		// Navigate back to home
		m.router.Back()
	}

	return m, nil
}

// getCurrentIndex finds the index of the currently selected option.
func (m Model) getCurrentIndex() int {
	for i, opt := range options {
		if opt.Value == m.selectedOption.Value {
			return i
		}
	}
	return 0
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	return "↑/k: up • ↓/j: down • enter: select • esc/q: back", view.HelpDisplayOptionAppend
}

func (m Model) View() string {
	items := make([]component.ListItem, len(options))
	for i, opt := range options {
		items[i] = component.Item(opt.Label, opt.Value, opt.Description)
	}
	return component.VStackC(
		component.T("You are using the EVM blockchain").Bold(true).Primary(),
		component.SpacerV(1),
		component.NewList(items).ShowDescription(true).Spacing(1).Selected(m.selectedOption.Value),

		component.SpacerV(1),
		component.T("EVM Blockchain").Muted(),
	).Render()
}
