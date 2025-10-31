package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GeneratorTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *GeneratorTestSuite) SetupTest() {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "routergen-test-*")
	s.Require().NoError(err)
	s.tempDir = tempDir
}

func (s *GeneratorTestSuite) TearDownTest() {
	// Clean up temporary directory
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *GeneratorTestSuite) createPageFile(relPath string) string {
	fullPath := filepath.Join(s.tempDir, relPath, "page.go")
	dir := filepath.Dir(fullPath)

	err := os.MkdirAll(dir, 0755)
	s.Require().NoError(err)

	err = os.WriteFile(fullPath, []byte("package page\n"), 0644)
	s.Require().NoError(err)

	return fullPath
}

func (s *GeneratorTestSuite) TestConvertToRoutePath_RootPath() {
	result := convertToRoutePath(".")
	s.Equal("/", result)
}

func (s *GeneratorTestSuite) TestConvertToRoutePath_SimplePath() {
	result := convertToRoutePath("users")
	s.Equal("/users", result)
}

func (s *GeneratorTestSuite) TestConvertToRoutePath_SingleDynamicSegment() {
	result := convertToRoutePath("users/_id")
	s.Equal("/users/:id", result)
}

func (s *GeneratorTestSuite) TestConvertToRoutePath_MultipleDynamicSegments() {
	result := convertToRoutePath("posts/_postId/comments/_commentId")
	s.Equal("/posts/:postId/comments/:commentId", result)
}

func (s *GeneratorTestSuite) TestConvertToRoutePath_NestedStaticPaths() {
	result := convertToRoutePath("admin/settings/profile")
	s.Equal("/admin/settings/profile", result)
}

func (s *GeneratorTestSuite) TestConvertToRoutePath_MixedStaticAndDynamic() {
	result := convertToRoutePath("api/users/_userId/posts")
	s.Equal("/api/users/:userId/posts", result)
}

func (s *GeneratorTestSuite) TestGeneratePackageAlias_RootPath() {
	result := generatePackageAlias(".")
	s.Equal("root_page", result)
}

func (s *GeneratorTestSuite) TestGeneratePackageAlias_SimplePath() {
	result := generatePackageAlias("users")
	s.Equal("users_page", result)
}

func (s *GeneratorTestSuite) TestGeneratePackageAlias_DynamicSegment() {
	result := generatePackageAlias("users/_id")
	s.Equal("users__id_page", result)
}

func (s *GeneratorTestSuite) TestGeneratePackageAlias_MultipleDynamicSegments() {
	result := generatePackageAlias("posts/_postId/comments/_commentId")
	s.Equal("posts__postid_comments__commentid_page", result)
}

func (s *GeneratorTestSuite) TestGeneratePackageAlias_ComplexNested() {
	result := generatePackageAlias("admin/users/_userId/settings")
	s.Equal("admin_users__userid_settings_page", result)
}

func (s *GeneratorTestSuite) TestScanAppFolder_EmptyDirectory() {
	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Empty(routes)
}

func (s *GeneratorTestSuite) TestScanAppFolder_RootPageOnly() {
	s.createPageFile(".")

	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Len(routes, 1)
	s.Equal("/", routes[0].Path)
	s.Equal("root_page", routes[0].PackageAlias)
}

func (s *GeneratorTestSuite) TestScanAppFolder_SimpleNestedPages() {
	s.createPageFile(".")
	s.createPageFile("users")
	s.createPageFile("posts")

	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Len(routes, 3)

	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}

	s.True(paths["/"])
	s.True(paths["/users"])
	s.True(paths["/posts"])
}

func (s *GeneratorTestSuite) TestScanAppFolder_DynamicSegments() {
	s.createPageFile(".")
	s.createPageFile("users")
	s.createPageFile("users/_id")

	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Len(routes, 3)

	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}

	s.True(paths["/"])
	s.True(paths["/users"])
	s.True(paths["/users/:id"])
}

func (s *GeneratorTestSuite) TestScanAppFolder_ComplexNestedStructure() {
	s.createPageFile(".")
	s.createPageFile("users")
	s.createPageFile("users/_id")
	s.createPageFile("posts")
	s.createPageFile("posts/_postId")
	s.createPageFile("posts/_postId/comments")
	s.createPageFile("posts/_postId/comments/_commentId")

	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Len(routes, 7)

	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}

	s.True(paths["/"])
	s.True(paths["/users"])
	s.True(paths["/users/:id"])
	s.True(paths["/posts"])
	s.True(paths["/posts/:postId"])
	s.True(paths["/posts/:postId/comments"])
	s.True(paths["/posts/:postId/comments/:commentId"])
}

func (s *GeneratorTestSuite) TestScanAppFolder_IgnoresNonPageFiles() {
	s.createPageFile(".")
	s.createPageFile("users")

	// Create non-page.go files that should be ignored
	otherFile := filepath.Join(s.tempDir, "users", "helper.go")
	err := os.WriteFile(otherFile, []byte("package users\n"), 0644)
	s.Require().NoError(err)

	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Len(routes, 2) // Only page.go files should be counted
}

func (s *GeneratorTestSuite) TestScanAppFolder_InvalidDirectory() {
	routes, err := ScanAppFolder("/non/existent/directory")
	s.Error(err)
	s.Nil(routes)
}

func (s *GeneratorTestSuite) TestScanAppFolder_PackagePathsAreCorrect() {
	s.createPageFile(".")
	usersPagePath := s.createPageFile("users")

	routes, err := ScanAppFolder(s.tempDir)
	s.NoError(err)
	s.Len(routes, 2)

	// Find the users route
	var usersRoute *RouteDefinition
	for i := range routes {
		if routes[i].Path == "/users" {
			usersRoute = &routes[i]
			break
		}
	}

	s.NotNil(usersRoute)
	s.Equal(usersPagePath, usersRoute.FilePath)
	s.Equal(filepath.Dir(usersPagePath), usersRoute.PackagePath)
}

func (s *GeneratorTestSuite) TestGenerateRoutesFile_EmptyRoutes() {
	code := GenerateRoutesFile([]RouteDefinition{}, "github.com/test/project")
	s.Contains(code, "package app")
	s.Contains(code, "func GetRoutes() []view.Route")
	s.Contains(code, "return []view.Route{")
}

func (s *GeneratorTestSuite) TestGenerateRoutesFile_SingleRoute() {
	routes := []RouteDefinition{
		{
			Path:         "/",
			FilePath:     "/test/app/page.go",
			PackagePath:  "github.com/test/project/app",
			PackageAlias: "root_page",
		},
	}

	code := GenerateRoutesFile(routes, "github.com/test/project")

	s.Contains(code, "package app")
	// Root package should not be imported (to avoid circular import)
	s.NotContains(code, `root_page "github.com/test/project/app"`)
	// Root route should call NewPage() directly without package prefix
	s.Contains(code, `{Path: "/", Component: func(r view.Router, sharedMemory storage.SharedMemory) view.View { return NewPage(r, sharedMemory) }}`)
}

func (s *GeneratorTestSuite) TestGenerateRoutesFile_MultipleRoutes() {
	routes := []RouteDefinition{
		{
			Path:         "/",
			PackagePath:  "github.com/test/project/app",
			PackageAlias: "root_page",
		},
		{
			Path:         "/users",
			PackagePath:  "github.com/test/project/app/users",
			PackageAlias: "users_page",
		},
		{
			Path:         "/users/:id",
			PackagePath:  "github.com/test/project/app/users/id",
			PackageAlias: "users_id_page",
		},
	}

	code := GenerateRoutesFile(routes, "github.com/test/project")

	// Check package declaration
	s.Contains(code, "package app")

	// Check imports - root package should not be imported to avoid circular import
	s.NotContains(code, `root_page "github.com/test/project/app"`)
	s.Contains(code, `users_page "github.com/test/project/app/users"`)
	s.Contains(code, `users_id_page "github.com/test/project/app/users/id"`)

	// Check route definitions - root route calls NewPage() directly
	s.Contains(code, `{Path: "/", Component: func(r view.Router, sharedMemory storage.SharedMemory) view.View { return NewPage(r, sharedMemory) }}`)
	s.Contains(code, `{Path: "/users", Component: func(r view.Router, sharedMemory storage.SharedMemory) view.View { return users_page.NewPage(r, sharedMemory) }}`)
	s.Contains(code, `{Path: "/users/:id", Component: func(r view.Router, sharedMemory storage.SharedMemory) view.View { return users_id_page.NewPage(r, sharedMemory) }}`)
}

func (s *GeneratorTestSuite) TestGenerateRoutesFile_HasCorrectStructure() {
	routes := []RouteDefinition{
		{
			Path:         "/users",
			PackagePath:  "github.com/test/project/app/users",
			PackageAlias: "users_page",
		},
	}

	code := GenerateRoutesFile(routes, "github.com/test/project")

	// Verify basic structure
	s.Contains(code, "package app\n\nimport (")
	s.Contains(code, `"github.com/rxtech-lab/smart-contract-cli/internal/view"`)
	s.Contains(code, "func GetRoutes() []view.Route {")
	s.Contains(code, "return []view.Route{")
}

func (s *GeneratorTestSuite) TestConvertAbsoluteToModulePath_Success() {
	moduleRoot := "/Users/test/project"
	moduleName := "github.com/test/project"
	absPath := "/Users/test/project/app/users"

	result, err := ConvertAbsoluteToModulePath(absPath, moduleRoot, moduleName)
	s.NoError(err)
	s.Equal("github.com/test/project/app/users", result)
}

func (s *GeneratorTestSuite) TestConvertAbsoluteToModulePath_NestedPath() {
	moduleRoot := "/Users/test/project"
	moduleName := "github.com/test/project"
	absPath := "/Users/test/project/app/users/settings/profile"

	result, err := ConvertAbsoluteToModulePath(absPath, moduleRoot, moduleName)
	s.NoError(err)
	s.Equal("github.com/test/project/app/users/settings/profile", result)
}

func (s *GeneratorTestSuite) TestConvertAbsoluteToModulePath_RootPath() {
	moduleRoot := "/Users/test/project"
	moduleName := "github.com/test/project"
	absPath := "/Users/test/project/app"

	result, err := ConvertAbsoluteToModulePath(absPath, moduleRoot, moduleName)
	s.NoError(err)
	s.Equal("github.com/test/project/app", result)
}

func TestGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(GeneratorTestSuite))
}
