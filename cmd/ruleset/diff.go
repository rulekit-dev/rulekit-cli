package ruleset

import (
	"context"
	"fmt"

	"github.com/rulekit-dev/rulekit-cli/internal/app/config"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/registry"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare local locked versions vs registry latest",
	RunE:  runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {
	lf, err := lock.Read(globals.LockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return globals.Exitf(1, "load lockfile: %v", err)
	}

	cfg := config.Resolve(globals.Registry, globals.Workspace, globals.Dir, globals.Token, lf.Registry, lf.Workspace)
	client := registry.NewClient(cfg.RegistryURL, cfg.Token)

	for key, entry := range lf.Rulesets {
		meta, err := client.GetLatestVersion(context.Background(), key, cfg.Workspace)
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
