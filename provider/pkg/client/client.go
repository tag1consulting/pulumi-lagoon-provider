package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
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
	maxDelay   time.Duration

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
		if n < 0 {
			n = 0
		}
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
		maxDelay:   30 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if !c.verifySSL {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,    //nolint:gosec // user-configured
			MinVersion:         tls.VersionTLS12,
		}
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
	var nextDelay time.Duration
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(nextDelay):
			}
		}

		data, err := c.executeOnce(ctx, query, variables)
		if err == nil {
			return data, nil
		}

		// Retry on rate limit errors, honoring the Retry-After header.
		var rlErr *LagoonRateLimitError
		if errors.As(err, &rlErr) {
			lastErr = err
			nextDelay = c.retryDelay(attempt+1, rlErr.RetryAfter)
			continue
		}
		// Retry on connection/transport errors with jittered exponential backoff.
		var connErr *LagoonConnectionError
		if isConnectionError(err, &connErr) {
			lastErr = err
			nextDelay = c.retryDelay(attempt+1, 0)
			continue
		}
		return nil, err
	}

	return nil, lastErr
}

// retryDelay returns the wait duration for a retry attempt.
// minDelay overrides the backoff when set (e.g. from a Retry-After header).
func (c *Client) retryDelay(attempt int, minDelay time.Duration) time.Duration {
	exp := c.baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
	if exp > c.maxDelay {
		exp = c.maxDelay
	}
	// Add up to 25% jitter so concurrent retries don't all fire at once.
	// Guard against exp==0 (e.g. baseDelay unset) to avoid rand.Int63n(0) panic.
	var jitter time.Duration
	if exp > 0 {
		jitter = time.Duration(rand.Int63n(int64(exp) / 4)) //nolint:gosec // non-crypto jitter
	}
	delay := exp + jitter
	if minDelay > delay {
		delay = minDelay
	}
	return delay
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
	return errors.As(err, target)
}

func (c *Client) executeOnce(ctx context.Context, query string, variables map[string]any) (json.RawMessage, error) {
	payload := graphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &LagoonConnectionError{Message: "failed to read response body", Cause: err}
	}

	if resp.StatusCode == 429 {
		var retryAfter time.Duration
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
				retryAfter = time.Duration(secs) * time.Second
			}
			// Non-integer Retry-After values (e.g. HTTP-date, fractional seconds) are
			// silently ignored; retryAfter stays 0 and backoff uses the exponential default.
		}
		return nil, &LagoonRateLimitError{RetryAfter: retryAfter}
	}
	if resp.StatusCode >= 500 {
		return nil, &LagoonConnectionError{
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}
	}
	if resp.StatusCode >= 400 {
		return nil, &LagoonAPIError{
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
	// Zero tokenExpAt means the token hasn't been refreshed yet — treat as needing refresh
	needsRefresh := c.tokenExpAt.IsZero() || time.Now().After(c.tokenExpAt.Add(-5*time.Minute))
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
		// Don't cache — transient errors shouldn't poison the version detection
		return "legacy"
	}

	var result struct {
		Type *struct {
			Name string `json:"name"`
		} `json:"__type"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		// Unmarshal failed — don't cache, may be transient
		return "legacy"
	}
	if result.Type == nil {
		// Valid response but type not found — this is a real legacy API
		c.apiVersion = "legacy"
		c.apiVersionSet = true
		return "legacy"
	}

	c.apiVersion = "new"
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
