package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

func TestProjectCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, input map[string]any) (*client.Project, error) {
			return &client.Project{ID: 42, Name: "myproject", Created: "2024-01-01", PublicKey: "ssh-rsa AAAA... lagoon"}, nil
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
	if resp.Output.PublicKey != "ssh-rsa AAAA... lagoon" {
		t.Errorf("expected PublicKey to be populated from API response, got %q", resp.Output.PublicKey)
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

func TestProjectCreate_DuplicateEntry_ReturnsError(t *testing.T) {
	mock := &mockLagoonClient{
		createProjectFn: func(_ context.Context, _ map[string]any) (*client.Project, error) {
			return nil, &client.LagoonAPIError{Message: "Duplicate entry"}
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	_, err := r.Create(ctx, infer.CreateRequest[ProjectArgs]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", DeploytargetID: 1},
	})
	if err == nil {
		t.Fatal("expected error when project already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "pulumi import") {
		t.Errorf("expected 'pulumi import' hint in error, got: %v", err)
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
			return &client.Project{ID: id, Name: "myproject", PublicKey: "ssh-rsa AAAA... lagoon"}, nil
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
	if resp.Output.PublicKey != "ssh-rsa AAAA... lagoon" {
		t.Errorf("expected PublicKey from API, got %q", resp.Output.PublicKey)
	}
}

func TestProjectUpdate_PublicKeyPreserved_WhenAPIOmits(t *testing.T) {
	mock := &mockLagoonClient{
		updateProjectFn: func(_ context.Context, id int, _ map[string]any) (*client.Project, error) {
			// Some Lagoon API versions or permission scopes may omit publicKey on update
			return &client.Project{ID: id, Name: "myproject"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Update(ctx, infer.UpdateRequest[ProjectArgs, ProjectState]{
		ID:     "42",
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@new.com:repo.git", DeploytargetID: 2},
		State: ProjectState{
			ProjectArgs: ProjectArgs{Name: "myproject", GitURL: "git@old.com:repo.git", DeploytargetID: 1},
			LagoonID:    42,
			PublicKey:   "ssh-rsa PRIOR... lagoon",
			Created:     "2024-01-01",
		},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Output.PublicKey != "ssh-rsa PRIOR... lagoon" {
		t.Errorf("expected PublicKey to fall back to prior state when API omits, got %q", resp.Output.PublicKey)
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

// fastProjectDeleteWait shortens the project delete wait so tests finish quickly.
// It changes package-level vars, so do not use it in tests that call t.Parallel.
func fastProjectDeleteWait(t *testing.T, timeout time.Duration) {
	t.Helper()
	oldTimeout, oldBase, oldMax, oldStall := projectDeleteTimeout, projectDeleteBaseDelay, projectDeleteMaxDelay, projectDeleteStallTimeout
	projectDeleteTimeout, projectDeleteBaseDelay, projectDeleteMaxDelay = timeout, time.Millisecond, 2*time.Millisecond
	projectDeleteStallTimeout = timeout // no stall detection before the overall timeout unless a test lowers it
	t.Cleanup(func() {
		projectDeleteTimeout, projectDeleteBaseDelay, projectDeleteMaxDelay = oldTimeout, oldBase, oldMax
		projectDeleteStallTimeout = oldStall
	})
}

var errEnvsRemaining = &client.LagoonAPIError{Message: "Unable to delete project, there are existing environments that need to be removed first"}

func projectDeleteReq() infer.DeleteRequest[ProjectState] {
	return infer.DeleteRequest[ProjectState]{
		ID:    "42",
		State: ProjectState{ProjectArgs: ProjectArgs{Name: "myproject"}, LagoonID: 42},
	}
}

func TestProjectDelete_RetriesWhileEnvironmentsRemain(t *testing.T) {
	fastProjectDeleteWait(t, time.Second)
	calls := 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			calls++
			if calls <= 2 {
				return errEnvsRemaining
			}
			return nil
		},
	}
	if _, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq()); err != nil {
		t.Fatalf("Delete should succeed after retries: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 delete calls, got %d", calls)
	}
}

func TestProjectDelete_NotFoundDuringRetryIsSuccess(t *testing.T) {
	fastProjectDeleteWait(t, time.Second)
	calls := 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			calls++
			if calls <= 2 {
				return errEnvsRemaining
			}
			return &client.LagoonNotFoundError{ResourceType: "Project", Identifier: "myproject"}
		},
	}
	if _, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq()); err != nil {
		t.Fatalf("NotFound during retry should succeed: %v", err)
	}
}

func TestProjectDelete_TimesOutWhileEnvironmentsRemain(t *testing.T) {
	fastProjectDeleteWait(t, 30*time.Millisecond)
	calls := 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			calls++
			return errEnvsRemaining
		},
	}
	_, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "still has active environments") {
		t.Errorf("error should explain active environments remain, got: %v", err)
	}
	if !client.IsEnvironmentsRemaining(err) {
		t.Errorf("error should wrap the last API error, got: %v", err)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 delete calls, got %d", calls)
	}
}

func TestProjectDelete_ContextCancelledDuringWait(t *testing.T) {
	fastProjectDeleteWait(t, time.Hour)
	projectDeleteBaseDelay, projectDeleteMaxDelay = time.Minute, time.Minute
	ctx, cancel := context.WithCancel(context.Background())
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			cancel()
			return errEnvsRemaining
		},
	}
	start := time.Now()
	_, err := (&Project{}).Delete(withTestClient(ctx, mock), projectDeleteReq())
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if !strings.Contains(err.Error(), "myproject") || !strings.Contains(err.Error(), "existing environments") {
		t.Errorf("cancel error should name the project and keep the last Lagoon error, got: %v", err)
	}
	if time.Since(start) > 5*time.Second {
		t.Errorf("cancel should end the wait quickly, took %s", time.Since(start))
	}
}

func TestProjectDelete_FailsEarlyWhenEnvironmentsDoNotGoAway(t *testing.T) {
	fastProjectDeleteWait(t, 30*time.Second)
	projectDeleteStallTimeout = 40 * time.Millisecond
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error { return errEnvsRemaining },
		getProjectEnvironmentNamesFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"main", "pr-7"}, nil
		},
	}
	start := time.Now()
	_, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq())
	if err == nil {
		t.Fatal("expected an error")
	}
	if time.Since(start) > 5*time.Second {
		t.Errorf("delete should fail after the stall timeout, not the overall timeout, took %s", time.Since(start))
	}
	for _, want := range []string{"main", "pr-7", "has not gone down", "does not manage"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should contain %q, got: %v", want, err)
		}
	}
	if !client.IsEnvironmentsRemaining(err) {
		t.Errorf("error should wrap the last API error, got: %v", err)
	}
}

func TestProjectDelete_KeepsWaitingWhileEnvironmentsDisappear(t *testing.T) {
	fastProjectDeleteWait(t, 30*time.Second)
	projectDeleteStallTimeout = 40 * time.Millisecond
	lists := [][]string{{"a", "b", "c"}, {"a", "b"}, {"a"}, {}}
	queries, deletes := 0, 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			deletes++
			if deletes <= len(lists) {
				return errEnvsRemaining
			}
			return nil
		},
		getProjectEnvironmentNamesFn: func(_ context.Context, _ string) ([]string, error) {
			l := lists[queries]
			queries++
			time.Sleep(25 * time.Millisecond) // each step is slower than half of the stall timeout
			return l, nil
		},
	}
	if _, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq()); err != nil {
		t.Fatalf("delete should succeed while environments keep disappearing: %v", err)
	}
	if deletes != len(lists)+1 {
		t.Errorf("expected %d delete calls, got %d", len(lists)+1, deletes)
	}
}

func TestProjectDelete_EnvironmentQueryFailureIsNotFatal(t *testing.T) {
	fastProjectDeleteWait(t, time.Second)
	deletes := 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			deletes++
			if deletes <= 2 {
				return errEnvsRemaining
			}
			return nil
		},
		getProjectEnvironmentNamesFn: func(_ context.Context, _ string) ([]string, error) {
			return nil, &client.LagoonAPIError{Message: "boom"}
		},
	}
	if _, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq()); err != nil {
		t.Fatalf("a failing environment query must not fail the delete: %v", err)
	}
	if deletes != 3 {
		t.Errorf("expected 3 delete calls, got %d", deletes)
	}
}

func TestProjectDelete_TimeoutErrorListsRemainingEnvironments(t *testing.T) {
	fastProjectDeleteWait(t, 30*time.Millisecond)
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error { return errEnvsRemaining },
		getProjectEnvironmentNamesFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"main"}, nil
		},
	}
	_, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq())
	if err == nil || !strings.Contains(err.Error(), "main") {
		t.Fatalf("timeout error should list the remaining environments, got: %v", err)
	}
}

func TestProjectDelete_NoEnvironmentsListedDoesNotCountAsStall(t *testing.T) {
	// An empty list after a refusal means the environments finished removing in between.
	fastProjectDeleteWait(t, time.Second)
	projectDeleteStallTimeout = 5 * time.Millisecond
	deletes := 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			deletes++
			if deletes <= 6 {
				time.Sleep(3 * time.Millisecond)
				return errEnvsRemaining
			}
			return nil
		},
	}
	if _, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq()); err != nil {
		t.Fatalf("an empty environment list must not trigger the stall error: %v", err)
	}
}

func TestNextProjectDeleteDelay(t *testing.T) {
	old := projectDeleteMaxDelay
	projectDeleteMaxDelay = 30 * time.Second
	t.Cleanup(func() { projectDeleteMaxDelay = old })

	tests := []struct {
		name     string
		in       time.Duration
		min, max time.Duration
	}{
		{"doubles below the cap", 5 * time.Second, 10 * time.Second, 12500 * time.Millisecond},
		{"capped at the max plus jitter", 20 * time.Second, 30 * time.Second, 37500 * time.Millisecond},
		{"already at the cap", 30 * time.Second, 30 * time.Second, 37500 * time.Millisecond},
		{"tiny delay does not panic", time.Nanosecond, 2 * time.Nanosecond, 2 * time.Nanosecond},
		{"zero delay does not panic", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 200; i++ {
				got := nextProjectDeleteDelay(tt.in)
				if got < tt.min || got > tt.max {
					t.Fatalf("nextProjectDeleteDelay(%s) = %s, want within [%s, %s]", tt.in, got, tt.min, tt.max)
				}
			}
		})
	}
}

func TestProjectDelete_OtherAPIErrorNotRetried(t *testing.T) {
	fastProjectDeleteWait(t, time.Second)
	calls := 0
	mock := &mockLagoonClient{
		deleteProjectFn: func(_ context.Context, _ string) error {
			calls++
			return &client.LagoonAPIError{Message: "forbidden"}
		},
	}
	if _, err := (&Project{}).Delete(testCtx(mock), projectDeleteReq()); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("non-matching error must not be retried, got %d calls", calls)
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
				PublicKey:             "ssh-rsa AAAA... lagoon",
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
	if resp.State.PublicKey != "ssh-rsa AAAA... lagoon" {
		t.Errorf("expected PublicKey to be read from API, got %q", resp.State.PublicKey)
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

func TestProjectRead_OptionalFields(t *testing.T) {
	autoIdle := 1
	storagecalc := 0
	mock := &mockLagoonClient{
		getProjectByIDFn: func(_ context.Context, id int) (*client.Project, error) {
			return &client.Project{
				ID:                      id,
				Name:                    "myproject",
				GitURL:                  "git@example.com:repo.git",
				OpenshiftID:             3,
				ProductionEnvironment:   "main",
				Branches:                "main|develop",
				Pullrequests:            "true",
				OpenshiftProjectPattern: "${project}-${environment}",
				AutoIdle:                &autoIdle,
				StorageCalc:             &storagecalc,
				Created:                 "2024-01-01",
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &Project{}
	resp, err := r.Read(ctx, infer.ReadRequest[ProjectArgs, ProjectState]{ID: "42"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if resp.Inputs.OpenshiftProjectPattern == nil || *resp.Inputs.OpenshiftProjectPattern != "${project}-${environment}" {
		t.Errorf("expected OpenshiftProjectPattern '${project}-${environment}', got %v", resp.Inputs.OpenshiftProjectPattern)
	}
	if resp.Inputs.AutoIdle == nil || *resp.Inputs.AutoIdle != 1 {
		t.Errorf("expected AutoIdle 1, got %v", resp.Inputs.AutoIdle)
	}
	if resp.Inputs.StorageCalc == nil || *resp.Inputs.StorageCalc != 0 {
		t.Errorf("expected StorageCalc 0, got %v", resp.Inputs.StorageCalc)
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

// ==================== Diff: DeployTargetConfig placeholder (issue #270) ====================

// TestProjectDiff_DeployTargetPlaceholder_NoDiff reproduces the scenario from
// issue #270: once a DeployTargetConfig is attached to a project, the Lagoon
// API overwrites the project's branches/pullrequests columns with a fixed
// placeholder string. A refresh then picks up that placeholder as state, and
// the next preview must not diff it against the user's actual desired regex.
func TestProjectDiff_DeployTargetPlaceholder_NoDiff(t *testing.T) {
	desiredBranches := "^(main|develop|feature/.*)$"
	desiredPRs := ".*"
	placeholder := deployTargetConfigPlaceholder

	r := &Project{}
	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{
		Inputs: ProjectArgs{
			Name:         "myproject",
			GitURL:       "git@example.com:repo.git",
			Branches:     &desiredBranches,
			Pullrequests: &desiredPRs,
		},
		State: ProjectState{
			ProjectArgs: ProjectArgs{
				Name:         "myproject",
				GitURL:       "git@example.com:repo.git",
				Branches:     &placeholder,
				Pullrequests: &placeholder,
			},
		},
	})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Errorf("expected no changes when state holds the DeployTargetConfig placeholder; got diff: %+v", resp.DetailedDiff)
	}
	if _, ok := resp.DetailedDiff["branches"]; ok {
		t.Error("expected 'branches' to be absent from DetailedDiff when state is the DeployTargetConfig placeholder")
	}
	if _, ok := resp.DetailedDiff["pullrequests"]; ok {
		t.Error("expected 'pullrequests' to be absent from DetailedDiff when state is the DeployTargetConfig placeholder")
	}
}

// TestProjectDiff_RealDrift_StillDetected ensures the placeholder special-case
// doesn't mask genuine drift: when the state holds anything other than the
// exact placeholder string, a real difference must still be reported.
func TestProjectDiff_RealDrift_StillDetected(t *testing.T) {
	desired := "^(main|develop)$"
	stateVal := "^(main)$"

	r := &Project{}
	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", Branches: &desired},
		State: ProjectState{
			ProjectArgs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", Branches: &stateVal},
		},
	})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes when branches genuinely differ and state is not the DeployTargetConfig placeholder")
	}
	if _, ok := resp.DetailedDiff["branches"]; !ok {
		t.Error("expected 'branches' in DetailedDiff for genuine drift")
	}
}

// TestProjectDiff_PlaceholderLookalike_StillDetected guards against a
// too-broad match: a value that merely resembles the placeholder (e.g. a
// user's own regex happens to equal it, or the API string changes slightly in
// a future Lagoon version) must not be silently swallowed. Only an exact match
// of the documented placeholder is suppressed.
func TestProjectDiff_PlaceholderLookalike_StillDetected(t *testing.T) {
	desired := "^(main)$"
	lookalike := "This project is configured with DeployTarget" // missing trailing "s"

	r := &Project{}
	resp, err := r.Diff(context.Background(), infer.DiffRequest[ProjectArgs, ProjectState]{
		Inputs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", Branches: &desired},
		State: ProjectState{
			ProjectArgs: ProjectArgs{Name: "myproject", GitURL: "git@example.com:repo.git", Branches: &lookalike},
		},
	})
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("expected changes when state is merely similar to, but not exactly, the DeployTargetConfig placeholder")
	}
}

// TestDiffProject_DispatchesToCustomDiff is a regression test for issue #270,
// following the same framework-dispatch pattern established for issue #267
// (see config_test.go's TestDiffConfig_DispatchesToCustomDiff): it builds the
// provider exactly as production does (infer.Resource(&resources.Project{}))
// and calls the resulting provider's real Diff RPC entrypoint with property.Map
// state/inputs, rather than calling (*Project).Diff directly. This exercises
// the actual interface-dispatch path the Pulumi engine uses on every
// refresh/preview/up, not just the Go method in isolation.
func TestDiffProject_DispatchesToCustomDiff(t *testing.T) {
	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&Project{})).
		Build()
	if err != nil {
		t.Fatalf("failed to build provider: %v", err)
	}
	if prov.Diff == nil {
		t.Fatal("provider.Diff is nil; infer.Resource did not wire up Diff")
	}

	req := p.DiffRequest{
		ID:  "1",
		Urn: "urn:pulumi:test::test::lagoon:lagoon:Project::myproject",
		State: property.NewMap(map[string]property.Value{
			"name":           property.New("myproject"),
			"gitUrl":         property.New("git@example.com:repo.git"),
			"deploytargetId": property.New(1.0),
			"branches":       property.New(deployTargetConfigPlaceholder),
			"pullrequests":   property.New(deployTargetConfigPlaceholder),
			"lagoonId":       property.New(1.0),
			"publicKey":      property.New(""),
			"created":        property.New(""),
		}),
		Inputs: property.NewMap(map[string]property.Value{
			"name":           property.New("myproject"),
			"gitUrl":         property.New("git@example.com:repo.git"),
			"deploytargetId": property.New(1.0),
			"branches":       property.New("^(main|develop|feature/.*)$"),
			"pullrequests":   property.New(".*"),
		}),
	}

	resp, err := prov.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff returned an error: %v", err)
	}
	for field, d := range resp.DetailedDiff {
		if d.Kind == p.UpdateReplace || d.Kind == p.AddReplace || d.Kind == p.DeleteReplace {
			t.Errorf("Diff marked field %q as %v for a DeployTargetConfig-placeholder-only difference; "+
				"expected no diff at all. Full DetailedDiff: %+v", field, d.Kind, resp.DetailedDiff)
		}
	}
	if resp.HasChanges {
		t.Errorf("expected no changes through the real dispatch path when state holds the DeployTargetConfig "+
			"placeholder for branches/pullrequests; got: %+v", resp.DetailedDiff)
	}
}
