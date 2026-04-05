package lock

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type RulesetLock struct {
	Version  int       `json:"version"`
	Checksum string    `json:"checksum"`
	PulledAt time.Time `json:"pulled_at"`
}

type LockFile struct {
	mu        sync.RWMutex
	Registry  string                 `json:"registry"`
	Dashboard string                 `json:"dashboard,omitempty"`
	Workspace string                 `json:"workspace"`
	Rulesets  map[string]RulesetLock `json:"rulesets"`
}

func Read(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read lockfile: %w", err)
	}

	var lf LockFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("parse lockfile: %w", err)
	}

	if lf.Rulesets == nil {
		lf.Rulesets = make(map[string]RulesetLock)
	}

	return &lf, nil
}

func Write(path string, lf *LockFile) error {
	lf.mu.RLock()
	defer lf.mu.RUnlock()

	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write lockfile: %w", err)
	}

	return nil
}

func Empty(registry, workspace string) *LockFile {
	return &LockFile{
		Registry:  registry,
		Workspace: workspace,
		Rulesets:  make(map[string]RulesetLock),
	}
}
