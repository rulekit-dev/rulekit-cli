package ruleset

import (
	"fmt"
	"os"

	"github.com/rulekit-dev/rulekit-cli/internal/domain/bundle"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify checksums of locally pulled rule bundles",
	RunE:  runVerify,
}

func runVerify(cmd *cobra.Command, args []string) error {
	lf, err := lock.Read(globals.LockfilePath)
	if err != nil {
		output.Error("load lockfile: %v", err)
		return globals.Exitf(1, "load lockfile: %v", err)
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
		return globals.Exitf(2, "checksum verification failed")
	}

	output.Info("all checksums verified (%d rulesets)", len(lf.Rulesets))
	return nil
}

func resolveDir() string {
	if globals.Dir != "" {
		return globals.Dir
	}
	if v := os.Getenv("RULEKIT_DIR"); v != "" {
		return v
	}
	return ".rulekit"
}
