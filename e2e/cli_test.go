package e2e

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type commandResult struct {
	stdout string
	stderr string
	err    error
}

type provenance struct {
	SourceID      string `json:"source_id"`
	SourceName    string `json:"source_name"`
	SourceType    string `json:"source_type"`
	BundleVersion string `json:"bundle_version"`
	PresetName    string `json:"preset_name"`
	Entrypoint    string `json:"entrypoint"`
	AppliedAt     string `json:"applied_at"`
}

func TestVersion(t *testing.T) {
	result := runOC(t, testEnv(t), "version")
	requireSuccess(t, result)
	if !strings.HasPrefix(result.stdout, "oc ") {
		t.Fatalf("expected version output, got stdout=%q stderr=%q", result.stdout, result.stderr)
	}
}

func TestLocalDirectoryFlow(t *testing.T) {
	env := testEnv(t)
	bundleDir := copyFixtureBundle(t)
	projectRoot := t.TempDir()

	addResult := runOC(t, env, "source", "add", bundleDir, "--name", "fixture-dir")
	requireSuccess(t, addResult)
	sourceID := extractSourceID(t, addResult.stdout)

	listResult := runOC(t, env, "source", "list")
	requireSuccess(t, listResult)
	requireContains(t, listResult.stdout, sourceID)
	requireContains(t, listResult.stdout, "fixture-dir")
	requireContains(t, listResult.stdout, "local-directory")

	applyResult := runOC(t, env, "bundle", "apply", sourceID, "--preset", "fixture", "--project-root", projectRoot)
	requireSuccess(t, applyResult)

	configPath := filepath.Join(projectRoot, "opencode.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read applied config: %v", err)
	}
	requireContains(t, string(configData), `"mode": "fixture"`)

	provenancePath := filepath.Join(projectRoot, ".opencode", "bundle-provenance.json")
	prov := readProvenance(t, provenancePath)
	if prov.SourceID != sourceID {
		t.Fatalf("expected source id %q, got %q", sourceID, prov.SourceID)
	}
	if prov.SourceType != "local-directory" {
		t.Fatalf("expected source type local-directory, got %q", prov.SourceType)
	}
	if prov.PresetName != "fixture" {
		t.Fatalf("expected preset fixture, got %q", prov.PresetName)
	}

	statusResult := runOC(t, env, "bundle", "status", "--project-root", projectRoot)
	requireSuccess(t, statusResult)
	requireContains(t, statusResult.stdout, "Bundle Provenance:")
	requireContains(t, statusResult.stdout, sourceID)
	requireContains(t, statusResult.stdout, "fixture-dir")
	requireContains(t, statusResult.stdout, "fixture")

	overwriteResult := runOC(t, env, "bundle", "apply", sourceID, "--preset", "fixture", "--project-root", projectRoot)
	requireFailure(t, overwriteResult)
	requireContains(t, overwriteResult.stderr, "output file exists")

	forceResult := runOC(t, env, "bundle", "apply", sourceID, "--preset", "fixture", "--project-root", projectRoot, "--force")
	requireSuccess(t, forceResult)
	updatedProv := readProvenance(t, provenancePath)
	if updatedProv.SourceID != sourceID {
		t.Fatalf("expected forced apply to preserve source id %q, got %q", sourceID, updatedProv.SourceID)
	}
	if updatedProv.SourceName != "fixture-dir" {
		t.Fatalf("expected forced apply source name fixture-dir, got %q", updatedProv.SourceName)
	}
}

func TestLocalArchiveFlow(t *testing.T) {
	env := testEnv(t)
	bundleDir := copyFixtureBundle(t)
	archivePath := filepath.Join(t.TempDir(), "fixture-bundle.tar.gz")
	createTarGzFromDir(t, bundleDir, archivePath)
	projectRoot := t.TempDir()

	addResult := runOC(t, env, "source", "add", archivePath, "--name", "fixture-archive")
	requireSuccess(t, addResult)
	sourceID := extractSourceID(t, addResult.stdout)

	applyResult := runOC(t, env, "bundle", "apply", sourceID, "--preset", "fixture", "--project-root", projectRoot)
	requireSuccess(t, applyResult)

	provenancePath := filepath.Join(projectRoot, ".opencode", "bundle-provenance.json")
	prov := readProvenance(t, provenancePath)
	if prov.SourceType != "local-archive" {
		t.Fatalf("expected source type local-archive, got %q", prov.SourceType)
	}
	if prov.SourceName != "fixture-archive" {
		t.Fatalf("expected source name fixture-archive, got %q", prov.SourceName)
	}
}

func TestBundleApplyFailsForUnknownSource(t *testing.T) {
	projectRoot := t.TempDir()
	result := runOC(t, testEnv(t), "bundle", "apply", "missing-id", "--preset", "fixture", "--project-root", projectRoot)
	requireFailure(t, result)
	requireContains(t, result.stderr, "source not found")
}

func TestSourceAddFailsWithoutManifest(t *testing.T) {
	bundleDir := t.TempDir()
	result := runOC(t, testEnv(t), "source", "add", bundleDir)
	requireFailure(t, result)
	requireContains(t, result.stderr, "bundle manifest not found")
}

func TestInvalidTarballFailsOnApply(t *testing.T) {
	env := testEnv(t)
	archivePath := filepath.Join(t.TempDir(), "invalid.tar.gz")
	if err := os.WriteFile(archivePath, []byte("not a tarball"), 0o644); err != nil {
		t.Fatalf("failed to write invalid archive: %v", err)
	}

	addResult := runOC(t, env, "source", "add", archivePath, "--name", "broken-archive")
	requireSuccess(t, addResult)
	sourceID := extractSourceID(t, addResult.stdout)

	projectRoot := t.TempDir()
	applyResult := runOC(t, env, "bundle", "apply", sourceID, "--preset", "fixture", "--project-root", projectRoot)
	requireFailure(t, applyResult)
	requireContains(t, applyResult.stderr, "failed to resolve source")
	if runtime.GOOS == "darwin" {
		requireContains(t, applyResult.stderr, "failed to extract tarball")
		return
	}
	requireContains(t, applyResult.stderr, "failed to extract tarball")
}

func testEnv(t *testing.T) []string {
	t.Helper()
	homeDir := t.TempDir()
	configHome := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configHome, 0o755); err != nil {
		t.Fatalf("failed to create config home: %v", err)
	}

	pathValue := os.Getenv("PATH")
	if pathValue == "" {
		t.Fatal("PATH is required for subprocess execution")
	}

	return []string{
		"HOME=" + homeDir,
		"XDG_CONFIG_HOME=" + configHome,
		"PATH=" + pathValue,
	}
}

func runOC(t *testing.T, env []string, args ...string) commandResult {
	t.Helper()
	binaryPath := os.Getenv("OC_E2E_BINARY")
	if binaryPath == "" {
		t.Skip("OC_E2E_BINARY not set; skipping black-box CLI E2E tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, args...)
	cmd.Env = env
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return commandResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		err:    err,
	}
}

func requireSuccess(t *testing.T, result commandResult) {
	t.Helper()
	if result.err != nil {
		t.Fatalf("expected success, got err=%v stdout=%q stderr=%q", result.err, result.stdout, result.stderr)
	}
}

func requireFailure(t *testing.T, result commandResult) {
	t.Helper()
	if result.err == nil {
		t.Fatalf("expected failure, got stdout=%q stderr=%q", result.stdout, result.stderr)
	}
}

func requireContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q to contain %q", haystack, needle)
	}
}

func extractSourceID(t *testing.T, stdout string) string {
	t.Helper()
	for _, line := range strings.Split(stdout, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ID:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "ID:"))
		}
	}
	t.Fatalf("failed to extract source id from stdout=%q", stdout)
	return ""
}

func readProvenance(t *testing.T, path string) provenance {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read provenance file: %v", err)
	}

	var prov provenance
	if err := json.Unmarshal(data, &prov); err != nil {
		t.Fatalf("failed to parse provenance file: %v", err)
	}
	return prov
}

func copyFixtureBundle(t *testing.T) string {
	t.Helper()
	sourceRoot := filepath.Join("testdata", "fixture-bundle")
	destRoot := filepath.Join(t.TempDir(), "fixture-bundle")
	copyDir(t, sourceRoot, destRoot)
	return destRoot
}

func copyDir(t *testing.T, srcDir, destDir string) {
	t.Helper()
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("failed to read fixture dir %s: %v", srcDir, err)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("failed to create fixture dir %s: %v", destDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())
		if entry.IsDir() {
			copyDir(t, srcPath, destPath)
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("failed to read fixture file %s: %v", srcPath, err)
		}
		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			t.Fatalf("failed to write fixture file %s: %v", destPath, err)
		}
	}
}

func createTarGzFromDir(t *testing.T, sourceDir, archivePath string) {
	t.Helper()
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create archive %s: %v", archivePath, err)
	}
	defer archiveFile.Close()

	gzipWriter := gzip.NewWriter(archiveFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name: filepath.ToSlash(filepath.Join(filepath.Base(sourceDir), relPath)),
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tarWriter.Write(content); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("failed to build tar.gz fixture: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("failed to finalize tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("failed to finalize gzip writer: %v", err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatalf("failed to close archive file: %v", err)
	}
}
