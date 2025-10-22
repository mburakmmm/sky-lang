package wing

import (
	"fmt"
	"sort"
	"strings"
)

// DependencyResolver handles dependency resolution and conflict resolution
type DependencyResolver struct {
	registry RegistryClient
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(registry RegistryClient) *DependencyResolver {
	return &DependencyResolver{
		registry: registry,
	}
}

// ResolveDependencies resolves all dependencies for a package
func (dr *DependencyResolver) ResolveDependencies(manifest *PackageManifest) (*ResolutionResult, error) {
	resolver := &resolutionState{
		resolved:   make(map[string]string),
		conflicts:  make(map[string][]string),
		registry:   dr.registry,
		processing: make(map[string]bool),
	}

	// Add direct dependencies
	for name, version := range manifest.Dependencies {
		if err := resolver.addDependency(name, version); err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s@%s: %v", name, version, err)
		}
	}

	// Add dev dependencies
	for name, version := range manifest.DevDependencies {
		if err := resolver.addDependency(name, version); err != nil {
			return nil, fmt.Errorf("failed to resolve dev dependency %s@%s: %v", name, version, err)
		}
	}

	return &ResolutionResult{
		Dependencies:   resolver.resolved,
		Conflicts:      resolver.conflicts,
		ResolutionTree: resolver.buildResolutionTree(),
	}, nil
}

// ResolutionResult represents the result of dependency resolution
type ResolutionResult struct {
	Dependencies   map[string]string          `json:"dependencies"`
	Conflicts      map[string][]string        `json:"conflicts"`
	ResolutionTree map[string]*DependencyNode `json:"resolution_tree"`
}

// DependencyNode represents a node in the dependency tree
type DependencyNode struct {
	Name         string                     `json:"name"`
	Version      string                     `json:"version"`
	Dependencies map[string]*DependencyNode `json:"dependencies"`
	Resolved     bool                       `json:"resolved"`
	Conflict     bool                       `json:"conflict"`
}

type resolutionState struct {
	resolved   map[string]string
	conflicts  map[string][]string
	registry   RegistryClient
	processing map[string]bool
}

func (rs *resolutionState) addDependency(name, version string) error {
	// Check for circular dependencies
	if rs.processing[name] {
		return fmt.Errorf("circular dependency detected: %s", name)
	}

	// If already resolved, check for conflicts
	if existingVersion, exists := rs.resolved[name]; exists {
		if existingVersion != version {
			rs.conflicts[name] = append(rs.conflicts[name], existingVersion, version)
			return fmt.Errorf("version conflict for %s: %s vs %s", name, existingVersion, version)
		}
		return nil
	}

	// Mark as processing
	rs.processing[name] = true
	defer delete(rs.processing, name)

	// Resolve version if needed
	if version == "latest" {
		latestVersion, err := rs.registry.GetLatestVersion(name)
		if err != nil {
			return fmt.Errorf("failed to get latest version for %s: %v", name, err)
		}
		version = latestVersion
	}

	// Add to resolved
	rs.resolved[name] = version

	// Get package dependencies
	deps, err := rs.registry.GetPackageDependencies(name, version)
	if err != nil {
		return fmt.Errorf("failed to get dependencies for %s@%s: %v", name, version, err)
	}

	// Recursively resolve dependencies
	for depName, depVersion := range deps {
		if err := rs.addDependency(depName, depVersion); err != nil {
			return err
		}
	}

	return nil
}

func (rs *resolutionState) buildResolutionTree() map[string]*DependencyNode {
	tree := make(map[string]*DependencyNode)

	for name, version := range rs.resolved {
		tree[name] = rs.buildNode(name, version)
	}

	return tree
}

func (rs *resolutionState) buildNode(name, version string) *DependencyNode {
	node := &DependencyNode{
		Name:         name,
		Version:      version,
		Dependencies: make(map[string]*DependencyNode),
		Resolved:     true,
		Conflict:     len(rs.conflicts[name]) > 0,
	}

	// Get package dependencies
	deps, err := rs.registry.GetPackageDependencies(name, version)
	if err != nil {
		return node
	}

	// Build child nodes
	for depName := range deps {
		if resolvedVersion, exists := rs.resolved[depName]; exists {
			node.Dependencies[depName] = rs.buildNode(depName, resolvedVersion)
		}
	}

	return node
}

// VersionComparator compares semantic versions
type VersionComparator struct{}

// Compare compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (vc *VersionComparator) Compare(v1, v2 string) int {
	parts1 := vc.parseVersion(v1)
	parts2 := vc.parseVersion(v2)

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int

		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 < p2 {
			return -1
		} else if p1 > p2 {
			return 1
		}
	}

	return 0
}

// parseVersion parses a semantic version string
func (vc *VersionComparator) parseVersion(version string) []int {
	// Remove 'v' prefix if present
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	// Split by dots
	parts := strings.Split(version, ".")
	var result []int

	for _, part := range parts {
		// Remove any non-numeric suffix (e.g., "1.0.0-beta", "1.0.0+123")
		if dashIndex := strings.Index(part, "-"); dashIndex != -1 {
			part = part[:dashIndex]
		}
		if plusIndex := strings.Index(part, "+"); plusIndex != -1 {
			part = part[:plusIndex]
		}

		// Convert to int
		var num int
		fmt.Sscanf(part, "%d", &num)
		result = append(result, num)
	}

	return result
}

// ResolveConflicts resolves version conflicts using various strategies
func (dr *DependencyResolver) ResolveConflicts(result *ResolutionResult, strategy ConflictResolutionStrategy) (*ResolutionResult, error) {
	if len(result.Conflicts) == 0 {
		return result, nil
	}

	resolved := &ResolutionResult{
		Dependencies:   make(map[string]string),
		Conflicts:      make(map[string][]string),
		ResolutionTree: make(map[string]*DependencyNode),
	}

	// Copy resolved dependencies
	for name, version := range result.Dependencies {
		resolved.Dependencies[name] = version
	}

	// Resolve conflicts
	for name, versions := range result.Conflicts {
		chosenVersion, err := strategy.ResolveConflict(name, versions)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve conflict for %s: %v", name, err)
		}

		resolved.Dependencies[name] = chosenVersion
	}

	// Rebuild resolution tree
	resolver := &resolutionState{
		resolved:   resolved.Dependencies,
		conflicts:  make(map[string][]string),
		registry:   dr.registry,
		processing: make(map[string]bool),
	}

	resolved.ResolutionTree = resolver.buildResolutionTree()

	return resolved, nil
}

// ConflictResolutionStrategy defines strategies for resolving version conflicts
type ConflictResolutionStrategy interface {
	ResolveConflict(name string, versions []string) (string, error)
}

// LatestVersionStrategy resolves conflicts by choosing the latest version
type LatestVersionStrategy struct {
	comparator *VersionComparator
}

// NewLatestVersionStrategy creates a new latest version strategy
func NewLatestVersionStrategy() *LatestVersionStrategy {
	return &LatestVersionStrategy{
		comparator: &VersionComparator{},
	}
}

// ResolveConflict resolves conflicts by choosing the latest version
func (lvs *LatestVersionStrategy) ResolveConflict(name string, versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions provided")
	}

	// Remove duplicates
	uniqueVersions := make(map[string]bool)
	for _, version := range versions {
		uniqueVersions[version] = true
	}

	var versionList []string
	for version := range uniqueVersions {
		versionList = append(versionList, version)
	}

	// Sort versions (latest first)
	sort.Slice(versionList, func(i, j int) bool {
		return lvs.comparator.Compare(versionList[i], versionList[j]) > 0
	})

	return versionList[0], nil
}

// ConservativeStrategy resolves conflicts by choosing the earliest version
type ConservativeStrategy struct {
	comparator *VersionComparator
}

// NewConservativeStrategy creates a new conservative strategy
func NewConservativeStrategy() *ConservativeStrategy {
	return &ConservativeStrategy{
		comparator: &VersionComparator{},
	}
}

// ResolveConflict resolves conflicts by choosing the earliest version
func (cs *ConservativeStrategy) ResolveConflict(name string, versions []string) (string, error) {
	if len(versions) == 0 {
		return "", fmt.Errorf("no versions provided")
	}

	// Remove duplicates
	uniqueVersions := make(map[string]bool)
	for _, version := range versions {
		uniqueVersions[version] = true
	}

	var versionList []string
	for version := range uniqueVersions {
		versionList = append(versionList, version)
	}

	// Sort versions (earliest first)
	sort.Slice(versionList, func(i, j int) bool {
		return cs.comparator.Compare(versionList[i], versionList[j]) < 0
	})

	return versionList[0], nil
}

// CustomStrategy resolves conflicts using a custom function
type CustomStrategy struct {
	resolver func(name string, versions []string) (string, error)
}

// NewCustomStrategy creates a new custom strategy
func NewCustomStrategy(resolver func(name string, versions []string) (string, error)) *CustomStrategy {
	return &CustomStrategy{
		resolver: resolver,
	}
}

// ResolveConflict resolves conflicts using the custom resolver
func (cs *CustomStrategy) ResolveConflict(name string, versions []string) (string, error) {
	return cs.resolver(name, versions)
}
