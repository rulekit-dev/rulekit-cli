package ruleset

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rulekit-dev/rulekit-cli/internal/globals"
)

func makeBundle(t *testing.T, key string, version int, tamper bool) []byte {
	t.Helper()

	dslContent := `{"dsl_version":"v1","strategy":"first_match","rules":[]}`
	sum := sha256.Sum256([]byte(dslContent))
	checksum := fmt.Sprintf("sha256:%x", sum)
	if tamper {
		checksum = "sha256:deadbeefdeadbeef000000000000000000000000000000000000000000000000"
	}

	manifest := map[string]any{
		"workspace":   "default",
		"ruleset_key": key,
		"version":     version,
		"checksum":    checksum,
		"created_at":  "2025-01-01T00:00:00Z",
	}
	manifestBytes, _ := json.Marshal(manifest)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	addEntry := func(name string, data []byte) {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		f.Write(data)
	}
	addEntry("manifest.json", manifestBytes)
	addEntry("dsl.json", []byte(dslContent))
	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func setupTempDir(t *testing.T) string {
	t.Helper()
	resetFlags()
	dir := t.TempDir()
	globals.LockfilePath = filepath.Join(dir, "rulekit.lock")
	t.Setenv("RULEKIT_DIR", filepath.Join(dir, ".rulekit"))
	t.Setenv("RULEKIT_REGISTRY_URL", "")
	t.Setenv("RULEKIT_WORKSPACE", "")
	t.Setenv("RULEKIT_TOKEN", "")
	t.Cleanup(func() { globals.LockfilePath = "rulekit.lock" })
	return dir
}

func resetFlags() {
	globals.Registry = ""
	globals.Workspace = ""
	globals.Dir = ""
	globals.Token = ""
	globals.Verbose = false
	pullKey = ""
	pullVersion = ""
	addVersion = "latest"
}

func runAddCmd(key, registry, version string) error {
	resetFlags()
	globals.Registry = registry
	addVersion = version
	return runAdd(addCmd, []string{key})
}

func runPullCmd(key, version string) error {
	resetFlags()
	pullKey = key
	pullVersion = version
	return runPull(pullCmd, nil)
}

func runVerifyCmd() error {
	resetFlags()
	return runVerify(verifyCmd, nil)
}

// --- pull tests ---

func TestPull_HappyPath(t *testing.T) {
	dir := setupTempDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(makeBundle(t, "pricing", 2, false))
	}))
	defer srv.Close()

	if err := runAddCmd("pricing", srv.URL, "latest"); err != nil {
		t.Fatalf("add: %v", err)
	}

	lf := readLockfile(t)
	entry, ok := lf["rulesets"].(map[string]any)["pricing"]
	if !ok {
		t.Fatal("pricing not in lockfile")
	}
	if v := int(entry.(map[string]any)["version"].(float64)); v != 2 {
		t.Errorf("version: got %d, want 2", v)
	}

	if _, err := os.Stat(filepath.Join(dir, ".rulekit", "pricing", "dsl.json")); err != nil {
		t.Errorf("dsl.json not found: %v", err)
	}
}

func TestPull_ChecksumMismatch(t *testing.T) {
	setupTempDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(makeBundle(t, "pricing", 2, true))
	}))
	defer srv.Close()

	err := runAddCmd("pricing", srv.URL, "latest")
	if err == nil {
		t.Fatal("expected checksum mismatch error, got nil")
	}
}

func TestPull_RegistryUnreachable(t *testing.T) {
	setupTempDir(t)

	err := runAddCmd("pricing", "http://127.0.0.1:1", "latest")
	if err == nil {
		t.Fatal("expected error when registry unreachable, got nil")
	}
}

func TestPull_AllFromLockfile(t *testing.T) {
	setupTempDir(t)

	calls := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var key string
		fmt.Sscanf(r.URL.Path, "/v1/rulesets/%s", &key)
		calls[r.URL.Path]++
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(makeBundle(t, "pricing", 1, false))
	}))
	defer srv.Close()

	writeLockfile(t, srv.URL, map[string]any{
		"pricing":       map[string]any{"version": 1, "checksum": "sha256:x", "pulled_at": "2025-01-01T00:00:00Z"},
		"fraud-scoring": map[string]any{"version": 1, "checksum": "sha256:y", "pulled_at": "2025-01-01T00:00:00Z"},
	})

	resetFlags()
	if err := runPull(pullCmd, nil); err != nil {
		t.Fatalf("pull all: %v", err)
	}

	if len(calls) != 2 {
		t.Errorf("expected 2 bundle requests, got %d", len(calls))
	}
}

// --- verify tests ---

func TestVerify_AllMatch(t *testing.T) {
	setupTempDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(makeBundle(t, "pricing", 1, false))
	}))
	defer srv.Close()

	if err := runAddCmd("pricing", srv.URL, "latest"); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := runVerifyCmd(); err != nil {
		t.Errorf("verify: expected no error, got %v", err)
	}
}

func TestVerify_OneMismatch(t *testing.T) {
	dir := setupTempDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(makeBundle(t, "pricing", 1, false))
	}))
	defer srv.Close()

	if err := runAddCmd("pricing", srv.URL, "latest"); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".rulekit", "pricing", "dsl.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}

	err := runVerifyCmd()
	if err == nil {
		t.Fatal("expected verify to fail after tampering, got nil")
	}
}

func TestVerify_MissingFile(t *testing.T) {
	dir := setupTempDir(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(makeBundle(t, "pricing", 1, false))
	}))
	defer srv.Close()

	if err := runAddCmd("pricing", srv.URL, "latest"); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := os.Remove(filepath.Join(dir, ".rulekit", "pricing", "dsl.json")); err != nil {
		t.Fatalf("remove: %v", err)
	}

	err := runVerifyCmd()
	if err == nil {
		t.Fatal("expected verify to fail for missing file, got nil")
	}
}

// --- helpers ---

func readLockfile(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(globals.LockfilePath)
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("parse lockfile: %v", err)
	}
	return out
}

func writeLockfile(t *testing.T, registry string, rulesets map[string]any) {
	t.Helper()
	lf := map[string]any{
		"registry":  registry,
		"workspace": "default",
		"rulesets":  rulesets,
	}
	data, _ := json.MarshalIndent(lf, "", "  ")
	if err := os.WriteFile(globals.LockfilePath, data, 0o644); err != nil {
		t.Fatalf("write lockfile: %v", err)
	}
}
