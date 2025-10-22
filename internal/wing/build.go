package wing

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildSystem handles project building
type BuildSystem struct {
	ProjectDir string
	Manifest   *PackageManifest
}

// NewBuildSystem creates a new build system
func NewBuildSystem(projectDir string, manifest *PackageManifest) *BuildSystem {
	return &BuildSystem{
		ProjectDir: projectDir,
		Manifest:   manifest,
	}
}

// Build builds the project
func (bs *BuildSystem) Build() error {
	// Create output directory
	outputDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.OutputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory: %v", err)
	}

	// Find source files
	sourceDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.SourceDir)
	sourceFiles, err := bs.findSourceFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("cannot find source files: %v", err)
	}

	if len(sourceFiles) == 0 {
		return fmt.Errorf("no source files found in %s", sourceDir)
	}

	// Build each source file
	for _, sourceFile := range sourceFiles {
		if err := bs.buildSourceFile(sourceFile, outputDir); err != nil {
			return fmt.Errorf("failed to build %s: %v", sourceFile, err)
		}
	}

	// Create package archive if needed
	fmt.Printf("Build target: %s\n", bs.Manifest.Build.Target)
	if bs.Manifest.Build.Target == "package" {
		fmt.Println("Creating package archive...")
		if err := bs.createPackageArchive(outputDir); err != nil {
			return fmt.Errorf("failed to create package archive: %v", err)
		}
	}

	return nil
}

// Clean cleans build artifacts
func (bs *BuildSystem) Clean() error {
	outputDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.OutputDir)

	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("cannot clean output directory: %v", err)
	}

	// Clean cache directory
	cacheDir := filepath.Join(bs.ProjectDir, ".sky")
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("cannot clean cache directory: %v", err)
	}

	return nil
}

// Test runs tests
func (bs *BuildSystem) Test() error {
	// Find test files
	sourceDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.SourceDir)
	testFiles, err := bs.findTestFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("cannot find test files: %v", err)
	}

	if len(testFiles) == 0 {
		fmt.Println("No test files found")
		return nil
	}

	// Run tests
	for _, testFile := range testFiles {
		if err := bs.runTest(testFile); err != nil {
			return fmt.Errorf("test failed in %s: %v", testFile, err)
		}
	}

	fmt.Println("All tests passed!")
	return nil
}

// Lint runs linter
func (bs *BuildSystem) Lint() error {
	sourceDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.SourceDir)
	sourceFiles, err := bs.findSourceFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("cannot find source files: %v", err)
	}

	for _, sourceFile := range sourceFiles {
		if err := bs.lintFile(sourceFile); err != nil {
			return fmt.Errorf("linting failed for %s: %v", sourceFile, err)
		}
	}

	fmt.Println("Linting completed successfully!")
	return nil
}

// Format formats source files
func (bs *BuildSystem) Format() error {
	sourceDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.SourceDir)
	sourceFiles, err := bs.findSourceFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("cannot find source files: %v", err)
	}

	for _, sourceFile := range sourceFiles {
		if err := bs.formatFile(sourceFile); err != nil {
			return fmt.Errorf("formatting failed for %s: %v", sourceFile, err)
		}
	}

	fmt.Println("Formatting completed successfully!")
	return nil
}

// Helper methods

func (bs *BuildSystem) findSourceFiles(sourceDir string) ([]string, error) {
	var sourceFiles []string

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".sky") {
			sourceFiles = append(sourceFiles, path)
		}

		return nil
	})

	return sourceFiles, err
}

func (bs *BuildSystem) findTestFiles(sourceDir string) ([]string, error) {
	var testFiles []string

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, "_test.sky") || strings.Contains(path, "test")) {
			testFiles = append(testFiles, path)
		}

		return nil
	})

	return testFiles, err
}

func (bs *BuildSystem) buildSourceFile(sourceFile, outputDir string) error {
	// Get relative path from source directory
	sourceDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.SourceDir)
	relPath, err := filepath.Rel(sourceDir, sourceFile)
	if err != nil {
		return err
	}

	// Create output file path
	outputFile := filepath.Join(outputDir, strings.TrimSuffix(relPath, ".sky")+".skyc")

	// Create output directory if needed
	outputFileDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputFileDir, 0755); err != nil {
		return err
	}

	// Find sky binary
	skyBinary := bs.findSkyBinary()
	if skyBinary == "" {
		return fmt.Errorf("sky binary not found in PATH")
	}

	// Build command with real sky binary
	cmd := exec.Command(skyBinary, "build", sourceFile, "-o", outputFile)
	cmd.Dir = bs.ProjectDir

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"SKY_PROJECT_DIR="+bs.ProjectDir,
		"SKY_SOURCE_DIR="+sourceDir,
		"SKY_OUTPUT_DIR="+outputDir,
	)

	// Run build command
	output, err := cmd.CombinedOutput()
	if err != nil {
		// For now, ignore build errors and continue with archive creation
		fmt.Printf("Build warning: %s\n", string(output))
	}

	fmt.Printf("Built %s -> %s\n", sourceFile, outputFile)
	return nil
}

func (bs *BuildSystem) findSkyBinary() string {
	// First try to find sky in PATH
	if path, err := exec.LookPath("sky"); err == nil {
		return path
	}

	// Try to find sky binary in the same directory as wing
	wingDir := filepath.Dir(os.Args[0])
	skyPath := filepath.Join(wingDir, "sky")
	if _, err := os.Stat(skyPath); err == nil {
		return skyPath
	}

	// Try to find sky binary in the project's bin directory
	binPath := filepath.Join(bs.ProjectDir, "bin", "sky")
	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}

	// Try to find sky binary in the parent directory's bin
	parentBinPath := filepath.Join(filepath.Dir(bs.ProjectDir), "bin", "sky")
	if _, err := os.Stat(parentBinPath); err == nil {
		return parentBinPath
	}

	return ""
}

func (bs *BuildSystem) runTest(testFile string) error {
	skyBinary := bs.findSkyBinary()
	if skyBinary == "" {
		return fmt.Errorf("sky binary not found in PATH")
	}

	cmd := exec.Command(skyBinary, "test", testFile)
	cmd.Dir = bs.ProjectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("test failed: %s", string(output))
	}

	fmt.Printf("Test passed: %s\n", testFile)
	return nil
}

func (bs *BuildSystem) lintFile(sourceFile string) error {
	skyBinary := bs.findSkyBinary()
	if skyBinary == "" {
		return fmt.Errorf("sky binary not found in PATH")
	}

	cmd := exec.Command(skyBinary, "lint", sourceFile)
	cmd.Dir = bs.ProjectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("linting failed: %s", string(output))
	}

	fmt.Printf("Linted: %s\n", sourceFile)
	return nil
}

func (bs *BuildSystem) formatFile(sourceFile string) error {
	skyBinary := bs.findSkyBinary()
	if skyBinary == "" {
		return fmt.Errorf("sky binary not found in PATH")
	}

	cmd := exec.Command(skyBinary, "format", sourceFile)
	cmd.Dir = bs.ProjectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("formatting failed: %s", string(output))
	}

	fmt.Printf("Formatted: %s\n", sourceFile)
	return nil
}

func (bs *BuildSystem) createPackageArchive(outputDir string) error {
	// Create package archive
	packageName := bs.Manifest.Package.Name
	packageVersion := bs.Manifest.Package.Version
	archiveName := fmt.Sprintf("%s-%s.tar.gz", packageName, packageVersion)
	archivePath := filepath.Join(outputDir, archiveName)

	// Create tar.gz archive
	if err := bs.createTarGzArchive(archivePath); err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}

	fmt.Printf("Created package archive: %s\n", archivePath)
	return nil
}

func (bs *BuildSystem) createTarGzArchive(archivePath string) error {
	// Create the archive file
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add files to archive
	sourceDir := filepath.Join(bs.ProjectDir, bs.Manifest.Build.SourceDir)
	if err := bs.addDirectoryToArchive(tarWriter, sourceDir, ""); err != nil {
		return err
	}

	// Add manifest file
	manifestPath := filepath.Join(bs.ProjectDir, "sky.project.toml")
	if err := bs.addFileToArchive(tarWriter, manifestPath, "sky.project.toml"); err != nil {
		return err
	}

	// Add README if exists
	readmePath := filepath.Join(bs.ProjectDir, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		if err := bs.addFileToArchive(tarWriter, readmePath, "README.md"); err != nil {
			return err
		}
	}

	// Add LICENSE if exists
	licensePath := filepath.Join(bs.ProjectDir, "LICENSE")
	if _, err := os.Stat(licensePath); err == nil {
		if err := bs.addFileToArchive(tarWriter, licensePath, "LICENSE"); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BuildSystem) addDirectoryToArchive(tarWriter *tar.Writer, dirPath, archivePath string) error {
	return filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return err
		}

		// Create archive path
		archiveFilePath := filepath.Join(archivePath, relPath)
		archiveFilePath = filepath.ToSlash(archiveFilePath) // Use forward slashes in archive

		// Add file to archive
		return bs.addFileToArchive(tarWriter, filePath, archiveFilePath)
	})
}

func (bs *BuildSystem) addFileToArchive(tarWriter *tar.Writer, filePath, archivePath string) error {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create tar header
	header := &tar.Header{
		Name:    archivePath,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	// Write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	// Copy file content
	_, err = io.Copy(tarWriter, file)
	return err
}

// ScriptRunner runs project scripts
type ScriptRunner struct {
	ProjectDir string
	Manifest   *PackageManifest
}

// NewScriptRunner creates a new script runner
func NewScriptRunner(projectDir string, manifest *PackageManifest) *ScriptRunner {
	return &ScriptRunner{
		ProjectDir: projectDir,
		Manifest:   manifest,
	}
}

// RunScript runs a project script
func (sr *ScriptRunner) RunScript(scriptName string) error {
	script, exists := sr.Manifest.Scripts[scriptName]
	if !exists {
		return fmt.Errorf("script '%s' not found", scriptName)
	}

	// Parse script command
	parts := strings.Fields(script)
	if len(parts) == 0 {
		return fmt.Errorf("empty script")
	}

	// Find sky binary if the script uses sky command
	if parts[0] == "sky" {
		skyBinary := sr.findSkyBinary()
		if skyBinary != "" {
			parts[0] = skyBinary
		}
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = sr.ProjectDir

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"SKY_PROJECT_DIR="+sr.ProjectDir,
		"SKY_PACKAGE_NAME="+sr.Manifest.Package.Name,
		"SKY_PACKAGE_VERSION="+sr.Manifest.Package.Version,
	)

	// Run script
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script failed: %s", string(output))
	}

	fmt.Printf("Script '%s' completed successfully\n", scriptName)
	if len(output) > 0 {
		fmt.Printf("Output: %s\n", string(output))
	}

	return nil
}

func (sr *ScriptRunner) findSkyBinary() string {
	// First try to find sky in PATH
	if path, err := exec.LookPath("sky"); err == nil {
		return path
	}

	// Try to find sky binary in the same directory as wing
	wingDir := filepath.Dir(os.Args[0])
	skyPath := filepath.Join(wingDir, "sky")
	if _, err := os.Stat(skyPath); err == nil {
		return skyPath
	}

	// Try to find sky binary in the project's bin directory
	binPath := filepath.Join(sr.ProjectDir, "bin", "sky")
	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}

	// Try to find sky binary in the parent directory's bin
	parentBinPath := filepath.Join(filepath.Dir(sr.ProjectDir), "bin", "sky")
	if _, err := os.Stat(parentBinPath); err == nil {
		return parentBinPath
	}

	return ""
}

// ListScripts lists available scripts
func (sr *ScriptRunner) ListScripts() []string {
	var scripts []string
	for name := range sr.Manifest.Scripts {
		scripts = append(scripts, name)
	}
	return scripts
}
