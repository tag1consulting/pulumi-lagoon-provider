package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// User represents a Lagoon user.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Comment   string `json:"comment"`
}

// UserGroupRole represents a user's role within a group.
type UserGroupRole struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type userWithGroupRoles struct {
	ID         string          `json:"id"`
	Email      string          `json:"email"`
	GroupRoles []UserGroupRole `json:"groupRoles"`
}

type userWithPlatformRoles struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	PlatformRoles []string `json:"platformRoles"`
}

// CreateUser creates a new Lagoon user.
func (c *Client) CreateUser(ctx context.Context, email string, firstName, lastName, comment *string) (*User, error) {
	input := map[string]any{"email": email}
	if firstName != nil {
		input["firstName"] = *firstName
	}
	if lastName != nil {
		input["lastName"] = *lastName
	}
	if comment != nil {
		input["comment"] = *comment
	}

	data, err := c.Execute(ctx, mutationAddUser, map[string]any{"input": input})
	if err != nil {
		return nil, err
	}

	u, err := unmarshalField[User](data, "addUser")
	if err != nil {
		return nil, fmt.Errorf("addUser: %w", err)
	}
	return &u, nil
}

// GetUserByEmail retrieves a user by email address.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	data, err := c.Execute(ctx, queryUserByEmail, map[string]any{"email": email})
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
		}
		return nil, err
	}

	u, err := unmarshalField[User](data, "userByEmail")
	if err != nil {
		return nil, fmt.Errorf("userByEmail: %w", err)
	}
	if u.Email == "" {
		return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
	}
	return &u, nil
}

// UpdateUser updates an existing Lagoon user.
func (c *Client) UpdateUser(ctx context.Context, email string, patch map[string]any) (*User, error) {
	input := map[string]any{
		"user":  map[string]any{"email": email},
		"patch": patch,
	}

	data, err := c.Execute(ctx, mutationUpdateUser, map[string]any{"input": input})
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
		}
		return nil, err
	}

	u, err := unmarshalField[User](data, "updateUser")
	if err != nil {
		return nil, fmt.Errorf("updateUser: %w", err)
	}
	return &u, nil
}

// DeleteUser deletes a Lagoon user by email.
func (c *Client) DeleteUser(ctx context.Context, email string) error {
	input := map[string]any{
		"user": map[string]any{"email": email},
	}

	_, err := c.Execute(ctx, mutationDeleteUser, map[string]any{"input": input})
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return &LagoonNotFoundError{ResourceType: "User", Identifier: email}
		}
	}
	return err
}

// AddUserToGroup assigns a user to a group with a role.
func (c *Client) AddUserToGroup(ctx context.Context, email, groupName, role string) error {
	input := map[string]any{
		"user":  map[string]any{"email": email},
		"group": map[string]any{"name": groupName},
		"role":  strings.ToUpper(role),
	}

	_, err := c.Execute(ctx, mutationAddUserToGroup, map[string]any{"input": input})
	return err
}

// RemoveUserFromGroup removes a user from a group.
func (c *Client) RemoveUserFromGroup(ctx context.Context, email, groupName string) error {
	input := map[string]any{
		"user":  map[string]any{"email": email},
		"group": map[string]any{"name": groupName},
	}

	_, err := c.Execute(ctx, mutationRemoveUserFromGroup, map[string]any{"input": input})
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return &LagoonNotFoundError{ResourceType: "UserGroupAssignment", Identifier: email + ":" + groupName}
		}
	}
	return err
}

// AddPlatformRoleToUser assigns a platform role to a user.
func (c *Client) AddPlatformRoleToUser(ctx context.Context, email, role string) error {
	vars := map[string]any{
		"user": map[string]any{"email": email},
		"role": strings.ToUpper(role),
	}

	_, err := c.Execute(ctx, mutationAddPlatformRoleToUser, vars)
	return err
}

// RemovePlatformRoleFromUser removes a platform role from a user.
func (c *Client) RemovePlatformRoleFromUser(ctx context.Context, email, role string) error {
	vars := map[string]any{
		"user": map[string]any{"email": email},
		"role": strings.ToUpper(role),
	}

	_, err := c.Execute(ctx, mutationRemovePlatformRoleFromUser, vars)
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return &LagoonNotFoundError{ResourceType: "UserPlatformRole", Identifier: email + ":" + role}
		}
	}
	return err
}

// GetUserGroupRoles retrieves a user's group role assignments.
func (c *Client) GetUserGroupRoles(ctx context.Context, email string) ([]UserGroupRole, error) {
	data, err := c.Execute(ctx, queryUserByEmailWithGroupRoles, map[string]any{"email": email})
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
		}
		return nil, err
	}

	u, err := unmarshalField[userWithGroupRoles](data, "userByEmail")
	if err != nil {
		return nil, fmt.Errorf("userByEmail: %w", err)
	}
	if u.Email == "" {
		return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
	}
	// Normalize nil to empty slice so callers can distinguish "user exists with no group
	// roles" (non-nil empty) from "user not found" (error returned above).
	if u.GroupRoles == nil {
		return []UserGroupRole{}, nil
	}
	return u.GroupRoles, nil
}

// GetUserPlatformRoles retrieves a user's platform roles.
func (c *Client) GetUserPlatformRoles(ctx context.Context, email string) ([]string, error) {
	data, err := c.Execute(ctx, queryUserByEmailWithPlatformRoles, map[string]any{"email": email})
	if err != nil {
		var apiErr *LagoonAPIError
		if errors.As(err, &apiErr) && containsNotFound(apiErr.Message) {
			return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
		}
		return nil, err
	}

	u, err := unmarshalField[userWithPlatformRoles](data, "userByEmail")
	if err != nil {
		return nil, fmt.Errorf("userByEmail: %w", err)
	}
	if u.Email == "" {
		return nil, &LagoonNotFoundError{ResourceType: "User", Identifier: email}
	}
	// Normalize nil to empty slice — see GetUserGroupRoles for rationale.
	if u.PlatformRoles == nil {
		return []string{}, nil
	}
	return u.PlatformRoles, nil
}

// containsNotFound checks if an error message indicates a user-not-found condition.
// Lagoon's GraphQL API returns human-readable error strings rather than typed codes, so
// we match a narrow set of known phrases. The substrings must co-occur with "user" to
// avoid misclassifying authorization errors (e.g., "access denied: no user permissions
// for this operation") as NotFound, which would cause silent state removal in Read and
// silent success on Delete — a privilege-retention risk for UserPlatformRole.
func containsNotFound(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "user not found") ||
		strings.Contains(lower, "user does not exist") ||
		strings.Contains(lower, "no user found")
}
