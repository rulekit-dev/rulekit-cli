package cmd

import (
	"fmt"
	"os"

	"github.com/rulekit-dev/rulekit-cli/internal/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/output"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <key>",
	Short: "Remove a ruleset from the lockfile and delete local files",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	key := args[0]

	lf, err := lock.Read(lockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return exitErr(1, "load lockfile: %v", err)
	}

	if _, ok := lf.Rulesets[key]; !ok {
		output.Error("ruleset %q not found in lockfile", key)
		return exitErr(1, "ruleset %q not found in lockfile", key)
	}

	dir := resolveDir()
	rulesetDir := fmt.Sprintf("%s/%s", dir, key)
	if err := os.RemoveAll(rulesetDir); err != nil {
		output.Error("remove directory %s: %v", rulesetDir, err)
		return exitErr(1, "remove directory %s: %v", rulesetDir, err)
	}

	delete(lf.Rulesets, key)

	if err := lock.Write(lockfilePath, lf); err != nil {
		output.Error("write lockfile: %v", err)
		return exitErr(1, "write lockfile: %v", err)
	}

	output.Info("removed %s", key)
	return nil
}
