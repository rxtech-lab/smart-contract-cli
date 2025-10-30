package component

import (
	"github.com/charmbracelet/lipgloss"
)

// Box is a container component that wraps content with optional borders and styling.
type Box struct {
	child Component
	style lipgloss.Style
}

// NewBox creates a new box container with the given child component.
func NewBox(child Component) *Box {
	return &Box{
		child: child,
		style: lipgloss.NewStyle(),
	}
}

// BoxC is a convenience function for creating a box.
func BoxC(child Component) *Box {
	return NewBox(child)
}

// WithStyle applies a lipgloss style to the box.
func (b *Box) WithStyle(style lipgloss.Style) *Box {
	b.style = style
	return b
}

// Border sets a border on the box.
func (b *Box) Border(border lipgloss.Border, sides ...bool) *Box {
	b.style = b.style.Border(border, sides...)
	return b
}

// BorderStyle sets the border style.
func (b *Box) BorderStyle(border lipgloss.Border) *Box {
	b.style = b.style.BorderStyle(border)
	return b
}

// RoundedBorder applies a rounded border.
func (b *Box) RoundedBorder() *Box {
	b.style = b.style.BorderStyle(lipgloss.RoundedBorder())
	return b
}

// NormalBorder applies a normal border.
func (b *Box) NormalBorder() *Box {
	b.style = b.style.BorderStyle(lipgloss.NormalBorder())
	return b
}

// ThickBorder applies a thick border.
func (b *Box) ThickBorder() *Box {
	b.style = b.style.BorderStyle(lipgloss.ThickBorder())
	return b
}

// DoubleBorder applies a double border.
func (b *Box) DoubleBorder() *Box {
	b.style = b.style.BorderStyle(lipgloss.DoubleBorder())
	return b
}

// HiddenBorder applies a hidden border (for spacing without visible border).
func (b *Box) HiddenBorder() *Box {
	b.style = b.style.BorderStyle(lipgloss.HiddenBorder())
	return b
}

// BorderColor sets the border color.
func (b *Box) BorderColor(color lipgloss.Color) *Box {
	b.style = b.style.BorderForeground(color)
	return b
}

// Padding sets padding inside the box.
func (b *Box) Padding(vertical, horizontal int) *Box {
	b.style = b.style.Padding(vertical, horizontal)
	return b
}

// PaddingAll sets equal padding on all sides.
func (b *Box) PaddingAll(padding int) *Box {
	b.style = b.style.Padding(padding)
	return b
}

// Margin sets margin outside the box.
func (b *Box) Margin(vertical, horizontal int) *Box {
	b.style = b.style.Margin(vertical, horizontal)
	return b
}

// MarginAll sets equal margin on all sides.
func (b *Box) MarginAll(margin int) *Box {
	b.style = b.style.Margin(margin)
	return b
}

// Width sets the width of the box.
func (b *Box) Width(width int) *Box {
	b.style = b.style.Width(width)
	return b
}

// Height sets the height of the box.
func (b *Box) Height(height int) *Box {
	b.style = b.style.Height(height)
	return b
}

// Background sets the background color.
func (b *Box) Background(color lipgloss.Color) *Box {
	b.style = b.style.Background(color)
	return b
}

// Foreground sets the foreground color.
func (b *Box) Foreground(color lipgloss.Color) *Box {
	b.style = b.style.Foreground(color)
	return b
}

// Align sets the alignment of content within the box.
func (b *Box) Align(horizontal, vertical lipgloss.Position) *Box {
	b.style = b.style.AlignHorizontal(horizontal).AlignVertical(vertical)
	return b
}

// Render renders the box with its child.
func (b *Box) Render() string {
	content := b.child.Render()
	return b.style.Render(content)
}

// Card creates a styled card container (box with rounded border and padding).
func Card(child Component) *Box {
	return NewBox(child).
		RoundedBorder().
		PaddingAll(1)
}

// Panel creates a styled panel container (box with normal border and padding).
func Panel(child Component) *Box {
	return NewBox(child).
		NormalBorder().
		PaddingAll(1)
}

// Center centers a component both horizontally and vertically.
type Center struct {
	child  Component
	width  int
	height int
	style  lipgloss.Style
}

// NewCenter creates a new center container.
func NewCenter(child Component, width, height int) *Center {
	return &Center{
		child:  child,
		width:  width,
		height: height,
		style:  lipgloss.NewStyle(),
	}
}

// CenterC is a convenience function for creating a centered component.
func CenterC(child Component, width, height int) Component {
	return NewCenter(child, width, height)
}

// WithStyle applies a lipgloss style to the center container.
func (c *Center) WithStyle(style lipgloss.Style) *Center {
	c.style = style
	return c
}

// Render renders the centered component.
func (c *Center) Render() string {
	content := c.child.Render()
	centered := lipgloss.Place(
		c.width,
		c.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
	return c.style.Render(centered)
}

// Padding is a component that adds padding around its child.
type Padding struct {
	child      Component
	top        int
	right      int
	bottom     int
	left       int
	style      lipgloss.Style
	useMargins bool // if true, uses margins instead of padding
}

// NewPadding creates a new padding container.
func NewPadding(child Component) *Padding {
	return &Padding{
		child:      child,
		top:        0,
		right:      0,
		bottom:     0,
		left:       0,
		style:      lipgloss.NewStyle(),
		useMargins: false,
	}
}

// PaddingC is a convenience function for creating a padding container.
func PaddingC(child Component) *Padding {
	return NewPadding(child)
}

// All sets equal padding on all sides.
func (p *Padding) All(padding int) *Padding {
	p.top = padding
	p.right = padding
	p.bottom = padding
	p.left = padding
	return p
}

// Vertical sets vertical (top and bottom) padding.
func (p *Padding) Vertical(padding int) *Padding {
	p.top = padding
	p.bottom = padding
	return p
}

// Horizontal sets horizontal (left and right) padding.
func (p *Padding) Horizontal(padding int) *Padding {
	p.left = padding
	p.right = padding
	return p
}

// Top sets top padding.
func (p *Padding) Top(padding int) *Padding {
	p.top = padding
	return p
}

// Right sets right padding.
func (p *Padding) Right(padding int) *Padding {
	p.right = padding
	return p
}

// Bottom sets bottom padding.
func (p *Padding) Bottom(padding int) *Padding {
	p.bottom = padding
	return p
}

// Left sets left padding.
func (p *Padding) Left(padding int) *Padding {
	p.left = padding
	return p
}

// UseMargins makes the padding use margins instead of padding.
// This affects spacing outside the component rather than inside.
func (p *Padding) UseMargins() *Padding {
	p.useMargins = true
	return p
}

// WithStyle applies a lipgloss style to the padding container.
func (p *Padding) WithStyle(style lipgloss.Style) *Padding {
	p.style = style
	return p
}

// Render renders the padded component.
func (p *Padding) Render() string {
	content := p.child.Render()

	var styled string
	if p.useMargins {
		styled = p.style.
			MarginTop(p.top).
			MarginRight(p.right).
			MarginBottom(p.bottom).
			MarginLeft(p.left).
			Render(content)
	} else {
		styled = p.style.
			PaddingTop(p.top).
			PaddingRight(p.right).
			PaddingBottom(p.bottom).
			PaddingLeft(p.left).
			Render(content)
	}

	return styled
}
