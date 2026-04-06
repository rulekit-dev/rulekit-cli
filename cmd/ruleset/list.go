package ruleset

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locked rulesets",
	GroupID: "ruleset",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	lf, err := lock.Read(globals.LockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return globals.Exitf(1, "load lockfile: %v", err)
	}

	if len(lf.Rulesets) == 0 {
		output.Info("no rulesets locked")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "key\tversion\tchecksum\tpulled_at")
	for key, entry := range lf.Rulesets {
		fmt.Fprintf(w, "%s\tv%d\t%s\t%s\n",
			key,
			entry.Version,
			entry.Checksum,
			entry.PulledAt.Format("2006-01-02"),
		)
	}
	w.Flush()

	return nil
}
