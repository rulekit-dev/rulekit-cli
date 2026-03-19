package config

import (
	"testing"
)

func TestResolve_Defaults(t *testing.T) {
	cfg := Resolve("", "", "", "", "", "")

	if cfg.RegistryURL != "http://localhost:8080" {
		t.Errorf("expected default registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Namespace != "default" {
		t.Errorf("expected default namespace, got %q", cfg.Namespace)
	}
	if cfg.Dir != ".rulekit" {
		t.Errorf("expected default dir, got %q", cfg.Dir)
	}
	if cfg.Token != "" {
		t.Errorf("expected empty token, got %q", cfg.Token)
	}
}

func TestResolve_LockfileOverridesDefaults(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "")
	t.Setenv("RULEKIT_NAMESPACE", "")
	t.Setenv("RULEKIT_DIR", "")
	t.Setenv("RULEKIT_TOKEN", "")
	cfg := Resolve("", "", "", "", "http://registry.example.com", "production")

	if cfg.RegistryURL != "http://registry.example.com" {
		t.Errorf("expected lockfile registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Namespace != "production" {
		t.Errorf("expected lockfile namespace, got %q", cfg.Namespace)
	}
}

func TestResolve_EnvOverridesLockfile(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "http://env.example.com")
	t.Setenv("RULEKIT_NAMESPACE", "staging")
	t.Setenv("RULEKIT_DIR", "/tmp/rules")
	t.Setenv("RULEKIT_TOKEN", "env-token")

	cfg := Resolve("", "", "", "", "http://lock.example.com", "production")

	if cfg.RegistryURL != "http://env.example.com" {
		t.Errorf("expected env registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Namespace != "staging" {
		t.Errorf("expected env namespace, got %q", cfg.Namespace)
	}
	if cfg.Dir != "/tmp/rules" {
		t.Errorf("expected env dir, got %q", cfg.Dir)
	}
	if cfg.Token != "env-token" {
		t.Errorf("expected env token, got %q", cfg.Token)
	}
}

func TestResolve_FlagsOverrideEnv(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "http://env.example.com")
	t.Setenv("RULEKIT_NAMESPACE", "staging")
	t.Setenv("RULEKIT_TOKEN", "env-token")

	cfg := Resolve("http://flag.example.com", "flag-ns", "flag-dir", "flag-token", "", "")

	if cfg.RegistryURL != "http://flag.example.com" {
		t.Errorf("expected flag registry URL, got %q", cfg.RegistryURL)
	}
	if cfg.Namespace != "flag-ns" {
		t.Errorf("expected flag namespace, got %q", cfg.Namespace)
	}
	if cfg.Dir != "flag-dir" {
		t.Errorf("expected flag dir, got %q", cfg.Dir)
	}
	if cfg.Token != "flag-token" {
		t.Errorf("expected flag token, got %q", cfg.Token)
	}
}

func TestResolve_EmptyFlagsDoNotOverride(t *testing.T) {
	t.Setenv("RULEKIT_REGISTRY_URL", "http://env.example.com")

	cfg := Resolve("", "", "", "", "", "")

	if cfg.RegistryURL != "http://env.example.com" {
		t.Errorf("expected env registry URL, got %q", cfg.RegistryURL)
	}
}
