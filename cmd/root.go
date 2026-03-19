package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagRegistry  string
	flagNamespace string
	flagDir       string
	flagToken     string
	flagVerbose   bool
)

var rootCmd = &cobra.Command{
	Use:   "rulekit",
	Short: "rulekit-cli pulls and manages rule bundles from the RuleKit registry",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if flagVerbose {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}
		return nil
	},
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagRegistry, "registry", "", "Registry base URL")
	rootCmd.PersistentFlags().StringVar(&flagNamespace, "namespace", "", "Namespace (default: \"default\")")
	rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "", "Local output directory (default: .rulekit)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Bearer token")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable structured logging")
}
