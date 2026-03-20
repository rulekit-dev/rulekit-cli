package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPoll_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{Status: "ok", Version: "v0.1.0"})
	}))
	defer srv.Close()

	resp, err := Poll(context.Background(), srv.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status: got %q, want %q", resp.Status, "ok")
	}
	if resp.Version != "v0.1.0" {
		t.Errorf("version: got %q, want %q", resp.Version, "v0.1.0")
	}
}

func TestPoll_Timeout(t *testing.T) {
	// Server that always returns 503.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, err := Poll(context.Background(), srv.URL, 1500*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestPoll_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := Poll(ctx, srv.URL, 5*time.Second)
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
}

func TestPoll_RetryOnError(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		json.NewEncoder(w).Encode(Response{Status: "ok", Version: "v1.0.0"})
	}))
	defer srv.Close()

	resp, err := Poll(context.Background(), srv.URL, 10*time.Second)
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status: got %q, want %q", resp.Status, "ok")
	}
	if attempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempts)
	}
}

func TestCheck_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{Status: "ok", Version: "v0.2.0"})
	}))
	defer srv.Close()

	resp, err := Check(srv.URL)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if resp.Version != "v0.2.0" {
		t.Errorf("version: got %q, want %q", resp.Version, "v0.2.0")
	}
}

func TestCheck_Unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{Status: "degraded"})
	}))
	defer srv.Close()

	_, err := Check(srv.URL)
	if err == nil {
		t.Fatal("expected error for unhealthy status, got nil")
	}
}

func TestCheck_Unreachable(t *testing.T) {
	_, err := Check("http://127.0.0.1:1")
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestReachable_True(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if !Reachable(srv.URL) {
		t.Error("expected Reachable to return true for running server")
	}
}

func TestReachable_False(t *testing.T) {
	if Reachable("http://127.0.0.1:1") {
		t.Error("expected Reachable to return false for unreachable server")
	}
}
