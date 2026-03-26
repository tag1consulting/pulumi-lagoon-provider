package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestGroupCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createGroupFn: func(_ context.Context, name string, _ *string) (*client.Group, error) {
			return &client.Group{ID: "uuid-42", Name: name}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	resp, err := r.Create(ctx, infer.CreateRequest[GroupArgs]{
		Inputs: GroupArgs{Name: "my-team"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "uuid-42" {
		t.Errorf("expected ID 'uuid-42', got %q", resp.ID)
	}
	if resp.Output.LagoonID != "uuid-42" {
		t.Errorf("expected LagoonID 'uuid-42', got %q", resp.Output.LagoonID)
	}
	if resp.Output.Name != "my-team" {
		t.Errorf("expected Name 'my-team', got %q", resp.Output.Name)
	}
}

func TestGroupCreate_WithParent(t *testing.T) {
	parent := "parent-group"
	var gotParent *string
	mock := &mockLagoonClient{
		createGroupFn: func(_ context.Context, name string, parentGroupName *string) (*client.Group, error) {
			gotParent = parentGroupName
			return &client.Group{ID: "uuid-43", Name: name}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Create(ctx, infer.CreateRequest[GroupArgs]{
		Inputs: GroupArgs{Name: "sub-team", ParentGroupName: &parent},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if gotParent == nil || *gotParent != "parent-group" {
		t.Errorf("expected parent group name 'parent-group' to be passed")
	}
}

func TestGroupCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		createGroupFn: func(_ context.Context, _ string, _ *string) (*client.Group, error) {
			called = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	resp, err := r.Create(ctx, infer.CreateRequest[GroupArgs]{
		Inputs: GroupArgs{Name: "my-team"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Create DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
	if resp.ID != "preview-id" {
		t.Errorf("expected ID 'preview-id', got %q", resp.ID)
	}
	if resp.Output.Name != "my-team" {
		t.Errorf("expected Name preserved in output, got %q", resp.Output.Name)
	}
}

func TestGroupCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		createGroupFn: func(_ context.Context, _ string, _ *string) (*client.Group, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Create(ctx, infer.CreateRequest[GroupArgs]{
		Inputs: GroupArgs{Name: "my-team"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getGroupByNameFn: func(_ context.Context, name string) (*client.Group, error) {
			return &client.Group{ID: "uuid-42", Name: name}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	resp, err := r.Read(ctx, infer.ReadRequest[GroupArgs, GroupState]{
		ID:    "uuid-42",
		State: GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "uuid-42" {
		t.Errorf("expected ID 'uuid-42', got %q", resp.ID)
	}
	if resp.Inputs.Name != "my-team" {
		t.Errorf("expected Name 'my-team', got %q", resp.Inputs.Name)
	}
}

func TestGroupRead_FallbackToID(t *testing.T) {
	// When state.Name is empty (e.g. import), Read should use req.ID as the name.
	mock := &mockLagoonClient{
		getGroupByNameFn: func(_ context.Context, name string) (*client.Group, error) {
			if name != "imported-group" {
				t.Errorf("expected fallback to ID 'imported-group', got %q", name)
			}
			return &client.Group{ID: "uuid-99", Name: name}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Read(ctx, infer.ReadRequest[GroupArgs, GroupState]{
		ID: "imported-group",
		// Empty state — simulates import
	})
	if err != nil {
		t.Fatalf("Read with fallback failed: %v", err)
	}
}

func TestGroupRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getGroupByNameFn: func(_ context.Context, _ string) (*client.Group, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Group", Identifier: "my-team"}
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	resp, err := r.Read(ctx, infer.ReadRequest[GroupArgs, GroupState]{
		ID:    "uuid-42",
		State: GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for deleted resource, got %q", resp.ID)
	}
}

func TestGroupRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getGroupByNameFn: func(_ context.Context, _ string) (*client.Group, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Read(ctx, infer.ReadRequest[GroupArgs, GroupState]{
		ID:    "uuid-42",
		State: GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err == nil {
		t.Fatal("expected error when API returns error")
	}
}

func TestGroupUpdate_HappyPath(t *testing.T) {
	newParent := "new-parent"
	oldParent := "old-parent"
	mock := &mockLagoonClient{
		updateGroupFn: func(_ context.Context, name string, patch map[string]any) (*client.Group, error) {
			if _, ok := patch["parentGroup"]; !ok {
				t.Error("expected parentGroup in patch")
			}
			return &client.Group{ID: "uuid-42", Name: name}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	resp, err := r.Update(ctx, infer.UpdateRequest[GroupArgs, GroupState]{
		ID:     "uuid-42",
		Inputs: GroupArgs{Name: "my-team", ParentGroupName: &newParent},
		State:  GroupState{GroupArgs: GroupArgs{Name: "my-team", ParentGroupName: &oldParent}, LagoonID: "uuid-42"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.LagoonID != "uuid-42" {
		t.Errorf("expected LagoonID preserved")
	}
}

func TestGroupUpdate_NoPatch(t *testing.T) {
	// When nothing changed, no API call should be made
	called := false
	mock := &mockLagoonClient{
		updateGroupFn: func(_ context.Context, _ string, _ map[string]any) (*client.Group, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Update(ctx, infer.UpdateRequest[GroupArgs, GroupState]{
		ID:     "uuid-42",
		Inputs: GroupArgs{Name: "my-team"},
		State:  GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err != nil {
		t.Fatalf("Update NoPatch failed: %v", err)
	}
	if called {
		t.Error("API should not be called when nothing changed")
	}
}

func TestGroupUpdate_DryRun(t *testing.T) {
	called := false
	newParent := "new-parent"
	oldParent := "old-parent"
	mock := &mockLagoonClient{
		updateGroupFn: func(_ context.Context, _ string, _ map[string]any) (*client.Group, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	resp, err := r.Update(ctx, infer.UpdateRequest[GroupArgs, GroupState]{
		ID:     "uuid-42",
		Inputs: GroupArgs{Name: "my-team", ParentGroupName: &newParent},
		State:  GroupState{GroupArgs: GroupArgs{Name: "my-team", ParentGroupName: &oldParent}, LagoonID: "uuid-42"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
	if resp.Output.LagoonID != "uuid-42" {
		t.Errorf("expected LagoonID preserved")
	}
}

func TestGroupUpdate_APIError(t *testing.T) {
	newParent := "new-parent"
	oldParent := "old-parent"
	mock := &mockLagoonClient{
		updateGroupFn: func(_ context.Context, _ string, _ map[string]any) (*client.Group, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Update(ctx, infer.UpdateRequest[GroupArgs, GroupState]{
		Inputs: GroupArgs{Name: "my-team", ParentGroupName: &newParent},
		State:  GroupState{GroupArgs: GroupArgs{Name: "my-team", ParentGroupName: &oldParent}, LagoonID: "uuid-42"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroupDelete_HappyPath(t *testing.T) {
	deleted := ""
	mock := &mockLagoonClient{
		deleteGroupFn: func(_ context.Context, name string) error {
			deleted = name
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Delete(ctx, infer.DeleteRequest[GroupState]{
		ID:    "uuid-42",
		State: GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if deleted != "my-team" {
		t.Errorf("expected delete called with 'my-team', got %q", deleted)
	}
}

func TestGroupDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteGroupFn: func(_ context.Context, _ string) error {
			return &client.LagoonNotFoundError{ResourceType: "Group", Identifier: "my-team"}
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Delete(ctx, infer.DeleteRequest[GroupState]{
		ID:    "uuid-42",
		State: GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestGroupDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteGroupFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("forbidden")
		},
	}
	ctx := testCtx(mock)
	r := &Group{}
	_, err := r.Delete(ctx, infer.DeleteRequest[GroupState]{
		ID:    "uuid-42",
		State: GroupState{GroupArgs: GroupArgs{Name: "my-team"}, LagoonID: "uuid-42"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
