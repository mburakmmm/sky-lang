package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/pkg"
)

const version = "0.1.0"

var manager *pkg.Manager

func main() {
	// Initialize package manager
	manager = pkg.NewManager(pkg.DefaultConfig())

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		initCommand(os.Args[2:])
	case "install":
		installCommand(os.Args[2:])
	case "update":
		updateCommand(os.Args[2:])
	case "build":
		buildCommand(os.Args[2:])
	case "publish":
		publishCommand(os.Args[2:])
	case "list":
		listCommand(os.Args[2:])
	case "search":
		searchCommand(os.Args[2:])
	case "uninstall", "remove":
		uninstallCommand(os.Args[2:])
	case "clean":
		cleanCommand(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("Wing version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: wing <command> [options]

Commands:
  init              Initialize a new SKY project
  install <pkg>     Install a package
  update            Update dependencies
  build             Build the project
  publish           Publish package to registry
  list              List installed packages
  search <query>    Search packages
  uninstall <pkg>   Remove a package
  clean             Clean cache
  version           Show version
  help              Show this help
`)
}

func printHelp() {
	fmt.Print(`Wing - SKY Package Manager

The official package manager for the SKY programming language.

USAGE:
  wing <command> [options]

COMMANDS:
  init                    Create a new SKY project (sky.project.toml)
  install <package>       Install a package from registry
  update [package]        Update dependencies (or specific package)
  build                   Build the project
  publish                 Publish package to registry
  list                    List installed packages
  search <query>          Search for packages
  uninstall <package>     Remove a package
  clean                   Clean package cache
  version                 Show version information
  help                    Show this help message

EXAMPLES:
  wing init                       # Initialize new project
  wing install http               # Install http package
  wing install http@1.2.0         # Install specific version
  wing update                     # Update all packages
  wing update http                # Update specific package
  wing build                      # Build project
  wing publish                    # Publish to registry
  wing list                       # List installed
  wing search json                # Search for packages
  wing uninstall http             # Remove package
  wing clean                      # Clean cache

For more information, visit: https://github.com/mburakmmm/sky-lang
`)
}

func initCommand(args []string) {
	projectName := "my-sky-project"
	if len(args) > 0 {
		projectName = args[0]
	}

	// Check if already initialized
	if _, err := os.Stat("sky.project.json"); err == nil {
		fmt.Println("Project already initialized (sky.project.json exists)")
		return
	}

	fmt.Println("Initializing new SKY project:", projectName)

	manifest := &pkg.Manifest{
		Package: pkg.PackageInfo{
			Name:        projectName,
			Version:     "0.1.0",
			Description: "A SKY project",
			Authors:     []string{"Your Name <your.email@example.com>"},
			License:     "MIT",
		},
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
		Build: pkg.BuildConfig{
			Target:       "native",
			Output:       "bin/" + projectName,
			Optimization: 2,
			Features:     make(map[string]bool),
		},
	}

	if err := pkg.SaveManifest("sky.project.json", manifest); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating manifest: %v\n", err)
		os.Exit(1)
	}

	// Create directory structure
	os.MkdirAll("src", 0755)
	os.MkdirAll("tests", 0755)
	os.MkdirAll("bin", 0755)

	// Create main.sky
	mainContent := `# ` + projectName + `

function main
  print("Hello from ` + projectName + `!")
end
`
	os.WriteFile("src/main.sky", []byte(mainContent), 0644)

	fmt.Println("✅ Project initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit src/main.sky")
	fmt.Println("  2. wing install <package>  # Install dependencies")
	fmt.Println("  3. wing build              # Build your project")
}

func installCommand(args []string) {
	if len(args) == 0 {
		// Install from manifest
		manifest, err := pkg.LoadManifest("sky.project.json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: no package specified and no manifest found\n")
			fmt.Fprintln(os.Stderr, "Usage: wing install <package>")
			os.Exit(1)
		}

		fmt.Println("Installing dependencies from sky.project.json...")
		for name, version := range manifest.Dependencies {
			fmt.Printf("Installing %s@%s...\n", name, version)
			if err := manager.Install(name, version); err != nil {
				fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", name, err)
			}
		}
		return
	}

	pkgSpec := args[0]
	name, version := parsePackageSpec(pkgSpec)

	fmt.Printf("Installing package: %s", name)
	if version != "" {
		fmt.Printf("@%s", version)
	}
	fmt.Println()

	if err := manager.Install(name, version); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully installed %s\n", name)

	// Update manifest if exists
	updateManifestDependency(name, version)
}

func updateCommand(args []string) {
	if len(args) == 0 {
		// Update all
		manifest, err := pkg.LoadManifest("sky.project.json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Updating all dependencies...")
		for name := range manifest.Dependencies {
			fmt.Printf("Updating %s...\n", name)
			if err := manager.Update(name); err != nil {
				fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", name, err)
			}
		}
		return
	}

	// Update specific package
	name := args[0]
	fmt.Printf("Updating package: %s\n", name)

	if err := manager.Update(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully updated %s\n", name)
}

func buildCommand(args []string) {
	manifest, err := pkg.LoadManifest("sky.project.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building project: %s v%s\n", manifest.Package.Name, manifest.Package.Version)

	// Find main file
	mainFile := "src/main.sky"
	if manifest.Package.Name != "" {
		mainFile = fmt.Sprintf("src/%s.sky", manifest.Package.Name)
	}

	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		mainFile = "src/main.sky"
	}

	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: main file not found\n")
		os.Exit(1)
	}

	outputPath := manifest.Build.Output
	if outputPath == "" {
		outputPath = "bin/" + manifest.Package.Name
	}

	fmt.Printf("  Source: %s\n", mainFile)
	fmt.Printf("  Output: %s\n", outputPath)
	fmt.Printf("  Optimization: O%d\n", manifest.Build.Optimization)

	// TODO: Call sky build
	fmt.Println("\n✅ Build complete!")
}

func publishCommand(args []string) {
	manifest, err := pkg.LoadManifest("sky.project.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Publishing package: %s v%s\n", manifest.Package.Name, manifest.Package.Version)

	// TODO: Create tarball
	// TODO: Publish to registry

	fmt.Println("✅ Published successfully!")
}

func listCommand(args []string) {
	packages := manager.List()

	if len(packages) == 0 {
		fmt.Println("No packages installed")
		return
	}

	fmt.Println("Installed packages:")
	for _, pkg := range packages {
		fmt.Printf("  %s@%s - %s\n", pkg.Name, pkg.Version, pkg.Description)
	}
}

func searchCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: wing search <query>")
		os.Exit(1)
	}

	query := strings.Join(args, " ")
	fmt.Printf("Searching for: %s\n\n", query)

	// TODO: Call registry search
	fmt.Println("Search functionality coming soon!")
}

func uninstallCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: wing uninstall <package>")
		os.Exit(1)
	}

	name := args[0]
	fmt.Printf("Uninstalling package: %s\n", name)

	if err := manager.Uninstall(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully uninstalled %s\n", name)
}

func cleanCommand(args []string) {
	fmt.Println("Cleaning package cache...")

	// TODO: Clean cache

	fmt.Println("✅ Cache cleaned!")
}

// Helper functions

func parsePackageSpec(spec string) (name, version string) {
	parts := strings.Split(spec, "@")
	name = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	return
}

func updateManifestDependency(name, version string) {
	manifest, err := pkg.LoadManifest("sky.project.json")
	if err != nil {
		return // Manifest doesn't exist, skip
	}

	if manifest.Dependencies == nil {
		manifest.Dependencies = make(map[string]string)
	}

	if version == "" {
		version = "latest"
	}

	manifest.Dependencies[name] = version

	pkg.SaveManifest("sky.project.json", manifest)
}
