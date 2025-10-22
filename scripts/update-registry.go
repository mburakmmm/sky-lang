package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// PackageIndex represents the registry index structure
type PackageIndex struct {
	Packages map[string]*PackageIndexEntry `json:"packages"`
	Updated  time.Time                     `json:"updated"`
	Version  string                        `json:"version"`
}

// PackageIndexEntry represents a package in the index
type PackageIndexEntry struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Latest       string            `json:"latest"`
	Versions     map[string]string `json:"versions"` // version -> download_url
	Authors      []string          `json:"authors"`
	License      string            `json:"license"`
	Repository   string            `json:"repository"`
	Homepage     string            `json:"homepage"`
	Keywords     []string          `json:"keywords"`
	Downloads    int64             `json:"downloads"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Dependencies map[string]string `json:"dependencies"`
}

func main() {
	fmt.Println("Updating package registry...")

	// Load existing index
	index, err := loadIndex()
	if err != nil {
		fmt.Printf("Error loading index: %v\n", err)
		os.Exit(1)
	}

	// Scan packages directory
	packagesDir := "packages"
	if err := filepath.Walk(packagesDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for sky.project.toml files
		if info.Name() == "sky.project.toml" {
			if err := processPackage(path, index); err != nil {
				fmt.Printf("Error processing package %s: %v\n", path, err)
			}
		}

		return nil
	}); err != nil {
		fmt.Printf("Error scanning packages: %v\n", err)
		os.Exit(1)
	}

	// Update timestamp
	index.Updated = time.Now()

	// Save updated index
	if err := saveIndex(index); err != nil {
		fmt.Printf("Error saving index: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Package registry updated successfully!")
}

func loadIndex() (*PackageIndex, error) {
	indexPath := "packages.json"

	// Check if index exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Create new index
		return &PackageIndex{
			Packages: make(map[string]*PackageIndexEntry),
			Updated:  time.Now(),
			Version:  "1.0.0",
		}, nil
	}

	// Load existing index
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var index PackageIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

func saveIndex(index *PackageIndex) error {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("packages.json", data, 0644)
}

func processPackage(manifestPath string, index *PackageIndex) error {
	// Read manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest PackageManifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return err
	}

	// Get package directory
	packageDir := filepath.Dir(manifestPath)
	packageName := manifest.Package.Name

	// Find tar.gz file
	tarGzPath := filepath.Join(packageDir, fmt.Sprintf("%s-%s.tar.gz", packageName, manifest.Package.Version))
	if _, err := os.Stat(tarGzPath); os.IsNotExist(err) {
		return fmt.Errorf("package archive not found: %s", tarGzPath)
	}

	// Create download URL (GitHub releases URL)
	downloadURL := fmt.Sprintf("https://github.com/mburakmmm/wing-packages/releases/download/v%s/%s-%s.tar.gz",
		manifest.Package.Version, packageName, manifest.Package.Version)

	// Create or update package entry
	entry := &PackageIndexEntry{
		Name:         packageName,
		Description:  manifest.Package.Description,
		Latest:       manifest.Package.Version,
		Versions:     make(map[string]string),
		Authors:      manifest.Package.Authors,
		License:      manifest.Package.License,
		Repository:   manifest.Package.Repository,
		Homepage:     manifest.Package.Homepage,
		Keywords:     manifest.Package.Keywords,
		Downloads:    0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Dependencies: manifest.Dependencies,
	}

	// If package exists, preserve some data
	if existing, exists := index.Packages[packageName]; exists {
		entry.Downloads = existing.Downloads
		entry.CreatedAt = existing.CreatedAt
		entry.Versions = existing.Versions
	}

	// Add new version
	entry.Versions[manifest.Package.Version] = downloadURL
	entry.UpdatedAt = time.Now()

	// Update index
	index.Packages[packageName] = entry

	fmt.Printf("Processed package: %s@%s\n", packageName, manifest.Package.Version)
	return nil
}
