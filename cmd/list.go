package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rulekit-dev/rulekit-cli/internal/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locked rulesets",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	lf, err := lock.Read(lockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		os.Exit(1)
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
