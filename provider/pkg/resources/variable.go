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

// Variable manages a Lagoon environment variable.
type Variable struct{}

// VariableArgs are the inputs for a Lagoon variable.
type VariableArgs struct {
	Name          string `pulumi:"name"`
	Value         string `pulumi:"value" provider:"secret"`
	ProjectID     int    `pulumi:"projectId"`
	Scope         string `pulumi:"scope"`
	EnvironmentID *int   `pulumi:"environmentId,optional"`
}

// VariableState is the full state of a Lagoon variable.
type VariableState struct {
	VariableArgs
	LagoonID int `pulumi:"lagoonId"`
}

func (r *Variable) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "Variable")
	a.Describe(&r, "Manages a Lagoon environment or project-level variable.")
}

func (a *VariableArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The variable name.")
	an.Describe(&a.Value, "The variable value (stored as secret).")
	an.Describe(&a.ProjectID, "The parent project ID.")
	an.Describe(&a.Scope, "Variable scope: 'build', 'runtime', 'global', 'container_registry', or 'internal_container_registry'.")
	an.Describe(&a.EnvironmentID, "Environment ID (omit for project-level variables).")
}

func (s *VariableState) Annotate(a infer.Annotator) {
	a.Describe(&s.LagoonID, "The Lagoon-assigned numeric ID of the variable.")
}

func (r *Variable) Create(ctx context.Context, req infer.CreateRequest[VariableArgs]) (infer.CreateResponse[VariableState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if req.DryRun {
		return infer.CreateResponse[VariableState]{
			ID:     "preview-id",
			Output: VariableState{VariableArgs: req.Inputs},
		}, nil
	}

	v, err := client.AddVariable(ctx, req.Inputs.Name, req.Inputs.Value, req.Inputs.ProjectID, req.Inputs.Scope, req.Inputs.EnvironmentID)
	if err != nil {
		return infer.CreateResponse[VariableState]{}, fmt.Errorf("failed to create variable: %w", err)
	}

	return infer.CreateResponse[VariableState]{
		ID: strconv.Itoa(v.ID),
		Output: VariableState{
			VariableArgs: req.Inputs,
			LagoonID:     v.ID,
		},
	}, nil
}

func (r *Variable) Update(ctx context.Context, req infer.UpdateRequest[VariableArgs, VariableState]) (infer.UpdateResponse[VariableState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if req.DryRun {
		return infer.UpdateResponse[VariableState]{
			Output: VariableState{
				VariableArgs: req.Inputs,
				LagoonID:     req.State.LagoonID,
			},
		}, nil
	}

	// Lagoon uses addOrUpdate semantics — same mutation for create and update
	v, err := client.AddVariable(ctx, req.Inputs.Name, req.Inputs.Value, req.Inputs.ProjectID, req.Inputs.Scope, req.Inputs.EnvironmentID)
	if err != nil {
		return infer.UpdateResponse[VariableState]{}, fmt.Errorf("failed to update variable: %w", err)
	}

	return infer.UpdateResponse[VariableState]{
		Output: VariableState{
			VariableArgs: req.Inputs,
			LagoonID:     v.ID,
		},
	}, nil
}

func (r *Variable) Delete(ctx context.Context, req infer.DeleteRequest[VariableState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	if err := c.DeleteVariable(ctx, req.State.Name, req.State.ProjectID, req.State.EnvironmentID); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete variable: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

func (r *Variable) Read(ctx context.Context, req infer.ReadRequest[VariableArgs, VariableState]) (infer.ReadResponse[VariableArgs, VariableState], error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	c := cfg.NewClient()

	// Parse import ID: {project_id}:{env_id}:{var_name} or {project_id}::{var_name}
	parts := strings.SplitN(req.ID, ":", 3)
	var projectID int
	var environmentID *int
	var varName string

	if len(parts) == 3 {
		// Import scenario
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return infer.ReadResponse[VariableArgs, VariableState]{}, fmt.Errorf("invalid variable import ID '%s': project_id must be numeric", req.ID)
		}
		projectID = pid
		varName = parts[2]

		if parts[1] != "" {
			eid, err := strconv.Atoi(parts[1])
			if err != nil {
				return infer.ReadResponse[VariableArgs, VariableState]{}, fmt.Errorf("invalid variable import ID '%s': env_id must be numeric", req.ID)
			}
			environmentID = &eid
		}
	} else if req.State.Name != "" {
		// Refresh from existing state
		projectID = req.State.ProjectID
		environmentID = req.State.EnvironmentID
		varName = req.State.Name
	} else {
		return infer.ReadResponse[VariableArgs, VariableState]{},
			fmt.Errorf("invalid variable import ID '%s': expected format {project_id}:{env_id}:{var_name} or {project_id}::{var_name}", req.ID)
	}

	v, err := c.GetVariable(ctx, varName, projectID, environmentID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[VariableArgs, VariableState]{}, nil
		}
		return infer.ReadResponse[VariableArgs, VariableState]{}, fmt.Errorf("failed to read variable: %w", err)
	}

	args := VariableArgs{
		Name:          v.Name,
		Value:         v.Value,
		ProjectID:     projectID,
		Scope:         strings.ToLower(v.Scope),
		EnvironmentID: environmentID,
	}

	st := VariableState{
		VariableArgs: args,
		LagoonID:     v.ID,
	}

	return infer.ReadResponse[VariableArgs, VariableState]{
		ID:     strconv.Itoa(v.ID),
		Inputs: args,
		State:  st,
	}, nil
}

func (r *Variable) Diff(ctx context.Context, req infer.DiffRequest[VariableArgs, VariableState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// forceNew fields
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if req.Inputs.ProjectID != req.State.ProjectID {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrIntDiffers(req.Inputs.EnvironmentID, req.State.EnvironmentID) {
		diff["environmentId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Updatable fields
	if req.Inputs.Value != req.State.Value {
		diff["value"] = p.PropertyDiff{Kind: p.Update}
	}
	if !strings.EqualFold(req.Inputs.Scope, req.State.Scope) {
		diff["scope"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}
