package component

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item that can be displayed in a list.
type ListItem interface {
	// GetLabel returns the display label for the item
	GetLabel() string
	// GetValue returns the value associated with the item
	GetValue() string
}

// SimpleItem is a basic implementation of ListItem.
type SimpleItem struct {
	Label string
	Value string
}

func (s SimpleItem) GetLabel() string { return s.Label }
func (s SimpleItem) GetValue() string { return s.Value }

// Item creates a simple list item.
func Item(label, value string) ListItem {
	return SimpleItem{Label: label, Value: value}
}

// List is a component that renders a list of selectable items.
type List struct {
	items            []ListItem
	selectedValue    string
	renderItem       func(item ListItem, isSelected bool) Component
	selectedPrefix   string
	unselectedPrefix string
	spacing          int
	style            lipgloss.Style
}

// NewList creates a new list component.
func NewList(items []ListItem) *List {
	return &List{
		items:            items,
		selectedValue:    "",
		renderItem:       nil,
		selectedPrefix:   "> ",
		unselectedPrefix: "  ",
		spacing:          0,
		style:            lipgloss.NewStyle(),
	}
}

// ListC is a convenience function for creating a list.
func ListC(items []ListItem) *List {
	return NewList(items)
}

// Selected sets the currently selected item by value.
func (l *List) Selected(value string) *List {
	l.selectedValue = value
	return l
}

// SelectedPrefix sets the prefix for selected items.
func (l *List) SelectedPrefix(prefix string) *List {
	l.selectedPrefix = prefix
	return l
}

// UnselectedPrefix sets the prefix for unselected items.
func (l *List) UnselectedPrefix(prefix string) *List {
	l.unselectedPrefix = prefix
	return l
}

// Spacing sets the vertical spacing between items.
func (l *List) Spacing(spacing int) *List {
	l.spacing = spacing
	return l
}

// RenderItem sets a custom renderer for list items.
// If not set, uses the default renderer with prefixes and labels.
func (l *List) RenderItem(fn func(item ListItem, isSelected bool) Component) *List {
	l.renderItem = fn
	return l
}

// WithStyle applies a lipgloss style to the list container.
func (l *List) WithStyle(style lipgloss.Style) *List {
	l.style = style
	return l
}

// Render renders the list.
func (l *List) Render() string {
	if len(l.items) == 0 {
		return l.style.Render("")
	}

	components := make([]Component, 0, len(l.items))

	for _, item := range l.items {
		isSelected := item.GetValue() == l.selectedValue

		var itemComponent Component
		if l.renderItem != nil {
			// Use custom renderer
			itemComponent = l.renderItem(item, isSelected)
		} else {
			// Use default renderer
			prefix := l.unselectedPrefix
			if isSelected {
				prefix = l.selectedPrefix
			}
			itemComponent = T(prefix + item.GetLabel())
		}

		components = append(components, itemComponent)
	}

	vstack := NewVStack(components...).Spacing(l.spacing)
	result := vstack.Render()
	return l.style.Render(result)
}

// StringList creates a list from a slice of strings.
// Each string becomes both the label and value.
func StringList(items []string) *List {
	listItems := make([]ListItem, len(items))
	for i, item := range items {
		listItems[i] = SimpleItem{Label: item, Value: item}
	}
	return NewList(listItems)
}

// NumberedList creates a numbered list component.
type NumberedList struct {
	items   []Component
	start   int
	spacing int
	style   lipgloss.Style
}

// NewNumberedList creates a new numbered list.
func NewNumberedList(items ...Component) *NumberedList {
	return &NumberedList{
		items:   items,
		start:   1,
		spacing: 0,
		style:   lipgloss.NewStyle(),
	}
}

// Start sets the starting number for the list.
func (n *NumberedList) Start(start int) *NumberedList {
	n.start = start
	return n
}

// Spacing sets the vertical spacing between items.
func (n *NumberedList) Spacing(spacing int) *NumberedList {
	n.spacing = spacing
	return n
}

// WithStyle applies a lipgloss style to the list.
func (n *NumberedList) WithStyle(style lipgloss.Style) *NumberedList {
	n.style = style
	return n
}

// Render renders the numbered list.
func (n *NumberedList) Render() string {
	if len(n.items) == 0 {
		return n.style.Render("")
	}

	components := make([]Component, 0, len(n.items))

	for i, item := range n.items {
		number := n.start + i
		prefix := fmt.Sprintf("%d. ", number)
		numbered := HStackC(T(prefix), item)
		components = append(components, numbered)
	}

	vstack := NewVStack(components...).Spacing(n.spacing)
	result := vstack.Render()
	return n.style.Render(result)
}

// BulletList creates a bulleted list component.
type BulletList struct {
	items   []Component
	bullet  string
	spacing int
	style   lipgloss.Style
}

// NewBulletList creates a new bulleted list.
func NewBulletList(items ...Component) *BulletList {
	return &BulletList{
		items:   items,
		bullet:  "• ",
		spacing: 0,
		style:   lipgloss.NewStyle(),
	}
}

// Bullet sets the bullet character/string.
func (b *BulletList) Bullet(bullet string) *BulletList {
	b.bullet = bullet
	return b
}

// Spacing sets the vertical spacing between items.
func (b *BulletList) Spacing(spacing int) *BulletList {
	b.spacing = spacing
	return b
}

// WithStyle applies a lipgloss style to the list.
func (b *BulletList) WithStyle(style lipgloss.Style) *BulletList {
	b.style = style
	return b
}

// Render renders the bulleted list.
func (b *BulletList) Render() string {
	if len(b.items) == 0 {
		return b.style.Render("")
	}

	components := make([]Component, 0, len(b.items))

	for _, item := range b.items {
		bulleted := HStackC(T(b.bullet), item)
		components = append(components, bulleted)
	}

	vstack := NewVStack(components...).Spacing(b.spacing)
	result := vstack.Render()
	return b.style.Render(result)
}
