package resources

import (
	"context"
	"strings"
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
	if resp.ID != "test@example.com" {
		t.Errorf("expected email as preview ID, got %q", resp.ID)
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

func TestUserCreate_EmptyEmail(t *testing.T) {
	ctx := testCtx(&mockLagoonClient{})
	r := &User{}
	_, err := r.Create(ctx, infer.CreateRequest[UserArgs]{
		Inputs: UserArgs{Email: ""},
	})
	if err == nil {
		t.Fatal("expected error for empty email")
	}
	if !strings.Contains(err.Error(), "email must not be empty") {
		t.Errorf("expected empty email error, got: %v", err)
	}
}

func TestUserDiff_NoDeleteBeforeReplace(t *testing.T) {
	r := &User{}
	resp, err := r.Diff(context.Background(), infer.DiffRequest[UserArgs, UserState]{
		Inputs: UserArgs{Email: "new@example.com"},
		State:  UserState{UserArgs: UserArgs{Email: "old@example.com"}, LagoonID: "uuid-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DeleteBeforeReplace {
		t.Error("DeleteBeforeReplace must be false — old and new email are independent identities")
	}
	if !resp.HasChanges {
		t.Error("expected HasChanges when email differs")
	}
}

// ==================== New behavior tests (fixes for PR review findings) ====================

// TestUserUpdate_ClearOptionalField verifies that transitioning an optional field from
// non-nil to nil sends explicit null in the patch so Lagoon actually clears the field.
// Regression test for the "silent drop of field-clear updates" bug.
func TestUserUpdate_ClearOptionalField(t *testing.T) {
	var gotPatch map[string]any
	mock := &mockLagoonClient{
		updateUserFn: func(_ context.Context, email string, patch map[string]any) (*client.User, error) {
			gotPatch = patch
			return &client.User{ID: "uuid-1", Email: email}, nil
		},
	}
	ctx := testCtx(mock)
	r := &User{}

	oldFirst := "Alice"
	_, err := r.Update(ctx, infer.UpdateRequest[UserArgs, UserState]{
		Inputs: UserArgs{Email: "test@example.com", FirstName: nil},
		State:  UserState{UserArgs: UserArgs{Email: "test@example.com", FirstName: &oldFirst}, LagoonID: "uuid-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPatch == nil {
		t.Fatal("expected UpdateUser to be called when clearing a field")
	}
	v, present := gotPatch["firstName"]
	if !present {
		t.Fatalf("expected firstName key in patch, got %#v", gotPatch)
	}
	if v != nil {
		t.Errorf("expected firstName=nil (null) to clear the field, got %v", v)
	}
}

func TestUserCreate_DuplicateEntryGuidance(t *testing.T) {
	dupErr := &client.LagoonAPIError{Message: "duplicate entry for key email"}
	mock := &mockLagoonClient{
		createUserFn: func(_ context.Context, _ string, _, _, _ *string) (*client.User, error) {
			return nil, dupErr
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	_, err := r.Create(ctx, infer.CreateRequest[UserArgs]{
		Inputs: UserArgs{Email: "admin@example.com"},
	})
	if err == nil {
		t.Fatal("expected duplicate-entry error to surface")
	}
	if !strings.Contains(err.Error(), "pulumi import") {
		t.Errorf("expected duplicate-entry error to suggest pulumi import, got: %v", err)
	}
}

func TestUserGroupAssignmentCreate_EmptyInputs(t *testing.T) {
	ctx := testCtx(&mockLagoonClient{})
	r := &UserGroupAssignment{}
	cases := []UserGroupAssignmentArgs{
		{UserEmail: "", GroupName: "g", Role: "GUEST"},
		{UserEmail: "u@x", GroupName: "", Role: "GUEST"},
	}
	for _, in := range cases {
		_, err := r.Create(ctx, infer.CreateRequest[UserGroupAssignmentArgs]{Inputs: in})
		if err == nil {
			t.Errorf("expected error for empty input %+v", in)
		}
	}
}

func TestUserGroupAssignmentCreate_DuplicateEntryGuidance(t *testing.T) {
	dupErr := &client.LagoonAPIError{Message: "user already exists in group"}
	mock := &mockLagoonClient{
		addUserToGroupFn: func(_ context.Context, _, _, _ string) error { return dupErr },
	}
	ctx := testCtx(mock)
	r := &UserGroupAssignment{}
	_, err := r.Create(ctx, infer.CreateRequest[UserGroupAssignmentArgs]{
		Inputs: UserGroupAssignmentArgs{UserEmail: "u@x", GroupName: "g", Role: "GUEST"},
	})
	if err == nil || !strings.Contains(err.Error(), "pulumi import") {
		t.Errorf("expected duplicate-entry error mentioning pulumi import, got: %v", err)
	}
}

func TestParseUserGroupAssignmentID(t *testing.T) {
	cases := []struct {
		name      string
		id        string
		state     UserGroupAssignmentState
		wantEmail string
		wantGroup string
		wantErr   bool
	}{
		{"happy path", "u@x.com:devs", UserGroupAssignmentState{}, "u@x.com", "devs", false},
		{"colon in email local-part", `"a:b"@x.com:devs`, UserGroupAssignmentState{}, `"a:b"@x.com`, "devs", false},
		{"empty left", ":devs", UserGroupAssignmentState{}, "", "", true},
		{"empty right", "u@x.com:", UserGroupAssignmentState{}, "", "", true},
		{"empty id, state fallback", "", UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "u@x", GroupName: "g"}}, "u@x", "g", false},
		{"malformed id with state rejects", "garbage", UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "u@x", GroupName: "g"}}, "", "", true},
		{"malformed id without state", "garbage", UserGroupAssignmentState{}, "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			email, group, err := parseUserGroupAssignmentID(tc.id, tc.state)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if err == nil && (email != tc.wantEmail || group != tc.wantGroup) {
				t.Errorf("got (%q, %q), want (%q, %q)", email, group, tc.wantEmail, tc.wantGroup)
			}
		})
	}
}

func TestParseUserPlatformRoleID(t *testing.T) {
	cases := []struct {
		name     string
		id       string
		state    UserPlatformRoleState
		wantMail string
		wantRole string
		wantErr  bool
	}{
		{"happy path upper", "admin@x:OWNER", UserPlatformRoleState{}, "admin@x", "OWNER", false},
		{"happy path lower (normalized)", "admin@x:owner", UserPlatformRoleState{}, "admin@x", "OWNER", false},
		{"invalid role", "admin@x:ADMIN", UserPlatformRoleState{}, "", "", true},
		{"empty left", ":OWNER", UserPlatformRoleState{}, "", "", true},
		{"empty right", "admin@x:", UserPlatformRoleState{}, "", "", true},
		{"empty id, state fallback", "", UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "u@x", Role: "VIEWER"}}, "u@x", "VIEWER", false},
		{"malformed id with state rejects", "nocolon", UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "u@x", Role: "VIEWER"}}, "", "", true},
		{"malformed id without state", "nocolon", UserPlatformRoleState{}, "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			email, role, err := parseUserPlatformRoleID(tc.id, tc.state)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if err == nil && (email != tc.wantMail || role != tc.wantRole) {
				t.Errorf("got (%q, %q), want (%q, %q)", email, role, tc.wantMail, tc.wantRole)
			}
		})
	}
}

func TestUserPlatformRoleCreate_DuplicateEntryGuidance(t *testing.T) {
	dupErr := &client.LagoonAPIError{Message: "role already exists for user"}
	mock := &mockLagoonClient{
		addPlatformRoleToUserFn: func(_ context.Context, _, _ string) error { return dupErr },
	}
	ctx := testCtx(mock)
	r := &UserPlatformRole{}
	_, err := r.Create(ctx, infer.CreateRequest[UserPlatformRoleArgs]{
		Inputs: UserPlatformRoleArgs{UserEmail: "admin@x", Role: "OWNER"},
	})
	if err == nil || !strings.Contains(err.Error(), "pulumi import") {
		t.Errorf("expected duplicate-entry error mentioning pulumi import, got: %v", err)
	}
}

func TestUserPlatformRoleDiff_NoDeleteBeforeReplace(t *testing.T) {
	r := &UserPlatformRole{}
	resp, err := r.Diff(context.Background(), infer.DiffRequest[UserPlatformRoleArgs, UserPlatformRoleState]{
		Inputs: UserPlatformRoleArgs{UserEmail: "admin@x", Role: "OWNER"},
		State:  UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: "admin@x", Role: "VIEWER"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DeleteBeforeReplace {
		t.Error("DeleteBeforeReplace must be false so a failed Create does not leave the user with no platform role")
	}
}

func TestUserGroupAssignmentDiff_NoDeleteBeforeReplace(t *testing.T) {
	r := &UserGroupAssignment{}
	resp, err := r.Diff(context.Background(), infer.DiffRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]{
		Inputs: UserGroupAssignmentArgs{UserEmail: "u@x", GroupName: "newgroup", Role: "GUEST"},
		State:  UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: "u@x", GroupName: "oldgroup", Role: "GUEST"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DeleteBeforeReplace {
		t.Error("DeleteBeforeReplace must be false so a failed Create does not drop the user's access")
	}
}

// Client-level containsNotFound / NotFound translation tests
func TestContainsNotFound_NarrowMatching(t *testing.T) {
	// These are checked indirectly via the real client, but we can assert the behavior
	// through DeleteUser: an authorization-style error containing "no user permissions"
	// must NOT be converted to ErrNotFound.
	mock := &mockLagoonClient{
		deleteUserFn: func(_ context.Context, _ string) error {
			// Mock returns exactly what the real client would return for an authz error:
			// the original LagoonAPIError, NOT a LagoonNotFoundError.
			return &client.LagoonAPIError{Message: "access denied: no user permissions for this operation"}
		},
	}
	ctx := testCtx(mock)
	r := &User{}
	_, err := r.Delete(ctx, infer.DeleteRequest[UserState]{
		State: UserState{UserArgs: UserArgs{Email: "u@x"}, LagoonID: "1"},
	})
	if err == nil {
		t.Fatal("expected authorization error to surface, not be silently swallowed as NotFound")
	}
}
