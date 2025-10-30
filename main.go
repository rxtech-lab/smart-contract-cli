package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

func main() {
	program := tea.NewProgram(view.NewHomeModel())
	_, err := program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
