package resources

import (
	"context"
	"errors"
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

// Group manages a Lagoon group.
type Group struct{}

type GroupArgs struct {
	Name            string  `pulumi:"name"`
	ParentGroupName *string `pulumi:"parentGroupName,optional"`
}

type GroupState struct {
	GroupArgs
	LagoonID string `pulumi:"lagoonId"`
}

func (r *Group) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "Group")
	a.Describe(&r, "Manages a Lagoon group for organizing projects and users.")
}

func (a *GroupArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The group name.")
	an.Describe(&a.ParentGroupName, "The name of the parent group, for creating subgroups.")
}

func (s *GroupState) Annotate(an infer.Annotator) {
	an.Describe(&s.LagoonID, "The Lagoon internal ID of the group.")
}

func (r *Group) Create(ctx context.Context, req infer.CreateRequest[GroupArgs]) (infer.CreateResponse[GroupState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	if req.DryRun {
		return infer.CreateResponse[GroupState]{
			ID:     "preview-id",
			Output: GroupState{GroupArgs: req.Inputs},
		}, nil
	}

	g, err := c.CreateGroup(ctx, req.Inputs.Name, req.Inputs.ParentGroupName)
	if err != nil {
		return infer.CreateResponse[GroupState]{}, fmt.Errorf("failed to create group: %w", err)
	}

	return infer.CreateResponse[GroupState]{
		ID:     g.ID,
		Output: GroupState{GroupArgs: req.Inputs, LagoonID: g.ID},
	}, nil
}

func (r *Group) Read(ctx context.Context, req infer.ReadRequest[GroupArgs, GroupState]) (infer.ReadResponse[GroupArgs, GroupState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	name := req.State.Name
	if name == "" {
		name = req.ID
	}

	g, err := c.GetGroupByName(ctx, name)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[GroupArgs, GroupState]{}, nil
		}
		return infer.ReadResponse[GroupArgs, GroupState]{}, fmt.Errorf("failed to read group: %w", err)
	}

	// parentGroupName is not available from the Lagoon API (parentGroup is not
	// exposed on GroupInterface), so preserve it from the prior state.
	args := GroupArgs{Name: g.Name, ParentGroupName: req.State.ParentGroupName}
	st := GroupState{GroupArgs: args, LagoonID: g.ID}

	return infer.ReadResponse[GroupArgs, GroupState]{
		ID:     g.ID,
		Inputs: args,
		State:  st,
	}, nil
}

func (r *Group) Update(ctx context.Context, req infer.UpdateRequest[GroupArgs, GroupState]) (infer.UpdateResponse[GroupState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	patch := map[string]any{}
	if ptrDiffers(req.Inputs.ParentGroupName, req.State.ParentGroupName) {
		if req.Inputs.ParentGroupName != nil {
			patch["parentGroup"] = map[string]any{"name": *req.Inputs.ParentGroupName}
		}
	}

	if req.DryRun || len(patch) == 0 {
		return infer.UpdateResponse[GroupState]{
			Output: GroupState{GroupArgs: req.Inputs, LagoonID: req.State.LagoonID},
		}, nil
	}

	_, err := c.UpdateGroup(ctx, req.State.Name, patch)
	if err != nil {
		return infer.UpdateResponse[GroupState]{}, fmt.Errorf("failed to update group: %w", err)
	}

	return infer.UpdateResponse[GroupState]{
		Output: GroupState{GroupArgs: req.Inputs, LagoonID: req.State.LagoonID},
	}, nil
}

func (r *Group) Delete(ctx context.Context, req infer.DeleteRequest[GroupState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	if err := c.DeleteGroup(ctx, req.State.Name); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete group: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *Group) Diff(ctx context.Context, req infer.DiffRequest[GroupArgs, GroupState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(req.Inputs.ParentGroupName, req.State.ParentGroupName) {
		diff["parentGroupName"] = p.PropertyDiff{Kind: p.Update}
	}
	return p.DiffResponse{HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true}, nil
}
