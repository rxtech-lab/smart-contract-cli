package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/suite"
)

type ComponentTestSuite struct {
	suite.Suite
}

func TestComponentTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentTestSuite))
}

// Test ComponentFunc
func (s *ComponentTestSuite) TestComponentFunc() {
	comp := ComponentFunc(func() string {
		return "Hello, World!"
	})

	s.Equal("Hello, World!", comp.Render())
}

// Test Empty
func (s *ComponentTestSuite) TestEmpty() {
	comp := Empty()
	s.Equal("", comp.Render())
}

// Test Raw
func (s *ComponentTestSuite) TestRaw() {
	comp := Raw("Raw content")
	s.Equal("Raw content", comp.Render())
}

// Test Join
func (s *ComponentTestSuite) TestJoin() {
	c1 := T("Hello")
	c2 := T("World")
	c3 := T("!")

	joined := Join(", ", c1, c2, c3)
	s.Equal("Hello, World, !", joined.Render())
}

func (s *ComponentTestSuite) TestJoinEmpty() {
	joined := Join(", ")
	s.Equal("", joined.Render())
}

func (s *ComponentTestSuite) TestJoinSingle() {
	c1 := T("Hello")
	joined := Join(", ", c1)
	s.Equal("Hello", joined.Render())
}

// Test Text component
func (s *ComponentTestSuite) TestText() {
	text := T("Hello")
	s.Contains(text.Render(), "Hello")
}

func (s *ComponentTestSuite) TestTextChaining() {
	text := T("Styled").Bold(true).Italic(true)
	rendered := text.Render()
	s.NotEmpty(rendered)
}

func (s *ComponentTestSuite) TestTextPresets() {
	s.NotEmpty(T("Primary").Primary().Render())
	s.NotEmpty(T("Success").Success().Render())
	s.NotEmpty(T("Error").Error().Render())
	s.NotEmpty(T("Warning").Warning().Render())
	s.NotEmpty(T("Info").Info().Render())
	s.NotEmpty(T("Muted").Muted().Render())
}

// Test VStack
func (s *ComponentTestSuite) TestVStack() {
	stack := NewVStack(
		T("Line 1"),
		T("Line 2"),
		T("Line 3"),
	)

	rendered := stack.Render()
	s.Contains(rendered, "Line 1")
	s.Contains(rendered, "Line 2")
	s.Contains(rendered, "Line 3")
}

func (s *ComponentTestSuite) TestVStackEmpty() {
	stack := NewVStack()
	s.Equal("", stack.Render())
}

func (s *ComponentTestSuite) TestVStackSpacing() {
	stack := NewVStack(
		T("Line 1"),
		T("Line 2"),
	).Spacing(1)

	rendered := stack.Render()
	// With spacing of 1, there should be an empty line between items
	lines := strings.Split(rendered, "\n")
	s.GreaterOrEqual(len(lines), 2)
}

// Test HStack
func (s *ComponentTestSuite) TestHStack() {
	stack := NewHStack(
		T("A"),
		T("B"),
		T("C"),
	)

	rendered := stack.Render()
	s.Contains(rendered, "A")
	s.Contains(rendered, "B")
	s.Contains(rendered, "C")
}

func (s *ComponentTestSuite) TestHStackEmpty() {
	stack := NewHStack()
	s.Equal("", stack.Render())
}

func (s *ComponentTestSuite) TestHStackSpacing() {
	stack := NewHStack(
		T("A"),
		T("B"),
	).Spacing(2)

	rendered := stack.Render()
	s.NotEmpty(rendered)
}

// Test Spacer
func (s *ComponentTestSuite) TestSpacerVertical() {
	spacer := SpacerV(2)
	rendered := spacer.Render()
	s.Equal("\n", rendered)
}

func (s *ComponentTestSuite) TestSpacerHorizontal() {
	spacer := SpacerH(5)
	rendered := spacer.Render()
	s.Equal("     ", rendered)
}

func (s *ComponentTestSuite) TestSpacerZero() {
	spacer := NewSpacer(0, 0)
	s.Equal("", spacer.Render())
}

// Test Divider
func (s *ComponentTestSuite) TestDivider() {
	div := NewDivider("─", 10)
	rendered := div.Render()
	s.Equal(strings.Repeat("─", 10), rendered)
}

func (s *ComponentTestSuite) TestDividerLine() {
	div := DividerLine(5)
	rendered := div.Render()
	s.Contains(rendered, "─")
}

// Test List
func (s *ComponentTestSuite) TestList() {
	items := []ListItem{
		Item("Item 1", "1", "Description 1"),
		Item("Item 2", "2", "Description 2"),
		Item("Item 3", "3", "Description 3"),
	}

	list := NewList(items).Selected("2")
	rendered := list.Render()

	s.Contains(rendered, "Item 1")
	s.Contains(rendered, "Item 2")
	s.Contains(rendered, "Item 3")
}

func (s *ComponentTestSuite) TestListEmpty() {
	list := NewList([]ListItem{})
	s.Equal("", list.Render())
}

func (s *ComponentTestSuite) TestListCustomPrefix() {
	items := []ListItem{
		Item("Item 1", "1", "Description 1"),
	}

	list := NewList(items).
		Selected("1").
		SelectedPrefix("→ ").
		UnselectedPrefix("  ")

	rendered := list.Render()
	s.Contains(rendered, "→")
}

func (s *ComponentTestSuite) TestStringList() {
	list := StringList([]string{"A", "B", "C"}).Selected("B")
	rendered := list.Render()

	s.Contains(rendered, "A")
	s.Contains(rendered, "B")
	s.Contains(rendered, "C")
}

func (s *ComponentTestSuite) TestListWithDescription() {
	items := []ListItem{
		Item("Item 1", "1", "Description 1"),
		Item("Item 2", "2", "Description 2"),
		Item("Item 3", "3", "Description 3"),
	}

	list := NewList(items).
		Selected("2").
		ShowDescription(true)

	rendered := list.Render()

	// Should contain the selected item
	s.Contains(rendered, "Item 2")
	// Should contain the description of the selected item
	s.Contains(rendered, "Description 2")
	// Should not contain descriptions of unselected items
	s.NotContains(rendered, "Description 1")
	s.NotContains(rendered, "Description 3")
}

func (s *ComponentTestSuite) TestListWithDescriptionDisabled() {
	items := []ListItem{
		Item("Item 1", "1", "Description 1"),
		Item("Item 2", "2", "Description 2"),
	}

	list := NewList(items).
		Selected("1").
		ShowDescription(false) // Explicitly disabled

	rendered := list.Render()

	// Should contain items
	s.Contains(rendered, "Item 1")
	s.Contains(rendered, "Item 2")
	// Should not contain any descriptions
	s.NotContains(rendered, "Description 1")
	s.NotContains(rendered, "Description 2")
}

func (s *ComponentTestSuite) TestListWithEmptyDescription() {
	items := []ListItem{
		Item("Item 1", "1", ""), // Empty description
		Item("Item 2", "2", "Description 2"),
	}

	list := NewList(items).
		Selected("1").
		ShowDescription(true)

	rendered := list.Render()

	// Should handle empty description gracefully
	s.Contains(rendered, "Item 1")
	s.NotContains(rendered, "Description 2")
}

func (s *ComponentTestSuite) TestListDescriptionSpacing() {
	items := []ListItem{
		Item("Item 1", "1", "Description 1"),
	}

	list := NewList(items).
		Selected("1").
		ShowDescription(true).
		DescriptionSpacing(2)

	rendered := list.Render()

	// Should contain both item and description
	s.Contains(rendered, "Item 1")
	s.Contains(rendered, "Description 1")
	// Check that there's spacing (multiple newlines)
	s.Contains(rendered, "\n")
}

func (s *ComponentTestSuite) TestListDescriptionStyle() {
	items := []ListItem{
		Item("Item 1", "1", "Description 1"),
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	list := NewList(items).
		Selected("1").
		ShowDescription(true).
		DescriptionStyle(style)

	rendered := list.Render()

	// Should contain the description
	s.Contains(rendered, "Description 1")
}

func (s *ComponentTestSuite) TestNumberedList() {
	list := NewNumberedList(
		T("First"),
		T("Second"),
		T("Third"),
	)

	rendered := list.Render()
	s.Contains(rendered, "1.")
	s.Contains(rendered, "2.")
	s.Contains(rendered, "3.")
}

func (s *ComponentTestSuite) TestNumberedListCustomStart() {
	list := NewNumberedList(
		T("First"),
		T("Second"),
	).Start(5)

	rendered := list.Render()
	s.Contains(rendered, "5.")
	s.Contains(rendered, "6.")
}

func (s *ComponentTestSuite) TestBulletList() {
	list := NewBulletList(
		T("First"),
		T("Second"),
		T("Third"),
	)

	rendered := list.Render()
	s.Contains(rendered, "•")
	s.Contains(rendered, "First")
	s.Contains(rendered, "Second")
	s.Contains(rendered, "Third")
}

func (s *ComponentTestSuite) TestBulletListCustomBullet() {
	list := NewBulletList(
		T("Item"),
	).Bullet("* ")

	rendered := list.Render()
	s.Contains(rendered, "*")
}

// Test Box
func (s *ComponentTestSuite) TestBox() {
	box := NewBox(T("Content"))
	rendered := box.Render()
	s.Contains(rendered, "Content")
}

func (s *ComponentTestSuite) TestBoxWithBorder() {
	box := NewBox(T("Content")).RoundedBorder()
	rendered := box.Render()
	s.NotEmpty(rendered)
}

func (s *ComponentTestSuite) TestCard() {
	card := Card(T("Card content"))
	rendered := card.Render()
	s.Contains(rendered, "Card content")
}

func (s *ComponentTestSuite) TestPanel() {
	panel := Panel(T("Panel content"))
	rendered := panel.Render()
	s.Contains(rendered, "Panel content")
}

// Test Padding
func (s *ComponentTestSuite) TestPadding() {
	padded := NewPadding(T("Content")).All(1)
	rendered := padded.Render()
	s.Contains(rendered, "Content")
}

func (s *ComponentTestSuite) TestPaddingVertical() {
	padded := NewPadding(T("Content")).Vertical(2)
	rendered := padded.Render()
	s.NotEmpty(rendered)
}

func (s *ComponentTestSuite) TestPaddingHorizontal() {
	padded := NewPadding(T("Content")).Horizontal(3)
	rendered := padded.Render()
	s.NotEmpty(rendered)
}

// Test Center
func (s *ComponentTestSuite) TestCenter() {
	centered := NewCenter(T("Centered"), 20, 5)
	rendered := centered.Render()
	s.Contains(rendered, "Centered")
}

// Test Conditional - If
func (s *ComponentTestSuite) TestIfTrue() {
	comp := IfC(true, T("True"), T("False"))
	s.Equal("True", comp.Render())
}

func (s *ComponentTestSuite) TestIfFalse() {
	comp := IfC(false, T("True"), T("False"))
	s.Equal("False", comp.Render())
}

// Test Conditional - IfThen
func (s *ComponentTestSuite) TestIfThenTrue() {
	comp := IfThenC(true, T("Shown"))
	s.Equal("Shown", comp.Render())
}

func (s *ComponentTestSuite) TestIfThenFalse() {
	comp := IfThenC(false, T("Hidden"))
	s.Equal("", comp.Render())
}

// Test Conditional - Unless
func (s *ComponentTestSuite) TestUnlessTrue() {
	comp := UnlessC(true, T("Hidden"))
	s.Equal("", comp.Render())
}

func (s *ComponentTestSuite) TestUnlessFalse() {
	comp := UnlessC(false, T("Shown"))
	s.Equal("Shown", comp.Render())
}

// Test Conditional - Switch
func (s *ComponentTestSuite) TestSwitch() {
	value := 2

	comp := SwitchC(
		Case(Match(value, 1), T("One")),
		Case(Match(value, 2), T("Two")),
		Case(Match(value, 3), T("Three")),
	).Default(T("Other"))

	s.Equal("Two", comp.Render())
}

func (s *ComponentTestSuite) TestSwitchDefault() {
	value := 99

	comp := SwitchC(
		Case(Match(value, 1), T("One")),
		Case(Match(value, 2), T("Two")),
	).Default(T("Other"))

	s.Equal("Other", comp.Render())
}

func (s *ComponentTestSuite) TestMatchAny() {
	value := 2

	comp := SwitchC(
		Case(MatchAny(value, 1, 2, 3), T("Low")),
		Case(MatchAny(value, 4, 5, 6), T("High")),
	).Default(T("Other"))

	s.Equal("Low", comp.Render())
}

func (s *ComponentTestSuite) TestMatchRange() {
	value := 5

	comp := SwitchC(
		Case(MatchRange(value, 1, 3), T("Low")),
		Case(MatchRange(value, 4, 6), T("Mid")),
		Case(MatchRange(value, 7, 9), T("High")),
	).Default(T("Other"))

	s.Equal("Mid", comp.Render())
}

// Test Show/Hide helpers
func (s *ComponentTestSuite) TestShow() {
	s.Equal("Shown", Show(true, T("Shown")).Render())
	s.Equal("", Show(false, T("Hidden")).Render())
}

func (s *ComponentTestSuite) TestHide() {
	s.Equal("", Hide(true, T("Hidden")).Render())
	s.Equal("Shown", Hide(false, T("Shown")).Render())
}

// Test Toggle
func (s *ComponentTestSuite) TestToggle() {
	s.Equal("On", Toggle(true, T("On"), T("Off")).Render())
	s.Equal("Off", Toggle(false, T("On"), T("Off")).Render())
}

// Integration tests - Complex compositions
func (s *ComponentTestSuite) TestComplexComposition() {
	comp := VStackC(
		T("Title").Bold(true),
		SpacerV(1),
		Card(
			VStackC(
				T("Card Title"),
				DividerLine(20),
				T("Card Content"),
			),
		),
		SpacerV(1),
		HStackC(
			T("Left"),
			SpacerH(5),
			T("Right"),
		),
	)

	rendered := comp.Render()
	s.Contains(rendered, "Title")
	s.Contains(rendered, "Card Title")
	s.Contains(rendered, "Card Content")
	s.Contains(rendered, "Left")
	s.Contains(rendered, "Right")
}

func (s *ComponentTestSuite) TestNestedConditionals() {
	isLoggedIn := true
	hasPermission := false

	comp := VStackC(
		When(isLoggedIn,
			VStackC(
				T("Welcome!"),
				IfC(hasPermission,
					T("Admin Panel"),
					T("User Panel"),
				),
			),
		),
		UnlessC(isLoggedIn,
			T("Please log in"),
		),
	)

	rendered := comp.Render()
	s.Contains(rendered, "Welcome!")
	s.Contains(rendered, "User Panel")
	s.NotContains(rendered, "Admin Panel")
	s.NotContains(rendered, "Please log in")
}

// Test style application
func (s *ComponentTestSuite) TestStyleApplication() {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	text := T("Styled").WithStyle(style)
	rendered := text.Render()
	s.NotEmpty(rendered)
}

// Test List with highlighting
func (s *ComponentTestSuite) TestListWithHighlighting() {
	items := []ListItem{
		Item("SQLite", "sqlite", "Local database"),
		Item("Postgres", "postgres", "Remote database"),
		Item("MySQL", "mysql", "Another database"),
	}

	list := NewList(items).
		Highlighted("postgres"). // Highlight postgres as active
		Selected("mysql")        // Select mysql with cursor

	rendered := list.Render()

	// Should contain all items
	s.Contains(rendered, "SQLite")
	s.Contains(rendered, "Postgres")
	s.Contains(rendered, "MySQL")

	// Should contain the highlighted marker
	s.Contains(rendered, "★")
}

func (s *ComponentTestSuite) TestListHighlightedOnly() {
	items := []ListItem{
		Item("Item 1", "1", ""),
		Item("Item 2", "2", ""),
		Item("Item 3", "3", ""),
	}

	// Highlight item 2 but don't select anything
	list := NewList(items).Highlighted("2")
	rendered := list.Render()

	// Should contain all items
	s.Contains(rendered, "Item 1")
	s.Contains(rendered, "Item 2")
	s.Contains(rendered, "Item 3")

	// Should contain the highlighted marker for item 2
	s.Contains(rendered, "★")
}

func (s *ComponentTestSuite) TestListMultipleHighlighted() {
	items := []ListItem{
		Item("Item 1", "1", ""),
		Item("Item 2", "2", ""),
		Item("Item 3", "3", ""),
		Item("Item 4", "4", ""),
	}

	// Highlight multiple items
	list := NewList(items).Highlighted("1", "3")
	rendered := list.Render()

	// Should contain all items
	s.Contains(rendered, "Item 1")
	s.Contains(rendered, "Item 2")
	s.Contains(rendered, "Item 3")
	s.Contains(rendered, "Item 4")

	// Counting occurrences of ★ symbol
	count := strings.Count(rendered, "★")
	s.Equal(2, count, "Should have exactly 2 highlighted markers")
}

func (s *ComponentTestSuite) TestListHighlightedWithDescription() {
	items := []ListItem{
		Item("SQLite", "sqlite", "Path: /data/db.sqlite"),
		Item("Postgres", "postgres", "URL: postgres://localhost"),
	}

	list := NewList(items).
		Highlighted("sqlite").
		ShowDescription(true)

	rendered := list.Render()

	// Should show description for highlighted item even if not selected
	s.Contains(rendered, "Path: /data/db.sqlite")
	s.NotContains(rendered, "URL: postgres://localhost")
}

func (s *ComponentTestSuite) TestListCustomHighlightedPrefix() {
	items := []ListItem{
		Item("Item 1", "1", ""),
		Item("Item 2", "2", ""),
	}

	list := NewList(items).
		Highlighted("1").
		HighlightedPrefix("✓ ")

	rendered := list.Render()

	// Should contain custom highlighted prefix
	s.Contains(rendered, "✓")
	// Should not contain default star
	s.NotContains(rendered, "★")
}

func (s *ComponentTestSuite) TestListHighlightedStyle() {
	items := []ListItem{
		Item("Item 1", "1", ""),
		Item("Item 2", "2", ""),
	}

	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	list := NewList(items).
		Highlighted("1").
		HighlightedStyle(customStyle)

	rendered := list.Render()

	// Should contain the item
	s.Contains(rendered, "Item 1")
	// Just verify it renders without error
	s.NotEmpty(rendered)
}

func (s *ComponentTestSuite) TestListSelectedStyle() {
	items := []ListItem{
		Item("Item 1", "1", ""),
		Item("Item 2", "2", ""),
	}

	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Pink
	list := NewList(items).
		Selected("2").
		SelectedStyle(customStyle)

	rendered := list.Render()

	// Should contain the item
	s.Contains(rendered, "Item 2")
	// Just verify it renders without error
	s.NotEmpty(rendered)
}

func (s *ComponentTestSuite) TestListSelectedAndHighlighted() {
	items := []ListItem{
		Item("SQLite", "sqlite", "Local database"),
		Item("Postgres", "postgres", "Remote database"),
	}

	list := NewList(items).
		Selected("sqlite").    // Cursor on SQLite
		Highlighted("sqlite"). // SQLite is also active
		ShowDescription(true)

	rendered := list.Render()

	// Should contain both item and description
	s.Contains(rendered, "SQLite")
	s.Contains(rendered, "Local database")
	// Should contain highlighted marker
	s.Contains(rendered, "★")
}
