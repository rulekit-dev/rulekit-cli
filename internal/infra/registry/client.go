package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	httpClient *http.Client
}

type VersionMeta struct {
	Workspace  string    `json:"workspace"`
	RulesetKey string    `json:"ruleset_key"`
	Version    int       `json:"version"`
	Checksum   string    `json:"checksum"`
	CreatedAt  time.Time `json:"created_at"`
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) newRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return req, nil
}

func (c *Client) DownloadBundle(ctx context.Context, key, version, workspace string) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/rulesets/%s/versions/%s/bundle?workspace=%s", c.BaseURL, key, version, workspace)

	req, err := c.newRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	return body, nil
}

func (c *Client) GetLatestVersion(ctx context.Context, key, workspace string) (*VersionMeta, error) {
	url := fmt.Sprintf("%s/v1/rulesets/%s/versions/latest?workspace=%s", c.BaseURL, key, workspace)

	req, err := c.newRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("registry unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	var meta VersionMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &meta, nil
}

func parseAPIError(statusCode int, body []byte) error {
	var apiErr apiError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("registry error %d: %s", statusCode, apiErr.Error.Message)
	}
	return fmt.Errorf("registry error %d", statusCode)
}
