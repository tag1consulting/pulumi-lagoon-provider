package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestVariableCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, name, value string, projectID int, scope string, envID *int) (*client.Variable, error) {
			if envID != nil {
				t.Errorf("expected nil environmentID for project-level variable")
			}
			return &client.Variable{ID: 5, Name: name, Value: value, Scope: scope}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Create(ctx, infer.CreateRequest[VariableArgs]{
		Inputs: VariableArgs{Name: "MY_VAR", Value: "secret", ProjectID: 1, Scope: "build"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "5" {
		t.Errorf("expected ID '5', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 5 {
		t.Errorf("expected LagoonID 5, got %d", resp.Output.LagoonID)
	}
}

func TestVariableCreate_WithEnvironment(t *testing.T) {
	envID := 42
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, _ string, _ string, _ int, _ string, eid *int) (*client.Variable, error) {
			if eid == nil || *eid != 42 {
				t.Errorf("expected environmentID 42")
			}
			return &client.Variable{ID: 5, Name: "MY_VAR", Value: "val", Scope: "BUILD"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Create(ctx, infer.CreateRequest[VariableArgs]{
		Inputs: VariableArgs{Name: "MY_VAR", Value: "val", ProjectID: 1, Scope: "build", EnvironmentID: &envID},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.Output.EnvironmentID == nil || *resp.Output.EnvironmentID != 42 {
		t.Errorf("expected EnvironmentID 42 in output")
	}
}

func TestVariableCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, _ string, _ string, _ int, _ string, _ *int) (*client.Variable, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Create(ctx, infer.CreateRequest[VariableArgs]{
		Inputs: VariableArgs{Name: "MY_VAR", Value: "val", ProjectID: 1, Scope: "build"},
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

func TestVariableUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, name, value string, _ int, _ string, _ *int) (*client.Variable, error) {
			return &client.Variable{ID: 5, Name: name, Value: value, Scope: "BUILD"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Update(ctx, infer.UpdateRequest[VariableArgs, VariableState]{
		ID:     "5",
		Inputs: VariableArgs{Name: "MY_VAR", Value: "new-value", ProjectID: 1, Scope: "build"},
		State:  VariableState{VariableArgs: VariableArgs{Name: "MY_VAR", Value: "old-value", ProjectID: 1, Scope: "build"}, LagoonID: 5},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.Value != "new-value" {
		t.Errorf("expected new value, got %q", resp.Output.Value)
	}
}

func TestVariableUpdate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, _ string, _ string, _ int, _ string, _ *int) (*client.Variable, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Update(ctx, infer.UpdateRequest[VariableArgs, VariableState]{
		Inputs: VariableArgs{Name: "MY_VAR", Value: "new", ProjectID: 1, Scope: "build"},
		State:  VariableState{VariableArgs: VariableArgs{Name: "MY_VAR"}, LagoonID: 5},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
}

func TestVariableDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteVariableFn: func(_ context.Context, name string, projectID int, envID *int) error {
			if name != "MY_VAR" {
				t.Errorf("expected name 'MY_VAR', got %q", name)
			}
			if projectID != 1 {
				t.Errorf("expected projectID 1, got %d", projectID)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Delete(ctx, infer.DeleteRequest[VariableState]{
		ID:    "5",
		State: VariableState{VariableArgs: VariableArgs{Name: "MY_VAR", ProjectID: 1, Scope: "build"}, LagoonID: 5},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestVariableDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteVariableFn: func(_ context.Context, _ string, _ int, _ *int) error {
			return &client.LagoonNotFoundError{ResourceType: "Variable", Identifier: "MY_VAR"}
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Delete(ctx, infer.DeleteRequest[VariableState]{
		ID:    "5",
		State: VariableState{VariableArgs: VariableArgs{Name: "MY_VAR", ProjectID: 1}, LagoonID: 5},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestVariableDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteVariableFn: func(_ context.Context, _ string, _ int, _ *int) error {
			return fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Delete(ctx, infer.DeleteRequest[VariableState]{
		ID:    "5",
		State: VariableState{VariableArgs: VariableArgs{Name: "MY_VAR", ProjectID: 1}, LagoonID: 5},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVariableRead_HappyPath(t *testing.T) {
	envID := 42
	mock := &mockLagoonClient{
		getVariableFn: func(_ context.Context, name string, projectID int, envIDPtr *int) (*client.Variable, error) {
			if name != "MY_VAR" {
				t.Errorf("expected name 'MY_VAR', got %q", name)
			}
			if projectID != 1 {
				t.Errorf("expected projectID 1, got %d", projectID)
			}
			if envIDPtr == nil || *envIDPtr != 42 {
				t.Errorf("expected envID 42")
			}
			return &client.Variable{ID: 5, Name: "MY_VAR", Value: "secret", Scope: "BUILD"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{ID: "1:42:MY_VAR"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "5" {
		t.Errorf("expected ID '5', got %q", resp.ID)
	}
	if resp.Inputs.Scope != "build" {
		t.Errorf("expected lowercased scope 'build', got %q", resp.Inputs.Scope)
	}
	if resp.Inputs.EnvironmentID == nil || *resp.Inputs.EnvironmentID != envID {
		t.Errorf("expected EnvironmentID 42")
	}
}

func TestVariableRead_ProjectLevelID(t *testing.T) {
	mock := &mockLagoonClient{
		getVariableFn: func(_ context.Context, name string, projectID int, envID *int) (*client.Variable, error) {
			if envID != nil {
				t.Errorf("expected nil environmentID for project-level variable")
			}
			return &client.Variable{ID: 5, Name: "MY_VAR", Value: "val", Scope: "RUNTIME"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{ID: "1::MY_VAR"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.Inputs.EnvironmentID != nil {
		t.Errorf("expected nil EnvironmentID for project-level variable")
	}
}

func TestVariableRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getVariableFn: func(_ context.Context, _ string, _ int, _ *int) (*client.Variable, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Variable", Identifier: "MY_VAR"}
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{ID: "1::MY_VAR"})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestVariableRead_InvalidID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{ID: "bad"})
	if err == nil {
		t.Fatal("expected error for invalid import ID")
	}
}

func TestVariableRead_InvalidProjectID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{ID: "abc:42:MY_VAR"})
	if err == nil {
		t.Fatal("expected error for non-numeric project_id")
	}
}

func TestVariableRead_InvalidEnvID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{ID: "1:abc:MY_VAR"})
	if err == nil {
		t.Fatal("expected error for non-numeric env_id")
	}
}

func TestVariableCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, _ string, _ string, _ int, _ string, _ *int) (*client.Variable, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Create(ctx, infer.CreateRequest[VariableArgs]{
		Inputs: VariableArgs{Name: "MY_VAR", Value: "secret", ProjectID: 1, Scope: "build"},
	})
	if err == nil {
		t.Fatal("expected error when create API fails")
	}
}

func TestVariableUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		addVariableFn: func(_ context.Context, _ string, _ string, _ int, _ string, _ *int) (*client.Variable, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	_, err := r.Update(ctx, infer.UpdateRequest[VariableArgs, VariableState]{
		ID:     "5",
		Inputs: VariableArgs{Name: "MY_VAR", Value: "new-value", ProjectID: 1, Scope: "build"},
		State:  VariableState{VariableArgs: VariableArgs{Name: "MY_VAR", Value: "old-value", ProjectID: 1, Scope: "build"}, LagoonID: 5},
	})
	if err == nil {
		t.Fatal("expected error when update API fails")
	}
}

func TestVariableRead_RefreshFromState(t *testing.T) {
	mock := &mockLagoonClient{
		getVariableFn: func(_ context.Context, name string, _ int, _ *int) (*client.Variable, error) {
			return &client.Variable{ID: 5, Name: name, Value: "val", Scope: "BUILD"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Variable{}
	resp, err := r.Read(ctx, infer.ReadRequest[VariableArgs, VariableState]{
		ID:    "5",
		State: VariableState{VariableArgs: VariableArgs{Name: "MY_VAR", ProjectID: 1, Scope: "build"}, LagoonID: 5},
	})
	if err != nil {
		t.Fatalf("Read from state failed: %v", err)
	}
	if resp.Inputs.Name != "MY_VAR" {
		t.Errorf("expected Name 'MY_VAR', got %q", resp.Inputs.Name)
	}
}
