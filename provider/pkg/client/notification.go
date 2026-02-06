package client

import (
	"context"
	"encoding/json"
	"fmt"
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

// getAllNotifications retrieves all notifications of all types.
func (c *Client) getAllNotifications(ctx context.Context) ([]Notification, error) {
	data, err := c.Execute(ctx, queryAllNotifications, nil)
	if err != nil {
		return nil, err
	}

	notifications, err := unmarshalField[[]Notification](data, "allNotifications")
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// --- Slack ---

// CreateNotificationSlack creates a Slack notification.
func (c *Client) CreateNotificationSlack(ctx context.Context, name, webhook, channel string) (*Notification, error) {
	data, err := c.Execute(ctx, mutationAddNotificationSlack, map[string]any{
		"input": map[string]any{"name": name, "webhook": webhook, "channel": channel},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "addNotificationSlack")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// GetNotificationSlackByName retrieves a Slack notification by name.
func (c *Client) GetNotificationSlackByName(ctx context.Context, name string) (*Notification, error) {
	all, err := c.getAllNotifications(ctx)
	if err != nil {
		return nil, err
	}
	for _, n := range all {
		if n.TypeName == "NotificationSlack" && n.Name == name {
			return &n, nil
		}
	}
	return nil, &LagoonNotFoundError{ResourceType: "NotificationSlack", Identifier: name}
}

// UpdateNotificationSlack updates a Slack notification.
func (c *Client) UpdateNotificationSlack(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	data, err := c.Execute(ctx, mutationUpdateNotificationSlack, map[string]any{
		"input": map[string]any{"name": name, "patch": patch},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "updateNotificationSlack")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteNotificationSlack deletes a Slack notification.
func (c *Client) DeleteNotificationSlack(ctx context.Context, name string) error {
	_, err := c.Execute(ctx, mutationDeleteNotificationSlack, map[string]any{
		"input": map[string]any{"name": name},
	})
	return err
}

// --- RocketChat ---

// CreateNotificationRocketChat creates a RocketChat notification.
func (c *Client) CreateNotificationRocketChat(ctx context.Context, name, webhook, channel string) (*Notification, error) {
	data, err := c.Execute(ctx, mutationAddNotificationRocketChat, map[string]any{
		"input": map[string]any{"name": name, "webhook": webhook, "channel": channel},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "addNotificationRocketChat")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// GetNotificationRocketChatByName retrieves a RocketChat notification by name.
func (c *Client) GetNotificationRocketChatByName(ctx context.Context, name string) (*Notification, error) {
	all, err := c.getAllNotifications(ctx)
	if err != nil {
		return nil, err
	}
	for _, n := range all {
		if n.TypeName == "NotificationRocketChat" && n.Name == name {
			return &n, nil
		}
	}
	return nil, &LagoonNotFoundError{ResourceType: "NotificationRocketChat", Identifier: name}
}

// UpdateNotificationRocketChat updates a RocketChat notification.
func (c *Client) UpdateNotificationRocketChat(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	data, err := c.Execute(ctx, mutationUpdateNotificationRocketChat, map[string]any{
		"input": map[string]any{"name": name, "patch": patch},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "updateNotificationRocketChat")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteNotificationRocketChat deletes a RocketChat notification.
func (c *Client) DeleteNotificationRocketChat(ctx context.Context, name string) error {
	_, err := c.Execute(ctx, mutationDeleteNotificationRocketChat, map[string]any{
		"input": map[string]any{"name": name},
	})
	return err
}

// --- Email ---

// CreateNotificationEmail creates an Email notification.
func (c *Client) CreateNotificationEmail(ctx context.Context, name, emailAddress string) (*Notification, error) {
	data, err := c.Execute(ctx, mutationAddNotificationEmail, map[string]any{
		"input": map[string]any{"name": name, "emailAddress": emailAddress},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "addNotificationEmail")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// GetNotificationEmailByName retrieves an Email notification by name.
func (c *Client) GetNotificationEmailByName(ctx context.Context, name string) (*Notification, error) {
	all, err := c.getAllNotifications(ctx)
	if err != nil {
		return nil, err
	}
	for _, n := range all {
		if n.TypeName == "NotificationEmail" && n.Name == name {
			return &n, nil
		}
	}
	return nil, &LagoonNotFoundError{ResourceType: "NotificationEmail", Identifier: name}
}

// UpdateNotificationEmail updates an Email notification.
func (c *Client) UpdateNotificationEmail(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	data, err := c.Execute(ctx, mutationUpdateNotificationEmail, map[string]any{
		"input": map[string]any{"name": name, "patch": patch},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "updateNotificationEmail")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteNotificationEmail deletes an Email notification.
func (c *Client) DeleteNotificationEmail(ctx context.Context, name string) error {
	_, err := c.Execute(ctx, mutationDeleteNotificationEmail, map[string]any{
		"input": map[string]any{"name": name},
	})
	return err
}

// --- Microsoft Teams ---

// CreateNotificationMicrosoftTeams creates a Microsoft Teams notification.
func (c *Client) CreateNotificationMicrosoftTeams(ctx context.Context, name, webhook string) (*Notification, error) {
	data, err := c.Execute(ctx, mutationAddNotificationMicrosoftTeams, map[string]any{
		"input": map[string]any{"name": name, "webhook": webhook},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "addNotificationMicrosoftTeams")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// GetNotificationMicrosoftTeamsByName retrieves a Microsoft Teams notification by name.
func (c *Client) GetNotificationMicrosoftTeamsByName(ctx context.Context, name string) (*Notification, error) {
	all, err := c.getAllNotifications(ctx)
	if err != nil {
		return nil, err
	}
	for _, n := range all {
		if n.TypeName == "NotificationMicrosoftTeams" && n.Name == name {
			return &n, nil
		}
	}
	return nil, &LagoonNotFoundError{ResourceType: "NotificationMicrosoftTeams", Identifier: name}
}

// UpdateNotificationMicrosoftTeams updates a Microsoft Teams notification.
func (c *Client) UpdateNotificationMicrosoftTeams(ctx context.Context, name string, patch map[string]any) (*Notification, error) {
	data, err := c.Execute(ctx, mutationUpdateNotificationMicrosoftTeams, map[string]any{
		"input": map[string]any{"name": name, "patch": patch},
	})
	if err != nil {
		return nil, err
	}

	n, err := unmarshalField[Notification](data, "updateNotificationMicrosoftTeams")
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// DeleteNotificationMicrosoftTeams deletes a Microsoft Teams notification.
func (c *Client) DeleteNotificationMicrosoftTeams(ctx context.Context, name string) error {
	_, err := c.Execute(ctx, mutationDeleteNotificationMicrosoftTeams, map[string]any{
		"input": map[string]any{"name": name},
	})
	return err
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

	var project struct {
		ID            int              `json:"id"`
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

	expectedTypeName := typeNameMap[notificationType]

	for _, rawN := range project.Notifications {
		var n struct {
			TypeName string `json:"__typename"`
			Name     string `json:"name"`
		}
		if err := json.Unmarshal(rawN, &n); err != nil {
			continue
		}
		if n.TypeName == expectedTypeName && n.Name == notificationName {
			return &ProjectNotificationInfo{ProjectID: project.ID, Exists: true}, nil
		}
	}

	return &ProjectNotificationInfo{ProjectID: project.ID, Exists: false}, nil
}
