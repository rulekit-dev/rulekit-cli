package wizard

import (
	"bytes"
	"strings"
	"testing"
)

// newTestPrompter returns a Prompter backed by the given input string.
// IsTTY() will return false for piped input, so all prompts use defaults.
func newTestPrompter(input string) (*Prompter, *bytes.Buffer) {
	out := &bytes.Buffer{}
	p := &Prompter{
		In:  strings.NewReader(input),
		Out: out,
	}
	return p, out
}

// --- PromptSelect ---

func TestPromptSelect_DefaultOnEmptyInput(t *testing.T) {
	p, _ := newTestPrompter("")
	idx, err := p.PromptSelect("Choose:", []string{"A", "B", "C"}, 1)
	if err != nil {
		t.Fatalf("PromptSelect: %v", err)
	}
	// Non-TTY: always returns defaultIdx
	if idx != 1 {
		t.Errorf("got idx %d, want 1", idx)
	}
}

func TestPromptSelect_NonTTY_ReturnsDefault(t *testing.T) {
	p, _ := newTestPrompter("2\n")
	// Even with "2" as input, non-TTY returns the default.
	idx, err := p.PromptSelect("Choose:", []string{"A", "B"}, 0)
	if err != nil {
		t.Fatalf("PromptSelect: %v", err)
	}
	if idx != 0 {
		t.Errorf("non-TTY: got idx %d, want 0 (default)", idx)
	}
}

// --- PromptText ---

func TestPromptText_DefaultOnNonTTY(t *testing.T) {
	p, _ := newTestPrompter("")
	val, err := p.PromptText("Enter something", "my-default")
	if err != nil {
		t.Fatalf("PromptText: %v", err)
	}
	if val != "my-default" {
		t.Errorf("got %q, want %q", val, "my-default")
	}
}

func TestPromptText_EmptyDefault(t *testing.T) {
	p, _ := newTestPrompter("")
	val, err := p.PromptText("Enter something", "")
	if err != nil {
		t.Fatalf("PromptText: %v", err)
	}
	if val != "" {
		t.Errorf("got %q, want empty", val)
	}
}

// --- PromptSecret ---

func TestPromptSecret_NonTTY_ReturnsEmpty(t *testing.T) {
	p, _ := newTestPrompter("supersecret\n")
	// Non-TTY: returns empty string (no echo reading in non-interactive mode).
	val, err := p.PromptSecret("Password")
	if err != nil {
		t.Fatalf("PromptSecret: %v", err)
	}
	if val != "" {
		t.Errorf("non-TTY PromptSecret: got %q, want empty", val)
	}
}

// --- PromptConfirm ---

func TestPromptConfirm_DefaultYes_NonTTY(t *testing.T) {
	p, _ := newTestPrompter("")
	ok, err := p.PromptConfirm("Continue?", true)
	if err != nil {
		t.Fatalf("PromptConfirm: %v", err)
	}
	if !ok {
		t.Error("defaultYes=true on non-TTY: expected true")
	}
}

func TestPromptConfirm_DefaultNo_NonTTY(t *testing.T) {
	p, _ := newTestPrompter("")
	ok, err := p.PromptConfirm("Continue?", false)
	if err != nil {
		t.Fatalf("PromptConfirm: %v", err)
	}
	if ok {
		t.Error("defaultYes=false on non-TTY: expected false")
	}
}
