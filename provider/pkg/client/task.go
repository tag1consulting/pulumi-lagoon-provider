package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// TaskDefinition represents a Lagoon advanced task definition.
type TaskDefinition struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Type             string           `json:"type"`
	Service          string           `json:"service"`
	Command          string           `json:"command,omitempty"`
	Image            string           `json:"image,omitempty"`
	Permission       string           `json:"permission"`
	ConfirmationText string           `json:"confirmationText"`
	ProjectID        *int             `json:"-"` // normalized
	EnvironmentID    *int             `json:"-"` // normalized
	GroupName        string           `json:"groupName"`
	Created          string           `json:"created"`
	Arguments        []TaskArgument   `json:"-"` // normalized
}

// TaskArgument represents an argument for a task definition.
type TaskArgument struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
}

// taskDefinitionRaw is used for unmarshaling the API response.
type taskDefinitionRaw struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Type             string          `json:"type"`
	Service          string          `json:"service"`
	Command          string          `json:"command,omitempty"`
	Image            string          `json:"image,omitempty"`
	Permission       string          `json:"permission"`
	ConfirmationText string          `json:"confirmationText"`
	Project          json.RawMessage `json:"project"`
	Environment      json.RawMessage `json:"environment"`
	GroupName        string          `json:"groupName"`
	Created          string          `json:"created"`
	Arguments        []TaskArgument  `json:"advancedTaskDefinitionArguments"`
}

func normalizeTaskDefinition(raw taskDefinitionRaw) TaskDefinition {
	td := TaskDefinition{
		ID:               raw.ID,
		Name:             raw.Name,
		Description:      raw.Description,
		Type:             raw.Type,
		Service:          raw.Service,
		Command:          raw.Command,
		Image:            raw.Image,
		Permission:       raw.Permission,
		ConfirmationText: raw.ConfirmationText,
		GroupName:        raw.GroupName,
		Created:          raw.Created,
		Arguments:        raw.Arguments,
	}

	// Project can be an object {id, name} (old API) or an int (new API)
	if len(raw.Project) > 0 && string(raw.Project) != "null" {
		var obj struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw.Project, &obj); err == nil && obj.ID != 0 {
			td.ProjectID = &obj.ID
		} else {
			var id int
			if err := json.Unmarshal(raw.Project, &id); err == nil && id != 0 {
				td.ProjectID = &id
			}
		}
	}

	// Same for environment
	if len(raw.Environment) > 0 && string(raw.Environment) != "null" {
		var obj struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw.Environment, &obj); err == nil && obj.ID != 0 {
			td.EnvironmentID = &obj.ID
		} else {
			var id int
			if err := json.Unmarshal(raw.Environment, &id); err == nil && id != 0 {
				td.EnvironmentID = &id
			}
		}
	}

	return td
}

// CreateTaskDefinition creates a new advanced task definition.
func (c *Client) CreateTaskDefinition(ctx context.Context, input map[string]any) (*TaskDefinition, error) {
	data, err := c.Execute(ctx, mutationAddAdvancedTaskDefinition, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[taskDefinitionRaw](data, "addAdvancedTaskDefinition")
	if err != nil {
		return nil, err
	}

	td := normalizeTaskDefinition(raw)
	return &td, nil
}

// GetTaskDefinitionByID retrieves a task definition by ID.
func (c *Client) GetTaskDefinitionByID(ctx context.Context, taskID int) (*TaskDefinition, error) {
	data, err := c.Execute(ctx, queryAdvancedTaskDefinitionById, map[string]any{"id": taskID})
	if err != nil {
		return nil, err
	}

	raw, err := unmarshalField[*taskDefinitionRaw](data, "advancedTaskDefinitionById")
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, &LagoonNotFoundError{
			ResourceType: "TaskDefinition",
			Identifier:   fmt.Sprintf("ID=%d", taskID),
		}
	}

	td := normalizeTaskDefinition(*raw)
	return &td, nil
}

// DeleteTaskDefinition deletes a task definition.
func (c *Client) DeleteTaskDefinition(ctx context.Context, taskID int) error {
	_, err := c.Execute(ctx, mutationDeleteAdvancedTaskDefinition, map[string]any{"id": taskID})
	return err
}

// GetTasksByEnvironment retrieves task definitions for an environment.
// Uses the new API with fallback to legacy.
func (c *Client) GetTasksByEnvironment(ctx context.Context, environmentID int) ([]TaskDefinition, error) {
	data, err := c.Execute(ctx, queryAdvancedTasksForEnvironmentNew, map[string]any{"environment": environmentID})
	if err != nil {
		var apiErr *LagoonAPIError
		if isAPIError(err, &apiErr) && (strings.Contains(apiErr.Message, "Cannot query field") || strings.Contains(apiErr.Message, "400")) {
			return c.getTasksByEnvironmentLegacy(ctx, environmentID)
		}
		return nil, err
	}

	return c.unmarshalTasks(data, "advancedTasksForEnvironment")
}

func (c *Client) getTasksByEnvironmentLegacy(ctx context.Context, environmentID int) ([]TaskDefinition, error) {
	data, err := c.Execute(ctx, queryAdvancedTasksByEnvironmentOld, map[string]any{"environment": environmentID})
	if err != nil {
		return nil, err
	}

	return c.unmarshalTasks(data, "advancedTasksByEnvironment")
}

func (c *Client) unmarshalTasks(data json.RawMessage, field string) ([]TaskDefinition, error) {
	raw, err := extractField(data, field)
	if err != nil {
		return nil, err
	}

	var raws []taskDefinitionRaw
	if err := json.Unmarshal(raw, &raws); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	tasks := make([]TaskDefinition, len(raws))
	for i, r := range raws {
		tasks[i] = normalizeTaskDefinition(r)
	}
	return tasks, nil
}
