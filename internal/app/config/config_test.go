package config

import (
	"testing"
)

// isolateHome points HOME at an empty temp dir so userconfig.Load finds no file.
func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestResolve_Defaults(t *testing.T) {
	isolateHome(t)
	t.Setenv("RULEKIT_REGISTRY_URL", "")
	t.Setenv("RULEKIT_WORKSPACE", "")
	t.Setenv("RULEKIT_DIR", "")
	t.Setenv("RULEKIT_API_KEY", "")
	cfg := Resolve("", "", "", "", "", "")

	if cfg.RegistryURL != "http://localhost:8080" {
		t.Errorf("expected default registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Workspace != "default" {
		t.Errorf("expected default workspace, got %q", cfg.Workspace)
	}
	if cfg.Dir != ".rulekit" {
		t.Errorf("expected default dir, got %q", cfg.Dir)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty api key, got %q", cfg.APIKey)
	}
}

func TestResolve_LockfileOverridesDefaults(t *testing.T) {
	isolateHome(t)
	t.Setenv("RULEKIT_REGISTRY_URL", "")
	t.Setenv("RULEKIT_WORKSPACE", "")
	t.Setenv("RULEKIT_DIR", "")
	t.Setenv("RULEKIT_API_KEY", "")
	cfg := Resolve("", "", "", "", "http://registry.example.com", "production")

	if cfg.RegistryURL != "http://registry.example.com" {
		t.Errorf("expected lockfile registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Workspace != "production" {
		t.Errorf("expected lockfile workspace, got %q", cfg.Workspace)
	}
}

func TestResolve_EnvOverridesLockfile(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "http://env.example.com")
	t.Setenv("RULEKIT_WORKSPACE", "staging")
	t.Setenv("RULEKIT_DIR", "/tmp/rules")
	t.Setenv("RULEKIT_API_KEY", "env-apikey")

	cfg := Resolve("", "", "", "", "http://lock.example.com", "production")

	if cfg.RegistryURL != "http://env.example.com" {
		t.Errorf("expected env registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Workspace != "staging" {
		t.Errorf("expected env workspace, got %q", cfg.Workspace)
	}
	if cfg.Dir != "/tmp/rules" {
		t.Errorf("expected env dir, got %q", cfg.Dir)
	}
	if cfg.APIKey != "env-apikey" {
		t.Errorf("expected env api key, got %q", cfg.APIKey)
	}
}

func TestResolve_FlagsOverrideEnv(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "http://env.example.com")
	t.Setenv("RULEKIT_WORKSPACE", "staging")
	t.Setenv("RULEKIT_API_KEY", "env-apikey")

	cfg := Resolve("http://flag.example.com", "flag-ws", "flag-dir", "flag-apikey", "", "")

	if cfg.RegistryURL != "http://flag.example.com" {
		t.Errorf("expected flag registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Workspace != "flag-ws" {
		t.Errorf("expected flag workspace, got %q", cfg.Workspace)
	}
	if cfg.Dir != "flag-dir" {
		t.Errorf("expected flag dir, got %q", cfg.Dir)
	}
	if cfg.APIKey != "flag-apikey" {
		t.Errorf("expected flag api key, got %q", cfg.APIKey)
	}
}

func TestResolve_EmptyFlagsDoNotOverride(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "http://env.example.com")

	cfg := Resolve("", "", "", "", "", "")

	if cfg.RegistryURL != "http://env.example.com" {
		t.Errorf("expected env registry URL, got %q", cfg.RegistryURL)
	}
}
