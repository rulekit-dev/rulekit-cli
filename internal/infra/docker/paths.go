package docker

import (
	"os"
	"path/filepath"
)

// ComposeDir returns the path to ~/.rulekit/compose/.
func ComposeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".rulekit", "compose")
}

// ComposePath returns the path to ~/.rulekit/compose/docker-compose.yml.
func ComposePath() string {
	return filepath.Join(ComposeDir(), "docker-compose.yml")
}

// EnvPath returns the path to ~/.rulekit/compose/.env.
func EnvPath() string {
	return filepath.Join(ComposeDir(), ".env")
}

// DataDir returns the path to ~/.rulekit/data/.
func DataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".rulekit", "data")
}

// SQLiteDBPath returns the displayed path for the SQLite database file.
func SQLiteDBPath() string {
	return filepath.Join(DataDir(), "rulekit.db")
}
