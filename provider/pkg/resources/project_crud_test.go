package resources

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestProjectCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, input map[string]any) (*client.Project, error) {
			return &client.Project{ID: 42, Name: "myproject", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID != "42" {
		t.Errorf("expected ID '42', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 42 {
		t.Errorf("expected LagoonID 42, got %d", resp.Output.LagoonID)
	}
	if resp.Output.Created != "2024-01-01" {
		t.Errorf("expected Created '2024-01-01', got %q", resp.Output.Created)
	}
	if resp.Output.Name != "myproject" {
		t.Errorf("expected Name 'myproject', got %q", resp.Output.Name)
	}
}

func TestProjectCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, _ map[string]any) (*client.Project, error) {
			called = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
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
	if resp.Output.Name != "myproject" {
		t.Errorf("expected Name preserved in output, got %q", resp.Output.Name)
	}
}

func TestProjectCreate_DuplicateEntry_Adopts(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, _ map[string]any) (*client.Project, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry"}
		},
		getProjectByNameFn: func(_ context.Context, name string) (*client.Project, error) {
			return &client.Project{ID: 99, Name: name}, nil
		},
		updateProjectFn: func(_ context.Context, id int, _ map[string]any) (*client.Project, error) {
			return &client.Project{ID: id, Name: "myproject", Created: "2024-01-01"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
	})
	if err != nil {
		t.Fatalf("Create with duplicate should adopt: %v", err)
	}
	if resp.ID != "99" {
		t.Errorf("expected adopted ID '99', got %q", resp.ID)
	}
	if resp.Output.LagoonID != 99 {
		t.Errorf("expected LagoonID 99, got %d", resp.Output.LagoonID)
	}
}

func TestProjectCreate_DuplicateEntry_LookupFails(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, _ map[string]any) (*client.Project, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry"}
		},
		getProjectByNameFn: func(_ context.Context, _ string) (*client.Project, error) {
			return nil, fmt.Errorf("network error")
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
	})
	if err == nil {
		t.Fatal("expected error when lookup fails")
	}
}

func TestProjectCreate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, _ map[string]any) (*client.Project, error) {
			return nil, &client.LagoonAPIError{Message: "permission denied"}
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProjectUpdate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		updateProjectFn: func(_ context.Context, id int, _ map[string]any) (*client.Project, error) {
			return &client.Project{ID: id, Name: "myproject"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Update(ctx, infer.UpdateRequest[ProjectArgs, ProjectState]{
		ID:     "42",
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@new.com:repo.git", DeploytargetID: 2},
		State:  ProjectState{ProjectArgs: ProjectArgs{Name: "myproject", GitURL: "git@old.com:repo.git", DeploytargetID: 1}, LagoonID: 42, Created: "2024-01-01"},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.LagoonID != 42 {
		t.Errorf("expected LagoonID 42, got %d", resp.Output.LagoonID)
	}
	if resp.Output.Created != "2024-01-01" {
		t.Errorf("expected Created preserved, got %q", resp.Output.Created)
	}
	if resp.Output.GitURL != "git@new.com:repo.git" {
		t.Errorf("expected updated GitURL, got %q", resp.Output.GitURL)
	}
}

func TestProjectUpdate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		updateProjectFn: func(_ context.Context, _ int, _ map[string]any) (*client.Project, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Update(ctx, infer.UpdateRequest[ProjectArgs, ProjectState]{
		ID:     "42",
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@new.com:repo.git", DeploytargetID: 2},
		State:  ProjectState{ProjectArgs: ProjectArgs{Name: "myproject"}, LagoonID: 42, Created: "2024-01-01"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}
	if called {
		t.Error("API should not be called during DryRun")
	}
	if resp.Output.LagoonID != 42 {
		t.Errorf("expected LagoonID preserved")
	}
}

func TestProjectUpdate_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		updateProjectFn: func(_ context.Context, _ int, _ map[string]any) (*client.Project, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Update(ctx, infer.UpdateRequest[ProjectArgs, ProjectState]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
		State:  ProjectState{ProjectArgs: ProjectArgs{Name: "myproject"}, LagoonID: 42},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProjectDelete_HappyPath(t *testing.T) {
	deleted := ""
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, name string) error {
			deleted = name
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Delete(ctx, infer.DeleteRequest[ProjectState]{
		ID:    "42",
		State: ProjectState{ProjectArgs: ProjectArgs{Name: "myproject"}, LagoonID: 42},
	})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if deleted != "myproject" {
		t.Errorf("expected delete called with 'myproject', got %q", deleted)
	}
}

func TestProjectDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			return &client.LagoonNotFoundError{ResourceType: "Project", Identifier: "myproject"}
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Delete(ctx, infer.DeleteRequest[ProjectState]{
		ID:    "42",
		State: ProjectState{ProjectArgs: ProjectArgs{Name: "myproject"}, LagoonID: 42},
	})
	if err != nil {
		t.Fatalf("Delete NotFound should succeed: %v", err)
	}
}

func TestProjectDelete_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			return &client.LagoonAPIError{Message: "forbidden"}
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Delete(ctx, infer.DeleteRequest[ProjectState]{
		ID:    "42",
		State: ProjectState{ProjectArgs: ProjectArgs{Name: "myproject"}, LagoonID: 42},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProjectRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, id int) (*client.Project, error) {
			return &client.Project{
				ID:                    id,
				Name:                  "myproject",
				GitURL:                "git@example.com:repo.git",
				OpenshiftID:           3,
				ProductionEnvironment: "main",
				Branches:              "main|develop",
				Pullrequests:          "true",
				Created:               "2024-01-01",
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectArgs, ProjectState]{ID: "42"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.ID != "42" {
		t.Errorf("expected ID '42', got %q", resp.ID)
	}
	if resp.State.LagoonID != 42 {
		t.Errorf("expected LagoonID 42, got %d", resp.State.LagoonID)
	}
	if resp.Inputs.Name != "myproject" {
		t.Errorf("expected Name 'myproject', got %q", resp.Inputs.Name)
	}
	if resp.Inputs.DeploytargetID != 3 {
		t.Errorf("expected DeploytargetID 3, got %d", resp.Inputs.DeploytargetID)
	}
	if resp.Inputs.ProductionEnvironment == nil || *resp.Inputs.ProductionEnvironment != "main" {
		t.Errorf("expected ProductionEnvironment 'main'")
	}
	if resp.Inputs.Branches == nil || *resp.Inputs.Branches != "main|develop" {
		t.Errorf("expected Branches 'main|develop'")
	}
	if resp.Inputs.Pullrequests == nil || *resp.Inputs.Pullrequests != "true" {
		t.Errorf("expected Pullrequests 'true'")
	}
}

func TestProjectRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, id int) (*client.Project, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "Project", Identifier: "42"}
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectArgs, ProjectState]{ID: "42"})
	if err != nil {
		t.Fatalf("Read NotFound should not error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for deleted resource, got %q", resp.ID)
	}
}

func TestProjectRead_InvalidID(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Read(ctx, infer.ReadRequest[ProjectArgs, ProjectState]{ID: "not-a-number"})
	if err == nil {
		t.Fatal("expected error for non-numeric ID")
	}
}

func TestProjectCreate_DuplicateAdopt_UpdateFails(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, _ map[string]any) (*client.Project, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry 'myproject' for key 'name'"}
		},
		getProjectByNameFn: func(_ context.Context, name string) (*client.Project, error) {
			return &client.Project{ID: 99, Name: name}, nil
		},
		updateProjectFn: func(_ context.Context, _ int, _ map[string]any) (*client.Project, error) {
			return nil, fmt.Errorf("update failed")
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
	})
	if err == nil {
		t.Fatal("expected error when adopt-update fails")
	}
	if !strings.Contains(err.Error(), "failed to update") {
		t.Errorf("expected error to mention update failure, got: %v", err)
	}
}

func TestProjectRead_APIError(t *testing.T) {
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, _ int) (*client.Project, error) {
			return nil, fmt.Errorf("api error")
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Read(ctx, infer.ReadRequest[ProjectArgs, ProjectState]{ID: "42"})
	if err == nil {
		t.Fatal("expected error when API returns error")
	}
}

func TestProjectCreate_OptionalFields(t *testing.T) {
	prodEnv := "main"
	branches := "main|develop"
	autoIdle := 1
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, input map[string]any) (*client.Project, error) {
			// Verify optional fields were passed through
			if input["productionEnvironment"] != "main" {
				t.Errorf("expected productionEnvironment in input")
			}
			if input["branches"] != "main|develop" {
				t.Errorf("expected branches in input")
			}
			if input["autoIdle"] != 1 {
				t.Errorf("expected autoIdle in input")
			}
			return &client.Project{ID: 42, Name: "myproject"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{
			Name:                  "myproject",
			GitURL:                "git@example.com:repo.git",
			DeploytargetID:        1,
			ProductionEnvironment: &prodEnv,
			Branches:              &branches,
			AutoIdle:              &autoIdle,
		},
	})
	if err != nil {
		t.Fatalf("Create with optional fields failed: %v", err)
	}
}
