package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Variable represents a Lagoon environment variable.
type Variable struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

// AddVariable creates or updates an environment variable.
// Uses the new API (v2.30.0+) with fallback to legacy.
func (c *Client) AddVariable(ctx context.Context, name, value string, projectID int, scope string, environmentID *int) (*Variable, error) {
	if c.IsNewAPI(ctx) {
		return c.addVariableNew(ctx, name, value, projectID, scope, environmentID)
	}
	return c.addVariableLegacy(ctx, name, value, projectID, scope, environmentID)
}

func (c *Client) addVariableNew(ctx context.Context, name, value string, projectID int, scope string, environmentID *int) (*Variable, error) {
	// New API uses project name, not ID
	project, err := c.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up project for variable: %w", err)
	}

	input := map[string]any{
		"name":    name,
		"value":   value,
		"scope":   strings.ToUpper(scope),
		"project": project.Name,
	}

	if environmentID != nil {
		envName, err := c.getEnvironmentName(ctx, *environmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to look up environment for variable: %w", err)
		}
		input["environment"] = envName
	}

	data, err := c.Execute(ctx, mutationAddOrUpdateEnvVariableByName, map[string]any{"input": input})
	if err != nil {
		// Fallback on "Cannot query field" errors
		var apiErr *LagoonAPIError
		if isAPIError(err, &apiErr) && isFieldNotFoundError(apiErr) {
			return c.addVariableLegacy(ctx, name, value, projectID, scope, environmentID)
		}
		return nil, err
	}

	v, err := unmarshalField[Variable](data, "addOrUpdateEnvVariableByName")
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) addVariableLegacy(ctx context.Context, name, value string, projectID int, scope string, environmentID *int) (*Variable, error) {
	input := map[string]any{
		"name":  name,
		"value": value,
		"scope": strings.ToUpper(scope),
	}

	if environmentID != nil {
		input["type"] = "ENVIRONMENT"
		input["typeId"] = *environmentID
	} else {
		input["type"] = "PROJECT"
		input["typeId"] = projectID
	}

	data, err := c.Execute(ctx, mutationAddEnvVariable, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	v, err := unmarshalField[Variable](data, "addEnvVariable")
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// GetVariable retrieves a variable by name within a project (and optionally environment).
func (c *Client) GetVariable(ctx context.Context, name string, projectID int, environmentID *int) (*Variable, error) {
	if c.IsNewAPI(ctx) {
		v, err := c.getVariableNew(ctx, name, projectID, environmentID)
		if err != nil {
			var apiErr *LagoonAPIError
			if isAPIError(err, &apiErr) && isFieldNotFoundError(apiErr) {
				return c.getVariableLegacy(ctx, name, projectID, environmentID)
			}
			return v, err
		}
		return v, nil
	}
	return c.getVariableLegacy(ctx, name, projectID, environmentID)
}

func (c *Client) getVariableNew(ctx context.Context, name string, projectID int, environmentID *int) (*Variable, error) {
	project, err := c.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	input := map[string]any{"project": project.Name}

	if environmentID != nil {
		envName, err := c.getEnvironmentName(ctx, *environmentID)
		if err != nil {
			return nil, err
		}
		input["environment"] = envName
	}

	data, err := c.Execute(ctx, queryGetEnvVariablesByProjectEnvironmentName, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	vars, err := unmarshalField[[]Variable](data, "getEnvVariablesByProjectEnvironmentName")
	if err != nil {
		return nil, err
	}

	for _, v := range vars {
		if v.Name == name {
			return &v, nil
		}
	}

	return nil, &LagoonNotFoundError{
		ResourceType: "Variable",
		Identifier:   fmt.Sprintf("%s (project=%d)", name, projectID),
	}
}

func (c *Client) getVariableLegacy(ctx context.Context, name string, projectID int, environmentID *int) (*Variable, error) {
	vars := map[string]any{"project": projectID}
	if environmentID != nil {
		vars["environment"] = *environmentID
	}

	data, err := c.Execute(ctx, queryEnvVariablesByProjectEnvironment, vars)
	if err != nil {
		return nil, err
	}

	allVars, err := unmarshalField[[]Variable](data, "envVariablesByProjectEnvironment")
	if err != nil {
		return nil, err
	}

	for _, v := range allVars {
		if v.Name == name {
			return &v, nil
		}
	}

	return nil, &LagoonNotFoundError{
		ResourceType: "Variable",
		Identifier:   fmt.Sprintf("%s (project=%d)", name, projectID),
	}
}

// DeleteVariable deletes an environment variable.
func (c *Client) DeleteVariable(ctx context.Context, name string, projectID int, environmentID *int) error {
	if c.IsNewAPI(ctx) {
		err := c.deleteVariableNew(ctx, name, projectID, environmentID)
		if err != nil {
			var apiErr *LagoonAPIError
			if isAPIError(err, &apiErr) && isFieldNotFoundError(apiErr) {
				return c.deleteVariableLegacy(ctx, name, projectID, environmentID)
			}
			return err
		}
		return nil
	}
	return c.deleteVariableLegacy(ctx, name, projectID, environmentID)
}

func (c *Client) deleteVariableNew(ctx context.Context, name string, projectID int, environmentID *int) error {
	project, err := c.GetProjectByID(ctx, projectID)
	if err != nil {
		return err
	}

	input := map[string]any{
		"name":    name,
		"project": project.Name,
	}

	if environmentID != nil {
		envName, err := c.getEnvironmentName(ctx, *environmentID)
		if err != nil {
			return err
		}
		input["environment"] = envName
	}

	_, err = c.Execute(ctx, mutationDeleteEnvVariableByName, map[string]any{"input": input})
	return err
}

func (c *Client) deleteVariableLegacy(ctx context.Context, name string, projectID int, environmentID *int) error {
	input := map[string]any{
		"name":    name,
		"project": projectID,
	}
	if environmentID != nil {
		input["environment"] = *environmentID
	}

	_, err := c.Execute(ctx, mutationDeleteEnvVariable, map[string]any{"input": input})
	return err
}

// getEnvironmentName looks up an environment name by ID.
func (c *Client) getEnvironmentName(ctx context.Context, envID int) (string, error) {
	data, err := c.Execute(ctx, queryEnvironmentById, map[string]any{"id": envID})
	if err != nil {
		return "", err
	}

	var env struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	raw, err := extractField(data, "environmentById")
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", fmt.Errorf("failed to unmarshal environmentById: %w", err)
	}

	return env.Name, nil
}

// isAPIError checks if an error is a LagoonAPIError and sets the target pointer.
func isAPIError(err error, target **LagoonAPIError) bool {
	e, ok := err.(*LagoonAPIError)
	if ok && target != nil {
		*target = e
	}
	return ok
}

// isFieldNotFoundError returns true if the API error is a "Cannot query field" error,
// which indicates the API version doesn't support the query.
func isFieldNotFoundError(err *LagoonAPIError) bool {
	return strings.Contains(err.Message, "Cannot query field") ||
		strings.Contains(err.Message, "Unknown argument")
}
