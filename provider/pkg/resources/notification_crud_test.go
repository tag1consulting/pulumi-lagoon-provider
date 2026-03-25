package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// ==================== Slack ====================

func TestNotificationSlackCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationSlackFn: func(_ context.Context, name, webhook, channel string) (*client.Notification, error) {
			return &client.Notification{ID: 10, Name: name, Webhook: webhook, Channel: channel}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	resp, err := r.Create(ctx, infer.CreateRequest[NotificationSlackArgs]{
		Inputs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/xxx", Channel: "#deployments"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "10" {
		t.Errorf("expected ID '10', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 10 {
		t.Errorf("expected LagoonID 10")
	}
}

func TestNotificationSlackCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		createNotificationSlackFn: func(_ context.Context, _, _, _ string) (*client.Notification, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	resp, err := r.Create(ctx, infer.CreateRequest[NotificationSlackArgs]{
		Inputs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/xxx", Channel: "#deployments"},
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

func TestNotificationSlackUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationSlackFn: func(_ context.Context, name string, patch map[string]any) (*client.Notification, error) {
			if _, ok := patch["webhook"]; !ok {
				t.Error("expected webhook in patch")
			}
			return &client.Notification{ID: 10, Name: name}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	resp, err := r.Update(ctx, infer.UpdateRequest[NotificationSlackArgs, NotificationSlackState]{
		ID:     "10",
		Inputs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/new", Channel: "#deployments"},
		State:  NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/old", Channel: "#deployments"}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.Webhook != "https://hooks.slack.com/new" {
		t.Errorf("expected updated webhook")
	}
}

func TestNotificationSlackUpdate_NoPatch(t *testing.T) {
	// When nothing changed, no API call should be made
	called := false
	mock := &mockLagoonClient{
		updateNotificationSlackFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationSlackArgs, NotificationSlackState]{
		ID:     "10",
		Inputs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/same", Channel: "#deployments"},
		State:  NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/same", Channel: "#deployments"}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if called {
		t.Error("API should not be called when nothing changed")
	}
}

func TestNotificationSlackDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationSlackFn: func(_ context.Context, name string) error {
			if name != "deploy-slack" {
				t.Errorf("expected 'deploy-slack', got %q", name)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationSlackState]{
		ID:    "10",
		State: NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack"}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNotificationSlackDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationSlackFn: func(_ context.Context, _ string) error {
			return &client.LagoonNotFoundError{ResourceType: "NotificationSlack", Identifier: "deploy-slack"}
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationSlackState]{
		ID:    "10",
		State: NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack"}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestNotificationSlackRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationSlackByNameFn: func(_ context.Context, name string) (*client.Notification, error) {
			return &client.Notification{ID: 10, Name: name, Webhook: "https://hooks.slack.com/xxx", Channel: "#deployments"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationSlackArgs, NotificationSlackState]{
		ID:    "10",
		State: NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack"}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "10" {
		t.Errorf("expected ID '10', got %q", resp.ID)
	}
	if resp.Inputs.Channel != "#deployments" {
		t.Errorf("expected Channel '#deployments'")
	}
}

func TestNotificationSlackRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationSlackByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "NotificationSlack", Identifier: "deploy-slack"}
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationSlackArgs, NotificationSlackState]{
		ID:    "10",
		State: NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack"}, LagoonID: 10},
	})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestNotificationSlackRead_FallbackToID(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationSlackByNameFn: func(_ context.Context, name string) (*client.Notification, error) {
			if name != "import-name" {
				t.Errorf("expected fallback to ID 'import-name', got %q", name)
			}
			return &client.Notification{ID: 10, Name: name, Webhook: "https://hooks.slack.com/xxx", Channel: "#test"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Read(ctx, infer.ReadRequest[NotificationSlackArgs, NotificationSlackState]{
		ID: "import-name",
		// Empty state — simulates import
	})
	if err != nil {
		t.Fatalf("Read with fallback failed: %v", err)
	}
}

// ==================== RocketChat ====================

func TestNotificationRocketChatCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationRocketChatFn: func(_ context.Context, name, webhook, channel string) (*client.Notification, error) {
			return &client.Notification{ID: 11, Name: name, Webhook: webhook, Channel: channel}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	resp, err := r.Create(ctx, infer.CreateRequest[NotificationRocketChatArgs]{
		Inputs: NotificationRocketChatArgs{Name: "rc-notif", Webhook: "https://rc.example.com/hook", Channel: "#general"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "11" {
		t.Errorf("expected ID '11', got %q", resp.ID)
	}
}

func TestNotificationRocketChatDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationRocketChatFn: func(_ context.Context, name string) error {
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationRocketChatState]{
		State: NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif"}, LagoonID: 11},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNotificationRocketChatRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationRocketChatByNameFn: func(_ context.Context, name string) (*client.Notification, error) {
			return &client.Notification{ID: 11, Name: name, Webhook: "https://rc.example.com/hook", Channel: "#general"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationRocketChatArgs, NotificationRocketChatState]{
		ID:    "11",
		State: NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif"}, LagoonID: 11},
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.Inputs.Channel != "#general" {
		t.Errorf("expected Channel '#general'")
	}
}

func TestNotificationRocketChatRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationRocketChatByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "NotificationRocketChat", Identifier: "rc-notif"}
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationRocketChatArgs, NotificationRocketChatState]{
		ID:    "11",
		State: NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif"}, LagoonID: 11},
	})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestNotificationRocketChatUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationRocketChatFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			return &client.Notification{ID: 11, Name: "rc-notif"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationRocketChatArgs, NotificationRocketChatState]{
		Inputs: NotificationRocketChatArgs{Name: "rc-notif", Webhook: "https://rc.example.com/new", Channel: "#general"},
		State:  NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif", Webhook: "https://rc.example.com/old", Channel: "#general"}, LagoonID: 11},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

// ==================== Email ====================

func TestNotificationEmailCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationEmailFn: func(_ context.Context, name, email string) (*client.Notification, error) {
			return &client.Notification{ID: 12, Name: name, EmailAddress: email}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	resp, err := r.Create(ctx, infer.CreateRequest[NotificationEmailArgs]{
		Inputs: NotificationEmailArgs{Name: "email-notif", EmailAddress: "admin@example.com"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "12" {
		t.Errorf("expected ID '12', got %q", resp.ID)
	}
}

func TestNotificationEmailDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationEmailFn: func(_ context.Context, name string) error {
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationEmailState]{
		State: NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif"}, LagoonID: 12},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNotificationEmailRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationEmailByNameFn: func(_ context.Context, name string) (*client.Notification, error) {
			return &client.Notification{ID: 12, Name: name, EmailAddress: "admin@example.com"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationEmailArgs, NotificationEmailState]{
		ID:    "12",
		State: NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif"}, LagoonID: 12},
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.Inputs.EmailAddress != "admin@example.com" {
		t.Errorf("expected EmailAddress")
	}
}

func TestNotificationEmailRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationEmailByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "NotificationEmail", Identifier: "email-notif"}
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationEmailArgs, NotificationEmailState]{
		ID:    "12",
		State: NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif"}, LagoonID: 12},
	})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestNotificationEmailUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationEmailFn: func(_ context.Context, _ string, patch map[string]any) (*client.Notification, error) {
			if patch["emailAddress"] != "new@example.com" {
				t.Errorf("expected emailAddress in patch")
			}
			return &client.Notification{ID: 12, Name: "email-notif"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationEmailArgs, NotificationEmailState]{
		Inputs: NotificationEmailArgs{Name: "email-notif", EmailAddress: "new@example.com"},
		State:  NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif", EmailAddress: "old@example.com"}, LagoonID: 12},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

// ==================== Microsoft Teams ====================

func TestNotificationMicrosoftTeamsCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationMicrosoftTeamsFn: func(_ context.Context, name, webhook string) (*client.Notification, error) {
			return &client.Notification{ID: 13, Name: name, Webhook: webhook}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	resp, err := r.Create(ctx, infer.CreateRequest[NotificationMicrosoftTeamsArgs]{
		Inputs: NotificationMicrosoftTeamsArgs{Name: "teams-notif", Webhook: "https://teams.example.com/hook"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "13" {
		t.Errorf("expected ID '13', got %q", resp.ID)
	}
}

func TestNotificationMicrosoftTeamsDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationMicrosoftTeamsFn: func(_ context.Context, name string) error {
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationMicrosoftTeamsState]{
		State: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif"}, LagoonID: 13},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestNotificationMicrosoftTeamsRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationMicrosoftTeamsByNameFn: func(_ context.Context, name string) (*client.Notification, error) {
			return &client.Notification{ID: 13, Name: name, Webhook: "https://teams.example.com/hook"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{
		ID:    "13",
		State: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif"}, LagoonID: 13},
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.Inputs.Webhook != "https://teams.example.com/hook" {
		t.Errorf("expected Webhook")
	}
}

func TestNotificationMicrosoftTeamsRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationMicrosoftTeamsByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "NotificationMicrosoftTeams", Identifier: "teams-notif"}
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	resp, err := r.Read(ctx, infer.ReadRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{
		ID:    "13",
		State: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif"}, LagoonID: 13},
	})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestNotificationMicrosoftTeamsUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationMicrosoftTeamsFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			return &client.Notification{ID: 13, Name: "teams-notif"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{
		Inputs: NotificationMicrosoftTeamsArgs{Name: "teams-notif", Webhook: "https://teams.example.com/new"},
		State:  NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif", Webhook: "https://teams.example.com/old"}, LagoonID: 13},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
}

func TestNotificationSlackCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationSlackFn: func(_ context.Context, _, _, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Create(ctx, infer.CreateRequest[NotificationSlackArgs]{
		Inputs: NotificationSlackArgs{Name: "fail", Webhook: "https://x", Channel: "#x"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationSlackDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationSlackFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationSlackState]{
		State: NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "fail"}, LagoonID: 10},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationSlackUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationSlackFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationSlackArgs, NotificationSlackState]{
		ID:     "10",
		Inputs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/new", Channel: "#deployments"},
		State:  NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack", Webhook: "https://hooks.slack.com/old", Channel: "#deployments"}, LagoonID: 10},
	})
	if err == nil {
		t.Fatal("expected error when update API fails")
	}
}

func TestNotificationSlackRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationSlackByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationSlack{}
	_, err := r.Read(ctx, infer.ReadRequest[NotificationSlackArgs, NotificationSlackState]{
		ID:    "10",
		State: NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "deploy-slack"}, LagoonID: 10},
	})
	if err == nil {
		t.Fatal("expected error when read API fails")
	}
}

// ==================== RocketChat Error Paths ====================

func TestNotificationRocketChatCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationRocketChatFn: func(_ context.Context, _, _, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	_, err := r.Create(ctx, infer.CreateRequest[NotificationRocketChatArgs]{
		Inputs: NotificationRocketChatArgs{Name: "fail", Webhook: "https://x", Channel: "#x"},
	})
	if err == nil {
		t.Fatal("expected error when create API fails")
	}
}

func TestNotificationRocketChatUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationRocketChatFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationRocketChatArgs, NotificationRocketChatState]{
		Inputs: NotificationRocketChatArgs{Name: "rc-notif", Webhook: "https://rc.example.com/new", Channel: "#general"},
		State:  NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif", Webhook: "https://rc.example.com/old", Channel: "#general"}, LagoonID: 11},
	})
	if err == nil {
		t.Fatal("expected error when update API fails")
	}
}

func TestNotificationRocketChatDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationRocketChatFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("delete failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationRocketChatState]{
		State: NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif"}, LagoonID: 11},
	})
	if err == nil {
		t.Fatal("expected error when delete API fails")
	}
}

func TestNotificationRocketChatRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationRocketChatByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationRocketChat{}
	_, err := r.Read(ctx, infer.ReadRequest[NotificationRocketChatArgs, NotificationRocketChatState]{
		ID:    "11",
		State: NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "rc-notif"}, LagoonID: 11},
	})
	if err == nil {
		t.Fatal("expected error when read API fails")
	}
}

// ==================== Email Error Paths ====================

func TestNotificationEmailCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationEmailFn: func(_ context.Context, _, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	_, err := r.Create(ctx, infer.CreateRequest[NotificationEmailArgs]{
		Inputs: NotificationEmailArgs{Name: "fail", EmailAddress: "fail@example.com"},
	})
	if err == nil {
		t.Fatal("expected error when create API fails")
	}
}

func TestNotificationEmailUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationEmailFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationEmailArgs, NotificationEmailState]{
		Inputs: NotificationEmailArgs{Name: "email-notif", EmailAddress: "new@example.com"},
		State:  NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif", EmailAddress: "old@example.com"}, LagoonID: 12},
	})
	if err == nil {
		t.Fatal("expected error when update API fails")
	}
}

func TestNotificationEmailDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationEmailFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("delete failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationEmailState]{
		State: NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif"}, LagoonID: 12},
	})
	if err == nil {
		t.Fatal("expected error when delete API fails")
	}
}

func TestNotificationEmailRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationEmailByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationEmail{}
	_, err := r.Read(ctx, infer.ReadRequest[NotificationEmailArgs, NotificationEmailState]{
		ID:    "12",
		State: NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "email-notif"}, LagoonID: 12},
	})
	if err == nil {
		t.Fatal("expected error when read API fails")
	}
}

// ==================== Microsoft Teams Error Paths ====================

func TestNotificationMicrosoftTeamsCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		createNotificationMicrosoftTeamsFn: func(_ context.Context, _, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	_, err := r.Create(ctx, infer.CreateRequest[NotificationMicrosoftTeamsArgs]{
		Inputs: NotificationMicrosoftTeamsArgs{Name: "fail", Webhook: "https://teams.example.com/hook"},
	})
	if err == nil {
		t.Fatal("expected error when create API fails")
	}
}

func TestNotificationMicrosoftTeamsUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		updateNotificationMicrosoftTeamsFn: func(_ context.Context, _ string, _ map[string]any) (*client.Notification, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	_, err := r.Update(ctx, infer.UpdateRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{
		Inputs: NotificationMicrosoftTeamsArgs{Name: "teams-notif", Webhook: "https://teams.example.com/new"},
		State:  NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif", Webhook: "https://teams.example.com/old"}, LagoonID: 13},
	})
	if err == nil {
		t.Fatal("expected error when update API fails")
	}
}

func TestNotificationMicrosoftTeamsDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteNotificationMicrosoftTeamsFn: func(_ context.Context, _ string) error {
			return fmt.Errorf("delete failed")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	_, err := r.Delete(ctx, infer.DeleteRequest[NotificationMicrosoftTeamsState]{
		State: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif"}, LagoonID: 13},
	})
	if err == nil {
		t.Fatal("expected error when delete API fails")
	}
}

func TestNotificationMicrosoftTeamsRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getNotificationMicrosoftTeamsByNameFn: func(_ context.Context, _ string) (*client.Notification, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &NotificationMicrosoftTeams{}
	_, err := r.Read(ctx, infer.ReadRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{
		ID:    "13",
		State: NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "teams-notif"}, LagoonID: 13},
	})
	if err == nil {
		t.Fatal("expected error when read API fails")
	}
}
