package stack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rulekit-dev/rulekit-cli/internal/app/wizard"
)

// TestResolveConfig_YesFlag verifies that --yes skips the wizard and returns defaults.
func TestResolveConfig_YesFlag(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	// --yes flag set, no existing .env
	upYes = true
	upReconfigure = false
	t.Cleanup(func() {
		upYes = false
		upReconfigure = false
	})

	cfg, err := resolveConfig(envPath)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}

	// Should get defaults without wizard.
	defaults := wizard.Defaults()
	if cfg.Store != defaults.Store {
		t.Errorf("store: got %q, want %q", cfg.Store, defaults.Store)
	}
	if cfg.RegistryPort != defaults.RegistryPort {
		t.Errorf("registryPort: got %d, want %d", cfg.RegistryPort, defaults.RegistryPort)
	}
}

// TestResolveConfig_ExistingEnvSkipsWizard verifies that an existing .env skips the wizard.
func TestResolveConfig_ExistingEnvSkipsWizard(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	// Write a minimal .env file.
	content := "RULEKIT_STORE=sqlite\nRULEKIT_DATA_DIR=/data\nRULEKIT_ADDR=:9999\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	upYes = false
	upReconfigure = false
	t.Cleanup(func() {
		upYes = false
		upReconfigure = false
	})

	cfg, err := resolveConfig(envPath)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}

	// Should load from existing .env, not run wizard.
	if cfg.RegistryPort != 9999 {
		t.Errorf("expected port 9999 from existing .env, got %d", cfg.RegistryPort)
	}
}

// TestResolveConfig_ReconfigureFlag verifies --reconfigure triggers the wizard
// even when .env exists. In non-TTY context (CI), it will use defaults from the existing file.
func TestResolveConfig_ReconfigureFlag(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")

	// Write an existing .env.
	content := "RULEKIT_STORE=sqlite\nRULEKIT_DATA_DIR=/data\nRULEKIT_ADDR=:9999\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}

	upYes = false
	upReconfigure = true
	t.Cleanup(func() {
		upYes = false
		upReconfigure = false
	})

	// In non-TTY (test environment), RunWizard returns defaults from the existing config.
	cfg, err := resolveConfig(envPath)
	if err != nil {
		t.Fatalf("resolveConfig with --reconfigure: %v", err)
	}

	// Should have loaded from the existing env as pre-fill.
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}
