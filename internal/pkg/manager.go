package pkg

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
	"time"
)

// Manager paket yöneticisi
type Manager struct {
	config      *Config
	registry    *Registry
	cache       *Cache
	installed   map[string]*Package
	installedMu sync.RWMutex
}

// Config paket yöneticisi ayarları
type Config struct {
	RegistryURL  string
	CacheDir     string
	InstallDir   string
	GlobalCache  bool
	OfflineMode  bool
	ParallelJobs int
}

// DefaultConfig varsayılan ayarlar
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		RegistryURL:  "https://registry.sky-lang.org",
		CacheDir:     filepath.Join(homeDir, ".sky", "cache"),
		InstallDir:   filepath.Join(homeDir, ".sky", "packages"),
		GlobalCache:  true,
		OfflineMode:  false,
		ParallelJobs: 4,
	}
}

// Package paket bilgisi
type Package struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Author          string            `json:"author"`
	License         string            `json:"license"`
	Homepage        string            `json:"homepage"`
	Repository      string            `json:"repository"`
	Keywords        []string          `json:"keywords"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Main            string            `json:"main"`
	Files           []string          `json:"files"`
	Scripts         map[string]string `json:"scripts"`
	Checksum        string            `json:"checksum,omitempty"`
}

// Manifest sky.project.toml manifestosu
type Manifest struct {
	Package         PackageInfo       `json:"package"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"dev_dependencies"`
	Build           BuildConfig       `json:"build"`
}

// PackageInfo paket metadata
type PackageInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Authors     []string `json:"authors"`
	License     string   `json:"license"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	Keywords    []string `json:"keywords"`
}

// BuildConfig build ayarları
type BuildConfig struct {
	Target       string          `json:"target"`
	Output       string          `json:"output"`
	Optimization int             `json:"optimization"`
	Features     map[string]bool `json:"features"`
}

// NewManager yeni bir paket yöneticisi oluşturur
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	// Create directories
	os.MkdirAll(config.CacheDir, 0755)
	os.MkdirAll(config.InstallDir, 0755)

	return &Manager{
		config:    config,
		registry:  NewRegistry(config.RegistryURL),
		cache:     NewCache(config.CacheDir),
		installed: make(map[string]*Package),
	}
}

// Install paketi kurar
func (m *Manager) Install(name, version string) error {
	// Version empty ise latest al
	if version == "" {
		version = "latest"
	}

	// Registry'den paket bilgisini al
	pkg, err := m.registry.GetPackage(name, version)
	if err != nil {
		return fmt.Errorf("failed to get package info: %w", err)
	}

	// Cache'de var mı kontrol et
	if !m.cache.Has(name, pkg.Version) {
		// Download
		if err := m.downloadPackage(pkg); err != nil {
			return fmt.Errorf("failed to download package: %w", err)
		}
	}

	// Extract ve install
	if err := m.extractPackage(pkg); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Dependencies kur
	if err := m.installDependencies(pkg); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Track installed
	m.installedMu.Lock()
	m.installed[name] = pkg
	m.installedMu.Unlock()

	return nil
}

// Update paketi günceller
func (m *Manager) Update(name string) error {
	// Latest versiyonu al
	latest, err := m.registry.GetLatestVersion(name)
	if err != nil {
		return err
	}

	return m.Install(name, latest)
}

// Uninstall paketi kaldırır
func (m *Manager) Uninstall(name string) error {
	installPath := filepath.Join(m.config.InstallDir, name)

	if err := os.RemoveAll(installPath); err != nil {
		return fmt.Errorf("failed to remove package: %w", err)
	}

	m.installedMu.Lock()
	delete(m.installed, name)
	m.installedMu.Unlock()

	return nil
}

// List kurulu paketleri listeler
func (m *Manager) List() []*Package {
	m.installedMu.RLock()
	defer m.installedMu.RUnlock()

	packages := make([]*Package, 0, len(m.installed))
	for _, pkg := range m.installed {
		packages = append(packages, pkg)
	}

	return packages
}

// downloadPackage paketi indirir
func (m *Manager) downloadPackage(pkg *Package) error {
	url := fmt.Sprintf("%s/packages/%s/%s.tar.gz",
		m.config.RegistryURL, pkg.Name, pkg.Version)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Save to cache
	cachePath := m.cache.Path(pkg.Name, pkg.Version)
	os.MkdirAll(filepath.Dir(cachePath), 0755)

	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Download with checksum verification
	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return err
	}

	// Verify checksum
	if pkg.Checksum != "" {
		checksum := hex.EncodeToString(hasher.Sum(nil))
		if checksum != pkg.Checksum {
			os.Remove(cachePath)
			return fmt.Errorf("checksum mismatch: expected %s, got %s",
				pkg.Checksum, checksum)
		}
	}

	return nil
}

// extractPackage paketi extract eder
func (m *Manager) extractPackage(pkg *Package) error {
	// TODO: Implement tar.gz extraction
	cachePath := m.cache.Path(pkg.Name, pkg.Version)
	installPath := filepath.Join(m.config.InstallDir, pkg.Name)

	os.MkdirAll(installPath, 0755)

	// Basit copy (gerçekte tar extraction)
	_, err := os.Stat(cachePath)
	return err
}

// installDependencies bağımlılıkları kurar
func (m *Manager) installDependencies(pkg *Package) error {
	if pkg.Dependencies == nil {
		return nil
	}

	// Parallel installation
	semaphore := make(chan struct{}, m.config.ParallelJobs)
	var wg sync.WaitGroup
	errCh := make(chan error, len(pkg.Dependencies))

	for name, version := range pkg.Dependencies {
		wg.Add(1)
		go func(n, v string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := m.Install(n, v); err != nil {
				errCh <- err
			}
		}(name, version)
	}

	wg.Wait()
	close(errCh)

	// Hataları topla
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadManifest sky.project.toml yükler
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// TODO: Parse TOML (şimdilik JSON)
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// SaveManifest manifest dosyasını kaydeder
func SaveManifest(path string, manifest *Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Registry paket registry client
type Registry struct {
	baseURL string
	client  *http.Client
	cache   sync.Map // package metadata cache
}

// NewRegistry yeni bir registry client oluşturur
func NewRegistry(baseURL string) *Registry {
	return &Registry{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPackage paket bilgisini alır
func (r *Registry) GetPackage(name, version string) (*Package, error) {
	cacheKey := fmt.Sprintf("%s@%s", name, version)

	// Cache'de var mı?
	if cached, ok := r.cache.Load(cacheKey); ok {
		return cached.(*Package), nil
	}

	// Registry'den al
	url := fmt.Sprintf("%s/api/packages/%s/%s", r.baseURL, name, version)
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("package not found: %s@%s", name, version)
	}

	var pkg Package
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, err
	}

	// Cache'e kaydet
	r.cache.Store(cacheKey, &pkg)

	return &pkg, nil
}

// GetLatestVersion en son versiyonu alır
func (r *Registry) GetLatestVersion(name string) (string, error) {
	url := fmt.Sprintf("%s/api/packages/%s/latest", r.baseURL, name)
	resp, err := r.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("package not found: %s", name)
	}

	var result struct {
		Version string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Version, nil
}

// Search paket arar
func (r *Registry) Search(query string) ([]*Package, error) {
	url := fmt.Sprintf("%s/api/search?q=%s", r.baseURL, query)
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results struct {
		Packages []*Package `json:"packages"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results.Packages, nil
}

// Publish paketi registry'ye yükler
func (r *Registry) Publish(pkg *Package, tarballPath string) error {
	// TODO: Implement multipart upload
	return fmt.Errorf("publish not yet implemented")
}

// Cache paket cache yöneticisi
type Cache struct {
	dir string
	mu  sync.RWMutex
}

// NewCache yeni bir cache oluşturur
func NewCache(dir string) *Cache {
	os.MkdirAll(dir, 0755)
	return &Cache{dir: dir}
}

// Path cache path döndürür
func (c *Cache) Path(name, version string) string {
	return filepath.Join(c.dir, name, version+".tar.gz")
}

// Has cache'de var mı kontrol eder
func (c *Cache) Has(name, version string) bool {
	path := c.Path(name, version)
	_, err := os.Stat(path)
	return err == nil
}

// Clear cache'i temizler
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}

// Size cache boyutunu döndürür
func (c *Cache) Size() (int64, error) {
	var size int64

	err := filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}
