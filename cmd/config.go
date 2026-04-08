package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rulekit-dev/rulekit-cli/internal/app/userconfig"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"golang.org/x/term"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set your registry URL, API key, and workspace",
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(_ *cobra.Command, _ []string) error {
	current, err := userconfig.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("rulekit config requires an interactive terminal")
	}

	orange := lipgloss.Color("#FF7800")
	theme := huh.ThemeBase()
	theme.Focused.Title = theme.Focused.Title.Foreground(orange).Bold(true)
	theme.Focused.Base = theme.Focused.Base.BorderForeground(orange)
	theme.Focused.Description = theme.Focused.Description.Faint(true)

	registryURL := current.RegistryURL
	if registryURL == "" {
		registryURL = "http://localhost:8080"
	}
	apiKey := current.APIKey
	workspace := current.Workspace
	if workspace == "" {
		workspace = "default"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Registry URL").
				Description("Base URL of your RuleKit registry (e.g. https://registry.example.com).").
				Value(&registryURL).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("registry URL is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("API key").
				Description("Your rk_* API key for registry authentication.").
				EchoMode(huh.EchoModePassword).
				Value(&apiKey),
			huh.NewInput().
				Title("Workspace").
				Description("Default workspace to use when none is specified.").
				Value(&workspace),
		),
	).WithTheme(theme)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	if err := userconfig.Save(userconfig.UserConfig{
		RegistryURL: registryURL,
		APIKey:      apiKey,
		Workspace:   workspace,
	}); err != nil {
		output.Error("save config: %v", err)
		return err
	}

	output.Success("Config saved to " + userconfig.Path())
	return nil
}
