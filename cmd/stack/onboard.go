package stack

import (
	"errors"

	"github.com/rulekit-dev/rulekit-cli/internal/app/wizard"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/spf13/cobra"
)

var (
	onboardPostgres       bool
	onboardPort           int
	onboardDashboardPort  int
	onboardRegistryImage  string
	onboardDashboardImage string
	onboardYes            bool
	onboardReconfigure    bool
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Configure the RuleKit stack (registry + dashboard)",
	GroupID: "stack",
	RunE:  runOnboard,
}

func init() {
	onboardCmd.Flags().BoolVar(&onboardPostgres, "postgres", false, "Use Postgres instead of SQLite")
	onboardCmd.Flags().IntVar(&onboardPort, "port", 8080, "Registry port")
	onboardCmd.Flags().IntVar(&onboardDashboardPort, "dashboard-port", 3001, "Dashboard port")
	onboardCmd.Flags().StringVar(&onboardRegistryImage, "registry-image", "ghcr.io/rulekit-dev/rulekit-registry:latest", "Registry Docker image")
	onboardCmd.Flags().StringVar(&onboardDashboardImage, "dashboard-image", "ghcr.io/rulekit-dev/rulekit-dashboard:latest", "Dashboard Docker image")
	onboardCmd.Flags().BoolVar(&onboardYes, "yes", false, "Skip wizard and accept all defaults (for CI/scripted use)")
	onboardCmd.Flags().BoolVar(&onboardReconfigure, "reconfigure", false, "Re-run the setup wizard even if config already exists")
}

func runOnboard(cmd *cobra.Command, args []string) error {
	envPath := docker.EnvPath()

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
		RegistryImage:  onboardRegistryImage,
		DashboardImage: onboardDashboardImage,
	}
	if err := docker.GenerateCompose(opts); err != nil {
		output.Error("generate compose: %v", err)
		return globals.Exitf(1, "generate compose: %v", err)
	}

	output.Info("configuration saved. run 'rulekit stack up' to start the stack.")
	return nil
}

func resolveConfig(envPath string) (*wizard.StackConfig, error) {
	envExists := wizard.EnvExists(envPath)

	switch {
	case onboardYes:
		return wizard.RunWithDefaults(wizard.Default), nil
	case onboardReconfigure:
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
