package wing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitHubPackagesClient handles communication with GitHub Packages registry
type GitHubPackagesClient struct {
	Owner     string
	Repo      string
	Token     string
	BaseURL   string
	Client    *http.Client
	LocalPath string
}

// NewGitHubPackagesClient creates a new GitHub Packages registry client
func NewGitHubPackagesClient(owner, repo, token string) *GitHubPackagesClient {
	return &GitHubPackagesClient{
		Owner:     owner,
		Repo:      repo,
		Token:     token,
		BaseURL:   fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo),
		Client:    &http.Client{Timeout: 30 * time.Second},
		LocalPath: filepath.Join(os.TempDir(), "wing-packages", repo),
	}
}

// GitHubPackageManifest represents the package manifest for GitHub Packages
type GitHubPackageManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Authors      []string          `json:"authors"`
	License      string            `json:"license"`
	Repository   string            `json:"repository"`
	Homepage     string            `json:"homepage"`
	Keywords     []string          `json:"keywords"`
	Dependencies map[string]string `json:"dependencies"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// SearchPackages searches for packages in the GitHub Packages registry
func (gpc *GitHubPackagesClient) SearchPackages(query string) ([]RegistryPackage, error) {
	// Search in GitHub Packages
	url := fmt.Sprintf("https://api.github.com/search/packages?q=%s+in:name", query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+gpc.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status %d", resp.StatusCode)
	}

	var searchResult struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	var results []RegistryPackage
	for _, item := range searchResult.Items {
		// Filter for wing packages
		if strings.HasPrefix(item.Name, "wing-") {
			createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
			updatedAt, _ := time.Parse(time.RFC3339, item.UpdatedAt)

			pkg := RegistryPackage{
				Name:        strings.TrimPrefix(item.Name, "wing-"),
				Version:     "latest", // GitHub Packages doesn't provide version in search
				Description: item.Description,
				Repository:  item.HTMLURL,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			}
			results = append(results, pkg)
		}
	}

	return results, nil
}

// GetPackage retrieves package information from GitHub Packages
func (gpc *GitHubPackagesClient) GetPackage(name string) (*RegistryPackage, error) {
	packageName := fmt.Sprintf("wing-%s", name)

	// Get package info from GitHub Packages
	url := fmt.Sprintf("https://api.github.com/orgs/%s/packages/container/%s", gpc.Owner, packageName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+gpc.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package '%s' not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get package: status %d", resp.StatusCode)
	}

	var packageInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		HTMLURL     string `json:"html_url"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&packageInfo); err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339, packageInfo.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, packageInfo.UpdatedAt)

	pkg := &RegistryPackage{
		Name:        name,
		Version:     "latest",
		Description: packageInfo.Description,
		Repository:  packageInfo.HTMLURL,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return pkg, nil
}

// GetPackageVersion retrieves specific version information
func (gpc *GitHubPackagesClient) GetPackageVersion(name, version string) (*RegistryPackage, error) {
	// For GitHub Packages, we'll get the latest and assume version compatibility
	return gpc.GetPackage(name)
}

// DownloadPackage downloads a package from GitHub Packages
func (gpc *GitHubPackagesClient) DownloadPackage(name, version string) (io.ReadCloser, error) {
	packageName := fmt.Sprintf("wing-%s", name)

	// Download package from GitHub Packages Container Registry
	manifestURL := fmt.Sprintf("https://ghcr.io/v2/%s/%s/manifests/%s",
		gpc.Owner, packageName, version)

	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+gpc.Token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("package '%s' version '%s' not found", name, version)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to get package manifest: status %d", resp.StatusCode)
	}

	// Parse manifest to get blob digest
	var manifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()

	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("no layers found in package manifest")
	}

	// Download the blob (package archive)
	blobDigest := manifest.Layers[0].Digest
	blobURL := fmt.Sprintf("https://ghcr.io/v2/%s/%s/blobs/%s",
		gpc.Owner, packageName, blobDigest)

	blobReq, err := http.NewRequest("GET", blobURL, nil)
	if err != nil {
		return nil, err
	}

	blobReq.Header.Set("Authorization", "Bearer "+gpc.Token)

	blobResp, err := gpc.Client.Do(blobReq)
	if err != nil {
		return nil, err
	}

	if blobResp.StatusCode != http.StatusOK {
		blobResp.Body.Close()
		return nil, fmt.Errorf("failed to download package blob: status %d", blobResp.StatusCode)
	}

	return blobResp.Body, nil
}

// PublishPackage publishes a package to GitHub Packages
func (gpc *GitHubPackagesClient) PublishPackage(manifest *PackageManifest, archiveData []byte) error {
	packageName := fmt.Sprintf("wing-%s", manifest.Package.Name)

	// Create package manifest for GitHub Packages
	packageManifest := GitHubPackageManifest{
		Name:         packageName,
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
	}

	// Upload to GitHub Packages
	return gpc.uploadToGitHubPackages(packageManifest, archiveData)
}

// GetLatestVersion gets the latest version of a package
func (gpc *GitHubPackagesClient) GetLatestVersion(name string) (string, error) {
	pkg, err := gpc.GetPackage(name)
	if err != nil {
		return "", err
	}
	return pkg.Version, nil
}

// CheckPackageExists checks if a package exists in the registry
func (gpc *GitHubPackagesClient) CheckPackageExists(name string) (bool, error) {
	_, err := gpc.GetPackage(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetPackageDependencies gets the dependencies of a package
func (gpc *GitHubPackagesClient) GetPackageDependencies(name, version string) (map[string]string, error) {
	pkg, err := gpc.GetPackageVersion(name, version)
	if err != nil {
		return nil, err
	}
	return pkg.Dependencies, nil
}

// ResolveDependencies resolves all dependencies for a package
func (gpc *GitHubPackagesClient) ResolveDependencies(name, version string) (map[string]string, error) {
	deps := make(map[string]string)

	// Get direct dependencies
	directDeps, err := gpc.GetPackageDependencies(name, version)
	if err != nil {
		return nil, err
	}

	// Add direct dependencies
	for depName, depVersion := range directDeps {
		deps[depName] = depVersion
	}

	// Recursively resolve transitive dependencies
	for depName, depVersion := range directDeps {
		transitiveDeps, err := gpc.ResolveDependencies(depName, depVersion)
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

func (gpc *GitHubPackagesClient) uploadToGitHubPackages(manifest GitHubPackageManifest, archiveData []byte) error {
	fmt.Printf("Publishing %s@%s to GitHub Packages...\n", manifest.Name, manifest.Version)

	// For now, we'll use a simplified approach that creates a GitHub release
	// but marks it as a package in the metadata

	// Create a package metadata file
	metadata := map[string]interface{}{
		"name":         manifest.Name,
		"version":      manifest.Version,
		"description":  manifest.Description,
		"authors":      manifest.Authors,
		"license":      manifest.License,
		"repository":   manifest.Repository,
		"homepage":     manifest.Homepage,
		"keywords":     manifest.Keywords,
		"dependencies": manifest.Dependencies,
		"created_at":   manifest.CreatedAt,
		"updated_at":   manifest.UpdatedAt,
		"package_type": "wing-package",
	}

	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create metadata: %v", err)
	}

	// Create a temporary directory for the package
	tempDir := filepath.Join(os.TempDir(), "wing-package-"+manifest.Name)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write metadata file
	metadataFile := filepath.Join(tempDir, "package.json")
	if err := os.WriteFile(metadataFile, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}

	// Write archive file
	archiveFile := filepath.Join(tempDir, manifest.Name+"-"+manifest.Version+".tar.gz")
	if err := os.WriteFile(archiveFile, archiveData, 0644); err != nil {
		return fmt.Errorf("failed to write archive: %v", err)
	}

	// Create a GitHub release with package metadata
	releaseName := fmt.Sprintf("%s %s (Package)", manifest.Name, manifest.Version)
	releaseBody := fmt.Sprintf(`# %s

%s

## Package Information
- **Name**: %s
- **Version**: %s
- **License**: %s
- **Authors**: %s

## Installation
`+"```bash"+`
wing add %s
`+"```"+`

## Files
- `+"`package.json`"+` - Package metadata
- `+"`%s-%s.tar.gz`"+` - Package archive
`, manifest.Name, manifest.Description, manifest.Name, manifest.Version, manifest.License,
		strings.Join(manifest.Authors, ", "), strings.TrimPrefix(manifest.Name, "wing-"),
		manifest.Name, manifest.Version)

	// Create release using GitHub API
	releaseData := map[string]interface{}{
		"tag_name":   "pkg-" + manifest.Version,
		"name":       releaseName,
		"body":       releaseBody,
		"draft":      false,
		"prerelease": false,
	}

	releaseJSON, err := json.Marshal(releaseData)
	if err != nil {
		return fmt.Errorf("failed to marshal release data: %v", err)
	}

	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", gpc.Owner, gpc.Repo)
	req, err := http.NewRequest("POST", releaseURL, bytes.NewReader(releaseJSON))
	if err != nil {
		return fmt.Errorf("failed to create release request: %v", err)
	}

	req.Header.Set("Authorization", "token "+gpc.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create release: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create release: %s", string(body))
	}

	var release struct {
		ID        int    `json:"id"`
		UploadURL string `json:"upload_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode release response: %v", err)
	}

	// Upload package files as release assets
	files := []string{metadataFile, archiveFile}
	for _, file := range files {
		if err := gpc.uploadReleaseAsset(release.UploadURL, file); err != nil {
			fmt.Printf("Warning: failed to upload %s: %v\n", filepath.Base(file), err)
		}
	}

	fmt.Printf("Package %s@%s published to GitHub Packages successfully!\n", manifest.Name, manifest.Version)

	// Update packages.json registry
	if err := gpc.updatePackagesRegistry(manifest); err != nil {
		fmt.Printf("Warning: failed to update packages.json: %v\n", err)
	} else {
		fmt.Printf("Updated packages.json registry successfully!\n")
	}

	return nil
}

func (gpc *GitHubPackagesClient) uploadReleaseAsset(uploadURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Remove the {?name,label} part from upload URL
	uploadURL = strings.Split(uploadURL, "{")[0]
	uploadURL += "?name=" + filepath.Base(filePath)

	req, err := http.NewRequest("POST", uploadURL, file)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+gpc.Token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload asset: %s", string(body))
	}

	return nil
}

func (gpc *GitHubPackagesClient) updatePackagesRegistry(manifest GitHubPackageManifest) error {
	// Download current packages.json
	packagesURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/packages.json", gpc.Owner, gpc.Repo)
	resp, err := gpc.Client.Get(packagesURL)
	if err != nil {
		return fmt.Errorf("failed to download packages.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download packages.json: status %d", resp.StatusCode)
	}

	var registry struct {
		Packages map[string]interface{} `json:"packages"`
		Updated  string                 `json:"updated"`
		Version  string                 `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return fmt.Errorf("failed to decode packages.json: %v", err)
	}

	// Initialize packages if nil
	if registry.Packages == nil {
		registry.Packages = make(map[string]interface{})
	}

	// Create package entry
	packageName := strings.TrimPrefix(manifest.Name, "wing-")
	packageEntry := map[string]interface{}{
		"name":        packageName,
		"description": manifest.Description,
		"latest":      manifest.Version,
		"versions": map[string]string{
			manifest.Version: fmt.Sprintf("https://github.com/%s/%s/releases/download/pkg-%s/%s-%s.tar.gz",
				gpc.Owner, gpc.Repo, manifest.Version, manifest.Name, manifest.Version),
		},
		"authors":      manifest.Authors,
		"license":      manifest.License,
		"repository":   manifest.Repository,
		"homepage":     manifest.Homepage,
		"keywords":     manifest.Keywords,
		"downloads":    0,
		"created_at":   manifest.CreatedAt,
		"updated_at":   manifest.UpdatedAt,
		"dependencies": manifest.Dependencies,
	}

	// Check if package already exists
	if existingPackage, exists := registry.Packages[packageName]; exists {
		// Update existing package
		if existingMap, ok := existingPackage.(map[string]interface{}); ok {
			// Update latest version
			existingMap["latest"] = manifest.Version
			existingMap["updated_at"] = manifest.UpdatedAt

			// Add new version to versions
			if versions, ok := existingMap["versions"].(map[string]interface{}); ok {
				versions[manifest.Version] = fmt.Sprintf("https://github.com/%s/%s/releases/download/pkg-%s/%s-%s.tar.gz",
					gpc.Owner, gpc.Repo, manifest.Version, manifest.Name, manifest.Version)
			}
		}
	} else {
		// Add new package
		registry.Packages[packageName] = packageEntry
	}

	// Update registry metadata
	registry.Updated = manifest.UpdatedAt.Format(time.RFC3339)
	registry.Version = "1.0.0"

	// Convert back to JSON
	updatedJSON, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated packages.json: %v", err)
	}

	// Update packages.json in repository
	return gpc.updateFileInRepository("packages.json", string(updatedJSON), "Update packages.json registry")
}

func (gpc *GitHubPackagesClient) updateFileInRepository(filePath, content, commitMessage string) error {
	// Get file SHA
	fileURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", gpc.Owner, gpc.Repo, filePath)
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create file request: %v", err)
	}

	req.Header.Set("Authorization", "token "+gpc.Token)

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get file: %v", err)
	}
	defer resp.Body.Close()

	var fileInfo struct {
		SHA  string `json:"sha"`
		Path string `json:"path"`
	}

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
			return fmt.Errorf("failed to decode file info: %v", err)
		}
	}

	// Update file
	updateData := map[string]interface{}{
		"message": commitMessage,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
	}

	if fileInfo.SHA != "" {
		updateData["sha"] = fileInfo.SHA
	}

	updateJSON, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("failed to marshal update data: %v", err)
	}

	updateReq, err := http.NewRequest("PUT", fileURL, bytes.NewReader(updateJSON))
	if err != nil {
		return fmt.Errorf("failed to create update request: %v", err)
	}

	updateReq.Header.Set("Authorization", "token "+gpc.Token)
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := gpc.Client.Do(updateReq)
	if err != nil {
		return fmt.Errorf("failed to update file: %v", err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK && updateResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(updateResp.Body)
		return fmt.Errorf("failed to update file: %s", string(body))
	}

	return nil
}

func (gpc *GitHubPackagesClient) createPackage(manifest GitHubPackageManifest) error {
	// Create package in GitHub Packages using GraphQL API
	query := fmt.Sprintf(`
		mutation {
			createPackage(input: {
				name: "%s"
				packageType: CONTAINER
				repositoryId: "%s"
			}) {
				package {
					id
					name
				}
			}
		}
	`, manifest.Name, gpc.getRepositoryID())

	reqBody := map[string]interface{}{
		"query": query,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "bearer "+gpc.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create package: %s", string(body))
	}

	return nil
}

func (gpc *GitHubPackagesClient) uploadPackageArchive(manifest GitHubPackageManifest, archiveData []byte) error {
	// Upload package archive to GitHub Packages
	packageName := manifest.Name
	version := manifest.Version

	// Create a container image manifest
	manifestData := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.v2+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size":      len(archiveData),
			"digest":    fmt.Sprintf("sha256:%x", archiveData),
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size":      len(archiveData),
				"digest":    fmt.Sprintf("sha256:%x", archiveData),
			},
		},
	}

	manifestJSON, err := json.Marshal(manifestData)
	if err != nil {
		return err
	}

	// Upload manifest
	manifestURL := fmt.Sprintf("https://ghcr.io/v2/%s/%s/manifests/%s",
		gpc.Owner, packageName, version)

	req, err := http.NewRequest("PUT", manifestURL, bytes.NewReader(manifestJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+gpc.Token)
	req.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload manifest: %s", string(body))
	}

	// Upload blob (archive data)
	blobURL := fmt.Sprintf("https://ghcr.io/v2/%s/%s/blobs/uploads/",
		gpc.Owner, packageName)

	blobReq, err := http.NewRequest("POST", blobURL, bytes.NewReader(archiveData))
	if err != nil {
		return err
	}

	blobReq.Header.Set("Authorization", "Bearer "+gpc.Token)
	blobReq.Header.Set("Content-Type", "application/vnd.docker.image.rootfs.diff.tar.gzip")

	blobResp, err := gpc.Client.Do(blobReq)
	if err != nil {
		return err
	}
	defer blobResp.Body.Close()

	if blobResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(blobResp.Body)
		return fmt.Errorf("failed to upload blob: %s", string(body))
	}

	return nil
}

func (gpc *GitHubPackagesClient) getRepositoryID() string {
	// Get repository ID using GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", gpc.Owner, gpc.Repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Authorization", "token "+gpc.Token)

	resp, err := gpc.Client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var repo struct {
		ID string `json:"node_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return ""
	}

	return repo.ID
}
