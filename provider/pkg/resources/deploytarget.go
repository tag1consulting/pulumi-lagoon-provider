package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

// DeployTarget manages a Lagoon Kubernetes deploy target.
type DeployTarget struct{}

type DeployTargetArgs struct {
	Name                string  `pulumi:"name"`
	ConsoleURL          string  `pulumi:"consoleUrl"`
	CloudProvider       *string `pulumi:"cloudProvider,optional"`
	CloudRegion         *string `pulumi:"cloudRegion,optional"`
	SSHHost             *string `pulumi:"sshHost,optional"`
	SSHPort             *string `pulumi:"sshPort,optional"`
	BuildImage          *string `pulumi:"buildImage,optional"`
	Disabled      *bool   `pulumi:"disabled,optional"`
	RouterPattern *string `pulumi:"routerPattern,optional"`
}

type DeployTargetState struct {
	DeployTargetArgs
	LagoonID int    `pulumi:"lagoonId"`
	Created  string `pulumi:"created"`
}

func (r *DeployTarget) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "DeployTarget")
	a.Describe(&r, "Manages a Lagoon Kubernetes deploy target (cluster).")
}

func (a *DeployTargetArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The deploy target name.")
	an.Describe(&a.ConsoleURL, "The Kubernetes API URL.")
	an.Describe(&a.CloudProvider, "Cloud provider (e.g., 'kind', 'aws', 'gcp'). Defaults to 'kind'.")
	an.Describe(&a.CloudRegion, "Cloud region (e.g., 'us-east-1', 'local'). Defaults to 'local'.")
	an.Describe(&a.SSHHost, "SSH host for builds.")
	an.Describe(&a.SSHPort, "SSH port for builds.")
	an.Describe(&a.BuildImage, "Custom build image.")
	an.Describe(&a.Disabled, "Whether the deploy target is disabled.")
	an.Describe(&a.RouterPattern, "Router pattern for the deploy target.")
}

func (r *DeployTarget) Create(ctx context.Context, req infer.CreateRequest[DeployTargetArgs]) (infer.CreateResponse[DeployTargetState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Compute effective values for fields that have API-side defaults.
	// Storing these explicitly in state prevents spurious drift on pulumi refresh.
	effectiveCloudProvider := "kind"
	if req.Inputs.CloudProvider != nil {
		effectiveCloudProvider = *req.Inputs.CloudProvider
	}
	effectiveCloudRegion := "local"
	if req.Inputs.CloudRegion != nil {
		effectiveCloudRegion = *req.Inputs.CloudRegion
	}

	input := map[string]any{
		"name":          req.Inputs.Name,
		"consoleUrl":    req.Inputs.ConsoleURL,
		"cloudProvider": effectiveCloudProvider,
		"cloudRegion":   effectiveCloudRegion,
	}
	setOptional(input, "sshHost", req.Inputs.SSHHost)
	setOptional(input, "sshPort", req.Inputs.SSHPort)
	setOptional(input, "buildImage", req.Inputs.BuildImage)
	setOptionalBool(input, "disabled", req.Inputs.Disabled)
	setOptional(input, "routerPattern", req.Inputs.RouterPattern)

	// Build normalized args with explicit effective values to avoid drift
	normalizedArgs := req.Inputs
	normalizedArgs.CloudProvider = &effectiveCloudProvider
	normalizedArgs.CloudRegion = &effectiveCloudRegion

	if req.DryRun {
		return infer.CreateResponse[DeployTargetState]{
			ID:     "preview-id",
			Output: DeployTargetState{DeployTargetArgs: normalizedArgs},
		}, nil
	}

	dt, err := c.CreateDeployTarget(ctx, input)
	if err != nil {
		if !client.IsDuplicateEntry(err) {
			return infer.CreateResponse[DeployTargetState]{}, fmt.Errorf("failed to create deploy target: %w", err)
		}

		// Resource already exists — adopt it by looking up by name and updating
		existing, lookupErr := c.GetDeployTargetByName(ctx, req.Inputs.Name)
		if lookupErr != nil {
			return infer.CreateResponse[DeployTargetState]{}, fmt.Errorf("deploy target %q already exists but failed to look up: %w", req.Inputs.Name, lookupErr)
		}

		// Update the existing resource to match desired inputs
		// Remove name — it's not updatable via the UpdateKubernetes mutation
		delete(input, "name")
		dt, err = c.UpdateDeployTarget(ctx, existing.ID, input)
		if err != nil {
			return infer.CreateResponse[DeployTargetState]{}, fmt.Errorf("deploy target %q already exists but failed to update: %w", req.Inputs.Name, err)
		}
	}

	return infer.CreateResponse[DeployTargetState]{
		ID: strconv.Itoa(dt.ID),
		Output: DeployTargetState{
			DeployTargetArgs: normalizedArgs,
			LagoonID:         dt.ID,
			Created:          dt.Created,
		},
	}, nil
}

func (r *DeployTarget) Update(ctx context.Context, req infer.UpdateRequest[DeployTargetArgs, DeployTargetState]) (infer.UpdateResponse[DeployTargetState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	input := map[string]any{
		"consoleUrl": req.Inputs.ConsoleURL,
	}
	setOptional(input, "cloudProvider", req.Inputs.CloudProvider)
	setOptional(input, "cloudRegion", req.Inputs.CloudRegion)
	setOptional(input, "sshHost", req.Inputs.SSHHost)
	setOptional(input, "sshPort", req.Inputs.SSHPort)
	setOptional(input, "buildImage", req.Inputs.BuildImage)
	setOptionalBool(input, "disabled", req.Inputs.Disabled)
	setOptional(input, "routerPattern", req.Inputs.RouterPattern)

	if req.DryRun {
		return infer.UpdateResponse[DeployTargetState]{
			Output: DeployTargetState{
				DeployTargetArgs: req.Inputs,
				LagoonID:         req.State.LagoonID,
				Created:          req.State.Created,
			},
		}, nil
	}

	_, err := c.UpdateDeployTarget(ctx, req.State.LagoonID, input)
	if err != nil {
		return infer.UpdateResponse[DeployTargetState]{}, fmt.Errorf("failed to update deploy target: %w", err)
	}

	return infer.UpdateResponse[DeployTargetState]{
		Output: DeployTargetState{
			DeployTargetArgs: req.Inputs,
			LagoonID:         req.State.LagoonID,
			Created:          req.State.Created,
		},
	}, nil
}

func (r *DeployTarget) Delete(ctx context.Context, req infer.DeleteRequest[DeployTargetState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	if err := c.DeleteDeployTarget(ctx, req.State.Name); err != nil {
		// Treat "not found" as success — resource is already gone
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete deploy target: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *DeployTarget) Read(ctx context.Context, req infer.ReadRequest[DeployTargetArgs, DeployTargetState]) (infer.ReadResponse[DeployTargetArgs, DeployTargetState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	dtID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.ReadResponse[DeployTargetArgs, DeployTargetState]{}, fmt.Errorf("invalid deploy target ID: %w", err)
	}

	dt, err := c.GetDeployTargetByID(ctx, dtID)
	if err != nil {
		// Return empty ID to signal the resource was deleted from Lagoon
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[DeployTargetArgs, DeployTargetState]{}, nil
		}
		return infer.ReadResponse[DeployTargetArgs, DeployTargetState]{}, fmt.Errorf("failed to read deploy target: %w", err)
	}

	args := DeployTargetArgs{
		Name:       dt.Name,
		ConsoleURL: dt.ConsoleURL,
	}
	if dt.CloudProvider != "" {
		args.CloudProvider = &dt.CloudProvider
	}
	if dt.CloudRegion != "" {
		args.CloudRegion = &dt.CloudRegion
	}
	if dt.SSHHost != "" {
		args.SSHHost = &dt.SSHHost
	}
	if dt.SSHPort != "" {
		args.SSHPort = &dt.SSHPort
	}
	if dt.BuildImage != "" {
		args.BuildImage = &dt.BuildImage
	}
	if dt.RouterPattern != "" {
		args.RouterPattern = &dt.RouterPattern
	}
	if dt.Disabled {
		args.Disabled = &dt.Disabled
	}

	st := DeployTargetState{
		DeployTargetArgs: args,
		LagoonID:         dt.ID,
		Created:          dt.Created,
	}

	return infer.ReadResponse[DeployTargetArgs, DeployTargetState]{
		ID:     req.ID,
		Inputs: args,
		State:  st,
	}, nil
}

func (r *DeployTarget) Diff(ctx context.Context, req infer.DiffRequest[DeployTargetArgs, DeployTargetState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.ConsoleURL != req.State.ConsoleURL {
		diff["consoleUrl"] = p.PropertyDiff{Kind: p.Update}
	}

	// Normalize defaulted fields before comparing: treat nil as the API default so
	// omitting an optional field in the program never triggers a spurious update.
	// Both sides are normalized so legacy state (nil) also matches the default.
	inputCP, stateCP := req.Inputs.CloudProvider, req.State.CloudProvider
	kind := "kind"
	if inputCP == nil {
		inputCP = &kind
	}
	if stateCP == nil {
		stateCP = &kind
	}
	if ptrDiffers(inputCP, stateCP) {
		diff["cloudProvider"] = p.PropertyDiff{Kind: p.Update}
	}

	inputCR, stateCR := req.Inputs.CloudRegion, req.State.CloudRegion
	local := "local"
	if inputCR == nil {
		inputCR = &local
	}
	if stateCR == nil {
		stateCR = &local
	}
	if ptrDiffers(inputCR, stateCR) {
		diff["cloudRegion"] = p.PropertyDiff{Kind: p.Update}
	}

	if ptrDiffers(req.Inputs.SSHHost, req.State.SSHHost) {
		diff["sshHost"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.SSHPort, req.State.SSHPort) {
		diff["sshPort"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.BuildImage, req.State.BuildImage) {
		diff["buildImage"] = p.PropertyDiff{Kind: p.Update}
	}

	// Normalize nil disabled to &false (the API default) so omitting the field
	// doesn't report a spurious update when disabled is already false in state.
	inputDis, stateDis := req.Inputs.Disabled, req.State.Disabled
	falseVal := false
	if inputDis == nil {
		inputDis = &falseVal
	}
	if stateDis == nil {
		stateDis = &falseVal
	}
	if ptrBoolDiffers(inputDis, stateDis) {
		diff["disabled"] = p.PropertyDiff{Kind: p.Update}
	}

	if ptrDiffers(req.Inputs.RouterPattern, req.State.RouterPattern) {
		diff["routerPattern"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true,
	}, nil
}
