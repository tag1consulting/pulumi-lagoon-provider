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

var validPlatformRoles = map[string]bool{
	"OWNER":  true,
	"VIEWER": true,
}

// UserPlatformRole assigns a platform-level role to a Lagoon user.
type UserPlatformRole struct{}

type UserPlatformRoleArgs struct {
	UserEmail string `pulumi:"userEmail"`
	Role      string `pulumi:"role"`
}

type UserPlatformRoleState struct {
	UserPlatformRoleArgs
}

func (r *UserPlatformRole) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "UserPlatformRole")
	a.Describe(&r, "Assigns a platform-level role (OWNER or VIEWER) to a Lagoon user.")
}

func (a *UserPlatformRoleArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.UserEmail, "The user's email address.")
	an.Describe(&a.Role, "The platform role: OWNER or VIEWER.")
}

func (r *UserPlatformRole) Create(ctx context.Context, req infer.CreateRequest[UserPlatformRoleArgs]) (infer.CreateResponse[UserPlatformRoleState], error) {
	c := clientFor(ctx)

	role := strings.ToUpper(req.Inputs.Role)
	if !validPlatformRoles[role] {
		return infer.CreateResponse[UserPlatformRoleState]{}, fmt.Errorf("invalid platform role %q: must be OWNER or VIEWER", req.Inputs.Role)
	}

	id := fmt.Sprintf("%s:%s", req.Inputs.UserEmail, role)

	if req.DryRun {
		return infer.CreateResponse[UserPlatformRoleState]{
			ID:     id,
			Output: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: req.Inputs.UserEmail, Role: role}},
		}, nil
	}

	if err := c.AddPlatformRoleToUser(ctx, req.Inputs.UserEmail, role); err != nil {
		return infer.CreateResponse[UserPlatformRoleState]{}, fmt.Errorf("failed to add platform role to user: %w", err)
	}

	return infer.CreateResponse[UserPlatformRoleState]{
		ID:     id,
		Output: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: req.Inputs.UserEmail, Role: role}},
	}, nil
}

func (r *UserPlatformRole) Read(ctx context.Context, req infer.ReadRequest[UserPlatformRoleArgs, UserPlatformRoleState]) (infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState], error) {
	c := clientFor(ctx)

	var email, role string
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) == 2 {
		email = parts[0]
		role = strings.ToUpper(parts[1])
	} else if req.State.UserEmail != "" {
		email = req.State.UserEmail
		role = strings.ToUpper(req.State.Role)
	} else {
		return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{},
			fmt.Errorf("invalid user platform role ID '%s': expected format {email}:{role}", req.ID)
	}

	roles, err := c.GetUserPlatformRoles(ctx, email)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, nil
		}
		return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, fmt.Errorf("failed to read user platform roles: %w", err)
	}

	for _, r := range roles {
		if strings.EqualFold(r, role) {
			args := UserPlatformRoleArgs{UserEmail: email, Role: role}
			st := UserPlatformRoleState{UserPlatformRoleArgs: args}
			return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{
				ID:     req.ID,
				Inputs: args,
				State:  st,
			}, nil
		}
	}

	return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, nil
}

// No Update — all fields are forceNew.

func (r *UserPlatformRole) Delete(ctx context.Context, req infer.DeleteRequest[UserPlatformRoleState]) (infer.DeleteResponse, error) {
	c := clientFor(ctx)

	if err := c.RemovePlatformRoleFromUser(ctx, req.State.UserEmail, req.State.Role); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to remove platform role from user: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *UserPlatformRole) Diff(ctx context.Context, req infer.DiffRequest[UserPlatformRoleArgs, UserPlatformRoleState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.UserEmail != req.State.UserEmail {
		diff["userEmail"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if !strings.EqualFold(req.Inputs.Role, req.State.Role) {
		diff["role"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
