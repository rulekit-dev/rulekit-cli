package config

import "os"

type Config struct {
	RegistryURL string
	Workspace   string
	Dir         string
	Token       string
	Verbose     bool
}

// Resolve builds a Config by layering sources in priority order:
// defaults → lockfile values → environment variables → CLI flags.
func Resolve(flagRegistry, flagWorkspace, flagDir, flagToken string, lockRegistry, lockWorkspace string) Config {
	cfg := Config{
		RegistryURL: "http://localhost:8080",
		Workspace:   "default",
		Dir:         ".rulekit",
	}

	if lockRegistry != "" {
		cfg.RegistryURL = lockRegistry
	}
	if lockWorkspace != "" {
		cfg.Workspace = lockWorkspace
	}

	if v := os.Getenv("RULEKIT_REGISTRY_URL"); v != "" {
		cfg.RegistryURL = v
	}
	if v := os.Getenv("RULEKIT_WORKSPACE"); v != "" {
		cfg.Workspace = v
	}
	if v := os.Getenv("RULEKIT_DIR"); v != "" {
		cfg.Dir = v
	}
	if v := os.Getenv("RULEKIT_TOKEN"); v != "" {
		cfg.Token = v
	}

	if flagRegistry != "" {
		cfg.RegistryURL = flagRegistry
	}
	if flagWorkspace != "" {
		cfg.Workspace = flagWorkspace
	}
	if flagDir != "" {
		cfg.Dir = flagDir
	}
	if flagToken != "" {
		cfg.Token = flagToken
	}

	return cfg
}
