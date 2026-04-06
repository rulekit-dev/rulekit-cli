package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rulekit-dev/rulekit-cli/cmd/ruleset"
	"github.com/rulekit-dev/rulekit-cli/cmd/stack"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/spf13/cobra"
)

const asciiArt = `
 ██████╗ ██╗   ██╗██╗     ███████╗██╗  ██╗██╗████████╗
 ██╔══██╗██║   ██║██║     ██╔════╝██║ ██╔╝██║╚══██╔══╝
 ██████╔╝██║   ██║██║     █████╗  █████╔╝ ██║   ██║
 ██╔══██╗██║   ██║██║     ██╔══╝  ██╔═██╗ ██║   ██║
 ██║  ██║╚██████╔╝███████╗███████╗██║  ██╗██║   ██║
 ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚═╝   ╚═╝`

var (
	clrOrange = lipgloss.Color("#FF7800")
	clrMuted  = lipgloss.Color("#666666")
	clrDesc   = lipgloss.Color("#999999")
)

// commandOrder defines the ordered list of commands shown in the interactive menu.
// Sorted by likely execution order: setup → daily use → maintenance.
// section marks the start of a new group; consecutive same-section entries share the header.
var commandOrder = []struct {
	name    string
	section string
}{
	{"onboard", "Setup"},
	{"up", ""},
	{"add", "Rulesets"},
	{"pull", ""},
	{"list", ""},
	{"diff", ""},
	{"verify", ""},
	{"status", "Maintenance"},
	{"dashboard", ""},
	{"logs", ""},
	{"restart", ""},
	{"upgrade", ""},
	{"down", ""},
	{"uninstall", ""},
}

var rootCmd = &cobra.Command{
	Use:           "rulekit",
	Short:         "Rule bundle manager",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if globals.Verbose {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isColorTerm() {
			fmt.Print(buildHelp(cmd))
			return nil
		}
		return runInteractive(cmd)
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

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Print(buildHelp(cmd))
	})
}

const sectionPrefix = "§" // sentinel prefix marking non-selectable section headers

// runInteractive shows the banner and a huh selector to pick and run a command.
func runInteractive(root *cobra.Command) error {
	fmt.Print(renderBanner(true))
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(clrMuted).Render("  Rule as a Service."))
	fmt.Println()

	cmdMap := make(map[string]*cobra.Command)
	for _, sub := range root.Commands() {
		cmdMap[sub.Name()] = sub
	}

	headerStyle := lipgloss.NewStyle().Foreground(clrOrange).Bold(true)
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(clrDesc)

	var options []huh.Option[string]
	for _, entry := range commandOrder {
		if entry.section != "" {
			label := headerStyle.Render("  " + strings.ToUpper(entry.section))
			options = append(options, huh.NewOption(label, sectionPrefix+entry.section))
		}
		sub, ok := cmdMap[entry.name]
		if !ok {
			continue
		}
		label := "  " + cmdStyle.Render(fmt.Sprintf("%-12s", sub.Name())) + " " + descStyle.Render(sub.Short)
		options = append(options, huh.NewOption(label, sub.Name()))
	}

	theme := interactiveTheme()

	for {
		var chosen string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What can I help you with?").
					Options(options...).
					Value(&chosen),
			),
		).WithTheme(theme)

		if err := form.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		if strings.HasPrefix(chosen, sectionPrefix) {
			// User selected a header — re-show the menu.
			continue
		}

		sub, _, err := root.Find([]string{chosen})
		if err != nil || sub == nil {
			return fmt.Errorf("unknown command: %s", chosen)
		}

		return sub.RunE(sub, []string{})
	}
}

func buildHelp(_ *cobra.Command) string {
	color := isColorTerm()
	var b strings.Builder

	b.WriteString(renderBanner(color))
	b.WriteString("\n\n")

	tagline := "  Rule as a Service."
	if color {
		tagline = lipgloss.NewStyle().Foreground(clrMuted).Render(tagline)
	}
	b.WriteString(tagline + "\n\n")

	hint := "  Run 'rulekit [command] --help' for more information."
	if color {
		hint = lipgloss.NewStyle().Foreground(clrMuted).Render(hint)
	}
	b.WriteString(hint + "\n")

	return b.String()
}

func interactiveTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Title = t.Focused.Title.Foreground(clrOrange).Bold(true)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(clrOrange)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(clrOrange).Bold(true)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(clrDesc)
	t.Focused.Base = t.Focused.Base.BorderForeground(clrOrange)
	return t
}

func renderBanner(color bool) string {
	if !color {
		return asciiArt
	}
	style := lipgloss.NewStyle().Foreground(clrOrange).Bold(true)
	lines := strings.Split(asciiArt, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = style.Render(l)
		}
	}
	return strings.Join(lines, "\n")
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
