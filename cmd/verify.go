package cmd

import (
	"errors"
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
		os.Exit(1)
	}

	mismatch := false
	for key, entry := range lf.Rulesets {
		cfg := resolveDir()
		dslPath := fmt.Sprintf("%s/%s/dsl.json", cfg, key)
		err := bundle.VerifyChecksum(dslPath, entry.Checksum)
		if err != nil {
			var csErr *bundle.ChecksumMismatchError
			if errors.As(err, &csErr) {
				output.Fail(fmt.Sprintf("%s: %v", key, err))
			} else {
				output.Fail(fmt.Sprintf("%s: %v", key, err))
			}
			mismatch = true
		} else {
			output.Success(fmt.Sprintf("%s: ok", key))
		}
	}

	if mismatch {
		os.Exit(2)
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
