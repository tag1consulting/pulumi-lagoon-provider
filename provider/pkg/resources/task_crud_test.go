package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestTaskCreate_Command(t *testing.T) {
	cmd := "drush cr"
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, input map[string]any) (*client.TaskDefinition, error) {
			if input["type"] != "COMMAND" {
				t.Errorf("expected type COMMAND, got %v", input["type"])
			}
			if input["command"] != "drush cr" {
				t.Errorf("expected command 'drush cr', got %v", input["command"])
			}
			return &client.TaskDefinition{ID: 20, Name: "clear-cache", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	resp, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "clear-cache", Type: "command", Service: "cli", Command: &cmd},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "20" {
		t.Errorf("expected ID '20', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 20 {
		t.Errorf("expected LagoonID 20, got %d", resp.Output.LagoonID)
	}
}

func TestTaskCreate_Image(t *testing.T) {
	img := "myregistry/task:latest"
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, input map[string]any) (*client.TaskDefinition, error) {
			if input["type"] != "IMAGE" {
				t.Errorf("expected type IMAGE, got %v", input["type"])
			}
			if input["image"] != "myregistry/task:latest" {
				t.Errorf("expected image, got %v", input["image"])
			}
			return &client.TaskDefinition{ID: 21, Name: "image-task"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	resp, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "image-task", Type: "image", Service: "cli", Image: &img},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "21" {
		t.Errorf("expected ID '21', got %q", resp.ID)
	}
}

func TestTaskCreate_InvalidType(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "bad-task", Type: "unknown", Service: "cli"},
	})
	if err == nil {
		t.Fatal("expected error for unknown task type")
	}
}

func TestTaskCreate_CommandMissing(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "bad-task", Type: "command", Service: "cli"},
	})
	if err == nil {
		t.Fatal("expected error when command is missing for type=command")
	}
}

func TestTaskCreate_ImageMissing(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "bad-task", Type: "image", Service: "cli"},
	})
	if err == nil {
		t.Fatal("expected error when image is missing for type=image")
	}
}

func TestTaskCreate_SystemWide(t *testing.T) {
	cmd := "echo hello"
	sw := true
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, input map[string]any) (*client.TaskDefinition, error) {
			if input["systemWide"] != true {
				t.Errorf("expected systemWide=true, got %v", input["systemWide"])
			}
			return &client.TaskDefinition{ID: 22, Name: "sys-task"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "sys-task", Type: "command", Service: "cli", Command: &cmd, SystemWide: &sw},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestTaskCreate_WithArguments(t *testing.T) {
	cmd := "echo"
	args := []TaskArgumentInput{
		{Name: "message", DisplayName: "Message", Type: "string"},
	}
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, input map[string]any) (*client.TaskDefinition, error) {
			rawArgs, ok := input["advancedTaskDefinitionArguments"].([]map[string]any)
			if !ok || len(rawArgs) != 1 {
				t.Errorf("expected 1 argument in input, got %v", input["advancedTaskDefinitionArguments"])
			} else {
				if rawArgs[0]["type"] != "STRING" {
					t.Errorf("expected argument type uppercased to STRING, got %v", rawArgs[0]["type"])
				}
			}
			return &client.TaskDefinition{ID: 23, Name: "arg-task"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "arg-task", Type: "command", Service: "cli", Command: &cmd, Arguments: &args},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestTaskCreate_WithPermission(t *testing.T) {
	cmd := "echo"
	perm := "developer"
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, input map[string]any) (*client.TaskDefinition, error) {
			if input["permission"] != "DEVELOPER" {
				t.Errorf("expected permission uppercased to DEVELOPER, got %v", input["permission"])
			}
			return &client.TaskDefinition{ID: 24, Name: "perm-task"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "perm-task", Type: "command", Service: "cli", Command: &cmd, Permission: &perm},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}

func TestTaskCreate_DryRun(t *testing.T) {
	cmd := "echo"
	called := false
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, _ map[string]any) (*client.TaskDefinition, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	resp, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "test-task", Type: "command", Service: "cli", Command: &cmd},
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

func TestTaskDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteTaskDefinitionFn: func(_ context.Context, taskID int) error {
			if taskID != 20 {
				t.Errorf("expected taskID 20, got %d", taskID)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Delete(ctx, infer.DeleteRequest[TaskState]{
		ID:    "20",
		State: TaskState{TaskArgs: TaskArgs{Name: "clear-cache"}, LagoonID: 20},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTaskDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteTaskDefinitionFn: func(_ context.Context, _ int) error {
			return &client.LagoonNotFoundError{ResourceType: "TaskDefinition", Identifier: "20"}
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Delete(ctx, infer.DeleteRequest[TaskState]{
		ID:    "20",
		State: TaskState{TaskArgs: TaskArgs{Name: "clear-cache"}, LagoonID: 20},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestTaskRead_HappyPath(t *testing.T) {
	sw := true
	mock := &mockLagoonClient{
		getTaskDefinitionByIDFn: func(_ context.Context, taskID int) (*client.TaskDefinition, error) {
			perm := "DEVELOPER"
			return &client.TaskDefinition{
				ID:          taskID,
				Name:        "clear-cache",
				Type:        "COMMAND",
				Service:     "cli",
				Command:     "drush cr",
				Permission:  perm,
				Description: "Clears the cache",
				Created:     "2024-01-01",
				Arguments: []client.TaskArgument{
					{ID: 1, Name: "site", DisplayName: "Site", Type: "STRING"},
				},
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	resp, err := r.Read(ctx, infer.ReadRequest[TaskArgs, TaskState]{
		ID:    "20",
		State: TaskState{TaskArgs: TaskArgs{SystemWide: &sw}},
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "20" {
		t.Errorf("expected ID '20', got %q", resp.ID)
	}
	if resp.Inputs.Type != "command" {
		t.Errorf("expected lowercased type 'command', got %q", resp.Inputs.Type)
	}
	if resp.Inputs.Command == nil || *resp.Inputs.Command != "drush cr" {
		t.Errorf("expected Command 'drush cr'")
	}
	if resp.Inputs.Permission == nil || *resp.Inputs.Permission != "developer" {
		t.Errorf("expected lowercased Permission 'developer'")
	}
	if resp.Inputs.Description == nil || *resp.Inputs.Description != "Clears the cache" {
		t.Errorf("expected Description")
	}
	if resp.Inputs.Arguments == nil || len(*resp.Inputs.Arguments) != 1 {
		t.Errorf("expected 1 argument")
	} else {
		arg := (*resp.Inputs.Arguments)[0]
		if arg.Type != "string" {
			t.Errorf("expected lowercased arg type 'string', got %q", arg.Type)
		}
	}
	// SystemWide should be carried from state
	if resp.Inputs.SystemWide == nil || *resp.Inputs.SystemWide != true {
		t.Errorf("expected SystemWide carried from state")
	}
}

func TestTaskRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getTaskDefinitionByIDFn: func(_ context.Context, _ int) (*client.TaskDefinition, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "TaskDefinition", Identifier: "20"}
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	resp, err := r.Read(ctx, infer.ReadRequest[TaskArgs, TaskState]{ID: "20"})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %q", resp.ID)
	}
}

func TestTaskRead_InvalidID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Read(ctx, infer.ReadRequest[TaskArgs, TaskState]{ID: "not-a-number"})
	if err == nil {
		t.Fatal("expected error for non-numeric ID")
	}
}

func TestTaskCreate_ProjectScoped(t *testing.T) {
	cmd := "echo"
	projID := 5
	mock := &mockLagoonClient{
		createTaskDefinitionFn: func(_ context.Context, input map[string]any) (*client.TaskDefinition, error) {
			if input["project"] != 5 {
				t.Errorf("expected project 5, got %v", input["project"])
			}
			return &client.TaskDefinition{ID: 25, Name: "proj-task"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Task{}
	_, err := r.Create(ctx, infer.CreateRequest[TaskArgs]{
		Inputs: TaskArgs{Name: "proj-task", Type: "command", Service: "cli", Command: &cmd, ProjectID: &projID},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
}
