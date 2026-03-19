package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/rulekit-dev/rulekit-cli/internal/bundle"
	"github.com/rulekit-dev/rulekit-cli/internal/config"
	"github.com/rulekit-dev/rulekit-cli/internal/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/output"
	"github.com/rulekit-dev/rulekit-cli/internal/registry"
	"github.com/spf13/cobra"
)

var (
	addVersion string
)

var addCmd = &cobra.Command{
	Use:   "add <key>",
	Short: "Add a ruleset to the lockfile and pull it",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addVersion, "version", "latest", "Version to pull (number or \"latest\")")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	key := args[0]

	lf, err := loadOrEmptyLock("")
	if err != nil {
		output.Error("%v", err)
		os.Exit(1)
	}

	cfg := config.Resolve(flagRegistry, flagNamespace, flagDir, flagToken, lf.Registry, lf.Namespace)
	lf.Registry = cfg.RegistryURL
	lf.Namespace = cfg.Namespace

	client := registry.NewClient(cfg.RegistryURL, cfg.Token)

	ver := addVersion
	if ver == "" {
		ver = "latest"
	}

	if err := pullOne(context.Background(), client, lf, cfg.Dir, key, ver, cfg.Namespace); err != nil {
		var csErr *bundle.ChecksumMismatchError
		if errors.As(err, &csErr) {
			output.Error("%v", err)
			os.Exit(2)
		}
		output.Error("%v", err)
		os.Exit(1)
	}

	if err := lock.Write(lockfilePath, lf); err != nil {
		output.Error("write lockfile: %v", err)
		os.Exit(1)
	}

	return nil
}
