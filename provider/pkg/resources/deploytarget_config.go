package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
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

func (r *DeployTargetConfig) Create(ctx context.Context, name string, inputs DeployTargetConfigArgs, preview bool) (string, DeployTargetConfigState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"project":      inputs.ProjectID,
		"deployTarget": inputs.DeployTargetID,
	}
	setOptional(input, "branches", inputs.Branches)
	setOptional(input, "pullrequests", inputs.Pullrequests)
	setOptionalInt(input, "weight", inputs.Weight)
	setOptional(input, "deployTargetProjectPattern", inputs.DeployTargetProjectPattern)

	if preview {
		return "preview-id", DeployTargetConfigState{DeployTargetConfigArgs: inputs}, nil
	}

	dtc, err := client.CreateDeployTargetConfig(ctx, input)
	if err != nil {
		return "", DeployTargetConfigState{}, fmt.Errorf("failed to create deploy target config: %w", err)
	}

	return strconv.Itoa(dtc.ID), DeployTargetConfigState{
		DeployTargetConfigArgs: inputs,
		LagoonID:               dtc.ID,
	}, nil
}

func (r *DeployTargetConfig) Update(ctx context.Context, id string, olds DeployTargetConfigState, news DeployTargetConfigArgs, preview bool) (DeployTargetConfigState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{}
	setOptional(input, "branches", news.Branches)
	setOptional(input, "pullrequests", news.Pullrequests)
	setOptionalInt(input, "weight", news.Weight)
	setOptional(input, "deployTargetProjectPattern", news.DeployTargetProjectPattern)

	if preview {
		return DeployTargetConfigState{
			DeployTargetConfigArgs: news,
			LagoonID:               olds.LagoonID,
		}, nil
	}

	_, err := client.UpdateDeployTargetConfig(ctx, olds.LagoonID, input)
	if err != nil {
		return DeployTargetConfigState{}, fmt.Errorf("failed to update deploy target config: %w", err)
	}

	return DeployTargetConfigState{
		DeployTargetConfigArgs: news,
		LagoonID:               olds.LagoonID,
	}, nil
}

func (r *DeployTargetConfig) Delete(ctx context.Context, id string, props DeployTargetConfigState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if err := client.DeleteDeployTargetConfig(ctx, props.LagoonID, props.ProjectID); err != nil {
		return fmt.Errorf("failed to delete deploy target config: %w", err)
	}
	return nil
}

func (r *DeployTargetConfig) Read(ctx context.Context, id string, inputs DeployTargetConfigArgs, state DeployTargetConfigState) (string, DeployTargetConfigArgs, DeployTargetConfigState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Import ID format: {project_id}:{config_id}
	parts := strings.SplitN(id, ":", 2)
	var projectID, configID int

	if len(parts) == 2 {
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", DeployTargetConfigArgs{}, DeployTargetConfigState{}, fmt.Errorf("invalid import ID: %w", err)
		}
		cid, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", DeployTargetConfigArgs{}, DeployTargetConfigState{}, fmt.Errorf("invalid import ID: %w", err)
		}
		projectID = pid
		configID = cid
	} else {
		projectID = state.ProjectID
		configID = state.LagoonID
	}

	dtc, err := client.GetDeployTargetConfigByID(ctx, configID, projectID)
	if err != nil {
		return "", DeployTargetConfigArgs{}, DeployTargetConfigState{}, fmt.Errorf("failed to read deploy target config: %w", err)
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
	if dtc.Weight != 0 {
		args.Weight = &dtc.Weight
	}
	if dtc.DeployTargetProjectPattern != "" {
		args.DeployTargetProjectPattern = &dtc.DeployTargetProjectPattern
	}

	st := DeployTargetConfigState{
		DeployTargetConfigArgs: args,
		LagoonID:               dtc.ID,
	}

	return strconv.Itoa(dtc.ID), args, st, nil
}

func (r *DeployTargetConfig) Diff(ctx context.Context, id string, olds DeployTargetConfigState, news DeployTargetConfigArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	if news.ProjectID != olds.ProjectID {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.DeployTargetID != olds.DeployTargetID {
		diff["deployTargetId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrDiffers(news.Branches, olds.Branches) {
		diff["branches"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.Pullrequests, olds.Pullrequests) {
		diff["pullrequests"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrIntDiffers(news.Weight, olds.Weight) {
		diff["weight"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.DeployTargetProjectPattern, olds.DeployTargetProjectPattern) {
		diff["deployTargetProjectPattern"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true,
	}, nil
}
