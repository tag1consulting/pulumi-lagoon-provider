package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

var validGroupRoles = map[string]bool{
	"GUEST":      true,
	"REPORTER":   true,
	"DEVELOPER":  true,
	"MAINTAINER": true,
	"OWNER":      true,
}

// UserGroupAssignment assigns a user to a group with a specific role.
type UserGroupAssignment struct{}

// UserGroupAssignmentArgs defines the input properties for a user-to-group role binding.
type UserGroupAssignmentArgs struct {
	UserEmail string `pulumi:"userEmail"`
	GroupName string `pulumi:"groupName"`
	Role      string `pulumi:"role"`
}

// UserGroupAssignmentState is the persisted state of a UserGroupAssignment.
type UserGroupAssignmentState struct {
	UserGroupAssignmentArgs
}

// validGroupRoleList returns the allowed group role values for error messages.
func validGroupRoleList() string {
	return "GUEST, REPORTER, DEVELOPER, MAINTAINER, OWNER"
}

func (r *UserGroupAssignment) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "UserGroupAssignment")
	a.Describe(&r, "Assigns a Lagoon user to a group with a specific role.")
}

func (a *UserGroupAssignmentArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.UserEmail, "The user's email address.")
	an.Describe(&a.GroupName, "The group name.")
	an.Describe(&a.Role, "The role within the group: GUEST, REPORTER, DEVELOPER, MAINTAINER, or OWNER.")
}

func (r *UserGroupAssignment) Create(ctx context.Context, req infer.CreateRequest[UserGroupAssignmentArgs]) (infer.CreateResponse[UserGroupAssignmentState], error) {
	c := clientFor(ctx)

	if req.Inputs.UserEmail == "" {
		return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf("userEmail must not be empty")
	}
	if req.Inputs.GroupName == "" {
		return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf("groupName must not be empty")
	}
	role := strings.ToUpper(req.Inputs.Role)
	if !validGroupRoles[role] {
		return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf("invalid group role %q: must be one of %s", req.Inputs.Role, validGroupRoleList())
	}

	id := fmt.Sprintf("%s:%s", req.Inputs.UserEmail, req.Inputs.GroupName)

	if req.DryRun {
		return infer.CreateResponse[UserGroupAssignmentState]{
			ID:     id,
			Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
		}, nil
	}

	if err := c.AddUserToGroup(ctx, req.Inputs.UserEmail, req.Inputs.GroupName, role); err != nil {
		if client.IsDuplicateEntry(err) {
			return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf(
				"user %q is already assigned to group %q; use `pulumi import lagoon:lagoon:UserGroupAssignment <name> %s` to adopt it: %w",
				req.Inputs.UserEmail, req.Inputs.GroupName, id, err)
		}
		return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf("failed to add user to group: %w", err)
	}

	return infer.CreateResponse[UserGroupAssignmentState]{
		ID:     id,
		Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
	}, nil
}

func (r *UserGroupAssignment) Read(ctx context.Context, req infer.ReadRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]) (infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState], error) {
	c := clientFor(ctx)

	email, groupName, err := parseUserGroupAssignmentID(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState]{}, err
	}

	roles, err := c.GetUserGroupRoles(ctx, email)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState]{}, nil
		}
		return infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState]{}, fmt.Errorf("failed to read user group roles: %w", err)
	}

	for _, gr := range roles {
		if gr.Name == groupName {
			role := strings.ToUpper(gr.Role)
			args := UserGroupAssignmentArgs{UserEmail: email, GroupName: groupName, Role: role}
			st := UserGroupAssignmentState{UserGroupAssignmentArgs: args}
			return infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState]{
				ID:     req.ID,
				Inputs: args,
				State:  st,
			}, nil
		}
	}

	return infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState]{}, nil
}

// parseUserGroupAssignmentID extracts email and groupName from a Pulumi resource ID.
// The ID format is "{email}:{groupName}". Because email local-parts may contain colons
// (RFC 5321 quoted strings), we split on the LAST colon rather than the first. If ID
// parsing fails, we fall back to prior state (useful during refresh/update). Returns an
// error when neither source yields non-empty email and groupName.
func parseUserGroupAssignmentID(id string, state UserGroupAssignmentState) (email, groupName string, err error) {
	if idx := strings.LastIndex(id, ":"); idx > 0 && idx < len(id)-1 {
		email = id[:idx]
		groupName = id[idx+1:]
	} else if state.UserEmail != "" && state.GroupName != "" {
		email = state.UserEmail
		groupName = state.GroupName
	} else {
		return "", "", fmt.Errorf("invalid user group assignment ID %q: expected non-empty {email}:{groupName}", id)
	}
	if email == "" || groupName == "" {
		return "", "", fmt.Errorf("invalid user group assignment ID %q: email and groupName must both be non-empty", id)
	}
	return email, groupName, nil
}

func (r *UserGroupAssignment) Update(ctx context.Context, req infer.UpdateRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]) (infer.UpdateResponse[UserGroupAssignmentState], error) {
	c := clientFor(ctx)

	role := strings.ToUpper(req.Inputs.Role)
	if !validGroupRoles[role] {
		return infer.UpdateResponse[UserGroupAssignmentState]{}, fmt.Errorf("invalid group role %q: must be one of %s", req.Inputs.Role, validGroupRoleList())
	}

	if req.DryRun {
		return infer.UpdateResponse[UserGroupAssignmentState]{
			Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
		}, nil
	}

	// addUserToGroup is upsert: re-adding the same user+group with a different role
	// updates the role in place (no remove+add needed).
	if err := c.AddUserToGroup(ctx, req.Inputs.UserEmail, req.Inputs.GroupName, role); err != nil {
		return infer.UpdateResponse[UserGroupAssignmentState]{}, fmt.Errorf("failed to update user group role: %w", err)
	}

	return infer.UpdateResponse[UserGroupAssignmentState]{
		Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
	}, nil
}

func (r *UserGroupAssignment) Delete(ctx context.Context, req infer.DeleteRequest[UserGroupAssignmentState]) (infer.DeleteResponse, error) {
	c := clientFor(ctx)

	if err := c.RemoveUserFromGroup(ctx, req.State.UserEmail, req.State.GroupName); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to remove user from group: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *UserGroupAssignment) Diff(ctx context.Context, req infer.DiffRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.UserEmail != req.State.UserEmail {
		diff["userEmail"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.GroupName != req.State.GroupName {
		diff["groupName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if !strings.EqualFold(req.Inputs.Role, req.State.Role) {
		diff["role"] = p.PropertyDiff{Kind: p.Update}
	}
	// Note: DeleteBeforeReplace is intentionally false. Group assignments are
	// independent (a user may hold a role in any number of groups), so creating the
	// new binding before removing the old one avoids a window of dropped access if
	// the create step fails mid-replace.
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}
