package bundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func makeTestZip(t *testing.T, dslContent string) []byte {
	t.Helper()

	manifest := `{
		"workspace": "default",
		"ruleset_key": "test-ruleset",
		"version": 3,
		"checksum": "placeholder",
		"created_at": "2025-01-01T00:00:00Z"
	}`

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	addFile := func(name, content string) {
		t.Helper()
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}

	addFile("manifest.json", manifest)
	addFile("dsl.json", dslContent)

	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	return buf.Bytes()
}

func makeTestZipWithChecksum(t *testing.T) ([]byte, string) {
	t.Helper()

	dslContent := `{"dsl_version":"v1","strategy":"first_match","rules":[]}`
	sum := sha256.Sum256([]byte(dslContent))
	checksum := fmt.Sprintf("sha256:%x", sum)

	manifest := fmt.Sprintf(`{
		"workspace": "default",
		"ruleset_key": "test-ruleset",
		"version": 3,
		"checksum": %q,
		"created_at": "2025-01-01T00:00:00Z"
	}`, checksum)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	addFile := func(name, content string) {
		t.Helper()
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}

	addFile("manifest.json", manifest)
	addFile("dsl.json", dslContent)

	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	return buf.Bytes(), checksum
}

func TestExtract_ReturnsManifest(t *testing.T) {
	dslContent := `{"dsl_version":"v1"}`
	zipBytes := makeTestZip(t, dslContent)

	destDir := t.TempDir()
	manifest, err := Extract(zipBytes, destDir)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if manifest.RulesetKey != "test-ruleset" {
		t.Errorf("ruleset_key: got %q, want %q", manifest.RulesetKey, "test-ruleset")
	}
	if manifest.Version != 3 {
		t.Errorf("version: got %d, want 3", manifest.Version)
	}
	if manifest.Workspace != "default" {
		t.Errorf("workspace: got %q, want %q", manifest.Workspace, "default")
	}

	dslPath := filepath.Join(destDir, "dsl.json")
	if _, err := os.Stat(dslPath); err != nil {
		t.Errorf("dsl.json not extracted: %v", err)
	}
}

func TestExtract_CreatedAt(t *testing.T) {
	zipBytes, _ := makeTestZipWithChecksum(t)
	destDir := t.TempDir()

	manifest, err := Extract(zipBytes, destDir)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !manifest.CreatedAt.Equal(expected) {
		t.Errorf("created_at: got %v, want %v", manifest.CreatedAt, expected)
	}
}

func TestVerifyChecksum_Correct(t *testing.T) {
	dslContent := `{"dsl_version":"v1","strategy":"first_match","rules":[]}`
	sum := sha256.Sum256([]byte(dslContent))
	expectedChecksum := fmt.Sprintf("sha256:%x", sum)

	path := filepath.Join(t.TempDir(), "dsl.json")
	if err := os.WriteFile(path, []byte(dslContent), 0o644); err != nil {
		t.Fatalf("write dsl: %v", err)
	}

	if err := VerifyChecksum(path, expectedChecksum); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	dslContent := `{"dsl_version":"v1"}`

	path := filepath.Join(t.TempDir(), "dsl.json")
	if err := os.WriteFile(path, []byte(dslContent), 0o644); err != nil {
		t.Fatalf("write dsl: %v", err)
	}

	wrongChecksum := "sha256:deadbeefdeadbeef"
	err := VerifyChecksum(path, wrongChecksum)
	if err == nil {
		t.Fatal("expected error on checksum mismatch, got nil")
	}

	var csErr *ChecksumMismatchError
	if _, ok := err.(*ChecksumMismatchError); !ok {
		t.Errorf("expected ChecksumMismatchError, got %T: %v", err, err)
	} else {
		_ = csErr
	}
}

func TestChecksumMismatchError_Message(t *testing.T) {
	err := &ChecksumMismatchError{
		Key:      "payout-routing",
		Expected: "sha256:aaa",
		Got:      "sha256:bbb",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestExtract_MissingManifest(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("dsl.json")
	f.Write([]byte(`{}`))
	w.Close()

	_, err := Extract(buf.Bytes(), t.TempDir())
	if err == nil {
		t.Fatal("expected error when manifest.json is missing")
	}
}
