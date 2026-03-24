package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	lagoon "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon"
	lagoonres "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon/lagoon"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		deploytargetId, err := cfg.TryInt("deploytargetId")
		if err != nil {
			deploytargetId = 1
		}

		lagoonCfg := config.New(ctx, "lagoon")
		apiUrl := lagoonCfg.Get("apiUrl")
		if apiUrl == "" {
			apiUrl = "http://localhost:7080/graphql"
		}
		token := lagoonCfg.GetSecret("token")

		prov, err := lagoon.NewProvider(ctx, "lagoon", &lagoon.ProviderArgs{
			ApiUrl:   pulumi.StringPtr(apiUrl),
			Token:    token,
			Insecure: pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}
		opts := pulumi.Provider(prov)

		project, err := lagoonres.NewProject(ctx, "go-test-project", &lagoonres.ProjectArgs{
			Name:                  pulumi.String("go-sdk-test"),
			GitUrl:                pulumi.String("git@github.com:example/go-sdk-test.git"),
			DeploytargetId:        pulumi.Int(deploytargetId),
			ProductionEnvironment: pulumi.StringPtr("main"),
			Branches:              pulumi.StringPtr("^main$"),
		}, opts)
		if err != nil {
			return err
		}

		devEnv, err := lagoonres.NewEnvironment(ctx, "go-test-dev-env", &lagoonres.EnvironmentArgs{
			Name:            pulumi.String("develop"),
			ProjectId:       project.LagoonId,
			DeployType:      pulumi.String("branch"),
			DeployBaseRef:   pulumi.StringPtr("develop"),
			EnvironmentType: pulumi.String("development"),
		}, pulumi.Provider(prov), pulumi.DependsOn([]pulumi.Resource{project}))
		if err != nil {
			return err
		}

		ctx.Export("projectId", project.LagoonId)
		ctx.Export("projectName", project.Name)
		ctx.Export("envName", devEnv.Name)

		return nil
	})
}
