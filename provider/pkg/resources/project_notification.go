package resources

import (
	"context"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

// ProjectNotification manages the link between a notification and a project.
type ProjectNotification struct{}

type ProjectNotificationArgs struct {
	ProjectName      string `pulumi:"projectName"`
	NotificationType string `pulumi:"notificationType"`
	NotificationName string `pulumi:"notificationName"`
}

type ProjectNotificationState struct {
	ProjectNotificationArgs
	ProjectID int `pulumi:"projectId"`
}

func (r *ProjectNotification) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "ProjectNotification")
	a.Describe(&r, "Links a notification to a Lagoon project.")
}

func (a *ProjectNotificationArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.ProjectName, "The project name.")
	an.Describe(&a.NotificationType, "Type of notification: 'slack', 'rocketchat', 'email', or 'microsoftteams'.")
	an.Describe(&a.NotificationName, "Name of the notification to link.")
}

func (s *ProjectNotificationState) Annotate(a infer.Annotator) {
	a.Describe(&s.ProjectID, "The Lagoon project ID.")
}

func (r *ProjectNotification) Create(ctx context.Context, name string, inputs ProjectNotificationArgs, preview bool) (string, ProjectNotificationState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	notifType := strings.ToUpper(inputs.NotificationType)
	id := fmt.Sprintf("%s:%s:%s", inputs.ProjectName, inputs.NotificationType, inputs.NotificationName)

	if preview {
		return id, ProjectNotificationState{ProjectNotificationArgs: inputs}, nil
	}

	if err := client.AddNotificationToProject(ctx, inputs.ProjectName, notifType, inputs.NotificationName); err != nil {
		return "", ProjectNotificationState{}, fmt.Errorf("failed to add notification to project: %w", err)
	}

	// Look up project ID
	info, err := client.CheckProjectNotificationExists(ctx, inputs.ProjectName, inputs.NotificationType, inputs.NotificationName)
	projectID := 0
	if err == nil && info != nil {
		projectID = info.ProjectID
	}

	return id, ProjectNotificationState{
		ProjectNotificationArgs: inputs,
		ProjectID:               projectID,
	}, nil
}

// No Update — all fields are forceNew, so any change triggers replace.

func (r *ProjectNotification) Delete(ctx context.Context, id string, props ProjectNotificationState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	notifType := strings.ToUpper(props.NotificationType)
	if err := client.RemoveNotificationFromProject(ctx, props.ProjectName, notifType, props.NotificationName); err != nil {
		return fmt.Errorf("failed to remove notification from project: %w", err)
	}
	return nil
}

func (r *ProjectNotification) Read(ctx context.Context, id string, inputs ProjectNotificationArgs, state ProjectNotificationState) (string, ProjectNotificationArgs, ProjectNotificationState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Import ID: {project_name}:{notification_type}:{notification_name}
	parts := strings.SplitN(id, ":", 3)
	var projectName, notificationType, notificationName string

	if len(parts) == 3 {
		projectName = parts[0]
		notificationType = parts[1]
		notificationName = parts[2]
	} else {
		projectName = state.ProjectName
		notificationType = state.NotificationType
		notificationName = state.NotificationName
	}

	info, err := client.CheckProjectNotificationExists(ctx, projectName, notificationType, notificationName)
	if err != nil {
		return "", ProjectNotificationArgs{}, ProjectNotificationState{}, fmt.Errorf("failed to read project notification: %w", err)
	}

	if !info.Exists {
		return "", ProjectNotificationArgs{}, ProjectNotificationState{},
			fmt.Errorf("notification '%s' (type=%s) not found on project '%s'", notificationName, notificationType, projectName)
	}

	args := ProjectNotificationArgs{
		ProjectName:      projectName,
		NotificationType: notificationType,
		NotificationName: notificationName,
	}
	st := ProjectNotificationState{
		ProjectNotificationArgs: args,
		ProjectID:               info.ProjectID,
	}

	return id, args, st, nil
}

func (r *ProjectNotification) Diff(ctx context.Context, id string, olds ProjectNotificationState, news ProjectNotificationArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// All fields are forceNew for project notifications
	if news.ProjectName != olds.ProjectName {
		diff["projectName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if !strings.EqualFold(news.NotificationType, olds.NotificationType) {
		diff["notificationType"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.NotificationName != olds.NotificationName {
		diff["notificationName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
