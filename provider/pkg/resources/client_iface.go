package resources

import (
	"context"

	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/client"
)

// LagoonClient defines the API operations used by resources.
// *client.Client satisfies this interface.
type LagoonClient interface {
	// Project methods
	CreateProject(ctx context.Context, input map[string]any) (*client.Project, error)
	GetProjectByName(ctx context.Context, name string) (*client.Project, error)
	GetProjectByID(ctx context.Context, id int) (*client.Project, error)
	UpdateProject(ctx context.Context, id int, input map[string]any) (*client.Project, error)
	DeleteProject(ctx context.Context, name string) error

	// Environment methods
	AddOrUpdateEnvironment(ctx context.Context, input map[string]any) (*client.Environment, error)
	GetEnvironmentByName(ctx context.Context, name string, projectID int) (*client.Environment, error)
	DeleteEnvironment(ctx context.Context, envName string, projectName string) error

	// Variable methods
	AddVariable(ctx context.Context, name, value string, projectID int, scope string, environmentID *int) (*client.Variable, error)
	GetVariable(ctx context.Context, name string, projectID int, environmentID *int) (*client.Variable, error)
	DeleteVariable(ctx context.Context, name string, projectID int, environmentID *int) error

	// Deploy target methods
	CreateDeployTarget(ctx context.Context, input map[string]any) (*client.DeployTarget, error)
	GetDeployTargetByID(ctx context.Context, id int) (*client.DeployTarget, error)
	GetDeployTargetByName(ctx context.Context, name string) (*client.DeployTarget, error)
	UpdateDeployTarget(ctx context.Context, id int, input map[string]any) (*client.DeployTarget, error)
	DeleteDeployTarget(ctx context.Context, name string) error

	// Deploy target config methods
	CreateDeployTargetConfig(ctx context.Context, input map[string]any) (*client.DeployTargetConfig, error)
	GetDeployTargetConfigsByProject(ctx context.Context, projectID int) ([]client.DeployTargetConfig, error)
	GetDeployTargetConfigByID(ctx context.Context, configID, projectID int) (*client.DeployTargetConfig, error)
	UpdateDeployTargetConfig(ctx context.Context, configID int, input map[string]any) (*client.DeployTargetConfig, error)
	DeleteDeployTargetConfig(ctx context.Context, configID, projectID int) error

	// Task methods
	CreateTaskDefinition(ctx context.Context, input map[string]any) (*client.TaskDefinition, error)
	GetTaskDefinitionByID(ctx context.Context, taskID int) (*client.TaskDefinition, error)
	DeleteTaskDefinition(ctx context.Context, taskID int) error

	// Notification methods — Slack
	CreateNotificationSlack(ctx context.Context, name, webhook, channel string) (*client.Notification, error)
	GetNotificationSlackByName(ctx context.Context, name string) (*client.Notification, error)
	UpdateNotificationSlack(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	DeleteNotificationSlack(ctx context.Context, name string) error

	// Notification methods — RocketChat
	CreateNotificationRocketChat(ctx context.Context, name, webhook, channel string) (*client.Notification, error)
	GetNotificationRocketChatByName(ctx context.Context, name string) (*client.Notification, error)
	UpdateNotificationRocketChat(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	DeleteNotificationRocketChat(ctx context.Context, name string) error

	// Notification methods — Email
	CreateNotificationEmail(ctx context.Context, name, emailAddress string) (*client.Notification, error)
	GetNotificationEmailByName(ctx context.Context, name string) (*client.Notification, error)
	UpdateNotificationEmail(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	DeleteNotificationEmail(ctx context.Context, name string) error

	// Notification methods — Microsoft Teams
	CreateNotificationMicrosoftTeams(ctx context.Context, name, webhook string) (*client.Notification, error)
	GetNotificationMicrosoftTeamsByName(ctx context.Context, name string) (*client.Notification, error)
	UpdateNotificationMicrosoftTeams(ctx context.Context, name string, patch map[string]any) (*client.Notification, error)
	DeleteNotificationMicrosoftTeams(ctx context.Context, name string) error

	// Project notification association methods
	AddNotificationToProject(ctx context.Context, projectName, notificationType, notificationName string) error
	RemoveNotificationFromProject(ctx context.Context, projectName, notificationType, notificationName string) error
	CheckProjectNotificationExists(ctx context.Context, projectName, notificationType, notificationName string) (*client.ProjectNotificationInfo, error)

	// Group methods
	CreateGroup(ctx context.Context, name string, parentGroupName *string) (*client.Group, error)
	GetGroupByName(ctx context.Context, name string) (*client.Group, error)
	UpdateGroup(ctx context.Context, name string, patch map[string]any) (*client.Group, error)
	DeleteGroup(ctx context.Context, name string) error

	// Route methods
	AddRouteToProject(ctx context.Context, input map[string]any) (*client.Route, error)
	GetRouteByDomain(ctx context.Context, projectName, domain string) (*client.Route, error)
	UpdateRouteOnProject(ctx context.Context, routeID int, patch map[string]any) (*client.Route, error)
	DeleteRoute(ctx context.Context, routeID int) error

	// Route sub-object methods
	AddRouteAlternativeDomains(ctx context.Context, routeID int, domains []string) (*client.Route, error)
	RemoveRouteAlternativeDomain(ctx context.Context, routeID int, domain string) (*client.Route, error)
	AddRouteAnnotations(ctx context.Context, routeID int, annotations []map[string]string) (*client.Route, error)
	RemoveRouteAnnotation(ctx context.Context, routeID int, key string) (*client.Route, error)
	AddPathRoutesToRoute(ctx context.Context, routeID int, pathRoutes []map[string]string) (*client.Route, error)
	RemovePathRouteFromRoute(ctx context.Context, routeID int, toService, path string) (*client.Route, error)

	// Route environment attachment
	AddOrUpdateRouteOnEnvironment(ctx context.Context, input map[string]any) (*client.Route, error)
	RemoveRouteFromEnvironment(ctx context.Context, domain, project, environment string) error

	// Autogenerated route config — project level
	UpdateAutogeneratedRouteConfigOnProject(ctx context.Context, projectName string, patch map[string]any) (*client.AutogeneratedRouteConfig, error)
	RemoveAutogeneratedRouteConfigFromProject(ctx context.Context, projectName string) error
	GetAutogeneratedRouteConfigForProject(ctx context.Context, projectName string) (*client.AutogeneratedRouteConfig, error)

	// Autogenerated route config — environment level
	UpdateAutogeneratedRouteConfigOnEnvironment(ctx context.Context, projectName, environmentName string, patch map[string]any) (*client.AutogeneratedRouteConfig, error)
	RemoveAutogeneratedRouteConfigFromEnvironment(ctx context.Context, projectName, environmentName string) error
	GetAutogeneratedRouteConfigForEnvironment(ctx context.Context, projectName, environmentName string) (*client.AutogeneratedRouteConfig, error)

	// User methods
	CreateUser(ctx context.Context, email string, firstName, lastName, comment *string) (*client.User, error)
	GetUserByEmail(ctx context.Context, email string) (*client.User, error)
	UpdateUser(ctx context.Context, email string, patch map[string]any) (*client.User, error)
	DeleteUser(ctx context.Context, email string) error

	// User group assignment methods
	AddUserToGroup(ctx context.Context, email, groupName, role string) error
	RemoveUserFromGroup(ctx context.Context, email, groupName string) error
	GetUserGroupRoles(ctx context.Context, email string) ([]client.UserGroupRole, error)

	// User platform role methods
	AddPlatformRoleToUser(ctx context.Context, email, role string) error
	RemovePlatformRoleFromUser(ctx context.Context, email, role string) error
	GetUserPlatformRoles(ctx context.Context, email string) ([]string, error)
}
