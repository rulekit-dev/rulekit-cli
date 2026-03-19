package cmd

import (
	"context"
	"fmt"

	"github.com/rulekit-dev/rulekit-cli/internal/config"
	"github.com/rulekit-dev/rulekit-cli/internal/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/output"
	"github.com/rulekit-dev/rulekit-cli/internal/registry"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show update status of all locked rulesets",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	lf, err := lock.Read(lockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return exitErr(1, "load lockfile: %v", err)
	}

	cfg := config.Resolve(flagRegistry, flagNamespace, flagDir, flagToken, lf.Registry, lf.Namespace)
	client := registry.NewClient(cfg.RegistryURL, cfg.Token)

	for key, entry := range lf.Rulesets {
		meta, err := client.GetLatestVersion(context.Background(), key, cfg.Namespace)
		if err != nil {
			output.Fail(fmt.Sprintf("%s: unverified (error: %v)", key, err))
			continue
		}

		if entry.Version >= meta.Version {
			output.Success(fmt.Sprintf("%s: up to date (v%d)", key, entry.Version))
		} else {
			output.Warn(fmt.Sprintf("%s: update available (local v%d, latest v%d)", key, entry.Version, meta.Version))
		}
	}

	return nil
}
