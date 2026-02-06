package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// --- Slack Notification Tests ---

func TestCreateNotificationSlack(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addNotificationSlack") {
			t.Errorf("expected addNotificationSlack mutation")
		}
		return map[string]any{
			"addNotificationSlack": map[string]any{
				"id":      1,
				"name":    "deploy-alerts",
				"webhook": "https://hooks.slack.com/xxx",
				"channel": "#deployments",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.CreateNotificationSlack(context.Background(), "deploy-alerts", "https://hooks.slack.com/xxx", "#deployments")
	if err != nil {
		t.Fatalf("CreateNotificationSlack failed: %v", err)
	}
	if n.ID != 1 {
		t.Errorf("expected ID=1, got %d", n.ID)
	}
	if n.Channel != "#deployments" {
		t.Errorf("expected channel=#deployments, got %s", n.Channel)
	}
}

func TestGetNotificationSlackByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{
				{"__typename": "NotificationSlack", "id": 1, "name": "slack-1", "webhook": "https://hook1.com", "channel": "#ch1"},
				{"__typename": "NotificationEmail", "id": 2, "name": "email-1", "emailAddress": "test@example.com"},
				{"__typename": "NotificationSlack", "id": 3, "name": "slack-2", "webhook": "https://hook2.com", "channel": "#ch2"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.GetNotificationSlackByName(context.Background(), "slack-2")
	if err != nil {
		t.Fatalf("GetNotificationSlackByName failed: %v", err)
	}
	if n.ID != 3 {
		t.Errorf("expected ID=3, got %d", n.ID)
	}
}

func TestGetNotificationSlackByName_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{
				{"__typename": "NotificationSlack", "id": 1, "name": "other", "webhook": "https://hook.com", "channel": "#ch"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetNotificationSlackByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

func TestUpdateNotificationSlack(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "updateNotificationSlack") {
			t.Errorf("expected updateNotificationSlack mutation")
		}
		return map[string]any{
			"updateNotificationSlack": map[string]any{
				"id":      1,
				"name":    "deploy-alerts",
				"webhook": "https://new-hook.com",
				"channel": "#new-channel",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.UpdateNotificationSlack(context.Background(), "deploy-alerts", map[string]any{"webhook": "https://new-hook.com"})
	if err != nil {
		t.Fatalf("UpdateNotificationSlack failed: %v", err)
	}
	if n.Webhook != "https://new-hook.com" {
		t.Errorf("expected updated webhook")
	}
}

func TestDeleteNotificationSlack(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"deleteNotificationSlack": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteNotificationSlack(context.Background(), "deploy-alerts")
	if err != nil {
		t.Fatalf("DeleteNotificationSlack failed: %v", err)
	}
}

// --- RocketChat Notification Tests ---

func TestCreateNotificationRocketChat(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"addNotificationRocketChat": map[string]any{
				"id":      2,
				"name":    "rc-alerts",
				"webhook": "https://rocket.example.com/hooks/xxx",
				"channel": "#deploys",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.CreateNotificationRocketChat(context.Background(), "rc-alerts", "https://rocket.example.com/hooks/xxx", "#deploys")
	if err != nil {
		t.Fatalf("CreateNotificationRocketChat failed: %v", err)
	}
	if n.ID != 2 {
		t.Errorf("expected ID=2, got %d", n.ID)
	}
}

func TestGetNotificationRocketChatByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{
				{"__typename": "NotificationRocketChat", "id": 5, "name": "rc-1", "webhook": "https://rc.com/hook", "channel": "#ch"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.GetNotificationRocketChatByName(context.Background(), "rc-1")
	if err != nil {
		t.Fatalf("GetNotificationRocketChatByName failed: %v", err)
	}
	if n.Name != "rc-1" {
		t.Errorf("expected name=rc-1, got %s", n.Name)
	}
}

// --- Email Notification Tests ---

func TestCreateNotificationEmail(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"addNotificationEmail": map[string]any{
				"id":           3,
				"name":         "email-alerts",
				"emailAddress": "alerts@example.com",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.CreateNotificationEmail(context.Background(), "email-alerts", "alerts@example.com")
	if err != nil {
		t.Fatalf("CreateNotificationEmail failed: %v", err)
	}
	if n.EmailAddress != "alerts@example.com" {
		t.Errorf("expected emailAddress=alerts@example.com, got %s", n.EmailAddress)
	}
}

func TestGetNotificationEmailByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{
				{"__typename": "NotificationEmail", "id": 10, "name": "email-1", "emailAddress": "test@example.com"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.GetNotificationEmailByName(context.Background(), "email-1")
	if err != nil {
		t.Fatalf("GetNotificationEmailByName failed: %v", err)
	}
	if n.ID != 10 {
		t.Errorf("expected ID=10, got %d", n.ID)
	}
}

func TestGetNotificationEmailByName_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetNotificationEmailByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

// --- Microsoft Teams Notification Tests ---

func TestCreateNotificationMicrosoftTeams(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"addNotificationMicrosoftTeams": map[string]any{
				"id":      4,
				"name":    "teams-alerts",
				"webhook": "https://teams.example.com/webhook/xxx",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.CreateNotificationMicrosoftTeams(context.Background(), "teams-alerts", "https://teams.example.com/webhook/xxx")
	if err != nil {
		t.Fatalf("CreateNotificationMicrosoftTeams failed: %v", err)
	}
	if n.ID != 4 {
		t.Errorf("expected ID=4, got %d", n.ID)
	}
}

func TestGetNotificationMicrosoftTeamsByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{
				{"__typename": "NotificationMicrosoftTeams", "id": 20, "name": "teams-1", "webhook": "https://teams.com/hook"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.GetNotificationMicrosoftTeamsByName(context.Background(), "teams-1")
	if err != nil {
		t.Fatalf("GetNotificationMicrosoftTeamsByName failed: %v", err)
	}
	if n.ID != 20 {
		t.Errorf("expected ID=20, got %d", n.ID)
	}
}

func TestGetNotificationMicrosoftTeamsByName_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allNotifications": []map[string]any{
				{"__typename": "NotificationSlack", "id": 1, "name": "teams-1", "webhook": "https://hook.com"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetNotificationMicrosoftTeamsByName(context.Background(), "teams-1")
	if err == nil {
		t.Fatal("expected error - wrong type name")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %T", err)
	}
}

// --- Project Notification Association Tests ---

func TestAddNotificationToProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addNotificationToProject") {
			t.Errorf("expected addNotificationToProject mutation")
		}
		return map[string]any{
			"addNotificationToProject": map[string]any{
				"id":   1,
				"name": "my-project",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.AddNotificationToProject(context.Background(), "my-project", "SLACK", "deploy-alerts")
	if err != nil {
		t.Fatalf("AddNotificationToProject failed: %v", err)
	}
}

func TestRemoveNotificationFromProject(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "removeNotificationFromProject") {
			t.Errorf("expected removeNotificationFromProject mutation")
		}
		return map[string]any{
			"removeNotificationFromProject": map[string]any{
				"id":   1,
				"name": "my-project",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.RemoveNotificationFromProject(context.Background(), "my-project", "SLACK", "deploy-alerts")
	if err != nil {
		t.Fatalf("RemoveNotificationFromProject failed: %v", err)
	}
}

func TestCheckProjectNotificationExists_Found(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"projectByName": map[string]any{
				"id":   42,
				"name": "my-project",
				"notifications": []map[string]any{
					{"__typename": "NotificationSlack", "id": 1, "name": "deploy-alerts", "webhook": "https://hook.com", "channel": "#ch"},
					{"__typename": "NotificationEmail", "id": 2, "name": "email-1", "emailAddress": "test@example.com"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	info, err := c.CheckProjectNotificationExists(context.Background(), "my-project", "slack", "deploy-alerts")
	if err != nil {
		t.Fatalf("CheckProjectNotificationExists failed: %v", err)
	}
	if !info.Exists {
		t.Error("expected notification to exist")
	}
	if info.ProjectID != 42 {
		t.Errorf("expected ProjectID=42, got %d", info.ProjectID)
	}
}

func TestCheckProjectNotificationExists_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"projectByName": map[string]any{
				"id":   42,
				"name": "my-project",
				"notifications": []map[string]any{
					{"__typename": "NotificationEmail", "id": 2, "name": "email-1", "emailAddress": "test@example.com"},
				},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	info, err := c.CheckProjectNotificationExists(context.Background(), "my-project", "slack", "deploy-alerts")
	if err != nil {
		t.Fatalf("CheckProjectNotificationExists failed: %v", err)
	}
	if info.Exists {
		t.Error("expected notification to NOT exist")
	}
	if info.ProjectID != 42 {
		t.Errorf("expected ProjectID=42 even when not found, got %d", info.ProjectID)
	}
}

func TestDeleteNotificationRocketChat(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"deleteNotificationRocketChat": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteNotificationRocketChat(context.Background(), "rc-alerts")
	if err != nil {
		t.Fatalf("DeleteNotificationRocketChat failed: %v", err)
	}
}

func TestDeleteNotificationEmail(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"deleteNotificationEmail": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteNotificationEmail(context.Background(), "email-alerts")
	if err != nil {
		t.Fatalf("DeleteNotificationEmail failed: %v", err)
	}
}

func TestDeleteNotificationMicrosoftTeams(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{"deleteNotificationMicrosoftTeams": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	err := c.DeleteNotificationMicrosoftTeams(context.Background(), "teams-alerts")
	if err != nil {
		t.Fatalf("DeleteNotificationMicrosoftTeams failed: %v", err)
	}
}

func TestUpdateNotificationRocketChat(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"updateNotificationRocketChat": map[string]any{
				"id": 5, "name": "rc-1", "webhook": "https://new.com", "channel": "#new",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.UpdateNotificationRocketChat(context.Background(), "rc-1", map[string]any{"webhook": "https://new.com"})
	if err != nil {
		t.Fatalf("UpdateNotificationRocketChat failed: %v", err)
	}
	if n.Webhook != "https://new.com" {
		t.Errorf("expected updated webhook")
	}
}

func TestUpdateNotificationEmail(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"updateNotificationEmail": map[string]any{
				"id": 10, "name": "email-1", "emailAddress": "new@example.com",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.UpdateNotificationEmail(context.Background(), "email-1", map[string]any{"emailAddress": "new@example.com"})
	if err != nil {
		t.Fatalf("UpdateNotificationEmail failed: %v", err)
	}
	if n.EmailAddress != "new@example.com" {
		t.Errorf("expected updated emailAddress")
	}
}

func TestUpdateNotificationMicrosoftTeams(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"updateNotificationMicrosoftTeams": map[string]any{
				"id": 20, "name": "teams-1", "webhook": "https://new-teams.com",
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	n, err := c.UpdateNotificationMicrosoftTeams(context.Background(), "teams-1", map[string]any{"webhook": "https://new-teams.com"})
	if err != nil {
		t.Fatalf("UpdateNotificationMicrosoftTeams failed: %v", err)
	}
	if n.Webhook != "https://new-teams.com" {
		t.Errorf("expected updated webhook")
	}
}
