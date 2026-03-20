package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- ToEnv ---

func TestToEnv_SQLite_FS_None(t *testing.T) {
	cfg := &StackConfig{
		Store:         "sqlite",
		DataDir:       "/data",
		BlobStore:     "fs",
		Auth:          "none",
		RegistryPort:  8080,
		DashboardPort: 3001,
	}
	env := cfg.ToEnv()

	assertContains(t, env, "RULEKIT_STORE=sqlite")
	assertContains(t, env, "RULEKIT_DATA_DIR=/data")
	assertContains(t, env, "RULEKIT_BLOB_STORE=fs")
	assertContains(t, env, "RULEKIT_AUTH=none")
	assertContains(t, env, "RULEKIT_ADDR=:8080")
	assertContains(t, env, "RULEKIT_CORS_ORIGINS=http://localhost:3001")

	assertNotContains(t, env, "RULEKIT_DATABASE_URL")
	assertNotContains(t, env, "RULEKIT_S3_")
	assertNotContains(t, env, "RULEKIT_JWT_SECRET")
	assertNotContains(t, env, "RULEKIT_SMTP_")
	assertNotContains(t, env, "RULEKIT_API_KEY")
}

func TestToEnv_Postgres_S3_JWT_SMTP(t *testing.T) {
	cfg := &StackConfig{
		Store:             "postgres",
		DatabaseURL:       "postgres://rulekit:rulekit@localhost:5432/rulekit",
		BlobStore:         "s3",
		S3Bucket:          "my-bucket",
		S3Region:          "us-east-1",
		S3Endpoint:        "https://r2.example.com",
		S3AccessKeyID:     "AKIA123",
		S3SecretAccessKey: "secret",
		Auth:              "jwt",
		JWTSecret:         "abc123",
		AdminEmail:        "admin@example.com",
		SMTPEnabled:       true,
		SMTPHost:          "smtp.example.com",
		SMTPPort:          587,
		SMTPUsername:      "user",
		SMTPPassword:      "pass",
		SMTPFrom:          "noreply@rulekit.dev",
		SMTPUseTLS:        false,
		RegistryPort:      8080,
		DashboardPort:     3001,
	}
	env := cfg.ToEnv()

	assertContains(t, env, "RULEKIT_STORE=postgres")
	assertContains(t, env, "RULEKIT_DATABASE_URL=postgres://rulekit:rulekit@localhost:5432/rulekit")
	assertContains(t, env, "RULEKIT_BLOB_STORE=s3")
	assertContains(t, env, "RULEKIT_S3_BUCKET=my-bucket")
	assertContains(t, env, "RULEKIT_S3_REGION=us-east-1")
	assertContains(t, env, "RULEKIT_S3_ENDPOINT=https://r2.example.com")
	assertContains(t, env, "RULEKIT_S3_ACCESS_KEY_ID=AKIA123")
	assertContains(t, env, "RULEKIT_S3_SECRET_ACCESS_KEY=secret")
	assertContains(t, env, "RULEKIT_AUTH=jwt")
	assertContains(t, env, "RULEKIT_JWT_SECRET=abc123")
	assertContains(t, env, "RULEKIT_ADMIN_EMAIL=admin@example.com")
	assertContains(t, env, "RULEKIT_SMTP_HOST=smtp.example.com")
	assertContains(t, env, "RULEKIT_SMTP_PORT=587")
	assertContains(t, env, "RULEKIT_SMTP_USERNAME=user")
	assertContains(t, env, "RULEKIT_SMTP_PASSWORD=pass")
	assertContains(t, env, "RULEKIT_SMTP_FROM=noreply@rulekit.dev")
	assertContains(t, env, "RULEKIT_SMTP_USE_TLS=false")

	assertNotContains(t, env, "RULEKIT_DATA_DIR")
}

func TestToEnv_SQLite_FS_APIKey(t *testing.T) {
	cfg := &StackConfig{
		Store:         "sqlite",
		DataDir:       "/data",
		BlobStore:     "fs",
		Auth:          "none",
		APIKey:        "myapikey",
		RegistryPort:  8080,
		DashboardPort: 3001,
	}
	env := cfg.ToEnv()

	assertContains(t, env, "RULEKIT_API_KEY=myapikey")
	assertNotContains(t, env, "RULEKIT_JWT_SECRET")
}

// --- LoadFromEnv roundtrip ---

func TestLoadFromEnv_Roundtrip(t *testing.T) {
	cfg := &StackConfig{
		Store:             "postgres",
		DatabaseURL:       "postgres://u:p@localhost:5432/db",
		BlobStore:         "s3",
		S3Bucket:          "bucket",
		S3Region:          "eu-west-1",
		S3Endpoint:        "",
		S3AccessKeyID:     "AKIA",
		S3SecretAccessKey: "sec",
		Auth:              "jwt",
		JWTSecret:         "jwt-secret",
		AdminEmail:        "a@b.com",
		SMTPEnabled:       true,
		SMTPHost:          "smtp.host",
		SMTPPort:          465,
		SMTPUsername:      "user",
		SMTPPassword:      "pass",
		SMTPFrom:          "from@host",
		SMTPUseTLS:        true,
		RegistryPort:      9090,
		DashboardPort:     4000,
	}

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := WriteEnv(path, cfg.ToEnv()); err != nil {
		t.Fatalf("WriteEnv: %v", err)
	}

	got, err := LoadFromEnv(path)
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}

	assertEqual(t, "Store", got.Store, cfg.Store)
	assertEqual(t, "DatabaseURL", got.DatabaseURL, cfg.DatabaseURL)
	assertEqual(t, "BlobStore", got.BlobStore, cfg.BlobStore)
	assertEqual(t, "S3Bucket", got.S3Bucket, cfg.S3Bucket)
	assertEqual(t, "S3Region", got.S3Region, cfg.S3Region)
	assertEqual(t, "S3AccessKeyID", got.S3AccessKeyID, cfg.S3AccessKeyID)
	assertEqual(t, "S3SecretAccessKey", got.S3SecretAccessKey, cfg.S3SecretAccessKey)
	assertEqual(t, "Auth", got.Auth, cfg.Auth)
	assertEqual(t, "JWTSecret", got.JWTSecret, cfg.JWTSecret)
	assertEqual(t, "AdminEmail", got.AdminEmail, cfg.AdminEmail)
	assertEqual(t, "SMTPHost", got.SMTPHost, cfg.SMTPHost)
	if got.SMTPPort != cfg.SMTPPort {
		t.Errorf("SMTPPort: got %d, want %d", got.SMTPPort, cfg.SMTPPort)
	}
	assertEqual(t, "SMTPUsername", got.SMTPUsername, cfg.SMTPUsername)
	assertEqual(t, "SMTPPassword", got.SMTPPassword, cfg.SMTPPassword)
	assertEqual(t, "SMTPFrom", got.SMTPFrom, cfg.SMTPFrom)
	if got.SMTPUseTLS != cfg.SMTPUseTLS {
		t.Errorf("SMTPUseTLS: got %v, want %v", got.SMTPUseTLS, cfg.SMTPUseTLS)
	}
	if got.RegistryPort != cfg.RegistryPort {
		t.Errorf("RegistryPort: got %d, want %d", got.RegistryPort, cfg.RegistryPort)
	}
	if got.DashboardPort != cfg.DashboardPort {
		t.Errorf("DashboardPort: got %d, want %d", got.DashboardPort, cfg.DashboardPort)
	}
}

func TestLoadFromEnv_MissingOptionalFields(t *testing.T) {
	minimal := "RULEKIT_STORE=sqlite\nRULEKIT_DATA_DIR=/data\n"

	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(minimal), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := LoadFromEnv(path)
	if err != nil {
		t.Fatalf("LoadFromEnv: %v", err)
	}

	if got.Store != "sqlite" {
		t.Errorf("Store: got %q, want %q", got.Store, "sqlite")
	}
	if got.BlobStore != "fs" {
		t.Errorf("BlobStore default: got %q, want %q", got.BlobStore, "fs")
	}
	if got.Auth != "none" {
		t.Errorf("Auth default: got %q, want %q", got.Auth, "none")
	}
}

// --- Summary masking ---

func TestSummary_MasksSecret(t *testing.T) {
	cfg := &StackConfig{
		Store:         "postgres",
		DatabaseURL:   "postgres://rulekit:supersecret@localhost:5432/rulekit",
		BlobStore:     "fs",
		Auth:          "jwt",
		JWTSecret:     "topsecretjwt",
		AdminEmail:    "admin@example.com",
		RegistryPort:  8080,
		DashboardPort: 3001,
	}
	summary := cfg.Summary()

	if strings.Contains(summary, "supersecret") {
		t.Error("Summary should mask password in postgres URL")
	}
	if strings.Contains(summary, "topsecretjwt") {
		t.Error("Summary should not contain raw JWT secret")
	}
	if !strings.Contains(summary, "***") {
		t.Error("Summary should contain *** for masked password")
	}
}

func TestSummary_APIKeyMasked(t *testing.T) {
	cfg := &StackConfig{
		Store:         "sqlite",
		BlobStore:     "fs",
		Auth:          "none",
		APIKey:        "myrawkey",
		RegistryPort:  8080,
		DashboardPort: 3001,
	}
	summary := cfg.Summary()

	if strings.Contains(summary, "myrawkey") {
		t.Error("Summary should mask API key")
	}
}

// --- WriteEnv permissions ---

func TestWriteEnv_Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	if err := WriteEnv(path, "KEY=val\n"); err != nil {
		t.Fatalf("WriteEnv: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions: got %o, want 0600", perm)
	}
}

// helpers

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected output to contain %q", sub)
	}
}

func assertNotContains(t *testing.T, s, sub string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Errorf("expected output NOT to contain %q", sub)
	}
}

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}
