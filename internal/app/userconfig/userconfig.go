package userconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UserConfig holds the user-level connection settings saved in ~/.rulekit/config.
type UserConfig struct {
	RegistryURL string
	APIKey      string
	Workspace   string
}

// Path returns the path to the user config file.
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".rulekit", "config")
}

// Load reads the config file. Returns an empty UserConfig if the file doesn't exist.
func Load() (UserConfig, error) {
	f, err := os.Open(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return UserConfig{}, nil
		}
		return UserConfig{}, fmt.Errorf("open user config: %w", err)
	}
	defer f.Close()

	var cfg UserConfig
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "registry_url":
			cfg.RegistryURL = strings.TrimSpace(v)
		case "api_key":
			cfg.APIKey = strings.TrimSpace(v)
		case "workspace":
			cfg.Workspace = strings.TrimSpace(v)
		}
	}
	return cfg, scanner.Err()
}

// Save writes the config to ~/.rulekit/config atomically.
func Save(cfg UserConfig) error {
	dir := filepath.Dir(Path())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	var b strings.Builder
	b.WriteString("# RuleKit user config — managed by 'rulekit config'\n\n")
	b.WriteString(fmt.Sprintf("registry_url=%s\n", cfg.RegistryURL))
	b.WriteString(fmt.Sprintf("api_key=%s\n", cfg.APIKey))
	b.WriteString(fmt.Sprintf("workspace=%s\n", cfg.Workspace))

	tmp := Path() + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return os.Rename(tmp, Path())
}
