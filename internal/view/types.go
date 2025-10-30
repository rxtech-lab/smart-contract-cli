package view

import tea "github.com/charmbracelet/bubbletea"

type RouterState struct {
	CurrentRoute string
}

type View interface {
	View() string
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
}

type Route struct {
	Path      string
	Component View
}
