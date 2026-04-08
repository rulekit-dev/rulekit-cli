package ruleset

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rulekit-dev/rulekit-cli/internal/app/config"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/bundle"
	"github.com/rulekit-dev/rulekit-cli/internal/domain/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/globals"
	"github.com/rulekit-dev/rulekit-cli/internal/infra/registry"
	"github.com/rulekit-dev/rulekit-cli/internal/ui/output"
	"github.com/spf13/cobra"
)

var (
	pullKey     string
	pullVersion string
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull rule bundles from the registry",
	GroupID: "ruleset",
	RunE:  runPull,
}

func init() {
	pullCmd.Flags().StringVar(&pullKey, "key", "", "Ruleset key to pull")
	pullCmd.Flags().StringVar(&pullVersion, "version", "", "Version to pull (number or \"latest\")")
}

func runPull(cmd *cobra.Command, args []string) error {
	lf, err := loadOrEmptyLock("")
	if err != nil {
		output.Error("%v", err)
		return globals.Exitf(1, "%v", err)
	}

	cfg, err := config.ResolveInteractive(globals.Registry, globals.Workspace, globals.Dir, globals.APIKey, lf.Registry, lf.Workspace)
	if err != nil {
		return err
	}
	lf.Registry = cfg.RegistryURL
	lf.Workspace = cfg.Workspace

	client := registry.NewClient(cfg.RegistryURL, cfg.APIKey)

	var keys []string
	if pullKey != "" {
		keys = []string{pullKey}
	} else {
		for k := range lf.Rulesets {
			keys = append(keys, k)
		}
		if len(keys) == 0 {
			output.Error("no rulesets in lockfile; use 'rulekit add <key>' first")
			return globals.Exitf(1, "no rulesets in lockfile")
		}
	}

	code := 0
	for _, key := range keys {
		ver := resolveVersion(key, pullVersion, lf)
		if err := pullOne(context.Background(), client, lf, cfg.Dir, key, ver, cfg.Workspace); err != nil {
			output.Error("%v", err)
			var csErr *bundle.ChecksumMismatchError
			if errors.As(err, &csErr) {
				code = 2
			} else {
				code = 1
			}
		}
	}

	if err := lock.Write(globals.LockfilePath, lf); err != nil {
		output.Error("write lockfile: %v", err)
		return globals.Exitf(1, "write lockfile: %v", err)
	}

	if code != 0 {
		return globals.Exitf(code, "one or more pulls failed")
	}
	return nil
}

func resolveVersion(key, flagVer string, lf *lock.LockFile) string {
	if flagVer != "" {
		return flagVer
	}
	if entry, ok := lf.Rulesets[key]; ok {
		return strconv.Itoa(entry.Version)
	}
	return "latest"
}

func pullOne(ctx context.Context, client *registry.Client, lf *lock.LockFile, dir, key, version, workspace string) error {
	output.Info("pulling %s@%s…", key, version)

	zipBytes, err := client.DownloadBundle(ctx, key, version, workspace)
	if err != nil {
		return fmt.Errorf("download %s: %w", key, err)
	}

	destDir := fmt.Sprintf("%s/%s", dir, key)
	manifest, err := bundle.Extract(zipBytes, destDir)
	if err != nil {
		return fmt.Errorf("extract %s: %w", key, err)
	}

	dslPath := fmt.Sprintf("%s/%s/dsl.json", dir, key)
	if err := bundle.VerifyChecksum(dslPath, manifest.Checksum); err != nil {
		return err
	}

	lf.Rulesets[key] = lock.RulesetLock{
		Version:  manifest.Version,
		Checksum: manifest.Checksum,
		PulledAt: time.Now().UTC(),
	}

	output.Success(fmt.Sprintf("locked %s %s",
		output.Highlight(key),
		output.Muted(fmt.Sprintf("v%d · %s", manifest.Version, manifest.Checksum)),
	))
	return nil
}

func loadOrEmptyLock(registryURL string) (*lock.LockFile, error) {
	lf, err := lock.Read(globals.LockfilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return lock.Empty(registryURL, ""), nil
		}
		return nil, err
	}
	return lf, nil
}
