package view

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rxtech-lab/smart-contract-cli/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MockView is a mock implementation of View for testing.
type MockView struct {
	name        string
	initCalled  bool
	viewContent string
}

func (m *MockView) Init() tea.Cmd {
	m.initCalled = true
	return nil
}

func (m *MockView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *MockView) View() string {
	return m.viewContent
}

func (m *MockView) Help() (string, HelpDisplayOption) {
	return "", HelpDisplayOptionAppend
}

// RouterTestSuite is the test suite for Router.
type RouterTestSuite struct {
	suite.Suite
	router Router
}

func (suite *RouterTestSuite) SetupTest() {
	suite.router = NewRouter()
}

// TestNewRouter tests the router initialization.
func (suite *RouterTestSuite) TestNewRouter() {
	router := NewRouter()
	assert.NotNil(suite.T(), router)
	assert.Empty(suite.T(), router.GetRoutes())
	assert.Equal(suite.T(), "", router.GetPath())
}

// TestAddRoute tests adding routes.
func (suite *RouterTestSuite) TestAddRoute() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	route := Route{
		Path:      "/",
		Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView },
	}

	suite.router.AddRoute(route)

	routes := suite.router.GetRoutes()
	assert.Len(suite.T(), routes, 1)
	assert.Equal(suite.T(), "/", routes[0].Path)
}

// TestSetRoutes tests setting multiple routes at once.
func (suite *RouterTestSuite) TestSetRoutes() {
	mockView1 := &MockView{name: "home", viewContent: "Home View"}
	mockView2 := &MockView{name: "about", viewContent: "About View"}

	routes := []Route{
		{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }},
		{Path: "/about", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }},
	}

	suite.router.SetRoutes(routes)

	retrievedRoutes := suite.router.GetRoutes()
	assert.Len(suite.T(), retrievedRoutes, 2)
	assert.Equal(suite.T(), "/", retrievedRoutes[0].Path)
	assert.Equal(suite.T(), "/about", retrievedRoutes[1].Path)
}

// TestRemoveRoute tests removing a route.
func (suite *RouterTestSuite) TestRemoveRoute() {
	mockView1 := &MockView{name: "home", viewContent: "Home View"}
	mockView2 := &MockView{name: "about", viewContent: "About View"}

	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }})
	suite.router.AddRoute(Route{Path: "/about", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }})

	suite.router.RemoveRoute("/about")

	routes := suite.router.GetRoutes()
	assert.Len(suite.T(), routes, 1)
	assert.Equal(suite.T(), "/", routes[0].Path)
}

// TestNavigateTo tests navigation to a route.
func (suite *RouterTestSuite) TestNavigateTo() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "/", suite.router.GetPath())

	currentRoute := suite.router.GetCurrentRoute()
	assert.Equal(suite.T(), "/", currentRoute.Path)
}

// TestNavigateToWithQueryParams tests navigation with query parameters.
func (suite *RouterTestSuite) TestNavigateToWithQueryParams() {
	mockView := &MockView{name: "users", viewContent: "Users View"}
	suite.router.AddRoute(Route{Path: "/users", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	queryParams := map[string]string{
		"id":   "123",
		"name": "john",
	}

	err := suite.router.NavigateTo("/users", queryParams)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "123", suite.router.GetQueryParam("id"))
	assert.Equal(suite.T(), "john", suite.router.GetQueryParam("name"))
}

// TestNavigateToInvalidRoute tests navigation to non-existent route.
func (suite *RouterTestSuite) TestNavigateToInvalidRoute() {
	err := suite.router.NavigateTo("/nonexistent", nil)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no route found")
}

// TestNavigateToWithPathParams tests navigation with path parameters.
func (suite *RouterTestSuite) TestNavigateToWithPathParams() {
	mockView := &MockView{name: "user", viewContent: "User View"}
	suite.router.AddRoute(Route{Path: "/users/:id", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	err := suite.router.NavigateTo("/users/123", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "123", suite.router.GetParam("id"))
	assert.Equal(suite.T(), "/users/123", suite.router.GetPath())
}

// TestNavigateToWithMultiplePathParams tests navigation with multiple path parameters.
func (suite *RouterTestSuite) TestNavigateToWithMultiplePathParams() {
	mockView := &MockView{name: "comment", viewContent: "Comment View"}
	suite.router.AddRoute(Route{Path: "/posts/:postId/comments/:commentId", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	err := suite.router.NavigateTo("/posts/456/comments/789", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "456", suite.router.GetParam("postId"))
	assert.Equal(suite.T(), "789", suite.router.GetParam("commentId"))
}

// TestReplaceRoute tests replacing the current route.
func (suite *RouterTestSuite) TestReplaceRoute() {
	mockView1 := &MockView{name: "home", viewContent: "Home View"}
	mockView2 := &MockView{name: "about", viewContent: "About View"}

	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }})
	suite.router.AddRoute(Route{Path: "/about", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }})

	// Navigate to home
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	// Replace with about (should not add to stack)
	err = suite.router.ReplaceRoute("/about")
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "/about", suite.router.GetPath())
	assert.False(suite.T(), suite.router.CanGoBack())
}

// TestBackNavigation tests navigating back.
func (suite *RouterTestSuite) TestBackNavigation() {
	mockView1 := &MockView{name: "home", viewContent: "Home View"}
	mockView2 := &MockView{name: "about", viewContent: "About View"}
	mockView3 := &MockView{name: "contact", viewContent: "Contact View"}

	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }})
	suite.router.AddRoute(Route{Path: "/about", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }})
	suite.router.AddRoute(Route{Path: "/contact", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView3 }})

	// Navigate through routes
	err := suite.router.NavigateTo("/", nil)
	suite.NoError(err)
	err = suite.router.NavigateTo("/about", nil)
	suite.NoError(err)
	err = suite.router.NavigateTo("/contact", nil)
	suite.NoError(err)

	assert.Equal(suite.T(), "/contact", suite.router.GetPath())
	assert.True(suite.T(), suite.router.CanGoBack())

	// Go back
	suite.router.Back()
	assert.Equal(suite.T(), "/about", suite.router.GetPath())
	assert.True(suite.T(), suite.router.CanGoBack())

	// Go back again
	suite.router.Back()
	assert.Equal(suite.T(), "/", suite.router.GetPath())
	assert.False(suite.T(), suite.router.CanGoBack())
}

// TestBackWithEmptyStack tests back navigation with empty stack.
func (suite *RouterTestSuite) TestBackWithEmptyStack() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.router.CanGoBack())

	// Should not panic
	suite.router.Back()
	assert.Equal(suite.T(), "/", suite.router.GetPath())
}

// TestCanGoBack tests the CanGoBack method.
func (suite *RouterTestSuite) TestCanGoBack() {
	mockView1 := &MockView{name: "home", viewContent: "Home View"}
	mockView2 := &MockView{name: "about", viewContent: "About View"}

	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }})
	suite.router.AddRoute(Route{Path: "/about", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }})

	assert.False(suite.T(), suite.router.CanGoBack())

	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), suite.router.CanGoBack())

	err = suite.router.NavigateTo("/about", nil)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), suite.router.CanGoBack())
}

// TestGetCurrentRoute tests getting the current route.
func (suite *RouterTestSuite) TestGetCurrentRoute() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	componentFunc := func(r Router, sharedMemory storage.SharedMemory) View { return mockView }
	route := Route{Path: "/", Component: componentFunc}

	suite.router.AddRoute(route)
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	currentRoute := suite.router.GetCurrentRoute()
	assert.Equal(suite.T(), "/", currentRoute.Path)
	assert.NotNil(suite.T(), currentRoute.Component)
}

// TestGetCurrentRouteEmpty tests getting current route when none is set.
func (suite *RouterTestSuite) TestGetCurrentRouteEmpty() {
	currentRoute := suite.router.GetCurrentRoute()
	assert.Equal(suite.T(), "", currentRoute.Path)
	assert.Nil(suite.T(), currentRoute.Component)
}

// TestGetQueryParamNotFound tests getting a non-existent query parameter.
func (suite *RouterTestSuite) TestGetQueryParamNotFound() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	param := suite.router.GetQueryParam("nonexistent")
	assert.Equal(suite.T(), "", param)
}

// TestGetParamNotFound tests getting a non-existent path parameter.
func (suite *RouterTestSuite) TestGetParamNotFound() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	param := suite.router.GetParam("nonexistent")
	assert.Equal(suite.T(), "", param)
}

// TestGetPathEmpty tests getting path when no route is active.
func (suite *RouterTestSuite) TestGetPathEmpty() {
	path := suite.router.GetPath()
	assert.Equal(suite.T(), "", path)
}

// TestRefresh tests refreshing the current route.
func (suite *RouterTestSuite) TestRefresh() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	mockView.initCalled = false
	suite.router.Refresh()

	assert.True(suite.T(), mockView.initCalled)
}

// TestViewMethod tests the View method.
func (suite *RouterTestSuite) TestViewMethod() {
	mockView := &MockView{name: "home", viewContent: "Home View Content"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	view := suite.router.View()
	// The view should contain the content wrapped in a box with helper text
	assert.Contains(suite.T(), view, "Home View Content")
	assert.Contains(suite.T(), view, "Ctrl + c to exit")
}

// TestViewMethodNoRoute tests View when no route is active.
func (suite *RouterTestSuite) TestViewMethodNoRoute() {
	view := suite.router.View()
	assert.Equal(suite.T(), "No route selected", view)
}

// TestInitMethod tests the Init method.
func (suite *RouterTestSuite) TestInitMethod() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	mockView.initCalled = false
	cmd := suite.router.Init()

	assert.Nil(suite.T(), cmd)
	assert.True(suite.T(), mockView.initCalled)
}

// TestUpdateMethod tests the Update method.
func (suite *RouterTestSuite) TestUpdateMethod() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})
	err := suite.router.NavigateTo("/", nil)
	assert.NoError(suite.T(), err)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd := suite.router.Update(msg)

	assert.NotNil(suite.T(), model)
	assert.Nil(suite.T(), cmd)
}

// TestMatchPatternExactMatch tests exact route matching.
func (suite *RouterTestSuite) TestMatchPatternExactMatch() {
	mockView := &MockView{name: "home", viewContent: "Home View"}
	suite.router.AddRoute(Route{Path: "/exact", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	err := suite.router.NavigateTo("/exact", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "/exact", suite.router.GetPath())
}

// TestMatchPatternComplexParams tests complex parameterized routes.
func (suite *RouterTestSuite) TestMatchPatternComplexParams() {
	mockView := &MockView{name: "complex", viewContent: "Complex View"}
	suite.router.AddRoute(Route{Path: "/api/:version/users/:userId/posts/:postId", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView }})

	err := suite.router.NavigateTo("/api/v1/users/100/posts/200", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v1", suite.router.GetParam("version"))
	assert.Equal(suite.T(), "100", suite.router.GetParam("userId"))
	assert.Equal(suite.T(), "200", suite.router.GetParam("postId"))
}

// TestNavigationStackIntegrity tests that navigation stack maintains integrity.
func (suite *RouterTestSuite) TestNavigationStackIntegrity() {
	mockView1 := &MockView{name: "view1", viewContent: "View 1"}
	mockView2 := &MockView{name: "view2", viewContent: "View 2"}
	mockView3 := &MockView{name: "view3", viewContent: "View 3"}

	suite.router.AddRoute(Route{Path: "/view1", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }})
	suite.router.AddRoute(Route{Path: "/view2", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }})
	suite.router.AddRoute(Route{Path: "/view3", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView3 }})

	// Navigate forward
	_ = suite.router.NavigateTo("/view1", nil)
	_ = suite.router.NavigateTo("/view2", nil)
	_ = suite.router.NavigateTo("/view3", nil)

	// Go back twice
	suite.router.Back()
	suite.router.Back()

	assert.Equal(suite.T(), "/view1", suite.router.GetPath())

	// Navigate forward again
	_ = suite.router.NavigateTo("/view3", nil)
	assert.Equal(suite.T(), "/view3", suite.router.GetPath())
	assert.True(suite.T(), suite.router.CanGoBack())
}

// TestParameterPersistenceAcrossNavigation tests that parameters are preserved during navigation.
func (suite *RouterTestSuite) TestParameterPersistenceAcrossNavigation() {
	mockView1 := &MockView{name: "users", viewContent: "Users View"}
	mockView2 := &MockView{name: "posts", viewContent: "Posts View"}

	suite.router.AddRoute(Route{Path: "/users/:id", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView1 }})
	suite.router.AddRoute(Route{Path: "/posts/:postId", Component: func(r Router, sharedMemory storage.SharedMemory) View { return mockView2 }})

	queryParams := map[string]string{"filter": "active"}
	_ = suite.router.NavigateTo("/users/123", queryParams)

	assert.Equal(suite.T(), "123", suite.router.GetParam("id"))
	assert.Equal(suite.T(), "active", suite.router.GetQueryParam("filter"))

	// Navigate to another route
	_ = suite.router.NavigateTo("/posts/456", nil)
	assert.Equal(suite.T(), "456", suite.router.GetParam("postId"))
	assert.Equal(suite.T(), "", suite.router.GetQueryParam("filter")) // Should be cleared

	// Go back
	suite.router.Back()
	assert.Equal(suite.T(), "123", suite.router.GetParam("id"))
	assert.Equal(suite.T(), "active", suite.router.GetQueryParam("filter")) // Should be restored
}

// Run the test suite.
func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
