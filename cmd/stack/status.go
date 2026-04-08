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
	Use:     "status",
	Short:   "Infra health check (registry, dashboard, db) and ruleset update status",
	GroupID: "stack",
	RunE:    runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	composePath := docker.ComposePath()
	registryURL := "http://localhost:8080"
	dashboardURL := "http://localhost:3001"

	lf, lockErr := lock.Read(globals.LockfilePath)
	if lockErr == nil {
		if lf.Registry != "" {
			registryURL = lf.Registry
		}
		if lf.Dashboard != "" {
			dashboardURL = lf.Dashboard
		}
	}

	fmt.Println(output.Label("  Infrastructure"))
	fmt.Println()

	infraOK := true
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	registryHR, registryErr := health.Check(registryURL)
	if registryErr == nil {
		ver := registryHR.Version
		if ver == "" {
			ver = "unknown"
		}
		fmt.Fprintf(w, "  %s\t%s  %s\n",
			output.Label("registry"),
			output.SymOK(),
			output.Highlight(registryURL)+" "+output.Muted("· "+ver),
		)
	} else {
		fmt.Fprintf(w, "  %s\t%s  %s\n",
			output.Label("registry"),
			output.SymFail(),
			output.Muted("not running"),
		)
		infraOK = false
	}

	if health.Reachable(dashboardURL) {
		fmt.Fprintf(w, "  %s\t%s  %s\n",
			output.Label("dashboard"),
			output.SymOK(),
			output.Highlight(dashboardURL),
		)
	} else {
		fmt.Fprintf(w, "  %s\t%s  %s\n",
			output.Label("dashboard"),
			output.SymFail(),
			output.Muted("not running"),
		)
		infraOK = false
	}

	dbType := docker.ParseDatabaseType(composePath)
	if dbType == "postgres" {
		client := docker.NewClient(composePath)
		if client.IsServiceRunning("postgres") {
			fmt.Fprintf(w, "  %s\t%s  %s\n",
				output.Label("database"),
				output.SymOK(),
				output.Muted("postgres"),
			)
		} else {
			fmt.Fprintf(w, "  %s\t%s  %s\n",
				output.Label("database"),
				output.SymFail(),
				output.Muted("postgres · not running"),
			)
			infraOK = false
		}
	} else {
		fmt.Fprintf(w, "  %s\t%s  %s\n",
			output.Label("database"),
			output.SymOK(),
			output.Muted("sqlite · "+docker.SQLiteDBPath()),
		)
	}

	w.Flush()

	if lockErr == nil && len(lf.Rulesets) > 0 {
		fmt.Println()
		fmt.Println(output.Label("  Rulesets"))
		fmt.Println()

		cfg, err := config.ResolveInteractive(globals.Registry, globals.Workspace, globals.Dir, globals.APIKey, lf.Registry, lf.Workspace)
		if err != nil {
			return err
		}
		regClient := registry.NewClient(cfg.RegistryURL, cfg.APIKey)

		w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for key, entry := range lf.Rulesets {
			meta, err := regClient.GetLatestVersion(context.Background(), key, cfg.Workspace)
			if err != nil {
				fmt.Fprintf(w2, "  %s\t%s  %s\n",
					output.Label(key),
					output.SymFail(),
					output.Muted("not found in registry"),
				)
				continue
			}
			if entry.Version >= meta.Version {
				fmt.Fprintf(w2, "  %s\t%s  %s\n",
					output.Label(key),
					output.SymOK(),
					output.Muted(fmt.Sprintf("v%d · up to date", entry.Version)),
				)
			} else {
				fmt.Fprintf(w2, "  %s\t%s  %s\n",
					output.Label(key),
					output.SymWarn(),
					output.Warn2(fmt.Sprintf("v%d → v%d available", entry.Version, meta.Version)),
				)
			}
		}
		w2.Flush()
	} else if errors.Is(lockErr, os.ErrNotExist) {
		// No lockfile — skip ruleset section silently.
	}

	fmt.Println()

	if !infraOK {
		return globals.Exitf(1, "one or more infra checks failed")
	}
	return nil
}
