package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCreateDeployTarget(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addKubernetes") {
			t.Errorf("expected addKubernetes mutation")
		}
		return map[string]any{
			"addKubernetes": map[string]any{
				"id":            1,
				"name":          "cluster-1",
				"consoleUrl":    "https://k8s.example.com",
				"cloudProvider": "kind",
				"cloudRegion":   "local",
				"created":       "2024-01-01T00:00:00Z",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dt, err := c.CreateDeployTarget(context.Background(), map[string]any{
		"name":          "cluster-1",
		"consoleUrl":    "https://k8s.example.com",
		"cloudProvider": "kind",
		"cloudRegion":   "local",
	})
	if err != nil {
		t.Fatalf("CreateDeployTarget failed: %v", err)
	}
	if dt.ID != 1 {
		t.Errorf("expected ID=1, got %d", dt.ID)
	}
	if dt.Name != "cluster-1" {
		t.Errorf("expected name=cluster-1, got %s", dt.Name)
	}
}

func TestGetDeployTargetByID(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allKubernetes": []map[string]any{
				{"id": 1, "name": "cluster-1", "consoleUrl": "https://c1.example.com"},
				{"id": 2, "name": "cluster-2", "consoleUrl": "https://c2.example.com"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dt, err := c.GetDeployTargetByID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetDeployTargetByID failed: %v", err)
	}
	if dt.Name != "cluster-2" {
		t.Errorf("expected cluster-2, got %s", dt.Name)
	}
}

func TestGetDeployTargetByID_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allKubernetes": []map[string]any{
				{"id": 1, "name": "cluster-1", "consoleUrl": "https://c1.example.com"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetDeployTargetByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

func TestGetDeployTargetByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allKubernetes": []map[string]any{
				{"id": 1, "name": "cluster-1", "consoleUrl": "https://c1.example.com"},
				{"id": 2, "name": "cluster-2", "consoleUrl": "https://c2.example.com"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dt, err := c.GetDeployTargetByName(context.Background(), "cluster-2")
	if err != nil {
		t.Fatalf("GetDeployTargetByName failed: %v", err)
	}
	if dt.ID != 2 {
		t.Errorf("expected ID=2, got %d", dt.ID)
	}
}

func TestGetDeployTargetByName_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allKubernetes": []map[string]any{},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetDeployTargetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

func TestUpdateDeployTarget(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "updateKubernetes") {
			t.Errorf("expected updateKubernetes mutation")
		}
		return map[string]any{
			"updateKubernetes": map[string]any{
				"id":         1,
				"name":       "cluster-1",
				"consoleUrl": "https://new.example.com",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dt, err := c.UpdateDeployTarget(context.Background(), 1, map[string]any{"consoleUrl": "https://new.example.com"})
	if err != nil {
		t.Fatalf("UpdateDeployTarget failed: %v", err)
	}
	if dt.ConsoleURL != "https://new.example.com" {
		t.Errorf("expected updated consoleUrl")
	}
}

func TestDeleteDeployTarget(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteKubernetes") {
			t.Errorf("expected deleteKubernetes mutation")
		}
		return map[string]any{"deleteKubernetes": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteDeployTarget(context.Background(), "cluster-1")
	if err != nil {
		t.Fatalf("DeleteDeployTarget failed: %v", err)
	}
}

func TestGetAllDeployTargets(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allKubernetes": []map[string]any{
				{"id": 1, "name": "c1", "consoleUrl": "https://c1.example.com"},
				{"id": 2, "name": "c2", "consoleUrl": "https://c2.example.com"},
				{"id": 3, "name": "c3", "consoleUrl": "https://c3.example.com"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	targets, err := c.GetAllDeployTargets(context.Background())
	if err != nil {
		t.Fatalf("GetAllDeployTargets failed: %v", err)
	}
	if len(targets) != 3 {
		t.Errorf("expected 3 targets, got %d", len(targets))
	}
}

// --- Deploy Target Config Tests ---

func TestCreateDeployTargetConfig(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addDeployTargetConfig") {
			t.Errorf("expected addDeployTargetConfig mutation")
		}
		return map[string]any{
			"addDeployTargetConfig": map[string]any{
				"id":       10,
				"weight":   1,
				"branches": "main",
				"deployTarget": map[string]any{
					"id":   1,
					"name": "cluster-1",
				},
				"project": map[string]any{
					"id":   5,
					"name": "my-project",
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dtc, err := c.CreateDeployTargetConfig(context.Background(), map[string]any{
		"project":      5,
		"deployTarget": 1,
		"branches":     "main",
	})
	if err != nil {
		t.Fatalf("CreateDeployTargetConfig failed: %v", err)
	}
	if dtc.ID != 10 {
		t.Errorf("expected ID=10, got %d", dtc.ID)
	}
	if dtc.DeployTargetID != 1 {
		t.Errorf("expected DeployTargetID=1, got %d", dtc.DeployTargetID)
	}
	if dtc.ProjectID != 5 {
		t.Errorf("expected ProjectID=5, got %d", dtc.ProjectID)
	}
}

func TestGetDeployTargetConfigByID(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"deployTargetConfigsByProjectId": []map[string]any{
				{
					"id":       10,
					"weight":   1,
					"branches": "main",
					"deployTarget": map[string]any{
						"id":   1,
						"name": "cluster-1",
					},
					"project": map[string]any{
						"id":   5,
						"name": "my-project",
					},
				},
				{
					"id":       20,
					"weight":   2,
					"branches": "develop",
					"deployTarget": map[string]any{
						"id":   2,
						"name": "cluster-2",
					},
					"project": map[string]any{
						"id":   5,
						"name": "my-project",
					},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dtc, err := c.GetDeployTargetConfigByID(context.Background(), 20, 5)
	if err != nil {
		t.Fatalf("GetDeployTargetConfigByID failed: %v", err)
	}
	if dtc.ID != 20 {
		t.Errorf("expected ID=20, got %d", dtc.ID)
	}
	if dtc.Branches != "develop" {
		t.Errorf("expected branches=develop, got %s", dtc.Branches)
	}
}

func TestGetDeployTargetConfigByID_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"deployTargetConfigsByProjectId": []map[string]any{},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetDeployTargetConfigByID(context.Background(), 999, 5)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

func TestUpdateDeployTargetConfig(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "updateDeployTargetConfig") {
			t.Errorf("expected updateDeployTargetConfig mutation")
		}
		return map[string]any{
			"updateDeployTargetConfig": map[string]any{
				"id":       10,
				"weight":   5,
				"branches": "main|develop",
				"deployTarget": map[string]any{
					"id":   1,
					"name": "cluster-1",
				},
				"project": map[string]any{
					"id":   5,
					"name": "my-project",
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	dtc, err := c.UpdateDeployTargetConfig(context.Background(), 10, map[string]any{"weight": 5})
	if err != nil {
		t.Fatalf("UpdateDeployTargetConfig failed: %v", err)
	}
	if dtc.Weight != 5 {
		t.Errorf("expected weight=5, got %d", dtc.Weight)
	}
}

func TestDeleteDeployTargetConfig(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteDeployTargetConfig") {
			t.Errorf("expected deleteDeployTargetConfig mutation")
		}
		return map[string]any{"deleteDeployTargetConfig": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteDeployTargetConfig(context.Background(), 10, 5)
	if err != nil {
		t.Fatalf("DeleteDeployTargetConfig failed: %v", err)
	}
}

func TestNormalizeDeployTargetConfig_NilNested(t *testing.T) {
	raw := deployTargetConfigRaw{
		ID:       1,
		Weight:   1,
		Branches: "main",
	}
	dtc := normalizeDeployTargetConfig(raw)
	if dtc.DeployTargetID != 0 {
		t.Errorf("expected DeployTargetID=0 for nil, got %d", dtc.DeployTargetID)
	}
	if dtc.ProjectID != 0 {
		t.Errorf("expected ProjectID=0 for nil, got %d", dtc.ProjectID)
	}
}

func TestNormalizeDeployTargetConfig_WithNested(t *testing.T) {
	raw := deployTargetConfigRaw{
		ID:       1,
		Weight:   1,
		Branches: "main",
		DeployTarget: &struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{ID: 3, Name: "cluster"},
		Project: &struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{ID: 7, Name: "project"},
	}
	dtc := normalizeDeployTargetConfig(raw)
	if dtc.DeployTargetID != 3 {
		t.Errorf("expected DeployTargetID=3, got %d", dtc.DeployTargetID)
	}
	if dtc.ProjectID != 7 {
		t.Errorf("expected ProjectID=7, got %d", dtc.ProjectID)
	}
}
