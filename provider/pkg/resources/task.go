package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

// Task manages a Lagoon advanced task definition.
type Task struct{}

// TaskArgumentInput represents an argument definition for a task.
type TaskArgumentInput struct {
	Name        string `pulumi:"name"`
	DisplayName string `pulumi:"displayName"`
	Type        string `pulumi:"type"`
}

type TaskArgs struct {
	Name             string              `pulumi:"name"`
	Type             string              `pulumi:"type"`
	Service          string              `pulumi:"service"`
	Command          *string             `pulumi:"command,optional"`
	Image            *string             `pulumi:"image,optional"`
	ProjectID        *int                `pulumi:"projectId,optional"`
	EnvironmentID    *int                `pulumi:"environmentId,optional"`
	GroupName        *string             `pulumi:"groupName,optional"`
	SystemWide       *bool               `pulumi:"systemWide,optional"`
	Description      *string             `pulumi:"description,optional"`
	Permission       *string             `pulumi:"permission,optional"`
	ConfirmationText *string             `pulumi:"confirmationText,optional"`
	Arguments        *[]TaskArgumentInput `pulumi:"arguments,optional"`
}

type TaskState struct {
	TaskArgs
	LagoonID int    `pulumi:"lagoonId"`
	Created  string `pulumi:"created"`
}

func (r *Task) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "Task")
	a.Describe(&r, "Manages a Lagoon advanced task definition.")
}

func (a *TaskArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The task definition name.")
	an.Describe(&a.Type, "Task type: 'command' or 'image'.")
	an.Describe(&a.Service, "Service container name to run the task in.")
	an.Describe(&a.Command, "Command to execute (required for 'command' type).")
	an.Describe(&a.Image, "Container image to run (required for 'image' type).")
	an.Describe(&a.ProjectID, "Project ID (for project-scoped tasks).")
	an.Describe(&a.EnvironmentID, "Environment ID (for environment-scoped tasks).")
	an.Describe(&a.GroupName, "Group name (for group-scoped tasks).")
	an.Describe(&a.SystemWide, "If true, task is available system-wide (platform admin only).")
	an.Describe(&a.Description, "Task description.")
	an.Describe(&a.Permission, "Permission level: 'guest', 'developer', or 'maintainer'.")
	an.Describe(&a.ConfirmationText, "Text to display for user confirmation.")
	an.Describe(&a.Arguments, "List of argument definitions for the task.")
}

func (r *Task) Create(ctx context.Context, name string, inputs TaskArgs, preview bool) (string, TaskState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"name":    inputs.Name,
		"type":    strings.ToUpper(inputs.Type),
		"service": inputs.Service,
	}
	setOptional(input, "command", inputs.Command)
	setOptional(input, "image", inputs.Image)
	setOptionalInt(input, "project", inputs.ProjectID)
	setOptionalInt(input, "environment", inputs.EnvironmentID)
	setOptional(input, "groupName", inputs.GroupName)
	if inputs.SystemWide != nil && *inputs.SystemWide {
		input["systemWide"] = true
	}
	setOptional(input, "description", inputs.Description)
	if inputs.Permission != nil {
		input["permission"] = strings.ToUpper(*inputs.Permission)
	}
	setOptional(input, "confirmationText", inputs.ConfirmationText)

	if inputs.Arguments != nil {
		args := make([]map[string]any, len(*inputs.Arguments))
		for i, arg := range *inputs.Arguments {
			args[i] = map[string]any{
				"name":        arg.Name,
				"displayName": arg.DisplayName,
				"type":        strings.ToUpper(arg.Type),
			}
		}
		input["advancedTaskDefinitionArguments"] = args
	}

	if preview {
		return "preview-id", TaskState{TaskArgs: inputs}, nil
	}

	td, err := client.CreateTaskDefinition(ctx, input)
	if err != nil {
		return "", TaskState{}, fmt.Errorf("failed to create task: %w", err)
	}

	return strconv.Itoa(td.ID), TaskState{
		TaskArgs: inputs,
		LagoonID: td.ID,
		Created:  td.Created,
	}, nil
}

// No Update — task definitions are replaced on any change since Lagoon API doesn't support update.

func (r *Task) Delete(ctx context.Context, id string, props TaskState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if err := client.DeleteTaskDefinition(ctx, props.LagoonID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *Task) Read(ctx context.Context, id string, inputs TaskArgs, state TaskState) (string, TaskArgs, TaskState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	tdID, err := strconv.Atoi(id)
	if err != nil {
		return "", TaskArgs{}, TaskState{}, fmt.Errorf("invalid task ID '%s': must be numeric", id)
	}

	td, err := client.GetTaskDefinitionByID(ctx, tdID)
	if err != nil {
		return "", TaskArgs{}, TaskState{}, fmt.Errorf("failed to read task: %w", err)
	}

	args := TaskArgs{
		Name:    td.Name,
		Type:    strings.ToLower(td.Type),
		Service: td.Service,
	}
	if td.Command != "" {
		args.Command = &td.Command
	}
	if td.Image != "" {
		args.Image = &td.Image
	}
	args.ProjectID = td.ProjectID
	args.EnvironmentID = td.EnvironmentID
	if td.GroupName != "" {
		args.GroupName = &td.GroupName
	}
	if td.Description != "" {
		args.Description = &td.Description
	}
	if td.Permission != "" {
		perm := strings.ToLower(td.Permission)
		args.Permission = &perm
	}
	if td.ConfirmationText != "" {
		args.ConfirmationText = &td.ConfirmationText
	}
	if len(td.Arguments) > 0 {
		taskArgs := make([]TaskArgumentInput, len(td.Arguments))
		for i, a := range td.Arguments {
			taskArgs[i] = TaskArgumentInput{
				Name:        a.Name,
				DisplayName: a.DisplayName,
				Type:        strings.ToLower(a.Type),
			}
		}
		args.Arguments = &taskArgs
	}

	st := TaskState{
		TaskArgs: args,
		LagoonID: td.ID,
		Created:  td.Created,
	}

	return id, args, st, nil
}

func (r *Task) Diff(ctx context.Context, id string, olds TaskState, news TaskArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// forceNew fields
	if !strings.EqualFold(news.Type, olds.Type) {
		diff["type"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrIntDiffers(news.ProjectID, olds.ProjectID) {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrIntDiffers(news.EnvironmentID, olds.EnvironmentID) {
		diff["environmentId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.GroupName, olds.GroupName) {
		diff["groupName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrBoolDiffers(news.SystemWide, olds.SystemWide) {
		diff["systemWide"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Since Lagoon doesn't support updating task definitions, all other changes also trigger replace
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.Service != olds.Service {
		diff["service"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.Command, olds.Command) {
		diff["command"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.Image, olds.Image) {
		diff["image"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.Description, olds.Description) {
		diff["description"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.Permission, olds.Permission) {
		diff["permission"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.ConfirmationText, olds.ConfirmationText) {
		diff["confirmationText"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
