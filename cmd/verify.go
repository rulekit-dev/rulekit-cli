package cmd

import (
	"fmt"
	"os"

	"github.com/rulekit-dev/rulekit-cli/internal/bundle"
	"github.com/rulekit-dev/rulekit-cli/internal/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/output"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify checksums of locally pulled rule bundles",
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) error {
	lf, err := lock.Read(lockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return exitErr(1, "load lockfile: %v", err)
	}

	mismatch := false
	for key, entry := range lf.Rulesets {
		dir := resolveDir()
		dslPath := fmt.Sprintf("%s/%s/dsl.json", dir, key)
		if err := bundle.VerifyChecksum(dslPath, entry.Checksum); err != nil {
			output.Fail(fmt.Sprintf("%s: %v", key, err))
			mismatch = true
		} else {
			output.Success(fmt.Sprintf("%s: ok", key))
		}
	}

	if mismatch {
		return exitErr(2, "checksum verification failed")
	}

	output.Info("all checksums verified (%d rulesets)", len(lf.Rulesets))
	return nil
}

func resolveDir() string {
	if flagDir != "" {
		return flagDir
	}
	if v := os.Getenv("RULEKIT_DIR"); v != "" {
		return v
	}
	return ".rulekit"
}
