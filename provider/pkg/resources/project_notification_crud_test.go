package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestProjectNotificationCreate_HappyPath(t *testing.T) {
	addCalled := false
	mock := &mockLagoonClient{
		addNotificationToProjectFn: func(_ context.Context, projectName, notifType, notifName string) error {
			addCalled = true
			if projectName != "myproject" {
				t.Errorf("expected projectName 'myproject', got %q", projectName)
			}
			if notifType != "SLACK" {
				t.Errorf("expected uppercased type 'SLACK', got %q", notifType)
			}
			return nil
		},
		checkProjectNotificationExistsFn: func(_ context.Context, _, _, _ string) (*client.ProjectNotificationInfo, error) {
			return &client.ProjectNotificationInfo{ProjectID: 42, Exists: true}, nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	resp, err := r.Create(ctx, infer.CreateRequest[ProjectNotificationArgs]{
		Inputs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !addCalled {
		t.Error("AddNotificationToProject should have been called")
	}
	if resp.ID != "myproject:slack:deploy-slack" {
		t.Errorf("expected composite ID, got %q", resp.ID)
	}
	if resp.Output.ProjectID != 42 {
		t.Errorf("expected ProjectID 42, got %d", resp.Output.ProjectID)
	}
}

func TestProjectNotificationCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addNotificationToProjectFn: func(_ context.Context, _, _, _ string) error {
			called = true
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	resp, err := r.Create(ctx, infer.CreateRequest[ProjectNotificationArgs]{
		Inputs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Create DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
	if resp.ID != "myproject:slack:deploy-slack" {
		t.Errorf("expected composite ID, got %q", resp.ID)
	}
}

func TestProjectNotificationCreate_VerificationFails(t *testing.T) {
	mock := &mockLagoonClient{
		addNotificationToProjectFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
		checkProjectNotificationExistsFn: func(_ context.Context, _, _, _ string) (*client.ProjectNotificationInfo, error) {
			return nil, fmt.Errorf("verification error")
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectNotificationArgs]{
		Inputs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
	})
	if err == nil {
		t.Fatal("expected error when verification fails")
	}
}

func TestProjectNotificationCreate_VerificationNotExists(t *testing.T) {
	mock := &mockLagoonClient{
		addNotificationToProjectFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
		checkProjectNotificationExistsFn: func(_ context.Context, _, _, _ string) (*client.ProjectNotificationInfo, error) {
			return &client.ProjectNotificationInfo{ProjectID: 42, Exists: false}, nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectNotificationArgs]{
		Inputs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
	})
	if err == nil {
		t.Fatal("expected error when notification not verified")
	}
}

func TestProjectNotificationCreate_AddFails(t *testing.T) {
	mock := &mockLagoonClient{
		addNotificationToProjectFn: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("add failed")
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectNotificationArgs]{
		Inputs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProjectNotificationDelete_HappyPath(t *testing.T) {
	removeCalled := false
	mock := &mockLagoonClient{
		removeNotificationFromProjectFn: func(_ context.Context, projectName, notifType, notifName string) error {
			removeCalled = true
			if notifType != "SLACK" {
				t.Errorf("expected uppercased type 'SLACK', got %q", notifType)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Delete(ctx, infer.DeleteRequest[ProjectNotificationState]{
		ID: "myproject:slack:deploy-slack",
		State: ProjectNotificationState{
			ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
			ProjectID:               42,
		},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !removeCalled {
		t.Error("RemoveNotificationFromProject should have been called")
	}
}

func TestProjectNotificationDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		removeNotificationFromProjectFn: func(_ context.Context, _, _, _ string) error {
			return &client.LagoonNotFoundError{ResourceType: "ProjectNotification", Identifier: "test"}
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Delete(ctx, infer.DeleteRequest[ProjectNotificationState]{
		State: ProjectNotificationState{
			ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
		},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestProjectNotificationRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		checkProjectNotificationExistsFn: func(_ context.Context, projectName, notifType, notifName string) (*client.ProjectNotificationInfo, error) {
			return &client.ProjectNotificationInfo{ProjectID: 42, Exists: true}, nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectNotificationArgs, ProjectNotificationState]{
		ID: "myproject:slack:deploy-slack",
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "myproject:slack:deploy-slack" {
		t.Errorf("expected ID preserved, got %q", resp.ID)
	}
	if resp.State.ProjectID != 42 {
		t.Errorf("expected ProjectID 42, got %d", resp.State.ProjectID)
	}
	if resp.Inputs.NotificationType != "slack" {
		t.Errorf("expected notificationType 'slack', got %q", resp.Inputs.NotificationType)
	}
}

func TestProjectNotificationRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		checkProjectNotificationExistsFn: func(_ context.Context, _, _, _ string) (*client.ProjectNotificationInfo, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Project", Identifier: "myproject"}
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectNotificationArgs, ProjectNotificationState]{
		ID: "myproject:slack:deploy-slack",
	})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestProjectNotificationRead_NotificationGone(t *testing.T) {
	mock := &mockLagoonClient{
		checkProjectNotificationExistsFn: func(_ context.Context, _, _, _ string) (*client.ProjectNotificationInfo, error) {
			return &client.ProjectNotificationInfo{ProjectID: 42, Exists: false}, nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectNotificationArgs, ProjectNotificationState]{
		ID: "myproject:slack:deploy-slack",
	})
	if err != nil {
		t.Fatalf("Read should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for vanished notification, got %q", resp.ID)
	}
}

func TestProjectNotificationRead_InvalidImportID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Read(ctx, infer.ReadRequest[ProjectNotificationArgs, ProjectNotificationState]{
		ID: "only-one-part",
	})
	if err == nil {
		t.Fatal("expected error for invalid import ID")
	}
}

func TestProjectNotificationDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		removeNotificationFromProjectFn: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("remove failed")
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	_, err := r.Delete(ctx, infer.DeleteRequest[ProjectNotificationState]{
		ID: "myproject:slack:deploy-slack",
		State: ProjectNotificationState{
			ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
			ProjectID:               42,
		},
	})
	if err == nil {
		t.Fatal("expected error when remove API fails")
	}
}

func TestProjectNotificationRead_RefreshFromState(t *testing.T) {
	mock := &mockLagoonClient{
		checkProjectNotificationExistsFn: func(_ context.Context, projectName, notifType, notifName string) (*client.ProjectNotificationInfo, error) {
			if projectName != "myproject" {
				t.Errorf("expected 'myproject', got %q", projectName)
			}
			return &client.ProjectNotificationInfo{ProjectID: 42, Exists: true}, nil
		},
	}
	ctx := testCtx(mock)
	r := &ProjectNotification{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectNotificationArgs, ProjectNotificationState]{
		ID: "some-id",
		State: ProjectNotificationState{
			ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "myproject", NotificationType: "slack", NotificationName: "deploy-slack"},
		},
	})
	if err != nil {
		t.Fatalf("Read from state failed: %v", err)
	}
	if resp.State.ProjectName != "myproject" {
		t.Errorf("expected ProjectName 'myproject', got %q", resp.State.ProjectName)
	}
}
