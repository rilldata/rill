<script lang="ts">
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
    createAdminServiceGetProject,
    type AdminServiceGetProjectParams,
    type RpcStatus,
    type V1GetProjectResponse,
  } from "@rilldata/web-admin/client";
  import httpClient from "@rilldata/web-admin/client/http-client";
  import { isProjectPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectDashboardsListener from "@rilldata/web-admin/features/projects/ProjectDashboardsListener.svelte";
  import ProjectTabs from "@rilldata/web-admin/features/projects/ProjectTabs.svelte";
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { hasAccessToOriginalDashboard } from "@rilldata/web-admin/features/shareable-urls/state";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import type { QueryObserverOptions } from "@tanstack/svelte-query";

  const user = createAdminServiceGetCurrentUser();

  $: ({ organization, project, token } = $page.params);

  $: onProjectPage = isProjectPage($page);
  $: onMagicLinkPage = !!token;

  // Get Project Query Options
  const PollTimeWhenProjectDeploymentPending = 1000;
  const PollTimeWhenProjectDeploymentError = 5000;

  const baseQueryOptions: QueryObserverOptions<
    V1GetProjectResponse,
    RpcStatus
  > = {
    cacheTime: Math.min(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL, 1000 * 60 * 5), // Make sure we don't keep a stale JWT in the cache
    refetchInterval: (data) => {
      switch (data?.prodDeployment?.status) {
        case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
          return PollTimeWhenProjectDeploymentPending;

        case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
        case V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED:
          return PollTimeWhenProjectDeploymentError;

        case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
          return RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires
      }
    },
    refetchOnMount: true,
    refetchOnReconnect: true,
    select: (data: V1GetProjectResponse) => {
      if (data?.prodDeployment?.runtimeHost) {
        // Hack: in development, the runtime host is actually on port 8081
        data.prodDeployment.runtimeHost =
          data.prodDeployment.runtimeHost.replace(
            "localhost:9091",
            "localhost:8081",
          );
      }
      return data;
    },
  };

  /**
   * Query for the project with cookies
   */
  $: projectQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    {
      query: {
        ...baseQueryOptions,
      },
    },
  );

  $: ({
    data: cookieProjectData,
    error: cookieProjectDataError,
    isLoading: cookieProjectDataIsLoading,
  } = $projectQuery);

  /**
   * Query for project with token
   */
  const adminServiceGetProjectWithBearerToken = (
    organizationName: string,
    name: string,
    params?: AdminServiceGetProjectParams,
    signal?: AbortSignal,
  ) => {
    return httpClient<V1GetProjectResponse>({
      url: `/v1/organizations/${organizationName}/projects/${name}`,
      method: "get",
      params,
      signal,
      headers: {
        Authorization: `Bearer ${token}`,
      },
      withCredentials: false,
    });
  };

  $: magicLinkRuntimeQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    {
      query: {
        ...baseQueryOptions,
        queryKey: ["project", organization, project, "magicLink", token],
        queryFn: ({ signal }) =>
          adminServiceGetProjectWithBearerToken(
            organization,
            project,
            undefined,
            signal,
          ),
        enabled: !!token,
      },
    },
  );

  $: ({
    data: magicLinkProjectData,
    isLoading: magicLinkProjectDataIsLoading,
    error: magicLinkProjectDataError,
  } = $magicLinkRuntimeQuery);

  /**
   * Query for the project runtime when the user is being spoofed via the "View As" functionality
   */
  $: mockedUserId = $viewAsUserStore?.id;

  $: mockedUserDeploymentCredentialsQuery =
    createAdminServiceGetDeploymentCredentials(
      organization,
      project,
      {
        userId: mockedUserId,
      },
      {
        query: {
          enabled: !!mockedUserId,
          select: (data) => {
            if (data?.runtimeHost) {
              // Hack: in development, the runtime host is actually on port 8081
              data.runtimeHost = data.runtimeHost.replace(
                "localhost:9091",
                "localhost:8081",
              );
            }
            return data;
          },
        },
      },
    );

  $: ({ data: mockedUserDeploymentCredentials } =
    $mockedUserDeploymentCredentialsQuery);

  // Depending on the page, we use the results from the cookie query or the magic link query
  $: isLoading = onMagicLinkPage
    ? magicLinkProjectDataIsLoading
    : cookieProjectDataIsLoading;
  $: typedError = (
    onMagicLinkPage ? magicLinkProjectDataError : cookieProjectDataError
  ) as HTTPError;
  $: projectData = onMagicLinkPage ? magicLinkProjectData : cookieProjectData;

  // Assess if the user has access to the original dashboard
  $: if (cookieProjectData?.jwt && magicLinkProjectData?.jwt) {
    // Tell children that the user has access to the original dashboard
    hasAccessToOriginalDashboard.set(true);
  }

  // Load telemetry client with relevant context
  $: if (project && $user.data?.user?.id) {
    metricsService.loadCloudFields({
      isDev: window.location.host.startsWith("localhost"),
      projectId: project,
      organizationId: organization,
      userId: $user.data?.user?.id,
    });
  }
</script>

{#if onProjectPage && projectData?.prodDeployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_OK}
  <ProjectTabs />
{/if}

{#if isLoading}
  <!-- TODO: Add a loading state -->
{:else if typedError}
  <ErrorPage
    statusCode={typedError.response.status}
    header="Error fetching deployment"
    body={typedError.response.data.message}
  />
{:else if !projectData.prodDeployment}
  <!-- No deployment = the project is "hibernating" -->
  <RedeployProjectCta {organization} {project} />
{:else if projectData.prodDeployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING}
  <ProjectBuilding />
{:else if projectData.prodDeployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR}
  <ErrorPage
    statusCode={500}
    header="Deployment Error"
    body={projectData.prodDeployment.statusMessage !== ""
      ? projectData.prodDeployment.statusMessage
      : "There was an error deploying your project. Please contact support."}
  />
{:else if projectData.prodDeployment.status === V1DeploymentStatus.DEPLOYMENT_STATUS_OK}
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
  >
    <ProjectDashboardsListener>
      <slot />
    </ProjectDashboardsListener>
  </RuntimeProvider>
{/if}
