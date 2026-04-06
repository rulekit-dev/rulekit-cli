package stack

import "github.com/spf13/cobra"

// Register adds all stack management commands to the root command.
func Register(root *cobra.Command) {
	root.AddCommand(
		onboardCmd,
		upCmd,
		downCmd,
		restartCmd,
		statusCmd,
		dashboardCmd,
		logsCmd,
		upgradeCmd,
		uninstallCmd,
	)
}
