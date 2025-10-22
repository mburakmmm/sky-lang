package wing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RegistryClient handles communication with the Sky package registry
type RegistryClient struct {
	BaseURL string
	Client  *http.Client
}

// NewRegistryClient creates a new registry client
func NewRegistryClient(baseURL string) *RegistryClient {
	return &RegistryClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchPackages searches for packages in the registry
func (rc *RegistryClient) SearchPackages(query string) ([]RegistryPackage, error) {
	url := fmt.Sprintf("%s/search?q=%s", rc.BaseURL, query)

	resp, err := rc.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var packages []RegistryPackage
	if err := json.Unmarshal(body, &packages); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return packages, nil
}

// GetPackage retrieves package information from the registry
func (rc *RegistryClient) GetPackage(name string) (*RegistryPackage, error) {
	url := fmt.Sprintf("%s/packages/%s", rc.BaseURL, name)

	resp, err := rc.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var pkg RegistryPackage
	if err := json.Unmarshal(body, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &pkg, nil
}

// GetPackageVersion retrieves specific version information
func (rc *RegistryClient) GetPackageVersion(name, version string) (*RegistryPackage, error) {
	url := fmt.Sprintf("%s/packages/%s/versions/%s", rc.BaseURL, name, version)

	resp, err := rc.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get package version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package '%s' version '%s' not found", name, version)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var pkg RegistryPackage
	if err := json.Unmarshal(body, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &pkg, nil
}

// DownloadPackage downloads a package archive
func (rc *RegistryClient) DownloadPackage(name, version string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/packages/%s/versions/%s/download", rc.BaseURL, name, version)

	resp, err := rc.Client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download package: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// PublishPackage publishes a package to the registry
func (rc *RegistryClient) PublishPackage(manifest *PackageManifest, archiveData []byte) error {
	url := fmt.Sprintf("%s/packages", rc.BaseURL)

	// Create multipart form data
	// In a real implementation, use multipart/form-data
	// For now, we'll use a simplified approach

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication header if needed
	// req.Header.Set("Authorization", "Bearer "+token)

	resp, err := rc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish package: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetLatestVersion gets the latest version of a package
func (rc *RegistryClient) GetLatestVersion(name string) (string, error) {
	pkg, err := rc.GetPackage(name)
	if err != nil {
		return "", err
	}

	if len(pkg.Versions) == 0 {
		return "", fmt.Errorf("no versions available for package '%s'", name)
	}

	// Return the latest version (assuming versions are sorted)
	return pkg.Versions[len(pkg.Versions)-1], nil
}

// CheckPackageExists checks if a package exists in the registry
func (rc *RegistryClient) CheckPackageExists(name string) (bool, error) {
	_, err := rc.GetPackage(name)
	if err != nil {
		// Check if it's a "not found" error
		if err.Error() == fmt.Sprintf("package '%s' not found", name) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetPackageDependencies gets the dependencies of a package
func (rc *RegistryClient) GetPackageDependencies(name, version string) (map[string]string, error) {
	pkg, err := rc.GetPackageVersion(name, version)
	if err != nil {
		return nil, err
	}

	return pkg.Dependencies, nil
}

// ResolveDependencies resolves all dependencies for a package
func (rc *RegistryClient) ResolveDependencies(name, version string) (map[string]string, error) {
	deps := make(map[string]string)

	// Get direct dependencies
	directDeps, err := rc.GetPackageDependencies(name, version)
	if err != nil {
		return nil, err
	}

	// Add direct dependencies
	for depName, depVersion := range directDeps {
		deps[depName] = depVersion
	}

	// Recursively resolve transitive dependencies
	for depName, depVersion := range directDeps {
		transitiveDeps, err := rc.ResolveDependencies(depName, depVersion)
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

// MockRegistryClient is a mock implementation for testing
type MockRegistryClient struct {
	Packages map[string]*RegistryPackage
}

// NewMockRegistryClient creates a mock registry client
func NewMockRegistryClient() *MockRegistryClient {
	return &MockRegistryClient{
		Packages: make(map[string]*RegistryPackage),
	}
}

// AddMockPackage adds a mock package to the registry
func (mrc *MockRegistryClient) AddMockPackage(pkg *RegistryPackage) {
	mrc.Packages[pkg.Name] = pkg
}

// SearchPackages implements RegistryClient interface
func (mrc *MockRegistryClient) SearchPackages(query string) ([]RegistryPackage, error) {
	var results []RegistryPackage

	for _, pkg := range mrc.Packages {
		if contains(pkg.Name, query) || contains(pkg.Description, query) {
			results = append(results, *pkg)
		}
	}

	return results, nil
}

// GetPackage implements RegistryClient interface
func (mrc *MockRegistryClient) GetPackage(name string) (*RegistryPackage, error) {
	if pkg, exists := mrc.Packages[name]; exists {
		return pkg, nil
	}
	return nil, fmt.Errorf("package '%s' not found", name)
}

// GetPackageVersion implements RegistryClient interface
func (mrc *MockRegistryClient) GetPackageVersion(name, version string) (*RegistryPackage, error) {
	if pkg, exists := mrc.Packages[name]; exists {
		if pkg.Version == version {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("package '%s' version '%s' not found", name, version)
}

// DownloadPackage implements RegistryClient interface
func (mrc *MockRegistryClient) DownloadPackage(name, version string) (io.ReadCloser, error) {
	// Return a mock reader
	return io.NopCloser(strings.NewReader("mock package data")), nil
}

// PublishPackage implements RegistryClient interface
func (mrc *MockRegistryClient) PublishPackage(manifest *PackageManifest, archiveData []byte) error {
	// Mock implementation - just add to packages
	pkg := &RegistryPackage{
		Name:         manifest.Package.Name,
		Version:      manifest.Package.Version,
		Description:  manifest.Package.Description,
		Authors:      manifest.Package.Authors,
		License:      manifest.Package.License,
		Repository:   manifest.Package.Repository,
		Homepage:     manifest.Package.Homepage,
		Keywords:     manifest.Package.Keywords,
		Dependencies: manifest.Dependencies,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Versions:     []string{manifest.Package.Version},
	}

	mrc.Packages[manifest.Package.Name] = pkg
	return nil
}

// GetLatestVersion implements RegistryClient interface
func (mrc *MockRegistryClient) GetLatestVersion(name string) (string, error) {
	if pkg, exists := mrc.Packages[name]; exists {
		if len(pkg.Versions) > 0 {
			return pkg.Versions[len(pkg.Versions)-1], nil
		}
		return pkg.Version, nil
	}
	return "", fmt.Errorf("package '%s' not found", name)
}

// CheckPackageExists implements RegistryClient interface
func (mrc *MockRegistryClient) CheckPackageExists(name string) (bool, error) {
	_, exists := mrc.Packages[name]
	return exists, nil
}

// GetPackageDependencies implements RegistryClient interface
func (mrc *MockRegistryClient) GetPackageDependencies(name, version string) (map[string]string, error) {
	if pkg, exists := mrc.Packages[name]; exists {
		return pkg.Dependencies, nil
	}
	return nil, fmt.Errorf("package '%s' not found", name)
}

// ResolveDependencies implements RegistryClient interface
func (mrc *MockRegistryClient) ResolveDependencies(name, version string) (map[string]string, error) {
	deps := make(map[string]string)

	if pkg, exists := mrc.Packages[name]; exists {
		// Add direct dependencies
		for depName, depVersion := range pkg.Dependencies {
			deps[depName] = depVersion
		}

		// Recursively resolve transitive dependencies
		for depName, depVersion := range pkg.Dependencies {
			transitiveDeps, err := mrc.ResolveDependencies(depName, depVersion)
			if err != nil {
				return nil, err
			}

			for transName, transVersion := range transitiveDeps {
				if existingVersion, exists := deps[transName]; exists {
					if transVersion > existingVersion {
						deps[transName] = transVersion
					}
				} else {
					deps[transName] = transVersion
				}
			}
		}
	}

	return deps, nil
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
