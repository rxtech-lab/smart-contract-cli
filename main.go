package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/app"
	"github.com/rxtech-lab/smart-contract-cli/internal/view"
)

func main() {
	router := view.NewRouter()
	router.SetRoutes(app.GetRoutes())
	program := tea.NewProgram(router)
	_, err := program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
