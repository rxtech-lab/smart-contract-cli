package view

import tea "github.com/charmbracelet/bubbletea"
import "github.com/rxtech-lab/smart-contract-cli/internal/storage"

type HelpDisplayOption string

const (
	HelpDisplayOptionOverride HelpDisplayOption = "override"
	HelpDisplayOptionAppend   HelpDisplayOption = "append"
)

type View interface {
	View() string
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	Help() (string, HelpDisplayOption)
}

type Router interface {
	// View returns the view of the router
	View() string
	// Init initializes the router
	Init() tea.Cmd
	// Update updates the router
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	// AddRoute adds a new route to the router
	AddRoute(route Route)
	// SetRoutes sets the routes for the router
	SetRoutes(routes []Route)
	// RemoveRoute removes a route from the router
	RemoveRoute(path string)
	// GetCurrentRoute gets the current route
	GetCurrentRoute() Route
	// GetRoutes gets all the routes
	GetRoutes() []Route
	// NavigateTo navigates to a new route. This will push the new route to the stack and set the current route to the new route
	NavigateTo(path string, queryParams map[string]string) error
	// ReplaceRoute replaces the current route with a new route. This will replace the current route on the stack and set the current route to the new route
	ReplaceRoute(path string) error
	// NavigateBack navigates back to the previous route
	Back()
	// CanGoBack checks if there is a previous route to navigate back to
	CanGoBack() bool
	// GetQueryParam gets a query parameter from the current route
	// For example, if the route is /users?id=123, GetQueryParam("id") will return the value of the id parameter
	GetQueryParam(key string) string
	// GetParam gets a parameter from the current route
	// For example, if the route is /users/:id, GetParam("id") will return the value of the id parameter
	GetParam(key string) string
	// GetPath gets the path of the current route
	GetPath() string
	// Refresh refreshes the current route
	Refresh()
}

type Route struct {
	Path      string
	Component func(router Router, sharedMemory storage.SharedMemory) View
}
