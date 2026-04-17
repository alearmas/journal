// Package ppi provides a client for the PPI (Prima Portafolio de Inversiones) REST API.
// Docs: https://itatppi.github.io/ppi-official-api-docs/api/documentacionRest/
package ppi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	sandboxBaseURL    = "https://clientapi_sandbox.portfoliopersonal.com"
	productionBaseURL = "https://clientapi.portfoliopersonal.com"
)

// Config holds the PPI API credentials.
type Config struct {
	Sandbox          bool   // true → sandbox, false → production
	AuthorizedClient string // AUTHORIZED_CLIENT env var
	ClientKey        string // CLIENT_KEY env var
	APIKey           string // PPI_PUBLIC_KEY env var
	APISecret        string // PPI_PRIVATE_KEY env var (optional)
}

type tokenData struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

// Client is a thread-safe PPI API HTTP client.
type Client struct {
	cfg     Config
	baseURL string
	http    *http.Client
	mu      sync.Mutex
	token   tokenData
}

// NewClient creates a new PPI API client.
func NewClient(cfg Config) *Client {
	baseURL := productionBaseURL
	if cfg.Sandbox {
		baseURL = sandboxBaseURL
	}
	return &Client{
		cfg:     cfg,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// loginResponse is the response from POST /api/1.0/Account/LoginApi.
type loginResponse struct {
	AccessToken    string    `json:"accessToken"`
	RefreshToken   string    `json:"refreshToken"`
	ExpirationDate time.Time `json:"expirationDate"`
	Expires        int       `json:"expires"` // seconds until expiry
	TokenType      string    `json:"tokenType"`
}

// Login authenticates against PPI and stores the bearer token.
func (c *Client) Login(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/1.0/Account/LoginApi", nil)
	if err != nil {
		return fmt.Errorf("ppi login: build request: %w", err)
	}

	req.Header.Set("AuthorizedClient", c.cfg.AuthorizedClient)
	req.Header.Set("ClientKey", c.cfg.ClientKey)
	req.Header.Set("ApiKey", c.cfg.APIKey)
	req.Header.Set("ApiSecret", c.cfg.APISecret)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("ppi login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ppi login: status %d – %s", resp.StatusCode, string(body))
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return fmt.Errorf("ppi login: decode response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(lr.Expires) * time.Second)
	if !lr.ExpirationDate.IsZero() {
		expiresAt = lr.ExpirationDate
	}

	c.mu.Lock()
	c.token = tokenData{
		accessToken:  lr.AccessToken,
		refreshToken: lr.RefreshToken,
		expiresAt:    expiresAt,
	}
	c.mu.Unlock()

	return nil
}

// ensureToken logs in or refreshes the token if necessary.
func (c *Client) ensureToken(ctx context.Context) error {
	c.mu.Lock()
	needsLogin := c.token.accessToken == "" ||
		time.Now().After(c.token.expiresAt.Add(-30*time.Second))
	c.mu.Unlock()

	if needsLogin {
		return c.Login(ctx)
	}
	return nil
}

// doGet performs an authenticated GET request.
func (c *Client) doGet(ctx context.Context, path string, params map[string]string) (*http.Response, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("ppi get %s: build request: %w", path, err)
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	c.mu.Lock()
	token := c.token.accessToken
	c.mu.Unlock()

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("AuthorizedClient", c.cfg.AuthorizedClient)
	req.Header.Set("ClientKey", c.cfg.ClientKey)
	req.Header.Set("Content-Type", "application/json")

	return c.http.Do(req)
}
