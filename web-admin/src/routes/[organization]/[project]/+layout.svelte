<script context="module" lang="ts">
  const PollTimeWhenProjectDeploymentPending = 1000;
  const PollTimeWhenProjectDeploymentError = 5000;
  const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires

  const baseGetProjectQueryOptions: Partial<
    CreateQueryOptions<V1GetProjectResponse, RpcStatus>
  > = {
    gcTime: Math.min(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL, 1000 * 60 * 5), // Make sure we don't keep a stale JWT in the cache
    refetchInterval: (query) => {
      switch (query.state.data?.deployment?.status) {
        case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
        case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
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
  import { onNavigate } from "$app/navigation";
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
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { AuthContext } from "@rilldata/web-common/runtime-client/v2/runtime-client";
  import type { CreateQueryOptions } from "@tanstack/svelte-query";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { onDestroy } from "svelte";

  const user = createAdminServiceGetCurrentUser();

  $: ({
    url: { pathname },
    params: { organization, project, token },
  } = $page);

  // Initialize view-as store for this project scope (loads from sessionStorage)
  $: if (organization && project) {
    viewAsUserStore.initForProject(organization, project);
  }

  // Clear view-as state when navigating to a different project
  onNavigate(({ from, to }) => {
    const changedProject =
      !from ||
      !to ||
      from.params?.organization !== to.params?.organization ||
      from.params?.project !== to.params?.project;
    if (changedProject) {
      viewAsUserStore.clear();
    }
  });

  // Clear view-as state when unmounting (e.g., navigating to org page)
  onDestroy(() => {
    viewAsUserStore.clear();
  });

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

  /**
   * When "View As" is active, fetch the project using the mocked user's JWT.
   * This returns the impersonated user's `projectPermissions` from the server.
   */
  $: mockedUserProjectQuery = createAdminServiceGetProjectWithBearerToken(
    organization,
    project,
    mockedUserDeploymentCredentials?.accessToken ?? "",
    undefined,
    {
      query: {
        enabled: !!mockedUserDeploymentCredentials?.accessToken,
      },
    },
  );

  $: ({ data: projectData, error: projectError } = $projectQuery);

  /**
   * Compute effective project permissions.
   * When "View As" is active, use the impersonated user's permissions (from server).
   * Otherwise, use the actual user's permissions.
   */
  $: effectiveProjectPermissions =
    mockedUserId && $mockedUserProjectQuery.data?.projectPermissions
      ? $mockedUserProjectQuery.data.projectPermissions
      : projectData?.projectPermissions;

  $: deploymentStatus = projectData?.deployment?.status;
  // A re-deploy triggers `DEPLOYMENT_STATUS_UPDATING` status. But we can still show the project UI.
  $: isProjectAvailable =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING;

  // Refetch list resource query when project query stops fetching.
  // This needs to happen when deployment status changes from updating to running after a redeploy.
  let prevDeploymentStatus: V1DeploymentStatus =
    V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
  $: if (prevDeploymentStatus !== deploymentStatus) {
    prevDeploymentStatus = deploymentStatus;
    if (deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING) {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          projectData.deployment.runtimeInstanceId,
        ),
      });
    }
  }

  $: error = projectError as HTTPError;

  $: authContext = (
    mockedUserId && mockedUserDeploymentCredentials
      ? "mock"
      : onPublicURLPage
        ? "magic"
        : "user"
  ) as AuthContext;

  // Derive effective runtime connection props
  $: effectiveHost =
    mockedUserId && mockedUserDeploymentCredentials
      ? mockedUserDeploymentCredentials.runtimeHost
      : projectData?.deployment?.runtimeHost;
  $: effectiveInstanceId =
    mockedUserId && mockedUserDeploymentCredentials
      ? mockedUserDeploymentCredentials.instanceId
      : projectData?.deployment?.runtimeInstanceId;
  $: effectiveJwt =
    mockedUserId && mockedUserDeploymentCredentials
      ? mockedUserDeploymentCredentials.accessToken
      : projectData?.jwt;

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

{#if onProjectPage && deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING}
  <ProjectTabs
    projectPermissions={effectiveProjectPermissions}
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
  {#if !projectData.deployment}
    <!-- No deployment = the project is "hibernating" -->
    <RedeployProjectCta {organization} {project} />
  {:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
    <ProjectBuilding />
  {:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED}
    <ErrorPage
      statusCode={500}
      header="Deployment Error"
      body={projectData.deployment.statusMessage !== ""
        ? projectData.deployment.statusMessage
        : "There was an error deploying your project. Please contact support."}
    />
  {:else if isProjectAvailable}
    {#key `${effectiveHost}::${effectiveInstanceId}`}
      <RuntimeProvider
        host={effectiveHost}
        instanceId={effectiveInstanceId}
        jwt={effectiveJwt}
        {authContext}
      >
        <slot />
      </RuntimeProvider>
    {/key}
  {/if}
{/if}
