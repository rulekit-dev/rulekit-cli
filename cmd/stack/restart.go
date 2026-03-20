package stack

import (
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Stop and restart the RuleKit stack (preserves existing config)",
	RunE:  runRestart,
}

func runRestart(cmd *cobra.Command, args []string) error {
	if err := docker.CheckDocker(); err != nil {
		output.Error("%v", err)
		return globals.Exitf(1, "%v", err)
	}

	// Stop silently.
	stopStack(true)

	// Start using the existing compose file — do not regenerate.
	return startStack(false)
}
