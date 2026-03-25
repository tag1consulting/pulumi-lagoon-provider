package client

import (
	"context"
	"fmt"
)

// Group represents a Lagoon group.
type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CreateGroup creates a new Lagoon group.
func (c *Client) CreateGroup(ctx context.Context, name string, parentGroupName *string) (*Group, error) {
	input := map[string]any{"name": name}
	if parentGroupName != nil {
		input["parentGroup"] = map[string]any{"name": *parentGroupName}
	}

	data, err := c.Execute(ctx, mutationAddGroup, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	g, err := unmarshalField[Group](data, "addGroup")
	if err != nil {
		return nil, fmt.Errorf("addGroup: %w", err)
	}
	return &g, nil
}

// GetGroupByName retrieves a group by name.
func (c *Client) GetGroupByName(ctx context.Context, name string) (*Group, error) {
	data, err := c.Execute(ctx, queryAllGroups, nil)
	if err != nil {
		return nil, err
	}

	groups, err := unmarshalField[[]Group](data, "allGroups")
	if err != nil {
		return nil, fmt.Errorf("allGroups: %w", err)
	}

	for _, g := range groups {
		if g.Name == name {
			return &g, nil
		}
	}

	return nil, &LagoonNotFoundError{ResourceType: "Group", Identifier: name}
}

// UpdateGroup updates an existing Lagoon group.
func (c *Client) UpdateGroup(ctx context.Context, name string, patch map[string]any) (*Group, error) {
	input := map[string]any{
		"group": map[string]any{"name": name},
		"patch": patch,
	}

	data, err := c.Execute(ctx, mutationUpdateGroup, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	g, err := unmarshalField[Group](data, "updateGroup")
	if err != nil {
		return nil, fmt.Errorf("updateGroup: %w", err)
	}
	return &g, nil
}

// DeleteGroup deletes a Lagoon group by name.
func (c *Client) DeleteGroup(ctx context.Context, name string) error {
	input := map[string]any{
		"group": map[string]any{"name": name},
	}

	_, err := c.Execute(ctx, mutationDeleteGroup, map[string]any{"input": input})
	return err
}
