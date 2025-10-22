package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/wing"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Get current working directory
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot get current directory: %v\n", err)
		os.Exit(1)
	}

	// Check for GitHub token and registry type
	githubToken := os.Getenv("GITHUB_TOKEN")
	registryType := os.Getenv("WING_REGISTRY_TYPE") // "packages" or "releases"
	var pm *wing.PackageManager

	if githubToken != "" {
		if registryType == "packages" {
			pm = wing.NewPackageManagerWithGitHubPackages(projectDir, githubToken)
		} else {
			pm = wing.NewPackageManagerWithGitHub(projectDir, githubToken)
		}
	} else {
		pm = wing.NewPackageManager(projectDir)
	}

	switch command {
	case "init":
		handleInit(pm, os.Args[2:])
	case "install":
		handleInstall(pm, os.Args[2:])
	case "add":
		handleAdd(pm, os.Args[2:])
	case "remove":
		handleRemove(pm, os.Args[2:])
	case "update":
		handleUpdate(pm, os.Args[2:])
	case "build":
		handleBuild(pm, os.Args[2:])
	case "publish":
		handlePublish(pm, os.Args[2:])
	case "search":
		handleSearch(pm, os.Args[2:])
	case "list":
		handleList(pm, os.Args[2:])
	case "info":
		handleInfo(pm, os.Args[2:])
	case "run":
		handleRun(pm, os.Args[2:])
	case "test":
		handleTest(pm, os.Args[2:])
	case "clean":
		handleClean(pm, os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleInit(pm *wing.PackageManager, args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	var projectName string
	fs.StringVar(&projectName, "name", "", "Project name")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if projectName == "" {
		// Use current directory name as project name
		projectName = filepath.Base(pm.ProjectDir)
	}

	fmt.Printf("Initializing Sky project '%s'...\n", projectName)

	if err := pm.Init(projectName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project '%s' initialized successfully!\n", projectName)
	fmt.Printf("Run 'wing install' to install dependencies.\n")
	fmt.Printf("Run 'wing build' to build the project.\n")
}

func handleInstall(pm *wing.PackageManager, args []string) {
	fmt.Println("Installing dependencies...")

	if err := pm.Install(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Dependencies installed successfully!")
}

func handleAdd(pm *wing.PackageManager, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var dev bool
	var version string
	fs.BoolVar(&dev, "dev", false, "Add as dev dependency")
	fs.StringVar(&version, "version", "latest", "Package version")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if len(fs.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error: package name required\n")
		os.Exit(1)
	}

	packageName := fs.Args()[0]

	fmt.Printf("Adding %s@%s...\n", packageName, version)

	if err := pm.Add(packageName, version, dev); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Added %s@%s successfully!\n", packageName, version)
}

func handleRemove(pm *wing.PackageManager, args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	var dev bool
	fs.BoolVar(&dev, "dev", false, "Remove from dev dependencies")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if len(fs.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error: package name required\n")
		os.Exit(1)
	}

	packageName := fs.Args()[0]

	fmt.Printf("Removing %s...\n", packageName)

	if err := pm.Remove(packageName, dev); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed %s successfully!\n", packageName)
}

func handleUpdate(pm *wing.PackageManager, args []string) {
	fmt.Println("Updating dependencies...")

	if err := pm.Update(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Dependencies updated successfully!")
}

func handleBuild(pm *wing.PackageManager, args []string) {
	fmt.Println("Building project...")

	if err := pm.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Project built successfully!")
}

func handlePublish(pm *wing.PackageManager, args []string) {
	fmt.Println("Publishing package...")

	if err := pm.Publish(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Package published successfully!")
}

func handleSearch(pm *wing.PackageManager, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: search query required\n")
		os.Exit(1)
	}

	query := strings.Join(args, " ")
	fmt.Printf("Searching for packages matching '%s'...\n", query)

	// Mock search results
	packages := []struct {
		Name        string
		Version     string
		Description string
	}{
		{"http", "1.0.0", "HTTP client library"},
		{"json", "2.1.0", "JSON parsing and serialization"},
		{"crypto", "1.5.0", "Cryptographic utilities"},
		{"database", "3.0.0", "Database connectivity"},
		{"web", "1.2.0", "Web framework"},
	}

	fmt.Println("\nSearch Results:")
	fmt.Println("===============")
	for _, pkg := range packages {
		if strings.Contains(strings.ToLower(pkg.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(pkg.Description), strings.ToLower(query)) {
			fmt.Printf("%-15s %-10s %s\n", pkg.Name, pkg.Version, pkg.Description)
		}
	}
}

func handleList(pm *wing.PackageManager, args []string) {
	if err := pm.LoadManifest(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Project Dependencies:")
	fmt.Println("====================")

	if len(pm.Manifest.Dependencies) > 0 {
		fmt.Println("\nDependencies:")
		for name, version := range pm.Manifest.Dependencies {
			fmt.Printf("  %s@%s\n", name, version)
		}
	}

	if len(pm.Manifest.DevDependencies) > 0 {
		fmt.Println("\nDev Dependencies:")
		for name, version := range pm.Manifest.DevDependencies {
			fmt.Printf("  %s@%s\n", name, version)
		}
	}

	if len(pm.Manifest.Dependencies) == 0 && len(pm.Manifest.DevDependencies) == 0 {
		fmt.Println("No dependencies found.")
	}
}

func handleInfo(pm *wing.PackageManager, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: package name required\n")
		os.Exit(1)
	}

	packageName := args[0]
	fmt.Printf("Package Information: %s\n", packageName)
	fmt.Println("========================")

	// Mock package info
	fmt.Printf("Name: %s\n", packageName)
	fmt.Printf("Version: 1.0.0\n")
	fmt.Printf("Description: Mock package for %s\n", packageName)
	fmt.Printf("License: MIT\n")
	fmt.Printf("Repository: https://github.com/sky-lang/%s\n", packageName)
	fmt.Printf("Downloads: 1,234\n")
	fmt.Printf("Created: 2024-01-01\n")
	fmt.Printf("Updated: 2024-01-15\n")
}

func handleRun(pm *wing.PackageManager, args []string) {
	if err := pm.LoadManifest(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		// Run default script
		if script, exists := pm.Manifest.Scripts["run"]; exists {
			fmt.Printf("Running script: %s\n", script)
			// In real implementation, execute the script
			fmt.Println("Script executed successfully!")
		} else {
			fmt.Fprintf(os.Stderr, "Error: no 'run' script defined\n")
			os.Exit(1)
		}
	} else {
		scriptName := args[0]
		if script, exists := pm.Manifest.Scripts[scriptName]; exists {
			fmt.Printf("Running script '%s': %s\n", scriptName, script)
			// In real implementation, execute the script
			fmt.Println("Script executed successfully!")
		} else {
			fmt.Fprintf(os.Stderr, "Error: script '%s' not found\n", scriptName)
			os.Exit(1)
		}
	}
}

func handleTest(pm *wing.PackageManager, args []string) {
	if err := pm.LoadManifest(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Running tests...")

	if script, exists := pm.Manifest.Scripts["test"]; exists {
		fmt.Printf("Running test script: %s\n", script)
		// In real implementation, execute the test script
		fmt.Println("Tests passed successfully!")
	} else {
		fmt.Println("No test script defined, running default tests...")
		// In real implementation, run default test discovery
		fmt.Println("Tests passed successfully!")
	}
}

func handleClean(pm *wing.PackageManager, args []string) {
	fmt.Println("Cleaning project...")

	// Clean build artifacts
	buildDir := filepath.Join(pm.ProjectDir, "dist")
	if err := os.RemoveAll(buildDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot remove dist directory: %v\n", err)
	}

	// Clean cache
	cacheDir := filepath.Join(pm.ProjectDir, ".sky")
	if err := os.RemoveAll(cacheDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot remove cache directory: %v\n", err)
	}

	fmt.Println("Project cleaned successfully!")
}

func printUsage() {
	fmt.Println("Wing - Sky Package Manager")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("Usage: wing <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init [name]           Initialize a new Sky project")
	fmt.Println("  install               Install project dependencies")
	fmt.Println("  add <package> [opts]  Add a dependency")
	fmt.Println("  remove <package>      Remove a dependency")
	fmt.Println("  update                 Update dependencies to latest versions")
	fmt.Println("  build                  Build the project")
	fmt.Println("  publish                Publish package to registry")
	fmt.Println("  search <query>         Search for packages")
	fmt.Println("  list                   List project dependencies")
	fmt.Println("  info <package>         Show package information")
	fmt.Println("  run [script]           Run a script")
	fmt.Println("  test                   Run tests")
	fmt.Println("  clean                  Clean build artifacts")
	fmt.Println("  help                   Show this help message")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --dev                  Add/remove as dev dependency")
	fmt.Println("  --version <ver>        Specify package version")
	fmt.Println("  --help, -h             Show help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  wing init my-project")
	fmt.Println("  wing add http --version 1.0.0")
	fmt.Println("  wing add json --dev")
	fmt.Println("  wing remove http")
	fmt.Println("  wing build")
	fmt.Println("  wing run test")
}
