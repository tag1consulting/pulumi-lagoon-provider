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

type UserGroupAssignmentArgs struct {
	UserEmail string `pulumi:"userEmail"`
	GroupName string `pulumi:"groupName"`
	Role      string `pulumi:"role"`
}

type UserGroupAssignmentState struct {
	UserGroupAssignmentArgs
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

	role := strings.ToUpper(req.Inputs.Role)
	if !validGroupRoles[role] {
		return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf("invalid group role %q: must be one of GUEST, REPORTER, DEVELOPER, MAINTAINER, OWNER", req.Inputs.Role)
	}

	id := fmt.Sprintf("%s:%s", req.Inputs.UserEmail, req.Inputs.GroupName)

	if req.DryRun {
		return infer.CreateResponse[UserGroupAssignmentState]{
			ID:     id,
			Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
		}, nil
	}

	if err := c.AddUserToGroup(ctx, req.Inputs.UserEmail, req.Inputs.GroupName, role); err != nil {
		return infer.CreateResponse[UserGroupAssignmentState]{}, fmt.Errorf("failed to add user to group: %w", err)
	}

	return infer.CreateResponse[UserGroupAssignmentState]{
		ID:     id,
		Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
	}, nil
}

func (r *UserGroupAssignment) Read(ctx context.Context, req infer.ReadRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]) (infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState], error) {
	c := clientFor(ctx)

	var email, groupName string
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) == 2 {
		email = parts[0]
		groupName = parts[1]
	} else if req.State.UserEmail != "" {
		email = req.State.UserEmail
		groupName = req.State.GroupName
	} else {
		return infer.ReadResponse[UserGroupAssignmentArgs, UserGroupAssignmentState]{},
			fmt.Errorf("invalid user group assignment ID '%s': expected format {email}:{groupName}", req.ID)
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
			args := UserGroupAssignmentArgs{UserEmail: email, GroupName: groupName, Role: strings.ToUpper(gr.Role)}
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

func (r *UserGroupAssignment) Update(ctx context.Context, req infer.UpdateRequest[UserGroupAssignmentArgs, UserGroupAssignmentState]) (infer.UpdateResponse[UserGroupAssignmentState], error) {
	c := clientFor(ctx)

	role := strings.ToUpper(req.Inputs.Role)
	if !validGroupRoles[role] {
		return infer.UpdateResponse[UserGroupAssignmentState]{}, fmt.Errorf("invalid group role %q: must be one of GUEST, REPORTER, DEVELOPER, MAINTAINER, OWNER", req.Inputs.Role)
	}

	if req.DryRun {
		return infer.UpdateResponse[UserGroupAssignmentState]{
			Output: UserGroupAssignmentState{UserGroupAssignmentArgs: UserGroupAssignmentArgs{UserEmail: req.Inputs.UserEmail, GroupName: req.Inputs.GroupName, Role: role}},
		}, nil
	}

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
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
