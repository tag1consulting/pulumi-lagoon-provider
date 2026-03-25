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

func TestGetVariable_NewAPI(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// API version probe — return new API
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		// GetProjectByID — allProjects
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// New API query for variables by project/environment name
		if strings.Contains(query, "getEnvVariablesByProjectEnvironmentName") {
			return map[string]any{
				"getEnvVariablesByProjectEnvironmentName": []map[string]any{
					{"id": 10, "name": "OTHER_VAR", "value": "other", "scope": "BUILD"},
					{"id": 20, "name": "MY_VAR", "value": "found-new", "scope": "RUNTIME"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	v, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err != nil {
		t.Fatalf("GetVariable (new API) failed: %v", err)
	}
	if v.Value != "found-new" {
		t.Errorf("expected value=found-new, got %s", v.Value)
	}
	if v.ID != 20 {
		t.Errorf("expected ID=20, got %d", v.ID)
	}
}

func TestGetVariable_NewAPI_WithEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// environmentById lookup for getEnvironmentName
		if strings.Contains(query, "environmentById") {
			return map[string]any{
				"environmentById": map[string]any{"id": 5, "name": "main"},
			}, nil
		}
		if strings.Contains(query, "getEnvVariablesByProjectEnvironmentName") {
			return map[string]any{
				"getEnvVariablesByProjectEnvironmentName": []map[string]any{
					{"id": 30, "name": "ENV_VAR", "value": "env-val", "scope": "RUNTIME"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	v, err := c.GetVariable(context.Background(), "ENV_VAR", 1, &envID)
	if err != nil {
		t.Fatalf("GetVariable (new API with env) failed: %v", err)
	}
	if v.Value != "env-val" {
		t.Errorf("expected value=env-val, got %s", v.Value)
	}
}

func TestGetVariable_NewAPI_FallbackToLegacy(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// API version probe — new API
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		// GetProjectByID for new API path
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// New API query returns "Cannot query field" error (triggers fallback)
		if strings.Contains(query, "getEnvVariablesByProjectEnvironmentName") {
			return nil, errors.New("Cannot query field 'getEnvVariablesByProjectEnvironmentName'")
		}
		// Legacy fallback
		if strings.Contains(query, "envVariablesByProjectEnvironment") {
			return map[string]any{
				"envVariablesByProjectEnvironment": []map[string]any{
					{"id": 50, "name": "MY_VAR", "value": "legacy-value", "scope": "BUILD"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	v, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err != nil {
		t.Fatalf("GetVariable (fallback to legacy) failed: %v", err)
	}
	if v.Value != "legacy-value" {
		t.Errorf("expected value=legacy-value (from legacy fallback), got %s", v.Value)
	}
	if v.ID != 50 {
		t.Errorf("expected ID=50 (from legacy), got %d", v.ID)
	}
}

func TestDeleteVariable_NewAPI(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "deleteEnvVariableByName") {
			return map[string]any{"deleteEnvVariableByName": "success"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, nil)
	if err != nil {
		t.Fatalf("DeleteVariable (new API) failed: %v", err)
	}
}

func TestDeleteVariable_NewAPI_WithEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "environmentById") {
			return map[string]any{
				"environmentById": map[string]any{"id": 5, "name": "main"},
			}, nil
		}
		if strings.Contains(query, "deleteEnvVariableByName") {
			return map[string]any{"deleteEnvVariableByName": "success"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, &envID)
	if err != nil {
		t.Fatalf("DeleteVariable (new API with env) failed: %v", err)
	}
}

func TestGetVariable_NewAPI_EnvironmentNotFound(t *testing.T) {
	// Tests getEnvironmentName returning LagoonNotFoundError when env name is empty
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// environmentById returns an env with an empty name
		if strings.Contains(query, "environmentById") {
			return map[string]any{
				"environmentById": map[string]any{"id": 999, "name": ""},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 999
	_, err := c.GetVariable(context.Background(), "MY_VAR", 1, &envID)
	if err == nil {
		t.Fatal("expected error when environment name is empty")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for missing environment, got %T: %v", err, err)
	}
}

func TestAddVariable_NewAPI_WithEnvironment(t *testing.T) {
	// Tests addVariableNew with environmentID set (the env lookup path)
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "environmentById") {
			return map[string]any{
				"environmentById": map[string]any{"id": 5, "name": "main"},
			}, nil
		}
		if strings.Contains(query, "addOrUpdateEnvVariableByName") {
			return map[string]any{
				"addOrUpdateEnvVariableByName": map[string]any{
					"id":    300,
					"name":  "ENV_VAR",
					"value": "env-val",
					"scope": "RUNTIME",
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	v, err := c.AddVariable(context.Background(), "ENV_VAR", "env-val", 1, "runtime", &envID)
	if err != nil {
		t.Fatalf("AddVariable (new API with env) failed: %v", err)
	}
	if v.ID != 300 {
		t.Errorf("expected ID=300, got %d", v.ID)
	}
}

// --- Category 2: Malformed response tests for legacy paths ---

func TestAddVariableLegacy_MalformedResponse(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil // legacy
		}
		if strings.Contains(query, "addEnvVariable") {
			return map[string]any{"addEnvVariable": "not-an-object"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", nil)
	if err == nil {
		t.Fatal("expected error for malformed addEnvVariable response")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

func TestGetVariableLegacy_MalformedResponse(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil // legacy
		}
		if strings.Contains(query, "envVariablesByProjectEnvironment") {
			return map[string]any{"envVariablesByProjectEnvironment": "bad"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err == nil {
		t.Fatal("expected error for malformed envVariablesByProjectEnvironment response")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

// --- Category 3: Variable new-API edge cases ---

func TestAddVariable_NewAPI_GetProjectError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		// GetProjectByID (allProjects) returns an error
		if strings.Contains(query, "allProjects") {
			return nil, errors.New("permission denied for allProjects")
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", nil)
	if err == nil {
		t.Fatal("expected error when GetProjectByID fails in addVariableNew")
	}
	if !strings.Contains(err.Error(), "failed to look up project") {
		t.Errorf("expected 'failed to look up project' in error, got: %v", err)
	}
}

func TestAddVariable_NewAPI_GetEnvironmentNameError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// getEnvironmentName fails
		if strings.Contains(query, "environmentById") {
			return nil, errors.New("environment query failed")
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	_, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", &envID)
	if err == nil {
		t.Fatal("expected error when getEnvironmentName fails in addVariableNew")
	}
	if !strings.Contains(err.Error(), "failed to look up environment") {
		t.Errorf("expected 'failed to look up environment' in error, got: %v", err)
	}
}

func TestAddVariable_NewAPI_UnmarshalError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// Mutation succeeds but returns malformed data
		if strings.Contains(query, "addOrUpdateEnvVariableByName") {
			return map[string]any{"addOrUpdateEnvVariableByName": "not-an-object"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", nil)
	if err == nil {
		t.Fatal("expected error for malformed addOrUpdateEnvVariableByName response")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

func TestDeleteVariable_NewAPI_FallbackToLegacy(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// deleteEnvVariableByName returns "Cannot query field" -> triggers fallback
		if strings.Contains(query, "deleteEnvVariableByName") {
			return nil, errors.New("Cannot query field 'deleteEnvVariableByName'")
		}
		// Legacy delete succeeds
		if strings.Contains(query, "deleteEnvVariable") {
			return map[string]any{"deleteEnvVariable": "success"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, nil)
	if err != nil {
		t.Fatalf("expected DeleteVariable to succeed via legacy fallback, got: %v", err)
	}
}

func TestDeleteVariable_NewAPI_GetProjectError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		// GetProjectByID fails
		if strings.Contains(query, "allProjects") {
			return nil, errors.New("project lookup failed")
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, nil)
	if err == nil {
		t.Fatal("expected error when GetProjectByID fails in deleteVariableNew")
	}
}

func TestDeleteVariable_NewAPI_GetEnvironmentNameError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// getEnvironmentName returns empty name (LagoonNotFoundError)
		if strings.Contains(query, "environmentById") {
			return map[string]any{
				"environmentById": map[string]any{"id": 999, "name": ""},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 999
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, &envID)
	if err == nil {
		t.Fatal("expected error when getEnvironmentName returns empty name")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestAddVariable_NewAPI_FallbackToLegacy(t *testing.T) {
	// Tests the addVariableNew fallback to legacy on "Cannot query field" error
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// addOrUpdateEnvVariableByName returns "Cannot query field" -> fallback
		if strings.Contains(query, "addOrUpdateEnvVariableByName") {
			return nil, errors.New("Cannot query field 'addOrUpdateEnvVariableByName'")
		}
		// Legacy add succeeds
		if strings.Contains(query, "addEnvVariable") {
			return map[string]any{
				"addEnvVariable": map[string]any{
					"id":    500,
					"name":  "MY_VAR",
					"value": "legacy-val",
					"scope": "BUILD",
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	v, err := c.AddVariable(context.Background(), "MY_VAR", "legacy-val", 1, "build", nil)
	if err != nil {
		t.Fatalf("expected AddVariable to succeed via legacy fallback, got: %v", err)
	}
	if v.ID != 500 {
		t.Errorf("expected ID=500 from legacy fallback, got %d", v.ID)
	}
}

func TestAddVariable_NewAPI_NonFallbackError(t *testing.T) {
	// Tests that a non-fallback error from addOrUpdateEnvVariableByName is propagated
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "addOrUpdateEnvVariableByName") {
			return nil, errors.New("permission denied")
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.AddVariable(context.Background(), "MY_VAR", "my-value", 1, "build", nil)
	if err == nil {
		t.Fatal("expected error to propagate from addVariableNew")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected 'permission denied' error, got: %v", err)
	}
}

func TestGetVariable_NewAPI_NotFound(t *testing.T) {
	// Tests getVariableNew returning not found when name doesn't match
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "getEnvVariablesByProjectEnvironmentName") {
			return map[string]any{
				"getEnvVariablesByProjectEnvironmentName": []map[string]any{
					{"id": 10, "name": "OTHER_VAR", "value": "other", "scope": "BUILD"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetVariable(context.Background(), "NONEXISTENT", 1, nil)
	if err == nil {
		t.Fatal("expected error for not-found variable in new API")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestGetVariable_NewAPI_ExecuteError(t *testing.T) {
	// Tests getVariableNew when Execute returns a non-fallback error
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "getEnvVariablesByProjectEnvironmentName") {
			return nil, errors.New("internal server error")
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err == nil {
		t.Fatal("expected error from getVariableNew Execute failure")
	}
}

func TestGetVariable_NewAPI_UnmarshalError(t *testing.T) {
	// Tests getVariableNew returning unmarshal error on bad field data
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		if strings.Contains(query, "getEnvVariablesByProjectEnvironmentName") {
			return map[string]any{"getEnvVariablesByProjectEnvironmentName": "not-an-array"}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err == nil {
		t.Fatal("expected error for malformed getEnvVariablesByProjectEnvironmentName response")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}

func TestGetVariable_NewAPI_GetProjectError(t *testing.T) {
	// Tests getVariableNew when GetProjectByID fails
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return nil, errors.New("project lookup failed")
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetVariable(context.Background(), "MY_VAR", 1, nil)
	if err == nil {
		t.Fatal("expected error when GetProjectByID fails in getVariableNew")
	}
}

func TestGetVariable_Legacy_WithEnvironment(t *testing.T) {
	// Tests getVariableLegacy with environment ID set
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil
		}
		if strings.Contains(query, "envVariablesByProjectEnvironment") {
			// Verify environment was passed
			if variables["environment"] == nil {
				t.Error("expected environment variable to be set")
			}
			return map[string]any{
				"envVariablesByProjectEnvironment": []map[string]any{
					{"id": 20, "name": "MY_VAR", "value": "env-found", "scope": "RUNTIME"},
				},
			}, nil
		}
		return nil, errors.New("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	v, err := c.GetVariable(context.Background(), "MY_VAR", 1, &envID)
	if err != nil {
		t.Fatalf("GetVariable (legacy with env) failed: %v", err)
	}
	if v.Value != "env-found" {
		t.Errorf("expected value=env-found, got %s", v.Value)
	}
}

func TestDeleteVariable_Legacy_WithEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": nil}, nil
		}
		if strings.Contains(query, "deleteEnvVariable") {
			// Verify environment was passed
			input, _ := variables["input"].(map[string]any)
			if input["environment"] == nil {
				t.Error("expected environment in delete input")
			}
			return map[string]any{"deleteEnvVariable": "success"}, nil
		}
		return nil, errors.New("unexpected query")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	err := c.DeleteVariable(context.Background(), "MY_VAR", 1, &envID)
	if err != nil {
		t.Fatalf("DeleteVariable (legacy with env) failed: %v", err)
	}
}

func TestGetEnvironmentName_UnmarshalError(t *testing.T) {
	// getEnvironmentName is tested indirectly via GetVariable (new API with env)
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if strings.Contains(query, "__type") {
			return map[string]any{"__type": map[string]any{"name": "EnvVariableByNameInput"}}, nil
		}
		if strings.Contains(query, "allProjects") {
			return map[string]any{
				"allProjects": []map[string]any{
					{"id": 1, "name": "my-project", "gitUrl": "git@example.com:repo.git", "openshift": map[string]any{"id": 1, "name": "c1"}},
				},
			}, nil
		}
		// environmentById returns a number instead of an object
		if strings.Contains(query, "environmentById") {
			return map[string]any{"environmentById": 12345}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	envID := 5
	_, err := c.GetVariable(context.Background(), "MY_VAR", 1, &envID)
	if err == nil {
		t.Fatal("expected error when environmentById returns a non-object")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("expected unmarshal error, got: %v", err)
	}
}
