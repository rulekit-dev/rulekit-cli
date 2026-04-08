package config

import (
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Config struct {
	RegistryURL string
	Workspace   string
	Dir         string
	APIKey      string
	Verbose     bool
}

// Resolve builds a Config by layering sources in priority order:
// defaults → lockfile values → environment variables → CLI flags.
func Resolve(flagRegistry, flagWorkspace, flagDir, flagAPIKey string, lockRegistry, lockWorkspace string) Config {
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
	if v := os.Getenv("RULEKIT_API_KEY"); v != "" {
		cfg.APIKey = v
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
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}

	return cfg
}

// ResolveInteractive calls Resolve, then prompts for any missing required fields
// (registry URL, api key) when running in an interactive terminal.
func ResolveInteractive(flagRegistry, flagWorkspace, flagDir, flagAPIKey string, lockRegistry, lockWorkspace string) (Config, error) {
	cfg := Resolve(flagRegistry, flagWorkspace, flagDir, flagAPIKey, lockRegistry, lockWorkspace)

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return cfg, nil
	}

	orange := lipgloss.Color("#FF7800")
	theme := huh.ThemeBase()
	theme.Focused.Title = theme.Focused.Title.Foreground(orange).Bold(true)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(orange)
	theme.Focused.Base = theme.Focused.Base.BorderForeground(orange)
	theme.Focused.Description = theme.Focused.Description.Faint(true)

	var fields []huh.Field

	if cfg.RegistryURL == "http://localhost:8080" && flagRegistry == "" &&
		os.Getenv("RULEKIT_REGISTRY_URL") == "" && lockRegistry == "" {
		fields = append(fields, huh.NewInput().
			Title("Registry URL").
			Description("Base URL of your RuleKit registry.").
			Value(&cfg.RegistryURL).
			Validate(func(s string) error {
				if s == "" {
					return os.ErrInvalid
				}
				return nil
			}),
		)
	}

	if cfg.APIKey == "" {
		fields = append(fields, huh.NewInput().
			Title("API key").
			Description("Your rk_* API key for registry authentication.").
			EchoMode(huh.EchoModePassword).
			Value(&cfg.APIKey),
		)
	}

	if len(fields) == 0 {
		return cfg, nil
	}

	if err := huh.NewForm(huh.NewGroup(fields...)).WithTheme(theme).Run(); err != nil {
		return cfg, err
	}

	return cfg, nil
}
