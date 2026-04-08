package ruleset

import (
	"fmt"
	"os"

	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <key>",
	Short:   "Remove a ruleset from the lockfile and delete local files",
	GroupID: "ruleset",
	RunE:    runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		output.Error("usage: rulekit remove <key>")
		return globals.Exitf(1, "missing required argument: key")
	}
	key := args[0]

	lf, err := lock.Read(globals.LockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return globals.Exitf(1, "load lockfile: %v", err)
	}

	if _, ok := lf.Rulesets[key]; !ok {
		output.Error("ruleset %q not found in lockfile", key)
		return globals.Exitf(1, "ruleset %q not found in lockfile", key)
	}

	dir := resolveDir()
	rulesetDir := fmt.Sprintf("%s/%s", dir, key)
	if err := os.RemoveAll(rulesetDir); err != nil {
		output.Error("remove directory %s: %v", rulesetDir, err)
		return globals.Exitf(1, "remove directory %s: %v", rulesetDir, err)
	}

	delete(lf.Rulesets, key)

	if err := lock.Write(globals.LockfilePath, lf); err != nil {
		output.Error("write lockfile: %v", err)
		return globals.Exitf(1, "write lockfile: %v", err)
	}

	output.Info("removed %s", key)
	return nil
}
