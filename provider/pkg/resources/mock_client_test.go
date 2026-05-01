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

	// Route
	addRouteToProjectFn              func(ctx context.Context, input map[string]any) (*client.Route, error)
	getRouteByDomainFn               func(ctx context.Context, projectName, domain string) (*client.Route, error)
	updateRouteOnProjectFn           func(ctx context.Context, routeID int, patch map[string]any) (*client.Route, error)
	deleteRouteFn                    func(ctx context.Context, routeID int) error
	addRouteAlternativeDomainsFn     func(ctx context.Context, routeID int, domains []string) (*client.Route, error)
	removeRouteAlternativeDomainFn   func(ctx context.Context, routeID int, domain string) (*client.Route, error)
	addRouteAnnotationsFn            func(ctx context.Context, routeID int, annotations []map[string]string) (*client.Route, error)
	removeRouteAnnotationFn          func(ctx context.Context, routeID int, key string) (*client.Route, error)
	addPathRoutesToRouteFn           func(ctx context.Context, routeID int, pathRoutes []map[string]string) (*client.Route, error)
	removePathRouteFromRouteFn       func(ctx context.Context, routeID int, toService, path string) (*client.Route, error)
	addOrUpdateRouteOnEnvironmentFn  func(ctx context.Context, input map[string]any) (*client.Route, error)
	removeRouteFromEnvironmentFn     func(ctx context.Context, domain, project, environment string) error

	// Autogenerated route config — project
	updateAutoRouteConfigProjectFn func(ctx context.Context, projectName string, patch map[string]any) (*client.AutogeneratedRouteConfig, error)
	removeAutoRouteConfigProjectFn func(ctx context.Context, projectName string) error
	getAutoRouteConfigProjectFn    func(ctx context.Context, projectName string) (*client.AutogeneratedRouteConfig, error)

	// Autogenerated route config — environment
	updateAutoRouteConfigEnvFn func(ctx context.Context, projectName, environmentName string, patch map[string]any) (*client.AutogeneratedRouteConfig, error)
	removeAutoRouteConfigEnvFn func(ctx context.Context, projectName, environmentName string) error
	getAutoRouteConfigEnvFn    func(ctx context.Context, projectName, environmentName string) (*client.AutogeneratedRouteConfig, error)

	// User
	createUserFn     func(ctx context.Context, email string, firstName, lastName, comment *string) (*client.User, error)
	getUserByEmailFn func(ctx context.Context, email string) (*client.User, error)
	updateUserFn     func(ctx context.Context, email string, patch map[string]any) (*client.User, error)
	deleteUserFn     func(ctx context.Context, email string) error

	// User group assignment
	addUserToGroupFn      func(ctx context.Context, email, groupName, role string) error
	removeUserFromGroupFn func(ctx context.Context, email, groupName string) error
	getUserGroupRolesFn   func(ctx context.Context, email string) ([]client.UserGroupRole, error)

	// User platform role
	addPlatformRoleToUserFn      func(ctx context.Context, email, role string) error
	removePlatformRoleFromUserFn func(ctx context.Context, email, role string) error
	getUserPlatformRolesFn       func(ctx context.Context, email string) ([]string, error)
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

// --- Route ---

func (m *mockLagoonClient) AddRouteToProject(ctx context.Context, input map[string]any) (*client.Route, error) {
	if m.addRouteToProjectFn != nil {
		return m.addRouteToProjectFn(ctx, input)
	}
	return &client.Route{ID: 1, Domain: "example.com", Source: "API"}, nil
}

func (m *mockLagoonClient) GetRouteByDomain(ctx context.Context, projectName, domain string) (*client.Route, error) {
	if m.getRouteByDomainFn != nil {
		return m.getRouteByDomainFn(ctx, projectName, domain)
	}
	return &client.Route{ID: 1, Domain: domain, ProjectName: projectName, Source: "API"}, nil
}

func (m *mockLagoonClient) UpdateRouteOnProject(ctx context.Context, routeID int, patch map[string]any) (*client.Route, error) {
	if m.updateRouteOnProjectFn != nil {
		return m.updateRouteOnProjectFn(ctx, routeID, patch)
	}
	return &client.Route{ID: routeID, Domain: "example.com", Source: "API"}, nil
}

func (m *mockLagoonClient) DeleteRoute(ctx context.Context, routeID int) error {
	if m.deleteRouteFn != nil {
		return m.deleteRouteFn(ctx, routeID)
	}
	return nil
}

func (m *mockLagoonClient) AddRouteAlternativeDomains(ctx context.Context, routeID int, domains []string) (*client.Route, error) {
	if m.addRouteAlternativeDomainsFn != nil {
		return m.addRouteAlternativeDomainsFn(ctx, routeID, domains)
	}
	return &client.Route{ID: routeID}, nil
}

func (m *mockLagoonClient) RemoveRouteAlternativeDomain(ctx context.Context, routeID int, domain string) (*client.Route, error) {
	if m.removeRouteAlternativeDomainFn != nil {
		return m.removeRouteAlternativeDomainFn(ctx, routeID, domain)
	}
	return &client.Route{ID: routeID}, nil
}

func (m *mockLagoonClient) AddRouteAnnotations(ctx context.Context, routeID int, annotations []map[string]string) (*client.Route, error) {
	if m.addRouteAnnotationsFn != nil {
		return m.addRouteAnnotationsFn(ctx, routeID, annotations)
	}
	return &client.Route{ID: routeID}, nil
}

func (m *mockLagoonClient) RemoveRouteAnnotation(ctx context.Context, routeID int, key string) (*client.Route, error) {
	if m.removeRouteAnnotationFn != nil {
		return m.removeRouteAnnotationFn(ctx, routeID, key)
	}
	return &client.Route{ID: routeID}, nil
}

func (m *mockLagoonClient) AddPathRoutesToRoute(ctx context.Context, routeID int, pathRoutes []map[string]string) (*client.Route, error) {
	if m.addPathRoutesToRouteFn != nil {
		return m.addPathRoutesToRouteFn(ctx, routeID, pathRoutes)
	}
	return &client.Route{ID: routeID}, nil
}

func (m *mockLagoonClient) RemovePathRouteFromRoute(ctx context.Context, routeID int, toService, path string) (*client.Route, error) {
	if m.removePathRouteFromRouteFn != nil {
		return m.removePathRouteFromRouteFn(ctx, routeID, toService, path)
	}
	return &client.Route{ID: routeID}, nil
}

func (m *mockLagoonClient) AddOrUpdateRouteOnEnvironment(ctx context.Context, input map[string]any) (*client.Route, error) {
	if m.addOrUpdateRouteOnEnvironmentFn != nil {
		return m.addOrUpdateRouteOnEnvironmentFn(ctx, input)
	}
	return &client.Route{ID: 1}, nil
}

func (m *mockLagoonClient) RemoveRouteFromEnvironment(ctx context.Context, domain, project, environment string) error {
	if m.removeRouteFromEnvironmentFn != nil {
		return m.removeRouteFromEnvironmentFn(ctx, domain, project, environment)
	}
	return nil
}

// --- Autogenerated Route Config — Project ---

func (m *mockLagoonClient) UpdateAutogeneratedRouteConfigOnProject(ctx context.Context, projectName string, patch map[string]any) (*client.AutogeneratedRouteConfig, error) {
	if m.updateAutoRouteConfigProjectFn != nil {
		return m.updateAutoRouteConfigProjectFn(ctx, projectName, patch)
	}
	return &client.AutogeneratedRouteConfig{}, nil
}

func (m *mockLagoonClient) RemoveAutogeneratedRouteConfigFromProject(ctx context.Context, projectName string) error {
	if m.removeAutoRouteConfigProjectFn != nil {
		return m.removeAutoRouteConfigProjectFn(ctx, projectName)
	}
	return nil
}

func (m *mockLagoonClient) GetAutogeneratedRouteConfigForProject(ctx context.Context, projectName string) (*client.AutogeneratedRouteConfig, error) {
	if m.getAutoRouteConfigProjectFn != nil {
		return m.getAutoRouteConfigProjectFn(ctx, projectName)
	}
	return nil, nil
}

// --- Autogenerated Route Config — Environment ---

func (m *mockLagoonClient) UpdateAutogeneratedRouteConfigOnEnvironment(ctx context.Context, projectName, environmentName string, patch map[string]any) (*client.AutogeneratedRouteConfig, error) {
	if m.updateAutoRouteConfigEnvFn != nil {
		return m.updateAutoRouteConfigEnvFn(ctx, projectName, environmentName, patch)
	}
	return &client.AutogeneratedRouteConfig{}, nil
}

func (m *mockLagoonClient) RemoveAutogeneratedRouteConfigFromEnvironment(ctx context.Context, projectName, environmentName string) error {
	if m.removeAutoRouteConfigEnvFn != nil {
		return m.removeAutoRouteConfigEnvFn(ctx, projectName, environmentName)
	}
	return nil
}

func (m *mockLagoonClient) GetAutogeneratedRouteConfigForEnvironment(ctx context.Context, projectName, environmentName string) (*client.AutogeneratedRouteConfig, error) {
	if m.getAutoRouteConfigEnvFn != nil {
		return m.getAutoRouteConfigEnvFn(ctx, projectName, environmentName)
	}
	return nil, nil
}

// --- User ---

func (m *mockLagoonClient) CreateUser(ctx context.Context, email string, firstName, lastName, comment *string) (*client.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, email, firstName, lastName, comment)
	}
	return &client.User{ID: "uuid-user-1", Email: email}, nil
}

func (m *mockLagoonClient) GetUserByEmail(ctx context.Context, email string) (*client.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return &client.User{ID: "uuid-user-1", Email: email}, nil
}

func (m *mockLagoonClient) UpdateUser(ctx context.Context, email string, patch map[string]any) (*client.User, error) {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, email, patch)
	}
	return &client.User{ID: "uuid-user-1", Email: email}, nil
}

func (m *mockLagoonClient) DeleteUser(ctx context.Context, email string) error {
	if m.deleteUserFn != nil {
		return m.deleteUserFn(ctx, email)
	}
	return nil
}

// --- User Group Assignment ---

func (m *mockLagoonClient) AddUserToGroup(ctx context.Context, email, groupName, role string) error {
	if m.addUserToGroupFn != nil {
		return m.addUserToGroupFn(ctx, email, groupName, role)
	}
	return nil
}

func (m *mockLagoonClient) RemoveUserFromGroup(ctx context.Context, email, groupName string) error {
	if m.removeUserFromGroupFn != nil {
		return m.removeUserFromGroupFn(ctx, email, groupName)
	}
	return nil
}

func (m *mockLagoonClient) GetUserGroupRoles(ctx context.Context, email string) ([]client.UserGroupRole, error) {
	if m.getUserGroupRolesFn != nil {
		return m.getUserGroupRolesFn(ctx, email)
	}
	return []client.UserGroupRole{}, nil
}

// --- User Platform Role ---

func (m *mockLagoonClient) AddPlatformRoleToUser(ctx context.Context, email, role string) error {
	if m.addPlatformRoleToUserFn != nil {
		return m.addPlatformRoleToUserFn(ctx, email, role)
	}
	return nil
}

func (m *mockLagoonClient) RemovePlatformRoleFromUser(ctx context.Context, email, role string) error {
	if m.removePlatformRoleFromUserFn != nil {
		return m.removePlatformRoleFromUserFn(ctx, email, role)
	}
	return nil
}

func (m *mockLagoonClient) GetUserPlatformRoles(ctx context.Context, email string) ([]string, error) {
	if m.getUserPlatformRolesFn != nil {
		return m.getUserPlatformRolesFn(ctx, email)
	}
	return []string{}, nil
}

// testCtx creates a context with the given mock client injected.
func testCtx(m *mockLagoonClient) context.Context {
	return withTestClient(context.Background(), m)
}
