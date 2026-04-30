package resources

import (
	"context"
	"errors"
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// User manages a Lagoon user.
type User struct{}

// UserArgs defines the input properties for a Lagoon User resource.
type UserArgs struct {
	Email     string  `pulumi:"email"`
	FirstName *string `pulumi:"firstName,optional"`
	LastName  *string `pulumi:"lastName,optional"`
	Comment   *string `pulumi:"comment,optional"`
}

// UserState is the persisted state of a Lagoon User resource.
type UserState struct {
	UserArgs
	LagoonID string `pulumi:"lagoonId"`
}

func (r *User) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "User")
	a.Describe(&r, "Manages a Lagoon user.")
}

func (a *UserArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Email, "The user's email address (Lagoon's primary user identifier).")
	an.Describe(&a.FirstName, "The user's first name.")
	an.Describe(&a.LastName, "The user's last name.")
	an.Describe(&a.Comment, "An optional comment about the user.")
}

func (s *UserState) Annotate(an infer.Annotator) {
	an.Describe(&s.LagoonID, "The Lagoon internal ID of the user.")
}

func (r *User) Create(ctx context.Context, req infer.CreateRequest[UserArgs]) (infer.CreateResponse[UserState], error) {
	c := clientFor(ctx)

	if req.DryRun {
		return infer.CreateResponse[UserState]{
			ID:     "preview-id",
			Output: UserState{UserArgs: req.Inputs},
		}, nil
	}

	u, err := c.CreateUser(ctx, req.Inputs.Email, req.Inputs.FirstName, req.Inputs.LastName, req.Inputs.Comment)
	if err != nil {
		if client.IsDuplicateEntry(err) {
			return infer.CreateResponse[UserState]{}, fmt.Errorf(
				"user %q already exists in Lagoon; use `pulumi import lagoon:lagoon:User <name> %s` to adopt it: %w",
				req.Inputs.Email, req.Inputs.Email, err)
		}
		return infer.CreateResponse[UserState]{}, fmt.Errorf("failed to create user: %w", err)
	}

	return infer.CreateResponse[UserState]{
		ID:     u.Email,
		Output: UserState{UserArgs: req.Inputs, LagoonID: u.ID},
	}, nil
}

func (r *User) Read(ctx context.Context, req infer.ReadRequest[UserArgs, UserState]) (infer.ReadResponse[UserArgs, UserState], error) {
	c := clientFor(ctx)

	email := req.State.Email
	if email == "" {
		email = req.ID
	}

	u, err := c.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[UserArgs, UserState]{}, nil
		}
		return infer.ReadResponse[UserArgs, UserState]{}, fmt.Errorf("failed to read user: %w", err)
	}

	args := UserArgs{
		Email:     u.Email,
		FirstName: nilIfEmpty(u.FirstName),
		LastName:  nilIfEmpty(u.LastName),
		Comment:   nilIfEmpty(u.Comment),
	}
	st := UserState{UserArgs: args, LagoonID: u.ID}

	return infer.ReadResponse[UserArgs, UserState]{
		ID:     u.Email,
		Inputs: args,
		State:  st,
	}, nil
}

func (r *User) Update(ctx context.Context, req infer.UpdateRequest[UserArgs, UserState]) (infer.UpdateResponse[UserState], error) {
	c := clientFor(ctx)

	// Build the patch. When an input transitions from non-nil to nil (user clears the
	// field), we send explicit GraphQL null (Go nil in the map marshals to JSON null) so
	// Lagoon actually clears the field server-side. If we used setOptional here, the key
	// would be omitted and the stale value would persist in Lagoon, causing permanent
	// drift on subsequent pulumi up runs.
	patch := map[string]any{}
	patchField := func(key string, input, state *string) {
		if !ptrDiffers(input, state) {
			return
		}
		if input == nil {
			patch[key] = nil
		} else {
			patch[key] = *input
		}
	}
	patchField("firstName", req.Inputs.FirstName, req.State.FirstName)
	patchField("lastName", req.Inputs.LastName, req.State.LastName)
	patchField("comment", req.Inputs.Comment, req.State.Comment)

	if req.DryRun || len(patch) == 0 {
		return infer.UpdateResponse[UserState]{
			Output: UserState{UserArgs: req.Inputs, LagoonID: req.State.LagoonID},
		}, nil
	}

	_, err := c.UpdateUser(ctx, req.State.Email, patch)
	if err != nil {
		return infer.UpdateResponse[UserState]{}, fmt.Errorf("failed to update user: %w", err)
	}

	return infer.UpdateResponse[UserState]{
		Output: UserState{UserArgs: req.Inputs, LagoonID: req.State.LagoonID},
	}, nil
}

func (r *User) Delete(ctx context.Context, req infer.DeleteRequest[UserState]) (infer.DeleteResponse, error) {
	c := clientFor(ctx)

	if err := c.DeleteUser(ctx, req.State.Email); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete user: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *User) Diff(ctx context.Context, req infer.DiffRequest[UserArgs, UserState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Email != req.State.Email {
		diff["email"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.FirstName, req.State.FirstName) {
		diff["firstName"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.LastName, req.State.LastName) {
		diff["lastName"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.Comment, req.State.Comment) {
		diff["comment"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
