package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// ==================== User ====================

func TestUserCreate_HappyPath(t *testing.T) {
	firstName := "Lagoon"
	lastName := "Admin"
	comment := "Bootstrap admin"
	mock := &mockLagoonClient{
		createUserFn: func(_ context.Context, email string, fn, ln, c *string) (*client.User, error) {
			if email != "admin@lagoon.example.com" {
				t.Errorf("unexpected email: %s", email)
			}
			return &client.User{ID: "uuid-1", Email: email, FirstName: "Lagoon", LastName: "Admin", Comment: "Bootstrap admin"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	resp, err := r.Create(ctx, infer.CreateRequest[UserArgs]{
		Inputs: UserArgs{Email: "admin@lagoon.example.com", FirstName: &firstName, LastName: &lastName, Comment: &comment},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "admin@lagoon.example.com" {
		t.Errorf("expected ID 'admin@lagoon.example.com', got %q", resp.ID)
	}
	if resp.Output.LagoonID != "uuid-1" {
		t.Errorf("expected LagoonID 'uuid-1', got %q", resp.Output.LagoonID)
	}
}

func TestUserCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		createUserFn: func(_ context.Context, _ string, _, _, _ *string) (*client.User, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	resp, err := r.Create(ctx, infer.CreateRequest[UserArgs]{
		Inputs: UserArgs{Email: "test@example.com"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("API should not be called during dry run")
	}
	if resp.ID != "preview-id" {
		t.Errorf("expected preview-id, got %q", resp.ID)
	}
}

func TestUserRead_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		getUserByEmailFn: func(_ context.Context, email string) (*client.User, error) {
			return &client.User{ID: "uuid-1", Email: email, FirstName: "Test", LastName: "User"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	resp, err := r.Read(ctx, infer.ReadRequest[UserArgs, UserState]{
		ID:    "test@example.com",
		State: UserState{UserArgs: UserArgs{Email: "test@example.com"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "test@example.com" {
		t.Errorf("expected ID 'test@example.com', got %q", resp.ID)
	}
	if resp.State.LagoonID != "uuid-1" {
		t.Errorf("expected LagoonID 'uuid-1', got %q", resp.State.LagoonID)
	}
}

func TestUserRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getUserByEmailFn: func(_ context.Context, email string) (*client.User, error) {
			return nil, &client.LagoonNotFoundError{ResourceType: "User", Identifier: email}
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	resp, err := r.Read(ctx, infer.ReadRequest[UserArgs, UserState]{
		ID:    "gone@example.com",
		State: UserState{UserArgs: UserArgs{Email: "gone@example.com"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for deleted resource, got %q", resp.ID)
	}
}

func TestUserUpdate_WithChanges(t *testing.T) {
	var gotPatch map[string]any
	mock := &mockLagoonClient{
		updateUserFn: func(_ context.Context, email string, patch map[string]any) (*client.User, error) {
			gotPatch = patch
			return &client.User{ID: "uuid-1", Email: email}, nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}

	oldFirst := "Old"
	newFirst := "New"
	resp, err := r.Update(ctx, infer.UpdateRequest[UserArgs, UserState]{
		Inputs: UserArgs{Email: "test@example.com", FirstName: &newFirst},
		State:  UserState{UserArgs: UserArgs{Email: "test@example.com", FirstName: &oldFirst}, LagoonID: "uuid-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPatch == nil {
		t.Fatal("expected API call with patch")
	}
	if gotPatch["firstName"] != "New" {
		t.Errorf("expected firstName=New in patch, got %v", gotPatch["firstName"])
	}
	if resp.Output.LagoonID != "uuid-1" {
		t.Errorf("expected preserved LagoonID, got %q", resp.Output.LagoonID)
	}
}

func TestUserUpdate_NoPatch(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		updateUserFn: func(_ context.Context, _ string, _ map[string]any) (*client.User, error) {
			called = true
			return nil, nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}

	first := "Same"
	_, err := r.Update(ctx, infer.UpdateRequest[UserArgs, UserState]{
		Inputs: UserArgs{Email: "test@example.com", FirstName: &first},
		State:  UserState{UserArgs: UserArgs{Email: "test@example.com", FirstName: &first}, LagoonID: "uuid-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("API should not be called when nothing changed")
	}
}

func TestUserDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		deleteUserFn: func(_ context.Context, email string) error {
			if email != "test@example.com" {
				t.Errorf("unexpected email: %s", email)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserState]{
		State: UserState{UserArgs: UserArgs{Email: "test@example.com"}, LagoonID: "uuid-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		deleteUserFn: func(_ context.Context, email string) error {
			return &client.LagoonNotFoundError{ResourceType: "User", Identifier: email}
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserState]{
		State: UserState{UserArgs: UserArgs{Email: "gone@example.com"}, LagoonID: "uuid-1"},
	})
	if err != nil {
		t.Error("delete of non-existent user should succeed (idempotent)")
	}
}

// ==================== UserGroupAssignment ====================

func TestUserGroupAssignmentCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		addUserToGroupFn: func(_ context.Context, email, groupName, role string) error {
			if email != "dev@example.com" {
				t.Errorf("unexpected email: %s", email)
			}
			if groupName != "project-mysite" {
				t.Errorf("unexpected group: %s", groupName)
			}
			if role != "DEVELOPER" {
				t.Errorf("unexpected role: %s", role)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	resp, err := r.Create(ctx, infer.CreateRequest[UserGroupAssignmentArgs]{
		Inputs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "project-mysite", Role: "developer"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "dev@example.com:project-mysite" {
		t.Errorf("unexpected ID: %s", resp.ID)
	}
	if resp.Output.Role != "DEVELOPER" {
		t.Errorf("expected normalized role DEVELOPER, got %q", resp.Output.Role)
	}
}

func TestUserGroupAssignmentCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addUserToGroupFn: func(_ context.Context, _, _, _ string) error {
			called = true
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	_, err := r.Create(ctx, infer.CreateRequest[UserGroupAssignmentArgs]{
		Inputs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "mygroup", Role: "GUEST"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("API should not be called during dry run")
	}
}

func TestUserGroupAssignmentCreate_InvalidRole(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	_, err := r.Create(ctx, infer.CreateRequest[UserGroupAssignmentArgs]{
		Inputs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "mygroup", Role: "SUPERADMIN"},
	})
	if err == nil {
		t.Error("expected error for invalid role")
	}
}

func TestUserGroupAssignmentRead_Found(t *testing.T) {
	mock := &mockLagoonClient{
		getUserGroupRolesFn: func(_ context.Context, email string) ([]client.UserGroupRole, error) {
			return []client.UserGroupRole{
				{Name: "project-mysite", Role: "DEVELOPER"},
				{Name: "project-other", Role: "GUEST"},
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	resp, err := r.Read(ctx, infer.ReadRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]{
		ID:    "dev@example.com:project-mysite",
		State: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "project-mysite", Role: "DEVELOPER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "dev@example.com:project-mysite" {
		t.Errorf("unexpected ID: %s", resp.ID)
	}
	if resp.State.Role != "DEVELOPER" {
		t.Errorf("expected role DEVELOPER, got %q", resp.State.Role)
	}
}

func TestUserGroupAssignmentRead_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		getUserGroupRolesFn: func(_ context.Context, _ string) ([]client.UserGroupRole, error) {
			return []client.UserGroupRole{
				{Name: "project-other", Role: "GUEST"},
			}, nil
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	resp, err := r.Read(ctx, infer.ReadRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]{
		ID:    "dev@example.com:project-mysite",
		State: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "project-mysite", Role: "DEVELOPER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for missing assignment, got %q", resp.ID)
	}
}

func TestUserGroupAssignmentUpdate_RoleChange(t *testing.T) {
	var gotRole string
	mock := &mockLagoonClient{
		addUserToGroupFn: func(_ context.Context, _, _, role string) error {
			gotRole = role
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	resp, err := r.Update(ctx, infer.UpdateRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]{
		Inputs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "mygroup", Role: "maintainer"},
		State:  UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "mygroup", Role: "DEVELOPER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotRole != "MAINTAINER" {
		t.Errorf("expected role MAINTAINER, got %q", gotRole)
	}
	if resp.Output.Role != "MAINTAINER" {
		t.Errorf("expected output role MAINTAINER, got %q", resp.Output.Role)
	}
}

func TestUserGroupAssignmentDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		removeUserFromGroupFn: func(_ context.Context, email, groupName string) error {
			if email != "dev@example.com" || groupName != "mygroup" {
				t.Errorf("unexpected args: %s, %s", email, groupName)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserGroupAssignmentState]{
		State: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "mygroup", Role: "DEVELOPER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserGroupAssignmentDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		removeUserFromGroupFn: func(_ context.Context, email, groupName string) error {
			return &client.LagoonNotFoundError{ResourceType: "UserGroupAssignment", Identifier: email + ":" + groupName}
		},
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserGroupAssignmentState]{
		State: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "dev@example.com", GroupName: "mygroup", Role: "DEVELOPER"}},
	})
	if err != nil {
		t.Error("delete of non-existent assignment should succeed (idempotent)")
	}
}

// ==================== UserPlatformRole ====================

func TestUserPlatformRoleCreate_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		addPlatformRoleToUserFn: func(_ context.Context, email, role string) error {
			if email != "admin@lagoon.example.com" {
				t.Errorf("unexpected email: %s", email)
			}
			if role != "OWNER" {
				t.Errorf("unexpected role: %s", role)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	resp, err := r.Create(ctx, infer.CreateRequest[UserPlatformRoleArgs]{
		Inputs: UserPlatformRoleArgs{UserEmail: "admin@lagoon.example.com", Role: "owner"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "admin@lagoon.example.com:OWNER" {
		t.Errorf("unexpected ID: %s", resp.ID)
	}
	if resp.Output.Role != "OWNER" {
		t.Errorf("expected normalized role OWNER, got %q", resp.Output.Role)
	}
}

func TestUserPlatformRoleCreate_DryRun(t *testing.T) {
	called := false
	mock := &mockLagoonClient{
		addPlatformRoleToUserFn: func(_ context.Context, _, _ string) error {
			called = true
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	_, err := r.Create(ctx, infer.CreateRequest[UserPlatformRoleArgs]{
		Inputs: UserPlatformRoleArgs{UserEmail: "admin@example.com", Role: "VIEWER"},
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("API should not be called during dry run")
	}
}

func TestUserPlatformRoleCreate_InvalidRole(t *testing.T) {
	mock := &mockLagoonClient{}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	_, err := r.Create(ctx, infer.CreateRequest[UserPlatformRoleArgs]{
		Inputs: UserPlatformRoleArgs{UserEmail: "admin@example.com", Role: "SUPERADMIN"},
	})
	if err == nil {
		t.Error("expected error for invalid platform role")
	}
}

func TestUserPlatformRoleRead_RolePresent(t *testing.T) {
	mock := &mockLagoonClient{
		getUserPlatformRolesFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"OWNER", "VIEWER"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	resp, err := r.Read(ctx, infer.ReadRequest[UserPlatformRoleArgs, UserPlatformRoleState]{
		ID:    "admin@example.com:OWNER",
		State: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "admin@example.com", Role: "OWNER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "admin@example.com:OWNER" {
		t.Errorf("unexpected ID: %s", resp.ID)
	}
	if resp.State.Role != "OWNER" {
		t.Errorf("expected role OWNER, got %q", resp.State.Role)
	}
}

func TestUserPlatformRoleRead_RoleAbsent(t *testing.T) {
	mock := &mockLagoonClient{
		getUserPlatformRolesFn: func(_ context.Context, _ string) ([]string, error) {
			return []string{"VIEWER"}, nil
		},
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	resp, err := r.Read(ctx, infer.ReadRequest[UserPlatformRoleArgs, UserPlatformRoleState]{
		ID:    "admin@example.com:OWNER",
		State: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "admin@example.com", Role: "OWNER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID for missing role, got %q", resp.ID)
	}
}

func TestUserPlatformRoleDelete_HappyPath(t *testing.T) {
	mock := &mockLagoonClient{
		removePlatformRoleFromUserFn: func(_ context.Context, email, role string) error {
			if email != "admin@example.com" || role != "OWNER" {
				t.Errorf("unexpected args: %s, %s", email, role)
			}
			return nil
		},
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserPlatformRoleState]{
		State: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "admin@example.com", Role: "OWNER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserPlatformRoleDelete_NotFound(t *testing.T) {
	mock := &mockLagoonClient{
		removePlatformRoleFromUserFn: func(_ context.Context, email, role string) error {
			return &client.LagoonNotFoundError{ResourceType: "UserPlatformRole", Identifier: email + ":" + role}
		},
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserPlatformRoleState]{
		State: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "admin@example.com", Role: "OWNER"}},
	})
	if err != nil {
		t.Error("delete of non-existent platform role should succeed (idempotent)")
	}
}
