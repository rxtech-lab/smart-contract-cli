package view

import (
	"fmt"
	"log"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// routeEntry represents a route in the navigation stack.
type routeEntry struct {
	route       Route
	queryParams map[string]string
	pathParams  map[string]string
	fullPath    string
}

type RouterImplementation struct {
	routes           []Route
	currentRoute     *routeEntry
	navigationStack  []routeEntry
	currentComponent View
}

func NewRouter() Router {
	return &RouterImplementation{
		routes:          []Route{},
		navigationStack: []routeEntry{},
	}
}

// Init implements Router.
func (r *RouterImplementation) Init() tea.Cmd {
	if r.currentRoute != nil && r.currentRoute.route.Component != nil {
		r.currentComponent = r.currentRoute.route.Component(r)
		return r.currentComponent.Init()
	}
	return nil
}

// Update implements Router.
func (r *RouterImplementation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// if ctrl + c, exit the program
	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
		return r, tea.Quit
	}

	if r.currentComponent != nil {
		updatedModel, cmd := r.currentComponent.Update(msg)
		r.currentComponent = updatedModel.(View)
		return r, cmd
	}
	return r, nil
}

// View implements Router.
func (r *RouterImplementation) View() string {
	if r.currentComponent != nil {
		// Create a blue box style with full width
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")). // Blue color
			Padding(1, 2).
			Width(lipgloss.Width(r.currentComponent.View()) + 4). // Add padding to width
			MaxWidth(120) // Set a reasonable max width

		// Get the component view
		componentView := boxStyle.Render(r.currentComponent.View())

		// Create helper text style
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray color
			Italic(true).
			MarginTop(1)

		// Combine component help and exit instruction
		componentHelp := r.currentComponent.Help()
		helpText := "Ctrl + c to exit"
		if componentHelp != "" {
			helpText = componentHelp + " • " + helpText
		}

		return componentView + "\n" + helpStyle.Render(helpText)
	}
	return "No route selected"
}

// AddRoute implements Router.
func (r *RouterImplementation) AddRoute(route Route) {
	r.routes = append(r.routes, route)
}

// SetRoutes implements Router.
func (r *RouterImplementation) SetRoutes(routes []Route) {
	r.routes = routes
	// set the current route to the /
	if err := r.NavigateTo("/", nil); err != nil {
		log.Fatalf("failed to set current route to /: %v", err)
	}
}

// RemoveRoute implements Router.
func (r *RouterImplementation) RemoveRoute(path string) {
	for i, route := range r.routes {
		if route.Path == path {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			return
		}
	}
}

// GetCurrentRoute implements Router.
func (r *RouterImplementation) GetCurrentRoute() Route {
	if r.currentRoute != nil {
		return r.currentRoute.route
	}
	return Route{}
}

// GetRoutes implements Router.
func (r *RouterImplementation) GetRoutes() []Route {
	return r.routes
}

// NavigateTo implements Router.
func (r *RouterImplementation) NavigateTo(path string, queryParams map[string]string) error {
	route, pathParams, err := r.matchRoute(path)
	if err != nil {
		return err
	}

	if queryParams == nil {
		queryParams = make(map[string]string)
	}

	entry := routeEntry{
		route:       *route,
		queryParams: queryParams,
		pathParams:  pathParams,
		fullPath:    path,
	}

	// Push current route to stack if it exists
	if r.currentRoute != nil {
		r.navigationStack = append(r.navigationStack, *r.currentRoute)
	}

	r.currentRoute = &entry
	r.currentComponent = route.Component(r)
	return nil
}

// ReplaceRoute implements Router.
func (r *RouterImplementation) ReplaceRoute(path string) error {
	route, pathParams, err := r.matchRoute(path)
	if err != nil {
		return err
	}

	entry := routeEntry{
		route:       *route,
		queryParams: make(map[string]string),
		pathParams:  pathParams,
		fullPath:    path,
	}

	// Replace current route without modifying the stack
	r.currentRoute = &entry
	r.currentComponent = route.Component(r)
	return nil
}

// Back implements Router.
func (r *RouterImplementation) Back() {
	if len(r.navigationStack) == 0 {
		return
	}

	// Pop the last entry from the stack
	lastIndex := len(r.navigationStack) - 1
	r.currentRoute = &r.navigationStack[lastIndex]
	r.navigationStack = r.navigationStack[:lastIndex]
	r.currentComponent = r.currentRoute.route.Component(r)
}

// CanGoBack implements Router.
func (r *RouterImplementation) CanGoBack() bool {
	return len(r.navigationStack) > 0
}

// GetQueryParam implements Router.
func (r *RouterImplementation) GetQueryParam(key string) string {
	if r.currentRoute == nil {
		return ""
	}
	return r.currentRoute.queryParams[key]
}

// GetParam implements Router.
func (r *RouterImplementation) GetParam(key string) string {
	if r.currentRoute == nil {
		return ""
	}
	return r.currentRoute.pathParams[key]
}

// GetPath implements Router.
func (r *RouterImplementation) GetPath() string {
	if r.currentRoute == nil {
		return ""
	}
	return r.currentRoute.fullPath
}

// Refresh implements Router.
func (r *RouterImplementation) Refresh() {
	if r.currentRoute != nil && r.currentRoute.route.Component != nil {
		r.currentComponent = r.currentRoute.route.Component(r)
		r.currentComponent.Init()
	}
}

// matchRoute finds a matching route and extracts path parameters.
func (r *RouterImplementation) matchRoute(path string) (*Route, map[string]string, error) {
	for _, route := range r.routes {
		if pathParams, matched := r.matchPattern(route.Path, path); matched {
			return &route, pathParams, nil
		}
	}
	return nil, nil, fmt.Errorf("no route found for path: %s", path)
}

// matchPattern matches a route pattern against a path and extracts parameters.
// Supports patterns like "/users/:id" or "/posts/:postId/comments/:commentId".
func (r *RouterImplementation) matchPattern(pattern, path string) (map[string]string, bool) {
	// Exact match
	if pattern == path {
		return make(map[string]string), true
	}

	// Build regex pattern from route pattern
	paramNames := []string{}
	regexPattern := "^" + pattern + "$"

	// Find all :param patterns
	paramRegex := regexp.MustCompile(`:(\w+)`)
	matches := paramRegex.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		paramNames = append(paramNames, match[1])
	}

	// Replace :param with capture groups
	regexPattern = paramRegex.ReplaceAllString(regexPattern, `([^/]+)`)

	// Compile and match
	re := regexp.MustCompile(regexPattern)
	pathMatches := re.FindStringSubmatch(path)

	if pathMatches == nil {
		return nil, false
	}

	// Extract parameters
	params := make(map[string]string)
	for i, name := range paramNames {
		if i+1 < len(pathMatches) {
			params[name] = pathMatches[i+1]
		}
	}

	return params, true
}
