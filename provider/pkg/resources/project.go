package resources

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// Project manages a Lagoon project resource.
type Project struct{}

// ProjectArgs are the inputs for a Lagoon project.
type ProjectArgs struct {
	Name                    string  `pulumi:"name"`
	GitURL                  string  `pulumi:"gitUrl"`
	DeploytargetID          int     `pulumi:"deploytargetId"`
	ProductionEnvironment   *string `pulumi:"productionEnvironment,optional"`
	Branches                *string `pulumi:"branches,optional"`
	Pullrequests            *string `pulumi:"pullrequests,optional"`
	OpenshiftProjectPattern *string `pulumi:"openshiftProjectPattern,optional"`
	AutoIdle                *int    `pulumi:"autoIdle,optional"`
	StorageCalc             *int    `pulumi:"storageCalc,optional"`
}

// ProjectState is the full state of a Lagoon project.
type ProjectState struct {
	ProjectArgs
	LagoonID int    `pulumi:"lagoonId"`
	Created  string `pulumi:"created"`
}

// Annotate sets the resource token and field descriptions.
func (r *Project) Annotate(a infer.Annotator) {
	a.SetToken("lagoon", "Project")
	a.Describe(&r, "Manages a Lagoon project (application/site).")
}

func (a *ProjectArgs) Annotate(an infer.Annotator) {
	an.Describe(&a.Name, "The project name. Must be unique, lowercase alphanumeric with hyphens.")
	an.Describe(&a.GitURL, "The Git repository URL (SSH or HTTPS format).")
	an.Describe(&a.DeploytargetID, "The deploy target (Kubernetes cluster) ID.")
	an.Describe(&a.ProductionEnvironment, "Name of the production branch/environment.")
	an.Describe(&a.Branches, "Regex pattern for branches to deploy.")
	an.Describe(&a.Pullrequests, "Regex pattern for pull requests to deploy.")
	an.Describe(&a.OpenshiftProjectPattern, "Namespace pattern for the project on the cluster.")
	an.Describe(&a.AutoIdle, "Whether to auto-idle environments (1=yes, 0=no).")
	an.Describe(&a.StorageCalc, "Whether to calculate storage (1=yes, 0=no).")
}

func (s *ProjectState) Annotate(a infer.Annotator) {
	a.Describe(&s.LagoonID, "The Lagoon-assigned numeric ID of the project.")
	a.Describe(&s.Created, "The creation timestamp of the project.")
}

// Create creates a new Lagoon project.
func (r *Project) Create(ctx context.Context, req infer.CreateRequest[ProjectArgs]) (infer.CreateResponse[ProjectState], error) {
	c := clientFor(ctx)

	input := map[string]any{
		"name":      req.Inputs.Name,
		"gitUrl":    req.Inputs.GitURL,
		"openshift": req.Inputs.DeploytargetID,
	}
	setOptional(input, "productionEnvironment", req.Inputs.ProductionEnvironment)
	setOptional(input, "branches", req.Inputs.Branches)
	setOptional(input, "pullrequests", req.Inputs.Pullrequests)
	setOptional(input, "openshiftProjectPattern", req.Inputs.OpenshiftProjectPattern)
	setOptionalInt(input, "autoIdle", req.Inputs.AutoIdle)
	setOptionalInt(input, "storageCalc", req.Inputs.StorageCalc)

	if req.DryRun {
		return infer.CreateResponse[ProjectState]{
			ID:     "preview-id",
			Output: ProjectState{ProjectArgs: req.Inputs},
		}, nil
	}

	project, err := c.CreateProject(ctx, input)
	if err != nil {
		if !client.IsDuplicateEntry(err) {
			return infer.CreateResponse[ProjectState]{}, fmt.Errorf("failed to create project: %w", err)
		}

		// Resource already exists — adopt it by looking up by name and updating
		existing, lookupErr := c.GetProjectByName(ctx, req.Inputs.Name)
		if lookupErr != nil {
			return infer.CreateResponse[ProjectState]{}, fmt.Errorf("project %q already exists but failed to look up: %w", req.Inputs.Name, lookupErr)
		}

		// Update the existing resource to match desired inputs
		// Remove name — it's not updatable
		delete(input, "name")
		project, err = c.UpdateProject(ctx, existing.ID, input)
		if err != nil {
			return infer.CreateResponse[ProjectState]{}, fmt.Errorf("project %q already exists but failed to update: %w", req.Inputs.Name, err)
		}
	}

	return infer.CreateResponse[ProjectState]{
		ID: strconv.Itoa(project.ID),
		Output: ProjectState{
			ProjectArgs: req.Inputs,
			LagoonID:    project.ID,
			Created:     project.Created,
		},
	}, nil
}

// Update updates an existing Lagoon project.
func (r *Project) Update(ctx context.Context, req infer.UpdateRequest[ProjectArgs, ProjectState]) (infer.UpdateResponse[ProjectState], error) {
	c := clientFor(ctx)

	input := map[string]any{
		"gitUrl":    req.Inputs.GitURL,
		"openshift": req.Inputs.DeploytargetID,
	}
	setOptional(input, "productionEnvironment", req.Inputs.ProductionEnvironment)
	setOptional(input, "branches", req.Inputs.Branches)
	setOptional(input, "pullrequests", req.Inputs.Pullrequests)
	setOptional(input, "openshiftProjectPattern", req.Inputs.OpenshiftProjectPattern)
	setOptionalInt(input, "autoIdle", req.Inputs.AutoIdle)
	setOptionalInt(input, "storageCalc", req.Inputs.StorageCalc)

	if req.DryRun {
		return infer.UpdateResponse[ProjectState]{
			Output: ProjectState{
				ProjectArgs: req.Inputs,
				LagoonID:    req.State.LagoonID,
				Created:     req.State.Created,
			},
		}, nil
	}

	_, err := c.UpdateProject(ctx, req.State.LagoonID, input)
	if err != nil {
		return infer.UpdateResponse[ProjectState]{}, fmt.Errorf("failed to update project: %w", err)
	}

	return infer.UpdateResponse[ProjectState]{
		Output: ProjectState{
			ProjectArgs: req.Inputs,
			LagoonID:    req.State.LagoonID,
			Created:     req.State.Created,
		},
	}, nil
}

// Delete deletes a Lagoon project.
func (r *Project) Delete(ctx context.Context, req infer.DeleteRequest[ProjectState]) (infer.DeleteResponse, error) {
	c := clientFor(ctx)

	if err := c.DeleteProject(ctx, req.State.Name); err != nil {
		// Treat "not found" as success — resource is already gone
		if errors.Is(err, client.ErrNotFound) {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete project: %w", err)
	}
	return infer.DeleteResponse{}, nil
}

// Read reads a Lagoon project for import/refresh.
func (r *Project) Read(ctx context.Context, req infer.ReadRequest[ProjectArgs, ProjectState]) (infer.ReadResponse[ProjectArgs, ProjectState], error) {
	c := clientFor(ctx)

	lagoonID, err := strconv.Atoi(req.ID)
	if err != nil {
		return infer.ReadResponse[ProjectArgs, ProjectState]{}, fmt.Errorf("invalid project ID '%s': must be numeric", req.ID)
	}

	project, err := c.GetProjectByID(ctx, lagoonID)
	if err != nil {
		// Return empty ID to signal the resource was deleted from Lagoon
		if errors.Is(err, client.ErrNotFound) {
			return infer.ReadResponse[ProjectArgs, ProjectState]{}, nil
		}
		return infer.ReadResponse[ProjectArgs, ProjectState]{}, fmt.Errorf("failed to read project: %w", err)
	}

	args := ProjectArgs{
		Name:           project.Name,
		GitURL:         project.GitURL,
		DeploytargetID: project.OpenshiftID,
	}
	if project.ProductionEnvironment != "" {
		args.ProductionEnvironment = &project.ProductionEnvironment
	}
	if project.Branches != "" {
		args.Branches = &project.Branches
	}
	if project.Pullrequests != "" {
		args.Pullrequests = &project.Pullrequests
	}
	if project.OpenshiftProjectPattern != "" {
		args.OpenshiftProjectPattern = &project.OpenshiftProjectPattern
	}
	args.AutoIdle = project.AutoIdle
	args.StorageCalc = project.StorageCalc

	st := ProjectState{
		ProjectArgs: args,
		LagoonID:    project.ID,
		Created:     project.Created,
	}

	return infer.ReadResponse[ProjectArgs, ProjectState]{
		ID:     req.ID,
		Inputs: args,
		State:  st,
	}, nil
}

// Diff computes the diff between old and new project state.
func (r *Project) Diff(ctx context.Context, req infer.DiffRequest[ProjectArgs, ProjectState]) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// name is forceNew
	if req.Inputs.Name != req.State.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// These can be updated in place
	if req.Inputs.GitURL != req.State.GitURL {
		diff["gitUrl"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.DeploytargetID != req.State.DeploytargetID {
		diff["deploytargetId"] = p.PropertyDiff{Kind: p.Update}
	}
	// All pointer fields use nil-means-unmanaged semantics: if the user's
	// input is nil they don't want to manage the field, so skip the diff even
	// if the API returns a value.  This prevents spurious updates on refresh.
	if req.Inputs.ProductionEnvironment != nil && ptrDiffers(req.Inputs.ProductionEnvironment, req.State.ProductionEnvironment) {
		diff["productionEnvironment"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Branches != nil && ptrDiffers(req.Inputs.Branches, req.State.Branches) {
		diff["branches"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.Pullrequests != nil && ptrDiffers(req.Inputs.Pullrequests, req.State.Pullrequests) {
		diff["pullrequests"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.OpenshiftProjectPattern != nil && ptrDiffers(req.Inputs.OpenshiftProjectPattern, req.State.OpenshiftProjectPattern) {
		diff["openshiftProjectPattern"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.AutoIdle != nil && ptrIntDiffers(req.Inputs.AutoIdle, req.State.AutoIdle) {
		diff["autoIdle"] = p.PropertyDiff{Kind: p.Update}
	}
	if req.Inputs.StorageCalc != nil && ptrIntDiffers(req.Inputs.StorageCalc, req.State.StorageCalc) {
		diff["storageCalc"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}
