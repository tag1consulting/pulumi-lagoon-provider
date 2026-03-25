package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestDeployTargetCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createDeployTargetFn: func(_ context.Context, input map[string]any) (*client.DeployTarget, error) {
			// Verify defaults are applied
			if input["cloudProvider"] != "kind" {
				t.Errorf("expected default cloudProvider 'kind', got %v", input["cloudProvider"])
			}
			if input["cloudRegion"] != "local" {
				t.Errorf("expected default cloudRegion 'local', got %v", input["cloudRegion"])
			}
			return &client.DeployTarget{ID: 7, Name: "mycluster", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetArgs]{
		Inputs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://k8s.example.com"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "7" {
		t.Errorf("expected ID '7', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 7 {
		t.Errorf("expected LagoonID 7, got %d", resp.Output.LagoonID)
	}
	// Verify defaults were stored in state
	if resp.Output.CloudProvider == nil || *resp.Output.CloudProvider != "kind" {
		t.Errorf("expected CloudProvider 'kind' in state")
	}
	if resp.Output.CloudRegion == nil || *resp.Output.CloudRegion != "local" {
		t.Errorf("expected CloudRegion 'local' in state")
	}
}

func TestDeployTargetCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		createDeployTargetFn: func(_ context.Context, _ map[string]any) (*client.DeployTarget, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetArgs]{
		Inputs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://k8s.example.com"},
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

func TestDeployTargetCreate_DuplicateAdopts(t *testing.T) {
	mock := &mockLagoonClient{
		createDeployTargetFn: func(_ context.Context, _ map[string]any) (*client.DeployTarget, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry"}
		},
		getDeployTargetByNameFn: func(_ context.Context, name string) (*client.DeployTarget, error) {
			return &client.DeployTarget{ID: 99, Name: name}, nil
		},
		updateDeployTargetFn: func(_ context.Context, id int, _ map[string]any) (*client.DeployTarget, error) {
			return &client.DeployTarget{ID: id, Name: "mycluster", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetArgs]{
		Inputs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://k8s.example.com"},
	})
	if err != nil {
		t.Fatalf("Create with duplicate should adopt: %v", err)
	}
	if resp.ID != "99" {
		t.Errorf("expected adopted ID '99', got %q", resp.ID)
	}
}

func TestDeployTargetCreate_ExplicitCloudProvider(t *testing.T) {
	cp := "aws"
	cr := "us-east-1"
	mock := &mockLagoonClient{
		createDeployTargetFn: func(_ context.Context, input map[string]any) (*client.DeployTarget, error) {
			if input["cloudProvider"] != "aws" {
				t.Errorf("expected cloudProvider 'aws', got %v", input["cloudProvider"])
			}
			if input["cloudRegion"] != "us-east-1" {
				t.Errorf("expected cloudRegion 'us-east-1', got %v", input["cloudRegion"])
			}
			return &client.DeployTarget{ID: 7, Name: "mycluster"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Create(ctx, infer.CreateRequest[DeployTargetArgs]{
		Inputs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://k8s.example.com", CloudProvider: &cp, CloudRegion: &cr},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.Output.CloudProvider == nil || *resp.Output.CloudProvider != "aws" {
		t.Errorf("expected CloudProvider 'aws' in state")
	}
}

func TestDeployTargetUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		updateDeployTargetFn: func(_ context.Context, id int, _ map[string]any) (*client.DeployTarget, error) {
			return &client.DeployTarget{ID: id, Name: "mycluster"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Update(ctx, infer.UpdateRequest[DeployTargetArgs, DeployTargetState]{
		ID:     "7",
		Inputs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://new.k8s.example.com"},
		State:  DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://old.k8s.example.com"}, LagoonID: 7, Created: "2024-01-01"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.ConsoleURL != "https://new.k8s.example.com" {
		t.Errorf("expected updated ConsoleURL")
	}
	if resp.Output.Created != "2024-01-01" {
		t.Errorf("expected Created preserved")
	}
}

func TestDeployTargetUpdate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		updateDeployTargetFn: func(_ context.Context, _ int, _ map[string]any) (*client.DeployTarget, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	_, err := r.Update(ctx, infer.UpdateRequest[DeployTargetArgs, DeployTargetState]{
		Inputs: DeployTargetArgs{Name: "mycluster", ConsoleURL: "https://k8s.example.com"},
		State:  DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "mycluster"}, LagoonID: 7},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
}

func TestDeployTargetDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteDeployTargetFn: func(_ context.Context, name string) error {
			if name != "mycluster" {
				t.Errorf("expected name 'mycluster', got %q", name)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	_, err := r.Delete(ctx, infer.DeleteRequest[DeployTargetState]{
		ID:    "7",
		State: DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "mycluster"}, LagoonID: 7},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestDeployTargetDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteDeployTargetFn: func(_ context.Context, _ string) error {
			return &client.LagoonNotFoundError{ResourceType: "DeployTarget", Identifier: "mycluster"}
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	_, err := r.Delete(ctx, infer.DeleteRequest[DeployTargetState]{
		ID:    "7",
		State: DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "mycluster"}, LagoonID: 7},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestDeployTargetDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteDeployTargetFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("forbidden")
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	_, err := r.Delete(ctx, infer.DeleteRequest[DeployTargetState]{
		ID:    "7",
		State: DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "mycluster"}, LagoonID: 7},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeployTargetRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getDeployTargetByIDFn: func(_ context.Context, id int) (*client.DeployTarget, error) {
			return &client.DeployTarget{
				ID:            id,
				Name:          "mycluster",
				ConsoleURL:    "https://k8s.example.com",
				CloudProvider: "aws",
				CloudRegion:   "us-east-1",
				SSHHost:       "ssh.example.com",
				SSHPort:       "22",
				Disabled:      false,
				Created:       "2024-01-01",
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Read(ctx, infer.ReadRequest[DeployTargetArgs, DeployTargetState]{ID: "7"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "7" {
		t.Errorf("expected ID '7', got %q", resp.ID)
	}
	if resp.Inputs.Name != "mycluster" {
		t.Errorf("expected Name 'mycluster', got %q", resp.Inputs.Name)
	}
	if resp.Inputs.CloudProvider == nil || *resp.Inputs.CloudProvider != "aws" {
		t.Errorf("expected CloudProvider 'aws'")
	}
	if resp.Inputs.SSHHost == nil || *resp.Inputs.SSHHost != "ssh.example.com" {
		t.Errorf("expected SSHHost")
	}
	if resp.Inputs.Disabled == nil || *resp.Inputs.Disabled != false {
		t.Errorf("expected Disabled false")
	}
}

func TestDeployTargetRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getDeployTargetByIDFn: func(_ context.Context, _ int) (*client.DeployTarget, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "DeployTarget", Identifier: "7"}
		},
	}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	resp, err := r.Read(ctx, infer.ReadRequest[DeployTargetArgs, DeployTargetState]{ID: "7"})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestDeployTargetRead_InvalidID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &DeployTarget{}
	_, err := r.Read(ctx, infer.ReadRequest[DeployTargetArgs, DeployTargetState]{ID: "not-a-number"})
	if err == nil {
		t.Fatal("expected error for non-numeric ID")
	}
}
