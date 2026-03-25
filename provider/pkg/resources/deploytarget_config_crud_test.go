package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestDeployTargetConfigCreate_HappyPath(t *testing.T) {
	branches := "main"
	mock := &mockLagoonClient{
		createDeployTargetConfigFn: func(_ context.Context, input map[string]any) (*client.DeployTargetConfig, error) {
			if input["project"] != 1 {
				t.Errorf("expected project 1, got %v", input["project"])
			}
			if input["deployTarget"] != 7 {
				t.Errorf("expected deployTarget 7, got %v", input["deployTarget"])
			}
			return &client.DeployTargetConfig{ID: 100, ProjectID: 1, DeployTargetID: 7}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetConfigArgs]{
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7, Branches: &branches},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "100" {
		t.Errorf("expected ID '100', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 100 {
		t.Errorf("expected LagoonID 100, got %d", resp.Output.LagoonID)
	}
}

func TestDeployTargetConfigCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		createDeployTargetConfigFn: func(_ context.Context, _ map[string]any) (*client.DeployTargetConfig, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetConfigArgs]{
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7},
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

func TestDeployTargetConfigCreate_DuplicateAdopts(t *testing.T) {
	mock := &mockLagoonClient{
		createDeployTargetConfigFn: func(_ context.Context, _ map[string]any) (*client.DeployTargetConfig, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry"}
		},
		getDeployTargetConfigsByProjectFn: func(_ context.Context, projectID int) ([]client.DeployTargetConfig, error) {
			return []client.DeployTargetConfig{
				{ID: 50, DeployTargetID: 99, ProjectID: projectID},
				{ID: 51, DeployTargetID: 7, ProjectID: projectID},
			}, nil
		},
		updateDeployTargetConfigFn: func(_ context.Context, configID int, _ map[string]any) (*client.DeployTargetConfig, error) {
			if configID != 51 {
				t.Errorf("expected to update config 51, got %d", configID)
			}
			return &client.DeployTargetConfig{ID: 51, ProjectID: 1, DeployTargetID: 7}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetConfigArgs]{
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7},
	})
	if err != nil {
		t.Fatalf("Create with duplicate should adopt: %v", err)
	}
	if resp.ID != "51" {
		t.Errorf("expected adopted ID '51', got %q", resp.ID)
	}
}

func TestDeployTargetConfigCreate_DuplicateNotFoundInList(t *testing.T) {
	mock := &mockLagoonClient{
		createDeployTargetConfigFn: func(_ context.Context, _ map[string]any) (*client.DeployTargetConfig, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry"}
		},
		getDeployTargetConfigsByProjectFn: func(_ context.Context, _ int) ([]client.DeployTargetConfig, error) {
			return []client.DeployTargetConfig{
				{ID: 50, DeployTargetID: 99},
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	_, err := r.Create(ctx, infer.CreateRequest[DeployTargetConfigArgs]{
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7},
	})
	if err == nil {
		t.Fatal("expected error when duplicate exists but matching config not found")
	}
}

func TestDeployTargetConfigUpdate_HappyPath(t *testing.T) {
	branches := "main|develop"
	mock := &mockLagoonClient{
		updateDeployTargetConfigFn: func(_ context.Context, configID int, _ map[string]any) (*client.DeployTargetConfig, error) {
			return &client.DeployTargetConfig{ID: configID, ProjectID: 1, DeployTargetID: 7}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Update(ctx, infer.UpdateRequest[DeployTargetConfigArgs, DeployTargetConfigState]{
		ID:     "100",
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7, Branches: &branches},
		State:  DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7}, LagoonID: 100},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.LagoonID != 100 {
		t.Errorf("expected LagoonID preserved")
	}
}

func TestDeployTargetConfigUpdate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		updateDeployTargetConfigFn: func(_ context.Context, _ int, _ map[string]any) (*client.DeployTargetConfig, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	_, err := r.Update(ctx, infer.UpdateRequest[DeployTargetConfigArgs, DeployTargetConfigState]{
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7},
		State:  DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7}, LagoonID: 100},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
}

func TestDeployTargetConfigDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteDeployTargetConfigFn: func(_ context.Context, configID, projectID int) error {
			if configID != 100 {
				t.Errorf("expected configID 100, got %d", configID)
			}
			if projectID != 1 {
				t.Errorf("expected projectID 1, got %d", projectID)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	_, err := r.Delete(ctx, infer.DeleteRequest[DeployTargetConfigState]{
		ID:    "100",
		State: DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7}, LagoonID: 100},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestDeployTargetConfigDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteDeployTargetConfigFn: func(_ context.Context, _ int, _ int) error {
			return &client.LagoonNotFoundError{ResourceType: "DeployTargetConfig", Identifier: "100"}
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	_, err := r.Delete(ctx, infer.DeleteRequest[DeployTargetConfigState]{
		ID:    "100",
		State: DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1}, LagoonID: 100},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestDeployTargetConfigRead_HappyPath(t *testing.T) {
	branches := "main"
	weight := 10
	mock := &mockLagoonClient{
		getDeployTargetConfigByIDFn: func(_ context.Context, configID, projectID int) (*client.DeployTargetConfig, error) {
			return &client.DeployTargetConfig{
				ID:             configID,
				ProjectID:      projectID,
				DeployTargetID: 7,
				Branches:       branches,
				Weight:         weight,
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Read(ctx, infer.ReadRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1:100"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "100" {
		t.Errorf("expected ID '100', got %q", resp.ID)
	}
	if resp.Inputs.DeployTargetID != 7 {
		t.Errorf("expected DeployTargetID 7, got %d", resp.Inputs.DeployTargetID)
	}
	if resp.Inputs.Branches == nil || *resp.Inputs.Branches != "main" {
		t.Errorf("expected Branches 'main'")
	}
	if resp.Inputs.Weight == nil || *resp.Inputs.Weight != 10 {
		t.Errorf("expected Weight 10")
	}
}

func TestDeployTargetConfigRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getDeployTargetConfigByIDFn: func(_ context.Context, _, _ int) (*client.DeployTargetConfig, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "DeployTargetConfig", Identifier: "100"}
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Read(ctx, infer.ReadRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1:100"})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestDeployTargetConfigRead_InvalidImportID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	_, err := r.Read(ctx, infer.ReadRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "abc:100"})
	if err == nil {
		t.Fatal("expected error for invalid project_id")
	}
}

func TestDeployTargetConfigRead_RefreshFromState(t *testing.T) {
	mock := &mockLagoonClient{
		getDeployTargetConfigByIDFn: func(_ context.Context, configID, projectID int) (*client.DeployTargetConfig, error) {
			return &client.DeployTargetConfig{ID: configID, ProjectID: projectID, DeployTargetID: 7}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	resp, err := r.Read(ctx, infer.ReadRequest[DeployTargetConfigArgs, DeployTargetConfigState]{
		ID:    "100",
		State: DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7}, LagoonID: 100},
	})
	if err != nil {
		t.Fatalf("Read from state failed: %v", err)
	}
	if resp.Inputs.ProjectID != 1 {
		t.Errorf("expected ProjectID 1, got %d", resp.Inputs.ProjectID)
	}
}

func TestDeployTargetConfigUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		updateDeployTargetConfigFn: func(_ context.Context, _ int, _ map[string]any) (*client.DeployTargetConfig, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &DeployTargetConfig{}
	_, err := r.Update(ctx, infer.UpdateRequest[DeployTargetConfigArgs, DeployTargetConfigState]{
		Inputs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7},
		State:  DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 7}, LagoonID: 100},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
