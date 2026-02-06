package resources

import (
	"context"
	"fmt"
	"strconv"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
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
func (r *Project) Create(ctx context.Context, name string, inputs ProjectArgs, preview bool) (string, ProjectState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"name":      inputs.Name,
		"gitUrl":    inputs.GitURL,
		"openshift": inputs.DeploytargetID,
	}
	setOptional(input, "productionEnvironment", inputs.ProductionEnvironment)
	setOptional(input, "branches", inputs.Branches)
	setOptional(input, "pullrequests", inputs.Pullrequests)
	setOptional(input, "openshiftProjectPattern", inputs.OpenshiftProjectPattern)
	setOptionalInt(input, "autoIdle", inputs.AutoIdle)
	setOptionalInt(input, "storageCalc", inputs.StorageCalc)

	if preview {
		return "preview-id", ProjectState{ProjectArgs: inputs}, nil
	}

	project, err := client.CreateProject(ctx, input)
	if err != nil {
		return "", ProjectState{}, fmt.Errorf("failed to create project: %w", err)
	}

	return strconv.Itoa(project.ID), ProjectState{
		ProjectArgs: inputs,
		LagoonID:    project.ID,
		Created:     project.Created,
	}, nil
}

// Update updates an existing Lagoon project.
func (r *Project) Update(ctx context.Context, id string, olds ProjectState, news ProjectArgs, preview bool) (ProjectState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	input := map[string]any{
		"gitUrl":    news.GitURL,
		"openshift": news.DeploytargetID,
	}
	setOptional(input, "productionEnvironment", news.ProductionEnvironment)
	setOptional(input, "branches", news.Branches)
	setOptional(input, "pullrequests", news.Pullrequests)
	setOptional(input, "openshiftProjectPattern", news.OpenshiftProjectPattern)
	setOptionalInt(input, "autoIdle", news.AutoIdle)
	setOptionalInt(input, "storageCalc", news.StorageCalc)

	if preview {
		return ProjectState{
			ProjectArgs: news,
			LagoonID:    olds.LagoonID,
			Created:     olds.Created,
		}, nil
	}

	_, err := client.UpdateProject(ctx, olds.LagoonID, input)
	if err != nil {
		return ProjectState{}, fmt.Errorf("failed to update project: %w", err)
	}

	return ProjectState{
		ProjectArgs: news,
		LagoonID:    olds.LagoonID,
		Created:     olds.Created,
	}, nil
}

// Delete deletes a Lagoon project.
func (r *Project) Delete(ctx context.Context, id string, props ProjectState) error {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	if err := client.DeleteProject(ctx, props.Name); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// Read reads a Lagoon project for import/refresh.
func (r *Project) Read(ctx context.Context, id string, inputs ProjectArgs, state ProjectState) (string, ProjectArgs, ProjectState, error) {
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	client := cfg.NewClient()

	lagoonID, err := strconv.Atoi(id)
	if err != nil {
		return "", ProjectArgs{}, ProjectState{}, fmt.Errorf("invalid project ID '%s': must be numeric", id)
	}

	project, err := client.GetProjectByID(ctx, lagoonID)
	if err != nil {
		return "", ProjectArgs{}, ProjectState{}, fmt.Errorf("failed to read project: %w", err)
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

	st := ProjectState{
		ProjectArgs: args,
		LagoonID:    project.ID,
		Created:     project.Created,
	}

	return id, args, st, nil
}

// Diff computes the diff between old and new project state.
func (r *Project) Diff(ctx context.Context, id string, olds ProjectState, news ProjectArgs) (p.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}

	// name is forceNew
	if news.Name != olds.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// These can be updated in place
	if news.GitURL != olds.GitURL {
		diff["gitUrl"] = p.PropertyDiff{Kind: p.Update}
	}
	if news.DeploytargetID != olds.DeploytargetID {
		diff["deploytargetId"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.ProductionEnvironment, olds.ProductionEnvironment) {
		diff["productionEnvironment"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.Branches, olds.Branches) {
		diff["branches"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.Pullrequests, olds.Pullrequests) {
		diff["pullrequests"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrDiffers(news.OpenshiftProjectPattern, olds.OpenshiftProjectPattern) {
		diff["openshiftProjectPattern"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrIntDiffers(news.AutoIdle, olds.AutoIdle) {
		diff["autoIdle"] = p.PropertyDiff{Kind: p.Update}
	}
	if ptrIntDiffers(news.StorageCalc, olds.StorageCalc) {
		diff["storageCalc"] = p.PropertyDiff{Kind: p.Update}
	}

	return p.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}
