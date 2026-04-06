package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	orange  = lipgloss.Color("#FF7800")
	green   = lipgloss.Color("#22C55E")
	red     = lipgloss.Color("#EF4444")
	yellow  = lipgloss.Color("#F59E0B")
	muted   = lipgloss.Color("#666666")
	white   = lipgloss.Color("#FFFFFF")
)

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// Info prints a general status message prefixed with the orange rulekit brand marker.
func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if isTerminal() {
		prefix := lipgloss.NewStyle().Foreground(orange).Bold(true).Render("▸")
		msg = lipgloss.NewStyle().Foreground(white).Render(msg)
		fmt.Println(prefix + " " + msg)
	} else {
		fmt.Println("rulekit: " + msg)
	}
}

// Error prints an error message to stderr.
func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if isTerminal() {
		prefix := lipgloss.NewStyle().Foreground(red).Bold(true).Render("✗")
		msg = lipgloss.NewStyle().Foreground(red).Render(msg)
		fmt.Fprintln(os.Stderr, prefix+" "+msg)
	} else {
		fmt.Fprintln(os.Stderr, "rulekit: error: "+msg)
	}
}

// Success prints a success message with a green checkmark.
func Success(msg string) {
	if isTerminal() {
		prefix := lipgloss.NewStyle().Foreground(green).Bold(true).Render("✓")
		fmt.Println(prefix + " " + lipgloss.NewStyle().Foreground(white).Render(msg))
	} else {
		fmt.Println("✓ " + msg)
	}
}

// Warn prints a warning message with a yellow arrow.
func Warn(msg string) {
	if isTerminal() {
		prefix := lipgloss.NewStyle().Foreground(yellow).Bold(true).Render("↑")
		fmt.Println(prefix + " " + lipgloss.NewStyle().Foreground(yellow).Render(msg))
	} else {
		fmt.Println("↑ " + msg)
	}
}

// Fail prints a failure message with a red cross.
func Fail(msg string) {
	if isTerminal() {
		prefix := lipgloss.NewStyle().Foreground(red).Bold(true).Render("✗")
		fmt.Println(prefix + " " + lipgloss.NewStyle().Foreground(red).Render(msg))
	} else {
		fmt.Println("✗ " + msg)
	}
}

// SymOK returns a colored ✓ on terminals, plain otherwise.
func SymOK() string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(green).Bold(true).Render("✓")
	}
	return "✓"
}

// SymWarn returns a colored ↑ on terminals, plain otherwise.
func SymWarn() string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(yellow).Bold(true).Render("↑")
	}
	return "↑"
}

// SymFail returns a colored ✗ on terminals, plain otherwise.
func SymFail() string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(red).Bold(true).Render("✗")
	}
	return "✗"
}

// Label returns a styled dim label string (for table headers etc).
func Label(s string) string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(muted).Render(s)
	}
	return s
}

// Highlight returns text in orange (for URLs, key values).
func Highlight(s string) string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(orange).Render(s)
	}
	return s
}

// Muted returns dimmed text.
func Muted(s string) string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(muted).Render(s)
	}
	return s
}

// Warn2 returns a yellow-styled string without printing it (for inline use in tables).
func Warn2(s string) string {
	if isTerminal() {
		return lipgloss.NewStyle().Foreground(yellow).Render(s)
	}
	return s
}
