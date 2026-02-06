package client

import (
	"context"
	"fmt"
)

// Environment represents a Lagoon environment.
type Environment struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	ProjectID       int    `json:"-"` // normalized from project object
	EnvironmentType string `json:"environmentType"`
	DeployType      string `json:"deployType"`
	DeployBaseRef   string `json:"deployBaseRef"`
	DeployHeadRef   string `json:"deployHeadRef"`
	DeployTitle     string `json:"deployTitle"`
	AutoIdle        *int   `json:"autoIdle"`
	Route           string `json:"route"`
	Routes          string `json:"routes"`
	Created         string `json:"created"`
}

// environmentRaw is used for unmarshaling the API response which nests project.
type environmentRaw struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Project         *struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	EnvironmentType string `json:"environmentType"`
	DeployType      string `json:"deployType"`
	DeployBaseRef   string `json:"deployBaseRef"`
	DeployHeadRef   string `json:"deployHeadRef"`
	DeployTitle     string `json:"deployTitle"`
	AutoIdle        *int   `json:"autoIdle"`
	Route           string `json:"route"`
	Routes          string `json:"routes"`
	Created         string `json:"created"`
}

func normalizeEnvironment(raw environmentRaw) Environment {
	e := Environment{
		ID:              raw.ID,
		Name:            raw.Name,
		EnvironmentType: raw.EnvironmentType,
		DeployType:      raw.DeployType,
		DeployBaseRef:   raw.DeployBaseRef,
		DeployHeadRef:   raw.DeployHeadRef,
		DeployTitle:     raw.DeployTitle,
		AutoIdle:        raw.AutoIdle,
		Route:           raw.Route,
		Routes:          raw.Routes,
		Created:         raw.Created,
	}
	if raw.Project != nil {
		e.ProjectID = raw.Project.ID
	}
	return e
}

// AddOrUpdateEnvironment creates or updates an environment.
func (c *Client) AddOrUpdateEnvironment(ctx context.Context, input map[string]any) (*Environment, error) {
	data, err := c.Execute(ctx, mutationAddOrUpdateEnvironment, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[environmentRaw](data, "addOrUpdateEnvironment")
	if err != nil {
		return nil, err
	}

	env := normalizeEnvironment(raw)
	return &env, nil
}

// GetEnvironmentByName retrieves an environment by name and project ID.
func (c *Client) GetEnvironmentByName(ctx context.Context, name string, projectID int) (*Environment, error) {
	data, err := c.Execute(ctx, queryEnvironmentByName, map[string]any{
		"name":    name,
		"project": projectID,
	})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[*environmentRaw](data, "environmentByName")
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, &LagoonNotFoundError{
			ResourceType: "Environment",
			Identifier:   fmt.Sprintf("%s (project=%d)", name, projectID),
		}
	}

	env := normalizeEnvironment(*raw)
	return &env, nil
}

// DeleteEnvironment deletes an environment.
func (c *Client) DeleteEnvironment(ctx context.Context, name string, projectID int) error {
	_, err := c.Execute(ctx, mutationDeleteEnvironment, map[string]any{
		"input": map[string]any{
			"name":    name,
			"project": projectID,
			"execute": true,
		},
	})
	return err
}
