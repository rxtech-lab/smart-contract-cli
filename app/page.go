package app

import (
	tea "github.com/charmbracelet/bubbletea"
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
	router         view.Router
	selectedOption Option
	selectedIndex  int
}

func NewPage(router view.Router) view.View {
	return Model{
		router:         router,
		selectedOption: options[0],
		selectedIndex:  0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	return "Use arrow keys to navigate and enter to select", view.HelpDisplayOptionAppend
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	// Convert options to ListItems
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
