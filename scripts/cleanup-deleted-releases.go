package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN not set")
		os.Exit(1)
	}

	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		repo = "mburakmmm/wing-packages"
	}

	fmt.Printf("Checking for deleted releases in %s\n", repo)

	// Get all releases
	releases, err := getAllReleases(repo, token)
	if err != nil {
		fmt.Printf("Error getting releases: %v\n", err)
		os.Exit(1)
	}

	// Get current packages.json
	packages, err := getCurrentPackages(repo)
	if err != nil {
		fmt.Printf("Error getting packages.json: %v\n", err)
		os.Exit(1)
	}

	// Find packages that reference deleted releases
	var packagesToUpdate []string

	for packageName, packageData := range packages.Packages {
		if packageMap, ok := packageData.(map[string]interface{}); ok {
			if versions, ok := packageMap["versions"].(map[string]interface{}); ok {
				for version, downloadURL := range versions {
					if url, ok := downloadURL.(string); ok {
						// Extract tag from URL
						tag := extractTagFromURL(url)
						if tag != "" && !releaseExists(tag, releases) {
							fmt.Printf("Found deleted release: %s (package: %s@%s)\n", tag, packageName, version)
							packagesToUpdate = append(packagesToUpdate, packageName)
							delete(versions, version)
						}
					}
				}

				// If no versions left, remove the entire package
				if len(versions) == 0 {
					delete(packages.Packages, packageName)
					fmt.Printf("Removed entire package: %s\n", packageName)
				} else {
					// Update latest version
					var latestVersion string
					for v := range versions {
						if latestVersion == "" || v > latestVersion {
							latestVersion = v
						}
					}
					packageMap["latest"] = latestVersion
					fmt.Printf("Updated latest version for %s to %s\n", packageName, latestVersion)
				}
			}
		}
	}

	if len(packagesToUpdate) > 0 {
		// Update packages.json
		packages.Updated = time.Now().Format(time.RFC3339)
		packages.Version = "1.0.0"

		updatedJSON, err := json.MarshalIndent(packages, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling packages: %v\n", err)
			os.Exit(1)
		}

		// Write to file
		if err := os.WriteFile("packages.json", updatedJSON, 0644); err != nil {
			fmt.Printf("Error writing packages.json: %v\n", err)
			os.Exit(1)
		}

		// Push to GitHub
		if err := pushToGitHub(repo, token, string(updatedJSON)); err != nil {
			fmt.Printf("Error pushing to GitHub: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Updated packages.json with %d package changes and pushed to GitHub\n", len(packagesToUpdate))
	} else {
		fmt.Println("No deleted releases found")
	}
}

func getAllReleases(repo, token string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get releases: status %d", resp.StatusCode)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	var tags []string
	for _, release := range releases {
		tags = append(tags, release.TagName)
	}

	return tags, nil
}

func getCurrentPackages(repo string) (*struct {
	Packages map[string]interface{} `json:"packages"`
	Updated  string                 `json:"updated"`
	Version  string                 `json:"version"`
}, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/packages.json", repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get packages.json: status %d", resp.StatusCode)
	}

	var packages struct {
		Packages map[string]interface{} `json:"packages"`
		Updated  string                 `json:"updated"`
		Version  string                 `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, err
	}

	return &packages, nil
}

func extractTagFromURL(url string) string {
	// Extract tag from URL like: https://github.com/mburakmmm/wing-packages/releases/download/pkg-0.5.0/wing-real-publish-test-0.5.0.tar.gz
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "download" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func releaseExists(tag string, releases []string) bool {
	for _, release := range releases {
		if release == tag {
			return true
		}
	}
	return false
}

func pushToGitHub(repo, token, content string) error {
	// Get file SHA
	fileURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/packages.json", repo)
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create file request: %v", err)
	}

	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
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
		"message": "Cleanup deleted releases from registry",
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

	updateReq.Header.Set("Authorization", "token "+token)
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := client.Do(updateReq)
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
