package component

import "github.com/charmbracelet/lipgloss"

// Component is the base interface for all UI components.
// All components must implement Render() which returns the final string representation.
type Component interface {
	Render() string
}

// Styleable represents a component that can have lipgloss styles applied.
type Styleable interface {
	Component
	// WithStyle applies a lipgloss style to the component
	WithStyle(style lipgloss.Style) Component
	// GetStyle returns the current style of the component
	GetStyle() lipgloss.Style
}

// Interactive represents a component that can handle selection/focus states.
type Interactive interface {
	Component
	// WithSelected sets whether the component is selected
	WithSelected(selected bool) Component
	// IsSelected returns whether the component is selected
	IsSelected() bool
}

// Modifier is a function that modifies a component.
type Modifier func(Component) Component

// ComponentFunc is a function type that implements Component.
// This allows simple functions to be used as components.
type ComponentFunc func() string

func (f ComponentFunc) Render() string {
	return f()
}

// Empty returns an empty component.
func Empty() Component {
	return ComponentFunc(func() string {
		return ""
	})
}

// Raw creates a component from a raw string.
// Useful for incorporating existing string-based rendering.
func Raw(s string) Component {
	return ComponentFunc(func() string {
		return s
	})
}

// Join joins multiple components with a separator.
func Join(separator string, components ...Component) Component {
	return ComponentFunc(func() string {
		if len(components) == 0 {
			return ""
		}

		result := components[0].Render()
		for i := 1; i < len(components); i++ {
			result += separator + components[i].Render()
		}
		return result
	})
}
