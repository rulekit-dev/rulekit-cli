package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response is the JSON body returned by GET /healthz.
type Response struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Poll polls GET <baseURL>/healthz every second until the registry responds
// with status "ok" or the context/timeout is exceeded.
func Poll(ctx context.Context, baseURL string, timeout time.Duration) (*Response, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	client := &http.Client{Timeout: 3 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := check(client, baseURL+"/healthz")
			if err == nil {
				return resp, nil
			}
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("health check timed out after %v: %w", timeout, err)
			}
		}
	}
}

// Check performs a single health check against baseURL/healthz.
// Returns an error if the service is unhealthy or unreachable.
func Check(baseURL string) (*Response, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	return check(client, baseURL+"/healthz")
}

func check(client *http.Client, url string) (*Response, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var hr Response
	if err := json.NewDecoder(resp.Body).Decode(&hr); err != nil {
		return nil, fmt.Errorf("decode healthz: %w", err)
	}

	if hr.Status != "ok" {
		return nil, fmt.Errorf("registry status: %q", hr.Status)
	}

	return &hr, nil
}

// Reachable returns true if a plain HTTP GET to url returns any 2xx response.
func Reachable(url string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
