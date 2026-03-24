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

func (r *Environment) Create(ctx context.Context, req infer.CreateRequest[EnvironmentArgs]) (infer.CreateResponse[EnvironmentState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Default deployBaseRef to the environment name if not provided.
	// The Lagoon API requires this field (String!).
	deployBaseRef := req.Inputs.Name
	if req.Inputs.DeployBaseRef != nil {
		deployBaseRef = *req.Inputs.DeployBaseRef
	}

	input := map[string]any{
		"name":            req.Inputs.Name,
		"project":         req.Inputs.ProjectID,
		"deployType":      strings.ToUpper(req.Inputs.DeployType),
		"environmentType": strings.ToUpper(req.Inputs.EnvironmentType),
		"deployBaseRef":   deployBaseRef,
	}
	setOptional(input, "deployHeadRef", req.Inputs.DeployHeadRef)
	setOptional(input, "deployTitle", req.Inputs.DeployTitle)
	setOptional(input, "openshiftProjectName", req.Inputs.OpenshiftProjectName)
	setOptionalInt(input, "autoIdle", req.Inputs.AutoIdle)

	// Store the effective deployBaseRef in state so it round-trips correctly.
	effectiveInputs := req.Inputs
	if effectiveInputs.DeployBaseRef == nil {
		effectiveInputs.DeployBaseRef = &deployBaseRef
	}

	if req.DryRun {
		return infer.CreateResponse[EnvironmentState]{
			ID:     "preview-id",
			Output: EnvironmentState{EnvironmentArgs: effectiveInputs},
		}, nil
	}

	env, err := client.AddOrUpdateEnvironment(ctx, input)
	if err != nil {
		return infer.CreateResponse[EnvironmentState]{}, fmt.Errorf("failed to create environment: %w", err)
	}

	return infer.CreateResponse[EnvironmentState]{
		ID: strconv.Itoa(env.ID),
		Output: EnvironmentState{
			EnvironmentArgs: effectiveInputs,
			LagoonID:        env.ID,
			Route:           env.Route,
			Routes:          env.Routes,
			Created:         env.Created,
		},
	}, nil
}

func (r *Environment) Update(ctx context.Context, req infer.UpdateRequest[EnvironmentArgs, EnvironmentState]) (infer.UpdateResponse[EnvironmentState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Default deployBaseRef to the environment name if not provided.
	deployBaseRef := req.Inputs.Name
	if req.Inputs.DeployBaseRef != nil {
		deployBaseRef = *req.Inputs.DeployBaseRef
	}

	input := map[string]any{
		"name":            req.Inputs.Name,
		"project":         req.Inputs.ProjectID,
		"deployType":      strings.ToUpper(req.Inputs.DeployType),
		"environmentType": strings.ToUpper(req.Inputs.EnvironmentType),
		"deployBaseRef":   deployBaseRef,
	}
	setOptional(input, "deployHeadRef", req.Inputs.DeployHeadRef)
	setOptional(input, "deployTitle", req.Inputs.DeployTitle)
	setOptional(input, "openshiftProjectName", req.Inputs.OpenshiftProjectName)
	setOptionalInt(input, "autoIdle", req.Inputs.AutoIdle)

	// Store the effective deployBaseRef in state so it round-trips correctly.
	effectiveInputs := req.Inputs
	if effectiveInputs.DeployBaseRef == nil {
		effectiveInputs.DeployBaseRef = &deployBaseRef
	}

	if req.DryRun {
		return infer.UpdateResponse[EnvironmentState]{
			Output: EnvironmentState{
				EnvironmentArgs: effectiveInputs,
				LagoonID:        req.State.LagoonID,
				Route:           req.State.Route,
				Routes:          req.State.Routes,
				Created:         req.State.Created,
			},
		}, nil
	}

	env, err := client.AddOrUpdateEnvironment(ctx, input)
	if err != nil {
		return infer.UpdateResponse[EnvironmentState]{}, fmt.Errorf("failed to update environment: %w", err)
	}

	return infer.UpdateResponse[EnvironmentState]{
		Output: EnvironmentState{
			EnvironmentArgs: effectiveInputs,
			LagoonID:        env.ID,
			Route:           env.Route,
			Routes:          env.Routes,
			Created:         env.Created,
		},
	}, nil
}

func (r *Environment) Delete(ctx context.Context, req infer.DeleteRequest[EnvironmentState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Lagoon's deleteEnvironment mutation requires the project name (String!), not the ID.
	proj, err := c.GetProjectByID(ctx, req.State.ProjectID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			// Parent project already gone, environment is implicitly deleted
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to look up project name for environment deletion: %w", err)
	}

	if err := c.DeleteEnvironment(ctx, req.State.Name, proj.Name); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete environment: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *Environment) Read(ctx context.Context, req infer.ReadRequest[EnvironmentArgs, EnvironmentState]) (infer.ReadResponse[EnvironmentArgs, EnvironmentState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Parse composite ID: {project_id}:{env_name}
	parts := strings.SplitN(req.ID, ":", 2)
	var projectID int
	var envName string

	if len(parts) == 2 {
		// Import scenario
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return infer.ReadResponse[EnvironmentArgs, EnvironmentState]{}, fmt.Errorf("invalid environment import ID '%s': project_id must be numeric", req.ID)
		}
		projectID = pid
		envName = parts[1]
	} else {
		// Refresh scenario — use state
		projectID = req.State.ProjectID
		envName = req.State.Name
	}

	env, err := c.GetEnvironmentByName(ctx, envName, projectID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[EnvironmentArgs, EnvironmentState]{ID: ""}, nil
		}
		return infer.ReadResponse[EnvironmentArgs, EnvironmentState]{}, fmt.Errorf("failed to read environment: %w", err)
	}

	// Preserve resolved project ID if API omits nested project
	resolvedProjectID := env.ProjectID
	if resolvedProjectID == 0 {
		resolvedProjectID = projectID
	}

	args := EnvironmentArgs{
		Name:            env.Name,
		ProjectID:       resolvedProjectID,
		DeployType:      strings.ToLower(env.DeployType),
		EnvironmentType: strings.ToLower(env.EnvironmentType),
	}
	if env.DeployBaseRef != "" {
		args.DeployBaseRef = &env.DeployBaseRef
	}
	if env.DeployHeadRef != "" {
		args.DeployHeadRef = &env.DeployHeadRef
	}
	if env.DeployTitle != "" {
		args.DeployTitle = &env.DeployTitle
	}
	if env.AutoIdle != nil {
		args.AutoIdle = env.AutoIdle
	}
	if env.OpenshiftProjectName != "" {
		args.OpenshiftProjectName = &env.OpenshiftProjectName
	}

	st := EnvironmentState{
		EnvironmentArgs: args,
		LagoonID:        env.ID,
		Route:           env.Route,
		Routes:          env.Routes,
		Created:         env.Created,
	}

	return infer.ReadResponse[EnvironmentArgs, EnvironmentState]{
		ID:     strconv.Itoa(env.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *Environment) Diff(ctx context.Context, req infer.DiffRequest[EnvironmentArgs, EnvironmentState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// forceNew fields
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.ProjectID != req.State.ProjectID {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Updatable fields
	if !strings.EqualFold(req.Inputs.DeployType, req.State.DeployType) {
		diff["deployType"] = p.PropertyDiff{Kind: p.Update}
	}
	if !strings.EqualFold(req.Inputs.EnvironmentType, req.State.EnvironmentType) {
		diff["environmentType"] = p.PropertyDiff{Kind: p.Update}
	}
	// Normalize deployBaseRef: treat nil as the environment name (same default as Create/Update)
	// to avoid spurious diffs when the user omits this optional field.
	inputDBR := req.Inputs.DeployBaseRef
	if inputDBR == nil {
		inputDBR = &req.Inputs.Name
	}
	stateDBR := req.State.DeployBaseRef
	if stateDBR == nil {
		stateDBR = &req.State.Name
	}
	if ptrDiffers(inputDBR, stateDBR) {
		diff["deployBaseRef"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.DeployHeadRef, req.State.DeployHeadRef) {
		diff["deployHeadRef"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.DeployTitle, req.State.DeployTitle) {
		diff["deployTitle"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(req.Inputs.OpenshiftProjectName, req.State.OpenshiftProjectName) {
		diff["openshiftProjectName"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrIntDiffers(req.Inputs.AutoIdle, req.State.AutoIdle) {
		diff["autoIdle"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}