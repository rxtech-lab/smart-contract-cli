package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	routes, err := scanAndProcessRoutes(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := generateAndWriteRoutes(routes, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	printSuccess(config.outputFile, routes)
}

// config holds the configuration parsed from CLI flags.
type config struct {
	appDir     string
	moduleRoot string
	outputFile string
	moduleName string
}

// parseFlags parses CLI flags and returns configuration.
func parseFlags() (*config, error) {
	appDirFlag := flag.String("dir", "./app", "Path to the app folder directory")
	outputFlag := flag.String("output", "", "Output file path (default: <app-dir>/routes.go)")
	moduleNameFlag := flag.String("module", "", "Go module name (auto-detected from go.mod if not provided)")
	moduleRootFlag := flag.String("module-root", ".", "Path to the module root directory")

	flag.Parse()

	appDir, err := filepath.Abs(*appDirFlag)
	if err != nil {
		return nil, fmt.Errorf("resolving app directory path: %w", err)
	}

	moduleRoot, err := filepath.Abs(*moduleRootFlag)
	if err != nil {
		return nil, fmt.Errorf("resolving module root path: %w", err)
	}

	outputFile := *outputFlag
	if outputFile == "" {
		outputFile = filepath.Join(appDir, "routes_gen.go")
	}

	moduleName := *moduleNameFlag
	if moduleName == "" {
		moduleName, err = detectModuleName(moduleRoot)
		if err != nil {
			return nil, fmt.Errorf("detecting module name: %w. Please provide module name using -module flag", err)
		}
	}

	return &config{
		appDir:     appDir,
		moduleRoot: moduleRoot,
		outputFile: outputFile,
		moduleName: moduleName,
	}, nil
}

// scanAndProcessRoutes scans the app folder and converts paths to import paths.
func scanAndProcessRoutes(cfg *config) ([]RouteDefinition, error) {
	fmt.Printf("Scanning app folder: %s\n", cfg.appDir)
	routes, err := ScanAppFolder(cfg.appDir)
	if err != nil {
		return nil, fmt.Errorf("scanning app folder: %w", err)
	}

	if len(routes) == 0 {
		fmt.Println("No routes found. Make sure your app folder contains page.go files.")
		os.Exit(0)
	}

	fmt.Printf("Found %d route(s)\n", len(routes))

	for i := range routes {
		importPath, err := ConvertAbsoluteToModulePath(routes[i].PackagePath, cfg.moduleRoot, cfg.moduleName)
		if err != nil {
			return nil, fmt.Errorf("converting path to import: %w", err)
		}
		routes[i].PackagePath = importPath
	}

	return routes, nil
}

// generateAndWriteRoutes generates the routes file and writes it to disk.
func generateAndWriteRoutes(routes []RouteDefinition, cfg *config) error {
	code := GenerateRoutesFile(routes, cfg.moduleName)

	if err := os.WriteFile(cfg.outputFile, []byte(code), 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}

// printSuccess prints success message with route details.
func printSuccess(outputFile string, routes []RouteDefinition) {
	fmt.Printf("Successfully generated routes file: %s\n", outputFile)
	fmt.Println("\nGenerated routes:")
	for _, route := range routes {
		fmt.Printf("  %s -> %s\n", route.Path, route.PackageAlias)
	}
}

// detectModuleName reads the go.mod file to determine the module name.
func detectModuleName(moduleRoot string) (string, error) {
	goModPath := filepath.Join(moduleRoot, "go.mod")

	// Validate path to prevent directory traversal
	cleaned := filepath.Clean(goModPath)
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("invalid file path: %s", goModPath)
	}

	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Parse the first line which should be "module <name>"
	lines := []byte{}
	for i, b := range data {
		if b == '\n' {
			lines = data[:i]
			break
		}
	}

	if len(lines) == 0 {
		return "", fmt.Errorf("go.mod is empty or invalid")
	}

	line := string(lines)
	if len(line) < 7 || line[:7] != "module " {
		return "", fmt.Errorf("go.mod does not start with 'module' declaration")
	}

	moduleName := line[7:]
	return moduleName, nil
}
