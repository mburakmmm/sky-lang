package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// RegistryServer serves packages
type RegistryServer struct {
	root     string
	packages map[string]*PackageInfo
	mu       sync.RWMutex
}

// PackageInfo stores package metadata
type PackageInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Checksum     string   `json:"checksum"`
	Dependencies []string `json:"dependencies"`
}

// NewRegistryServer creates a new registry server
func NewRegistryServer(root string) *RegistryServer {
	return &RegistryServer{
		root:     root,
		packages: make(map[string]*PackageInfo),
	}
}

// Start starts the HTTP server
func (rs *RegistryServer) Start(addr string) error {
	http.HandleFunc("/packages", rs.handleList)
	http.HandleFunc("/package/", rs.handleGet)
	http.HandleFunc("/publish", rs.handlePublish)

	fmt.Printf("Registry server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleList lists all packages
func (rs *RegistryServer) handleList(w http.ResponseWriter, r *http.Request) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	packages := make([]*PackageInfo, 0, len(rs.packages))
	for _, pkg := range rs.packages {
		packages = append(packages, pkg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packages)
}

// handleGet downloads a package
func (rs *RegistryServer) handleGet(w http.ResponseWriter, r *http.Request) {
	pkgName := r.URL.Path[len("/package/"):]

	rs.mu.RLock()
	pkg, exists := rs.packages[pkgName]
	rs.mu.RUnlock()

	if !exists {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Send package tarball
	tarball := filepath.Join(rs.root, pkgName+".tar.gz")
	http.ServeFile(w, r, tarball)

	w.Header().Set("X-Package-Checksum", pkg.Checksum)
}

// handlePublish publishes a new package
func (rs *RegistryServer) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get package file
	file, header, err := r.FormFile("package")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Calculate checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Save package
	pkgPath := filepath.Join(rs.root, header.Filename)
	out, err := os.Create(pkgPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	file.Seek(0, 0) // Reset file pointer
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Parse package metadata and update registry

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "published",
		"checksum": checksum,
	})
}
