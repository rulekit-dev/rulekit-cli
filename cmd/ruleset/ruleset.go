package ruleset

import "github.com/spf13/cobra"

// Register adds all ruleset management commands to the root command.
func Register(root *cobra.Command) {
	root.AddCommand(
		pullCmd,
		addCmd,
		removeCmd,
		listCmd,
		verifyCmd,
		diffCmd,
	)
}
