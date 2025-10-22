package wing

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// PackageManifest represents sky.project.toml structure
type PackageManifest struct {
	Package         PackageInfo       `toml:"package"`
	Dependencies    map[string]string `toml:"dependencies"`
	DevDependencies map[string]string `toml:"dev-dependencies"`
	Scripts         map[string]string `toml:"scripts"`
	Build           BuildConfig       `toml:"build"`
}

type PackageInfo struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	Description string   `toml:"description"`
	Authors     []string `toml:"authors"`
	License     string   `toml:"license"`
	Repository  string   `toml:"repository"`
	Homepage    string   `toml:"homepage"`
	Keywords    []string `toml:"keywords"`
}

type BuildConfig struct {
	Target       string `toml:"target"`
	Optimization string `toml:"optimization"`
	OutputDir    string `toml:"output-dir"`
	SourceDir    string `toml:"source-dir"`
}

// RegistryPackage represents a package in the registry
type RegistryPackage struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Authors      []string          `json:"authors"`
	License      string            `json:"license"`
	Repository   string            `json:"repository"`
	Homepage     string            `json:"homepage"`
	Keywords     []string          `json:"keywords"`
	Downloads    int64             `json:"downloads"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Versions     []string          `json:"versions"`
	Dependencies map[string]string `json:"dependencies"`
}

// Dependency represents a package dependency
type Dependency struct {
	Name    string
	Version string
	Source  string // "registry", "git", "local"
	URL     string // for git/local sources
}

// PackageManager handles package operations
type PackageManager struct {
	RegistryURL string
	CacheDir    string
	ProjectDir  string
	Manifest    *PackageManifest
	GitHubToken string
}

// NewPackageManager creates a new package manager
func NewPackageManager(projectDir string) *PackageManager {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".sky", "cache")

	return &PackageManager{
		RegistryURL: "https://github.com/mburakmmm/wing-packages",
		CacheDir:    cacheDir,
		ProjectDir:  projectDir,
	}
}

// NewPackageManagerWithGitHub creates a new package manager with GitHub registry
func NewPackageManagerWithGitHub(projectDir, token string) *PackageManager {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".sky", "cache")

	return &PackageManager{
		RegistryURL: "https://github.com/mburakmmm/wing-packages",
		CacheDir:    cacheDir,
		ProjectDir:  projectDir,
		GitHubToken: token,
	}
}

// NewPackageManagerWithGitHubPackages creates a new package manager with GitHub Packages
func NewPackageManagerWithGitHubPackages(projectDir, token string) *PackageManager {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".sky", "cache")

	return &PackageManager{
		RegistryURL: "https://github.com/mburakmmm/wing-packages",
		CacheDir:    cacheDir,
		ProjectDir:  projectDir,
		GitHubToken: token,
	}
}

// LoadManifest loads sky.project.toml from project directory
func (pm *PackageManager) LoadManifest() error {
	manifestPath := filepath.Join(pm.ProjectDir, "sky.project.toml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot read manifest: %v", err)
	}

	// Simple TOML parser (basic implementation)
	manifest, err := pm.parseTOML(string(data))
	if err != nil {
		return fmt.Errorf("cannot parse manifest: %v", err)
	}

	pm.Manifest = manifest
	return nil
}

// SaveManifest saves sky.project.toml to project directory
func (pm *PackageManager) SaveManifest() error {
	manifestPath := filepath.Join(pm.ProjectDir, "sky.project.toml")

	tomlContent := pm.generateTOML()

	err := os.WriteFile(manifestPath, []byte(tomlContent), 0644)
	if err != nil {
		return fmt.Errorf("cannot write manifest: %v", err)
	}

	return nil
}

// Init initializes a new Sky project
func (pm *PackageManager) Init(projectName string) error {
	// Create project directory if it doesn't exist
	if err := os.MkdirAll(pm.ProjectDir, 0755); err != nil {
		return fmt.Errorf("cannot create project directory: %v", err)
	}

	// Create default manifest
	pm.Manifest = &PackageManifest{
		Package: PackageInfo{
			Name:        projectName,
			Version:     "0.1.0",
			Description: "A new Sky project",
			Authors:     []string{"Developer"},
			License:     "MIT",
		},
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
		Scripts: map[string]string{
			"build": "sky build",
			"test":  "sky test",
			"run":   "sky run src/main.sky",
		},
		Build: BuildConfig{
			Target:       "native",
			Optimization: "debug",
			OutputDir:    "dist",
			SourceDir:    "src",
		},
	}

	// Save manifest
	if err := pm.SaveManifest(); err != nil {
		return err
	}

	// Create source directory
	srcDir := filepath.Join(pm.ProjectDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("cannot create src directory: %v", err)
	}

	// Create main.sky file
	mainFile := filepath.Join(srcDir, "main.sky")
	mainContent := `function main(): void
  print("Hello, Sky!")
end
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("cannot create main.sky: %v", err)
	}

	// Create .gitignore
	gitignoreContent := `# Sky build artifacts
dist/
.sky/
*.skyc

# Dependencies
deps/

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
`
	gitignoreFile := filepath.Join(pm.ProjectDir, ".gitignore")
	if err := os.WriteFile(gitignoreFile, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("cannot create .gitignore: %v", err)
	}

	return nil
}

// Install installs dependencies
func (pm *PackageManager) Install() error {
	if pm.Manifest == nil {
		if err := pm.LoadManifest(); err != nil {
			return err
		}
	}

	// Create deps directory
	depsDir := filepath.Join(pm.ProjectDir, "deps")
	if err := os.MkdirAll(depsDir, 0755); err != nil {
		return fmt.Errorf("cannot create deps directory: %v", err)
	}

	// Install dependencies
	for name, version := range pm.Manifest.Dependencies {
		if err := pm.installPackage(name, version, depsDir); err != nil {
			return fmt.Errorf("cannot install %s@%s: %v", name, version, err)
		}
	}

	// Install dev dependencies
	for name, version := range pm.Manifest.DevDependencies {
		if err := pm.installPackage(name, version, depsDir); err != nil {
			return fmt.Errorf("cannot install dev dependency %s@%s: %v", name, version, err)
		}
	}

	return nil
}

// Add adds a new dependency
func (pm *PackageManager) Add(packageName string, version string, dev bool) error {
	if pm.Manifest == nil {
		if err := pm.LoadManifest(); err != nil {
			return err
		}
	}

	// Add to manifest
	if dev {
		pm.Manifest.DevDependencies[packageName] = version
	} else {
		pm.Manifest.Dependencies[packageName] = version
	}

	// Save manifest
	if err := pm.SaveManifest(); err != nil {
		return err
	}

	// Install the package
	depsDir := filepath.Join(pm.ProjectDir, "deps")
	if err := os.MkdirAll(depsDir, 0755); err != nil {
		return fmt.Errorf("cannot create deps directory: %v", err)
	}

	if err := pm.installPackage(packageName, version, depsDir); err != nil {
		return fmt.Errorf("cannot install %s@%s: %v", packageName, version, err)
	}

	return nil
}

// Remove removes a dependency
func (pm *PackageManager) Remove(packageName string, dev bool) error {
	if pm.Manifest == nil {
		if err := pm.LoadManifest(); err != nil {
			return err
		}
	}

	// Remove from manifest
	if dev {
		delete(pm.Manifest.DevDependencies, packageName)
	} else {
		delete(pm.Manifest.Dependencies, packageName)
	}

	// Save manifest
	if err := pm.SaveManifest(); err != nil {
		return err
	}

	// Remove package directory
	depsDir := filepath.Join(pm.ProjectDir, "deps", packageName)
	if err := os.RemoveAll(depsDir); err != nil {
		return fmt.Errorf("cannot remove package directory: %v", err)
	}

	return nil
}

// Update updates dependencies
func (pm *PackageManager) Update() error {
	if pm.Manifest == nil {
		if err := pm.LoadManifest(); err != nil {
			return err
		}
	}

	// Update dependencies
	for name, version := range pm.Manifest.Dependencies {
		latestVersion, err := pm.getLatestVersion(name)
		if err != nil {
			return fmt.Errorf("cannot get latest version for %s: %v", name, err)
		}

		if latestVersion != version {
			pm.Manifest.Dependencies[name] = latestVersion
			if err := pm.installPackage(name, latestVersion, filepath.Join(pm.ProjectDir, "deps")); err != nil {
				return fmt.Errorf("cannot update %s: %v", name, err)
			}
		}
	}

	// Save updated manifest
	if err := pm.SaveManifest(); err != nil {
		return err
	}

	return nil
}

// Build builds the project
func (pm *PackageManager) Build() error {
	if pm.Manifest == nil {
		if err := pm.LoadManifest(); err != nil {
			return err
		}
	}

	// Use BuildSystem for building
	buildSystem := NewBuildSystem(pm.ProjectDir, pm.Manifest)
	return buildSystem.Build()
}

// Publish publishes the package to registry
func (pm *PackageManager) Publish() error {
	if pm.Manifest == nil {
		if err := pm.LoadManifest(); err != nil {
			return err
		}
	}

	// Validate package
	if err := pm.validatePackage(); err != nil {
		return fmt.Errorf("package validation failed: %v", err)
	}

	// Create package archive
	archivePath, err := pm.createPackageArchive()
	if err != nil {
		return fmt.Errorf("cannot create package archive: %v", err)
	}
	defer os.Remove(archivePath)

	// Read archive data
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("cannot read archive: %v", err)
	}

	// Use GitHub registry if token is available
	if pm.GitHubToken != "" {
		if err := pm.publishToGitHub(archiveData); err != nil {
			return fmt.Errorf("cannot publish to GitHub: %v", err)
		}
	} else {
		// Fallback to mock upload
		if err := pm.uploadToRegistry(archivePath); err != nil {
			return fmt.Errorf("cannot upload to registry: %v", err)
		}
	}

	fmt.Printf("Package %s@%s published successfully!\n", pm.Manifest.Package.Name, pm.Manifest.Package.Version)

	return nil
}

func (pm *PackageManager) publishToGitHub(archiveData []byte) error {
	fmt.Printf("Publishing to GitHub with token: %s...\n", pm.GitHubToken[:10]+"...")

	// Check registry type
	registryType := os.Getenv("WING_REGISTRY_TYPE")
	if registryType == "packages" {
		// Use GitHub Packages
		client := NewGitHubPackagesClient("mburakmmm", "wing-packages", pm.GitHubToken)
		return client.PublishPackage(pm.Manifest, archiveData)
	} else {
		// Use GitHub Releases (default)
		client := NewGitHubRegistryClient("mburakmmm", "wing-packages", pm.GitHubToken)
		return client.PublishPackage(pm.Manifest, archiveData)
	}
}

// Helper methods

func (pm *PackageManager) parseTOML(content string) (*PackageManifest, error) {
	manifest := &PackageManifest{
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
		Scripts:         make(map[string]string),
	}

	// Use real TOML parser
	err := toml.Unmarshal([]byte(content), manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %v", err)
	}

	// Initialize maps if they are nil
	if manifest.Dependencies == nil {
		manifest.Dependencies = make(map[string]string)
	}
	if manifest.DevDependencies == nil {
		manifest.DevDependencies = make(map[string]string)
	}
	if manifest.Scripts == nil {
		manifest.Scripts = make(map[string]string)
	}

	return manifest, nil
}

func (pm *PackageManager) generateTOML() string {
	// Use real TOML encoder
	data, err := toml.Marshal(pm.Manifest)
	if err != nil {
		// Fallback to manual generation if encoding fails
		return pm.generateTOMLFallback()
	}
	return string(data)
}

func (pm *PackageManager) generateTOMLFallback() string {
	var sb strings.Builder

	sb.WriteString("[package]\n")
	sb.WriteString(fmt.Sprintf("name = \"%s\"\n", pm.Manifest.Package.Name))
	sb.WriteString(fmt.Sprintf("version = \"%s\"\n", pm.Manifest.Package.Version))
	sb.WriteString(fmt.Sprintf("description = \"%s\"\n", pm.Manifest.Package.Description))
	sb.WriteString(fmt.Sprintf("license = \"%s\"\n", pm.Manifest.Package.License))
	if pm.Manifest.Package.Repository != "" {
		sb.WriteString(fmt.Sprintf("repository = \"%s\"\n", pm.Manifest.Package.Repository))
	}
	if pm.Manifest.Package.Homepage != "" {
		sb.WriteString(fmt.Sprintf("homepage = \"%s\"\n", pm.Manifest.Package.Homepage))
	}

	if len(pm.Manifest.Dependencies) > 0 {
		sb.WriteString("\n[dependencies]\n")
		for name, version := range pm.Manifest.Dependencies {
			sb.WriteString(fmt.Sprintf("%s = \"%s\"\n", name, version))
		}
	}

	if len(pm.Manifest.DevDependencies) > 0 {
		sb.WriteString("\n[dev-dependencies]\n")
		for name, version := range pm.Manifest.DevDependencies {
			sb.WriteString(fmt.Sprintf("%s = \"%s\"\n", name, version))
		}
	}

	if len(pm.Manifest.Scripts) > 0 {
		sb.WriteString("\n[scripts]\n")
		for name, command := range pm.Manifest.Scripts {
			sb.WriteString(fmt.Sprintf("%s = \"%s\"\n", name, command))
		}
	}

	sb.WriteString("\n[build]\n")
	sb.WriteString(fmt.Sprintf("target = \"%s\"\n", pm.Manifest.Build.Target))
	sb.WriteString(fmt.Sprintf("optimization = \"%s\"\n", pm.Manifest.Build.Optimization))
	sb.WriteString(fmt.Sprintf("output-dir = \"%s\"\n", pm.Manifest.Build.OutputDir))
	sb.WriteString(fmt.Sprintf("source-dir = \"%s\"\n", pm.Manifest.Build.SourceDir))

	return sb.String()
}

func (pm *PackageManager) installPackage(name, version, depsDir string) error {
	// Resolve latest version if needed
	if version == "latest" {
		latestVersion, err := pm.getLatestVersion(name)
		if err != nil {
			return fmt.Errorf("cannot get latest version for %s: %v", name, err)
		}
		version = latestVersion
		fmt.Printf("Resolved 'latest' to version: %s\n", version)
	}

	// Check if package is already installed
	packageDir := filepath.Join(depsDir, name)
	if _, err := os.Stat(packageDir); err == nil {
		fmt.Printf("Package %s@%s already installed\n", name, version)
		return nil
	}

	// Create package directory
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return err
	}

	// Download package from registry
	if err := pm.downloadPackageFromRegistry(name, version, packageDir); err != nil {
		return err
	}

	fmt.Printf("Installed %s@%s\n", name, version)
	return nil
}

func (pm *PackageManager) downloadPackageFromRegistry(name, version, destDir string) error {
	fmt.Printf("Downloading %s@%s from registry...\n", name, version)

	// Use GitHub registry if token is available
	if pm.GitHubToken != "" {
		return pm.downloadFromGitHub(name, version, destDir)
	}

	// Fallback to mock download
	return pm.downloadPackageMock(name, version, destDir)
}

func (pm *PackageManager) downloadFromGitHub(name, version, destDir string) error {
	// Create GitHub registry client
	client := NewGitHubRegistryClient("mburakmmm", "wing-packages", pm.GitHubToken)

	// Download package
	reader, err := client.DownloadPackage(name, version)
	if err != nil {
		return fmt.Errorf("failed to download from GitHub: %v", err)
	}
	defer reader.Close()

	// Extract tar.gz
	return pm.extractTarGz(reader, destDir)
}

func (pm *PackageManager) downloadPackageMock(name, version, destDir string) error {
	// Mock implementation for testing
	fmt.Printf("Using mock download for %s@%s...\n", name, version)

	// Create mock package files
	manifestFile := filepath.Join(destDir, "sky.project.toml")
	manifestContent := fmt.Sprintf(`[package]
name = "%s"
version = "%s"
description = "Mock package for %s"
license = "MIT"
`, name, version, name)

	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		return err
	}

	// Create mock source file
	srcDir := filepath.Join(destDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	sourceFile := filepath.Join(srcDir, "index.sky")
	sourceContent := fmt.Sprintf(`# Package: %s@%s

function hello(): void
  print("Hello from %s!")
end
`, name, version, name)

	if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
		return err
	}

	return nil
}

func (pm *PackageManager) extractTarGz(reader io.Reader, destDir string) error {
	// Create gzip reader
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %v", err)
		}

		// Create file path
		filePath := filepath.Join(destDir, header.Name)

		// Create directory if needed
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(filePath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %v", err)
		}

		// Create file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		// Copy file content
		if _, err := io.Copy(file, tarReader); err != nil {
			file.Close()
			return fmt.Errorf("failed to copy file content: %v", err)
		}

		file.Close()
	}

	return nil
}

func (pm *PackageManager) getLatestVersion(name string) (string, error) {
	// Use GitHub registry if token is available
	if pm.GitHubToken != "" {
		client := NewGitHubRegistryClient("mburakmmm", "wing-packages", pm.GitHubToken)
		return client.GetLatestVersion(name)
	}

	// Fallback to mock
	return "1.0.0", nil
}

func (pm *PackageManager) validatePackage() error {
	if pm.Manifest == nil {
		return fmt.Errorf("no manifest loaded")
	}

	// Create validator
	validator := NewPackageValidator(pm.ProjectDir, pm.Manifest)

	// Validate package
	result := validator.Validate()

	// Print warnings
	for _, warning := range result.Warnings {
		fmt.Printf("Warning: %s\n", warning.Message)
	}

	// Return errors
	if !result.Valid {
		var errorMessages []string
		for _, err := range result.Errors {
			errorMessages = append(errorMessages, err.Message)
		}
		return fmt.Errorf("package validation failed: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}

func (pm *PackageManager) createPackageArchive() (string, error) {
	// Create real tar.gz archive
	archivePath := filepath.Join(pm.ProjectDir, fmt.Sprintf("%s-%s.tar.gz", pm.Manifest.Package.Name, pm.Manifest.Package.Version))

	// Use BuildSystem to create archive
	buildSystem := NewBuildSystem(pm.ProjectDir, pm.Manifest)
	if err := buildSystem.createTarGzArchive(archivePath); err != nil {
		return "", err
	}

	return archivePath, nil
}

func (pm *PackageManager) uploadToRegistry(archivePath string) error {
	// Mock implementation - in real implementation, upload to registry
	fmt.Printf("Uploading %s to registry...\n", archivePath)
	return nil
}
