package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

type Model struct {
	router view.Router
}

func NewPage(router view.Router) view.View {
	return Model{router: router}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return "Home Page"
}
