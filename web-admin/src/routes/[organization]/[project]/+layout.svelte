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
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
    type RpcStatus,
    type V1GetProjectResponse,
  } from "@rilldata/web-admin/client";
  import { isProjectPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectTabs from "@rilldata/web-admin/features/projects/ProjectTabs.svelte";
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { cloudVersion } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";

  import { type CreateQueryOptions } from "@tanstack/svelte-query";

  const user = createAdminServiceGetCurrentUser();

  export let data;

  $: ({
    url: { pathname },
    params: { organization, project },
  } = $page);

  $: onProjectPage = isProjectPage($page);

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

  // $: ({ data: projectData, error: projectError } = $projectQuery);

  $: projectData = data.project;

  // A re-deploy triggers `DEPLOYMENT_STATUS_UPDATING` status. But we can still show the project UI.
  $: isProjectAvailable =
    projectData?.prodDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
    projectData?.prodDeployment?.status ===
      V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING;

  let error: HTTPError | null = null;

  // $: authContext = (
  //   mockedUserId && mockedUserDeploymentCredentials
  //     ? "mock"
  //     : onPublicURLPage
  //       ? "magic"
  //       : "user"
  // ) as AuthContext;

  $: if (mockedUserId && mockedUserDeploymentCredentials) {
    void httpClient.updateQuerySettings({
      instanceId: mockedUserDeploymentCredentials.instanceId,
      host: mockedUserDeploymentCredentials.runtimeHost,
      token: mockedUserDeploymentCredentials.accessToken,
      authContext: "mock",
    });
  }

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
  {:else if isProjectAvailable}
    <slot />
  {/if}
{/if}
