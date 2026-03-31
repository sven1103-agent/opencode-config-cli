// Package source provides functionality for managing OpenCode config sources.
package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// SourceType represents the type of a config source.
type SourceType string

const (
	// SourceTypeLocalDirectory represents a local directory source.
	SourceTypeLocalDirectory SourceType = "local-directory"
	// SourceTypeLocalArchive represents a local archive (.tar.gz) source.
	SourceTypeLocalArchive SourceType = "local-archive"
	// SourceTypeGitHubRelease represents a GitHub release source.
	SourceTypeGitHubRelease SourceType = "github-release"
)

// Source represents a registered config source.
type Source struct {
	ID        string     `json:"id"`
	Location  string     `json:"location"`
	Type      SourceType `json:"type"`
	Name      string     `json:"name"`
	CreatedAt string     `json:"created_at"`
}

// Registry represents the source registry file format.
type Registry struct {
	Version int      `json:"version"`
	Sources []Source `json:"sources"`
}

// RegistryPath returns the path to the source registry file.
func RegistryPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	return filepath.Join(configDir, "opencode-helper", "sources.json"), nil
}

// LoadRegistry loads the source registry from disk.
func LoadRegistry() (*Registry, error) {
	path, err := RegistryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Version: 1, Sources: []Source{}}, nil
		}
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	return &registry, nil
}

// SaveRegistry saves the source registry to disk.
func SaveRegistry(registry *Registry) error {
	path, err := RegistryPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// DetectSourceType detects the source type from a location path.
func DetectSourceType(location string) (SourceType, error) {
	info, err := os.Stat(location)
	if err != nil {
		return "", fmt.Errorf("location does not exist: %s", location)
	}

	if info.IsDir() {
		return SourceTypeLocalDirectory, nil
	}

	// Check for archive extensions
	ext := filepath.Ext(location)
	if ext == ".gz" || ext == ".tar" || location[len(location)-7:] == ".tar.gz" || location[len(location)-4:] == ".tgz" {
		return SourceTypeLocalArchive, nil
	}

	// Check if it's a GitHub URL or reference
	if isGitHubRef(location) {
		return SourceTypeGitHubRelease, nil
	}

	return SourceTypeLocalArchive, nil
}

// isGitHubRef checks if a location appears to be a GitHub reference.
func isGitHubRef(location string) bool {
	githubPrefixes := []string{
		"https://github.com/",
		"http://github.com/",
		"github.com/",
		"git@github.com:",
	}
	for _, prefix := range githubPrefixes {
		if len(location) >= len(prefix) && location[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// ValidateSource validates that a source location is accessible.
func ValidateSource(location string, sourceType SourceType) error {
	switch sourceType {
	case SourceTypeLocalDirectory:
		info, err := os.Stat(location)
		if err != nil {
			return fmt.Errorf("directory does not exist: %s", location)
		}
		if !info.IsDir() {
			return fmt.Errorf("location is not a directory: %s", location)
		}
		// Check for manifest file
		manifestPath := filepath.Join(location, "opencode-bundle.manifest.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return fmt.Errorf("bundle manifest not found: %s (run 'oc bundle init' to create one)", manifestPath)
		}
	case SourceTypeLocalArchive:
		info, err := os.Stat(location)
		if err != nil {
			return fmt.Errorf("archive does not exist: %s", location)
		}
		if info.IsDir() {
			return fmt.Errorf("archive must be a file, not a directory: %s", location)
		}
	case SourceTypeGitHubRelease:
		// GitHub sources are validated at bundle install time
		// For now, just check it's not empty
		if location == "" {
			return fmt.Errorf("GitHub source location cannot be empty")
		}
	default:
		return fmt.Errorf("unknown source type: %s", sourceType)
	}
	return nil
}

// AddSource adds a new source to the registry.
func AddSource(location string, name string) (*Source, error) {
	// Detect source type
	sourceType, err := DetectSourceType(location)
	if err != nil {
		return nil, fmt.Errorf("failed to detect source type: %w", err)
	}

	// Validate the source
	if err := ValidateSource(location, sourceType); err != nil {
		return nil, err
	}

	// Load existing registry
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	// Check for duplicate locations
	for _, s := range registry.Sources {
		if s.Location == location {
			return nil, fmt.Errorf("source already registered at location: %s (id: %s)", location, s.ID)
		}
	}

	// Generate ID if name not provided
	id := uuid.New().String()[:8]
	if name == "" {
		name = filepath.Base(location)
	}

	// Create new source
	source := Source{
		ID:        id,
		Location:  location,
		Type:      sourceType,
		Name:      name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	registry.Sources = append(registry.Sources, source)

	if err := SaveRegistry(registry); err != nil {
		return nil, err
	}

	return &source, nil
}

// ListSources returns all registered sources.
func ListSources() ([]Source, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}
	return registry.Sources, nil
}

// RemoveSource removes a source from the registry by ID.
func RemoveSource(id string) error {
	registry, err := LoadRegistry()
	if err != nil {
		return err
	}

	found := false
	newSources := make([]Source, 0, len(registry.Sources))
	for _, s := range registry.Sources {
		if s.ID == id {
			found = true
		} else {
			newSources = append(newSources, s)
		}
	}

	if !found {
		return fmt.Errorf("source not found: %s", id)
	}

	registry.Sources = newSources
	return SaveRegistry(registry)
}

// GetSource returns a source by ID.
func GetSource(id string) (*Source, error) {
	registry, err := LoadRegistry()
	if err != nil {
		return nil, err
	}

	for _, s := range registry.Sources {
		if s.ID == id {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("source not found: %s", id)
}
