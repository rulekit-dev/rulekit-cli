package stack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rulekit-dev/rulekit-cli/internal/app/wizard"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/health"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the RuleKit stack (registry + dashboard) via Docker",
	RunE:  runUp,
}

func runUp(cmd *cobra.Command, args []string) error {
	return startStack()
}

func startStack() error {
	if err := docker.CheckDocker(); err != nil {
		output.Error("%v", err)
		return globals.Exitf(1, "%v", err)
	}

	composePath := docker.ComposePath()
	client := docker.NewClient(composePath)

	running, _ := client.IsRunning()
	if running {
		output.Info("stack is already running. use 'rulekit stack restart' to restart.")
		return nil
	}

	if err := client.Up(); err != nil {
		output.Error("docker compose up: %v", err)
		return globals.Exitf(1, "docker compose up: %v", err)
	}

	registryURL, dashboardURL := resolveStackURLs()

	output.Info("waiting for registry to be ready…")
	if _, err := health.Poll(context.Background(), registryURL, 30*time.Second); err != nil {
		output.Error("registry did not become healthy: %v", err)
		return globals.Exitf(1, "registry health check failed")
	}

	output.Info("registry  ready · %s", registryURL)
	output.Info("dashboard ready · %s", dashboardURL)
	output.Info("run 'rulekit dashboard' to open the editor")

	writeDashboardToLock(registryURL, dashboardURL)

	return nil
}

func resolveStackURLs() (registryURL, dashboardURL string) {
	registryURL = fmt.Sprintf("http://localhost:%d", 8080)
	dashboardURL = fmt.Sprintf("http://localhost:%d", 3001)

	if cfg, err := wizard.LoadFromEnv(docker.EnvPath()); err == nil {
		if cfg.RegistryPort > 0 {
			registryURL = fmt.Sprintf("http://localhost:%d", cfg.RegistryPort)
		}
		if cfg.DashboardPort > 0 {
			dashboardURL = fmt.Sprintf("http://localhost:%d", cfg.DashboardPort)
		}
	}

	return registryURL, dashboardURL
}

func writeDashboardToLock(registryURL, dashboardURL string) {
	lf, err := lock.Read(globals.LockfilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		return
	}
	lf.Registry = registryURL
	lf.Dashboard = dashboardURL
	lock.Write(globals.LockfilePath, lf) //nolint:errcheck
}
