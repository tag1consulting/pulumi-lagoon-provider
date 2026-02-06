package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the GraphQL API client for Lagoon.
type Client struct {
	apiURL    string
	token     string
	verifySSL bool
	client    *http.Client

	// Retry configuration
	maxRetries int
	baseDelay  time.Duration

	// Token refresh
	tokenFunc  func() (string, error) // optional: generates a fresh token
	tokenMu    sync.RWMutex
	tokenExpAt time.Time

	// API version detection
	apiVersion    string // "new" (v2.30.0+) or "legacy"
	apiVersionMu  sync.RWMutex
	apiVersionSet bool
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithInsecureSSL disables SSL certificate verification.
func WithInsecureSSL() ClientOption {
	return func(c *Client) {
		c.verifySSL = false
	}
}

// WithMaxRetries sets the number of retries for transient errors.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithTokenFunc sets a function for automatic token refresh.
// The function should return a fresh JWT token.
func WithTokenFunc(f func() (string, error)) ClientOption {
	return func(c *Client) {
		c.tokenFunc = f
	}
}

// NewClient creates a new Lagoon GraphQL API client.
func NewClient(apiURL, token string, opts ...ClientOption) *Client {
	c := &Client{
		apiURL:     apiURL,
		token:      token,
		verifySSL:  true,
		maxRetries: 3,
		baseDelay:  1 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := &http.Transport{}
	if !c.verifySSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-configured
	}

	c.client = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return c
}

// graphQLRequest represents a GraphQL request payload.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLResponse represents a GraphQL response.
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors"`
}

// Execute runs a GraphQL query or mutation and returns the raw data.
func (c *Client) Execute(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	// Refresh token if needed
	if err := c.refreshTokenIfNeeded(); err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		data, err := c.executeOnce(ctx, query, variables)
		if err == nil {
			return data, nil
		}

		// Only retry on connection/transport errors, not API errors
		var connErr *LagoonConnectionError
		if isConnectionError(err, &connErr) {
			lastErr = err
			continue
		}
		return nil, err
	}

	return nil, lastErr
}

func isConnectionError(err error, target **LagoonConnectionError) bool {
	var connErr *LagoonConnectionError
	if ok := asError(err, &connErr); ok {
		if target != nil {
			*target = connErr
		}
		return true
	}
	return false
}

// asError is a typed wrapper around errors.As for LagoonConnectionError.
func asError(err error, target **LagoonConnectionError) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*LagoonConnectionError)
	if ok {
		*target = e
		return true
	}
	return false
}

func (c *Client) executeOnce(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	payload := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &LagoonConnectionError{Message: "failed to marshal request", Cause: err}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, &LagoonConnectionError{Message: "failed to create request", Cause: err}
	}

	c.tokenMu.RLock()
	token := c.token
	c.tokenMu.RUnlock()

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &LagoonConnectionError{Message: fmt.Sprintf("HTTP request failed: %s", err), Cause: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &LagoonConnectionError{Message: "failed to read response body", Cause: err}
	}

	if resp.StatusCode >= 400 {
		return nil, &LagoonConnectionError{
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, &LagoonAPIError{Message: fmt.Sprintf("invalid JSON response: %s", err)}
	}

	if len(gqlResp.Errors) > 0 {
		messages := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			messages[i] = e.Message
		}
		return nil, &LagoonAPIError{
			Message: strings.Join(messages, "; "),
			Errors:  gqlResp.Errors,
		}
	}

	return gqlResp.Data, nil
}

func (c *Client) refreshTokenIfNeeded() error {
	if c.tokenFunc == nil {
		return nil
	}

	c.tokenMu.RLock()
	needsRefresh := !c.tokenExpAt.IsZero() && time.Now().After(c.tokenExpAt.Add(-5*time.Minute))
	c.tokenMu.RUnlock()

	if !needsRefresh {
		return nil
	}

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check after acquiring write lock
	if !c.tokenExpAt.IsZero() && time.Now().Before(c.tokenExpAt.Add(-5*time.Minute)) {
		return nil
	}

	newToken, err := c.tokenFunc()
	if err != nil {
		return &LagoonConnectionError{Message: "failed to refresh token", Cause: err}
	}
	c.token = newToken
	c.tokenExpAt = time.Now().Add(1 * time.Hour)

	return nil
}

// SetToken updates the client's auth token.
func (c *Client) SetToken(token string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = token
	c.tokenExpAt = time.Now().Add(1 * time.Hour)
}

// DetectAPIVersion probes the Lagoon API to determine if it's v2.30.0+ (new)
// or an older version (legacy). The result is cached for the client lifetime.
func (c *Client) DetectAPIVersion(ctx context.Context) string {
	c.apiVersionMu.RLock()
	if c.apiVersionSet {
		v := c.apiVersion
		c.apiVersionMu.RUnlock()
		return v
	}
	c.apiVersionMu.RUnlock()

	c.apiVersionMu.Lock()
	defer c.apiVersionMu.Unlock()

	// Double-check
	if c.apiVersionSet {
		return c.apiVersion
	}

	// Probe: try a new-API-only query with a minimal payload
	probeQuery := `query { __type(name: "EnvVariableByNameInput") { name } }`
	data, err := c.Execute(ctx, probeQuery, nil)
	if err != nil {
		c.apiVersion = "legacy"
		c.apiVersionSet = true
		return c.apiVersion
	}

	var result struct {
		Type *struct {
			Name string `json:"name"`
		} `json:"__type"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Type == nil {
		c.apiVersion = "legacy"
	} else {
		c.apiVersion = "new"
	}

	c.apiVersionSet = true
	return c.apiVersion
}

// IsNewAPI returns true if the API is v2.30.0+.
func (c *Client) IsNewAPI(ctx context.Context) bool {
	return c.DetectAPIVersion(ctx) == "new"
}

// extractField extracts a top-level field from a raw JSON data response.
func extractField(data json.RawMessage, field string) (json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse response data: %w", err)
	}
	v, ok := m[field]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found in response", field)
	}
	return v, nil
}

// unmarshalField extracts a field from the GraphQL response and unmarshals it.
func unmarshalField[T any](data json.RawMessage, field string) (T, error) {
	var zero T
	raw, err := extractField(data, field)
	if err != nil {
		return zero, err
	}
	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return zero, fmt.Errorf("failed to unmarshal '%s': %w", field, err)
	}
	return result, nil
}
