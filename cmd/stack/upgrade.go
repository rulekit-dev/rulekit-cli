package stack

import (
	"context"
	"fmt"
	"time"

	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/health"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Pull latest images and perform a rolling restart",
	GroupID: "stack",
	RunE:  runUpgrade,
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	if err := docker.CheckDocker(); err != nil {
		output.Error("%v", err)
		return globals.Exitf(1, "%v", err)
	}

	composePath := docker.ComposePath()
	client := docker.NewClient(composePath)

	output.Info("pulling latest images…")
	if err := client.Pull(); err != nil {
		output.Error("docker compose pull: %v", err)
		return globals.Exitf(1, "docker compose pull: %v", err)
	}

	if err := client.Up(); err != nil {
		output.Error("docker compose up: %v", err)
		return globals.Exitf(1, "docker compose up: %v", err)
	}

	registryURL := "http://localhost:8080"
	dashboardURL := "http://localhost:3000"

	if lf, err := lock.Read(globals.LockfilePath); err == nil {
		if lf.Registry != "" {
			registryURL = lf.Registry
		}
		if lf.Dashboard != "" {
			dashboardURL = lf.Dashboard
		}
	}

	output.Info("waiting for registry to be ready…")
	if _, err := health.Poll(context.Background(), registryURL, 30*time.Second); err != nil {
		output.Error("registry did not become healthy: %v", err)
		return globals.Exitf(1, "registry health check failed")
	}

	output.Info("images updated.")
	output.Info("registry  ready · %s", fmt.Sprint(registryURL))
	output.Info("dashboard ready · %s", fmt.Sprint(dashboardURL))

	return nil
}
