package config

import "os"

type Config struct {
	RegistryURL string
	Namespace   string
	Dir         string
	Token       string
	Verbose     bool
}

// Resolve builds a Config by layering sources in priority order:
// defaults → lockfile values → environment variables → CLI flags.
func Resolve(flagRegistry, flagNamespace, flagDir, flagToken string, lockRegistry, lockNamespace string) Config {
	cfg := Config{
		RegistryURL: "http://localhost:8080",
		Namespace:   "default",
		Dir:         ".rulekit",
	}

	if lockRegistry != "" {
		cfg.RegistryURL = lockRegistry
	}
	if lockNamespace != "" {
		cfg.Namespace = lockNamespace
	}

	if v := os.Getenv("RULEKIT_REGISTRY_URL"); v != "" {
		cfg.RegistryURL = v
	}
	if v := os.Getenv("RULEKIT_NAMESPACE"); v != "" {
		cfg.Namespace = v
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
	if flagNamespace != "" {
		cfg.Namespace = flagNamespace
	}
	if flagDir != "" {
		cfg.Dir = flagDir
	}
	if flagToken != "" {
		cfg.Token = flagToken
	}

	return cfg
}
