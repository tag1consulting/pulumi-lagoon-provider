package client

// GraphQL query and mutation constants for the Lagoon API.
// Organized by resource type.

// --- Project Operations ---

const mutationAddProject = `
mutation AddProject($input: AddProjectInput!) {
    addProject(input: $input) {
        id
        name
        gitUrl
        openshift {
            id
            name
        }
        productionEnvironment
        branches
        pullrequests
        openshiftProjectPattern
        autoIdle
        storageCalc
        created
    }
}`

const queryProjectByName = `
query ProjectByName($name: String!) {
    projectByName(name: $name) {
        id
        name
        gitUrl
        openshift {
            id
            name
        }
        productionEnvironment
        branches
        pullrequests
        openshiftProjectPattern
        autoIdle
        storageCalc
        created
    }
}`

const queryAllProjects = `
query AllProjects {
    allProjects {
        id
        name
        gitUrl
        openshift {
            id
            name
        }
        productionEnvironment
        branches
        pullrequests
        openshiftProjectPattern
        autoIdle
        storageCalc
        created
    }
}`

const mutationUpdateProject = `
mutation UpdateProject($input: UpdateProjectInput!) {
    updateProject(input: $input) {
        id
        name
        gitUrl
        openshift {
            id
            name
        }
        productionEnvironment
        branches
        pullrequests
        openshiftProjectPattern
        autoIdle
        storageCalc
    }
}`

const mutationDeleteProject = `
mutation DeleteProject($input: DeleteProjectInput!) {
    deleteProject(input: $input)
}`

// --- Environment Operations ---

const mutationAddOrUpdateEnvironment = `
mutation AddOrUpdateEnvironment($input: AddEnvironmentInput!) {
    addOrUpdateEnvironment(input: $input) {
        id
        name
        project {
            id
            name
        }
        environmentType
        deployType
        deployBaseRef
        deployHeadRef
        deployTitle
        autoIdle
        openshiftProjectName
        route
        routes
        created
    }
}`

const queryEnvironmentByName = `
query EnvironmentByName($name: String!, $project: Int!) {
    environmentByName(name: $name, project: $project) {
        id
        name
        project {
            id
            name
        }
        environmentType
        deployType
        deployBaseRef
        deployHeadRef
        deployTitle
        autoIdle
        openshiftProjectName
        route
        routes
        created
    }
}`

const mutationDeleteEnvironment = `
mutation DeleteEnvironment($input: DeleteEnvironmentInput!) {
    deleteEnvironment(input: $input)
}`

// --- Variable Operations (New API: v2.30.0+) ---

const mutationAddOrUpdateEnvVariableByName = `
mutation AddOrUpdateEnvVariableByName($input: EnvVariableByNameInput!) {
    addOrUpdateEnvVariableByName(input: $input) {
        id
        name
        value
        scope
    }
}`

const queryGetEnvVariablesByProjectEnvironmentName = `
query GetEnvVariablesByProjectEnvironmentName($input: EnvVariableByProjectEnvironmentNameInput!) {
    getEnvVariablesByProjectEnvironmentName(input: $input) {
        id
        name
        value
        scope
    }
}`

const mutationDeleteEnvVariableByName = `
mutation DeleteEnvVariableByName($input: DeleteEnvVariableByNameInput!) {
    deleteEnvVariableByName(input: $input)
}`

// --- Variable Operations (Legacy API: < v2.30.0) ---

const mutationAddEnvVariable = `
mutation AddEnvVariable($input: EnvVariableInput!) {
    addEnvVariable(input: $input) {
        id
        name
        value
        scope
    }
}`

const queryEnvVariablesByProjectEnvironment = `
query EnvVariablesByProjectEnvironment($project: Int!, $environment: Int) {
    envVariablesByProjectEnvironment(input: {project: $project, environment: $environment}) {
        id
        name
        value
        scope
        project {
            id
            name
        }
        environment {
            id
            name
        }
    }
}`

const mutationDeleteEnvVariable = `
mutation DeleteEnvVariable($input: DeleteEnvVariableInput!) {
    deleteEnvVariable(input: $input)
}`

const queryEnvironmentById = `
query EnvironmentById($id: Int!) {
    environmentById(id: $id) {
        id
        name
    }
}`

// --- Deploy Target (Kubernetes) Operations ---

const mutationAddKubernetes = `
mutation AddKubernetes($input: AddKubernetesInput!) {
    addKubernetes(input: $input) {
        id
        name
        consoleUrl
        cloudProvider
        cloudRegion
        sshHost
        sshPort
        buildImage
        disabled
        routerPattern
        created
    }
}`

const queryAllKubernetes = `
query AllKubernetes {
    allKubernetes {
        id
        name
        consoleUrl
        cloudProvider
        cloudRegion
        sshHost
        sshPort
        buildImage
        disabled
        routerPattern
        created
    }
}`

const mutationUpdateKubernetes = `
mutation UpdateKubernetes($input: UpdateKubernetesInput!) {
    updateKubernetes(input: $input) {
        id
        name
        consoleUrl
        cloudProvider
        cloudRegion
        sshHost
        sshPort
        buildImage
        disabled
        routerPattern
    }
}`

const mutationDeleteKubernetes = `
mutation DeleteKubernetes($input: DeleteKubernetesInput!) {
    deleteKubernetes(input: $input)
}`

// --- Deploy Target Config Operations ---

const mutationAddDeployTargetConfig = `
mutation AddDeployTargetConfig($input: AddDeployTargetConfigInput!) {
    addDeployTargetConfig(input: $input) {
        id
        weight
        branches
        pullrequests
        deployTargetProjectPattern
        deployTarget {
            id
            name
        }
        project {
            id
            name
        }
    }
}`

const queryDeployTargetConfigsByProjectId = `
query DeployTargetConfigsByProjectId($project: Int!) {
    deployTargetConfigsByProjectId(project: $project) {
        id
        weight
        branches
        pullrequests
        deployTargetProjectPattern
        deployTarget {
            id
            name
        }
        project {
            id
            name
        }
    }
}`

const mutationUpdateDeployTargetConfig = `
mutation UpdateDeployTargetConfig($input: UpdateDeployTargetConfigInput!) {
    updateDeployTargetConfig(input: $input) {
        id
        weight
        branches
        pullrequests
        deployTargetProjectPattern
        deployTarget {
            id
            name
        }
        project {
            id
            name
        }
    }
}`

const mutationDeleteDeployTargetConfig = `
mutation DeleteDeployTargetConfig($input: DeleteDeployTargetConfigInput!) {
    deleteDeployTargetConfig(input: $input)
}`

// --- Notification Operations ---

const mutationAddNotificationSlack = `
mutation AddNotificationSlack($input: AddNotificationSlackInput!) {
    addNotificationSlack(input: $input) {
        id
        name
        webhook
        channel
    }
}`

const mutationUpdateNotificationSlack = `
mutation UpdateNotificationSlack($input: UpdateNotificationSlackInput!) {
    updateNotificationSlack(input: $input) {
        id
        name
        webhook
        channel
    }
}`

const mutationDeleteNotificationSlack = `
mutation DeleteNotificationSlack($input: DeleteNotificationSlackInput!) {
    deleteNotificationSlack(input: $input)
}`

const mutationAddNotificationRocketChat = `
mutation AddNotificationRocketChat($input: AddNotificationRocketChatInput!) {
    addNotificationRocketChat(input: $input) {
        id
        name
        webhook
        channel
    }
}`

const mutationUpdateNotificationRocketChat = `
mutation UpdateNotificationRocketChat($input: UpdateNotificationRocketChatInput!) {
    updateNotificationRocketChat(input: $input) {
        id
        name
        webhook
        channel
    }
}`

const mutationDeleteNotificationRocketChat = `
mutation DeleteNotificationRocketChat($input: DeleteNotificationRocketChatInput!) {
    deleteNotificationRocketChat(input: $input)
}`

const mutationAddNotificationEmail = `
mutation AddNotificationEmail($input: AddNotificationEmailInput!) {
    addNotificationEmail(input: $input) {
        id
        name
        emailAddress
    }
}`

const mutationUpdateNotificationEmail = `
mutation UpdateNotificationEmail($input: UpdateNotificationEmailInput!) {
    updateNotificationEmail(input: $input) {
        id
        name
        emailAddress
    }
}`

const mutationDeleteNotificationEmail = `
mutation DeleteNotificationEmail($input: DeleteNotificationEmailInput!) {
    deleteNotificationEmail(input: $input)
}`

// --- Group Operations ---

const mutationAddGroup = `
mutation AddGroup($input: AddGroupInput!) {
    addGroup(input: $input) {
        id
        name
    }
}`

const queryAllGroups = `
query AllGroups {
    allGroups {
        id
        name
    }
}`

const mutationUpdateGroup = `
mutation UpdateGroup($input: UpdateGroupInput!) {
    updateGroup(input: $input) {
        id
        name
    }
}`

const mutationDeleteGroup = `
mutation DeleteGroup($input: DeleteGroupInput!) {
    deleteGroup(input: $input)
}`

const mutationAddNotificationMicrosoftTeams = `
mutation AddNotificationMicrosoftTeams($input: AddNotificationMicrosoftTeamsInput!) {
    addNotificationMicrosoftTeams(input: $input) {
        id
        name
        webhook
    }
}`

const mutationUpdateNotificationMicrosoftTeams = `
mutation UpdateNotificationMicrosoftTeams($input: UpdateNotificationMicrosoftTeamsInput!) {
    updateNotificationMicrosoftTeams(input: $input) {
        id
        name
        webhook
    }
}`

const mutationDeleteNotificationMicrosoftTeams = `
mutation DeleteNotificationMicrosoftTeams($input: DeleteNotificationMicrosoftTeamsInput!) {
    deleteNotificationMicrosoftTeams(input: $input)
}`

const queryAllNotifications = `
query AllNotifications {
    allNotifications {
        __typename
        ... on NotificationSlack {
            id
            name
            webhook
            channel
        }
        ... on NotificationRocketChat {
            id
            name
            webhook
            channel
        }
        ... on NotificationEmail {
            id
            name
            emailAddress
        }
        ... on NotificationMicrosoftTeams {
            id
            name
            webhook
        }
    }
}`

// --- Project Notification Association Operations ---

const mutationAddNotificationToProject = `
mutation AddNotificationToProject($input: AddNotificationToProjectInput!) {
    addNotificationToProject(input: $input) {
        id
        name
    }
}`

const mutationRemoveNotificationFromProject = `
mutation RemoveNotificationFromProject($input: RemoveNotificationFromProjectInput!) {
    removeNotificationFromProject(input: $input) {
        id
        name
    }
}`

const queryProjectNotifications = `
query ProjectByName($name: String!) {
    projectByName(name: $name) {
        id
        name
        notifications {
            ... on NotificationSlack {
                __typename
                id
                name
                webhook
                channel
            }
            ... on NotificationRocketChat {
                __typename
                id
                name
                webhook
                channel
            }
            ... on NotificationEmail {
                __typename
                id
                name
                emailAddress
            }
            ... on NotificationMicrosoftTeams {
                __typename
                id
                name
                webhook
            }
        }
    }
}`

// --- Advanced Task Definition Operations ---

const taskResponseFields = `
... on AdvancedTaskDefinitionCommand {
    id
    name
    description
    type
    service
    command
    permission
    confirmationText
    advancedTaskDefinitionArguments {
        id
        name
        displayName
        type
    }
    project {
        id
        name
    }
    environment {
        id
        name
    }
    groupName
    created
}
... on AdvancedTaskDefinitionImage {
    id
    name
    description
    type
    service
    image
    permission
    confirmationText
    advancedTaskDefinitionArguments {
        id
        name
        displayName
        type
    }
    project {
        id
        name
    }
    environment {
        id
        name
    }
    groupName
    created
}`

var mutationAddAdvancedTaskDefinition = `
mutation AddAdvancedTaskDefinition($input: AddAdvancedTaskDefinitionInput!) {
    addAdvancedTaskDefinition(input: $input) {
        ` + taskResponseFields + `
    }
}`

var queryAdvancedTaskDefinitionById = `
query AdvancedTaskDefinitionById($id: Int!) {
    advancedTaskDefinitionById(id: $id) {
        ` + taskResponseFields + `
    }
}`

const mutationDeleteAdvancedTaskDefinition = `
mutation DeleteAdvancedTaskDefinition($id: Int!) {
    deleteAdvancedTaskDefinition(id: $id)
}`

// --- Task queries with API version fallback ---

const queryAdvancedTasksForEnvironmentNew = `
query AdvancedTasksForEnvironment($environment: Int!) {
    advancedTasksForEnvironment(environment: $environment) {
        ... on AdvancedTaskDefinitionCommand {
            id
            name
            description
            type
            service
            command
            permission
            confirmationText
            advancedTaskDefinitionArguments {
                id
                name
                displayName
                type
            }
            project
            environment
            groupName
        }
        ... on AdvancedTaskDefinitionImage {
            id
            name
            description
            type
            service
            image
            permission
            confirmationText
            advancedTaskDefinitionArguments {
                id
                name
                displayName
                type
            }
            project
            environment
            groupName
        }
    }
}`

const queryAdvancedTasksByEnvironmentOld = `
query AdvancedTasksByEnvironment($environment: Int!) {
    advancedTasksByEnvironment(environment: $environment) {
        ... on AdvancedTaskDefinitionCommand {
            id
            name
            description
            type
            service
            command
            permission
            confirmationText
            advancedTaskDefinitionArguments {
                id
                name
                displayName
                type
            }
            project {
                id
                name
            }
            environment {
                id
                name
            }
            groupName
        }
        ... on AdvancedTaskDefinitionImage {
            id
            name
            description
            type
            service
            image
            permission
            confirmationText
            advancedTaskDefinitionArguments {
                id
                name
                displayName
                type
            }
            project {
                id
                name
            }
            environment {
                id
                name
            }
            groupName
        }
    }
}`

// --- Route Operations ---

// Shared response fields for route queries and mutations.
var routeResponseFields = `
    id
    domain
    service
    tlsAcme
    insecure
    primary
    source
    type
    disableRequestVerification
    hstsEnabled
    hstsPreload
    hstsIncludeSubdomains
    hstsMaxAge
    monitoringPath
    created
    updated
    annotations { key value }
    alternativeNames { id domain }
    pathRoutes { id path toService }
    project { name }
    environment { name }
`

var mutationAddRouteToProject = `
mutation AddRouteToProject($input: AddRouteToProjectInput!) {
    addRouteToProject(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationUpdateRouteOnProject = `
mutation UpdateRouteOnProject($input: UpdateRouteInput!) {
    updateRouteOnProject(input: $input) {
        ` + routeResponseFields + `
    }
}`

const mutationDeleteRoute = `
mutation DeleteRoute($input: DeleteRouteInput!) {
    deleteRoute(input: $input)
}`

var queryGetRouteByDomain = `
query GetRouteByDomain($name: String!, $domain: String!) {
    projectByName(name: $name) {
        apiRoutes(name: $domain) {
            ` + routeResponseFields + `
        }
    }
}`

var mutationAddOrUpdateRouteOnEnvironment = `
mutation AddOrUpdateRouteOnEnvironment($input: AddOrUpdateEnvironmentRouteInput!) {
    addOrUpdateRouteOnEnvironment(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationRemoveRouteFromEnvironment = `
mutation RemoveRouteFromEnvironment($input: RemoveRouteFromEnvironmentInput!) {
    removeRouteFromEnvironment(input: $input) {
        ` + routeResponseFields + `
    }
}`

// --- Route Sub-object Operations ---

var mutationAddRouteAlternativeDomains = `
mutation AddRouteAlternativeDomains($input: AddRouteAlternativeDomainInput!) {
    addRouteAlternativeDomains(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationRemoveRouteAlternativeDomain = `
mutation RemoveRouteAlternativeDomain($input: RemoveRouteAlternativeDomainInput!) {
    removeRouteAlternativeDomain(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationAddRouteAnnotation = `
mutation AddRouteAnnotation($input: AddRouteAnnotationsInput!) {
    addRouteAnnotation(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationRemoveRouteAnnotation = `
mutation RemoveRouteAnnotation($input: RemoveRouteAnnotationInput!) {
    removeRouteAnnotation(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationAddPathRoutesToRoute = `
mutation AddPathRoutesToRoute($input: AddPathRoutesInput!) {
    addPathRoutesToRoute(input: $input) {
        ` + routeResponseFields + `
    }
}`

var mutationRemovePathRouteFromRoute = `
mutation RemovePathRouteFromRoute($input: RemovePathRoutesInput!) {
    removePathRouteFromRoute(input: $input) {
        ` + routeResponseFields + `
    }
}`

// --- Autogenerated Route Config Operations ---

const mutationUpdateAutogeneratedRouteConfigOnProject = `
mutation UpdateAutogeneratedRouteConfigOnProject($input: UpdateProjectAutogeneratedRouteConfigInput!) {
    updateAutogeneratedRouteConfigOnProject(input: $input) {
        enabled
        allowPullRequests
        prefixes
        pathRoutes { fromService toService path }
        disableRequestVerification
        insecure
        tlsAcme
    }
}`

const mutationRemoveAutogeneratedRouteConfigFromProject = `
mutation RemoveAutogeneratedRouteConfigFromProject($project: String!) {
    removeAutogeneratedRouteConfigFromProject(project: $project)
}`

const queryGetProjectAutogeneratedRouteConfig = `
query GetProjectAutogeneratedRouteConfig($name: String!) {
    projectByName(name: $name) {
        autogeneratedRouteConfig {
            enabled
            allowPullRequests
            prefixes
            pathRoutes { fromService toService path }
            disableRequestVerification
            insecure
            tlsAcme
        }
    }
}`

const mutationUpdateAutogeneratedRouteConfigOnEnvironment = `
mutation UpdateAutogeneratedRouteConfigOnEnvironment($input: UpdateEnvironmentAutogeneratedRouteConfigInput!) {
    updateAutogeneratedRouteConfigOnEnvironment(input: $input) {
        enabled
        allowPullRequests
        prefixes
        pathRoutes { fromService toService path }
        disableRequestVerification
        insecure
        tlsAcme
    }
}`

const mutationRemoveAutogeneratedRouteConfigFromEnvironment = `
mutation RemoveAutogeneratedRouteConfigFromEnvironment($project: String!, $environment: String!) {
    removeAutogeneratedRouteConfigFromEnvironment(project: $project, environment: $environment)
}`

const queryGetEnvironmentAutogeneratedRouteConfig = `
query GetEnvironmentAutogeneratedRouteConfig($name: String!, $project: Int!) {
    environmentByName(name: $name, project: $project) {
        autogeneratedRouteConfig {
            enabled
            allowPullRequests
            prefixes
            pathRoutes { fromService toService path }
            disableRequestVerification
            insecure
            tlsAcme
        }
    }
}`
