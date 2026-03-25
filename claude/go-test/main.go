package main

import (
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	lagoon "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon"
	lagoonres "github.com/tag1consulting/pulumi-lagoon-provider/sdk/go/lagoon/lagoon"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		deploytargetId := cfg.RequireInt("deploytargetId")

		lagoonCfg := config.New(ctx, "lagoon")
		apiUrl := lagoonCfg.Get("apiUrl")
		if apiUrl == "" {
			apiUrl = "http://localhost:7080/graphql"
		}
		token := lagoonCfg.GetSecret("token")

		// Only disable TLS for local development endpoints
		isLocal := strings.HasPrefix(apiUrl, "http://localhost") || strings.HasPrefix(apiUrl, "http://127.0.0.1")
		prov, err := lagoon.NewProvider(ctx, "lagoon", &lagoon.ProviderArgs{
			ApiUrl:   pulumi.StringPtr(apiUrl),
			Token:    token,
			Insecure: pulumi.BoolPtr(isLocal),
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
		}, opts, pulumi.DependsOn([]pulumi.Resource{project}))
		if err != nil {
			return err
		}

		// Test Variable resource
		_, err = lagoonres.NewVariable(ctx, "go-test-var", &lagoonres.VariableArgs{
			ProjectId: project.LagoonId,
			Name:      pulumi.String("TEST_VAR"),
			Value:     pulumi.String("hello-v022"),
			Scope:     pulumi.String("runtime"),
		}, opts, pulumi.DependsOn([]pulumi.Resource{project}))
		if err != nil {
			return err
		}

		// Test Group resource (new in v0.2.2)
		group, err := lagoonres.NewGroup(ctx, "go-test-group", &lagoonres.GroupArgs{
			Name: pulumi.String("v022-test-group"),
		}, opts)
		if err != nil {
			return err
		}

		// Test subgroup (exercises parentGroupName round-trip fix)
		_, err = lagoonres.NewGroup(ctx, "go-test-subgroup", &lagoonres.GroupArgs{
			Name:            pulumi.String("v022-test-subgroup"),
			ParentGroupName: pulumi.StringPtr("v022-test-group"),
		}, opts, pulumi.DependsOn([]pulumi.Resource{group}))
		if err != nil {
			return err
		}

		ctx.Export("projectId", project.LagoonId)
		ctx.Export("projectName", project.Name)
		ctx.Export("envName", devEnv.Name)
		ctx.Export("groupId", group.LagoonId)

		return nil
	})
}
