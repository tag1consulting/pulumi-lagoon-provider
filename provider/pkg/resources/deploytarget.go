package resources

import (
	"context"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
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
	Disabled            *bool   `pulumi:"disabled,optional"`
	RouterPattern       *string `pulumi:"routerPattern,optional"`
	SharedBastionSecret *string `pulumi:"sharedBastionSecret,optional"`
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

func (r *DeployTarget) Create(ctx context.Context, name string, inputs DeployTargetArgs, preview bool) (string, DeployTargetState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"name":       inputs.Name,
		"consoleUrl": inputs.ConsoleURL,
	}
	if inputs.CloudProvider != nil {
		input["cloudProvider"] = *inputs.CloudProvider
	} else {
		input["cloudProvider"] = "kind"
	}
	if inputs.CloudRegion != nil {
		input["cloudRegion"] = *inputs.CloudRegion
	} else {
		input["cloudRegion"] = "local"
	}
	setOptional(input, "sshHost", inputs.SSHHost)
	setOptional(input, "sshPort", inputs.SSHPort)
	setOptional(input, "buildImage", inputs.BuildImage)
	setOptionalBool(input, "disabled", inputs.Disabled)
	setOptional(input, "routerPattern", inputs.RouterPattern)

	if preview {
		return "preview-id", DeployTargetState{DeployTargetArgs: inputs}, nil
	}

	dt, err := client.CreateDeployTarget(ctx, input)
	if err != nil {
		return "", DeployTargetState{}, fmt.Errorf("failed to create deploy target: %w", err)
	}

	return strconv.Itoa(dt.ID), DeployTargetState{
		DeployTargetArgs: inputs,
		LagoonID:         dt.ID,
		Created:          dt.Created,
	}, nil
}

func (r *DeployTarget) Update(ctx context.Context, id string, olds DeployTargetState, news DeployTargetArgs, preview bool) (DeployTargetState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"consoleUrl": news.ConsoleURL,
	}
	setOptional(input, "cloudProvider", news.CloudProvider)
	setOptional(input, "cloudRegion", news.CloudRegion)
	setOptional(input, "sshHost", news.SSHHost)
	setOptional(input, "sshPort", news.SSHPort)
	setOptional(input, "buildImage", news.BuildImage)
	setOptionalBool(input, "disabled", news.Disabled)
	setOptional(input, "routerPattern", news.RouterPattern)

	if preview {
		return DeployTargetState{
			DeployTargetArgs: news,
			LagoonID:         olds.LagoonID,
			Created:          olds.Created,
		}, nil
	}

	_, err := client.UpdateDeployTarget(ctx, olds.LagoonID, input)
	if err != nil {
		return DeployTargetState{}, fmt.Errorf("failed to update deploy target: %w", err)
	}

	return DeployTargetState{
		DeployTargetArgs: news,
		LagoonID:         olds.LagoonID,
		Created:          olds.Created,
	}, nil
}

func (r *DeployTarget) Delete(ctx context.Context, id string, props DeployTargetState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if err := client.DeleteDeployTarget(ctx, props.Name); err != nil {
		return fmt.Errorf("failed to delete deploy target: %w", err)
	}
	return nil
}

func (r *DeployTarget) Read(ctx context.Context, id string, inputs DeployTargetArgs, state DeployTargetState) (string, DeployTargetArgs, DeployTargetState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	dtID, err := strconv.Atoi(id)
	if err != nil {
		return "", DeployTargetArgs{}, DeployTargetState{}, fmt.Errorf("invalid deploy target ID: %w", err)
	}

	dt, err := client.GetDeployTargetByID(ctx, dtID)
	if err != nil {
		return "", DeployTargetArgs{}, DeployTargetState{}, fmt.Errorf("failed to read deploy target: %w", err)
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

	st := DeployTargetState{
		DeployTargetArgs: args,
		LagoonID:         dt.ID,
		Created:          dt.Created,
	}

	return id, args, st, nil
}

func (r *DeployTarget) Diff(ctx context.Context, id string, olds DeployTargetState, news DeployTargetArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.ConsoleURL != olds.ConsoleURL {
		diff["consoleUrl"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.CloudProvider, olds.CloudProvider) {
		diff["cloudProvider"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.CloudRegion, olds.CloudRegion) {
		diff["cloudRegion"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.SSHHost, olds.SSHHost) {
		diff["sshHost"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.SSHPort, olds.SSHPort) {
		diff["sshPort"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.BuildImage, olds.BuildImage) {
		diff["buildImage"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrBoolDiffers(news.Disabled, olds.Disabled) {
		diff["disabled"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.RouterPattern, olds.RouterPattern) {
		diff["routerPattern"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges: len(diff) > 0, DetailedDiff: diff, DeleteBeforeReplace: true,
	}, nil
}
