package main

// RouteDefinition represents a route discovered from the app folder structure.
type RouteDefinition struct {
	// Path is the router path (e.g., "/users/:id")
	Path string
	// FilePath is the absolute path to the page.go file
	FilePath string
	// PackagePath is the Go import path for the page package
	PackagePath string
	// PackageAlias is the alias to use for this package in imports (e.g., "users_id_page")
	PackageAlias string
}
