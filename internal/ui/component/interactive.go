package component

import (
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Spinner wraps the bubbles spinner component as a declarative component.
type Spinner struct {
	model spinner.Model
	style lipgloss.Style
}

// NewSpinner creates a new spinner component.
func NewSpinner() *Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return &Spinner{
		model: s,
		style: lipgloss.NewStyle(),
	}
}

// SpinnerC is a convenience function for creating a spinner.
func SpinnerC() *Spinner {
	return NewSpinner()
}

// WithSpinnerType sets the spinner animation type.
func (s *Spinner) WithSpinnerType(spinnerType spinner.Spinner) *Spinner {
	s.model.Spinner = spinnerType
	return s
}

// WithStyle applies a lipgloss style to the spinner.
func (s *Spinner) WithStyle(style lipgloss.Style) *Spinner {
	s.style = style
	s.model.Style = style
	return s
}

// Render renders the spinner.
func (s *Spinner) Render() string {
	return s.style.Render(s.model.View())
}

// GetModel returns the underlying bubbles model for state management.
func (s *Spinner) GetModel() spinner.Model {
	return s.model
}

// Progress wraps the bubbles progress component as a declarative component.
type Progress struct {
	model   progress.Model
	percent float64
	style   lipgloss.Style
}

// NewProgress creates a new progress bar component.
func NewProgress(percent float64) *Progress {
	m := progress.New(progress.WithDefaultGradient())
	return &Progress{
		model:   m,
		percent: percent,
		style:   lipgloss.NewStyle(),
	}
}

// ProgressC is a convenience function for creating a progress bar.
func ProgressC(percent float64) *Progress {
	return NewProgress(percent)
}

// Width sets the width of the progress bar.
func (p *Progress) Width(width int) *Progress {
	p.model.Width = width
	return p
}

// WithStyle applies a lipgloss style to the progress bar.
func (p *Progress) WithStyle(style lipgloss.Style) *Progress {
	p.style = style
	return p
}

// ShowPercentage controls whether to show the percentage text.
func (p *Progress) ShowPercentage(show bool) *Progress {
	p.model.ShowPercentage = show
	return p
}

// Render renders the progress bar.
func (p *Progress) Render() string {
	return p.style.Render(p.model.ViewAs(p.percent))
}

// GetModel returns the underlying bubbles model for state management.
func (p *Progress) GetModel() progress.Model {
	return p.model
}

// Paginator wraps the bubbles paginator component as a declarative component.
type Paginator struct {
	model paginator.Model
	style lipgloss.Style
}

// NewPaginator creates a new paginator component.
func NewPaginator(totalPages int) *Paginator {
	p := paginator.New()
	p.SetTotalPages(totalPages)
	p.Type = paginator.Dots
	return &Paginator{
		model: p,
		style: lipgloss.NewStyle(),
	}
}

// PaginatorC is a convenience function for creating a paginator.
func PaginatorC(totalPages int) *Paginator {
	return NewPaginator(totalPages)
}

// Page sets the current page.
func (p *Paginator) Page(page int) *Paginator {
	p.model.Page = page
	return p
}

// PerPage sets items per page.
func (p *Paginator) PerPage(perPage int) *Paginator {
	p.model.PerPage = perPage
	return p
}

// PaginatorType sets the paginator style.
func (p *Paginator) PaginatorType(pType paginator.Type) *Paginator {
	p.model.Type = pType
	return p
}

// WithStyle applies a lipgloss style to the paginator.
func (p *Paginator) WithStyle(style lipgloss.Style) *Paginator {
	p.style = style
	return p
}

// Render renders the paginator.
func (p *Paginator) Render() string {
	return p.style.Render(p.model.View())
}

// GetModel returns the underlying bubbles model for state management.
func (p *Paginator) GetModel() paginator.Model {
	return p.model
}

// Table wraps the bubbles table component as a declarative component.
type Table struct {
	model table.Model
	style lipgloss.Style
}

// NewTable creates a new table component.
func NewTable(columns []table.Column, rows []table.Row) *Table {
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	return &Table{
		model: tbl,
		style: lipgloss.NewStyle(),
	}
}

// TableC is a convenience function for creating a table.
func TableC(columns []table.Column, rows []table.Row) *Table {
	return NewTable(columns, rows)
}

// Width sets the width of the table.
func (t *Table) Width(width int) *Table {
	t.model.SetWidth(width)
	return t
}

// Height sets the height of the table.
func (t *Table) Height(height int) *Table {
	t.model.SetHeight(height)
	return t
}

// WithStyle applies a lipgloss style to the table.
func (t *Table) WithStyle(style lipgloss.Style) *Table {
	t.style = style
	return t
}

// Focused sets whether the table is focused.
func (t *Table) Focused(focused bool) *Table {
	if focused {
		t.model.Focus()
	} else {
		t.model.Blur()
	}
	return t
}

// Render renders the table.
func (t *Table) Render() string {
	return t.style.Render(t.model.View())
}

// GetModel returns the underlying bubbles model for state management.
func (t *Table) GetModel() table.Model {
	return t.model
}

// TextInput wraps the bubbles textinput component as a declarative component.
type TextInput struct {
	model textinput.Model
	style lipgloss.Style
}

// NewTextInput creates a new text input component.
func NewTextInput() *TextInput {
	ti := textinput.New()
	ti.Focus()
	return &TextInput{
		model: ti,
		style: lipgloss.NewStyle(),
	}
}

// TextInputC is a convenience function for creating a text input.
func TextInputC() *TextInput {
	return NewTextInput()
}

// Placeholder sets the placeholder text.
func (t *TextInput) Placeholder(placeholder string) *TextInput {
	t.model.Placeholder = placeholder
	return t
}

// Value sets the current value.
func (t *TextInput) Value(value string) *TextInput {
	t.model.SetValue(value)
	return t
}

// Width sets the width of the input.
func (t *TextInput) Width(width int) *TextInput {
	t.model.Width = width
	return t
}

// Prompt sets the prompt string.
func (t *TextInput) Prompt(prompt string) *TextInput {
	t.model.Prompt = prompt
	return t
}

// WithStyle applies a lipgloss style to the text input.
func (t *TextInput) WithStyle(style lipgloss.Style) *TextInput {
	t.style = style
	return t
}

// Focused sets whether the input is focused.
func (t *TextInput) Focused(focused bool) *TextInput {
	if focused {
		t.model.Focus()
	} else {
		t.model.Blur()
	}
	return t
}

// CharLimit sets the maximum number of characters.
func (t *TextInput) CharLimit(limit int) *TextInput {
	t.model.CharLimit = limit
	return t
}

// EchoMode sets the echo mode for the input (useful for password fields).
func (t *TextInput) EchoMode(mode textinput.EchoMode) *TextInput {
	t.model.EchoMode = mode
	return t
}

// Render renders the text input.
func (t *TextInput) Render() string {
	return t.style.Render(t.model.View())
}

// GetModel returns the underlying bubbles model for state management.
func (t *TextInput) GetModel() textinput.Model {
	return t.model
}

// TextArea wraps the bubbles textarea component as a declarative component.
type TextArea struct {
	model textarea.Model
	style lipgloss.Style
}

// NewTextArea creates a new text area component.
func NewTextArea() *TextArea {
	ta := textarea.New()
	ta.Focus()
	return &TextArea{
		model: ta,
		style: lipgloss.NewStyle(),
	}
}

// TextAreaC is a convenience function for creating a text area.
func TextAreaC() *TextArea {
	return NewTextArea()
}

// Placeholder sets the placeholder text.
func (t *TextArea) Placeholder(placeholder string) *TextArea {
	t.model.Placeholder = placeholder
	return t
}

// Value sets the current value.
func (t *TextArea) Value(value string) *TextArea {
	t.model.SetValue(value)
	return t
}

// Width sets the width of the text area.
func (t *TextArea) Width(width int) *TextArea {
	t.model.SetWidth(width)
	return t
}

// Height sets the height of the text area.
func (t *TextArea) Height(height int) *TextArea {
	t.model.SetHeight(height)
	return t
}

// WithStyle applies a lipgloss style to the text area.
func (t *TextArea) WithStyle(style lipgloss.Style) *TextArea {
	t.style = style
	return t
}

// Focused sets whether the text area is focused.
func (t *TextArea) Focused(focused bool) *TextArea {
	if focused {
		t.model.Focus()
	} else {
		t.model.Blur()
	}
	return t
}

// MaxHeight sets the maximum height.
func (t *TextArea) MaxHeight(height int) *TextArea {
	t.model.MaxHeight = height
	return t
}

// ShowLineNumbers controls whether to show line numbers.
func (t *TextArea) ShowLineNumbers(show bool) *TextArea {
	t.model.ShowLineNumbers = show
	return t
}

// Render renders the text area.
func (t *TextArea) Render() string {
	return t.style.Render(t.model.View())
}

// GetModel returns the underlying bubbles model for state management.
func (t *TextArea) GetModel() textarea.Model {
	return t.model
}
