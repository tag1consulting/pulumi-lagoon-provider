package resources

import (
	"context"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
)

// --- Project Diff Tests ---

func TestProjectDiff_NoChanges(t *testing.T) {
	r := &Project{}
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestProjectDiff_NameForceNew(t *testing.T) {
	r := &Project{}
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "old-name", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "new-name", GitURL: "git@example.com:repo.git", DeploytargetID: 1}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes")
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

func TestProjectDiff_GitURLUpdate(t *testing.T) {
	r := &Project{}
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@old.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@new.com:repo.git", DeploytargetID: 1}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes")
	}
	if d, ok := resp.DetailedDiff["gitUrl"]; !ok || d.Kind != p.Update {
		t.Error("expected gitUrl to be Update")
	}
}

func TestProjectDiff_OptionalFields(t *testing.T) {
	r := &Project{}
	branch := "main"
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1, Branches: &branch}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes")
	}
	if d, ok := resp.DetailedDiff["branches"]; !ok || d.Kind != p.Update {
		t.Error("expected branches to be Update")
	}
}

func TestProjectDiff_DeleteBeforeReplace(t *testing.T) {
	r := &Project{}
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "new-proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.DeleteBeforeReplace {
		t.Error("expected DeleteBeforeReplace to be true")
	}
}

// --- Environment Diff Tests ---

func TestEnvironmentDiff_NoChanges(t *testing.T) {
	r := &Environment{}
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestEnvironmentDiff_NameForceNew(t *testing.T) {
	r := &Environment{}
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "old", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "new", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

func TestEnvironmentDiff_ProjectIDForceNew(t *testing.T) {
	r := &Environment{}
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 2, DeployType: "branch", EnvironmentType: "production"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["projectId"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected projectId to be UpdateReplace")
	}
}

func TestEnvironmentDiff_DeployTypeUpdate(t *testing.T) {
	r := &Environment{}
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "pullrequest", EnvironmentType: "production"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["deployType"]; !ok || d.Kind != p.Update {
		t.Error("expected deployType to be Update")
	}
}

func TestEnvironmentDiff_CaseInsensitiveDeployType(t *testing.T) {
	r := &Environment{}
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "BRANCH", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes for case-insensitive deploy type")
	}
}

// --- Variable Diff Tests ---

func TestVariableDiff_NoChanges(t *testing.T) {
	r := &Variable{}
	olds := VariableState{VariableArgs: VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build"}}
	news := VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestVariableDiff_NameForceNew(t *testing.T) {
	r := &Variable{}
	olds := VariableState{VariableArgs: VariableArgs{Name: "OLD_VAR", Value: "val", ProjectID: 1, Scope: "build"}}
	news := VariableArgs{Name: "NEW_VAR", Value: "val", ProjectID: 1, Scope: "build"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

func TestVariableDiff_ValueUpdate(t *testing.T) {
	r := &Variable{}
	olds := VariableState{VariableArgs: VariableArgs{Name: "VAR", Value: "old", ProjectID: 1, Scope: "build"}}
	news := VariableArgs{Name: "VAR", Value: "new", ProjectID: 1, Scope: "build"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["value"]; !ok || d.Kind != p.Update {
		t.Error("expected value to be Update")
	}
}

func TestVariableDiff_EnvironmentIDForceNew(t *testing.T) {
	r := &Variable{}
	envID := 5
	olds := VariableState{VariableArgs: VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build"}}
	news := VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build", EnvironmentID: &envID}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["environmentId"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected environmentId to be UpdateReplace")
	}
}

func TestVariableDiff_ScopeCaseInsensitive(t *testing.T) {
	r := &Variable{}
	olds := VariableState{VariableArgs: VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "BUILD"}}
	news := VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes for case-insensitive scope")
	}
}

// --- DeployTarget Diff Tests ---

func TestDeployTargetDiff_NoChanges(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestDeployTargetDiff_NameForceNew(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "old", ConsoleURL: "https://k8s.example.com"}}
	news := DeployTargetArgs{Name: "new", ConsoleURL: "https://k8s.example.com"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

func TestDeployTargetDiff_ConsoleURLUpdate(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://old.example.com"}}
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://new.example.com"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["consoleUrl"]; !ok || d.Kind != p.Update {
		t.Error("expected consoleUrl to be Update")
	}
}

// --- DeployTargetConfig Diff Tests ---

func TestDeployTargetConfigDiff_NoChanges(t *testing.T) {
	r := &DeployTargetConfig{}
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestDeployTargetConfigDiff_ProjectIDForceNew(t *testing.T) {
	r := &DeployTargetConfig{}
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 3, DeployTargetID: 2}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["projectId"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected projectId to be UpdateReplace")
	}
}

func TestDeployTargetConfigDiff_BranchesUpdate(t *testing.T) {
	r := &DeployTargetConfig{}
	b := "main"
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2, Branches: &b}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["branches"]; !ok || d.Kind != p.Update {
		t.Error("expected branches to be Update")
	}
}

// --- NotificationSlack Diff Tests ---

func TestNotificationSlackDiff_NoChanges(t *testing.T) {
	r := &NotificationSlack{}
	olds := NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "test", Webhook: "https://hooks.slack.com/x", Channel: "#test"}}
	news := NotificationSlackArgs{Name: "test", Webhook: "https://hooks.slack.com/x", Channel: "#test"}

	resp, err := r.Diff(context.Background(), "test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestNotificationSlackDiff_NameForceNew(t *testing.T) {
	r := &NotificationSlack{}
	olds := NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "old", Webhook: "https://hooks.slack.com/x", Channel: "#test"}}
	news := NotificationSlackArgs{Name: "new", Webhook: "https://hooks.slack.com/x", Channel: "#test"}

	resp, err := r.Diff(context.Background(), "old", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

func TestNotificationSlackDiff_WebhookUpdate(t *testing.T) {
	r := &NotificationSlack{}
	olds := NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "test", Webhook: "https://old.com", Channel: "#test"}}
	news := NotificationSlackArgs{Name: "test", Webhook: "https://new.com", Channel: "#test"}

	resp, err := r.Diff(context.Background(), "test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["webhook"]; !ok || d.Kind != p.Update {
		t.Error("expected webhook to be Update")
	}
}

func TestNotificationSlackDiff_ChannelUpdate(t *testing.T) {
	r := &NotificationSlack{}
	olds := NotificationSlackState{NotificationSlackArgs: NotificationSlackArgs{Name: "test", Webhook: "https://hook.com", Channel: "#old"}}
	news := NotificationSlackArgs{Name: "test", Webhook: "https://hook.com", Channel: "#new"}

	resp, err := r.Diff(context.Background(), "test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["channel"]; !ok || d.Kind != p.Update {
		t.Error("expected channel to be Update")
	}
}

// --- NotificationRocketChat Diff Tests ---

func TestNotificationRocketChatDiff_NameForceNew(t *testing.T) {
	r := &NotificationRocketChat{}
	olds := NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "old", Webhook: "https://rc.com/hook", Channel: "#test"}}
	news := NotificationRocketChatArgs{Name: "new", Webhook: "https://rc.com/hook", Channel: "#test"}

	resp, err := r.Diff(context.Background(), "old", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

// --- NotificationEmail Diff Tests ---

func TestNotificationEmailDiff_NoChanges(t *testing.T) {
	r := &NotificationEmail{}
	olds := NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "test", EmailAddress: "test@example.com"}}
	news := NotificationEmailArgs{Name: "test", EmailAddress: "test@example.com"}

	resp, err := r.Diff(context.Background(), "test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestNotificationEmailDiff_EmailUpdate(t *testing.T) {
	r := &NotificationEmail{}
	olds := NotificationEmailState{NotificationEmailArgs: NotificationEmailArgs{Name: "test", EmailAddress: "old@example.com"}}
	news := NotificationEmailArgs{Name: "test", EmailAddress: "new@example.com"}

	resp, err := r.Diff(context.Background(), "test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["emailAddress"]; !ok || d.Kind != p.Update {
		t.Error("expected emailAddress to be Update")
	}
}

// --- NotificationMicrosoftTeams Diff Tests ---

func TestNotificationMicrosoftTeamsDiff_NameForceNew(t *testing.T) {
	r := &NotificationMicrosoftTeams{}
	olds := NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "old", Webhook: "https://teams.com/hook"}}
	news := NotificationMicrosoftTeamsArgs{Name: "new", Webhook: "https://teams.com/hook"}

	resp, err := r.Diff(context.Background(), "old", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
}

func TestNotificationMicrosoftTeamsDiff_WebhookUpdate(t *testing.T) {
	r := &NotificationMicrosoftTeams{}
	olds := NotificationMicrosoftTeamsState{NotificationMicrosoftTeamsArgs: NotificationMicrosoftTeamsArgs{Name: "test", Webhook: "https://old.com"}}
	news := NotificationMicrosoftTeamsArgs{Name: "test", Webhook: "https://new.com"}

	resp, err := r.Diff(context.Background(), "test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["webhook"]; !ok || d.Kind != p.Update {
		t.Error("expected webhook to be Update")
	}
}

// --- ProjectNotification Diff Tests ---

func TestProjectNotificationDiff_NoChanges(t *testing.T) {
	r := &ProjectNotification{}
	olds := ProjectNotificationState{ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "test"}}
	news := ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "test"}

	resp, err := r.Diff(context.Background(), "proj:slack:test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestProjectNotificationDiff_AllFieldsForceNew(t *testing.T) {
	r := &ProjectNotification{}
	olds := ProjectNotificationState{ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "test"}}

	tests := []struct {
		name  string
		news  ProjectNotificationArgs
		field string
	}{
		{"projectName", ProjectNotificationArgs{ProjectName: "other", NotificationType: "slack", NotificationName: "test"}, "projectName"},
		{"notificationType", ProjectNotificationArgs{ProjectName: "proj", NotificationType: "email", NotificationName: "test"}, "notificationType"},
		{"notificationName", ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "other"}, "notificationName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := r.Diff(context.Background(), "proj:slack:test", olds, tt.news)
			if err != nil {
				t.Fatalf("Diff failed: %v", err)
			}
			if d, ok := resp.DetailedDiff[tt.field]; !ok || d.Kind != p.UpdateReplace {
				t.Errorf("expected %s to be UpdateReplace", tt.field)
			}
		})
	}
}

func TestProjectNotificationDiff_CaseInsensitiveType(t *testing.T) {
	r := &ProjectNotification{}
	olds := ProjectNotificationState{ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "proj", NotificationType: "SLACK", NotificationName: "test"}}
	news := ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "test"}

	resp, err := r.Diff(context.Background(), "proj:slack:test", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes for case-insensitive notification type")
	}
}

// --- Task Diff Tests ---

func TestTaskDiff_NoChanges(t *testing.T) {
	r := &Task{}
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli"}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestTaskDiff_TypeForceNew(t *testing.T) {
	r := &Task{}
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli"}}
	news := TaskArgs{Name: "task", Type: "image", Service: "cli"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["type"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected type to be UpdateReplace")
	}
}

func TestTaskDiff_TypeCaseInsensitive(t *testing.T) {
	r := &Task{}
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "COMMAND", Service: "cli"}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes for case-insensitive type")
	}
}

func TestTaskDiff_AllReplace(t *testing.T) {
	// Since Lagoon doesn't support updating task definitions, ALL changes trigger replace
	r := &Task{}
	olds := TaskState{TaskArgs: TaskArgs{Name: "old", Type: "command", Service: "cli"}}
	news := TaskArgs{Name: "new", Type: "command", Service: "cli"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["name"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected name to be UpdateReplace")
	}
	if !resp.DeleteBeforeReplace {
		t.Error("expected DeleteBeforeReplace for task")
	}
}

func TestTaskDiff_ServiceReplace(t *testing.T) {
	r := &Task{}
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli"}}
	news := TaskArgs{Name: "task", Type: "command", Service: "node"}

	resp, err := r.Diff(context.Background(), "1", olds, news)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["service"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected service to be UpdateReplace")
	}
}
