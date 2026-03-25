package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCreateGroup(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "addGroup") {
			t.Errorf("expected addGroup mutation")
		}
		input, ok := variables["input"].(map[string]any)
		if !ok {
			t.Fatalf("expected variables[\"input\"] to be a map, got %T", variables["input"])
		}
		if name, _ := input["name"].(string); name != "my-group" {
			t.Errorf("expected input name=my-group, got %v", input["name"])
		}
		if _, hasParent := input["parentGroup"]; hasParent {
			t.Error("expected no parentGroup for root group")
		}
		return map[string]any{
			"addGroup": map[string]any{"id": 1, "name": "my-group"},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	g, err := c.CreateGroup(context.Background(), "my-group", nil)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	if g.ID != 1 {
		t.Errorf("expected ID=1, got %d", g.ID)
	}
	if g.Name != "my-group" {
		t.Errorf("expected Name=my-group, got %s", g.Name)
	}
}

func TestCreateGroup_WithParent(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		input, ok := variables["input"].(map[string]any)
		if !ok {
			t.Fatalf("expected variables[\"input\"] to be a map, got %T", variables["input"])
		}
		parentGroup, ok := input["parentGroup"].(map[string]any)
		if !ok {
			t.Fatalf("expected parentGroup to be a map, got %T", input["parentGroup"])
		}
		if name, _ := parentGroup["name"].(string); name != "parent-group" {
			t.Errorf("expected parentGroup name=parent-group, got %v", parentGroup["name"])
		}
		return map[string]any{
			"addGroup": map[string]any{"id": 2, "name": "child-group"},
		}, nil
	})
	defer server.Close()

	parent := "parent-group"
	c := NewClient(server.URL, "token")
	g, err := c.CreateGroup(context.Background(), "child-group", &parent)
	if err != nil {
		t.Fatalf("CreateGroup with parent failed: %v", err)
	}
	if g.ID != 2 {
		t.Errorf("expected ID=2, got %d", g.ID)
	}
}

func TestGetGroupByName(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "allGroups") {
			t.Errorf("expected allGroups query")
		}
		return map[string]any{
			"allGroups": []map[string]any{
				{"id": 1, "name": "group-a"},
				{"id": 2, "name": "group-b"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	g, err := c.GetGroupByName(context.Background(), "group-b")
	if err != nil {
		t.Fatalf("GetGroupByName failed: %v", err)
	}
	if g.ID != 2 {
		t.Errorf("expected ID=2, got %d", g.ID)
	}
}

func TestGetGroupByName_NotFound(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		return map[string]any{
			"allGroups": []map[string]any{
				{"id": 1, "name": "other-group"},
			},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	_, err := c.GetGroupByName(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing group")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateGroup(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "updateGroup") {
			t.Errorf("expected updateGroup mutation")
		}
		input, ok := variables["input"].(map[string]any)
		if !ok {
			t.Fatalf("expected variables[\"input\"] to be a map, got %T", variables["input"])
		}
		group, _ := input["group"].(map[string]any)
		if name, _ := group["name"].(string); name != "my-group" {
			t.Errorf("expected group name=my-group, got %v", group["name"])
		}
		return map[string]any{
			"updateGroup": map[string]any{"id": 1, "name": "renamed-group"},
		}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	g, err := c.UpdateGroup(context.Background(), "my-group", map[string]any{"name": "renamed-group"})
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}
	if g.Name != "renamed-group" {
		t.Errorf("expected Name=renamed-group, got %s", g.Name)
	}
}

func TestDeleteGroup(t *testing.T) {
	server := mockGraphQLServer(t, func(query string, variables map[string]any) (any, error) {
		if !strings.Contains(query, "deleteGroup") {
			t.Errorf("expected deleteGroup mutation")
		}
		input, ok := variables["input"].(map[string]any)
		if !ok {
			t.Fatalf("expected variables[\"input\"] to be a map, got %T", variables["input"])
		}
		group, _ := input["group"].(map[string]any)
		if name, _ := group["name"].(string); name != "my-group" {
			t.Errorf("expected group name=my-group, got %v", group["name"])
		}
		return map[string]any{"deleteGroup": "success"}, nil
	})
	defer server.Close()

	c := NewClient(server.URL, "token")
	if err := c.DeleteGroup(context.Background(), "my-group"); err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}
