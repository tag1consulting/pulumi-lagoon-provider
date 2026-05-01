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

// validPlatformRoleList returns the allowed platform role values for error messages.
func validPlatformRoleList() string {
	return "OWNER, VIEWER"
}

// UserPlatformRole assigns a platform-level role to a Lagoon user.
type UserPlatformRole struct{}

// UserPlatformRoleArgs defines the input properties for a platform role binding.
type UserPlatformRoleArgs struct {
	UserEmail string `pulumi:"userEmail"`
	Role      string `pulumi:"role"`
}

// UserPlatformRoleState is the persisted state of a UserPlatformRole.
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

	if req.Inputs.UserEmail == "" {
		return infer.CreateResponse[UserPlatformRoleState]{}, fmt.Errorf("userEmail must not be empty")
	}
	role := strings.ToUpper(req.Inputs.Role)
	if !validPlatformRoles[role] {
		return infer.CreateResponse[UserPlatformRoleState]{}, fmt.Errorf("invalid platform role %q: must be %s", req.Inputs.Role, validPlatformRoleList())
	}

	id := fmt.Sprintf("%s:%s", req.Inputs.UserEmail, role)

	if req.DryRun {
		return infer.CreateResponse[UserPlatformRoleState]{
			ID:     id,
			Output: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: req.Inputs.UserEmail, Role: role}},
		}, nil
	}

	if err := c.AddPlatformRoleToUser(ctx, req.Inputs.UserEmail, role); err != nil {
		if client.IsDuplicateEntry(err) {
			return infer.CreateResponse[UserPlatformRoleState]{}, fmt.Errorf(
				"user %q already has platform role %q; use `pulumi import lagoon:lagoon:UserPlatformRole <name> %s` to adopt it: %w",
				req.Inputs.UserEmail, role, id, err)
		}
		return infer.CreateResponse[UserPlatformRoleState]{}, fmt.Errorf("failed to add platform role to user: %w", err)
	}

	return infer.CreateResponse[UserPlatformRoleState]{
		ID:     id,
		Output: UserPlatformRoleState{UserPlatformRoleArgs: UserPlatformRoleArgs{UserEmail: req.Inputs.UserEmail, Role: role}},
	}, nil
}

func (r *UserPlatformRole) Read(ctx context.Context, req infer.ReadRequest[UserPlatformRoleArgs, UserPlatformRoleState]) (infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState], error) {
	c := clientFor(ctx)

	email, role, err := parseUserPlatformRoleID(req.ID, req.State)
	if err != nil {
		return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, err
	}

	roles, err := c.GetUserPlatformRoles(ctx, email)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, nil
		}
		return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, fmt.Errorf("failed to read user platform roles: %w", err)
	}

	for _, platformRole := range roles {
		if strings.EqualFold(platformRole, role) {
			args := UserPlatformRoleArgs{UserEmail: email, Role: role}
			st := UserPlatformRoleState{UserPlatformRoleArgs: args}
			return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{
				ID:     req.ID,
				Inputs: args,
				State:  st,
			}, nil
		}
	}

	p.GetLogger(ctx).Warningf("UserPlatformRole %s: user %q does not have platform role %q — removing from state", req.ID, email, role)
	return infer.ReadResponse[UserPlatformRoleArgs, UserPlatformRoleState]{}, nil
}

// parseUserPlatformRoleID extracts email and role from a Pulumi resource ID.
// The ID format is "{email}:{role}". Because email local-parts may contain colons
// (RFC 5321 quoted strings), we split on the LAST colon. The parsed role is validated
// against the known platform-role enum so that a typo in an imported ID surfaces as a
// clear error instead of silently reporting the resource as deleted.
func parseUserPlatformRoleID(id string, state UserPlatformRoleState) (email, role string, err error) {
	if idx := strings.LastIndex(id, ":"); idx > 0 && idx < len(id)-1 {
		email = id[:idx]
		role = strings.ToUpper(id[idx+1:])
	} else if id != "" {
		return "", "", fmt.Errorf("invalid user platform role ID %q: expected {email}:{role}", id)
	} else if state.UserEmail != "" && state.Role != "" {
		email = state.UserEmail
		role = strings.ToUpper(state.Role)
	} else {
		return "", "", fmt.Errorf("invalid user platform role ID %q: expected non-empty {email}:{role}", id)
	}
	if email == "" || role == "" {
		return "", "", fmt.Errorf("invalid user platform role ID %q: email and role must both be non-empty", id)
	}
	if !validPlatformRoles[role] {
		return "", "", fmt.Errorf("invalid user platform role ID %q: role must be one of %s", id, validPlatformRoleList())
	}
	return email, role, nil
}

// No Update method — the Lagoon API exposes only addPlatformRoleToUser and
// removePlatformRoleFromUser mutations, so any change to userEmail or role requires a
// replace (see Diff).

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
	// Note: DeleteBeforeReplace is intentionally false. Lagoon allows multiple
	// platform roles concurrently, so creating the new assignment before removing the
	// old one avoids a window where the user has no platform role if the create
	// step fails mid-replace.
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff}, nil
}
