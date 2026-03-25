package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestEnvironmentCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, input map[string]any) (*client.Environment, error) {
			if input["deployType"] != "BRANCH" {
				t.Errorf("expected deployType uppercased to BRANCH, got %v", input["deployType"])
			}
			if input["environmentType"] != "DEVELOPMENT" {
				t.Errorf("expected environmentType uppercased to DEVELOPMENT, got %v", input["environmentType"])
			}
			return &client.Environment{ID: 10, Name: "develop", Route: "https://develop.example.com", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "10" {
		t.Errorf("expected ID '10', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 10 {
		t.Errorf("expected LagoonID 10, got %d", resp.Output.LagoonID)
	}
	if resp.Output.Route != "https://develop.example.com" {
		t.Errorf("expected Route, got %q", resp.Output.Route)
	}
}

func TestEnvironmentCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, _ map[string]any) (*client.Environment, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Create DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
	if resp.ID != "preview-id" {
		t.Errorf("expected 'preview-id', got %q", resp.ID)
	}
}

func TestEnvironmentCreate_InvalidDeployType(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "invalid", EnvironmentType: "development"},
	})
	if err == nil {
		t.Fatal("expected error for invalid deployType")
	}
}

func TestEnvironmentCreate_InvalidEnvironmentType(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "staging"},
	})
	if err == nil {
		t.Fatal("expected error for invalid environmentType")
	}
}

func TestEnvironmentCreate_DefaultDeployBaseRef(t *testing.T) {
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, input map[string]any) (*client.Environment, error) {
			if input["deployBaseRef"] != "develop" {
				t.Errorf("expected deployBaseRef to default to name 'develop', got %v", input["deployBaseRef"])
			}
			return &client.Environment{ID: 10, Name: "develop"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	// Verify the effective deployBaseRef was stored in state
	if resp.Output.DeployBaseRef == nil || *resp.Output.DeployBaseRef != "develop" {
		t.Errorf("expected DeployBaseRef to be 'develop' in state")
	}
}

func TestEnvironmentCreate_ExplicitDeployBaseRef(t *testing.T) {
	ref := "main"
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, input map[string]any) (*client.Environment, error) {
			if input["deployBaseRef"] != "main" {
				t.Errorf("expected deployBaseRef 'main', got %v", input["deployBaseRef"])
			}
			return &client.Environment{ID: 10, Name: "develop"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development", DeployBaseRef: &ref},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestEnvironmentUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, _ map[string]any) (*client.Environment, error) {
			return &client.Environment{ID: 10, Name: "develop", Route: "https://new.example.com"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Update(ctx, infer.UpdateRequest[EnvironmentArgs, EnvironmentState]{
		ID:     "10",
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
		State:  EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"}, LagoonID: 10, Created: "2024-01-01"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.Route != "https://new.example.com" {
		t.Errorf("expected new Route, got %q", resp.Output.Route)
	}
}

func TestEnvironmentUpdate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, _ map[string]any) (*client.Environment, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Update(ctx, infer.UpdateRequest[EnvironmentArgs, EnvironmentState]{
		ID:     "10",
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
		State:  EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop"}, LagoonID: 10, Route: "https://old.example.com", Created: "2024-01-01"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
	if resp.Output.Route != "https://old.example.com" {
		t.Errorf("expected old Route preserved during DryRun")
	}
}

func TestEnvironmentDelete_HappyPath(t *testing.T) {
	deletedEnv := ""
	deletedProj := ""
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, id int) (*client.Project, error) {
			return &client.Project{ID: id, Name: "myproject"}, nil
		},
		deleteEnvironmentFn: func(_ context.Context, envName, projectName string) error {
			deletedEnv = envName
			deletedProj = projectName
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[EnvironmentState]{
		ID:    "10",
		State: EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if deletedEnv != "develop" {
		t.Errorf("expected env 'develop', got %q", deletedEnv)
	}
	if deletedProj != "myproject" {
		t.Errorf("expected project 'myproject', got %q", deletedProj)
	}
}

func TestEnvironmentDelete_ProjectNotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, _ int) (*client.Project, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Project", Identifier: "1"}
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[EnvironmentState]{
		ID:    "10",
		State: EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Delete with missing project should succeed: %v", err)
	}
}

func TestEnvironmentDelete_EnvironmentNotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, id int) (*client.Project, error) {
			return &client.Project{ID: id, Name: "myproject"}, nil
		},
		deleteEnvironmentFn: func(_ context.Context, _, _ string) error {
			return &client.LagoonNotFoundError{ResourceType: "Environment", Identifier: "develop"}
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[EnvironmentState]{
		ID:    "10",
		State: EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Delete with missing environment should succeed: %v", err)
	}
}

func TestEnvironmentRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getEnvironmentByNameFn: func(_ context.Context, name string, projectID int) (*client.Environment, error) {
			return &client.Environment{
				ID:              10,
				Name:            name,
				ProjectID:       projectID,
				DeployType:      "BRANCH",
				EnvironmentType: "DEVELOPMENT",
				DeployBaseRef:   "develop",
				Route:           "https://develop.example.com",
				Created:         "2024-01-01",
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Read(ctx, infer.ReadRequest[EnvironmentArgs, EnvironmentState]{ID: "1:develop"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "10" {
		t.Errorf("expected ID '10', got %q", resp.ID)
	}
	if resp.Inputs.DeployType != "branch" {
		t.Errorf("expected lowercased deployType 'branch', got %q", resp.Inputs.DeployType)
	}
	if resp.Inputs.ProjectID != 1 {
		t.Errorf("expected ProjectID 1, got %d", resp.Inputs.ProjectID)
	}
}

func TestEnvironmentRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getEnvironmentByNameFn: func(_ context.Context, _ string, _ int) (*client.Environment, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Environment", Identifier: "test"}
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Read(ctx, infer.ReadRequest[EnvironmentArgs, EnvironmentState]{ID: "1:develop"})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for deleted resource, got %q", resp.ID)
	}
}

func TestEnvironmentRead_InvalidImportID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Read(ctx, infer.ReadRequest[EnvironmentArgs, EnvironmentState]{ID: "notanumber:develop"})
	if err == nil {
		t.Fatal("expected error for invalid project_id in import ID")
	}
}

func TestEnvironmentRead_RefreshFromState(t *testing.T) {
	mock := &mockLagoonClient{
		getEnvironmentByNameFn: func(_ context.Context, name string, projectID int) (*client.Environment, error) {
			return &client.Environment{ID: 10, Name: name, ProjectID: projectID, DeployType: "BRANCH", EnvironmentType: "DEVELOPMENT"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	resp, err := r.Read(ctx, infer.ReadRequest[EnvironmentArgs, EnvironmentState]{
		ID:    "10",
		State: EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1}},
	})
	if err != nil {
		t.Fatalf("Read from state failed: %v", err)
	}
	if resp.Inputs.Name != "develop" {
		t.Errorf("expected Name 'develop', got %q", resp.Inputs.Name)
	}
}

func TestEnvironmentCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, _ map[string]any) (*client.Environment, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Create(ctx, infer.CreateRequest[EnvironmentArgs]{
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnvironmentUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		addOrUpdateEnvironmentFn: func(_ context.Context, _ map[string]any) (*client.Environment, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Update(ctx, infer.UpdateRequest[EnvironmentArgs, EnvironmentState]{
		ID:     "10",
		Inputs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"},
		State:  EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"}, LagoonID: 10},
	})
	if err == nil {
		t.Fatal("expected error when update API fails")
	}
}

func TestEnvironmentDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, id int) (*client.Project, error) {
			return &client.Project{ID: id, Name: "myproject"}, nil
		},
		deleteEnvironmentFn: func(_ context.Context, _, _ string) error {
			return fmt.Errorf("delete failed")
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[EnvironmentState]{
		ID:    "10",
		State: EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1}, LagoonID: 10},
	})
	if err == nil {
		t.Fatal("expected error when delete API fails")
	}
}

func TestEnvironmentDelete_ProjectLookupError(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, _ int) (*client.Project, error) {
			return nil, fmt.Errorf("project lookup failed")
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[EnvironmentState]{
		ID:    "10",
		State: EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "develop", ProjectID: 1}, LagoonID: 10},
	})
	if err == nil {
		t.Fatal("expected error when project lookup fails")
	}
}

func TestEnvironmentRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getEnvironmentByNameFn: func(_ context.Context, _ string, _ int) (*client.Environment, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &Environment{}
	_, err := r.Read(ctx, infer.ReadRequest[EnvironmentArgs, EnvironmentState]{ID: "1:develop"})
	if err == nil {
		t.Fatal("expected error when read API fails")
	}
}
