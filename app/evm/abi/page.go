package abi

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/rxtech-lab/smart-contract-cli/internal/ui/component"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

type Model struct {
	router       view.Router
	sharedMemory storage.SharedMemory
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
	return Model{
		router:       router,
		sharedMemory: sharedMemory,
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
	case "esc", "q":
		m.router.Back()
	}

	return m, nil
}

func (m Model) Help() (string, view.HelpDisplayOption) {
	return "esc/q: back", view.HelpDisplayOptionAppend
}

func (m Model) View() string {
	return component.VStackC(
		component.T("ABI Management").Bold(true).Primary(),
		component.SpacerV(1),
		component.T("Coming soon...").Muted(),
	).Render()
}
