package stack

import (
	"errors"
	"os"

	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/browser"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/health"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the RuleKit dashboard in your default browser",
	GroupID: "stack",
	RunE:  runDashboard,
}

func runDashboard(cmd *cobra.Command, args []string) error {
	dashboardURL := "http://localhost:3001"

	lf, err := lock.Read(globals.LockfilePath)
	if err == nil && lf.Dashboard != "" {
		dashboardURL = lf.Dashboard
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		output.Error("read lockfile: %v", err)
	}

	if !health.Reachable(dashboardURL) {
		output.Warn("warning: dashboard does not appear to be running. start with 'rulekit up'.")
	}

	output.Info("opening dashboard · %s", dashboardURL)
	browser.Open(dashboardURL) //nolint:errcheck

	return nil
}
