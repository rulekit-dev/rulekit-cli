package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rulekit-dev/rulekit-cli/internal/bundle"
	"github.com/rulekit-dev/rulekit-cli/internal/config"
	"github.com/rulekit-dev/rulekit-cli/internal/lock"
	"github.com/rulekit-dev/rulekit-cli/internal/output"
	"github.com/rulekit-dev/rulekit-cli/internal/registry"
	"github.com/spf13/cobra"
)

var lockfilePath = "rulekit.lock"

var (
	pullKey     string
	pullVersion string
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull rule bundles from the registry",
	RunE:  runPull,
}

func init() {
	pullCmd.Flags().StringVar(&pullKey, "key", "", "Ruleset key to pull")
	pullCmd.Flags().StringVar(&pullVersion, "version", "", "Version to pull (number or \"latest\")")
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	lf, err := loadOrEmptyLock("")
	if err != nil {
		output.Error("%v", err)
		return exitErr(1, "%v", err)
	}

	cfg := config.Resolve(flagRegistry, flagNamespace, flagDir, flagToken, lf.Registry, lf.Namespace)
	lf.Registry = cfg.RegistryURL
	lf.Namespace = cfg.Namespace

	client := registry.NewClient(cfg.RegistryURL, cfg.Token)

	var keys []string
	if pullKey != "" {
		keys = []string{pullKey}
	} else {
		for k := range lf.Rulesets {
			keys = append(keys, k)
		}
		if len(keys) == 0 {
			output.Error("no rulesets in lockfile; use 'rulekit add <key>' first")
			return exitErr(1, "no rulesets in lockfile")
		}
	}

	code := 0
	for _, key := range keys {
		ver := resolveVersion(key, pullVersion, lf)
		if err := pullOne(context.Background(), client, lf, cfg.Dir, key, ver, cfg.Namespace); err != nil {
			output.Error("%v", err)
			var csErr *bundle.ChecksumMismatchError
			if errors.As(err, &csErr) {
				code = 2
			} else {
				code = 1
			}
		}
	}

	if err := lock.Write(lockfilePath, lf); err != nil {
		output.Error("write lockfile: %v", err)
		return exitErr(1, "write lockfile: %v", err)
	}

	if code != 0 {
		return exitErr(code, "one or more pulls failed")
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

func pullOne(ctx context.Context, client *registry.Client, lf *lock.LockFile, dir, key, version, namespace string) error {
	output.Info("pulling %s@%s…", key, version)

	zipBytes, err := client.DownloadBundle(ctx, key, version, namespace)
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

	output.Info("locked %s v%d · %s", key, manifest.Version, manifest.Checksum)
	return nil
}

func loadOrEmptyLock(registryURL string) (*lock.LockFile, error) {
	lf, err := lock.Read(lockfilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return lock.Empty(registryURL, "default"), nil
		}
		return nil, err
	}
	return lf, nil
}
