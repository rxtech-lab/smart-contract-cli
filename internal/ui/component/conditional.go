package component

// If represents a conditional component that renders one of two components based on a condition. //nolint:godot
// Similar to ternary operator: condition ? trueComponent : falseComponent.
type If struct {
	condition      bool
	trueComponent  Component
	falseComponent Component
}

// NewIf creates a new conditional component.
func NewIf(condition bool, trueComponent, falseComponent Component) *If {
	return &If{
		condition:      condition,
		trueComponent:  trueComponent,
		falseComponent: falseComponent,
	}
}

// IfC is a convenience function for conditional rendering.
func IfC(condition bool, trueComponent, falseComponent Component) Component {
	return NewIf(condition, trueComponent, falseComponent)
}

// IfElse is an alias for IfC for better readability.
func IfElse(condition bool, trueComponent, falseComponent Component) Component {
	return NewIf(condition, trueComponent, falseComponent)
}

// Render renders the appropriate component based on the condition.
func (i *If) Render() string {
	if i.condition {
		return i.trueComponent.Render()
	}
	return i.falseComponent.Render()
}

// IfThen renders a component only if the condition is true.
// If false, renders an empty component.
type IfThen struct {
	condition bool
	component Component
}

// NewIfThen creates a new conditional component that only renders when true.
func NewIfThen(condition bool, component Component) *IfThen {
	return &IfThen{
		condition: condition,
		component: component,
	}
}

// IfThenC is a convenience function for conditional rendering without an else branch.
func IfThenC(condition bool, component Component) Component {
	return NewIfThen(condition, component)
}

// When is an alias for IfThenC for better readability.
func When(condition bool, component Component) Component {
	return NewIfThen(condition, component)
}

// Render renders the component if the condition is true, otherwise renders nothing.
func (i *IfThen) Render() string {
	if i.condition {
		return i.component.Render()
	}
	return ""
}

// Unless renders a component only if the condition is false.
// Opposite of IfThen.
type Unless struct {
	condition bool
	component Component
}

// NewUnless creates a new conditional component that only renders when false.
func NewUnless(condition bool, component Component) *Unless {
	return &Unless{
		condition: condition,
		component: component,
	}
}

// UnlessC is a convenience function for inverse conditional rendering.
func UnlessC(condition bool, component Component) Component {
	return NewUnless(condition, component)
}

// Render renders the component if the condition is false, otherwise renders nothing.
func (u *Unless) Render() string {
	if !u.condition {
		return u.component.Render()
	}
	return ""
}

// SwitchCase represents a case in a switch statement.
type SwitchCase struct {
	matches   func() bool
	component Component
}

// Case creates a new switch case.
func Case(matches func() bool, component Component) SwitchCase {
	return SwitchCase{
		matches:   matches,
		component: component,
	}
}

// Switch renders the first matching case.
// Similar to a switch statement or pattern matching.
type Switch struct {
	cases          []SwitchCase
	defaultCase    Component
	evaluateAll    bool // if true, evaluates all cases (useful for side effects)
	renderMultiple bool // if true, renders all matching cases, not just the first
}

// NewSwitch creates a new switch component.
func NewSwitch(cases ...SwitchCase) *Switch {
	return &Switch{
		cases:          cases,
		defaultCase:    Empty(),
		evaluateAll:    false,
		renderMultiple: false,
	}
}

// SwitchC is a convenience function for creating a switch component.
func SwitchC(cases ...SwitchCase) *Switch {
	return NewSwitch(cases...)
}

// Default sets the default component to render if no cases match.
func (s *Switch) Default(component Component) *Switch {
	s.defaultCase = component
	return s
}

// EvaluateAll makes the switch evaluate all cases even after finding a match.
// Useful if cases have side effects.
func (s *Switch) EvaluateAll() *Switch {
	s.evaluateAll = true
	return s
}

// RenderMultiple makes the switch render all matching cases, not just the first.
func (s *Switch) RenderMultiple() *Switch {
	s.renderMultiple = true
	return s
}

// Render renders the first (or all) matching case(s).
func (s *Switch) Render() string {
	var matched []Component

	for _, c := range s.cases {
		if c.matches() {
			matched = append(matched, c.component)
			if !s.evaluateAll && !s.renderMultiple {
				// Found first match, stop evaluating
				break
			}
			if s.renderMultiple && !s.evaluateAll {
				// Continue evaluating to find all matches, but don't evaluate unnecessary cases
				continue
			}
		}
	}

	if len(matched) == 0 {
		return s.defaultCase.Render()
	}

	if len(matched) == 1 {
		return matched[0].Render()
	}

	// Multiple matches - render them all
	return VStackC(matched...).Render()
}

// Match is a helper for creating match conditions based on equality.
func Match[T comparable](value, target T) func() bool {
	return func() bool {
		return value == target
	}
}

// MatchAny is a helper for creating match conditions based on multiple possible values.
func MatchAny[T comparable](value T, targets ...T) func() bool {
	return func() bool {
		for _, target := range targets {
			if value == target {
				return true
			}
		}
		return false
	}
}

// MatchRange is a helper for creating match conditions based on a range.
func MatchRange[T int | int64 | float64](value, minVal, maxVal T) func() bool {
	return func() bool {
		return value >= minVal && value <= maxVal
	}
}

// Show renders a component if the condition is true, otherwise renders nothing.
// This is a simple helper that's more concise than IfThenC for simple cases.
func Show(condition bool, component Component) Component {
	if condition {
		return component
	}
	return Empty()
}

// Hide renders a component if the condition is false, otherwise renders nothing.
// Opposite of Show.
func Hide(condition bool, component Component) Component {
	if !condition {
		return component
	}
	return Empty()
}

// Toggle switches between two components based on a boolean state.
// Similar to If, but with a more semantic name for toggle-like UIs.
func Toggle(state bool, onComponent, offComponent Component) Component {
	return IfC(state, onComponent, offComponent)
}
