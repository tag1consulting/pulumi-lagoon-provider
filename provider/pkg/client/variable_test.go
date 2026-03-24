package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAddVariable_Legacy(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// First call: DetectAPIVersion probe
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil
		}
		// Legacy add
		if strings.Contains(query, "addEnvVariable") {
			return map[string]any{
				"addEnvVariable": map[string]any{
					"id":    100,
					"name":  "MY_VAR",
					"value": "my-value",
					"scope": "BUILD",
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	v, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", nil)
	if err != nil {
		t.Fatalf("AddVariable failed: %v", err)
	}
	if v.ID != 100 {
		t.Errorf("expected ID=100, got %d", v.ID)
	}
	if v.Name != "MY_VAR" {
		t.Errorf("expected name=MY_VAR, got %s", v.Name)
	}
}

func TestAddVariable_New(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// API version probe
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		// Need project lookup for new API
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// New API add
		if strings.Contains(query, "addOrUpdateEnvVariableByName") {
			return map[string]any{
				"addOrUpdateEnvVariableByName": map[string]any{
					"id":    200,
					"name":  "MY_VAR",
					"value": "my-value",
					"scope": "BUILD",
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	v, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", nil)
	if err != nil {
		t.Fatalf("AddVariable failed: %v", err)
	}
	if v.ID != 200 {
		t.Errorf("expected ID=200, got %d", v.ID)
	}
}

func TestAddVariable_WithEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil // legacy
		}
		if strings.Contains(query, "addEnvVariable") {
			return map[string]any{
				"addEnvVariable": map[string]any{
					"id":    101,
					"name":  "ENV_VAR",
					"value": "env-val",
					"scope": "RUNTIME",
				},
			}, nil
		}
		return nil, errors.New("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	v, err := c.AddVariable(context.Background(), "ENV_VAR", "env-val", 1, "runtime", &envID)
	if err != nil {
		t.Fatalf("AddVariable with environment failed: %v", err)
	}
	if v.ID != 101 {
		t.Errorf("expected ID=101, got %d", v.ID)
	}
}

func TestGetVariable_Legacy(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil
		}
		if strings.Contains(query, "envVariablesByProjectEnvironment") {
			return map[string]any{
				"envVariablesByProjectEnvironment": []map[string]any{
					{"id": 10, "name": "OTHER_VAR", "value": "other", "scope": "BUILD"},
					{"id": 20, "name": "MY_VAR", "value": "found", "scope": "RUNTIME"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	v, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err != nil {
		t.Fatalf("GetVariable failed: %v", err)
	}
	if v.Value != "found" {
		t.Errorf("expected value=found, got %s", v.Value)
	}
}

func TestGetVariable_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil
		}
		if strings.Contains(query, "envVariablesByProjectEnvironment") {
			return map[string]any{
				"envVariablesByProjectEnvironment": []map[string]any{
					{"id": 10, "name": "OTHER_VAR", "value": "other", "scope": "BUILD"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetVariable(context.Background(), "NONEXISTENT", 1, nil)
	if err == nil {
		t.Fatal("expected error for not found variable")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDeleteVariable_Legacy(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil
		}
		if strings.Contains(query, "deleteEnvVariable") {
			return map[string]any{"deleteEnvVariable": "success"}, nil
		}
		return nil, errors.New("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, nil)
	if err != nil {
		t.Fatalf("DeleteVariable failed: %v", err)
	}
}

func TestIsAPIError(t *testing.T) {
	apiErr := &LagoonAPIError{Message: "test"}
	var target *LagoonAPIError
	if !isAPIError(apiErr, &target) {
		t.Error("expected isAPIError to return true")
	}
	if target == nil || target.Message != "test" {
		t.Error("expected target to be set")
	}
}

func TestIsAPIError_NotAPIError(t *testing.T) {
	err := errors.New("not an API error")
	var target *LagoonAPIError
	if isAPIError(err, &target) {
		t.Error("expected isAPIError to return false for non-API error")
	}
}

func TestIsFieldNotFoundError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"Cannot query field 'addOrUpdateEnvVariableByName'", true},
		{"Unknown argument 'environment'", true},
		{"Something else entirely", false},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			err := &LagoonAPIError{Message: tt.msg}
			got := isFieldNotFoundError(err)
			if got != tt.want {
				t.Errorf("isFieldNotFoundError(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}
