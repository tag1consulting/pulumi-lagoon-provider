package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Notification represents a generic Lagoon notification.
type Notification struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Webhook      string `json:"webhook,omitempty"`
	Channel      string `json:"channel,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	TypeName     string `json:"__typename,omitempty"`
}

// notificationTypeConfig describes the GraphQL operations for a notification type.
type notificationTypeConfig struct {
	typeName       string // e.g. "NotificationSlack"
	createMutation string
	createField    string // response field for create, e.g. "addNotificationSlack"
	updateMutation string
	updateField    string // response field for update
	deleteMutation string
}

var (
	notifSlack = notificationTypeConfig{
		typeName:       "NotificationSlack",
		createMutation: mutationAddNotificationSlack,
		createField:    "addNotificationSlack",
		updateMutation: mutationUpdateNotificationSlack,
		updateField:    "updateNotificationSlack",
		deleteMutation: mutationDeleteNotificationSlack,
	}
	notifRocketChat = notificationTypeConfig{
		typeName:       "NotificationRocketChat",
		createMutation: mutationAddNotificationRocketChat,
		createField:    "addNotificationRocketChat",
		updateMutation: mutationUpdateNotificationRocketChat,
		updateField:    "updateNotificationRocketChat",
		deleteMutation: mutationDeleteNotificationRocketChat,
	}
	notifEmail = notificationTypeConfig{
		typeName:       "NotificationEmail",
		createMutation: mutationAddNotificationEmail,
		createField:    "addNotificationEmail",
		updateMutation: mutationUpdateNotificationEmail,
		updateField:    "updateNotificationEmail",
		deleteMutation: mutationDeleteNotificationEmail,
	}
	notifMicrosoftTeams = notificationTypeConfig{
		typeName:       "NotificationMicrosoftTeams",
		createMutation: mutationAddNotificationMicrosoftTeams,
		createField:    "addNotificationMicrosoftTeams",
		updateMutation: mutationUpdateNotificationMicrosoftTeams,
		updateField:    "updateNotificationMicrosoftTeams",
		deleteMutation: mutationDeleteNotificationMicrosoftTeams,
	}
)

// getAllNotifications retrieves all notifications of all types.
func (c *Client) getAllNotifications(ctx context.Context) ([]Notification, error) {
	data, err := c.Execute(ctx, queryAllNotifications, nil)
	if err != nil {
		return nil, err
	}
	return unmarshalField[[]Notification](data, "allNotifications")
}

// createNotification creates a notification of the given type with the provided fields.
func (c *Client) createNotification(ctx context.Context, cfg notificationTypeConfig, fields map[string]any) (*Notification, error) {
	data, err := c.Execute(ctx, cfg.createMutation, map[string]any{"input": fields})
	if err != nil {
		return nil, err
	}
	n, err := unmarshalField[Notification](data, cfg.createField)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// getNotificationByName retrieves a notification of the given type by name.
func (c *Client) getNotificationByName(ctx context.Context, cfg notificationTypeConfig, name string) (*Notification, error) {
	all, err := c.getAllNotifications(ctx)
	if err != nil {
		return nil, err
	}
	// An empty result from allNotifications is suspicious when we're looking for a
	// specific notification: it may indicate an API permissions issue rather than
	// genuine deletion. Return an error so callers don't silently remove state.
	if len(all) == 0 {
		return nil, fmt.Errorf("allNotifications returned no results; cannot confirm %s %q was deleted (possible API permissions issue)", cfg.typeName, name)
	}
	for _, n := range all {
		if n.TypeName == cfg.typeName && n.Name == name {
			return &n, nil
		}
	}
	return nil, &LagoonNotFoundError{ResourceType: cfg.typeName, Identifier: name}
}

// updateNotification updates a notification of the given type.
func (c *Client) updateNotification(ctx context.Context, cfg notificationTypeConfig, name string, patch map[string]any) (*Notification, error) {
	data, err := c.Execute(ctx, cfg.updateMutation, map[string]any{
		"input": map[string]any{"name": name, "patch": patch},
	})
	if err != nil {
		return nil, err
	}
	n, err := unmarshalField[Notification](data, cfg.updateField)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// deleteNotification deletes a notification of the given type by name.
func (c *Client) deleteNotification(ctx context.Context, cfg notificationTypeConfig, name string) error {
	_, err := c.Execute(ctx, cfg.deleteMutation, map[string]any{
		"input": map[string]any{"name": name},
	})
	return err
}

// --- Slack ---

// CreateNotificationSlack creates a Slack notification.
func (c *Client) CreateNotificationSlack(ctx context.Context, name, webhook, channel string) (*Notification, error) {
	return c.createNotification(ctx, notifSlack, map[string]any{"name": name, "webhook": webhook, "channel": channel})
}

// GetNotificationSlackByName retrieves a Slack notification by name.
func (c *Client) GetNotificationSlackByName(ctx context.Context, name string) (*Notification, error) {
	return c.getNotificationByName(ctx, notifSlack, name)
}

// UpdateNotificationSlack updates a Slack notification.
func (c *Client) UpdateNotificationSlack(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	return c.updateNotification(ctx, notifSlack, name, patch)
}

// DeleteNotificationSlack deletes a Slack notification.
func (c *Client) DeleteNotificationSlack(ctx context.Context, name string) error {
	return c.deleteNotification(ctx, notifSlack, name)
}

// --- RocketChat ---

// CreateNotificationRocketChat creates a RocketChat notification.
func (c *Client) CreateNotificationRocketChat(ctx context.Context, name, webhook, channel string) (*Notification, error) {
	return c.createNotification(ctx, notifRocketChat, map[string]any{"name": name, "webhook": webhook, "channel": channel})
}

// GetNotificationRocketChatByName retrieves a RocketChat notification by name.
func (c *Client) GetNotificationRocketChatByName(ctx context.Context, name string) (*Notification, error) {
	return c.getNotificationByName(ctx, notifRocketChat, name)
}

// UpdateNotificationRocketChat updates a RocketChat notification.
func (c *Client) UpdateNotificationRocketChat(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	return c.updateNotification(ctx, notifRocketChat, name, patch)
}

// DeleteNotificationRocketChat deletes a RocketChat notification.
func (c *Client) DeleteNotificationRocketChat(ctx context.Context, name string) error {
	return c.deleteNotification(ctx, notifRocketChat, name)
}

// --- Email ---

// CreateNotificationEmail creates an Email notification.
func (c *Client) CreateNotificationEmail(ctx context.Context, name, emailAddress string) (*Notification, error) {
	return c.createNotification(ctx, notifEmail, map[string]any{"name": name, "emailAddress": emailAddress})
}

// GetNotificationEmailByName retrieves an Email notification by name.
func (c *Client) GetNotificationEmailByName(ctx context.Context, name string) (*Notification, error) {
	return c.getNotificationByName(ctx, notifEmail, name)
}

// UpdateNotificationEmail updates an Email notification.
func (c *Client) UpdateNotificationEmail(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	return c.updateNotification(ctx, notifEmail, name, patch)
}

// DeleteNotificationEmail deletes an Email notification.
func (c *Client) DeleteNotificationEmail(ctx context.Context, name string) error {
	return c.deleteNotification(ctx, notifEmail, name)
}

// --- Microsoft Teams ---

// CreateNotificationMicrosoftTeams creates a Microsoft Teams notification.
func (c *Client) CreateNotificationMicrosoftTeams(ctx context.Context, name, webhook string) (*Notification, error) {
	return c.createNotification(ctx, notifMicrosoftTeams, map[string]any{"name": name, "webhook": webhook})
}

// GetNotificationMicrosoftTeamsByName retrieves a Microsoft Teams notification by name.
func (c *Client) GetNotificationMicrosoftTeamsByName(ctx context.Context, name string) (*Notification, error) {
	return c.getNotificationByName(ctx, notifMicrosoftTeams, name)
}

// UpdateNotificationMicrosoftTeams updates a Microsoft Teams notification.
func (c *Client) UpdateNotificationMicrosoftTeams(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	return c.updateNotification(ctx, notifMicrosoftTeams, name, patch)
}

// DeleteNotificationMicrosoftTeams deletes a Microsoft Teams notification.
func (c *Client) DeleteNotificationMicrosoftTeams(ctx context.Context, name string) error {
	return c.deleteNotification(ctx, notifMicrosoftTeams, name)
}

// --- Project Notification Associations ---

// ProjectNotificationInfo holds the result of looking up a project's notification association.
type ProjectNotificationInfo struct {
	ProjectID int
	Exists    bool
}

// AddNotificationToProject links a notification to a project.
func (c *Client) AddNotificationToProject(ctx context.Context, projectName, notificationType, notificationName string) error {
	_, err := c.Execute(ctx, mutationAddNotificationToProject, map[string]any{
		"input": map[string]any{
			"project":          projectName,
			"notificationType": notificationType,
			"notificationName": notificationName,
		},
	})
	return err
}

// RemoveNotificationFromProject unlinks a notification from a project.
func (c *Client) RemoveNotificationFromProject(ctx context.Context, projectName, notificationType, notificationName string) error {
	_, err := c.Execute(ctx, mutationRemoveNotificationFromProject, map[string]any{
		"input": map[string]any{
			"project":          projectName,
			"notificationType": notificationType,
			"notificationName": notificationName,
		},
	})
	return err
}

// CheckProjectNotificationExists checks if a specific notification is linked to a project.
func (c *Client) CheckProjectNotificationExists(ctx context.Context, projectName, notificationType, notificationName string) (*ProjectNotificationInfo, error) {
	data, err := c.Execute(ctx, queryProjectNotifications, map[string]any{"name": projectName})
	if err != nil {
		return nil, err
	}

	raw, err := extractField(data, "projectByName")
	if err != nil {
		return nil, err
	}

	// Check for null projectByName response (project doesn't exist)
	if strings.TrimSpace(string(raw)) == "null" {
		return nil, &LagoonNotFoundError{ResourceType: "Project", Identifier: projectName}
	}

	var project struct {
		ID            int               `json:"id"`
		Notifications []json.RawMessage `json:"notifications"`
	}
	if err := json.Unmarshal(raw, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project notifications: %w", err)
	}

	typeNameMap := map[string]string{
		"slack":          "NotificationSlack",
		"rocketchat":     "NotificationRocketChat",
		"email":          "NotificationEmail",
		"microsoftteams": "NotificationMicrosoftTeams",
	}

	expectedTypeName := typeNameMap[strings.ToLower(notificationType)]
	if expectedTypeName == "" {
		return nil, fmt.Errorf("unsupported notification type: %s", notificationType)
	}

	for _, rawN := range project.Notifications {
		var n struct {
			TypeName string `json:"__typename"`
			Name     string `json:"name"`
		}
		if err := json.Unmarshal(rawN, &n); err != nil {
			return nil, fmt.Errorf("malformed notification payload: %w", err)
		}
		if n.TypeName == expectedTypeName && n.Name == notificationName {
			return &ProjectNotificationInfo{ProjectID: project.ID, Exists: true}, nil
		}
	}

	return &ProjectNotificationInfo{ProjectID: project.ID, Exists: false}, nil
}
