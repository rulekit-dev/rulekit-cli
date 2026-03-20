package stack

import (
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop all RuleKit stack containers",
	RunE:  runDown,
}

func runDown(cmd *cobra.Command, args []string) error {
	return stopStack(false)
}

// stopStack stops the stack. If silent is true, output is suppressed (used by restart).
func stopStack(silent bool) error {
	client := docker.NewClient(docker.ComposePath())
	if silent {
		client.DownSilent()
		return nil
	}
	if err := client.Down(); err != nil {
		output.Error("docker compose down: %v", err)
		return globals.Exitf(1, "docker compose down: %v", err)
	}
	output.Info("stack stopped.")
	return nil
}
