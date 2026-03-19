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
