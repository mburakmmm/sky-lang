package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// GitHubWebhookPayload represents the webhook payload from GitHub
type GitHubWebhookPayload struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		Body        string `json:"body"`
		CreatedAt   string `json:"created_at"`
		PublishedAt string `json:"published_at"`
	} `json:"release"`
}

func main() {
	http.HandleFunc("/webhook", handleWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Webhook server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify GitHub webhook signature (optional but recommended)
	if !verifyGitHubSignature(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read the payload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	// Parse the webhook payload
	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Only process release events
	if !strings.Contains(r.Header.Get("X-GitHub-Event"), "release") {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process the webhook
	switch payload.Action {
	case "deleted":
		fmt.Printf("Release deleted: %s\n", payload.Release.TagName)
		if err := handleReleaseDeleted(payload); err != nil {
			fmt.Printf("Error handling release deletion: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	case "published":
		fmt.Printf("Release published: %s\n", payload.Release.TagName)
		// Optionally handle release published events
	case "edited":
		fmt.Printf("Release edited: %s\n", payload.Release.TagName)
		// Optionally handle release edited events
	default:
		fmt.Printf("Unhandled release action: %s\n", payload.Action)
	}

	w.WriteHeader(http.StatusOK)
}

func verifyGitHubSignature(r *http.Request) bool {
	// Get the signature from the header
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		return false
	}

	// Get the webhook secret from environment
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		// If no secret is configured, allow all requests (not recommended for production)
		return true
	}

	// TODO: Implement proper signature verification
	// This would involve computing HMAC-SHA256 of the request body
	// and comparing it with the signature header

	return true // For now, always return true
}

func handleReleaseDeleted(payload GitHubWebhookPayload) error {
	// Check if this is a package release (starts with "pkg-")
	if !strings.HasPrefix(payload.Release.TagName, "pkg-") {
		fmt.Printf("Not a package release, skipping: %s\n", payload.Release.TagName)
		return nil
	}

	// Extract package name from release name
	// Format: "wing-package-name version (Package)"
	releaseName := payload.Release.Name
	if !strings.Contains(releaseName, " (Package)") {
		fmt.Printf("Not a package release, skipping: %s\n", releaseName)
		return nil
	}

	// Extract package name
	packageName := strings.TrimSuffix(releaseName, " (Package)")
	packageName = strings.TrimPrefix(packageName, "wing-")

	// Extract version from tag
	version := strings.TrimPrefix(payload.Release.TagName, "pkg-")

	fmt.Printf("Removing package %s@%s from registry\n", packageName, version)

	// Update packages.json
	return updatePackagesRegistry(packageName, version, "delete")
}

func updatePackagesRegistry(packageName, version, action string) error {
	// Get GitHub token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN not set")
	}

	// Get repository info
	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		repo = "mburakmmm/wing-packages"
	}

	// Download current packages.json
	packagesURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/packages.json", repo)

	client := &http.Client{}
	req, err := http.NewRequest("GET", packagesURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Do(req)
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

	// Process the action
	switch action {
	case "delete":
		if pkg, exists := registry.Packages[packageName]; exists {
			if packageMap, ok := pkg.(map[string]interface{}); ok {
				// Remove the specific version
				if versions, ok := packageMap["versions"].(map[string]interface{}); ok {
					delete(versions, version)

					// If no versions left, remove the entire package
					if len(versions) == 0 {
						delete(registry.Packages, packageName)
						fmt.Printf("Removed entire package: %s\n", packageName)
					} else {
						// Update latest version to the highest remaining version
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
	}

	// Update registry metadata
	registry.Updated = fmt.Sprintf("%d", time.Now().Unix())
	registry.Version = "1.0.0"

	// Convert back to JSON
	updatedJSON, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated packages.json: %v", err)
	}

	// Update packages.json in repository
	return updateFileInRepository(repo, "packages.json", string(updatedJSON), fmt.Sprintf("Remove package %s@%s from registry", packageName, version), token)
}

func updateFileInRepository(repo, filePath, content, commitMessage, token string) error {
	// Get file SHA
	fileURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, filePath)
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
