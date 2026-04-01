package resources

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// --- Create tests ---

func TestRouteCreate_HappyPath(t *testing.T) {
	svc := "nginx"
	mock := &mockLagoonClient{
		addRouteToProjectFn: func(_ context.Context, input map[string]any) (*client.Route, error) {
			if input["domain"] != "example.com" {
				t.Errorf("expected domain example.com, got %v", input["domain"])
			}
			if input["project"] != "my-project" {
				t.Errorf("expected project my-project, got %v", input["project"])
			}
			return &client.Route{ID: 42, Domain: "example.com", ProjectName: "my-project", Source: "API", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		Inputs: RouteArgs{
			ProjectName: "my-project",
			Domain:      "example.com",
			Service:     &svc,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "42" {
		t.Errorf("expected ID 42, got %s", resp.ID)
	}
	if resp.Output.LagoonID != 42 {
		t.Errorf("expected LagoonID 42, got %d", resp.Output.LagoonID)
	}
	if resp.Output.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", resp.Output.Domain)
	}
}

func TestRouteCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addRouteToProjectFn: func(_ context.Context, _ map[string]any) (*client.Route, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		DryRun: true,
		Inputs: RouteArgs{ProjectName: "my-project", Domain: "example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("API should not be called on dry run")
	}
	if resp.ID != "preview-id" {
		t.Errorf("expected preview-id, got %s", resp.ID)
	}
}

func TestRouteCreate_WithMonitoringPath(t *testing.T) {
	updateCalled := false
	mp := "/health"
	mock := &mockLagoonClient{
		addRouteToProjectFn: func(_ context.Context, _ map[string]any) (*client.Route, error) {
			return &client.Route{ID: 5, Domain: "example.com", ProjectName: "p"}, nil
		},
		updateRouteOnProjectFn: func(_ context.Context, routeID int, patch map[string]any) (*client.Route, error) {
			updateCalled = true
			p, ok := patch["patch"].(map[string]any)
			if !ok {
				t.Error("patch missing 'patch' key")
			} else if p["monitoringPath"] != "/health" {
				t.Errorf("expected monitoringPath /health in patch, got %v", p["monitoringPath"])
			}
			return &client.Route{ID: routeID}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", MonitoringPath: &mp},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdateRouteOnProject to be called for monitoringPath")
	}
	if resp.Output.MonitoringPath == nil || *resp.Output.MonitoringPath != "/health" {
		t.Errorf("expected monitoringPath /health in final state, got %v", resp.Output.MonitoringPath)
	}
}

func TestRouteCreate_MonitoringPathUpdateFails(t *testing.T) {
	updateCalled := false
	mp := "/health"
	mock := &mockLagoonClient{
		addRouteToProjectFn: func(_ context.Context, _ map[string]any) (*client.Route, error) {
			return &client.Route{ID: 5, Domain: "example.com", ProjectName: "p", Source: "API", Created: "2024-01-01"}, nil
		},
		updateRouteOnProjectFn: func(_ context.Context, _ int, _ map[string]any) (*client.Route, error) {
			updateCalled = true
			return nil, fmt.Errorf("API error: internal server error")
		},
	}

	// Capture slog output (GetLogger falls back to slog.Default in tests)
	var logBuf bytes.Buffer
	oldDefault := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	defer slog.SetDefault(oldDefault)

	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", MonitoringPath: &mp},
	})
	if err != nil {
		t.Fatalf("expected create to succeed despite monitoringPath failure, got: %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdateRouteOnProject to be called for monitoringPath")
	}
	if resp.Output.MonitoringPath != nil {
		t.Errorf("expected monitoringPath to be nil in state, got %v", *resp.Output.MonitoringPath)
	}
	if resp.Output.LagoonID != 5 {
		t.Errorf("expected LagoonID 5, got %d", resp.Output.LagoonID)
	}
	if resp.ID != "5" {
		t.Errorf("expected ID '5', got %s", resp.ID)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "failed to set monitoringPath") {
		t.Errorf("expected warning about monitoringPath failure in log output, got: %s", logOutput)
	}
}

func TestRouteCreate_WithEnvironment(t *testing.T) {
	env := "main"
	mock := &mockLagoonClient{
		addRouteToProjectFn: func(_ context.Context, input map[string]any) (*client.Route, error) {
			if input["environment"] != "main" {
				t.Errorf("expected environment main in create input, got %v", input["environment"])
			}
			return &client.Route{ID: 10, Domain: "example.com", ProjectName: "p", EnvironmentName: "main"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", Environment: &env},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output.Environment == nil || *resp.Output.Environment != "main" {
		t.Errorf("expected environment main in state, got %v", resp.Output.Environment)
	}
}

func TestRouteCreate_InvalidInsecure(t *testing.T) {
	bad := "BadValue"
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Route{}
	_, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", Insecure: &bad},
	})
	if err == nil {
		t.Error("expected validation error for invalid insecure value")
	}
}

func TestRouteCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		addRouteToProjectFn: func(_ context.Context, _ map[string]any) (*client.Route, error) {
			return nil, &client.LagoonAPIError{Message: "forbidden"}
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	_, err := r.Create(ctx, infer.CreateRequest[RouteArgs]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com"},
	})
	if err == nil {
		t.Error("expected error from API")
	}
}

// --- Update tests ---

func TestRouteUpdate_ScalarFields(t *testing.T) {
	updateCalled := false
	tlsAcme := false
	mock := &mockLagoonClient{
		updateRouteOnProjectFn: func(_ context.Context, _ int, patch map[string]any) (*client.Route, error) {
			updateCalled = true
			return &client.Route{ID: 1, Domain: "example.com"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	oldTLS := true
	_, err := r.Update(ctx, infer.UpdateRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", TLSAcme: &tlsAcme},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com", TLSAcme: &oldTLS}, LagoonID: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdateRouteOnProject to be called")
	}
}

func TestRouteUpdate_AnnotationReconciliation(t *testing.T) {
	removeCalled := false
	addCalled := false
	mock := &mockLagoonClient{
		removeRouteAnnotationFn: func(_ context.Context, _ int, key string) (*client.Route, error) {
			removeCalled = true
			if key != "old-key" {
				t.Errorf("expected old-key, got %s", key)
			}
			return &client.Route{ID: 1}, nil
		},
		addRouteAnnotationsFn: func(_ context.Context, _ int, annotations []map[string]string) (*client.Route, error) {
			addCalled = true
			if len(annotations) != 1 || annotations[0]["key"] != "new-key" {
				t.Errorf("unexpected annotations: %v", annotations)
			}
			return &client.Route{ID: 1}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}

	oldAnns := []RouteAnnotationInput{{Key: "old-key", Value: "old-val"}}
	newAnns := []RouteAnnotationInput{{Key: "new-key", Value: "new-val"}}

	_, err := r.Update(ctx, infer.UpdateRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", Annotations: &newAnns},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com", Annotations: &oldAnns}, LagoonID: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removeCalled {
		t.Error("expected RemoveRouteAnnotation to be called")
	}
	if !addCalled {
		t.Error("expected AddRouteAnnotations to be called")
	}
}

func TestRouteUpdate_AlternativeNamesReconciliation(t *testing.T) {
	removeCalled := false
	addCalled := false
	mock := &mockLagoonClient{
		removeRouteAlternativeDomainFn: func(_ context.Context, _ int, domain string) (*client.Route, error) {
			removeCalled = true
			return &client.Route{ID: 1}, nil
		},
		addRouteAlternativeDomainsFn: func(_ context.Context, _ int, domains []string) (*client.Route, error) {
			addCalled = true
			return &client.Route{ID: 1}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}

	oldNames := []string{"old.example.com"}
	newNames := []string{"new.example.com"}

	_, err := r.Update(ctx, infer.UpdateRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", AlternativeNames: &newNames},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com", AlternativeNames: &oldNames}, LagoonID: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !removeCalled {
		t.Error("expected RemoveRouteAlternativeDomain to be called")
	}
	if !addCalled {
		t.Error("expected AddRouteAlternativeDomains to be called")
	}
}

func TestRouteUpdate_EnvironmentDetach(t *testing.T) {
	detachCalled := false
	mock := &mockLagoonClient{
		removeRouteFromEnvironmentFn: func(_ context.Context, domain, project, environment string) error {
			detachCalled = true
			if environment != "main" {
				t.Errorf("expected environment main, got %s", environment)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}

	oldEnv := "main"
	_, err := r.Update(ctx, infer.UpdateRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com"},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com", Environment: &oldEnv}, LagoonID: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !detachCalled {
		t.Error("expected RemoveRouteFromEnvironment to be called when clearing environment")
	}
}

func TestRouteUpdate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		updateRouteOnProjectFn: func(_ context.Context, _ int, _ map[string]any) (*client.Route, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	tlsFalse := false
	tlsTrue := true
	_, err := r.Update(ctx, infer.UpdateRequest[RouteArgs, RouteState]{
		DryRun: true,
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com", TLSAcme: &tlsFalse},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com", TLSAcme: &tlsTrue}, LagoonID: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("API should not be called on dry run")
	}
}

// --- Delete tests ---

func TestRouteDelete_HappyPath(t *testing.T) {
	deleteCalled := false
	mock := &mockLagoonClient{
		deleteRouteFn: func(_ context.Context, routeID int) error {
			deleteCalled = true
			if routeID != 42 {
				t.Errorf("expected routeID 42, got %d", routeID)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	_, err := r.Delete(ctx, infer.DeleteRequest[RouteState]{
		State: RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com"}, LagoonID: 42},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DeleteRoute to be called")
	}
}

func TestRouteDelete_WithEnvironment_DetachesFirst(t *testing.T) {
	callOrder := []string{}
	env := "main"
	mock := &mockLagoonClient{
		removeRouteFromEnvironmentFn: func(_ context.Context, _, _, _ string) error {
			callOrder = append(callOrder, "detach")
			return nil
		},
		deleteRouteFn: func(_ context.Context, _ int) error {
			callOrder = append(callOrder, "delete")
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	_, err := r.Delete(ctx, infer.DeleteRequest[RouteState]{
		State: RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com", Environment: &env}, LagoonID: 5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(callOrder) != 2 || callOrder[0] != "detach" || callOrder[1] != "delete" {
		t.Errorf("expected detach then delete, got %v", callOrder)
	}
}

func TestRouteDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteRouteFn: func(_ context.Context, _ int) error {
			return &client.LagoonNotFoundError{ResourceType: "Route", Identifier: "example.com"}
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	_, err := r.Delete(ctx, infer.DeleteRequest[RouteState]{
		State: RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com"}, LagoonID: 1},
	})
	if err != nil {
		t.Errorf("expected not-found to be tolerated, got: %v", err)
	}
}

// --- Read tests ---

func TestRouteRead_ByState(t *testing.T) {
	mock := &mockLagoonClient{
		getRouteByDomainFn: func(_ context.Context, projectName, domain string) (*client.Route, error) {
			return &client.Route{ID: 7, Domain: domain, ProjectName: projectName, Source: "API"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Read(ctx, infer.ReadRequest[RouteArgs, RouteState]{
		ID:    "7",
		State: RouteState{RouteArgs: RouteArgs{ProjectName: "my-proj", Domain: "example.com"}, LagoonID: 7},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "7" {
		t.Errorf("expected ID 7, got %s", resp.ID)
	}
}

func TestRouteRead_ByImportID(t *testing.T) {
	mock := &mockLagoonClient{
		getRouteByDomainFn: func(_ context.Context, projectName, domain string) (*client.Route, error) {
			if projectName != "my-proj" || domain != "example.com" {
				return nil, fmt.Errorf("unexpected lookup: %s:%s", projectName, domain)
			}
			return &client.Route{ID: 3, Domain: "example.com", ProjectName: "my-proj", Source: "API"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Read(ctx, infer.ReadRequest[RouteArgs, RouteState]{
		ID:    "my-proj:example.com",
		State: RouteState{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Inputs.ProjectName != "my-proj" {
		t.Errorf("expected projectName my-proj, got %s", resp.Inputs.ProjectName)
	}
}

func TestRouteRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getRouteByDomainFn: func(_ context.Context, _, _ string) (*client.Route, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Route", Identifier: "example.com"}
		},
	}
	ctx := testCtx(mock)
	r := &Route{}
	resp, err := r.Read(ctx, infer.ReadRequest[RouteArgs, RouteState]{
		ID:    "7",
		State: RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com"}, LagoonID: 7},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Error("expected empty ID for not-found route")
	}
}

// --- Diff tests ---

func TestRouteDiff_DomainIsForceNew(t *testing.T) {
	r := &Route{}
	resp, err := r.Diff(testCtx(&mockLagoonClient{}), infer.DiffRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "new.com"},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "old.com"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected HasChanges")
	}
	if d, ok := resp.DetailedDiff["domain"]; !ok || d.Kind != p.UpdateReplace { // UpdateReplace = 3
		t.Errorf("expected domain to be UpdateReplace, got %v", resp.DetailedDiff)
	}
}

func TestRouteDiff_ProjectNameIsForceNew(t *testing.T) {
	r := &Route{}
	resp, err := r.Diff(testCtx(&mockLagoonClient{}), infer.DiffRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "new-proj", Domain: "example.com"},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "old-proj", Domain: "example.com"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d, ok := resp.DetailedDiff["projectName"]; !ok || d.Kind != p.UpdateReplace {
		t.Errorf("expected projectName to be UpdateReplace, got %v", resp.DetailedDiff)
	}
}

func TestRouteDiff_NoChanges(t *testing.T) {
	r := &Route{}
	resp, err := r.Diff(testCtx(&mockLagoonClient{}), infer.DiffRequest[RouteArgs, RouteState]{
		Inputs: RouteArgs{ProjectName: "p", Domain: "example.com"},
		State:  RouteState{RouteArgs: RouteArgs{ProjectName: "p", Domain: "example.com"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HasChanges {
		t.Errorf("expected no changes, got: %v", resp.DetailedDiff)
	}
}
