package stack

import (
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var (
	logsService string
	logsFollow  bool
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail container logs for the RuleKit stack",
	GroupID: "stack",
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().StringVar(&logsService, "service", "", "Filter to one service (registry, dashboard, postgres)")
	logsCmd.Flags().BoolVar(&logsFollow, "follow", true, "Follow log output")
}

func runLogs(cmd *cobra.Command, args []string) error {
	client := docker.NewClient(docker.ComposePath())
	if err := client.Logs(logsFollow, logsService); err != nil {
		output.Error("docker compose logs: %v", err)
		return globals.Exitf(1, "docker compose logs: %v", err)
	}
	return nil
}
