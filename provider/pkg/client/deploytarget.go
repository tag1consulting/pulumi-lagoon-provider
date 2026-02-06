package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// DeployTarget represents a Lagoon Kubernetes deploy target.
type DeployTarget struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ConsoleURL    string `json:"consoleUrl"`
	CloudProvider string `json:"cloudProvider"`
	CloudRegion   string `json:"cloudRegion"`
	SSHHost       string `json:"sshHost"`
	SSHPort       string `json:"sshPort"`
	BuildImage    string `json:"buildImage"`
	Disabled      bool   `json:"disabled"`
	RouterPattern string `json:"routerPattern"`
	Created       string `json:"created"`
}

// CreateDeployTarget creates a new Kubernetes deploy target.
func (c *Client) CreateDeployTarget(ctx context.Context, input map[string]any) (*DeployTarget, error) {
	data, err := c.Execute(ctx, mutationAddKubernetes, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	dt, err := unmarshalField[DeployTarget](data, "addKubernetes")
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

// GetDeployTargetByID retrieves a deploy target by ID (queries all and filters).
func (c *Client) GetDeployTargetByID(ctx context.Context, id int) (*DeployTarget, error) {
	targets, err := c.GetAllDeployTargets(ctx)
	if err != nil {
		return nil, err
	}

	for _, dt := range targets {
		if dt.ID == id {
			return &dt, nil
		}
	}

	return nil, &LagoonNotFoundError{
		ResourceType: "DeployTarget",
		Identifier:   fmt.Sprintf("ID=%d", id),
	}
}

// GetDeployTargetByName retrieves a deploy target by name.
func (c *Client) GetDeployTargetByName(ctx context.Context, name string) (*DeployTarget, error) {
	targets, err := c.GetAllDeployTargets(ctx)
	if err != nil {
		return nil, err
	}

	for _, dt := range targets {
		if dt.Name == name {
			return &dt, nil
		}
	}

	return nil, &LagoonNotFoundError{
		ResourceType: "DeployTarget",
		Identifier:   name,
	}
}

// GetAllDeployTargets retrieves all Kubernetes deploy targets.
func (c *Client) GetAllDeployTargets(ctx context.Context) ([]DeployTarget, error) {
	data, err := c.Execute(ctx, queryAllKubernetes, nil)
	if err != nil {
		return nil, err
	}

	targets, err := unmarshalField[[]DeployTarget](data, "allKubernetes")
	if err != nil {
		return nil, err
	}
	return targets, nil
}

// UpdateDeployTarget updates a deploy target.
func (c *Client) UpdateDeployTarget(ctx context.Context, id int, input map[string]any) (*DeployTarget, error) {
	input["id"] = id
	data, err := c.Execute(ctx, mutationUpdateKubernetes, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	dt, err := unmarshalField[DeployTarget](data, "updateKubernetes")
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

// DeleteDeployTarget deletes a deploy target by name.
func (c *Client) DeleteDeployTarget(ctx context.Context, name string) error {
	_, err := c.Execute(ctx, mutationDeleteKubernetes, map[string]any{
		"input": map[string]any{"name": name},
	})
	return err
}

// DeployTargetConfig represents a deploy target configuration.
type DeployTargetConfig struct {
	ID                         int    `json:"id"`
	Weight                     int    `json:"weight"`
	Branches                   string `json:"branches"`
	Pullrequests               string `json:"pullrequests"`
	DeployTargetProjectPattern string `json:"deployTargetProjectPattern"`
	DeployTargetID             int    `json:"-"` // normalized
	ProjectID                  int    `json:"-"` // normalized
}

type deployTargetConfigRaw struct {
	ID                         int    `json:"id"`
	Weight                     int    `json:"weight"`
	Branches                   string `json:"branches"`
	Pullrequests               string `json:"pullrequests"`
	DeployTargetProjectPattern string `json:"deployTargetProjectPattern"`
	DeployTarget               *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"deployTarget"`
	Project *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
}

func normalizeDeployTargetConfig(raw deployTargetConfigRaw) DeployTargetConfig {
	dtc := DeployTargetConfig{
		ID:                         raw.ID,
		Weight:                     raw.Weight,
		Branches:                   raw.Branches,
		Pullrequests:               raw.Pullrequests,
		DeployTargetProjectPattern: raw.DeployTargetProjectPattern,
	}
	if raw.DeployTarget != nil {
		dtc.DeployTargetID = raw.DeployTarget.ID
	}
	if raw.Project != nil {
		dtc.ProjectID = raw.Project.ID
	}
	return dtc
}

// CreateDeployTargetConfig creates a new deploy target configuration.
func (c *Client) CreateDeployTargetConfig(ctx context.Context, input map[string]any) (*DeployTargetConfig, error) {
	data, err := c.Execute(ctx, mutationAddDeployTargetConfig, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[deployTargetConfigRaw](data, "addDeployTargetConfig")
	if err != nil {
		return nil, err
	}

	dtc := normalizeDeployTargetConfig(raw)
	return &dtc, nil
}

// GetDeployTargetConfigsByProject retrieves all deploy target configs for a project.
func (c *Client) GetDeployTargetConfigsByProject(ctx context.Context, projectID int) ([]DeployTargetConfig, error) {
	data, err := c.Execute(ctx, queryDeployTargetConfigsByProjectId, map[string]any{"project": projectID})
	if err != nil {
		return nil, err
	}

	raw, err := extractField(data, "deployTargetConfigsByProjectId")
	if err != nil {
		return nil, err
	}

	var raws []deployTargetConfigRaw
	if err := json.Unmarshal(raw, &raws); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deploy target configs: %w", err)
	}

	configs := make([]DeployTargetConfig, len(raws))
	for i, r := range raws {
		configs[i] = normalizeDeployTargetConfig(r)
	}
	return configs, nil
}

// GetDeployTargetConfigByID retrieves a deploy target config by ID within a project.
func (c *Client) GetDeployTargetConfigByID(ctx context.Context, configID, projectID int) (*DeployTargetConfig, error) {
	configs, err := c.GetDeployTargetConfigsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		if cfg.ID == configID {
			return &cfg, nil
		}
	}

	return nil, &LagoonNotFoundError{
		ResourceType: "DeployTargetConfig",
		Identifier:   fmt.Sprintf("ID=%d (project=%d)", configID, projectID),
	}
}

// UpdateDeployTargetConfig updates a deploy target configuration.
func (c *Client) UpdateDeployTargetConfig(ctx context.Context, configID int, input map[string]any) (*DeployTargetConfig, error) {
	input["id"] = configID
	data, err := c.Execute(ctx, mutationUpdateDeployTargetConfig, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[deployTargetConfigRaw](data, "updateDeployTargetConfig")
	if err != nil {
		return nil, err
	}

	dtc := normalizeDeployTargetConfig(raw)
	return &dtc, nil
}

// DeleteDeployTargetConfig deletes a deploy target configuration.
func (c *Client) DeleteDeployTargetConfig(ctx context.Context, configID, projectID int) error {
	_, err := c.Execute(ctx, mutationDeleteDeployTargetConfig, map[string]any{
		"input": map[string]any{
			"id":      configID,
			"project": projectID,
		},
	})
	return err
}
