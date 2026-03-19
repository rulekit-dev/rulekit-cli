package registry

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeValidBundle(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	manifest := map[string]any{
		"namespace":   "default",
		"ruleset_key": "payout-routing",
		"version":     4,
		"checksum":    "sha256:a3f1c8deadbeef",
		"created_at":  "2025-01-01T00:00:00Z",
	}
	manifestBytes, _ := json.Marshal(manifest)

	addEntry := func(name string, data []byte) {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		f.Write(data)
	}

	addEntry("manifest.json", manifestBytes)
	addEntry("dsl.json", []byte(`{"dsl_version":"v1","rules":[]}`))

	if err := w.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	return buf.Bytes()
}

func TestDownloadBundle_HappyPath(t *testing.T) {
	bundle := makeValidBundle(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/rulesets/payout-routing/versions/latest/bundle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("namespace") != "default" {
			t.Errorf("unexpected namespace: %s", r.URL.Query().Get("namespace"))
		}
		w.Header().Set("Content-Type", "application/zip")
		w.WriteHeader(http.StatusOK)
		w.Write(bundle)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	data, err := client.DownloadBundle(context.Background(), "payout-routing", "latest", "default")
	if err != nil {
		t.Fatalf("DownloadBundle: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty bundle data")
	}
}

func TestDownloadBundle_AuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", auth)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(makeValidBundle(t))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")
	_, err := client.DownloadBundle(context.Background(), "payout-routing", "latest", "default")
	if err != nil {
		t.Fatalf("DownloadBundle: %v", err)
	}
}

func TestDownloadBundle_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "ruleset not found",
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.DownloadBundle(context.Background(), "missing-ruleset", "latest", "default")
	if err == nil {
		t.Fatal("expected error on 404, got nil")
	}
}

func TestDownloadBundle_Unreachable(t *testing.T) {
	client := NewClient("http://127.0.0.1:1", "")
	_, err := client.DownloadBundle(context.Background(), "payout-routing", "latest", "default")
	if err == nil {
		t.Fatal("expected error when registry is unreachable, got nil")
	}
}

func TestGetLatestVersion_HappyPath(t *testing.T) {
	meta := VersionMeta{
		Namespace:  "default",
		RulesetKey: "payout-routing",
		Version:    7,
		Checksum:   "sha256:cafebabe",
		CreatedAt:  time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/rulesets/payout-routing/versions/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	got, err := client.GetLatestVersion(context.Background(), "payout-routing", "default")
	if err != nil {
		t.Fatalf("GetLatestVersion: %v", err)
	}

	if got.Version != meta.Version {
		t.Errorf("version: got %d, want %d", got.Version, meta.Version)
	}
	if got.Checksum != meta.Checksum {
		t.Errorf("checksum: got %q, want %q", got.Checksum, meta.Checksum)
	}
}

func TestGetLatestVersion_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "INTERNAL",
				"message": "internal server error",
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.GetLatestVersion(context.Background(), "payout-routing", "default")
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}
