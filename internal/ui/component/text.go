package component

import (
	"github.com/charmbracelet/lipgloss"
)

// Text is a styled text component with chainable modifiers.
type Text struct {
	content string
	style   lipgloss.Style
}

// NewText creates a new text component with the given content.
func NewText(content string) *Text {
	return &Text{
		content: content,
		style:   lipgloss.NewStyle(),
	}
}

// T is a convenience function for creating a text component.
func T(content string) *Text {
	return NewText(content)
}

// Render renders the text with applied styles.
func (t *Text) Render() string {
	return t.style.Render(t.content)
}

// WithStyle applies a lipgloss style to the text.
func (t *Text) WithStyle(style lipgloss.Style) *Text {
	t.style = style
	return t
}

// GetStyle returns the current style.
func (t *Text) GetStyle() lipgloss.Style {
	return t.style
}

// Style Modifiers - Chainable methods for common styling

// Bold makes the text bold.
func (t *Text) Bold(bold bool) *Text {
	t.style = t.style.Bold(bold)
	return t
}

// Italic makes the text italic.
func (t *Text) Italic(italic bool) *Text {
	t.style = t.style.Italic(italic)
	return t
}

// Underline adds underline to the text.
func (t *Text) Underline(underline bool) *Text {
	t.style = t.style.Underline(underline)
	return t
}

// Strikethrough adds strikethrough to the text.
func (t *Text) Strikethrough(strikethrough bool) *Text {
	t.style = t.style.Strikethrough(strikethrough)
	return t
}

// Blink makes the text blink.
func (t *Text) Blink(blink bool) *Text {
	t.style = t.style.Blink(blink)
	return t
}

// Faint makes the text faint/dim.
func (t *Text) Faint(faint bool) *Text {
	t.style = t.style.Faint(faint)
	return t
}

// Color Modifiers

// Foreground sets the foreground color.
func (t *Text) Foreground(color lipgloss.Color) *Text {
	t.style = t.style.Foreground(color)
	return t
}

// Background sets the background color.
func (t *Text) Background(color lipgloss.Color) *Text {
	t.style = t.style.Background(color)
	return t
}

// Color sets the foreground color (alias for Foreground).
func (t *Text) Color(color lipgloss.Color) *Text {
	return t.Foreground(color)
}

// BgColor sets the background color (alias for Background).
func (t *Text) BgColor(color lipgloss.Color) *Text {
	return t.Background(color)
}

// Layout Modifiers

// Width sets the width of the text.
func (t *Text) Width(width int) *Text {
	t.style = t.style.Width(width)
	return t
}

// Height sets the height of the text.
func (t *Text) Height(height int) *Text {
	t.style = t.style.Height(height)
	return t
}

// MaxWidth sets the maximum width of the text.
func (t *Text) MaxWidth(width int) *Text {
	t.style = t.style.MaxWidth(width)
	return t
}

// MaxHeight sets the maximum height of the text.
func (t *Text) MaxHeight(height int) *Text {
	t.style = t.style.MaxHeight(height)
	return t
}

// Align sets the horizontal alignment of the text.
func (t *Text) Align(position lipgloss.Position) *Text {
	t.style = t.style.Align(position)
	return t
}

// AlignHorizontal sets the horizontal alignment.
func (t *Text) AlignHorizontal(position lipgloss.Position) *Text {
	t.style = t.style.AlignHorizontal(position)
	return t
}

// AlignVertical sets the vertical alignment.
func (t *Text) AlignVertical(position lipgloss.Position) *Text {
	t.style = t.style.AlignVertical(position)
	return t
}

// Padding Modifiers

// Padding sets padding on all sides.
func (t *Text) Padding(vertical, horizontal int) *Text {
	t.style = t.style.Padding(vertical, horizontal)
	return t
}

// PaddingAll sets equal padding on all sides.
func (t *Text) PaddingAll(padding int) *Text {
	t.style = t.style.Padding(padding)
	return t
}

// PaddingTop sets top padding.
func (t *Text) PaddingTop(padding int) *Text {
	t.style = t.style.PaddingTop(padding)
	return t
}

// PaddingBottom sets bottom padding.
func (t *Text) PaddingBottom(padding int) *Text {
	t.style = t.style.PaddingBottom(padding)
	return t
}

// PaddingLeft sets left padding.
func (t *Text) PaddingLeft(padding int) *Text {
	t.style = t.style.PaddingLeft(padding)
	return t
}

// PaddingRight sets right padding.
func (t *Text) PaddingRight(padding int) *Text {
	t.style = t.style.PaddingRight(padding)
	return t
}

// Margin Modifiers

// Margin sets margin on all sides.
func (t *Text) Margin(vertical, horizontal int) *Text {
	t.style = t.style.Margin(vertical, horizontal)
	return t
}

// MarginAll sets equal margin on all sides.
func (t *Text) MarginAll(margin int) *Text {
	t.style = t.style.Margin(margin)
	return t
}

// MarginTop sets top margin.
func (t *Text) MarginTop(margin int) *Text {
	t.style = t.style.MarginTop(margin)
	return t
}

// MarginBottom sets bottom margin.
func (t *Text) MarginBottom(margin int) *Text {
	t.style = t.style.MarginBottom(margin)
	return t
}

// MarginLeft sets left margin.
func (t *Text) MarginLeft(margin int) *Text {
	t.style = t.style.MarginLeft(margin)
	return t
}

// MarginRight sets right margin.
func (t *Text) MarginRight(margin int) *Text {
	t.style = t.style.MarginRight(margin)
	return t
}

// Border Modifiers

// Border sets a border with the given style.
func (t *Text) Border(border lipgloss.Border, sides ...bool) *Text {
	t.style = t.style.Border(border, sides...)
	return t
}

// BorderStyle sets the border style.
func (t *Text) BorderStyle(border lipgloss.Border) *Text {
	t.style = t.style.BorderStyle(border)
	return t
}

// BorderTop sets the top border.
func (t *Text) BorderTop(border bool) *Text {
	t.style = t.style.BorderTop(border)
	return t
}

// BorderBottom sets the bottom border.
func (t *Text) BorderBottom(border bool) *Text {
	t.style = t.style.BorderBottom(border)
	return t
}

// BorderLeft sets the left border.
func (t *Text) BorderLeft(border bool) *Text {
	t.style = t.style.BorderLeft(border)
	return t
}

// BorderRight sets the right border.
func (t *Text) BorderRight(border bool) *Text {
	t.style = t.style.BorderRight(border)
	return t
}

// BorderForeground sets the border color.
func (t *Text) BorderForeground(color lipgloss.Color) *Text {
	t.style = t.style.BorderForeground(color)
	return t
}

// BorderBackground sets the border background color.
func (t *Text) BorderBackground(color lipgloss.Color) *Text {
	t.style = t.style.BorderBackground(color)
	return t
}

// Common preset styles

// Primary applies a primary style (bold and colored).
func (t *Text) Primary() *Text {
	return t.Bold(true).Foreground(lipgloss.Color("81"))
}

// Secondary applies a secondary style (faint).
func (t *Text) Secondary() *Text {
	return t.Faint(true)
}

// Success applies a success style (green).
func (t *Text) Success() *Text {
	return t.Foreground(lipgloss.Color("42"))
}

// Warning applies a warning style (yellow).
func (t *Text) Warning() *Text {
	return t.Foreground(lipgloss.Color("226"))
}

// Error applies an error style (red and bold).
func (t *Text) Error() *Text {
	return t.Bold(true).Foreground(lipgloss.Color("196"))
}

// Info applies an info style (blue).
func (t *Text) Info() *Text {
	return t.Foreground(lipgloss.Color("39"))
}

// Muted applies a muted style (faint and gray).
func (t *Text) Muted() *Text {
	return t.Faint(true).Foreground(lipgloss.Color("240"))
}
