package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// VendorMode manages offline package vendoring
type VendorMode struct {
	rootDir   string
	vendorDir string
}

// NewVendorMode creates a new vendor mode manager
func NewVendorMode(rootDir string) *VendorMode {
	return &VendorMode{
		rootDir:   rootDir,
		vendorDir: filepath.Join(rootDir, "vendor"),
	}
}

// Vendor copies all dependencies to vendor directory
func (vm *VendorMode) Vendor() error {
	// Create vendor directory
	if err := os.MkdirAll(vm.vendorDir, 0755); err != nil {
		return err
	}

	// Load lockfile
	lockfile, err := LoadLockfile(filepath.Join(vm.rootDir, "wing.lock"))
	if err != nil {
		return err
	}

	// Copy each package
	for name, pkg := range lockfile.Packages {
		src := pkg.Resolved
		dst := filepath.Join(vm.vendorDir, name)

		if err := vm.copyPackage(src, dst); err != nil {
			return fmt.Errorf("failed to vendor %s: %v", name, err)
		}

		fmt.Printf("Vendored: %s@%s\n", name, pkg.Version)
	}

	return nil
}

// copyPackage copies a package directory
func (vm *VendorMode) copyPackage(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Walk source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return vm.copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func (vm *VendorMode) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// IsVendored checks if vendor directory exists
func (vm *VendorMode) IsVendored() bool {
	info, err := os.Stat(vm.vendorDir)
	return err == nil && info.IsDir()
}

// Clean removes the vendor directory
func (vm *VendorMode) Clean() error {
	return os.RemoveAll(vm.vendorDir)
}
