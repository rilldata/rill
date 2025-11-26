<script context="module" lang="ts">
  const PollTimeWhenProjectDeploymentPending = 1000;
  const PollTimeWhenProjectDeploymentError = 5000;
  const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires

  const baseGetProjectQueryOptions: Partial<
    CreateQueryOptions<V1GetProjectResponse, RpcStatus>
  > = {
    gcTime: Math.min(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL, 1000 * 60 * 5), // Make sure we don't keep a stale JWT in the cache
    refetchInterval: (query) => {
      switch (query.state.data?.prodDeployment?.status) {
        case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
          return PollTimeWhenProjectDeploymentPending;
        case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
          return PollTimeWhenProjectDeploymentError;
        case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
          return PollTimeWhenProjectDeploymentOk;
        default:
          return false;
      }
    },
    refetchOnMount: true,
    refetchOnReconnect: true,
    refetchOnWindowFocus: true,
  };
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
    createAdminServiceGetProject,
    type RpcStatus,
    type V1GetProjectResponse,
  } from "@rilldata/web-admin/client";
  import {
    isProjectPage,
    isPublicAlertPage,
    isPublicReportPage,
    isPublicURLPage,
  } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectTabs from "@rilldata/web-admin/features/projects/ProjectTabs.svelte";
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { createAdminServiceGetProjectWithBearerToken } from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
  import { cloudVersion } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { AuthContext } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryOptions } from "@tanstack/svelte-query";

  const user = createAdminServiceGetCurrentUser();

  $: ({
    url: { pathname },
    params: { organization, project, token },
  } = $page);

  $: onProjectPage = isProjectPage($page);
  $: onPublicURLPage = isPublicURLPage($page);
  $: onPublicReportOrAlertPage =
    isPublicReportPage($page) || isPublicAlertPage($page);
  $: if (onPublicReportOrAlertPage) {
    token = $page.url.searchParams.get("token");
  }

  /**
   * `GetProject` with default cookie-based auth.
   * This returns the deployment credentials for the current logged-in user.
   */
  $: cookieProjectQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    {
      query: baseGetProjectQueryOptions,
    },
  );

  /**
   * `GetProject` with token-based auth.
   * This returns the deployment credentials for anonymous users who visit a Public URL.
   * The token is provided via the `[organization]/[project]/-/share/[token]` URL.
   */
  $: tokenProjectQuery = createAdminServiceGetProjectWithBearerToken(
    organization,
    project,
    token,
    undefined,
    {
      query: baseGetProjectQueryOptions,
    },
  );

  $: projectQuery = onPublicURLPage ? tokenProjectQuery : cookieProjectQuery;

  /**
   * `GetDeploymentCredentials`
   * This returns the deployment credentials for mocked/simulated users (aka the "View As" functionality).
   */
  $: mockedUserId = $viewAsUserStore?.id;
  $: mockedUserDeploymentCredentialsQuery =
    createAdminServiceGetDeploymentCredentials(
      organization,
      project,
      { userId: mockedUserId },
      {
        query: {
          enabled: !!mockedUserId,
        },
      },
    );
  $: ({ data: mockedUserDeploymentCredentials } =
    $mockedUserDeploymentCredentialsQuery);

  $: ({ data: projectData, error: projectError } = $projectQuery);

  $: error = projectError as HTTPError;

  $: authContext = (
    mockedUserId && mockedUserDeploymentCredentials
      ? "mock"
      : onPublicURLPage
        ? "magic"
        : "user"
  ) as AuthContext;

  // Load telemetry client with relevant context
  $: if (project && $user.data?.user?.id) {
    metricsService?.loadCloudFields({
      isDev: window.location.host.startsWith("localhost"),
      projectId: project,
      organizationId: organization,
      userId: $user.data?.user?.id,
      version: cloudVersion,
    });
  }
</script>

{#if onProjectPage && projectData?.prodDeployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING}
  <ProjectTabs
    projectPermissions={projectData.projectPermissions}
    {organization}
    {pathname}
    {project}
  />
{/if}

{#if error}
  <ErrorPage
    statusCode={error.response.status}
    header="Error fetching deployment"
    body={error.response.data?.message}
  />
{:else if projectData}
  {#if !projectData.prodDeployment}
    <!-- No deployment = the project is "hibernating" -->
    <RedeployProjectCta {organization} {project} />
  {:else if projectData.prodDeployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
    <ProjectBuilding />
  {:else if projectData.prodDeployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED}
    <ErrorPage
      statusCode={500}
      header="Deployment Error"
      body={projectData.prodDeployment.statusMessage !== ""
        ? projectData.prodDeployment.statusMessage
        : "There was an error deploying your project. Please contact support."}
    />
  {:else if projectData.prodDeployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING}
    <RuntimeProvider
      instanceId={mockedUserId && mockedUserDeploymentCredentials
        ? mockedUserDeploymentCredentials.instanceId
        : projectData.prodDeployment.runtimeInstanceId}
      host={mockedUserId && mockedUserDeploymentCredentials
        ? mockedUserDeploymentCredentials.runtimeHost
        : projectData.prodDeployment.runtimeHost}
      jwt={mockedUserId && mockedUserDeploymentCredentials
        ? mockedUserDeploymentCredentials.accessToken
        : projectData.jwt}
      {authContext}
    >
      <slot />
    </RuntimeProvider>
  {/if}
{/if}
