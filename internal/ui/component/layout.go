package component

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// VStack arranges components vertically (like SwiftUI's VStack).
type VStack struct {
	children  []Component
	spacing   int
	alignment lipgloss.Position
	style     lipgloss.Style
}

// NewVStack creates a new vertical stack with the given components.
func NewVStack(children ...Component) *VStack {
	return &VStack{
		children:  children,
		spacing:   0,
		alignment: lipgloss.Left,
		style:     lipgloss.NewStyle(),
	}
}

// VStack is a convenience function for creating a vertical stack.
func VStackC(children ...Component) Component {
	return NewVStack(children...)
}

// Spacing sets the vertical spacing between children.
func (v *VStack) Spacing(spacing int) *VStack {
	v.spacing = spacing
	return v
}

// Align sets the horizontal alignment of children.
func (v *VStack) Align(position lipgloss.Position) *VStack {
	v.alignment = position
	return v
}

// WithStyle applies a lipgloss style to the container.
func (v *VStack) WithStyle(style lipgloss.Style) *VStack {
	v.style = style
	return v
}

// Render renders the vertical stack.
func (v *VStack) Render() string {
	if len(v.children) == 0 {
		return v.style.Render("")
	}

	parts := make([]string, 0, len(v.children)*2)
	for i, child := range v.children {
		rendered := child.Render()
		parts = append(parts, rendered)

		// Add spacing between elements (but not after the last one)
		if i < len(v.children)-1 && v.spacing > 0 {
			for j := 0; j < v.spacing; j++ {
				parts = append(parts, "")
			}
		}
	}

	result := strings.Join(parts, "\n")
	return v.style.Render(result)
}

// HStack arranges components horizontally (like SwiftUI's HStack).
type HStack struct {
	children  []Component
	spacing   int
	alignment lipgloss.Position
	style     lipgloss.Style
}

// NewHStack creates a new horizontal stack with the given components.
func NewHStack(children ...Component) *HStack {
	return &HStack{
		children:  children,
		spacing:   0,
		alignment: lipgloss.Top,
		style:     lipgloss.NewStyle(),
	}
}

// HStack is a convenience function for creating a horizontal stack.
func HStackC(children ...Component) Component {
	return NewHStack(children...)
}

// Spacing sets the horizontal spacing between children.
func (h *HStack) Spacing(spacing int) *HStack {
	h.spacing = spacing
	return h
}

// Align sets the vertical alignment of children.
func (h *HStack) Align(position lipgloss.Position) *HStack {
	h.alignment = position
	return h
}

// WithStyle applies a lipgloss style to the container.
func (h *HStack) WithStyle(style lipgloss.Style) *HStack {
	h.style = style
	return h
}

// Render renders the horizontal stack.
func (h *HStack) Render() string {
	if len(h.children) == 0 {
		return h.style.Render("")
	}

	rendered := make([]string, len(h.children))
	for i, child := range h.children {
		rendered[i] = child.Render()
	}

	// Join with spacing if specified
	spacer := ""
	if h.spacing > 0 {
		spacer = strings.Repeat(" ", h.spacing)
	}

	result := lipgloss.JoinHorizontal(h.alignment, rendered...)
	if h.spacing > 0 && len(rendered) > 1 {
		// Re-join with spacing
		parts := make([]string, 0, len(rendered)*2-1)
		for i, r := range rendered {
			parts = append(parts, r)
			if i < len(rendered)-1 {
				parts = append(parts, spacer)
			}
		}
		result = lipgloss.JoinHorizontal(h.alignment, parts...)
	}

	return h.style.Render(result)
}

// ZStack overlays components on top of each other (like SwiftUI's ZStack).
type ZStack struct {
	children  []Component
	alignment lipgloss.Position
	style     lipgloss.Style
}

// NewZStack creates a new overlay stack with the given components.
// Components are layered from bottom to top (first component is at the back).
func NewZStack(children ...Component) *ZStack {
	return &ZStack{
		children:  children,
		alignment: lipgloss.Center,
		style:     lipgloss.NewStyle(),
	}
}

// ZStack is a convenience function for creating an overlay stack.
func ZStackC(children ...Component) Component {
	return NewZStack(children...)
}

// Align sets the alignment of overlaid children.
func (z *ZStack) Align(position lipgloss.Position) *ZStack {
	z.alignment = position
	return z
}

// WithStyle applies a lipgloss style to the container.
func (z *ZStack) WithStyle(style lipgloss.Style) *ZStack {
	z.style = style
	return z
}

// Render renders the overlay stack.
func (z *ZStack) Render() string {
	if len(z.children) == 0 {
		return z.style.Render("")
	}

	// For ZStack, we layer components using lipgloss.Place
	// Start with the first (bottom) component
	result := z.children[0].Render()

	// TODO: True overlay support would require more sophisticated rendering
	// For now, we'll just render them vertically as a fallback
	// In a real TUI, true overlay would need cursor positioning
	for index := 1; index < len(z.children); index++ {
		result = lipgloss.Place(
			lipgloss.Width(result),
			lipgloss.Height(result),
			z.alignment,
			z.alignment,
			z.children[index].Render(),
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}),
		)
	}

	return z.style.Render(result)
}

// Spacer creates flexible or fixed spacing.
type Spacer struct {
	height int // vertical spacing (number of empty lines)
	width  int // horizontal spacing (number of spaces)
}

// NewSpacer creates a new spacer with the given dimensions.
// If height > 0, creates vertical spacing (empty lines).
// If width > 0, creates horizontal spacing (spaces).
func NewSpacer(height, width int) *Spacer {
	return &Spacer{
		height: height,
		width:  width,
	}
}

// SpacerV creates a vertical spacer (empty lines).
func SpacerV(lines int) Component {
	return NewSpacer(lines, 0)
}

// SpacerH creates a horizontal spacer (spaces).
func SpacerH(spaces int) Component {
	return NewSpacer(0, spaces)
}

// Render renders the spacer.
func (s *Spacer) Render() string {
	if s.height > 0 {
		// Vertical spacer: return empty lines
		return strings.Repeat("\n", s.height-1)
	}
	if s.width > 0 {
		// Horizontal spacer: return spaces
		return strings.Repeat(" ", s.width)
	}
	return ""
}

// Divider creates a horizontal divider line.
type Divider struct {
	char  string
	width int
	style lipgloss.Style
}

// NewDivider creates a new divider with the given character and width.
func NewDivider(char string, width int) *Divider {
	return &Divider{
		char:  char,
		width: width,
		style: lipgloss.NewStyle(),
	}
}

// DividerLine creates a divider line with the default character "─".
func DividerLine(width int) Component {
	return NewDivider("─", width)
}

// WithStyle applies a lipgloss style to the divider.
func (d *Divider) WithStyle(style lipgloss.Style) *Divider {
	d.style = style
	return d
}

// Render renders the divider.
func (d *Divider) Render() string {
	if d.width <= 0 {
		return ""
	}
	line := strings.Repeat(d.char, d.width)
	return d.style.Render(line)
}
