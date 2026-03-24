package client

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestCreateProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addProject") {
			t.Errorf("expected addProject mutation, got: %s", query)
		}
		return map[string]any{
			"addProject": map[string]any{
				"id":      42,
				"name":    "test-project",
				"gitUrl":  "git@example.com:repo.git",
				"created": "2024-01-01T00:00:00Z",
				"openshift": map[string]any{
					"id":   1,
					"name": "cluster-1",
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	p, err := c.CreateProject(context.Background(), map[string]any{
		"name":      "test-project",
		"gitUrl":    "git@example.com:repo.git",
		"openshift": 1,
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if p.ID != 42 {
		t.Errorf("expected ID=42, got %d", p.ID)
	}
	if p.Name != "test-project" {
		t.Errorf("expected name=test-project, got %s", p.Name)
	}
	if p.OpenshiftID != 1 {
		t.Errorf("expected OpenshiftID=1 (normalized from nested), got %d", p.OpenshiftID)
	}
}

func TestGetProjectByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "projectByName") {
			t.Errorf("expected projectByName query")
		}
		return map[string]any{
			"projectByName": map[string]any{
				"id":      10,
				"name":    "my-project",
				"gitUrl":  "git@example.com:my.git",
				"created": "2024-01-01T00:00:00Z",
				"openshift": map[string]any{
					"id":   2,
					"name": "cluster-2",
				},
				"productionEnvironment": "main",
				"branches":              "main|develop",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	p, err := c.GetProjectByName(context.Background(), "my-project")
	if err != nil {
		t.Fatalf("GetProjectByName failed: %v", err)
	}
	if p.ID != 10 {
		t.Errorf("expected ID=10, got %d", p.ID)
	}
	if p.ProductionEnvironment != "main" {
		t.Errorf("expected productionEnvironment=main, got %s", p.ProductionEnvironment)
	}
	if p.OpenshiftID != 2 {
		t.Errorf("expected OpenshiftID=2, got %d", p.OpenshiftID)
	}
}

func TestGetProjectByID(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "allProjects") {
			t.Errorf("expected allProjects query")
		}
		return map[string]any{
			"allProjects": []map[string]any{
				{
					"id":        1,
					"name":      "project-1",
					"gitUrl":    "git@example.com:p1.git",
					"openshift": map[string]any{"id": 1, "name": "c1"},
				},
				{
					"id":        2,
					"name":      "project-2",
					"gitUrl":    "git@example.com:p2.git",
					"openshift": map[string]any{"id": 1, "name": "c1"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	p, err := c.GetProjectByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetProjectByID failed: %v", err)
	}
	if p.Name != "project-2" {
		t.Errorf("expected project-2, got %s", p.Name)
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allProjects": []map[string]any{
				{
					"id":        1,
					"name":      "project-1",
					"gitUrl":    "git@example.com:p1.git",
					"openshift": map[string]any{"id": 1, "name": "c1"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetProjectByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent project")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestUpdateProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "updateProject") {
			t.Errorf("expected updateProject mutation")
		}
		// Verify id is included (JSON numbers decode as float64)
		if input, ok := variables["input"].(map[string]any); ok {
			if id, ok := input["id"].(float64); !ok || int(id) != 42 {
				t.Errorf("expected id=42 in input, got %v", input["id"])
			}
		}
		return map[string]any{
			"updateProject": map[string]any{
				"id":        42,
				"name":      "my-project",
				"gitUrl":    "git@new.com:repo.git",
				"openshift": map[string]any{"id": 1, "name": "c1"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	p, err := c.UpdateProject(context.Background(), 42, map[string]any{"gitUrl": "git@new.com:repo.git"})
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}
	if p.GitURL != "git@new.com:repo.git" {
		t.Errorf("expected updated git URL")
	}
}

func TestDeleteProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteProject") {
			t.Errorf("expected deleteProject mutation")
		}
		return map[string]any{"deleteProject": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteProject(context.Background(), "my-project")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
}

func TestNormalizeProject_NilOpenshift(t *testing.T) {
	raw := projectRaw{
		ID:     1,
		Name:   "test",
		GitURL: "git@example.com:repo.git",
		// Openshift is empty json.RawMessage
	}
	p := normalizeProject(raw)
	if p.OpenshiftID != 0 {
		t.Errorf("expected OpenshiftID=0 for nil openshift, got %d", p.OpenshiftID)
	}
}

func TestNormalizeProject_WithOpenshift(t *testing.T) {
	raw := projectRaw{
		ID:        1,
		Name:      "test",
		GitURL:    "git@example.com:repo.git",
		Openshift: json.RawMessage(`{"id": 5, "name": "cluster-5"}`),
	}
	p := normalizeProject(raw)
	if p.OpenshiftID != 5 {
		t.Errorf("expected OpenshiftID=5, got %d", p.OpenshiftID)
	}
}

func TestCreateProject_APIError(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return nil, errors.New("duplicate project name")
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.CreateProject(context.Background(), map[string]any{
		"name":      "existing-project",
		"gitUrl":    "git@example.com:repo.git",
		"openshift": 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
