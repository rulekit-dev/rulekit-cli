package cmd

import (
	"errors"
	"log/slog"
	"os"

	"github.com/rulekit-dev/rulekit-cli/cmd/ruleset"
	"github.com/rulekit-dev/rulekit-cli/cmd/stack"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rulekit",
	Short: "rulekit-cli pulls and manages rule bundles from the rulekit-registry",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if globals.Verbose {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}
		return nil
	},
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		var exitErr *globals.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}
		return 1
	}
	return 0
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globals.Registry, "registry", "", "Registry base URL")
	rootCmd.PersistentFlags().StringVar(&globals.Namespace, "namespace", "", "Namespace (default: \"default\")")
	rootCmd.PersistentFlags().StringVar(&globals.Dir, "dir", "", "Local output directory (default: .rulekit)")
	rootCmd.PersistentFlags().StringVar(&globals.Token, "token", "", "Bearer token")
	rootCmd.PersistentFlags().BoolVar(&globals.Verbose, "verbose", false, "Enable structured logging")

	stack.Register(rootCmd)
	ruleset.Register(rootCmd)
}
