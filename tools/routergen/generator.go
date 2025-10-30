package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScanAppFolder scans the app folder directory structure and returns route definitions.
// It follows Next.js app folder conventions:
// - page.go files define routes.
// - [param] folders become :param in routes.
// - Nested folders create nested routes.
func ScanAppFolder(rootDir string) ([]RouteDefinition, error) {
	routes := []RouteDefinition{}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a page.go file
		if info.IsDir() || info.Name() != "page.go" {
			return nil
		}

		// Calculate the route path relative to rootDir
		relPath, err := filepath.Rel(rootDir, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Convert filesystem path to route path
		routePath := convertToRoutePath(relPath)

		// Calculate package path and alias
		packagePath := filepath.Dir(path)
		packageAlias := generatePackageAlias(relPath)

		route := RouteDefinition{
			Path:         routePath,
			FilePath:     path,
			PackagePath:  packagePath,
			PackageAlias: packageAlias,
		}

		routes = append(routes, route)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan app folder: %w", err)
	}

	return routes, nil
}

// convertToRoutePath converts a filesystem path to a route path.
// Examples:
//   - "." -> "/"
//   - "users" -> "/users"
//   - "users/_id" -> "/users/:id"
//   - "posts/_postId/comments/_commentId" -> "/posts/:postId/comments/:commentId"
func convertToRoutePath(fsPath string) string {
	if fsPath == "." {
		return "/"
	}

	// Split path into segments
	segments := strings.Split(fsPath, string(os.PathSeparator))

	// Convert _param to :param (underscore prefix indicates dynamic segment)
	for i, segment := range segments {
		if strings.HasPrefix(segment, "_") && len(segment) > 1 {
			segments[i] = ":" + segment[1:]
		}
	}

	// Join with forward slashes and prepend /
	return "/" + strings.Join(segments, "/")
}

// generatePackageAlias creates a unique package alias from the filesystem path.
// Examples:
//   - "." -> "root_page"
//   - "users" -> "users_page"
//   - "users/_id" -> "users_id_page"
//   - "posts/_postId/comments/_commentId" -> "posts_postid_comments_commentid_page"
func generatePackageAlias(fsPath string) string {
	if fsPath == "." {
		return "root_page"
	}

	// Replace path separators with underscores
	alias := strings.ReplaceAll(fsPath, string(os.PathSeparator), "_")

	// Convert to lowercase for consistency
	alias = strings.ToLower(alias)

	// Append _page suffix
	return alias + "_page"
}

// GenerateRoutesFile generates the Go source code for the routes.go file.
func GenerateRoutesFile(routes []RouteDefinition, moduleName string) string {
	var sb strings.Builder

	// Package declaration
	sb.WriteString("package app\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/rxtech-lab/smart-contract-cli/internal/view\"\n")
	sb.WriteString("\t\"github.com/rxtech-lab/smart-contract-cli/internal/storage\"\n")

	// Import each page package (skip root package to avoid import cycle)
	appPackagePath := moduleName + "/app"
	for _, route := range routes {
		// Skip importing the app package itself to avoid circular import
		if route.PackagePath != appPackagePath {
			sb.WriteString(fmt.Sprintf("\t%s \"%s\"\n", route.PackageAlias, route.PackagePath))
		}
	}
	sb.WriteString(")\n\n")

	// GetRoutes function
	sb.WriteString("// GetRoutes returns all routes generated from the app folder structure.\n")
	sb.WriteString("func GetRoutes() []view.Route {\n")
	sb.WriteString("\treturn []view.Route{\n")

	for _, route := range routes {
		// For root package, call NewPage() directly without package prefix
		if route.PackagePath == appPackagePath {
			sb.WriteString(fmt.Sprintf("\t\t{Path: %q, Component: func(r view.Router, sharedMemory storage.SharedMemory) view.View { return NewPage(r) }},\n", route.Path))
		} else {
			sb.WriteString(fmt.Sprintf("\t\t{Path: %q, Component: func(r view.Router, sharedMemory storage.SharedMemory) view.View { return %s.NewPage(r) }},\n",
				route.Path, route.PackageAlias))
		}
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// ConvertAbsoluteToModulePath converts an absolute file path to a module-relative import path.
// Example: /Users/user/project/app/users -> github.com/user/project/app/users.
func ConvertAbsoluteToModulePath(absPath, moduleRoot, moduleName string) (string, error) {
	relPath, err := filepath.Rel(moduleRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	// Convert filesystem path to Go import path (use forward slashes)
	importPath := filepath.ToSlash(relPath)

	return moduleName + "/" + importPath, nil
}
