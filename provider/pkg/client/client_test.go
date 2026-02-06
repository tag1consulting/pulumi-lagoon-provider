package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://api.test/graphql", "test-token")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.apiURL != "https://api.test/graphql" {
		t.Errorf("expected api URL to be set")
	}
	if c.token != "test-token" {
		t.Errorf("expected token to be set")
	}
	if c.maxRetries != 3 {
		t.Errorf("expected default maxRetries=3, got %d", c.maxRetries)
	}
	if c.verifySSL != true {
		t.Error("expected verifySSL to default to true")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	tokenFunc := func() (string, error) { return "new-token", nil }
	c := NewClient("https://api.test/graphql", "test-token",
		WithInsecureSSL(),
		WithMaxRetries(5),
		WithTokenFunc(tokenFunc),
	)
	if c.verifySSL {
		t.Error("expected verifySSL to be false with WithInsecureSSL")
	}
	if c.maxRetries != 5 {
		t.Errorf("expected maxRetries=5, got %d", c.maxRetries)
	}
	if c.tokenFunc == nil {
		t.Error("expected tokenFunc to be set")
	}
}

func TestExecute_Success(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"projectByName": map[string]any{
				"id":   1,
				"name": "test-project",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	data, err := c.Execute(context.Background(), "query { projectByName(name: \"test\") { id name } }", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil data")
	}
}

func TestExecute_GraphQLError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return nil, fmt.Errorf("Project not found")
	})
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	_, err := c.Execute(context.Background(), "query { projectByName(name: \"x\") { id } }", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrAPI) {
		t.Errorf("expected LagoonAPIError, got %T: %v", err, err)
	}
}

func TestExecute_HTTPError(t *testing.T) {
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})
	defer server.Close()

	c := NewClient(server.URL, "test-token", WithMaxRetries(0))
	_, err := c.Execute(context.Background(), "query { test }", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrConnection) {
		t.Errorf("expected LagoonConnectionError for HTTP 500, got %T: %v", err, err)
	}
}

func TestExecute_RetryOnConnectionError(t *testing.T) {
	var attempts int32
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			// Simulate server error (triggers retry since it's a connection-level issue)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		// Success on 3rd attempt
		resp := map[string]any{
			"data": map[string]any{
				"test": "ok",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	c := NewClient(server.URL, "test-token", WithMaxRetries(3))
	c.baseDelay = 1 * time.Millisecond // speed up retries for test
	data, err := c.Execute(context.Background(), "query { test }", nil)
	if err != nil {
		t.Fatalf("Execute failed after retries: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil data after retries")
	}
}

func TestExecute_NoRetryOnAPIError(t *testing.T) {
	var attempts int32
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		atomic.AddInt32(&attempts, 1)
		return nil, fmt.Errorf("Validation error: name is required")
	})
	defer server.Close()

	c := NewClient(server.URL, "test-token", WithMaxRetries(3))
	c.baseDelay = 1 * time.Millisecond
	_, err := c.Execute(context.Background(), "mutation { test }", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt (no retry on API error), got %d", atomic.LoadInt32(&attempts))
	}
}

func TestExecute_ContextCancellation(t *testing.T) {
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	})
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	c := NewClient(server.URL, "test-token", WithMaxRetries(0))
	_, err := c.Execute(ctx, "query { test }", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestExecute_BearerToken(t *testing.T) {
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-token" {
			t.Errorf("expected 'Bearer my-token', got '%s'", auth)
		}
		resp := map[string]any{"data": map[string]any{"ok": true}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	c := NewClient(server.URL, "my-token")
	_, err := c.Execute(context.Background(), "query { ok }", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestExecute_SendsVariables(t *testing.T) {
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Query == "" {
			t.Error("expected query to be set")
		}
		input, ok := req.Variables["input"]
		if !ok {
			t.Error("expected 'input' variable")
		}
		inputMap, ok := input.(map[string]any)
		if !ok {
			t.Errorf("expected input to be map, got %T", input)
		}
		if inputMap["name"] != "test" {
			t.Errorf("expected name=test, got %v", inputMap["name"])
		}

		resp := map[string]any{"data": map[string]any{"result": "ok"}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.Execute(context.Background(), "mutation { test }", map[string]any{
		"input": map[string]any{"name": "test"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestSetToken(t *testing.T) {
	c := NewClient("https://api.test/graphql", "old-token")
	c.SetToken("new-token")
	if c.token != "new-token" {
		t.Errorf("expected token to be 'new-token', got %s", c.token)
	}
}

func TestDetectAPIVersion_New(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{
				"__type": map[string]any{
					"name": "EnvVariableByNameInput",
				},
			}, nil
		}
		return nil, fmt.Errorf("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	version := c.DetectAPIVersion(context.Background())
	if version != "new" {
		t.Errorf("expected 'new', got %s", version)
	}

	// Should cache the result
	version2 := c.DetectAPIVersion(context.Background())
	if version2 != "new" {
		t.Errorf("expected cached 'new', got %s", version2)
	}
}

func TestDetectAPIVersion_Legacy(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{
				"__type": nil,
			}, nil
		}
		return nil, fmt.Errorf("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	version := c.DetectAPIVersion(context.Background())
	if version != "legacy" {
		t.Errorf("expected 'legacy', got %s", version)
	}
}

func TestDetectAPIVersion_Error(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return nil, fmt.Errorf("some error")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	version := c.DetectAPIVersion(context.Background())
	if version != "legacy" {
		t.Errorf("expected 'legacy' on error, got %s", version)
	}
}

func TestIsNewAPI(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	if !c.IsNewAPI(context.Background()) {
		t.Error("expected IsNewAPI to return true")
	}
}

func TestExtractField(t *testing.T) {
	data := json.RawMessage(`{"project":{"id":1,"name":"test"}}`)
	raw, err := extractField(data, "project")
	if err != nil {
		t.Fatalf("extractField failed: %v", err)
	}
	var result struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if result.ID != 1 || result.Name != "test" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestExtractField_NotFound(t *testing.T) {
	data := json.RawMessage(`{"project":{"id":1}}`)
	_, err := extractField(data, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing field")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestExtractField_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)
	_, err := extractField(data, "field")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestUnmarshalField(t *testing.T) {
	data := json.RawMessage(`{"project":{"id":1,"name":"test"}}`)
	type P struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	result, err := unmarshalField[P](data, "project")
	if err != nil {
		t.Fatalf("unmarshalField failed: %v", err)
	}
	if result.ID != 1 || result.Name != "test" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestUnmarshalField_Array(t *testing.T) {
	data := json.RawMessage(`{"items":[{"id":1},{"id":2}]}`)
	type Item struct {
		ID int `json:"id"`
	}
	result, err := unmarshalField[[]Item](data, "items")
	if err != nil {
		t.Fatalf("unmarshalField failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestUnmarshalField_Null(t *testing.T) {
	data := json.RawMessage(`{"project":null}`)
	type P struct {
		ID int `json:"id"`
	}
	result, err := unmarshalField[*P](data, "project")
	if err != nil {
		t.Fatalf("unmarshalField failed: %v", err)
	}
	if result != nil {
		t.Error("expected nil for null field")
	}
}

func TestTokenRefresh(t *testing.T) {
	var callCount int32
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"ok": true}, nil
	})
	defer server.Close()

	tokenFunc := func() (string, error) {
		atomic.AddInt32(&callCount, 1)
		return "refreshed-token", nil
	}

	c := NewClient(server.URL, "old-token", WithTokenFunc(tokenFunc))
	// Force token expiry
	c.tokenExpAt = time.Now().Add(-1 * time.Minute)

	_, err := c.Execute(context.Background(), "query { ok }", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected token refresh to be called once, got %d", atomic.LoadInt32(&callCount))
	}
	if c.token != "refreshed-token" {
		t.Errorf("expected token to be 'refreshed-token', got %s", c.token)
	}
}

func TestTokenRefresh_NotNeeded(t *testing.T) {
	var callCount int32
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"ok": true}, nil
	})
	defer server.Close()

	tokenFunc := func() (string, error) {
		atomic.AddInt32(&callCount, 1)
		return "refreshed", nil
	}

	c := NewClient(server.URL, "current-token", WithTokenFunc(tokenFunc))
	// Token is not expired, tokenExpAt is zero (first use never triggers refresh)

	_, err := c.Execute(context.Background(), "query { ok }", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 0 {
		t.Error("expected no token refresh when token is not expired")
	}
}

func TestTokenRefresh_Error(t *testing.T) {
	tokenFunc := func() (string, error) {
		return "", fmt.Errorf("refresh failed")
	}

	c := NewClient("https://api.test/graphql", "old-token", WithTokenFunc(tokenFunc))
	c.tokenExpAt = time.Now().Add(-1 * time.Minute)

	_, err := c.Execute(context.Background(), "query { ok }", nil)
	if err == nil {
		t.Fatal("expected error from failed token refresh")
	}
	if !errors.Is(err, ErrConnection) {
		t.Errorf("expected LagoonConnectionError, got %T", err)
	}
}

func TestExecute_InvalidJSON(t *testing.T) {
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	})
	defer server.Close()

	c := NewClient(server.URL, "token", WithMaxRetries(0))
	_, err := c.Execute(context.Background(), "query { test }", nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestExecute_MultipleErrors(t *testing.T) {
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"errors": []map[string]any{
				{"message": "error 1"},
				{"message": "error 2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	c := NewClient(server.URL, "token", WithMaxRetries(0))
	_, err := c.Execute(context.Background(), "query { test }", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "error 1") || !strings.Contains(err.Error(), "error 2") {
		t.Errorf("expected both errors in message, got: %v", err)
	}
}
