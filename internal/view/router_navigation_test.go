package view

import (
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/stretchr/testify/suite"
)

type HomeModel struct {
	router Router
}

type SubPageModel struct {
	router Router
}

func NewSubPage(router Router) View {
	return SubPageModel{router: router}
}

func (m SubPageModel) Init() tea.Cmd {
	return nil
}

func (m SubPageModel) Help() (string, HelpDisplayOption) {
	return "Use arrow keys to navigate and enter to select", HelpDisplayOptionAppend
}

func (m SubPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc", " ":
		m.router.Back()
	}
	return m, nil
}

func (m SubPageModel) View() string {
	return "Sub Page"
}

func NewPage(router Router) View {
	return HomeModel{
		router: router,
	}
}

func (m HomeModel) Init() tea.Cmd {
	return nil
}

func (m HomeModel) Help() (string, HelpDisplayOption) {
	return "Use arrow keys to navigate and enter to select", HelpDisplayOptionAppend
}

func (m HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "enter", " ":
		err := m.router.NavigateTo("/page2", nil)
		if err != nil {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m HomeModel) View() string {
	return "Home Page"
}

// RouterNavigationTestSuite tests router navigation using teatest.
type RouterNavigationTestSuite struct {
	suite.Suite
}

func TestRouterNavigationTestSuite(t *testing.T) {
	suite.Run(t, new(RouterNavigationTestSuite))
}

func (s *RouterNavigationTestSuite) getOutput(tm *teatest.TestModel) string {
	output, err := io.ReadAll(tm.Output())
	s.NoError(err, "Should be able to read output")
	return string(output)
}

// TestEnterKeyNavigation tests that pressing Enter navigates to the sub page.
func (s *RouterNavigationTestSuite) TestEnterKeyNavigation() {
	router := NewRouter()
	router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return NewPage(r) }})
	router.AddRoute(Route{Path: "/page2", Component: func(r Router, sharedMemory storage.SharedMemory) View { return NewSubPage(r) }})
	err := router.NavigateTo("/", nil)
	s.NoError(err, "Should navigate to root")

	testModel := teatest.NewTestModel(
		s.T(),
		router,
		teatest.WithInitialTermSize(300, 100),
	)

	// Verify initial state
	s.Equal("/", router.GetPath(), "Should be on root path")

	// Send Enter key to navigate to page2
	testModel.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Give time for the update to process
	time.Sleep(200 * time.Millisecond)

	// Verify navigation occurred by checking the router's current path
	// Note: We verify the navigation worked by checking the router state directly
	// since the teatest FinalOutput captures the entire terminal session
	s.Equal("/page2", router.GetPath(), "Should navigate to /page2 after Enter")
	s.Contains(s.getOutput(testModel), "Sub Page", "Should contain sub page content")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}

// TestEscKeyNavigation tests that pressing Esc navigates back to the previous page.
func (s *RouterNavigationTestSuite) TestEscKeyNavigation() {
	router := NewRouter()
	router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return NewPage(r) }})
	router.AddRoute(Route{Path: "/page2", Component: func(r Router, sharedMemory storage.SharedMemory) View { return NewSubPage(r) }})

	// Navigate to sub page first
	err := router.NavigateTo("/page2", nil)
	s.NoError(err, "Should navigate to page2")

	testModel := teatest.NewTestModel(
		s.T(),
		router,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait a moment for initial render to complete
	time.Sleep(100 * time.Millisecond)

	// Verify we're on sub page
	s.Equal("/page2", router.GetPath(), "Should be on /page2")
	s.Contains(s.getOutput(testModel), "Sub Page", "Should contain sub page content")

	// Send Esc key to go back
	testModel.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Give time for the update to process
	time.Sleep(500 * time.Millisecond)

	// Verify back navigation occurred by checking the router's current path
	s.Equal("/", router.GetPath(), "Should navigate back to / after Esc")
	output := s.getOutput(testModel)
	s.T().Logf("Output: %s", output)
	s.Contains(output, "Home Page", "Should contain home page content")

	// Quit
	testModel.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	testModel.WaitFinished(s.T(), teatest.WithFinalTimeout(time.Second))
}
