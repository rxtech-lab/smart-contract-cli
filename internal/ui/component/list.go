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
	// GetDescription returns the description for the item
	GetDescription() string
}

// SimpleItem is a basic implementation of ListItem.
type SimpleItem struct {
	Label       string
	Value       string
	Description string
}

func (s SimpleItem) GetLabel() string       { return s.Label }
func (s SimpleItem) GetValue() string       { return s.Value }
func (s SimpleItem) GetDescription() string { return s.Description }

// Item creates a simple list item.
func Item(label, value, description string) ListItem {
	return SimpleItem{Label: label, Value: value, Description: description}
}

// List is a component that renders a list of selectable items.
type List struct {
	items              []ListItem
	selectedValue      string
	highlightedValues  []string
	renderItem         func(item ListItem, isSelected bool, isHighlighted bool) Component
	selectedPrefix     string
	unselectedPrefix   string
	highlightedPrefix  string
	spacing            int
	style              lipgloss.Style
	selectedStyle      lipgloss.Style
	highlightedStyle   lipgloss.Style
	showDescription    bool
	descriptionStyle   lipgloss.Style
	descriptionSpacing int
}

// NewList creates a new list component.
func NewList(items []ListItem) *List {
	return &List{
		items:              items,
		selectedValue:      "",
		highlightedValues:  []string{},
		renderItem:         nil,
		selectedPrefix:     "> ",
		unselectedPrefix:   "  ",
		highlightedPrefix:  "★ ",
		spacing:            0,
		style:              lipgloss.NewStyle(),
		selectedStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("51")),            // Cyan
		highlightedStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true), // Green + Bold
		showDescription:    false,
		descriptionStyle:   lipgloss.NewStyle().Faint(true),
		descriptionSpacing: 0,
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

// HighlightedPrefix sets the prefix for highlighted items.
func (l *List) HighlightedPrefix(prefix string) *List {
	l.highlightedPrefix = prefix
	return l
}

// Highlighted marks certain items as highlighted/active by their values.
func (l *List) Highlighted(values ...string) *List {
	l.highlightedValues = values
	return l
}

// SelectedStyle sets the style for cursor-selected items.
func (l *List) SelectedStyle(style lipgloss.Style) *List {
	l.selectedStyle = style
	return l
}

// HighlightedStyle sets the style for highlighted/active items.
func (l *List) HighlightedStyle(style lipgloss.Style) *List {
	l.highlightedStyle = style
	return l
}

// Spacing sets the vertical spacing between items.
func (l *List) Spacing(spacing int) *List {
	l.spacing = spacing
	return l
}

// RenderItem sets a custom renderer for list items.
// If not set, uses the default renderer with prefixes and labels.
func (l *List) RenderItem(fn func(item ListItem, isSelected bool, isHighlighted bool) Component) *List {
	l.renderItem = fn
	return l
}

// WithStyle applies a lipgloss style to the list container.
func (l *List) WithStyle(style lipgloss.Style) *List {
	l.style = style
	return l
}

// ShowDescription enables rendering descriptions for selected items.
func (l *List) ShowDescription(show bool) *List {
	l.showDescription = show
	return l
}

// DescriptionStyle sets the style for description text.
func (l *List) DescriptionStyle(style lipgloss.Style) *List {
	l.descriptionStyle = style
	return l
}

// DescriptionSpacing sets the spacing between label and description.
func (l *List) DescriptionSpacing(spacing int) *List {
	l.descriptionSpacing = spacing
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
		isHighlighted := l.isHighlighted(item.GetValue())
		itemComponent := l.renderListItem(item, isSelected, isHighlighted)
		components = append(components, itemComponent)
	}

	vstack := NewVStack(components...).Spacing(l.spacing)
	result := vstack.Render()
	return l.style.Render(result)
}

// isHighlighted checks if an item value is in the highlighted list.
func (l *List) isHighlighted(value string) bool {
	for _, hv := range l.highlightedValues {
		if hv == value {
			return true
		}
	}
	return false
}

// renderListItem renders a single list item.
func (l *List) renderListItem(item ListItem, isSelected bool, isHighlighted bool) Component {
	var itemComponent Component
	if l.renderItem != nil {
		itemComponent = l.renderItem(item, isSelected, isHighlighted)
	} else {
		itemComponent = l.renderDefaultItem(item, isSelected, isHighlighted)
	}

	if l.showDescription && (isSelected || isHighlighted) {
		itemComponent = l.addDescription(itemComponent, item, isHighlighted)
	}

	return itemComponent
}

// renderDefaultItem renders an item using the default renderer.
func (l *List) renderDefaultItem(item ListItem, isSelected bool, isHighlighted bool) Component {
	// Determine prefix based on selection state
	prefix := l.unselectedPrefix
	if isSelected {
		prefix = l.selectedPrefix
	}

	label := item.GetLabel()

	// Build the text with prefix
	text := prefix + label

	// Add highlighted marker if highlighted
	if isHighlighted {
		text = text + " " + l.highlightedPrefix
	}

	// Create text component
	textComp := T(text)

	// Apply styles based on state priority
	if isSelected && isHighlighted {
		// Both: combine styles (selected background + highlighted foreground)
		combinedStyle := l.selectedStyle.Inherit(l.highlightedStyle)
		textComp = textComp.WithStyle(combinedStyle)
	} else if isSelected {
		// Only selected: use selected style
		textComp = textComp.WithStyle(l.selectedStyle)
	} else if isHighlighted {
		// Only highlighted: use highlighted style
		textComp = textComp.WithStyle(l.highlightedStyle)
	}

	return textComp
}

// addDescription adds a description below the item component.
func (l *List) addDescription(itemComponent Component, item ListItem, isHighlighted bool) Component {
	description := item.GetDescription()
	if description == "" {
		return itemComponent
	}

	descStyle := l.descriptionStyle
	// Apply highlighted style to description if item is highlighted
	if isHighlighted {
		descStyle = l.highlightedStyle.Faint(true)
	}

	descComponent := NewText(description).WithStyle(descStyle)
	prefixPadding := l.createPrefixPadding()
	descWithPadding := HStackC(T(prefixPadding), descComponent)

	return NewVStack(itemComponent, descWithPadding).
		Spacing(l.descriptionSpacing)
}

// createPrefixPadding creates padding equal to the selected prefix length.
func (l *List) createPrefixPadding() string {
	if l.selectedPrefix == "" {
		return ""
	}
	prefixPadding := ""
	for i := 0; i < len(l.selectedPrefix); i++ {
		prefixPadding += " "
	}
	return prefixPadding
}

// StringList creates a list from a slice of strings.
// Each string becomes both the label and value.
func StringList(items []string) *List {
	listItems := make([]ListItem, len(items))
	for i, item := range items {
		listItems[i] = SimpleItem{Label: item, Value: item, Description: ""}
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
