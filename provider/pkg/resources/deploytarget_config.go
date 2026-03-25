package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// DeployTargetConfig manages a Lagoon deploy target configuration.
type DeployTargetConfig struct{}

type DeployTargetConfigArgs struct {
	ProjectID                  int     `pulumi:"projectId"`
	DeployTargetID             int     `pulumi:"deployTargetId"`
	Branches                   *string `pulumi:"branches,optional"`
	Pullrequests               *string `pulumi:"pullrequests,optional"`
	Weight                     *int    `pulumi:"weight,optional"`
	DeployTargetProjectPattern *string `pulumi:"deployTargetProjectPattern,optional"`
}

type DeployTargetConfigState struct {
	DeployTargetConfigArgs
	LagoonID int `pulumi:"lagoonId"`
}

func (r *DeployTargetConfig) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "DeployTargetConfig")
	a.Describe(&r, "Manages a deploy target configuration to route branches/PRs to specific clusters.")
}

func (a *DeployTargetConfigArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.ProjectID, "The project ID.")
	an.Describe(&a.DeployTargetID, "The deploy target (Kubernetes cluster) ID.")
	an.Describe(&a.Branches, "Regex pattern for branches to match.")
	an.Describe(&a.Pullrequests, "Whether to handle PRs ('true' or 'false').")
	an.Describe(&a.Weight, "Priority weight (higher = more priority).")
	an.Describe(&a.DeployTargetProjectPattern, "Optional namespace pattern.")
}

func (r *DeployTargetConfig) Create(ctx context.Context, req infer.CreateRequest[DeployTargetConfigArgs]) (infer.CreateResponse[DeployTargetConfigState], error) {
	c := clientFor(ctx)

	input := map[string]any{
		"project":      req.Inputs.ProjectID,
		"deployTarget": req.Inputs.DeployTargetID,
	}
	setOptional(input, "branches", req.Inputs.Branches)
	setOptional(input, "pullrequests", req.Inputs.Pullrequests)
	setOptionalInt(input, "weight", req.Inputs.Weight)
	setOptional(input, "deployTargetProjectPattern", req.Inputs.DeployTargetProjectPattern)

	if req.DryRun {
		return infer.CreateResponse[DeployTargetConfigState]{
			ID:     "preview-id",
			Output: DeployTargetConfigState{DeployTargetConfigArgs: req.Inputs},
		}, nil
	}

	dtc, err := c.CreateDeployTargetConfig(ctx, input)
	if err != nil {
		if !client.IsDuplicateEntry(err) {
			return infer.CreateResponse[DeployTargetConfigState]{}, fmt.Errorf("failed to create deploy target config: %w", err)
		}

		// Resource already exists — adopt it by finding the matching config and updating
		configs, lookupErr := c.GetDeployTargetConfigsByProject(ctx, req.Inputs.ProjectID)
		if lookupErr != nil {
			return infer.CreateResponse[DeployTargetConfigState]{}, fmt.Errorf("deploy target config already exists but failed to list configs: %w", lookupErr)
		}

		var existingID int
		for _, cfg := range configs {
			if cfg.DeployTargetID == req.Inputs.DeployTargetID {
				existingID = cfg.ID
				break
			}
		}
		if existingID == 0 {
			return infer.CreateResponse[DeployTargetConfigState]{}, fmt.Errorf("deploy target config for project %d / target %d already exists but could not be found", req.Inputs.ProjectID, req.Inputs.DeployTargetID)
		}

		// Update the existing resource to match desired inputs
		dtc, err = c.UpdateDeployTargetConfig(ctx, existingID, input)
		if err != nil {
			return infer.CreateResponse[DeployTargetConfigState]{}, fmt.Errorf("deploy target config already exists but failed to update: %w", err)
		}
	}

	return infer.CreateResponse[DeployTargetConfigState]{
		ID: strconv.Itoa(dtc.ID),
		Output: DeployTargetConfigState{
			DeployTargetConfigArgs: req.Inputs,
			LagoonID:               dtc.ID,
		},
	}, nil
}

func (r *DeployTargetConfig) Update(ctx context.Context, req infer.UpdateRequest[DeployTargetConfigArgs, DeployTargetConfigState]) (infer.UpdateResponse[DeployTargetConfigState], error) {
	c := clientFor(ctx)

	input := map[string]any{}
	setOptional(input, "branches", req.Inputs.Branches)
	setOptional(input, "pullrequests", req.Inputs.Pullrequests)
	setOptionalInt(input, "weight", req.Inputs.Weight)
	setOptional(input, "deployTargetProjectPattern", req.Inputs.DeployTargetProjectPattern)

	if req.DryRun {
		return infer.UpdateResponse[DeployTargetConfigState]{
			Output: DeployTargetConfigState{
				DeployTargetConfigArgs: req.Inputs,
				LagoonID:               req.State.LagoonID,
			},
		}, nil
	}

	_, err := c.UpdateDeployTargetConfig(ctx, req.State.LagoonID, input)
	if err != nil {
		return infer.UpdateResponse[DeployTargetConfigState]{}, fmt.Errorf("failed to update deploy target config: %w", err)
	}

	return infer.UpdateResponse[DeployTargetConfigState]{
		Output: DeployTargetConfigState{
			DeployTargetConfigArgs: req.Inputs,
			LagoonID:               req.State.LagoonID,
		},
	}, nil
}

func (r *DeployTargetConfig) Delete(ctx context.Context, req infer.DeleteRequest[DeployTargetConfigState]) (infer.DeleteResponse, error) {
	c := clientFor(ctx)

	if err := c.DeleteDeployTargetConfig(ctx, req.State.LagoonID, req.State.ProjectID); err != nil {
		// Treat "not found" as success — resource is already gone
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete deploy target config: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *DeployTargetConfig) Read(ctx context.Context, req infer.ReadRequest[DeployTargetConfigArgs, DeployTargetConfigState]) (infer.ReadResponse[DeployTargetConfigArgs, DeployTargetConfigState], error) {
	c := clientFor(ctx)

	// Import ID format: {project_id}:{config_id}
	parts := strings.SplitN(req.ID, ":", 2)
	var projectID, configID int

	if len(parts) == 2 {
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return infer.ReadResponse[DeployTargetConfigArgs, DeployTargetConfigState]{}, fmt.Errorf("invalid import ID: %w", err)
		}
		cid, err := strconv.Atoi(parts[1])
		if err != nil {
			return infer.ReadResponse[DeployTargetConfigArgs, DeployTargetConfigState]{}, fmt.Errorf("invalid import ID: %w", err)
		}
		projectID = pid
		configID = cid
	} else {
		projectID = req.State.ProjectID
		configID = req.State.LagoonID
	}

	dtc, err := c.GetDeployTargetConfigByID(ctx, configID, projectID)
	if err != nil {
		// Return empty ID to signal the resource was deleted from Lagoon
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[DeployTargetConfigArgs, DeployTargetConfigState]{}, nil
		}
		return infer.ReadResponse[DeployTargetConfigArgs, DeployTargetConfigState]{}, fmt.Errorf("failed to read deploy target config: %w", err)
	}

	args := DeployTargetConfigArgs{
		ProjectID:      dtc.ProjectID,
		DeployTargetID: dtc.DeployTargetID,
	}
	if dtc.Branches != "" {
		args.Branches = &dtc.Branches
	}
	if dtc.Pullrequests != "" {
		args.Pullrequests = &dtc.Pullrequests
	}
	// Always populate Weight so explicit 0 doesn't drift
	args.Weight = &dtc.Weight
	if dtc.DeployTargetProjectPattern != "" {
		args.DeployTargetProjectPattern = &dtc.DeployTargetProjectPattern
	}

	st := DeployTargetConfigState{
		DeployTargetConfigArgs: args,
		LagoonID:               dtc.ID,
	}

	return infer.ReadResponse[DeployTargetConfigArgs, DeployTargetConfigState]{
		ID:     strconv.Itoa(dtc.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *DeployTargetConfig) Diff(ctx context.Context, req infer.DiffRequest[DeployTargetConfigArgs, DeployTargetConfigState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	if req.Inputs.ProjectID != req.State.ProjectID {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.DeployTargetID != req.State.DeployTargetID {
		diff["deployTargetId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	// Normalize optional fields: nil and empty/zero are equivalent (API defaults)
	inBranches := ptrOrDefault(req.Inputs.Branches, "")
	stBranches := ptrOrDefault(req.State.Branches, "")
	if inBranches != stBranches {
		diff["branches"] = p.PropertyDiff{Kind: p.Update}
	}
	inPR := ptrOrDefault(req.Inputs.Pullrequests, "")
	stPR := ptrOrDefault(req.State.Pullrequests, "")
	if inPR != stPR {
		diff["pullrequests"] = p.PropertyDiff{Kind: p.Update}
	}
	inWeight := ptrIntOrDefault(req.Inputs.Weight, 0)
	stWeight := ptrIntOrDefault(req.State.Weight, 0)
	if inWeight != stWeight {
		diff["weight"] = p.PropertyDiff{Kind: p.Update}
	}
	inPattern := ptrOrDefault(req.Inputs.DeployTargetProjectPattern, "")
	stPattern := ptrOrDefault(req.State.DeployTargetProjectPattern, "")
	if inPattern != stPattern {
		diff["deployTargetProjectPattern"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true,
	}, nil
}
