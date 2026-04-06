package cmd

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/rulekit-dev/rulekit-cli/cmd/ruleset"
	"github.com/rulekit-dev/rulekit-cli/cmd/stack"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/spf13/cobra"
)

const banner = `
 ██████╗ ██╗   ██╗██╗     ███████╗██╗  ██╗██╗████████╗
 ██╔══██╗██║   ██║██║     ██╔════╝██║ ██╔╝██║╚══██╔══╝
 ██████╔╝██║   ██║██║     █████╗  █████╔╝ ██║   ██║
 ██╔══██╗██║   ██║██║     ██╔══╝  ██╔═██╗ ██║   ██║
 ██║  ██║╚██████╔╝███████╗███████╗██║  ██╗██║   ██║
 ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚═╝   ╚═╝
`

var rootCmd = &cobra.Command{
	Use:   "rulekit",
	Short: "Rule bundle manager",
	Long:  colorBanner(),
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
	rootCmd.PersistentFlags().StringVar(&globals.Workspace, "workspace", "", "Workspace (default: \"default\")")
	rootCmd.PersistentFlags().StringVar(&globals.Dir, "dir", "", "Local output directory (default: .rulekit)")
	rootCmd.PersistentFlags().StringVar(&globals.Token, "token", "", "Bearer token")
	rootCmd.PersistentFlags().BoolVar(&globals.Verbose, "verbose", false, "Enable structured logging")

	rootCmd.AddGroup(
		&cobra.Group{ID: "stack", Title: "Stack"},
		&cobra.Group{ID: "ruleset", Title: "Rulesets"},
	)

	stack.Register(rootCmd)
	ruleset.Register(rootCmd)
}

// colorBanner returns the banner with orange ANSI color if the terminal supports it.
func colorBanner() string {
	if !isColorTerm() {
		return banner + "\n"
	}
	const orange = "\033[38;2;255;120;0m"
	const reset = "\033[0m"
	lines := strings.Split(banner, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = orange + l + reset
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func isColorTerm() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
