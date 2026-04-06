package stack

import (
	"errors"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Stop containers and remove ~/.rulekit/compose/",
	GroupID: "stack",
	RunE:    runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	confirm := false

	warning := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(
		"This will stop all containers and remove ~/.rulekit/compose/.\nYour rulekit.lock and .rulekit/ rule files will NOT be removed.",
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Uninstall RuleKit stack?").
				Description(warning).
				Value(&confirm),
		),
	).WithTheme(uninstallTheme())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			output.Info("aborted.")
			return nil
		}
		return err
	}

	if !confirm {
		output.Info("aborted.")
		return nil
	}

	composePath := docker.ComposePath()
	client := docker.NewClient(composePath)
	client.DownVolumes() //nolint:errcheck

	if err := os.RemoveAll(docker.ComposeDir()); err != nil {
		output.Error("remove compose dir: %v", err)
		return globals.Exitf(1, "remove compose dir: %v", err)
	}

	output.Success("stack removed. rulekit.lock and rule files are intact.")
	return nil
}

func uninstallTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Title = t.Focused.Title.Foreground(lipgloss.Color("#FF7800")).Bold(true)
	t.Focused.Base = t.Focused.Base.BorderForeground(lipgloss.Color("#FF7800"))
	return t
}
