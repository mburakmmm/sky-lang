package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
)

// Lockfile represents wing.lock
type Lockfile struct {
	Version  string                    `json:"version"`
	Packages map[string]*LockedPackage `json:"packages"`
}

// LockedPackage represents a locked package version
type LockedPackage struct {
	Version      string            `json:"version"`
	Resolved     string            `json:"resolved"` // URL or path
	Checksum     string            `json:"checksum"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// NewLockfile creates a new lockfile
func NewLockfile() *Lockfile {
	return &Lockfile{
		Version:  "1.0",
		Packages: make(map[string]*LockedPackage),
	}
}

// AddPackage adds a package to the lockfile
func (lf *Lockfile) AddPackage(name, version, resolved string, deps map[string]string) error {
	// Calculate checksum of resolved file
	checksum, err := calculateChecksum(resolved)
	if err != nil {
		return err
	}

	lf.Packages[name] = &LockedPackage{
		Version:      version,
		Resolved:     resolved,
		Checksum:     checksum,
		Dependencies: deps,
	}

	return nil
}

// Save saves the lockfile to disk
func (lf *Lockfile) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(lf)
}

// Load loads a lockfile from disk
func LoadLockfile(path string) (*Lockfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lf Lockfile
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&lf)
	return &lf, err
}

// Verify verifies all checksums in the lockfile
func (lf *Lockfile) Verify() error {
	for name, pkg := range lf.Packages {
		checksum, err := calculateChecksum(pkg.Resolved)
		if err != nil {
			return err
		}

		if checksum != pkg.Checksum {
			return &ChecksumMismatchError{
				Package:  name,
				Expected: pkg.Checksum,
				Actual:   checksum,
			}
		}
	}

	return nil
}

// ChecksumMismatchError represents a checksum verification failure
type ChecksumMismatchError struct {
	Package  string
	Expected string
	Actual   string
}

func (e *ChecksumMismatchError) Error() string {
	return "checksum mismatch for " + e.Package + ": expected " + e.Expected + ", got " + e.Actual
}

// calculateChecksum calculates SHA-256 checksum of a file
func calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
