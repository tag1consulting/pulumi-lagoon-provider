package resources

import (
	"context"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// --- Project Diff Tests ---

func TestProjectDiff_NoChanges(t *testing.T) {
	r := &Project{}
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationSlackArgs, NotificationSlackState]{ID: "test", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationSlackArgs, NotificationSlackState]{ID: "old", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationSlackArgs, NotificationSlackState]{ID: "test", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationSlackArgs, NotificationSlackState]{ID: "test", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationRocketChatArgs, NotificationRocketChatState]{ID: "old", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationEmailArgs, NotificationEmailState]{ID: "test", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationEmailArgs, NotificationEmailState]{ID: "test", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{ID: "old", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationMicrosoftTeamsArgs, NotificationMicrosoftTeamsState]{ID: "test", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["webhook"]; !ok || d.Kind != p.Update {
		t.Error("expected webhook to be Update")
	}
}

// --- Group Diff Tests ---

func TestGroupDiff_NoChanges(t *testing.T) {
	r := &Group{}
	olds := GroupState{GroupArgs: GroupArgs{Name: "my-group"}}
	news := GroupArgs{Name: "my-group"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[GroupArgs, GroupState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestGroupDiff_NameForceNew(t *testing.T) {
	r := &Group{}
	olds := GroupState{GroupArgs: GroupArgs{Name: "old-group"}}
	news := GroupArgs{Name: "new-group"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[GroupArgs, GroupState]{ID: "1", State: olds, Inputs: news})
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

func TestGroupDiff_ParentGroupNameChange(t *testing.T) {
	r := &Group{}
	parent := "parent-a"
	newParent := "parent-b"
	olds := GroupState{GroupArgs: GroupArgs{Name: "my-group", ParentGroupName: &parent}}
	news := GroupArgs{Name: "my-group", ParentGroupName: &newParent}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[GroupArgs, GroupState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes")
	}
	if d, ok := resp.DetailedDiff["parentGroupName"]; !ok || d.Kind != p.Update {
		t.Error("expected parentGroupName to be Update (not replace)")
	}
	if _, ok := resp.DetailedDiff["name"]; ok {
		t.Error("expected name to not be in diff")
	}
}

// --- ProjectNotification Diff Tests ---

func TestProjectNotificationDiff_NoChanges(t *testing.T) {
	r := &ProjectNotification{}
	olds := ProjectNotificationState{ProjectNotificationArgs: ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "test"}}
	news := ProjectNotificationArgs{ProjectName: "proj", NotificationType: "slack", NotificationName: "test"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectNotificationArgs, ProjectNotificationState]{ID: "proj:slack:test", State: olds, Inputs: news})
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
			resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectNotificationArgs, ProjectNotificationState]{ID: "proj:slack:test", State: olds, Inputs: tt.news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectNotificationArgs, ProjectNotificationState]{ID: "proj:slack:test", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
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

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["service"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected service to be UpdateReplace")
	}
}

// --- DeployTarget: optional field update tests ---

func TestDeployTargetDiff_CloudProviderUpdate(t *testing.T) {
	r := &DeployTarget{}
	// nil CloudProvider normalizes to "kind"; changing to "aws" should detect a change
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	aws := "aws"
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", CloudProvider: &aws}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["cloudProvider"]; !ok || d.Kind != p.Update {
		t.Error("expected cloudProvider to be Update")
	}
}

func TestDeployTargetDiff_CloudRegionUpdate(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	region := "us-east-1"
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", CloudRegion: &region}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["cloudRegion"]; !ok || d.Kind != p.Update {
		t.Error("expected cloudRegion to be Update")
	}
}

func TestDeployTargetDiff_SSHHostUpdate(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	host := "ssh.example.com"
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", SSHHost: &host}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["sshHost"]; !ok || d.Kind != p.Update {
		t.Error("expected sshHost to be Update")
	}
}

func TestDeployTargetDiff_SSHPortUpdate(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	port := "2222"
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", SSHPort: &port}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["sshPort"]; !ok || d.Kind != p.Update {
		t.Error("expected sshPort to be Update")
	}
}

func TestDeployTargetDiff_BuildImageUpdate(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	img := "custom/build:latest"
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", BuildImage: &img}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["buildImage"]; !ok || d.Kind != p.Update {
		t.Error("expected buildImage to be Update")
	}
}

func TestDeployTargetDiff_DisabledUpdate(t *testing.T) {
	r := &DeployTarget{}
	// nil Disabled normalizes to &false; changing to &true should detect a change
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	trueVal := true
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", Disabled: &trueVal}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["disabled"]; !ok || d.Kind != p.Update {
		t.Error("expected disabled to be Update")
	}
}

func TestDeployTargetDiff_RouterPatternUpdate(t *testing.T) {
	r := &DeployTarget{}
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	pattern := "${environment}.${project}.example.com"
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com", RouterPattern: &pattern}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["routerPattern"]; !ok || d.Kind != p.Update {
		t.Error("expected routerPattern to be Update")
	}
}

func TestDeployTargetDiff_BothSidesNilNoChange(t *testing.T) {
	r := &DeployTarget{}
	// Both sides nil for all optional fields: normalized to same defaults, no diff
	olds := DeployTargetState{DeployTargetArgs: DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}}
	news := DeployTargetArgs{Name: "target", ConsoleURL: "https://k8s.example.com"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetArgs, DeployTargetState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Errorf("expected no changes when both sides have nil optional fields, got diff: %v", resp.DetailedDiff)
	}
}

// --- DeployTargetConfig: missing field tests ---

func TestDeployTargetConfigDiff_DeployTargetIDForceNew(t *testing.T) {
	r := &DeployTargetConfig{}
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 5}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["deployTargetId"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected deployTargetId to be UpdateReplace")
	}
}

func TestDeployTargetConfigDiff_PullrequestsUpdate(t *testing.T) {
	r := &DeployTargetConfig{}
	pr := "true"
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2, Pullrequests: &pr}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["pullrequests"]; !ok || d.Kind != p.Update {
		t.Error("expected pullrequests to be Update")
	}
}

func TestDeployTargetConfigDiff_WeightUpdate(t *testing.T) {
	r := &DeployTargetConfig{}
	w := 10
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2, Weight: &w}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["weight"]; !ok || d.Kind != p.Update {
		t.Error("expected weight to be Update")
	}
}

func TestDeployTargetConfigDiff_DeployTargetProjectPatternUpdate(t *testing.T) {
	r := &DeployTargetConfig{}
	pat := "${project}-${environment}"
	olds := DeployTargetConfigState{DeployTargetConfigArgs: DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2}}
	news := DeployTargetConfigArgs{ProjectID: 1, DeployTargetID: 2, DeployTargetProjectPattern: &pat}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["deployTargetProjectPattern"]; !ok || d.Kind != p.Update {
		t.Error("expected deployTargetProjectPattern to be Update")
	}
}

// --- NotificationRocketChat: missing tests ---

func TestNotificationRocketChatDiff_NoChanges(t *testing.T) {
	r := &NotificationRocketChat{}
	olds := NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "test", Webhook: "https://rc.com/hook", Channel: "#test"}}
	news := NotificationRocketChatArgs{Name: "test", Webhook: "https://rc.com/hook", Channel: "#test"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationRocketChatArgs, NotificationRocketChatState]{ID: "test", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("expected no changes")
	}
}

func TestNotificationRocketChatDiff_WebhookUpdate(t *testing.T) {
	r := &NotificationRocketChat{}
	olds := NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "test", Webhook: "https://old.rc.com/hook", Channel: "#test"}}
	news := NotificationRocketChatArgs{Name: "test", Webhook: "https://new.rc.com/hook", Channel: "#test"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationRocketChatArgs, NotificationRocketChatState]{ID: "test", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["webhook"]; !ok || d.Kind != p.Update {
		t.Error("expected webhook to be Update")
	}
}

func TestNotificationRocketChatDiff_ChannelUpdate(t *testing.T) {
	r := &NotificationRocketChat{}
	olds := NotificationRocketChatState{NotificationRocketChatArgs: NotificationRocketChatArgs{Name: "test", Webhook: "https://rc.com/hook", Channel: "#old-channel"}}
	news := NotificationRocketChatArgs{Name: "test", Webhook: "https://rc.com/hook", Channel: "#new-channel"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[NotificationRocketChatArgs, NotificationRocketChatState]{ID: "test", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["channel"]; !ok || d.Kind != p.Update {
		t.Error("expected channel to be Update")
	}
}

// --- Task: remaining Diff field tests ---

func TestTaskDiff_CommandReplace(t *testing.T) {
	r := &Task{}
	oldCmd := "drush cr"
	newCmd := "drush updb"
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", Command: &oldCmd}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", Command: &newCmd}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["command"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected command to be UpdateReplace")
	}
}

func TestTaskDiff_ImageReplace(t *testing.T) {
	r := &Task{}
	oldImg := "alpine:3.18"
	newImg := "alpine:3.19"
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "image", Service: "cli", Image: &oldImg}}
	news := TaskArgs{Name: "task", Type: "image", Service: "cli", Image: &newImg}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["image"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected image to be UpdateReplace")
	}
}

func TestTaskDiff_DescriptionReplace(t *testing.T) {
	r := &Task{}
	oldDesc := "old description"
	newDesc := "new description"
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", Description: &oldDesc}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", Description: &newDesc}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["description"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected description to be UpdateReplace")
	}
}

func TestTaskDiff_PermissionCaseInsensitiveNoChange(t *testing.T) {
	r := &Task{}
	upper := "MAINTAINER"
	lower := "maintainer"
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", Permission: &upper}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", Permission: &lower}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if _, ok := resp.DetailedDiff["permission"]; ok {
		t.Error("expected no permission diff for case-insensitive match")
	}
}

func TestTaskDiff_PermissionReplace(t *testing.T) {
	r := &Task{}
	oldPerm := "guest"
	newPerm := "maintainer"
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", Permission: &oldPerm}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", Permission: &newPerm}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["permission"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected permission to be UpdateReplace")
	}
}

func TestTaskDiff_ArgumentsReplace(t *testing.T) {
	r := &Task{}
	oldArgs := []TaskArgumentInput{{Name: "arg1", DisplayName: "Arg 1", Type: "string"}}
	newArgs := []TaskArgumentInput{{Name: "arg1", DisplayName: "Arg 1 Updated", Type: "string"}}
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", Arguments: &oldArgs}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", Arguments: &newArgs}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["arguments"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected arguments to be UpdateReplace")
	}
}

func TestTaskDiff_ProjectIDReplace(t *testing.T) {
	r := &Task{}
	oldPID := 1
	newPID := 2
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", ProjectID: &oldPID}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", ProjectID: &newPID}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["projectId"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected projectId to be UpdateReplace")
	}
}

func TestTaskDiff_SystemWideReplace(t *testing.T) {
	r := &Task{}
	oldSW := false
	newSW := true
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", SystemWide: &oldSW}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", SystemWide: &newSW}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["systemWide"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected systemWide to be UpdateReplace")
	}
}

func TestTaskDiff_GroupNameReplace(t *testing.T) {
	r := &Task{}
	oldGN := "group-a"
	newGN := "group-b"
	olds := TaskState{TaskArgs: TaskArgs{Name: "task", Type: "command", Service: "cli", GroupName: &oldGN}}
	news := TaskArgs{Name: "task", Type: "command", Service: "cli", GroupName: &newGN}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[TaskArgs, TaskState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["groupName"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected groupName to be UpdateReplace")
	}
}

// --- Project: additional optional field tests ---

func TestProjectDiff_AutoIdleUpdate(t *testing.T) {
	r := &Project{}
	autoIdle := 1
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1, AutoIdle: &autoIdle}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["autoIdle"]; !ok || d.Kind != p.Update {
		t.Error("expected autoIdle to be Update")
	}
}

func TestProjectDiff_StorageCalcUpdate(t *testing.T) {
	r := &Project{}
	sc := 1
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1, StorageCalc: &sc}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["storageCalc"]; !ok || d.Kind != p.Update {
		t.Error("expected storageCalc to be Update")
	}
}

func TestProjectDiff_OpenshiftProjectPatternUpdate(t *testing.T) {
	r := &Project{}
	pattern := "${project}-${environment}"
	olds := ProjectState{ProjectArgs: ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1}}
	news := ProjectArgs{Name: "proj", GitURL: "git@example.com:repo.git", DeploytargetID: 1, OpenshiftProjectPattern: &pattern}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["openshiftProjectPattern"]; !ok || d.Kind != p.Update {
		t.Error("expected openshiftProjectPattern to be Update")
	}
}

// --- Environment: additional field tests ---

func TestEnvironmentDiff_EnvironmentTypeUpdate(t *testing.T) {
	r := &Environment{}
	// Case-insensitive comparison: changing from "production" to "development" is a real change
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "development"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["environmentType"]; !ok || d.Kind != p.Update {
		t.Error("expected environmentType to be Update")
	}
}

func TestEnvironmentDiff_EnvironmentTypeCaseInsensitiveNoChange(t *testing.T) {
	r := &Environment{}
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "PRODUCTION"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if _, ok := resp.DetailedDiff["environmentType"]; ok {
		t.Error("expected no environmentType diff for case-insensitive match")
	}
}

func TestEnvironmentDiff_AutoIdleUpdate(t *testing.T) {
	r := &Environment{}
	autoIdle := 1
	olds := EnvironmentState{EnvironmentArgs: EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production"}}
	news := EnvironmentArgs{Name: "main", ProjectID: 1, DeployType: "branch", EnvironmentType: "production", AutoIdle: &autoIdle}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[EnvironmentArgs, EnvironmentState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["autoIdle"]; !ok || d.Kind != p.Update {
		t.Error("expected autoIdle to be Update")
	}
}

// --- Variable: additional field tests ---

func TestVariableDiff_ProjectIDForceNew(t *testing.T) {
	r := &Variable{}
	olds := VariableState{VariableArgs: VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build"}}
	news := VariableArgs{Name: "VAR", Value: "val", ProjectID: 2, Scope: "build"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["projectId"]; !ok || d.Kind != p.UpdateReplace {
		t.Error("expected projectId to be UpdateReplace")
	}
}

func TestVariableDiff_ScopeRealChange(t *testing.T) {
	r := &Variable{}
	olds := VariableState{VariableArgs: VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "build"}}
	news := VariableArgs{Name: "VAR", Value: "val", ProjectID: 1, Scope: "runtime"}

	resp, err := r.Diff(context.Background(), infer.DiffRequest[VariableArgs, VariableState]{ID: "1", State: olds, Inputs: news})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if d, ok := resp.DetailedDiff["scope"]; !ok || d.Kind != p.Update {
		t.Error("expected scope to be Update when values actually differ")
	}
}
