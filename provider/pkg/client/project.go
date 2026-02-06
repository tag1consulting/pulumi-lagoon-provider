package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Project represents a Lagoon project.
type Project struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	GitURL                string `json:"gitUrl"`
	OpenshiftID           int    `json:"-"` // normalized from openshift object
	ProductionEnvironment string `json:"productionEnvironment"`
	Branches              string `json:"branches"`
	Pullrequests          string `json:"pullrequests"`
	Created               string `json:"created"`
}

// projectRaw is used for unmarshaling the API response which nests openshift.
type projectRaw struct {
	ID                    int             `json:"id"`
	Name                  string          `json:"name"`
	GitURL                string          `json:"gitUrl"`
	Openshift             json.RawMessage `json:"openshift"`
	ProductionEnvironment string          `json:"productionEnvironment"`
	Branches              string          `json:"branches"`
	Pullrequests          string          `json:"pullrequests"`
	Created               string          `json:"created"`
}

func normalizeProject(raw projectRaw) Project {
	p := Project{
		ID:                    raw.ID,
		Name:                  raw.Name,
		GitURL:                raw.GitURL,
		ProductionEnvironment: raw.ProductionEnvironment,
		Branches:              raw.Branches,
		Pullrequests:          raw.Pullrequests,
		Created:               raw.Created,
	}

	// Normalize openshift to just the ID
	if len(raw.Openshift) > 0 {
		var osObj struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw.Openshift, &osObj); err == nil {
			p.OpenshiftID = osObj.ID
		}
	}

	return p
}

// CreateProject creates a new Lagoon project.
func (c *Client) CreateProject(ctx context.Context, input map[string]any) (*Project, error) {
	data, err := c.Execute(ctx, mutationAddProject, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[projectRaw](data, "addProject")
	if err != nil {
		return nil, err
	}

	p := normalizeProject(raw)
	return &p, nil
}

// GetProjectByName retrieves a project by name.
func (c *Client) GetProjectByName(ctx context.Context, name string) (*Project, error) {
	data, err := c.Execute(ctx, queryProjectByName, map[string]any{"name": name})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[*projectRaw](data, "projectByName")
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, &LagoonNotFoundError{ResourceType: "Project", Identifier: name}
	}

	p := normalizeProject(*raw)
	return &p, nil
}

// GetProjectByID retrieves a project by ID (queries all projects and filters).
func (c *Client) GetProjectByID(ctx context.Context, projectID int) (*Project, error) {
	data, err := c.Execute(ctx, queryAllProjects, nil)
	if err != nil {
		return nil, err
	}

	var rawProjects []projectRaw
	raw, err := extractField(data, "allProjects")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &rawProjects); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allProjects: %w", err)
	}

	for _, rp := range rawProjects {
		if rp.ID == projectID {
			p := normalizeProject(rp)
			return &p, nil
		}
	}

	return nil, &LagoonNotFoundError{ResourceType: "Project", Identifier: fmt.Sprintf("ID=%d", projectID)}
}

// UpdateProject updates an existing project.
func (c *Client) UpdateProject(ctx context.Context, projectID int, input map[string]any) (*Project, error) {
	input["id"] = projectID
	data, err := c.Execute(ctx, mutationUpdateProject, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[projectRaw](data, "updateProject")
	if err != nil {
		return nil, err
	}

	p := normalizeProject(raw)
	return &p, nil
}

// DeleteProject deletes a project by name.
func (c *Client) DeleteProject(ctx context.Context, name string) error {
	_, err := c.Execute(ctx, mutationDeleteProject, map[string]any{
		"input": map[string]any{"project": name},
	})
	return err
}
