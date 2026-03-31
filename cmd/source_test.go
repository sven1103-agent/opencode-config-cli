package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sven1103-agent/opencode-config-cli/internal/source"
)

// saveRegistry saves the current registry and returns a function to restore it
func saveRegistry(t *testing.T) func() {
	t.Helper()
	original, err := source.LoadRegistry()
	if err != nil {
		t.Logf("No existing registry to save: %v", err)
		return func() {}
	}

	return func() {
		// Restore the original registry
		if err := source.SaveRegistry(original); err != nil {
			t.Errorf("Failed to restore registry: %v", err)
		}
	}
}

// setupTestBundle creates a temp directory with a valid bundle manifest
func setupTestBundle(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a bundle manifest for testing
	manifestContent := `{
  "manifest_version": 1,
  "bundle_name": "test-bundle",
  "bundle_version": "v1.0.0",
  "presets": [
    {
      "name": "test",
      "description": "Test preset",
      "entrypoint": "test.json",
      "prompt_files": []
    }
  ]
}`
	manifestPath := filepath.Join(tempDir, "opencode-bundle.manifest.json")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create manifest: %v", err)
	}

	return tempDir
}

func TestSourceAddCommand(t *testing.T) {
	// Save and restore registry
	restore := saveRegistry(t)
	defer restore()

	// Clear any existing sources for clean test
	registry, _ := source.LoadRegistry()
	registry.Sources = []source.Source{}
	if err := source.SaveRegistry(registry); err != nil {
		t.Fatalf("failed to save registry: %v", err)
	}

	// Create a temporary directory for testing
	tempDir := setupTestBundle(t)
	defer os.RemoveAll(tempDir)

	// Reset the sourceName flag for testing
	sourceName = ""

	// Test adding a source
	err := runSourceAdd(tempDir)
	if err != nil {
		t.Errorf("runSourceAdd() error = %v", err)
	}

	// Verify source was added
	sources, err := source.ListSources()
	if err != nil {
		t.Errorf("source.ListSources() error = %v", err)
	}
	if len(sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Location != tempDir {
		t.Errorf("expected location %s, got %s", tempDir, sources[0].Location)
	}
}

func TestSourceListCommand(t *testing.T) {
	// Save and restore registry
	restore := saveRegistry(t)
	defer restore()

	// Test listing sources (should not error)
	err := runSourceList()
	if err != nil {
		t.Errorf("runSourceList() error = %v", err)
	}
}

func TestSourceAddInvalidLocation(t *testing.T) {
	// Test adding an invalid location
	err := runSourceAdd("/nonexistent/path/12345")
	if err == nil {
		t.Error("runSourceAdd() expected error for invalid location")
	}
}

func TestSourceRemoveCommand(t *testing.T) {
	// Save and restore registry
	restore := saveRegistry(t)
	defer restore()

	// Clear any existing sources for clean test
	registry, _ := source.LoadRegistry()
	registry.Sources = []source.Source{}
	if err := source.SaveRegistry(registry); err != nil {
		t.Fatalf("failed to save registry: %v", err)
	}

	// First add a source to remove
	tempDir := setupTestBundle(t)
	defer os.RemoveAll(tempDir)

	sourceName = ""
	err := runSourceAdd(tempDir)
	if err != nil {
		t.Fatalf("runSourceAdd() error = %v", err)
	}

	// Get the source ID
	sources, err := source.ListSources()
	if err != nil {
		t.Fatalf("source.ListSources() error = %v", err)
	}

	if len(sources) == 0 {
		t.Fatal("expected at least one source after add")
	}

	sourceID := sources[0].ID

	// Test removing the source
	err = runSourceRemove(sourceID)
	if err != nil {
		t.Errorf("runSourceRemove() error = %v", err)
	}

	// Verify source was removed
	sources, err = source.ListSources()
	if err != nil {
		t.Errorf("source.ListSources() error = %v", err)
	}
	if len(sources) != 0 {
		t.Errorf("expected 0 sources after remove, got %d", len(sources))
	}
}

func TestSourceRemoveNonexistent(t *testing.T) {
	// Test removing a nonexistent source
	err := runSourceRemove("nonexistent-id")
	if err == nil {
		t.Error("runSourceRemove() expected error for nonexistent source")
	}
}

func TestSourceAddWithoutManifest(t *testing.T) {
	// Create a temp directory without manifest
	tempDir, err := os.MkdirTemp("", "opencode-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test adding a source without manifest
	err = runSourceAdd(tempDir)
	if err == nil {
		t.Error("runSourceAdd() expected error for directory without manifest")
	}
}

func TestSourceCommandsFlags(t *testing.T) {
	// Test that flags are properly configured
	if sourceAddCmd.Flags().Lookup("name") == nil {
		t.Error("name flag should exist on source add command")
	}
}
