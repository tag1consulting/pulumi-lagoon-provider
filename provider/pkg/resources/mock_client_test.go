package resources

import (
	"context"

	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// mockLagoonClient is a controllable mock of LagoonClient for tests.
// Set the function fields to customize behavior; nil fields return safe defaults.
type mockLagoonClient struct {
	// Project
	createProjectFn     func(ctx context.Context, input map[string]any) (*client.Project, error)
	getProjectByNameFn  func(ctx context.Context, name string) (*client.Project, error)
	getProjectByIDFn    func(ctx context.Context, id int) (*client.Project, error)
	updateProjectFn     func(ctx context.Context, id int, input map[string]any) (*client.Project, error)
	deleteProjectFn     func(ctx context.Context, name string) error

	// Environment
	addOrUpdateEnvironmentFn func(ctx context.Context, input map[string]any) (*client.Environment, error)
	getEnvironmentByNameFn   func(ctx context.Context, name string, projectID int) (*client.Environment, error)
	deleteEnvironmentFn      func(ctx context.Context, envName string, projectName string) error

	// Variable
	addVariableFn    func(ctx context.Context, name, value string, projectID int, scope string, environmentID *int) (*client.Variable, error)
	getVariableFn    func(ctx context.Context, name string, projectID int, environmentID *int) (*client.Variable, error)
	deleteVariableFn func(ctx context.Context, name string, projectID int, environmentID *int) error

	// DeployTarget
	createDeployTargetFn     func(ctx context.Context, input map[string]any) (*client.DeployTarget, error)
	getDeployTargetByIDFn    func(ctx context.Context, id int) (*client.DeployTarget, error)
	getDeployTargetByNameFn  func(ctx context.Context, name string) (*client.DeployTarget, error)
	updateDeployTargetFn     func(ctx context.Context, id int, input map[string]any) (*client.DeployTarget, error)
	deleteDeployTargetFn     func(ctx context.Context, name string) error

	// DeployTargetConfig
	createDeployTargetConfigFn          func(ctx context.Context, input map[string]any) (*client.DeployTargetConfig, error)
	getDeployTargetConfigsByProjectFn   func(ctx context.Context, projectID int) ([]client.DeployTargetConfig, error)
	getDeployTargetConfigByIDFn         func(ctx context.Context, configID, projectID int) (*client.DeployTargetConfig, error)
	updateDeployTargetConfigFn          func(ctx context.Context, configID int, input map[string]any) (*client.DeployTargetConfig, error)
	deleteDeployTargetConfigFn          func(ctx context.Context, configID, projectID int) error

	// Task
	createTaskDefinitionFn func(ctx context.Context, input map[string]any) (*client.TaskDefinition, error)
	getTaskDefinitionByIDFn func(ctx context.Context, taskID int) (*client.TaskDefinition, error)
	deleteTaskDefinitionFn func(ctx context.Context, taskID int) error

	// Notification — Slack
	createNotificationSlackFn     func(ctx context.Context, name, webhook, channel string) (*client.Notification, error)
	getNotificationSlackByNameFn  func(ctx context.Context, name string) (*client.Notification, error)
	updateNotificationSlackFn     func(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	deleteNotificationSlackFn     func(ctx context.Context, name string) error

	// Notification — RocketChat
	createNotificationRocketChatFn     func(ctx context.Context, name, webhook, channel string) (*client.Notification, error)
	getNotificationRocketChatByNameFn  func(ctx context.Context, name string) (*client.Notification, error)
	updateNotificationRocketChatFn     func(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	deleteNotificationRocketChatFn     func(ctx context.Context, name string) error

	// Notification — Email
	createNotificationEmailFn     func(ctx context.Context, name, emailAddress string) (*client.Notification, error)
	getNotificationEmailByNameFn  func(ctx context.Context, name string) (*client.Notification, error)
	updateNotificationEmailFn     func(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	deleteNotificationEmailFn     func(ctx context.Context, name string) error

	// Notification — Microsoft Teams
	createNotificationMicrosoftTeamsFn     func(ctx context.Context, name, webhook string) (*client.Notification, error)
	getNotificationMicrosoftTeamsByNameFn  func(ctx context.Context, name string) (*client.Notification, error)
	updateNotificationMicrosoftTeamsFn     func(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	deleteNotificationMicrosoftTeamsFn     func(ctx context.Context, name string) error

	// Project notification associations
	addNotificationToProjectFn           func(ctx context.Context, projectName, notificationType, notificationName string) error
	removeNotificationFromProjectFn      func(ctx context.Context, projectName, notificationType, notificationName string) error
	checkProjectNotificationExistsFn     func(ctx context.Context, projectName, notificationType, notificationName string) (*client.ProjectNotificationInfo, error)

	// Group
	createGroupFn     func(ctx context.Context, name string, parentGroupName *string) (*client.Group, error)
	getGroupByNameFn  func(ctx context.Context, name string) (*client.Group, error)
	updateGroupFn     func(ctx context.Context, name string, patch map[string]any) (*client.Group, error)
	deleteGroupFn     func(ctx context.Context, name string) error
}

// --- Project ---

func (m *mockLagoonClient) CreateProject(ctx context.Context, input map[string]any) (*client.Project, error) {
	if m.createProjectFn != nil {
		return m.createProjectFn(ctx, input)
	}
	return &client.Project{ID: 1, Name: "test"}, nil
}

func (m *mockLagoonClient) GetProjectByName(ctx context.Context, name string) (*client.Project, error) {
	if m.getProjectByNameFn != nil {
		return m.getProjectByNameFn(ctx, name)
	}
	return &client.Project{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) GetProjectByID(ctx context.Context, id int) (*client.Project, error) {
	if m.getProjectByIDFn != nil {
		return m.getProjectByIDFn(ctx, id)
	}
	return &client.Project{ID: id, Name: "test"}, nil
}

func (m *mockLagoonClient) UpdateProject(ctx context.Context, id int, input map[string]any) (*client.Project, error) {
	if m.updateProjectFn != nil {
		return m.updateProjectFn(ctx, id, input)
	}
	return &client.Project{ID: id, Name: "test"}, nil
}

func (m *mockLagoonClient) DeleteProject(ctx context.Context, name string) error {
	if m.deleteProjectFn != nil {
		return m.deleteProjectFn(ctx, name)
	}
	return nil
}

// --- Environment ---

func (m *mockLagoonClient) AddOrUpdateEnvironment(ctx context.Context, input map[string]any) (*client.Environment, error) {
	if m.addOrUpdateEnvironmentFn != nil {
		return m.addOrUpdateEnvironmentFn(ctx, input)
	}
	return &client.Environment{ID: 1, Name: "test"}, nil
}

func (m *mockLagoonClient) GetEnvironmentByName(ctx context.Context, name string, projectID int) (*client.Environment, error) {
	if m.getEnvironmentByNameFn != nil {
		return m.getEnvironmentByNameFn(ctx, name, projectID)
	}
	return &client.Environment{ID: 1, Name: name, ProjectID: projectID}, nil
}

func (m *mockLagoonClient) DeleteEnvironment(ctx context.Context, envName string, projectName string) error {
	if m.deleteEnvironmentFn != nil {
		return m.deleteEnvironmentFn(ctx, envName, projectName)
	}
	return nil
}

// --- Variable ---

func (m *mockLagoonClient) AddVariable(ctx context.Context, name, value string, projectID int, scope string, environmentID *int) (*client.Variable, error) {
	if m.addVariableFn != nil {
		return m.addVariableFn(ctx, name, value, projectID, scope, environmentID)
	}
	return &client.Variable{ID: 1, Name: name, Value: value, Scope: scope}, nil
}

func (m *mockLagoonClient) GetVariable(ctx context.Context, name string, projectID int, environmentID *int) (*client.Variable, error) {
	if m.getVariableFn != nil {
		return m.getVariableFn(ctx, name, projectID, environmentID)
	}
	return &client.Variable{ID: 1, Name: name, Value: "val", Scope: "BUILD"}, nil
}

func (m *mockLagoonClient) DeleteVariable(ctx context.Context, name string, projectID int, environmentID *int) error {
	if m.deleteVariableFn != nil {
		return m.deleteVariableFn(ctx, name, projectID, environmentID)
	}
	return nil
}

// --- DeployTarget ---

func (m *mockLagoonClient) CreateDeployTarget(ctx context.Context, input map[string]any) (*client.DeployTarget, error) {
	if m.createDeployTargetFn != nil {
		return m.createDeployTargetFn(ctx, input)
	}
	return &client.DeployTarget{ID: 1, Name: "test"}, nil
}

func (m *mockLagoonClient) GetDeployTargetByID(ctx context.Context, id int) (*client.DeployTarget, error) {
	if m.getDeployTargetByIDFn != nil {
		return m.getDeployTargetByIDFn(ctx, id)
	}
	return &client.DeployTarget{ID: id, Name: "test"}, nil
}

func (m *mockLagoonClient) GetDeployTargetByName(ctx context.Context, name string) (*client.DeployTarget, error) {
	if m.getDeployTargetByNameFn != nil {
		return m.getDeployTargetByNameFn(ctx, name)
	}
	return &client.DeployTarget{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) UpdateDeployTarget(ctx context.Context, id int, input map[string]any) (*client.DeployTarget, error) {
	if m.updateDeployTargetFn != nil {
		return m.updateDeployTargetFn(ctx, id, input)
	}
	return &client.DeployTarget{ID: id, Name: "test"}, nil
}

func (m *mockLagoonClient) DeleteDeployTarget(ctx context.Context, name string) error {
	if m.deleteDeployTargetFn != nil {
		return m.deleteDeployTargetFn(ctx, name)
	}
	return nil
}

// --- DeployTargetConfig ---

func (m *mockLagoonClient) CreateDeployTargetConfig(ctx context.Context, input map[string]any) (*client.DeployTargetConfig, error) {
	if m.createDeployTargetConfigFn != nil {
		return m.createDeployTargetConfigFn(ctx, input)
	}
	return &client.DeployTargetConfig{ID: 1}, nil
}

func (m *mockLagoonClient) GetDeployTargetConfigsByProject(ctx context.Context, projectID int) ([]client.DeployTargetConfig, error) {
	if m.getDeployTargetConfigsByProjectFn != nil {
		return m.getDeployTargetConfigsByProjectFn(ctx, projectID)
	}
	return []client.DeployTargetConfig{}, nil
}

func (m *mockLagoonClient) GetDeployTargetConfigByID(ctx context.Context, configID, projectID int) (*client.DeployTargetConfig, error) {
	if m.getDeployTargetConfigByIDFn != nil {
		return m.getDeployTargetConfigByIDFn(ctx, configID, projectID)
	}
	return &client.DeployTargetConfig{ID: configID, ProjectID: projectID}, nil
}

func (m *mockLagoonClient) UpdateDeployTargetConfig(ctx context.Context, configID int, input map[string]any) (*client.DeployTargetConfig, error) {
	if m.updateDeployTargetConfigFn != nil {
		return m.updateDeployTargetConfigFn(ctx, configID, input)
	}
	return &client.DeployTargetConfig{ID: configID}, nil
}

func (m *mockLagoonClient) DeleteDeployTargetConfig(ctx context.Context, configID, projectID int) error {
	if m.deleteDeployTargetConfigFn != nil {
		return m.deleteDeployTargetConfigFn(ctx, configID, projectID)
	}
	return nil
}

// --- Task ---

func (m *mockLagoonClient) CreateTaskDefinition(ctx context.Context, input map[string]any) (*client.TaskDefinition, error) {
	if m.createTaskDefinitionFn != nil {
		return m.createTaskDefinitionFn(ctx, input)
	}
	return &client.TaskDefinition{ID: 1, Name: "test"}, nil
}

func (m *mockLagoonClient) GetTaskDefinitionByID(ctx context.Context, taskID int) (*client.TaskDefinition, error) {
	if m.getTaskDefinitionByIDFn != nil {
		return m.getTaskDefinitionByIDFn(ctx, taskID)
	}
	return &client.TaskDefinition{ID: taskID, Name: "test"}, nil
}

func (m *mockLagoonClient) DeleteTaskDefinition(ctx context.Context, taskID int) error {
	if m.deleteTaskDefinitionFn != nil {
		return m.deleteTaskDefinitionFn(ctx, taskID)
	}
	return nil
}

// --- Notification Slack ---

func (m *mockLagoonClient) CreateNotificationSlack(ctx context.Context, name, webhook, channel string) (*client.Notification, error) {
	if m.createNotificationSlackFn != nil {
		return m.createNotificationSlackFn(ctx, name, webhook, channel)
	}
	return &client.Notification{ID: 1, Name: name, Webhook: webhook, Channel: channel}, nil
}

func (m *mockLagoonClient) GetNotificationSlackByName(ctx context.Context, name string) (*client.Notification, error) {
	if m.getNotificationSlackByNameFn != nil {
		return m.getNotificationSlackByNameFn(ctx, name)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) UpdateNotificationSlack(ctx context.Context, name string, patch map[string]any) (*client.Notification, error) {
	if m.updateNotificationSlackFn != nil {
		return m.updateNotificationSlackFn(ctx, name, patch)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) DeleteNotificationSlack(ctx context.Context, name string) error {
	if m.deleteNotificationSlackFn != nil {
		return m.deleteNotificationSlackFn(ctx, name)
	}
	return nil
}

// --- Notification RocketChat ---

func (m *mockLagoonClient) CreateNotificationRocketChat(ctx context.Context, name, webhook, channel string) (*client.Notification, error) {
	if m.createNotificationRocketChatFn != nil {
		return m.createNotificationRocketChatFn(ctx, name, webhook, channel)
	}
	return &client.Notification{ID: 1, Name: name, Webhook: webhook, Channel: channel}, nil
}

func (m *mockLagoonClient) GetNotificationRocketChatByName(ctx context.Context, name string) (*client.Notification, error) {
	if m.getNotificationRocketChatByNameFn != nil {
		return m.getNotificationRocketChatByNameFn(ctx, name)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) UpdateNotificationRocketChat(ctx context.Context, name string, patch map[string]any) (*client.Notification, error) {
	if m.updateNotificationRocketChatFn != nil {
		return m.updateNotificationRocketChatFn(ctx, name, patch)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) DeleteNotificationRocketChat(ctx context.Context, name string) error {
	if m.deleteNotificationRocketChatFn != nil {
		return m.deleteNotificationRocketChatFn(ctx, name)
	}
	return nil
}

// --- Notification Email ---

func (m *mockLagoonClient) CreateNotificationEmail(ctx context.Context, name, emailAddress string) (*client.Notification, error) {
	if m.createNotificationEmailFn != nil {
		return m.createNotificationEmailFn(ctx, name, emailAddress)
	}
	return &client.Notification{ID: 1, Name: name, EmailAddress: emailAddress}, nil
}

func (m *mockLagoonClient) GetNotificationEmailByName(ctx context.Context, name string) (*client.Notification, error) {
	if m.getNotificationEmailByNameFn != nil {
		return m.getNotificationEmailByNameFn(ctx, name)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) UpdateNotificationEmail(ctx context.Context, name string, patch map[string]any) (*client.Notification, error) {
	if m.updateNotificationEmailFn != nil {
		return m.updateNotificationEmailFn(ctx, name, patch)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) DeleteNotificationEmail(ctx context.Context, name string) error {
	if m.deleteNotificationEmailFn != nil {
		return m.deleteNotificationEmailFn(ctx, name)
	}
	return nil
}

// --- Notification Microsoft Teams ---

func (m *mockLagoonClient) CreateNotificationMicrosoftTeams(ctx context.Context, name, webhook string) (*client.Notification, error) {
	if m.createNotificationMicrosoftTeamsFn != nil {
		return m.createNotificationMicrosoftTeamsFn(ctx, name, webhook)
	}
	return &client.Notification{ID: 1, Name: name, Webhook: webhook}, nil
}

func (m *mockLagoonClient) GetNotificationMicrosoftTeamsByName(ctx context.Context, name string) (*client.Notification, error) {
	if m.getNotificationMicrosoftTeamsByNameFn != nil {
		return m.getNotificationMicrosoftTeamsByNameFn(ctx, name)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) UpdateNotificationMicrosoftTeams(ctx context.Context, name string, patch map[string]any) (*client.Notification, error) {
	if m.updateNotificationMicrosoftTeamsFn != nil {
		return m.updateNotificationMicrosoftTeamsFn(ctx, name, patch)
	}
	return &client.Notification{ID: 1, Name: name}, nil
}

func (m *mockLagoonClient) DeleteNotificationMicrosoftTeams(ctx context.Context, name string) error {
	if m.deleteNotificationMicrosoftTeamsFn != nil {
		return m.deleteNotificationMicrosoftTeamsFn(ctx, name)
	}
	return nil
}

// --- Project Notification Associations ---

func (m *mockLagoonClient) AddNotificationToProject(ctx context.Context, projectName, notificationType, notificationName string) error {
	if m.addNotificationToProjectFn != nil {
		return m.addNotificationToProjectFn(ctx, projectName, notificationType, notificationName)
	}
	return nil
}

func (m *mockLagoonClient) RemoveNotificationFromProject(ctx context.Context, projectName, notificationType, notificationName string) error {
	if m.removeNotificationFromProjectFn != nil {
		return m.removeNotificationFromProjectFn(ctx, projectName, notificationType, notificationName)
	}
	return nil
}

func (m *mockLagoonClient) CheckProjectNotificationExists(ctx context.Context, projectName, notificationType, notificationName string) (*client.ProjectNotificationInfo, error) {
	if m.checkProjectNotificationExistsFn != nil {
		return m.checkProjectNotificationExistsFn(ctx, projectName, notificationType, notificationName)
	}
	return &client.ProjectNotificationInfo{ProjectID: 1, Exists: true}, nil
}

// --- Group ---

func (m *mockLagoonClient) CreateGroup(ctx context.Context, name string, parentGroupName *string) (*client.Group, error) {
	if m.createGroupFn != nil {
		return m.createGroupFn(ctx, name, parentGroupName)
	}
	return &client.Group{ID: "uuid-1", Name: name}, nil
}

func (m *mockLagoonClient) GetGroupByName(ctx context.Context, name string) (*client.Group, error) {
	if m.getGroupByNameFn != nil {
		return m.getGroupByNameFn(ctx, name)
	}
	return &client.Group{ID: "uuid-1", Name: name}, nil
}

func (m *mockLagoonClient) UpdateGroup(ctx context.Context, name string, patch map[string]any) (*client.Group, error) {
	if m.updateGroupFn != nil {
		return m.updateGroupFn(ctx, name, patch)
	}
	return &client.Group{ID: "uuid-1", Name: name}, nil
}

func (m *mockLagoonClient) DeleteGroup(ctx context.Context, name string) error {
	if m.deleteGroupFn != nil {
		return m.deleteGroupFn(ctx, name)
	}
	return nil
}

// testCtx creates a context with the given mock client injected.
func testCtx(m *mockLagoonClient) context.Context {
	return withTestClient(context.Background(), m)
}
