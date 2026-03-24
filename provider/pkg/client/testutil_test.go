package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockGraphQLServer creates a test server that handles GraphQL requests.
// The handler receives parsed query and variables, and should write the response.
func mockGraphQLServer(t *testing.T, handler func(query string, variables map[string]any) (any, error)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req graphQLRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to unmarshal request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		data, handlerErr := handler(req.Query, req.Variables)
		if handlerErr != nil {
			resp := map[string]any{
				"errors": []map[string]any{{"message": handlerErr.Error()}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := map[string]any{"data": data}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// mockGraphQLServerRaw creates a test server that returns raw bytes.
func mockGraphQLServerRaw(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(handler))
}

// intPtr returns a pointer to an int.
func intPtr(i int) *int { return &i }

// strPtr returns a pointer to a string.
func strPtr(s string) *string { return &s }
