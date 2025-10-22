package wing

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PackageValidator handles package validation
type PackageValidator struct {
	ProjectDir string
	Manifest   *PackageManifest
}

// NewPackageValidator creates a new package validator
func NewPackageValidator(projectDir string, manifest *PackageManifest) *PackageValidator {
	return &PackageValidator{
		ProjectDir: projectDir,
		Manifest:   manifest,
	}
}

// ValidationResult represents the result of package validation
type ValidationResult struct {
	Valid   bool
	Errors  []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error
type ValidationError struct {
	Type    string
	Message string
	File    string
	Line    int
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type    string
	Message string
	File    string
	Line    int
}

// Validate performs comprehensive package validation
func (pv *PackageValidator) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Validate manifest
	pv.validateManifest(result)

	// Validate source code
	pv.validateSourceCode(result)

	// Validate dependencies
	pv.validateDependencies(result)

	// Validate build configuration
	pv.validateBuildConfig(result)

	// Check for circular dependencies
	pv.validateCircularDependencies(result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// validateManifest validates the package manifest
func (pv *PackageValidator) validateManifest(result *ValidationResult) {
	manifest := pv.Manifest

	// Validate package name
	if manifest.Package.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "manifest",
			Message: "Package name is required",
			File:    "sky.project.toml",
		})
	} else if !pv.isValidPackageName(manifest.Package.Name) {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "manifest",
			Message: "Invalid package name. Must contain only lowercase letters, numbers, and hyphens",
			File:    "sky.project.toml",
		})
	}

	// Validate version
	if manifest.Package.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "manifest",
			Message: "Package version is required",
			File:    "sky.project.toml",
		})
	} else if !pv.isValidVersion(manifest.Package.Version) {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "manifest",
			Message: "Invalid version format. Must be semantic version (e.g., 1.0.0)",
			File:    "sky.project.toml",
		})
	}

	// Validate description
	if manifest.Package.Description == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "manifest",
			Message: "Package description is recommended",
			File:    "sky.project.toml",
		})
	}

	// Validate authors
	if len(manifest.Package.Authors) == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "manifest",
			Message: "Package authors are recommended",
			File:    "sky.project.toml",
		})
	}

	// Validate license
	if manifest.Package.License == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "manifest",
			Message: "Package license is recommended",
			File:    "sky.project.toml",
		})
	}
}

// validateSourceCode validates the source code
func (pv *PackageValidator) validateSourceCode(result *ValidationResult) {
	sourceDir := filepath.Join(pv.ProjectDir, pv.Manifest.Build.SourceDir)
	
	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "source",
			Message: fmt.Sprintf("Source directory '%s' does not exist", sourceDir),
			File:    "sky.project.toml",
		})
		return
	}

	// Find all .sky files
	skyFiles, err := pv.findSkyFiles(sourceDir)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "source",
			Message: fmt.Sprintf("Error scanning source directory: %v", err),
			File:    sourceDir,
		})
		return
	}

	if len(skyFiles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "source",
			Message: "No .sky files found in source directory",
			File:    sourceDir,
		})
		return
	}

	// Validate each .sky file
	for _, file := range skyFiles {
		pv.validateSkyFile(file, result)
	}

	// Check for main function
	if !pv.hasMainFunction(skyFiles) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Type:    "source",
			Message: "No main function found. Package may not be executable",
			File:    sourceDir,
		})
	}
}

// validateDependencies validates dependencies
func (pv *PackageValidator) validateDependencies(result *ValidationResult) {
	manifest := pv.Manifest

	// Validate dependency versions
	for name, version := range manifest.Dependencies {
		if !pv.isValidVersion(version) && version != "latest" {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "dependencies",
				Message: fmt.Sprintf("Invalid version '%s' for dependency '%s'", version, name),
				File:    "sky.project.toml",
			})
		}
	}

	// Validate dev dependencies
	for name, version := range manifest.DevDependencies {
		if !pv.isValidVersion(version) && version != "latest" {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "dev-dependencies",
				Message: fmt.Sprintf("Invalid version '%s' for dev dependency '%s'", version, name),
				File:    "sky.project.toml",
			})
		}
	}
}

// validateBuildConfig validates build configuration
func (pv *PackageValidator) validateBuildConfig(result *ValidationResult) {
	build := pv.Manifest.Build

	// Validate target
	validTargets := []string{"native", "package", "wasm"}
	if !pv.contains(validTargets, build.Target) {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "build",
			Message: fmt.Sprintf("Invalid build target '%s'. Must be one of: %s", build.Target, strings.Join(validTargets, ", ")),
			File:    "sky.project.toml",
		})
	}

	// Validate optimization level
	validOptimizations := []string{"debug", "release", "size"}
	if !pv.contains(validOptimizations, build.Optimization) {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "build",
			Message: fmt.Sprintf("Invalid optimization level '%s'. Must be one of: %s", build.Optimization, strings.Join(validOptimizations, ", ")),
			File:    "sky.project.toml",
		})
	}

	// Validate output directory
	if build.OutputDir == "" {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "build",
			Message: "Output directory is required",
			File:    "sky.project.toml",
		})
	}
}

// validateCircularDependencies checks for circular dependencies
func (pv *PackageValidator) validateCircularDependencies(result *ValidationResult) {
	// This would require a more sophisticated dependency graph analysis
	// For now, we'll do a simple check
	manifest := pv.Manifest
	
	// Check if package depends on itself
	for name := range manifest.Dependencies {
		if name == manifest.Package.Name {
			result.Errors = append(result.Errors, ValidationError{
				Type:    "dependencies",
				Message: "Package cannot depend on itself",
				File:    "sky.project.toml",
			})
		}
	}
}

// Helper methods

func (pv *PackageValidator) isValidPackageName(name string) bool {
	// Package names should be lowercase, contain only letters, numbers, and hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, name)
	return matched && len(name) > 0 && len(name) <= 50
}

func (pv *PackageValidator) isValidVersion(version string) bool {
	// Basic semantic version validation
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`, version)
	return matched
}

func (pv *PackageValidator) findSkyFiles(dir string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".sky") {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

func (pv *PackageValidator) validateSkyFile(filePath string, result *ValidationResult) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "source",
			Message: fmt.Sprintf("Cannot read file: %v", err),
			File:    filePath,
		})
		return
	}

	// Basic syntax validation
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		lineNum := i + 1
		
		// Check for common syntax issues
		if strings.Contains(line, "function ") && !strings.Contains(line, "end") {
			// Check if function has proper ending in subsequent lines
			hasEnd := false
			for j := i + 1; j < len(lines) && j < i+10; j++ {
				if strings.TrimSpace(lines[j]) == "end" {
					hasEnd = true
					break
				}
			}
			if !hasEnd {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:    "syntax",
					Message: "Function may be missing 'end' keyword",
					File:    filePath,
					Line:    lineNum,
				})
			}
		}
	}
}

func (pv *PackageValidator) hasMainFunction(files []string) bool {
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		if strings.Contains(string(content), "function main") {
			return true
		}
	}
	return false
}

func (pv *PackageValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
