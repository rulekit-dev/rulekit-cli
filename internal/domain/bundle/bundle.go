package bundle

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Manifest struct {
	Namespace  string    `json:"namespace"`
	RulesetKey string    `json:"ruleset_key"`
	Version    int       `json:"version"`
	Checksum   string    `json:"checksum"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChecksumMismatchError struct {
	Key      string
	Expected string
	Got      string
}

func (e *ChecksumMismatchError) Error() string {
	return fmt.Sprintf("checksum mismatch for %s (expected %s got %s)", e.Key, e.Expected, e.Got)
}

func Extract(zipBytes []byte, destDir string) (*Manifest, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("create dest dir: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}

	var manifest *Manifest

	for _, f := range r.File {
		if err := extractFile(f, destDir); err != nil {
			return nil, err
		}

		if f.Name == "manifest.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("open manifest: %w", err)
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("read manifest: %w", err)
			}
			manifest = &Manifest{}
			if err := json.Unmarshal(data, manifest); err != nil {
				return nil, fmt.Errorf("parse manifest: %w", err)
			}
		}
	}

	if manifest == nil {
		return nil, fmt.Errorf("bundle missing manifest.json")
	}

	return manifest, nil
}

func extractFile(f *zip.File, destDir string) error {
	destPath := filepath.Join(destDir, filepath.FromSlash(f.Name))

	// Guard against zip slip attacks.
	if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path in zip: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %s: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", destPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("write file %s: %w", destPath, err)
	}

	return nil
}

func VerifyChecksum(dslPath string, expectedChecksum string) error {
	data, err := os.ReadFile(dslPath)
	if err != nil {
		return fmt.Errorf("read dsl file: %w", err)
	}

	sum := sha256.Sum256(data)
	got := fmt.Sprintf("sha256:%x", sum)

	if got != expectedChecksum {
		key := filepath.Base(filepath.Dir(dslPath))
		return &ChecksumMismatchError{
			Key:      key,
			Expected: expectedChecksum,
			Got:      got,
		}
	}

	return nil
}
