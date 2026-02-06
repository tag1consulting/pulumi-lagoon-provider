package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAddOrUpdateEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addOrUpdateEnvironment") {
			t.Errorf("expected addOrUpdateEnvironment mutation")
		}
		return map[string]any{
			"addOrUpdateEnvironment": map[string]any{
				"id":              10,
				"name":            "main",
				"environmentType": "PRODUCTION",
				"deployType":      "BRANCH",
				"route":           "https://main.example.com",
				"routes":          "https://main.example.com",
				"created":         "2024-01-01T00:00:00Z",
				"project": map[string]any{
					"id":   1,
					"name": "test-project",
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	env, err := c.AddOrUpdateEnvironment(context.Background(), map[string]any{
		"name":            "main",
		"project":         1,
		"deployType":      "BRANCH",
		"environmentType": "PRODUCTION",
	})
	if err != nil {
		t.Fatalf("AddOrUpdateEnvironment failed: %v", err)
	}
	if env.ID != 10 {
		t.Errorf("expected ID=10, got %d", env.ID)
	}
	if env.ProjectID != 1 {
		t.Errorf("expected ProjectID=1 (normalized), got %d", env.ProjectID)
	}
	if env.Route != "https://main.example.com" {
		t.Errorf("expected route, got %s", env.Route)
	}
}

func TestGetEnvironmentByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "environmentByName") {
			t.Errorf("expected environmentByName query")
		}
		return map[string]any{
			"environmentByName": map[string]any{
				"id":              5,
				"name":            "develop",
				"environmentType": "DEVELOPMENT",
				"deployType":      "BRANCH",
				"route":           "https://develop.example.com",
				"created":         "2024-01-01T00:00:00Z",
				"project": map[string]any{
					"id":   1,
					"name": "test-project",
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	env, err := c.GetEnvironmentByName(context.Background(), "develop", 1)
	if err != nil {
		t.Fatalf("GetEnvironmentByName failed: %v", err)
	}
	if env.Name != "develop" {
		t.Errorf("expected name=develop, got %s", env.Name)
	}
	if env.ProjectID != 1 {
		t.Errorf("expected ProjectID=1, got %d", env.ProjectID)
	}
}

func TestGetEnvironmentByName_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"environmentByName": nil,
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetEnvironmentByName(context.Background(), "nonexistent", 1)
	if err == nil {
		t.Fatal("expected error for not found environment")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDeleteEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteEnvironment") {
			t.Errorf("expected deleteEnvironment mutation")
		}
		return map[string]any{"deleteEnvironment": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteEnvironment(context.Background(), "develop", 1)
	if err != nil {
		t.Fatalf("DeleteEnvironment failed: %v", err)
	}
}

func TestNormalizeEnvironment_NilProject(t *testing.T) {
	raw := environmentRaw{
		ID:   1,
		Name: "main",
	}
	env := normalizeEnvironment(raw)
	if env.ProjectID != 0 {
		t.Errorf("expected ProjectID=0 for nil project, got %d", env.ProjectID)
	}
}

func TestNormalizeEnvironment_WithProject(t *testing.T) {
	raw := environmentRaw{
		ID:   1,
		Name: "main",
		Project: &struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{ID: 5, Name: "test-project"},
		EnvironmentType: "PRODUCTION",
		DeployType:      "BRANCH",
	}
	env := normalizeEnvironment(raw)
	if env.ProjectID != 5 {
		t.Errorf("expected ProjectID=5, got %d", env.ProjectID)
	}
	if env.EnvironmentType != "PRODUCTION" {
		t.Errorf("expected PRODUCTION, got %s", env.EnvironmentType)
	}
}

func TestNormalizeEnvironment_AllFields(t *testing.T) {
	autoIdle := 1
	raw := environmentRaw{
		ID:              1,
		Name:            "main",
		EnvironmentType: "PRODUCTION",
		DeployType:      "BRANCH",
		DeployBaseRef:   "main",
		DeployHeadRef:   "feature",
		DeployTitle:     "PR #1",
		AutoIdle:        &autoIdle,
		Route:           "https://main.example.com",
		Routes:          "https://main.example.com,https://www.example.com",
		Created:         "2024-01-01T00:00:00Z",
	}
	env := normalizeEnvironment(raw)
	if env.DeployBaseRef != "main" {
		t.Errorf("expected deployBaseRef=main, got %s", env.DeployBaseRef)
	}
	if env.DeployHeadRef != "feature" {
		t.Errorf("expected deployHeadRef=feature, got %s", env.DeployHeadRef)
	}
	if env.DeployTitle != "PR #1" {
		t.Errorf("expected deployTitle=PR #1, got %s", env.DeployTitle)
	}
	if env.AutoIdle == nil || *env.AutoIdle != 1 {
		t.Errorf("expected autoIdle=1, got %v", env.AutoIdle)
	}
}
