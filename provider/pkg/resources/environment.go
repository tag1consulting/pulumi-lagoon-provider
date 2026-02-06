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

// Environment manages a Lagoon environment resource.
type Environment struct{}

// EnvironmentArgs are the inputs for a Lagoon environment.
type EnvironmentArgs struct {
	Name                 string  `pulumi:"name"`
	ProjectID            int     `pulumi:"projectId"`
	DeployType           string  `pulumi:"deployType"`
	EnvironmentType      string  `pulumi:"environmentType"`
	DeployBaseRef        *string `pulumi:"deployBaseRef,optional"`
	DeployHeadRef        *string `pulumi:"deployHeadRef,optional"`
	DeployTitle          *string `pulumi:"deployTitle,optional"`
	OpenshiftProjectName *string `pulumi:"openshiftProjectName,optional"`
	AutoIdle             *int    `pulumi:"autoIdle,optional"`
}

// EnvironmentState is the full state of a Lagoon environment.
type EnvironmentState struct {
	EnvironmentArgs
	LagoonID int    `pulumi:"lagoonId"`
	Route    string `pulumi:"route"`
	Routes   string `pulumi:"routes"`
	Created  string `pulumi:"created"`
}

func (r *Environment) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "Environment")
	a.Describe(&r, "Manages a Lagoon environment (branch/PR deployment).")
}

func (a *EnvironmentArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The environment name (typically the branch name).")
	an.Describe(&a.ProjectID, "The parent project ID.")
	an.Describe(&a.DeployType, "Deployment type: 'branch' or 'pullrequest'.")
	an.Describe(&a.EnvironmentType, "Environment type: 'production', 'development', or 'standby'.")
	an.Describe(&a.DeployBaseRef, "The base ref for the deployment.")
	an.Describe(&a.DeployHeadRef, "The head ref for pull request deployments.")
	an.Describe(&a.DeployTitle, "Title for pull request deployments.")
	an.Describe(&a.OpenshiftProjectName, "Override namespace name on the cluster.")
	an.Describe(&a.AutoIdle, "Whether to auto-idle this environment (1=yes, 0=no).")
}

func (s *EnvironmentState) Annotate(a infer.Annotator) {
	a.Describe(&s.LagoonID, "The Lagoon-assigned numeric ID of the environment.")
	a.Describe(&s.Route, "The primary route URL for this environment.")
	a.Describe(&s.Routes, "All route URLs for this environment.")
	a.Describe(&s.Created, "The creation timestamp.")
}

func (r *Environment) Create(ctx context.Context, name string, inputs EnvironmentArgs, preview bool) (string, EnvironmentState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"name":            inputs.Name,
		"project":         inputs.ProjectID,
		"deployType":      strings.ToUpper(inputs.DeployType),
		"environmentType": strings.ToUpper(inputs.EnvironmentType),
	}
	setOptional(input, "deployBaseRef", inputs.DeployBaseRef)
	setOptional(input, "deployHeadRef", inputs.DeployHeadRef)
	setOptional(input, "deployTitle", inputs.DeployTitle)
	setOptional(input, "openshiftProjectName", inputs.OpenshiftProjectName)
	setOptionalInt(input, "autoIdle", inputs.AutoIdle)

	if preview {
		return "preview-id", EnvironmentState{EnvironmentArgs: inputs}, nil
	}

	env, err := client.AddOrUpdateEnvironment(ctx, input)
	if err != nil {
		return "", EnvironmentState{}, fmt.Errorf("failed to create environment: %w", err)
	}

	return strconv.Itoa(env.ID), EnvironmentState{
		EnvironmentArgs: inputs,
		LagoonID:        env.ID,
		Route:           env.Route,
		Routes:          env.Routes,
		Created:         env.Created,
	}, nil
}

func (r *Environment) Update(ctx context.Context, id string, olds EnvironmentState, news EnvironmentArgs, preview bool) (EnvironmentState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"name":            news.Name,
		"project":         news.ProjectID,
		"deployType":      strings.ToUpper(news.DeployType),
		"environmentType": strings.ToUpper(news.EnvironmentType),
	}
	setOptional(input, "deployBaseRef", news.DeployBaseRef)
	setOptional(input, "deployHeadRef", news.DeployHeadRef)
	setOptional(input, "deployTitle", news.DeployTitle)
	setOptional(input, "openshiftProjectName", news.OpenshiftProjectName)
	setOptionalInt(input, "autoIdle", news.AutoIdle)

	if preview {
		return EnvironmentState{
			EnvironmentArgs: news,
			LagoonID:        olds.LagoonID,
			Route:           olds.Route,
			Routes:          olds.Routes,
			Created:         olds.Created,
		}, nil
	}

	env, err := client.AddOrUpdateEnvironment(ctx, input)
	if err != nil {
		return EnvironmentState{}, fmt.Errorf("failed to update environment: %w", err)
	}

	return EnvironmentState{
		EnvironmentArgs: news,
		LagoonID:        env.ID,
		Route:           env.Route,
		Routes:          env.Routes,
		Created:         env.Created,
	}, nil
}

func (r *Environment) Delete(ctx context.Context, id string, props EnvironmentState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if err := client.DeleteEnvironment(ctx, props.Name, props.ProjectID); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	return nil
}

func (r *Environment) Read(ctx context.Context, id string, inputs EnvironmentArgs, state EnvironmentState) (string, EnvironmentArgs, EnvironmentState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Parse composite ID: {project_id}:{env_name}
	parts := strings.SplitN(id, ":", 2)
	var projectID int
	var envName string

	if len(parts) == 2 {
		// Import scenario
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", EnvironmentArgs{}, EnvironmentState{}, fmt.Errorf("invalid environment import ID '%s': project_id must be numeric", id)
		}
		projectID = pid
		envName = parts[1]
	} else {
		// Refresh scenario — use state
		projectID = state.ProjectID
		envName = state.Name
	}

	env, err := client.GetEnvironmentByName(ctx, envName, projectID)
	if err != nil {
		return "", EnvironmentArgs{}, EnvironmentState{}, fmt.Errorf("failed to read environment: %w", err)
	}

	args := EnvironmentArgs{
		Name:            env.Name,
		ProjectID:       env.ProjectID,
		DeployType:      strings.ToLower(env.DeployType),
		EnvironmentType: strings.ToLower(env.EnvironmentType),
	}

	st := EnvironmentState{
		EnvironmentArgs: args,
		LagoonID:        env.ID,
		Route:           env.Route,
		Routes:          env.Routes,
		Created:         env.Created,
	}

	return strconv.Itoa(env.ID), args, st, nil
}

func (r *Environment) Diff(ctx context.Context, id string, olds EnvironmentState, news EnvironmentArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// forceNew fields
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.ProjectID != olds.ProjectID {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Updatable fields
	if !strings.EqualFold(news.DeployType, olds.DeployType) {
		diff["deployType"] = p.PropertyDiff{Kind: p.Update}
	}
	if !strings.EqualFold(news.EnvironmentType, olds.EnvironmentType) {
		diff["environmentType"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.DeployBaseRef, olds.DeployBaseRef) {
		diff["deployBaseRef"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.DeployHeadRef, olds.DeployHeadRef) {
		diff["deployHeadRef"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.DeployTitle, olds.DeployTitle) {
		diff["deployTitle"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.OpenshiftProjectName, olds.OpenshiftProjectName) {
		diff["openshiftProjectName"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrIntDiffers(news.AutoIdle, olds.AutoIdle) {
		diff["autoIdle"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}
