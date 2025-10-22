package wing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitHubRegistryClient handles communication with GitHub-based registry
type GitHubRegistryClient struct {
	Owner     string
	Repo      string
	Token     string
	BaseURL   string
	Client    *http.Client
	LocalPath string // Local clone path
}

// NewGitHubRegistryClient creates a new GitHub registry client
func NewGitHubRegistryClient(owner, repo, token string) *GitHubRegistryClient {
	return &GitHubRegistryClient{
		Owner:     owner,
		Repo:      repo,
		Token:     token,
		BaseURL:   fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo),
		Client:    &http.Client{Timeout: 30 * time.Second},
		LocalPath: filepath.Join(os.TempDir(), "wing-registry", repo),
	}
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

// SearchPackages searches for packages in the GitHub registry
func (grc *GitHubRegistryClient) SearchPackages(query string) ([]RegistryPackage, error) {
	// Load local index
	index, err := grc.loadLocalIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %v", err)
	}

	var results []RegistryPackage
	for name, entry := range index.Packages {
		if strings.Contains(strings.ToLower(name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(entry.Description), strings.ToLower(query)) {

			// Convert to RegistryPackage
			pkg := &RegistryPackage{
				Name:         name,
				Version:      entry.Latest,
				Description:  entry.Description,
				Authors:      entry.Authors,
				License:      entry.License,
				Repository:   entry.Repository,
				Homepage:     entry.Homepage,
				Keywords:     entry.Keywords,
				Downloads:    entry.Downloads,
				CreatedAt:    entry.CreatedAt,
				UpdatedAt:    entry.UpdatedAt,
				Versions:     make([]string, 0, len(entry.Versions)),
				Dependencies: entry.Dependencies,
			}

			// Add versions
			for version := range entry.Versions {
				pkg.Versions = append(pkg.Versions, version)
			}

			results = append(results, *pkg)
		}
	}

	return results, nil
}

// GetPackage retrieves package information from the GitHub registry
func (grc *GitHubRegistryClient) GetPackage(name string) (*RegistryPackage, error) {
	index, err := grc.loadLocalIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %v", err)
	}

	fmt.Printf("Looking for package '%s' in registry with %d packages\n", name, len(index.Packages))
	for pkgName := range index.Packages {
		fmt.Printf("  - %s\n", pkgName)
	}

	entry, exists := index.Packages[name]
	if !exists {
		return nil, fmt.Errorf("package '%s' not found in registry", name)
	}

	// Convert to RegistryPackage
	pkg := &RegistryPackage{
		Name:         name,
		Version:      entry.Latest,
		Description:  entry.Description,
		Authors:      entry.Authors,
		License:      entry.License,
		Repository:   entry.Repository,
		Homepage:     entry.Homepage,
		Keywords:     entry.Keywords,
		Downloads:    entry.Downloads,
		CreatedAt:    entry.CreatedAt,
		UpdatedAt:    entry.UpdatedAt,
		Versions:     make([]string, 0, len(entry.Versions)),
		Dependencies: entry.Dependencies,
	}

	// Add versions
	for version := range entry.Versions {
		pkg.Versions = append(pkg.Versions, version)
	}

	return pkg, nil
}

// GetPackageVersion retrieves specific version information
func (grc *GitHubRegistryClient) GetPackageVersion(name, version string) (*RegistryPackage, error) {
	index, err := grc.loadLocalIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %v", err)
	}

	entry, exists := index.Packages[name]
	if !exists {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	if _, exists := entry.Versions[version]; !exists {
		return nil, fmt.Errorf("package '%s' version '%s' not found", name, version)
	}

	// Convert to RegistryPackage
	pkg := &RegistryPackage{
		Name:         name,
		Version:      version,
		Description:  entry.Description,
		Authors:      entry.Authors,
		License:      entry.License,
		Repository:   entry.Repository,
		Homepage:     entry.Homepage,
		Keywords:     entry.Keywords,
		Downloads:    entry.Downloads,
		CreatedAt:    entry.CreatedAt,
		UpdatedAt:    entry.UpdatedAt,
		Versions:     make([]string, 0, len(entry.Versions)),
		Dependencies: entry.Dependencies,
	}

	// Add versions
	for v := range entry.Versions {
		pkg.Versions = append(pkg.Versions, v)
	}

	return pkg, nil
}

// DownloadPackage downloads a package archive
func (grc *GitHubRegistryClient) DownloadPackage(name, version string) (io.ReadCloser, error) {
	index, err := grc.loadLocalIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %v", err)
	}

	entry, exists := index.Packages[name]
	if !exists {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	downloadURL, exists := entry.Versions[version]
	if !exists {
		return nil, fmt.Errorf("package '%s' version '%s' not found", name, version)
	}

	// Download from GitHub releases
	resp, err := grc.Client.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download package: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// PublishPackage publishes a package to the GitHub registry
func (grc *GitHubRegistryClient) PublishPackage(manifest *PackageManifest, archiveData []byte) error {
	// Create GitHub release
	release, err := grc.createGitHubRelease(manifest, archiveData)
	if err != nil {
		return fmt.Errorf("failed to create release: %v", err)
	}

	// Update local index
	if err := grc.updateLocalIndex(manifest, release.HTMLURL); err != nil {
		return fmt.Errorf("failed to update index: %v", err)
	}

	// Commit and push changes
	if err := grc.commitAndPushIndex(); err != nil {
		return fmt.Errorf("failed to push changes: %v", err)
	}

	return nil
}

// GetLatestVersion gets the latest version of a package
func (grc *GitHubRegistryClient) GetLatestVersion(name string) (string, error) {
	pkg, err := grc.GetPackage(name)
	if err != nil {
		return "", err
	}

	return pkg.Version, nil
}

// CheckPackageExists checks if a package exists in the registry
func (grc *GitHubRegistryClient) CheckPackageExists(name string) (bool, error) {
	_, err := grc.GetPackage(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetPackageDependencies gets the dependencies of a package
func (grc *GitHubRegistryClient) GetPackageDependencies(name, version string) (map[string]string, error) {
	pkg, err := grc.GetPackageVersion(name, version)
	if err != nil {
		return nil, err
	}

	return pkg.Dependencies, nil
}

// ResolveDependencies resolves all dependencies for a package
func (grc *GitHubRegistryClient) ResolveDependencies(name, version string) (map[string]string, error) {
	deps := make(map[string]string)

	// Get direct dependencies
	directDeps, err := grc.GetPackageDependencies(name, version)
	if err != nil {
		return nil, err
	}

	// Add direct dependencies
	for depName, depVersion := range directDeps {
		deps[depName] = depVersion
	}

	// Recursively resolve transitive dependencies
	for depName, depVersion := range directDeps {
		transitiveDeps, err := grc.ResolveDependencies(depName, depVersion)
		if err != nil {
			return nil, err
		}

		// Add transitive dependencies (avoid conflicts)
		for transName, transVersion := range transitiveDeps {
			if existingVersion, exists := deps[transName]; exists {
				// Version conflict - use the higher version
				if transVersion > existingVersion {
					deps[transName] = transVersion
				}
			} else {
				deps[transName] = transVersion
			}
		}
	}

	return deps, nil
}

// Helper methods

func (grc *GitHubRegistryClient) loadLocalIndex() (*PackageIndex, error) {
	// First try to download from GitHub
	index, err := grc.downloadIndexFromGitHub()
	if err != nil {
		// Fallback to local file
		indexPath := filepath.Join(grc.LocalPath, "packages.json")

		// Check if local index exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			// Initialize empty index
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

		var localIndex PackageIndex
		if err := json.Unmarshal(data, &localIndex); err != nil {
			return nil, err
		}

		return &localIndex, nil
	}

	return index, nil
}

func (grc *GitHubRegistryClient) downloadIndexFromGitHub() (*PackageIndex, error) {
	// Download packages.json from GitHub repository
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/packages.json", grc.Owner, grc.Repo)
	fmt.Printf("Downloading index from: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if grc.Token != "" {
		req.Header.Set("Authorization", "token "+grc.Token)
	}

	resp, err := grc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("GitHub response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download index: status %d", resp.StatusCode)
	}

	var index PackageIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, err
	}

	fmt.Printf("Downloaded index with %d packages\n", len(index.Packages))
	return &index, nil
}

func (grc *GitHubRegistryClient) updateLocalIndex(manifest *PackageManifest, downloadURL string) error {
	index, err := grc.loadLocalIndex()
	if err != nil {
		return err
	}

	// Create or update package entry
	entry := &PackageIndexEntry{
		Name:         manifest.Package.Name,
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
	if existing, exists := index.Packages[manifest.Package.Name]; exists {
		entry.Downloads = existing.Downloads
		entry.CreatedAt = existing.CreatedAt
		entry.Versions = existing.Versions
	}

	// Add new version
	entry.Versions[manifest.Package.Version] = downloadURL
	entry.UpdatedAt = time.Now()

	// Update index
	index.Packages[manifest.Package.Name] = entry
	index.Updated = time.Now()

	// Save index
	return grc.saveLocalIndex(index)
}

func (grc *GitHubRegistryClient) saveLocalIndex(index *PackageIndex) error {
	// Ensure local directory exists
	if err := os.MkdirAll(grc.LocalPath, 0755); err != nil {
		return err
	}

	// Save index
	indexPath := filepath.Join(grc.LocalPath, "packages.json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0644)
}

func (grc *GitHubRegistryClient) createGitHubRelease(manifest *PackageManifest, archiveData []byte) (*GitHubRelease, error) {
	// Create release request
	releaseReq := &GitHubReleaseRequest{
		TagName:    fmt.Sprintf("v%s", manifest.Package.Version),
		Name:       fmt.Sprintf("%s %s", manifest.Package.Name, manifest.Package.Version),
		Body:       manifest.Package.Description,
		Draft:      false,
		Prerelease: false,
	}

	// Create release
	release, err := grc.createRelease(releaseReq)
	if err != nil {
		return nil, err
	}

	// Upload asset
	assetName := fmt.Sprintf("%s-%s.tar.gz", manifest.Package.Name, manifest.Package.Version)
	assetURL, err := grc.uploadAsset(release.ID, assetName, archiveData)
	if err != nil {
		return nil, err
	}

	release.HTMLURL = assetURL
	return release, nil
}

func (grc *GitHubRegistryClient) createRelease(req *GitHubReleaseRequest) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/releases", grc.BaseURL)

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "token "+grc.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := grc.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create release: %s", string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func (grc *GitHubRegistryClient) uploadAsset(releaseID int64, assetName string, assetData []byte) (string, error) {
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets?name=%s",
		grc.Owner, grc.Repo, releaseID, assetName)

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(assetData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "token "+grc.Token)
	httpReq.Header.Set("Content-Type", "application/gzip")

	resp, err := grc.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload asset: %s", string(body))
	}

	var asset GitHubAsset
	if err := json.NewDecoder(resp.Body).Decode(&asset); err != nil {
		return "", err
	}

	return asset.BrowserDownloadURL, nil
}

func (grc *GitHubRegistryClient) commitAndPushIndex() error {
	// This would require git operations
	// For now, we'll just save locally
	// In a real implementation, you'd use go-git or exec.Command to git operations
	return nil
}

// GitHub API types
type GitHubReleaseRequest struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

type GitHubRelease struct {
	ID        int64  `json:"id"`
	TagName   string `json:"tag_name"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	HTMLURL   string `json:"html_url"`
	UploadURL string `json:"upload_url"`
}

type GitHubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}
