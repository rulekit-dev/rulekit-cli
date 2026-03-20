package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateCompose_SQLite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	opts := ComposeOptions{
		RegistryPort:   8080,
		DashboardPort:  3000,
		UsePostgres:    false,
		RegistryImage:  "ghcr.io/rulekit-dev/rulekit-registry:latest",
		DashboardImage: "ghcr.io/rulekit-dev/rulekit-dashboard:latest",
	}

	if err := GenerateCompose(opts); err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}

	data, err := os.ReadFile(ComposePath())
	if err != nil {
		t.Fatalf("read compose file: %v", err)
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		t.Fatalf("parse compose YAML: %v", err)
	}

	if _, ok := cf.Services["registry"]; !ok {
		t.Error("missing registry service")
	}
	if _, ok := cf.Services["dashboard"]; !ok {
		t.Error("missing dashboard service")
	}
	if _, ok := cf.Services["postgres"]; ok {
		t.Error("postgres service should not be present for SQLite mode")
	}

	reg := cf.Services["registry"]
	// Registry config flows through env_file, not inline environment.
	if len(reg.EnvFile) == 0 {
		t.Error("registry service missing env_file")
	}
	if len(reg.Ports) == 0 || !strings.Contains(reg.Ports[0], "8080") {
		t.Errorf("registry port missing 8080: %v", reg.Ports)
	}
	if reg.HealthCheck == nil {
		t.Error("registry healthcheck missing")
	}

	if _, ok := cf.Volumes["registry-data"]; !ok {
		t.Error("registry-data volume missing")
	}
}

func TestGenerateCompose_Postgres(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	opts := ComposeOptions{
		RegistryPort:   9090,
		DashboardPort:  4000,
		UsePostgres:    true,
		RegistryImage:  "ghcr.io/rulekit-dev/rulekit-registry:latest",
		DashboardImage: "ghcr.io/rulekit-dev/rulekit-dashboard:latest",
	}

	if err := GenerateCompose(opts); err != nil {
		t.Fatalf("GenerateCompose: %v", err)
	}

	data, err := os.ReadFile(ComposePath())
	if err != nil {
		t.Fatalf("read compose file: %v", err)
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		t.Fatalf("parse compose YAML: %v", err)
	}

	if _, ok := cf.Services["postgres"]; !ok {
		t.Error("postgres service missing")
	}

	reg := cf.Services["registry"]
	// Registry config flows through env_file, not inline environment.
	if len(reg.EnvFile) == 0 {
		t.Error("registry service missing env_file")
	}
	if len(reg.Ports) == 0 || !strings.Contains(reg.Ports[0], "9090") {
		t.Errorf("registry port missing 9090: %v", reg.Ports)
	}

	pg := cf.Services["postgres"]
	if pg.HealthCheck == nil {
		t.Error("postgres healthcheck missing")
	}

	if _, ok := cf.Volumes["postgres-data"]; !ok {
		t.Error("postgres-data volume missing")
	}
}

func TestGenerateCompose_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	opts := DefaultComposeOptions()
	if err := GenerateCompose(opts); err != nil {
		t.Fatalf("first write: %v", err)
	}

	// Verify no .tmp file is left behind.
	tmp := ComposePath() + ".tmp"
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error("tmp file left behind after atomic write")
	}

	// File must exist at the expected path.
	if _, err := os.Stat(ComposePath()); err != nil {
		t.Errorf("compose file not found: %v", err)
	}
}

func TestParseDatabaseType(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// SQLite compose
	opts := DefaultComposeOptions()
	opts.UsePostgres = false
	if err := GenerateCompose(opts); err != nil {
		t.Fatalf("generate sqlite compose: %v", err)
	}
	if got := ParseDatabaseType(ComposePath()); got != "sqlite" {
		t.Errorf("SQLite compose: got %q, want %q", got, "sqlite")
	}

	// Postgres compose
	opts.UsePostgres = true
	if err := GenerateCompose(opts); err != nil {
		t.Fatalf("generate postgres compose: %v", err)
	}
	if got := ParseDatabaseType(ComposePath()); got != "postgres" {
		t.Errorf("Postgres compose: got %q, want %q", got, "postgres")
	}
}

func TestParseDatabaseType_MissingFile(t *testing.T) {
	got := ParseDatabaseType(filepath.Join(t.TempDir(), "nonexistent.yml"))
	if got != "sqlite" {
		t.Errorf("missing file: got %q, want %q", got, "sqlite")
	}
}
