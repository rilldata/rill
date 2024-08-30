<script lang="ts">
  import { page } from "$app/stores";
  import {
    V1DeploymentStatus,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
    createAdminServiceGetProject,
  } from "@rilldata/web-admin/client";
  import { isProjectPage } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectTabs from "@rilldata/web-admin/features/projects/ProjectTabs.svelte";
  import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/status/selectors";
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import { createAdminServiceGetProjectWithBearerToken } from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";
  import type { AuthContext } from "@rilldata/web-common/runtime-client/runtime-store";

  const user = createAdminServiceGetCurrentUser();

  $: ({ organization, project, token } = $page.params);
  $: onProjectPage = isProjectPage($page);
  $: onPublicURLPage = !!token;

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
      {
        userId: mockedUserId,
      },
      {
        query: {
          enabled: !!mockedUserId,
          select: (data) => {
            if (data?.runtimeHost) {
              data.runtimeHost = fixLocalhostRuntimePort(data.runtimeHost);
            }
            return data;
          },
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
      {authContext}
    >
      <slot />
    </RuntimeProvider>
  {/if}
{/if}
