package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/resources"
)

// NewProvider constructs the Lagoon provider with all resources wired.
func NewProvider(version string) p.Provider {
	return infer.Provider(infer.Options{
		Config: infer.Config[config.LagoonConfig](),
		Resources: []infer.InferredResource{
			infer.Resource[*resources.Project, resources.ProjectArgs, resources.ProjectState](),
			infer.Resource[*resources.Environment, resources.EnvironmentArgs, resources.EnvironmentState](),
			infer.Resource[*resources.Variable, resources.VariableArgs, resources.VariableState](),
			infer.Resource[*resources.DeployTarget, resources.DeployTargetArgs, resources.DeployTargetState](),
			infer.Resource[*resources.DeployTargetConfig, resources.DeployTargetConfigArgs, resources.DeployTargetConfigState](),
			infer.Resource[*resources.Task, resources.TaskArgs, resources.TaskState](),
			infer.Resource[*resources.NotificationSlack, resources.NotificationSlackArgs, resources.NotificationSlackState](),
			infer.Resource[*resources.NotificationRocketChat, resources.NotificationRocketChatArgs, resources.NotificationRocketChatState](),
			infer.Resource[*resources.NotificationEmail, resources.NotificationEmailArgs, resources.NotificationEmailState](),
			infer.Resource[*resources.NotificationMicrosoftTeams, resources.NotificationMicrosoftTeamsArgs, resources.NotificationMicrosoftTeamsState](),
			infer.Resource[*resources.ProjectNotification, resources.ProjectNotificationArgs, resources.ProjectNotificationState](),
		},
	})
}
