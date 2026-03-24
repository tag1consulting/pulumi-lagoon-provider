import * as pulumi from "@pulumi/pulumi";
import * as lagoon from "@tag1consulting/pulumi-lagoon";

const config = new pulumi.Config();
const deploytargetId = config.requireNumber("deploytargetId");

const lagoonCfg = new pulumi.Config("lagoon");
const apiUrl = lagoonCfg.get("apiUrl") || "http://localhost:7080/graphql";
const token = lagoonCfg.getSecret("token");

// Create Lagoon provider
const lagoonProvider = new lagoon.Provider("lagoon", {
    apiUrl: apiUrl,
    token: token,
    insecure: true,
});
const opts: pulumi.ResourceOptions = { provider: lagoonProvider };

// Create a test project
const project = new lagoon.lagoon.Project("ts-test-project", {
    name: "ts-sdk-test",
    gitUrl: "git@github.com:example/ts-sdk-test.git",
    deploytargetId: deploytargetId,
    productionEnvironment: "main",
    branches: "^main$",
}, opts);

// Create a development environment
const devEnv = new lagoon.lagoon.Environment("ts-test-dev-env", {
    name: "develop",
    projectId: project.lagoonId,
    deployType: "branch",
    deployBaseRef: "develop",
    environmentType: "development",
}, { ...opts, dependsOn: [project] });

export const projectId = project.lagoonId;
export const projectName = project.name;
export const envName = devEnv.name;
