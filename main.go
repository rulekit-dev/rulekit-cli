package main

import (
	"os"

	"github.com/rulekit-dev/rulekit-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
