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

func (r *Variable) Create(ctx context.Context, name string, inputs VariableArgs, preview bool) (string, VariableState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if preview {
		return "preview-id", VariableState{VariableArgs: inputs}, nil
	}

	v, err := client.AddVariable(ctx, inputs.Name, inputs.Value, inputs.ProjectID, inputs.Scope, inputs.EnvironmentID)
	if err != nil {
		return "", VariableState{}, fmt.Errorf("failed to create variable: %w", err)
	}

	return strconv.Itoa(v.ID), VariableState{
		VariableArgs: inputs,
		LagoonID:     v.ID,
	}, nil
}

func (r *Variable) Update(ctx context.Context, id string, olds VariableState, news VariableArgs, preview bool) (VariableState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if preview {
		return VariableState{
			VariableArgs: news,
			LagoonID:     olds.LagoonID,
		}, nil
	}

	// Lagoon uses addOrUpdate semantics — same mutation for create and update
	v, err := client.AddVariable(ctx, news.Name, news.Value, news.ProjectID, news.Scope, news.EnvironmentID)
	if err != nil {
		return VariableState{}, fmt.Errorf("failed to update variable: %w", err)
	}

	return VariableState{
		VariableArgs: news,
		LagoonID:     v.ID,
	}, nil
}

func (r *Variable) Delete(ctx context.Context, id string, props VariableState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if err := client.DeleteVariable(ctx, props.Name, props.ProjectID, props.EnvironmentID); err != nil {
		return fmt.Errorf("failed to delete variable: %w", err)
	}
	return nil
}

func (r *Variable) Read(ctx context.Context, id string, inputs VariableArgs, state VariableState) (string, VariableArgs, VariableState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	// Parse import ID: {project_id}:{env_id}:{var_name} or {project_id}::{var_name}
	parts := strings.SplitN(id, ":", 3)
	var projectID int
	var environmentID *int
	var varName string

	if len(parts) == 3 {
		// Import scenario
		pid, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", VariableArgs{}, VariableState{}, fmt.Errorf("invalid variable import ID '%s': project_id must be numeric", id)
		}
		projectID = pid
		varName = parts[2]

		if parts[1] != "" {
			eid, err := strconv.Atoi(parts[1])
			if err != nil {
				return "", VariableArgs{}, VariableState{}, fmt.Errorf("invalid variable import ID '%s': env_id must be numeric", id)
			}
			environmentID = &eid
		}
	} else {
		// Refresh
		projectID = state.ProjectID
		environmentID = state.EnvironmentID
		varName = state.Name
	}

	v, err := client.GetVariable(ctx, varName, projectID, environmentID)
	if err != nil {
		return "", VariableArgs{}, VariableState{}, fmt.Errorf("failed to read variable: %w", err)
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

	return strconv.Itoa(v.ID), args, st, nil
}

func (r *Variable) Diff(ctx context.Context, id string, olds VariableState, news VariableArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// forceNew fields
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if news.ProjectID != olds.ProjectID {
		diff["projectId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}
	if ptrIntDiffers(news.EnvironmentID, olds.EnvironmentID) {
		diff["environmentId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Updatable fields
	if news.Value != olds.Value {
		diff["value"] = p.PropertyDiff{Kind: p.Update}
	}
	if !strings.EqualFold(news.Scope, olds.Scope) {
		diff["scope"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}
