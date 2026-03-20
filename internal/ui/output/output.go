package output

import (
	"fmt"
	"os"
)

func Info(format string, args ...any) {
	fmt.Printf("rulekit: "+format+"\n", args...)
}

func Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "rulekit: error: "+format+"\n", args...)
}

func Success(msg string) {
	fmt.Println("rulekit: ✓ " + msg)
}

func Warn(msg string) {
	fmt.Println("rulekit: → " + msg)
}

func Fail(msg string) {
	fmt.Println("rulekit: ✗ " + msg)
}

// isTerminal returns true when stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func colorize(code, s string) string {
	if isTerminal() {
		return code + s + "\033[0m"
	}
	return s
}

// SymOK returns a green ✓ on terminals, plain ✓ otherwise.
func SymOK() string { return colorize("\033[32m", "✓") }

// SymWarn returns a yellow ↑ on terminals, plain ↑ otherwise.
func SymWarn() string { return colorize("\033[33m", "↑") }

// SymFail returns a red ✗ on terminals, plain ✗ otherwise.
func SymFail() string { return colorize("\033[31m", "✗") }
