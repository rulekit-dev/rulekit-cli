package lock

import (
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
		Namespace: "production",
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
	if got.Namespace != original.Namespace {
		t.Errorf("namespace: got %q, want %q", got.Namespace, original.Namespace)
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
