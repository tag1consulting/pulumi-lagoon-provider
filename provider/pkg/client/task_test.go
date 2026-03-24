package client

import (
	"context"
	"encoding/json"
	"errors"
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
