package docker

import (
	"os"
	"os/exec"
	"testing"
)

func TestCheckDocker_NoBinary(t *testing.T) {
	// Override PATH to an empty dir so `docker` cannot be found.
	empty := t.TempDir()
	t.Setenv("PATH", empty)

	err := CheckDocker()
	if err == nil {
		t.Fatal("expected error when docker is not in PATH, got nil")
	}
}

func TestCheckDocker_BinaryExists(t *testing.T) {
	// Skip if docker is not installed on this machine.
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed, skipping binary-exists test")
	}
	// We cannot guarantee the daemon is running in CI, so we only verify
	// that CheckDocker returns an error OR nil (not panic).
	_ = CheckDocker()
}

func TestComposeDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := ComposeDir()
	if dir == "" {
		t.Error("ComposeDir returned empty string")
	}
}

func TestComposePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := ComposePath()
	if path == "" {
		t.Error("ComposePath returned empty string")
	}
}

func TestDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := DataDir()
	if dir == "" {
		t.Error("DataDir returned empty string")
	}
}

func TestSQLiteDBPath_ContainsExpectedSuffix(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := SQLiteDBPath()
	expected := "rulekit.db"
	if len(path) < len(expected) || path[len(path)-len(expected):] != expected {
		t.Errorf("SQLiteDBPath %q does not end with %q", path, expected)
	}
}

func TestGenerateCompose_CreatesDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	opts := DefaultComposeOptions()
	if err := GenerateCompose(opts); err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}

	if _, err := os.Stat(ComposeDir()); err != nil {
		t.Errorf("compose dir not created: %v", err)
	}
}
