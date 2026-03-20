package wizard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter holds the I/O streams used by all prompt functions.
// Using os.Stdin/os.Stdout by default; replaced in tests.
type Prompter struct {
	In  io.Reader
	Out io.Writer
}

// Default is the package-level prompter backed by real stdin/stdout.
var Default = &Prompter{In: os.Stdin, Out: os.Stdout}

// IsTTY returns true when stdin is an interactive terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// PromptSelect shows a numbered menu and returns the index of the chosen option.
// If stdin is not a TTY, defaultIdx is returned immediately.
func (p *Prompter) PromptSelect(question string, options []string, defaultIdx int) (int, error) {
	fmt.Fprintln(p.Out)
	fmt.Fprintln(p.Out, question)
	fmt.Fprintln(p.Out)
	for i, opt := range options {
		marker := " "
		if i == defaultIdx {
			marker = "*"
		}
		fmt.Fprintf(p.Out, "  [%d]%s %s\n", i+1, marker, opt)
	}
	fmt.Fprintln(p.Out)

	if !IsTTY() {
		return defaultIdx, nil
	}

	scanner := bufio.NewScanner(p.In)
	for {
		fmt.Fprint(p.Out, "  > ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return 0, err
			}
			// EOF — use default.
			return defaultIdx, nil
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return defaultIdx, nil
		}
		for i := range options {
			if input == fmt.Sprintf("%d", i+1) {
				return i, nil
			}
		}
		fmt.Fprintf(p.Out, "  please enter a number between 1 and %d\n", len(options))
	}
}

// PromptText prompts for a free-form string, showing defaultVal in brackets.
// If stdin is not a TTY, defaultVal is returned immediately.
func (p *Prompter) PromptText(question, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(p.Out, "  %s [%s]: ", question, defaultVal)
	} else {
		fmt.Fprintf(p.Out, "  %s: ", question)
	}

	if !IsTTY() {
		fmt.Fprintln(p.Out)
		return defaultVal, nil
	}

	scanner := bufio.NewScanner(p.In)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return defaultVal, nil
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// PromptSecret prompts for a secret value with echo disabled.
// If stdin is not a TTY, returns empty string immediately.
func (p *Prompter) PromptSecret(question string) (string, error) {
	fmt.Fprintf(p.Out, "  %s: ", question)

	if !IsTTY() {
		fmt.Fprintln(p.Out)
		return "", nil
	}

	// Use terminal raw mode to disable echo.
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(p.Out)
	if err != nil {
		return "", fmt.Errorf("read secret: %w", err)
	}
	return string(b), nil
}

// PromptConfirm asks a yes/no question. defaultYes controls which answer [Y/n] or [y/N] is shown.
// If stdin is not a TTY, the default is returned immediately.
func (p *Prompter) PromptConfirm(question string, defaultYes bool) (bool, error) {
	choices := "[y/N]"
	if defaultYes {
		choices = "[Y/n]"
	}
	fmt.Fprintf(p.Out, "  %s %s: ", question, choices)

	if !IsTTY() {
		fmt.Fprintln(p.Out)
		return defaultYes, nil
	}

	scanner := bufio.NewScanner(p.In)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return defaultYes, nil
	}
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return defaultYes, nil
	default:
		return defaultYes, nil
	}
}
