package ruleset

import (
	"context"
	"errors"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
	"github.com/rulekit-dev/rulekit-cli/internal/app/config"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/bundle"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/registry"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var addVersion string

var addCmd = &cobra.Command{
	Use:     "add [ruleset-key]",
	Short:   "Add a ruleset to the lockfile and pull it",
	GroupID: "ruleset",
	RunE:    runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addVersion, "version", "latest", "Version to pull (number or \"latest\")")
}

func runAdd(cmd *cobra.Command, args []string) error {
	lf, err := loadOrEmptyLock("")
	if err != nil {
		output.Error("%v", err)
		return globals.Exitf(1, "%v", err)
	}

	cfg, err := config.ResolveInteractive(globals.Registry, globals.Workspace, globals.Dir, globals.APIKey, lf.Registry, lf.Workspace)
	if err != nil {
		return err
	}

	rulesetKey := ""
	if len(args) > 0 {
		rulesetKey = args[0]
	}

	if err := promptAddInputs(&rulesetKey, &cfg.Workspace, &cfg.APIKey); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	lf.Registry = cfg.RegistryURL
	lf.Workspace = cfg.Workspace

	client := registry.NewClient(cfg.RegistryURL, cfg.APIKey)

	ver := addVersion
	if ver == "" {
		ver = "latest"
	}

	if err := pullOne(context.Background(), client, lf, cfg.Dir, rulesetKey, ver, cfg.Workspace); err != nil {
		output.Error("%v", err)
		var csErr *bundle.ChecksumMismatchError
		if errors.As(err, &csErr) {
			return globals.Exitf(2, "%v", err)
		}
		return globals.Exitf(1, "%v", err)
	}

	if err := lock.Write(globals.LockfilePath, lf); err != nil {
		output.Error("write lockfile: %v", err)
		return globals.Exitf(1, "write lockfile: %v", err)
	}

	return nil
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptAddInputs prompts for any inputs not already provided via args/flags/env.
func promptAddInputs(rulesetKey, workspace, apiKey *string) error {
	orange := lipgloss.Color("#FF7800")
	theme := huh.ThemeBase()
	theme.Focused.Title = theme.Focused.Title.Foreground(orange).Bold(true)
	theme.Focused.Base = theme.Focused.Base.BorderForeground(orange)
	theme.Focused.Description = theme.Focused.Description.Faint(true)

	var fields []huh.Field

	if *rulesetKey == "" {
		fields = append(fields, huh.NewInput().
			Title("Ruleset key").
			Description("The unique identifier of the ruleset to add.").
			Value(rulesetKey).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("ruleset key is required")
				}
				return nil
			}),
		)
	}

	if *workspace == "" || *workspace == "default" {
		fields = append(fields, huh.NewInput().
			Title("Workspace").
			Description("The workspace this ruleset belongs to (default: default).").
			Value(workspace),
		)
	}

	if *apiKey == "" {
		fields = append(fields, huh.NewInput().
			Title("API key").
			Description("Your rk_* API key for registry authentication.").
			EchoMode(huh.EchoModePassword).
			Value(apiKey),
		)
	}

	if len(fields) == 0 {
		return nil
	}

	if !isTTY() {
		return nil
	}

	return huh.NewForm(huh.NewGroup(fields...)).WithTheme(theme).Run()
}
