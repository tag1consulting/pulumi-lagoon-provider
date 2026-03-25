package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreateTaskDefinition(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addAdvancedTaskDefinition") {
			t.Errorf("expected addAdvancedTaskDefinition mutation")
		}
		return map[string]any{
			"addAdvancedTaskDefinition": map[string]any{
				"id":          100,
				"name":        "my-task",
				"description": "A test task",
				"type":        "COMMAND",
				"service":     "cli",
				"command":     "drush cr",
				"permission":  "DEVELOPER",
				"groupName":   "",
				"created":     "2024-01-01T00:00:00Z",
				"project":     nil,
				"environment": nil,
				"advancedTaskDefinitionArguments": []map[string]any{},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	td, err := c.CreateTaskDefinition(context.Background(), map[string]any{
		"name":    "my-task",
		"type":    "COMMAND",
		"service": "cli",
		"command": "drush cr",
	})
	if err != nil {
		t.Fatalf("CreateTaskDefinition failed: %v", err)
	}
	if td.ID != 100 {
		t.Errorf("expected ID=100, got %d", td.ID)
	}
	if td.Name != "my-task" {
		t.Errorf("expected name=my-task, got %s", td.Name)
	}
	if td.Type != "COMMAND" {
		t.Errorf("expected type=COMMAND, got %s", td.Type)
	}
}

func TestGetTaskDefinitionByID(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "advancedTaskDefinitionById") {
			t.Errorf("expected advancedTaskDefinitionById query")
		}
		return map[string]any{
			"advancedTaskDefinitionById": map[string]any{
				"id":          50,
				"name":        "db-backup",
				"description": "Backup database",
				"type":        "IMAGE",
				"service":     "cli",
				"image":       "amazeeio/lagoon-cli:latest",
				"permission":  "MAINTAINER",
				"created":     "2024-01-01T00:00:00Z",
				"project": map[string]any{
					"id":   5,
					"name": "my-project",
				},
				"environment": nil,
				"advancedTaskDefinitionArguments": []map[string]any{
					{"id": 1, "name": "db_name", "displayName": "Database Name", "type": "STRING"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	td, err := c.GetTaskDefinitionByID(context.Background(), 50)
	if err != nil {
		t.Fatalf("GetTaskDefinitionByID failed: %v", err)
	}
	if td.Name != "db-backup" {
		t.Errorf("expected name=db-backup, got %s", td.Name)
	}
	if td.ProjectID == nil || *td.ProjectID != 5 {
		t.Error("expected ProjectID=5")
	}
	if td.EnvironmentID != nil {
		t.Error("expected EnvironmentID to be nil")
	}
	if len(td.Arguments) != 1 {
		t.Errorf("expected 1 argument, got %d", len(td.Arguments))
	}
	if td.Arguments[0].Name != "db_name" {
		t.Errorf("expected argument name=db_name, got %s", td.Arguments[0].Name)
	}
}

func TestGetTaskDefinitionByID_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"advancedTaskDefinitionById": nil,
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetTaskDefinitionByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

func TestDeleteTaskDefinition(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteAdvancedTaskDefinition") {
			t.Errorf("expected deleteAdvancedTaskDefinition mutation")
		}
		return map[string]any{"deleteAdvancedTaskDefinition": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteTaskDefinition(context.Background(), 100)
	if err != nil {
		t.Fatalf("DeleteTaskDefinition failed: %v", err)
	}
}

func TestNormalizeTaskDefinition_ObjectProject(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:      1,
		Name:    "task-1",
		Type:    "COMMAND",
		Service: "cli",
		Project: json.RawMessage(`{"id": 10, "name": "project-10"}`),
	}
	td := normalizeTaskDefinition(raw)
	if td.ProjectID == nil || *td.ProjectID != 10 {
		t.Errorf("expected ProjectID=10 from object, got %v", td.ProjectID)
	}
}

func TestNormalizeTaskDefinition_IntProject(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:      1,
		Name:    "task-1",
		Type:    "COMMAND",
		Service: "cli",
		Project: json.RawMessage(`42`),
	}
	td := normalizeTaskDefinition(raw)
	if td.ProjectID == nil || *td.ProjectID != 42 {
		t.Errorf("expected ProjectID=42 from int, got %v", td.ProjectID)
	}
}

func TestNormalizeTaskDefinition_NullProject(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:      1,
		Name:    "task-1",
		Type:    "COMMAND",
		Service: "cli",
		Project: json.RawMessage(`null`),
	}
	td := normalizeTaskDefinition(raw)
	if td.ProjectID != nil {
		t.Errorf("expected nil ProjectID for null, got %v", td.ProjectID)
	}
}

func TestNormalizeTaskDefinition_EmptyProject(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:      1,
		Name:    "task-1",
		Type:    "COMMAND",
		Service: "cli",
		// Project is nil (empty RawMessage)
	}
	td := normalizeTaskDefinition(raw)
	if td.ProjectID != nil {
		t.Errorf("expected nil ProjectID for empty, got %v", td.ProjectID)
	}
}

func TestNormalizeTaskDefinition_ObjectEnvironment(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:          1,
		Name:        "task-1",
		Type:        "COMMAND",
		Service:     "cli",
		Environment: json.RawMessage(`{"id": 20, "name": "main"}`),
	}
	td := normalizeTaskDefinition(raw)
	if td.EnvironmentID == nil || *td.EnvironmentID != 20 {
		t.Errorf("expected EnvironmentID=20 from object, got %v", td.EnvironmentID)
	}
}

func TestNormalizeTaskDefinition_IntEnvironment(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:          1,
		Name:        "task-1",
		Type:        "COMMAND",
		Service:     "cli",
		Environment: json.RawMessage(`99`),
	}
	td := normalizeTaskDefinition(raw)
	if td.EnvironmentID == nil || *td.EnvironmentID != 99 {
		t.Errorf("expected EnvironmentID=99 from int, got %v", td.EnvironmentID)
	}
}

func TestNormalizeTaskDefinition_Arguments(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:      1,
		Name:    "task-1",
		Type:    "COMMAND",
		Service: "cli",
		Arguments: []TaskArgument{
			{ID: 1, Name: "arg1", DisplayName: "Argument 1", Type: "STRING"},
			{ID: 2, Name: "arg2", DisplayName: "Argument 2", Type: "ENVIRONMENT_SOURCE_NAME"},
		},
	}
	td := normalizeTaskDefinition(raw)
	if len(td.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(td.Arguments))
	}
	if td.Arguments[0].DisplayName != "Argument 1" {
		t.Errorf("expected displayName=Argument 1, got %s", td.Arguments[0].DisplayName)
	}
}

func TestNormalizeTaskDefinition_AllFields(t *testing.T) {
	raw := taskDefinitionRaw{
		ID:               1,
		Name:             "full-task",
		Description:      "A full task",
		Type:             "IMAGE",
		Service:          "cli",
		Image:            "myimage:latest",
		Command:          "",
		Permission:       "MAINTAINER",
		ConfirmationText: "Are you sure?",
		GroupName:        "my-group",
		Created:          "2024-01-01T00:00:00Z",
	}
	td := normalizeTaskDefinition(raw)
	if td.Description != "A full task" {
		t.Errorf("expected description, got %s", td.Description)
	}
	if td.Image != "myimage:latest" {
		t.Errorf("expected image, got %s", td.Image)
	}
	if td.ConfirmationText != "Are you sure?" {
		t.Errorf("expected confirmationText, got %s", td.ConfirmationText)
	}
	if td.GroupName != "my-group" {
		t.Errorf("expected groupName, got %s", td.GroupName)
	}
}

func TestGetTasksByEnvironment_Success(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// New API query — advancedTasksForEnvironment
		if strings.Contains(query, "advancedTasksForEnvironment") {
			return map[string]any{
				"advancedTasksForEnvironment": []map[string]any{
					{
						"id":          10,
						"name":        "task-1",
						"description": "First task",
						"type":        "COMMAND",
						"service":     "cli",
						"command":     "drush cr",
						"permission":  "DEVELOPER",
						"project":     nil,
						"environment": nil,
						"advancedTaskDefinitionArguments": []map[string]any{},
					},
					{
						"id":          20,
						"name":        "task-2",
						"description": "Second task",
						"type":        "IMAGE",
						"service":     "cli",
						"image":       "myimage:latest",
						"permission":  "MAINTAINER",
						"project":     5,
						"environment": 10,
						"advancedTaskDefinitionArguments": []map[string]any{},
					},
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	tasks, err := c.GetTasksByEnvironment(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetTasksByEnvironment failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Name != "task-1" {
		t.Errorf("expected task-1, got %s", tasks[0].Name)
	}
	if tasks[1].Name != "task-2" {
		t.Errorf("expected task-2, got %s", tasks[1].Name)
	}
	// New API returns project/environment as ints
	if tasks[1].ProjectID == nil || *tasks[1].ProjectID != 5 {
		t.Errorf("expected ProjectID=5 for task-2, got %v", tasks[1].ProjectID)
	}
	if tasks[1].EnvironmentID == nil || *tasks[1].EnvironmentID != 10 {
		t.Errorf("expected EnvironmentID=10 for task-2, got %v", tasks[1].EnvironmentID)
	}
}

func TestGetTasksByEnvironment_FallbackToLegacy_CannotQueryField(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// New API query returns "Cannot query field" error
		if strings.Contains(query, "advancedTasksForEnvironment") {
			return nil, errors.New("Cannot query field 'advancedTasksForEnvironment'")
		}
		// Legacy fallback
		if strings.Contains(query, "advancedTasksByEnvironment") {
			return map[string]any{
				"advancedTasksByEnvironment": []map[string]any{
					{
						"id":          30,
						"name":        "legacy-task",
						"description": "Legacy task",
						"type":        "COMMAND",
						"service":     "cli",
						"command":     "drush status",
						"permission":  "DEVELOPER",
						"project":     map[string]any{"id": 1, "name": "project-1"},
						"environment": map[string]any{"id": 5, "name": "main"},
						"advancedTaskDefinitionArguments": []map[string]any{},
					},
				},
			}, nil
		}
		return nil, errors.New("unexpected query: " + query)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	tasks, err := c.GetTasksByEnvironment(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTasksByEnvironment (fallback) failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "legacy-task" {
		t.Errorf("expected legacy-task, got %s", tasks[0].Name)
	}
	// Legacy API returns project as object
	if tasks[0].ProjectID == nil || *tasks[0].ProjectID != 1 {
		t.Errorf("expected ProjectID=1, got %v", tasks[0].ProjectID)
	}
}

func TestGetTasksByEnvironment_FallbackToLegacy_HTTP400(t *testing.T) {
	callCount := 0
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		body, _ := io.ReadAll(r.Body)
		var req graphQLRequest
		json.Unmarshal(body, &req)

		if callCount == 1 {
			// First call (new API) — return a GraphQL error with "HTTP 400" in the message
			resp := map[string]any{
				"errors": []map[string]any{{"message": "HTTP 400: Bad Request"}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// Second call (legacy fallback)
		resp := map[string]any{
			"data": map[string]any{
				"advancedTasksByEnvironment": []map[string]any{
					{
						"id":          40,
						"name":        "fallback-task",
						"description": "Fallback",
						"type":        "COMMAND",
						"service":     "cli",
						"command":     "echo ok",
						"permission":  "DEVELOPER",
						"project":     nil,
						"environment": nil,
						"advancedTaskDefinitionArguments": []map[string]any{},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	tasks, err := c.GetTasksByEnvironment(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTasksByEnvironment (HTTP 400 fallback) failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "fallback-task" {
		t.Errorf("expected fallback-task, got %s", tasks[0].Name)
	}
}

func TestGetTasksByEnvironment_FallbackToLegacy_NestedUnknownArgument(t *testing.T) {
	callCount := 0
	server := mockGraphQLServerRaw(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First call — return nested "Unknown argument" error
			resp := map[string]any{
				"errors": []map[string]any{
					{"message": "Unknown argument 'environment' on field 'Query.advancedTasksForEnvironment'"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		// Second call (legacy)
		resp := map[string]any{
			"data": map[string]any{
				"advancedTasksByEnvironment": []map[string]any{
					{
						"id":          50,
						"name":        "nested-fallback",
						"description": "",
						"type":        "COMMAND",
						"service":     "cli",
						"command":     "true",
						"permission":  "GUEST",
						"project":     nil,
						"environment": nil,
						"advancedTaskDefinitionArguments": []map[string]any{},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	tasks, err := c.GetTasksByEnvironment(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetTasksByEnvironment (nested Unknown argument fallback) failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Name != "nested-fallback" {
		t.Errorf("expected nested-fallback, got %s", tasks[0].Name)
	}
}

func TestGetTasksByEnvironment_NonFallbackError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		// Return a real API error that should NOT trigger fallback
		return nil, errors.New("Internal server processing error")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetTasksByEnvironment(context.Background(), 5)
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
	if !errors.Is(err, ErrAPI) {
		t.Errorf("expected ErrAPI, got %T: %v", err, err)
	}
	// Ensure it's not a "not found" type error — it should be the actual API error
	if strings.Contains(err.Error(), "Cannot query field") {
		t.Error("should not contain 'Cannot query field' — this was a real error")
	}
}

func TestIsFieldNotFoundOrLegacyError(t *testing.T) {
	tests := []struct {
		name    string
		apiErr  *LagoonAPIError
		want    bool
	}{
		{
			name:   "top-level Cannot query field",
			apiErr: &LagoonAPIError{Message: "Cannot query field 'advancedTasksForEnvironment'"},
			want:   true,
		},
		{
			name:   "top-level HTTP 400",
			apiErr: &LagoonAPIError{Message: "HTTP 400: Bad Request"},
			want:   true,
		},
		{
			name: "nested Cannot query field",
			apiErr: &LagoonAPIError{
				Message: "some wrapper error",
				Errors:  []GraphQLError{{Message: "Cannot query field 'advancedTasksForEnvironment'"}},
			},
			want: true,
		},
		{
			name: "nested Unknown argument",
			apiErr: &LagoonAPIError{
				Message: "some wrapper error",
				Errors:  []GraphQLError{{Message: "Unknown argument 'environment' on field 'Query.advancedTasksForEnvironment'"}},
			},
			want: true,
		},
		{
			name: "nested mixed — one matches",
			apiErr: &LagoonAPIError{
				Message: "multiple errors",
				Errors: []GraphQLError{
					{Message: "Something else"},
					{Message: "Cannot query field 'foo'"},
				},
			},
			want: true,
		},
		{
			name:   "unrelated top-level error",
			apiErr: &LagoonAPIError{Message: "Internal server error"},
			want:   false,
		},
		{
			name: "unrelated nested error",
			apiErr: &LagoonAPIError{
				Message: "wrapper",
				Errors:  []GraphQLError{{Message: "Permission denied"}},
			},
			want: false,
		},
		{
			name:   "empty message",
			apiErr: &LagoonAPIError{Message: ""},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFieldNotFoundOrLegacyError(tt.apiErr)
			if got != tt.want {
				t.Errorf("isFieldNotFoundOrLegacyError() = %v, want %v", got, tt.want)
			}
		})
	}
}
