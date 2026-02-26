package notionclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://www.notion.so"

type Options struct {
	BaseURL      string
	TokenV2      string
	NotionUserID string
	ActiveUserID string
	Cookie       string
	HTTPClient   *http.Client
}

type Client struct {
	baseURL      *url.URL
	httpClient   *http.Client
	tokenV2      string
	notionUserID string
	activeUserID string
	cookie       string
}

func New(opts Options) (*Client, error) {
	baseURL := strings.TrimSpace(opts.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		baseURL:      u,
		httpClient:   hc,
		tokenV2:      strings.TrimSpace(opts.TokenV2),
		notionUserID: strings.TrimSpace(opts.NotionUserID),
		activeUserID: strings.TrimSpace(opts.ActiveUserID),
		cookie:       strings.TrimSpace(opts.Cookie),
	}, nil
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload any) (map[string]any, error) {
	rel, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	u := c.baseURL.ResolveReference(rel)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", c.baseURL.String())
	req.Header.Set("Referer", c.baseURL.String()+"/")
	req.Header.Set("User-Agent", "Mozilla/5.0 notion-cli/0.1")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	if cookie := c.cookieHeader(); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	if c.activeUserID != "" {
		req.Header.Set("x-notion-active-user-header", c.activeUserID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("notion request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var out map[string]any
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode response json: %w", err)
	}

	return out, nil
}

func (c *Client) cookieHeader() string {
	if c.cookie != "" {
		return c.cookie
	}

	parts := make([]string, 0, 2)
	if c.tokenV2 != "" {
		parts = append(parts, "token_v2="+c.tokenV2)
	}
	if c.notionUserID != "" {
		parts = append(parts, "notion_user_id="+c.notionUserID)
	}
	return strings.Join(parts, "; ")
}
