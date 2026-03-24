package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
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

func (r *Task) Create(ctx context.Context, req infer.CreateRequest[TaskArgs]) (infer.CreateResponse[TaskState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Validate task type and type-specific required fields
	taskType := strings.ToLower(req.Inputs.Type)
	switch taskType {
	case "command":
		if req.Inputs.Command == nil || *req.Inputs.Command == "" {
			return infer.CreateResponse[TaskState]{}, fmt.Errorf("'command' is required when type is 'command'")
		}
	case "image":
		if req.Inputs.Image == nil || *req.Inputs.Image == "" {
			return infer.CreateResponse[TaskState]{}, fmt.Errorf("'image' is required when type is 'image'")
		}
	default:
		return infer.CreateResponse[TaskState]{}, fmt.Errorf("unknown task type %q: must be 'command' or 'image'", req.Inputs.Type)
	}

	input := map[string]any{
		"name":    req.Inputs.Name,
		"type":    strings.ToUpper(req.Inputs.Type),
		"service": req.Inputs.Service,
	}
	setOptional(input, "command", req.Inputs.Command)
	setOptional(input, "image", req.Inputs.Image)
	setOptionalInt(input, "project", req.Inputs.ProjectID)
	setOptionalInt(input, "environment", req.Inputs.EnvironmentID)
	setOptional(input, "groupName", req.Inputs.GroupName)
	if req.Inputs.SystemWide != nil && *req.Inputs.SystemWide {
		input["systemWide"] = true
	}
	setOptional(input, "description", req.Inputs.Description)
	if req.Inputs.Permission != nil {
		input["permission"] = strings.ToUpper(*req.Inputs.Permission)
	}
	setOptional(input, "confirmationText", req.Inputs.ConfirmationText)

	if req.Inputs.Arguments != nil {
		args := make([]map[string]any, len(*req.Inputs.Arguments))
		for i, arg := range *req.Inputs.Arguments {
			args[i] = map[string]any{
				"name":        arg.Name,
				"displayName": arg.DisplayName,
				"type":        strings.ToUpper(arg.Type),
			}
		}
		input["advancedTaskDefinitionArguments"] = args
	}

	if req.DryRun {
		return infer.CreateResponse[TaskState]{
			ID:     "preview-id",
			Output: TaskState{TaskArgs: req.Inputs},
		}, nil
	}

	td, err := client.CreateTaskDefinition(ctx, input)
	if err != nil {
		return infer.CreateResponse[TaskState]{}, fmt.Errorf("failed to create task: %w", err)
	}

	return infer.CreateResponse[TaskState]{
		ID: strconv.Itoa(td.ID),
		Output: TaskState{
			TaskArgs: req.Inputs,
			LagoonID: td.ID,
			Created:  td.Created,
		},
	}, nil
}

// No Update — task definitions are replaced on any change since Lagoon API doesn't support update.

func (r *Task) Delete(ctx context.Context, req infer.DeleteRequest[TaskState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	if err := c.DeleteTaskDefinition(ctx, req.State.LagoonID); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete task: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *Task) Read(ctx context.Context, req infer.ReadRequest[TaskArgs, TaskState]) (infer.ReadResponse[TaskArgs, TaskState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	tdID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.ReadResponse[TaskArgs, TaskState]{}, fmt.Errorf("invalid task ID '%s': must be numeric", req.ID)
	}

	td, err := c.GetTaskDefinitionByID(ctx, tdID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[TaskArgs, TaskState]{}, nil
		}
		return infer.ReadResponse[TaskArgs, TaskState]{}, fmt.Errorf("failed to read task: %w", err)
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
	// Carry forward systemWide from state — the API doesn't return it
	args.SystemWide = req.State.SystemWide
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

	return infer.ReadResponse[TaskArgs, TaskState]{
		ID:     req.ID,
		Inputs: args,
		State:  st,
	}, nil
}

func (r *Task) Diff(ctx context.Context, req infer.DiffRequest[TaskArgs, TaskState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// forceNew fields
	if !strings.EqualFold(req.Inputs.Type, req.State.Type) {
		diff["type"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrIntDiffers(req.Inputs.ProjectID, req.State.ProjectID) {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrIntDiffers(req.Inputs.EnvironmentID, req.State.EnvironmentID) {
		diff["environmentId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.GroupName, req.State.GroupName) {
		diff["groupName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrBoolDiffers(req.Inputs.SystemWide, req.State.SystemWide) {
		diff["systemWide"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Since Lagoon doesn't support updating task definitions, all other changes also trigger replace
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.Service != req.State.Service {
		diff["service"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.Command, req.State.Command) {
		diff["command"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.Image, req.State.Image) {
		diff["image"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.Description, req.State.Description) {
		diff["description"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	// Use case-insensitive comparison for permission (lowered on read, uppered on create)
	inPerm := strings.ToLower(ptrOrDefault(req.Inputs.Permission, ""))
	stPerm := strings.ToLower(ptrOrDefault(req.State.Permission, ""))
	if inPerm != stPerm {
		diff["permission"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.ConfirmationText, req.State.ConfirmationText) {
		diff["confirmationText"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if taskArgumentsDiffer(req.Inputs.Arguments, req.State.Arguments) {
		diff["arguments"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}

// taskArgumentsDiffer returns true if two optional TaskArgumentInput slices differ.
// Treats nil and empty slice as equivalent.
func taskArgumentsDiffer(a, b *[]TaskArgumentInput) bool {
	aLen := 0
	if a != nil {
		aLen = len(*a)
	}
	bLen := 0
	if b != nil {
		bLen = len(*b)
	}
	if aLen == 0 && bLen == 0 {
		return false
	}
	if aLen != bLen {
		return true
	}
	if len(*a) != len(*b) {
		return true
	}
	for i := range *a {
		ai, bi := (*a)[i], (*b)[i]
		if ai.Name != bi.Name || ai.DisplayName != bi.DisplayName || !strings.EqualFold(ai.Type, bi.Type) {
			return true
		}
	}
	return false
}
