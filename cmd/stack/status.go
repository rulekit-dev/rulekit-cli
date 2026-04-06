package stack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rulekit-dev/rulekit-cli/internal/app/config"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/health"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/registry"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Infra health check (registry, dashboard, db) and ruleset update status",
	GroupID: "stack",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	composePath := docker.ComposePath()
	registryURL := "http://localhost:8080"
	dashboardURL := "http://localhost:3000"

	lf, lockErr := lock.Read(globals.LockfilePath)
	if lockErr == nil {
		if lf.Registry != "" {
			registryURL = lf.Registry
		}
		if lf.Dashboard != "" {
			dashboardURL = lf.Dashboard
		}
	}

	output.Info("checking stack…")
	fmt.Println()

	infraOK := true
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	registryHR, registryErr := health.Check(registryURL)
	if registryErr == nil {
		ver := registryHR.Version
		if ver == "" {
			ver = "unknown"
		}
		fmt.Fprintf(w, "registry\t%s running · %s · %s\n", output.SymOK(), ver, registryURL)
	} else {
		fmt.Fprintf(w, "registry\t%s not running\n", output.SymFail())
		infraOK = false
	}

	if health.Reachable(dashboardURL) {
		fmt.Fprintf(w, "dashboard\t%s running · %s\n", output.SymOK(), dashboardURL)
	} else {
		fmt.Fprintf(w, "dashboard\t%s not running\n", output.SymFail())
		infraOK = false
	}

	dbType := docker.ParseDatabaseType(composePath)
	if dbType == "postgres" {
		client := docker.NewClient(composePath)
		if client.IsServiceRunning("postgres") {
			fmt.Fprintf(w, "database\t%s postgres · running\n", output.SymOK())
		} else {
			fmt.Fprintf(w, "database\t%s postgres · not running\n", output.SymFail())
			infraOK = false
		}
	} else {
		dbPath := docker.SQLiteDBPath()
		fmt.Fprintf(w, "database\t%s sqlite · %s\n", output.SymOK(), dbPath)
	}

	w.Flush()

	if lockErr == nil && len(lf.Rulesets) > 0 {
		fmt.Println()
		output.Info("checking rulesets…")
		fmt.Println()

		cfg := config.Resolve(globals.Registry, globals.Workspace, globals.Dir, globals.Token, lf.Registry, lf.Workspace)
		regClient := registry.NewClient(cfg.RegistryURL, cfg.Token)

		w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for key, entry := range lf.Rulesets {
			meta, err := regClient.GetLatestVersion(context.Background(), key, cfg.Workspace)
			if err != nil {
				fmt.Fprintf(w2, "%s\t%s not found in registry\n", key, output.SymFail())
				continue
			}
			if entry.Version >= meta.Version {
				fmt.Fprintf(w2, "%s\t%s v%d · up to date\n", key, output.SymOK(), entry.Version)
			} else {
				fmt.Fprintf(w2, "%s\t%s v%d → v%d available\n", key, output.SymWarn(), entry.Version, meta.Version)
			}
		}
		w2.Flush()
	} else if errors.Is(lockErr, os.ErrNotExist) {
		// No lockfile — skip ruleset section silently.
	}

	if !infraOK {
		return globals.Exitf(1, "one or more infra checks failed")
	}
	return nil
}
