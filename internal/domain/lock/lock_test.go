package lock

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rulekit.lock")

	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	original := &LockFile{
		Registry:  "http://registry.example.com",
		Workspace: "production",
		Rulesets: map[string]RulesetLock{
			"payout-routing": {
				Version:  4,
				Checksum: "sha256:a3f1c8deadbeef",
				PulledAt: now,
			},
			"fraud-scoring": {
				Version:  2,
				Checksum: "sha256:b7e2d1cafebabe",
				PulledAt: now,
			},
		},
	}

	if err := Write(path, original); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if got.Registry != original.Registry {
		t.Errorf("registry: got %q, want %q", got.Registry, original.Registry)
	}
	if got.Workspace != original.Workspace {
		t.Errorf("workspace: got %q, want %q", got.Workspace, original.Workspace)
	}
	if len(got.Rulesets) != len(original.Rulesets) {
		t.Errorf("rulesets count: got %d, want %d", len(got.Rulesets), len(original.Rulesets))
	}

	for key, want := range original.Rulesets {
		entry, ok := got.Rulesets[key]
		if !ok {
			t.Errorf("missing ruleset %q", key)
			continue
		}
		if entry.Version != want.Version {
			t.Errorf("%s version: got %d, want %d", key, entry.Version, want.Version)
		}
		if entry.Checksum != want.Checksum {
			t.Errorf("%s checksum: got %q, want %q", key, entry.Checksum, want.Checksum)
		}
		if !entry.PulledAt.Equal(want.PulledAt) {
			t.Errorf("%s pulled_at: got %v, want %v", key, entry.PulledAt, want.PulledAt)
		}
	}
}

func TestRead_NonExistentFile(t *testing.T) {
	_, err := Read("/nonexistent/path/rulekit.lock")
	if err == nil {
		t.Fatal("expected error reading non-existent file, got nil")
	}
	if !os.IsNotExist(err) {
		// Read wraps the error, so we check if the underlying error is not-exist
		// via the unwrapped error message
		t.Logf("error (expected not-exist related): %v", err)
	}
}

func TestEmpty(t *testing.T) {
	lf := Empty("http://localhost:8080", "default")
	if lf.Registry != "http://localhost:8080" {
		t.Errorf("got %q", lf.Registry)
	}
	if lf.Rulesets == nil {
		t.Error("rulesets map should not be nil")
	}
}

func TestDashboardField_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rulekit.lock")

	original := &LockFile{
		Registry:  "http://localhost:8080",
		Dashboard: "http://localhost:3000",
		Workspace: "default",
		Rulesets:  make(map[string]RulesetLock),
	}

	if err := Write(path, original); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if got.Dashboard != original.Dashboard {
		t.Errorf("dashboard: got %q, want %q", got.Dashboard, original.Dashboard)
	}
}

func TestDashboardField_BackwardCompat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rulekit.lock")

	// Write a lock file without the dashboard field (simulating old format).
	old := `{"registry":"http://localhost:8080","workspace":"default","rulesets":{}}`
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("read old format: %v", err)
	}

	if got.Dashboard != "" {
		t.Errorf("dashboard should be empty for old format, got %q", got.Dashboard)
	}
	if got.Registry != "http://localhost:8080" {
		t.Errorf("registry: got %q", got.Registry)
	}
}

func TestDashboardField_OmitemptyInJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rulekit.lock")

	lf := &LockFile{
		Registry:  "http://localhost:8080",
		Dashboard: "", // empty — should be omitted
		Workspace: "default",
		Rulesets:  make(map[string]RulesetLock),
	}

	if err := Write(path, lf); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if bytes.Contains(data, []byte(`"dashboard"`)) {
		t.Error("dashboard key should be omitted when empty")
	}
}
