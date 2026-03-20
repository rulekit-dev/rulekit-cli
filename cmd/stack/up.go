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

var (
	upPostgres       bool
	upPort           int
	upDashboardPort  int
	upRegistryImage  string
	upDashboardImage string
	upYes            bool
	upReconfigure    bool
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the RuleKit stack (registry + dashboard) via Docker",
	RunE:  runUp,
}

func init() {
	upCmd.Flags().BoolVar(&upPostgres, "postgres", false, "Use Postgres instead of SQLite")
	upCmd.Flags().IntVar(&upPort, "port", 8080, "Registry port")
	upCmd.Flags().IntVar(&upDashboardPort, "dashboard-port", 3001, "Dashboard port")
	upCmd.Flags().StringVar(&upRegistryImage, "registry-image", "ghcr.io/rulekit-dev/rulekit-registry:latest", "Registry Docker image")
	upCmd.Flags().StringVar(&upDashboardImage, "dashboard-image", "ghcr.io/rulekit-dev/rulekit-dashboard:latest", "Dashboard Docker image")
	upCmd.Flags().BoolVar(&upYes, "yes", false, "Skip wizard and accept all defaults (for CI/scripted use)")
	upCmd.Flags().BoolVar(&upReconfigure, "reconfigure", false, "Re-run the setup wizard even if config already exists")
}

func runUp(cmd *cobra.Command, args []string) error {
	return startStack(true)
}

func startStack(regenerate bool) error {
	if err := docker.CheckDocker(); err != nil {
		output.Error("%v", err)
		return globals.Exitf(1, "%v", err)
	}

	envPath := docker.EnvPath()
	composePath := docker.ComposePath()

	if regenerate {
		cfg, err := resolveConfig(envPath)
		if err != nil {
			if errors.Is(err, wizard.ErrCancelled) {
				output.Info("setup cancelled. nothing was saved.")
				return nil
			}
			output.Error("%v", err)
			return globals.Exitf(1, "%v", err)
		}

		if err := wizard.WriteEnv(envPath, cfg.ToEnv()); err != nil {
			output.Error("write .env: %v", err)
			return globals.Exitf(1, "write .env: %v", err)
		}

		opts := docker.ComposeOptions{
			RegistryPort:   cfg.RegistryPort,
			DashboardPort:  cfg.DashboardPort,
			UsePostgres:    cfg.Store == "postgres",
			RegistryImage:  upRegistryImage,
			DashboardImage: upDashboardImage,
		}
		if err := docker.GenerateCompose(opts); err != nil {
			output.Error("generate compose: %v", err)
			return globals.Exitf(1, "generate compose: %v", err)
		}

		if upReconfigure {
			output.Info("stopping stack…")
			docker.NewClient(composePath).DownSilent()
			output.Info("writing new configuration…")
			output.Info("starting stack…")
		}
	}

	client := docker.NewClient(composePath)

	if !upReconfigure {
		running, _ := client.IsRunning()
		if running {
			output.Info("stack is already running. use 'rulekit restart' to restart.")
			return nil
		}
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

func resolveConfig(envPath string) (*wizard.StackConfig, error) {
	envExists := wizard.EnvExists(envPath)

	switch {
	case upYes:
		return wizard.RunWithDefaults(wizard.Default), nil
	case upReconfigure:
		var existing *wizard.StackConfig
		if envExists {
			if loaded, err := wizard.LoadFromEnv(envPath); err == nil {
				existing = loaded
			}
		}
		return wizard.RunWizard(wizard.Default, existing)
	case envExists:
		return wizard.LoadFromEnv(envPath)
	default:
		return wizard.RunWizard(wizard.Default, nil)
	}
}

func resolveStackURLs() (registryURL, dashboardURL string) {
	registryURL = fmt.Sprintf("http://localhost:%d", upPort)
	dashboardURL = fmt.Sprintf("http://localhost:%d", upDashboardPort)

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
