package client

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// --- Route CRUD Tests ---

func TestAddRouteToProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addRouteToProject") {
			t.Errorf("expected addRouteToProject mutation")
		}
		return map[string]any{
			"addRouteToProject": map[string]any{
				"id":      100,
				"domain":  "example.com",
				"service": "nginx",
				"project": map[string]any{"name": "my-project"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.AddRouteToProject(context.Background(), map[string]any{
		"project": "my-project",
		"domain":  "example.com",
		"service": "nginx",
	})
	if err != nil {
		t.Fatalf("AddRouteToProject failed: %v", err)
	}
	if r.ID != 100 {
		t.Errorf("expected ID=100, got %d", r.ID)
	}
	if r.Domain != "example.com" {
		t.Errorf("expected domain=example.com, got %s", r.Domain)
	}
	if r.ProjectName != "my-project" {
		t.Errorf("expected ProjectName=my-project, got %s", r.ProjectName)
	}
}

func TestGetRouteByDomain(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "projectByName") {
			t.Errorf("expected projectByName query")
		}
		return map[string]any{
			"projectByName": map[string]any{
				"apiRoutes": []map[string]any{
					{
						"id":      50,
						"domain":  "test.example.com",
						"service": "nginx",
						"project": map[string]any{"name": "test-project"},
					},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.GetRouteByDomain(context.Background(), "test-project", "test.example.com")
	if err != nil {
		t.Fatalf("GetRouteByDomain failed: %v", err)
	}
	if r.ID != 50 {
		t.Errorf("expected ID=50, got %d", r.ID)
	}
	if r.Domain != "test.example.com" {
		t.Errorf("expected domain=test.example.com, got %s", r.Domain)
	}
}

func TestGetRouteByDomain_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"projectByName": map[string]any{
				"apiRoutes": []map[string]any{
					{"id": 1, "domain": "other.com"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetRouteByDomain(context.Background(), "test-project", "nonexistent.com")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *LagoonNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected LagoonNotFoundError, got %T", err)
	}
}

func TestGetRouteByDomain_ProjectNotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"projectByName": nil}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetRouteByDomain(context.Background(), "nonexistent", "example.com")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *LagoonNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected LagoonNotFoundError, got %T", err)
	}
}

func TestUpdateRouteOnProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "updateRouteOnProject") {
			t.Errorf("expected updateRouteOnProject mutation")
		}
		return map[string]any{
			"updateRouteOnProject": map[string]any{
				"id":      100,
				"domain":  "example.com",
				"service": "updated-service",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.UpdateRouteOnProject(context.Background(), 100, map[string]any{"service": "updated-service"})
	if err != nil {
		t.Fatalf("UpdateRouteOnProject failed: %v", err)
	}
	if r.Service != "updated-service" {
		t.Errorf("expected service=updated-service, got %s", r.Service)
	}
}

func TestDeleteRoute(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteRoute") {
			t.Errorf("expected deleteRoute mutation")
		}
		return map[string]any{"deleteRoute": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteRoute(context.Background(), 100)
	if err != nil {
		t.Fatalf("DeleteRoute failed: %v", err)
	}
}

func TestDeleteRoute_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return nil, &LagoonAPIError{Message: "Route not found"}
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteRoute(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *LagoonNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected LagoonNotFoundError, got %T", err)
	}
}

// --- Environment Association Tests ---

func TestAddOrUpdateRouteOnEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addOrUpdateRouteOnEnvironment") {
			t.Errorf("expected addOrUpdateRouteOnEnvironment mutation")
		}
		return map[string]any{
			"addOrUpdateRouteOnEnvironment": map[string]any{
				"id":          100,
				"domain":      "example.com",
				"environment": map[string]any{"name": "main"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.AddOrUpdateRouteOnEnvironment(context.Background(), map[string]any{
		"domain":      "example.com",
		"environment": "main",
	})
	if err != nil {
		t.Fatalf("AddOrUpdateRouteOnEnvironment failed: %v", err)
	}
	if r.EnvironmentName != "main" {
		t.Errorf("expected EnvironmentName=main, got %s", r.EnvironmentName)
	}
}

func TestRemoveRouteFromEnvironment(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "removeRouteFromEnvironment") {
			t.Errorf("expected removeRouteFromEnvironment mutation")
		}
		return map[string]any{"removeRouteFromEnvironment": map[string]any{"id": 100}}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.RemoveRouteFromEnvironment(context.Background(), "example.com", "my-project", "main")
	if err != nil {
		t.Fatalf("RemoveRouteFromEnvironment failed: %v", err)
	}
}

func TestRemoveRouteFromEnvironment_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return nil, &LagoonAPIError{Message: "Route not found"}
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.RemoveRouteFromEnvironment(context.Background(), "nonexistent.com", "project", "main")
	if err == nil {
		t.Fatal("expected error")
	}
	var notFoundErr *LagoonNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected LagoonNotFoundError, got %T", err)
	}
}

// --- Alternative Domains Tests ---

func TestAddRouteAlternativeDomains(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addRouteAlternativeDomains") {
			t.Errorf("expected addRouteAlternativeDomains mutation")
		}
		return map[string]any{
			"addRouteAlternativeDomains": map[string]any{
				"id": 100,
				"alternativeNames": []map[string]any{
					{"id": 1, "domain": "alt1.example.com"},
					{"id": 2, "domain": "alt2.example.com"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.AddRouteAlternativeDomains(context.Background(), 100, []string{"alt1.example.com", "alt2.example.com"})
	if err != nil {
		t.Fatalf("AddRouteAlternativeDomains failed: %v", err)
	}
	if len(r.AlternativeNames) != 2 {
		t.Errorf("expected 2 alternative names, got %d", len(r.AlternativeNames))
	}
}

func TestRemoveRouteAlternativeDomain(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "removeRouteAlternativeDomain") {
			t.Errorf("expected removeRouteAlternativeDomain mutation")
		}
		return map[string]any{
			"removeRouteAlternativeDomain": map[string]any{
				"id":               100,
				"alternativeNames": []map[string]any{},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.RemoveRouteAlternativeDomain(context.Background(), 100, "alt1.example.com")
	if err != nil {
		t.Fatalf("RemoveRouteAlternativeDomain failed: %v", err)
	}
	if len(r.AlternativeNames) != 0 {
		t.Errorf("expected 0 alternative names, got %d", len(r.AlternativeNames))
	}
}

// --- Annotations Tests ---

func TestAddRouteAnnotations(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addRouteAnnotation") {
			t.Errorf("expected addRouteAnnotation mutation")
		}
		return map[string]any{
			"addRouteAnnotation": map[string]any{
				"id": 100,
				"annotations": []map[string]any{
					{"key": "key1", "value": "value1"},
					{"key": "key2", "value": "value2"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	annotations := []map[string]string{
		{"key": "key1", "value": "value1"},
		{"key": "key2", "value": "value2"},
	}
	r, err := c.AddRouteAnnotations(context.Background(), 100, annotations)
	if err != nil {
		t.Fatalf("AddRouteAnnotations failed: %v", err)
	}
	if len(r.Annotations) != 2 {
		t.Errorf("expected 2 annotations, got %d", len(r.Annotations))
	}
}

func TestRemoveRouteAnnotation(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "removeRouteAnnotation") {
			t.Errorf("expected removeRouteAnnotation mutation")
		}
		return map[string]any{
			"removeRouteAnnotation": map[string]any{
				"id":          100,
				"annotations": []map[string]any{},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.RemoveRouteAnnotation(context.Background(), 100, "key1")
	if err != nil {
		t.Fatalf("RemoveRouteAnnotation failed: %v", err)
	}
	if len(r.Annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(r.Annotations))
	}
}

// --- Path Routes Tests ---

func TestAddPathRoutesToRoute(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addPathRoutesToRoute") {
			t.Errorf("expected addPathRoutesToRoute mutation")
		}
		return map[string]any{
			"addPathRoutesToRoute": map[string]any{
				"id": 100,
				"pathRoutes": []map[string]any{
					{"id": 1, "path": "/api", "toService": "api-service"},
					{"id": 2, "path": "/admin", "toService": "admin-service"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	pathRoutes := []map[string]string{
		{"path": "/api", "toService": "api-service"},
		{"path": "/admin", "toService": "admin-service"},
	}
	r, err := c.AddPathRoutesToRoute(context.Background(), 100, pathRoutes)
	if err != nil {
		t.Fatalf("AddPathRoutesToRoute failed: %v", err)
	}
	if len(r.PathRoutes) != 2 {
		t.Errorf("expected 2 path routes, got %d", len(r.PathRoutes))
	}
}

func TestRemovePathRouteFromRoute(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "removePathRouteFromRoute") {
			t.Errorf("expected removePathRouteFromRoute mutation")
		}
		return map[string]any{
			"removePathRouteFromRoute": map[string]any{
				"id":         100,
				"pathRoutes": []map[string]any{},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	r, err := c.RemovePathRouteFromRoute(context.Background(), 100, "api-service", "/api")
	if err != nil {
		t.Fatalf("RemovePathRouteFromRoute failed: %v", err)
	}
	if len(r.PathRoutes) != 0 {
		t.Errorf("expected 0 path routes, got %d", len(r.PathRoutes))
	}
}

// --- normalizeRoute Tests ---

func TestNormalizeRoute_WithProjectAndEnvironment(t *testing.T) {
	raw := routeRaw{
		ID:      1,
		Domain:  "example.com",
		Project: json.RawMessage(`{"name": "test-project"}`),
		Environment: json.RawMessage(`{"name": "main"}`),
	}
	r := normalizeRoute(raw)
	if r.ProjectName != "test-project" {
		t.Errorf("expected ProjectName=test-project, got %s", r.ProjectName)
	}
	if r.EnvironmentName != "main" {
		t.Errorf("expected EnvironmentName=main, got %s", r.EnvironmentName)
	}
}

func TestNormalizeRoute_NullProject(t *testing.T) {
	raw := routeRaw{
		ID:      1,
		Domain:  "example.com",
		Project: json.RawMessage(`null`),
	}
	r := normalizeRoute(raw)
	if r.ProjectName != "" {
		t.Errorf("expected empty ProjectName, got %s", r.ProjectName)
	}
}

func TestNormalizeRoute_EmptyProject(t *testing.T) {
	raw := routeRaw{
		ID:     1,
		Domain: "example.com",
	}
	r := normalizeRoute(raw)
	if r.ProjectName != "" {
		t.Errorf("expected empty ProjectName, got %s", r.ProjectName)
	}
}

func TestNormalizeRoute_AllFields(t *testing.T) {
	tlsAcme := true
	primary := true
	disableReqVerif := false
	hstsEnabled := true
	hstsPreload := true
	hstsIncludeSubdomains := false
	hstsMaxAge := 31536000

	raw := routeRaw{
		ID:                         1,
		Domain:                     "example.com",
		Service:                    "nginx",
		TLSAcme:                    &tlsAcme,
		Insecure:                   "Redirect",
		Primary:                    &primary,
		Source:                     "API",
		Type:                       "STANDARD",
		DisableRequestVerification: &disableReqVerif,
		HSTSEnabled:                &hstsEnabled,
		HSTSPreload:                &hstsPreload,
		HSTSIncludeSubdomains:      &hstsIncludeSubdomains,
		HSTSMaxAge:                 &hstsMaxAge,
		MonitoringPath:             "/health",
		Annotations: []RouteAnnotation{
			{Key: "key1", Value: "value1"},
		},
		AlternativeNames: []AlternativeName{
			{ID: 1, Domain: "alt.example.com"},
		},
		PathRoutes: []RoutePathRoute{
			{ID: 1, Path: "/api", ToService: "api"},
		},
		Created:     "2024-01-01T00:00:00Z",
		Updated:     "2024-01-02T00:00:00Z",
		Project:     json.RawMessage(`{"name": "my-project"}`),
		Environment: json.RawMessage(`{"name": "prod"}`),
	}

	r := normalizeRoute(raw)
	if r.Domain != "example.com" {
		t.Errorf("expected domain=example.com, got %s", r.Domain)
	}
	if r.TLSAcme == nil || !*r.TLSAcme {
		t.Error("expected TLSAcme=true")
	}
	if r.Insecure != "Redirect" {
		t.Errorf("expected Insecure=Redirect, got %s", r.Insecure)
	}
	if r.MonitoringPath != "/health" {
		t.Errorf("expected MonitoringPath=/health, got %s", r.MonitoringPath)
	}
	if len(r.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(r.Annotations))
	}
	if len(r.AlternativeNames) != 1 {
		t.Errorf("expected 1 alternative name, got %d", len(r.AlternativeNames))
	}
	if len(r.PathRoutes) != 1 {
		t.Errorf("expected 1 path route, got %d", len(r.PathRoutes))
	}
	if r.ProjectName != "my-project" {
		t.Errorf("expected ProjectName=my-project, got %s", r.ProjectName)
	}
	if r.EnvironmentName != "prod" {
		t.Errorf("expected EnvironmentName=prod, got %s", r.EnvironmentName)
	}
}
