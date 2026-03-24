package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
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

func (r *ProjectNotification) Create(ctx context.Context, req infer.CreateRequest[ProjectNotificationArgs]) (infer.CreateResponse[ProjectNotificationState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	notifType := strings.ToUpper(req.Inputs.NotificationType)
	id := fmt.Sprintf("%s:%s:%s", req.Inputs.ProjectName, req.Inputs.NotificationType, req.Inputs.NotificationName)

	if req.DryRun {
		return infer.CreateResponse[ProjectNotificationState]{
			ID:     id,
			Output: ProjectNotificationState{ProjectNotificationArgs: req.Inputs},
		}, nil
	}

	if err := c.AddNotificationToProject(ctx, req.Inputs.ProjectName, notifType, req.Inputs.NotificationName); err != nil {
		return infer.CreateResponse[ProjectNotificationState]{}, fmt.Errorf("failed to add notification to project: %w", err)
	}

	// Look up project ID — surface errors rather than silently persisting projectId=0
	info, err := c.CheckProjectNotificationExists(ctx, req.Inputs.ProjectName, req.Inputs.NotificationType, req.Inputs.NotificationName)
	if err != nil {
		return infer.CreateResponse[ProjectNotificationState]{}, fmt.Errorf("failed to verify project notification after create: %w", err)
	}
	if info == nil || !info.Exists {
		return infer.CreateResponse[ProjectNotificationState]{}, fmt.Errorf("project notification was created but could not be verified")
	}

	return infer.CreateResponse[ProjectNotificationState]{
		ID: id,
		Output: ProjectNotificationState{
			ProjectNotificationArgs: req.Inputs,
			ProjectID:               info.ProjectID,
		},
	}, nil
}

// No Update — all fields are forceNew, so any change triggers replace.

func (r *ProjectNotification) Delete(ctx context.Context, req infer.DeleteRequest[ProjectNotificationState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	notifType := strings.ToUpper(req.State.NotificationType)
	if err := c.RemoveNotificationFromProject(ctx, req.State.ProjectName, notifType, req.State.NotificationName); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to remove notification from project: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *ProjectNotification) Read(ctx context.Context, req infer.ReadRequest[ProjectNotificationArgs, ProjectNotificationState]) (infer.ReadResponse[ProjectNotificationArgs, ProjectNotificationState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Import ID: {project_name}:{notification_type}:{notification_name}
	parts := strings.SplitN(req.ID, ":", 3)
	var projectName, notificationType, notificationName string

	if len(parts) == 3 {
		projectName = parts[0]
		notificationType = parts[1]
		notificationName = parts[2]
	} else {
		projectName = req.State.ProjectName
		notificationType = req.State.NotificationType
		notificationName = req.State.NotificationName
	}

	info, err := c.CheckProjectNotificationExists(ctx, projectName, notificationType, notificationName)
	if err != nil {
		return infer.ReadResponse[ProjectNotificationArgs, ProjectNotificationState]{}, fmt.Errorf("failed to read project notification: %w", err)
	}

	if info == nil || !info.Exists {
		// Resource was deleted out-of-band — return empty response so Pulumi removes it from state
		return infer.ReadResponse[ProjectNotificationArgs, ProjectNotificationState]{}, nil
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

	return infer.ReadResponse[ProjectNotificationArgs, ProjectNotificationState]{
		ID:     req.ID,
		Inputs: args,
		State:  st,
	}, nil
}

func (r *ProjectNotification) Diff(ctx context.Context, req infer.DiffRequest[ProjectNotificationArgs, ProjectNotificationState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// All fields are forceNew for project notifications
	if req.Inputs.ProjectName != req.State.ProjectName {
		diff["projectName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if !strings.EqualFold(req.Inputs.NotificationType, req.State.NotificationType) {
		diff["notificationType"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.NotificationName != req.State.NotificationName {
		diff["notificationName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
