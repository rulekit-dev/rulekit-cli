package stack

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/docker"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Stop containers and remove ~/.rulekit/compose/",
	GroupID: "stack",
	RunE:  runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("rulekit: this will stop all containers and remove ~/.rulekit/compose/.")
	fmt.Println("rulekit: your rulekit.lock and .rulekit/ rule files will NOT be removed.")
	fmt.Print("continue? [y/N] ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())

	if answer != "y" && answer != "Y" {
		output.Info("aborted.")
		return nil
	}

	composePath := docker.ComposePath()
	client := docker.NewClient(composePath)
	client.DownVolumes() //nolint:errcheck

	if err := os.RemoveAll(docker.ComposeDir()); err != nil {
		output.Error("remove compose dir: %v", err)
		return globals.Exitf(1, "remove compose dir: %v", err)
	}

	output.Info("stack removed. rulekit.lock and rule files are intact.")
	return nil
}
