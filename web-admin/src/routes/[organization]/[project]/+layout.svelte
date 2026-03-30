<!--
  Project layout: connects to the project's runtime when the deployment is running,
  or shows a status page (building, error, hibernating) when it isn't.

  Four dimensions converge here:
    1. Auth mode: cookie (logged-in), bearer token (public URL), or mock (View As)
    2. Deployment lifecycle: running, pending, errored, hibernating
    3. Runtime connection: host, instanceId, JWT passed to RuntimeProvider
    4. Navigation chrome: full header + tabs when running, slim header otherwise
-->
<script lang="ts">
  import { page } from "$app/state";
  import { untrack } from "svelte";
  import type { Snippet } from "svelte";
  import {
    V1DeploymentStatus,
    type V1Organization,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
    createAdminServiceGetProject,
  } from "@rilldata/web-admin/client";
  import {
    isProjectPage,
    isPublicAlertPage,
    isPublicReportPage,
    isPublicURLPage,
  } from "@rilldata/web-admin/features/navigation/nav-utils";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectHeader from "@rilldata/web-admin/features/projects/ProjectHeader.svelte";
  import ProjectTabs from "@rilldata/web-admin/features/projects/ProjectTabs.svelte";
  import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/project-query-options";
  import { resolveRuntimeConnection } from "@rilldata/web-admin/features/projects/project-runtime";
  import RedeployProjectCta from "@rilldata/web-admin/features/projects/RedeployProjectCTA.svelte";
  import SlimProjectHeader from "@rilldata/web-admin/features/projects/SlimProjectHeader.svelte";
  import { createAdminServiceGetProjectWithBearerToken } from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
  import { cloudVersion } from "@rilldata/web-admin/features/telemetry/initCloudMetrics";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { metricsService } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";

  let { children }: { children: Snippet } = $props();

  // --- Page state ---

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);
  let pathname = $derived(page.url.pathname);
  let pageData = $derived(page.data);

  // Token: from route params, or from search params on report/alert pages
  let token = $derived.by(() => {
    if (isPublicReportPage(page) || isPublicAlertPage(page)) {
      return page.url.searchParams.get("token") ?? page.params.token;
    }
    return page.params.token;
  });

  let onProjectPage = $derived(isProjectPage(page));
  let onPublicURLPage = $derived(isPublicURLPage(page));

  // From root layout; passed through to header components
  let organizationPermissions = $derived(
    pageData?.organizationPermissions ?? {},
  );
  let planDisplayName = $derived(pageData?.planDisplayName);
  let organizationLogoUrl = $derived(
    getThemedLogoUrl(
      $themeControl,
      pageData?.organization as V1Organization | undefined,
    ),
  );

  // --- View As (admin impersonation of another user's permissions) ---

  let mockedUserId = $derived($viewAsUserStore?.id);

  // Initialize view-as store for current project scope; clear on scope change or unmount
  $effect(() => {
    if (organization && project) {
      viewAsUserStore.initForProject(organization, project);
    }
    return () => {
      viewAsUserStore.clear();
    };
  });

  // --- Queries (three auth strategies; cookie and token are mutually exclusive,
  //     mock is an overlay that runs in parallel when View As is active) ---

  const user = createAdminServiceGetCurrentUser();

  /**
   * `GetProject` with default cookie-based auth.
   */
  let cookieProjectQuery = $derived(
    createAdminServiceGetProject(organization, project, undefined, {
      query: baseGetProjectQueryOptions,
    }),
  );

  /**
   * `GetProject` with token-based auth.
   * Returns deployment credentials for anonymous users visiting a Public URL.
   */
  let tokenProjectQuery = $derived(
    createAdminServiceGetProjectWithBearerToken(
      organization,
      project,
      token,
      undefined,
      { query: baseGetProjectQueryOptions },
    ),
  );

  let projectQuery = $derived(
    onPublicURLPage ? tokenProjectQuery : cookieProjectQuery,
  );

  /**
   * `GetDeploymentCredentials` for "View As" (mocked/simulated user).
   */
  let mockedUserDeploymentCredentialsQuery = $derived(
    createAdminServiceGetDeploymentCredentials(
      organization,
      project,
      { userId: mockedUserId },
      { query: { enabled: !!mockedUserId } },
    ),
  );
  let mockedUserDeploymentCredentials = $derived(
    $mockedUserDeploymentCredentialsQuery.data,
  );

  /**
   * When "View As" is active, fetch the project using the mocked user's JWT.
   * Returns the impersonated user's `projectPermissions` from the server.
   */
  let mockedUserProjectQuery = $derived(
    createAdminServiceGetProjectWithBearerToken(
      organization,
      project,
      mockedUserDeploymentCredentials?.accessToken ?? "",
      undefined,
      {
        query: { enabled: !!mockedUserDeploymentCredentials?.accessToken },
      },
    ),
  );

  // --- Derived state (resolve effective runtime connection from whichever auth mode is active) ---

  let projectData = $derived($projectQuery.data);
  let projectError = $derived($projectQuery.error);
  let error = $derived(projectError as HTTPError);

  let deploymentStatus = $derived(projectData?.deployment?.status);
  let isProjectAvailable = $derived(
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING,
  );

  let runtime = $derived(
    resolveRuntimeConnection(
      projectData,
      mockedUserId,
      mockedUserDeploymentCredentials,
      $mockedUserProjectQuery.data?.projectPermissions,
      onPublicURLPage,
    ),
  );

  // --- Side effects (cache invalidation and telemetry) ---

  // Track previous status so we invalidate only on *transitions*, not on every
  // render where the status happens to be RUNNING.
  let prevDeploymentStatus: V1DeploymentStatus | undefined = $state(
    V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
  );
  $effect(() => {
    const currentStatus = deploymentStatus;
    const prevStatus = untrack(() => prevDeploymentStatus);
    if (currentStatus === prevStatus) return;

    prevDeploymentStatus = currentStatus;

    if (currentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING) {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          projectData?.deployment?.runtimeInstanceId,
        ),
      });
    }
  });

  $effect(() => {
    if (project && $user.data?.user?.id) {
      metricsService?.loadCloudFields({
        isDev: window.location.host.startsWith("localhost"),
        projectId: project,
        organizationId: organization,
        userId: $user.data?.user?.id,
        version: cloudVersion,
      });
    }
  });
</script>

{#if error}
  <SlimProjectHeader
    {organization}
    {project}
    readProjects={organizationPermissions?.readProjects}
    {planDisplayName}
    {organizationLogoUrl}
  />
  <ErrorPage
    statusCode={error.response.status}
    header="Error fetching deployment"
    body={error.response.data?.message}
  />
{:else if projectData}
  {#if isProjectAvailable && runtime.host != null && runtime.instanceId}
    <!-- Re-key on host::instanceId to force RuntimeProvider to tear down and
         reconnect when the deployment changes (e.g. View As). -->
    {#key `${runtime.host}::${runtime.instanceId}`}
      <RuntimeProvider
        host={runtime.host}
        instanceId={runtime.instanceId}
        jwt={runtime.jwt}
        authContext={runtime.authContext}
      >
        <ProjectHeader
          {organization}
          {project}
          projectPermissions={runtime.projectPermissions}
          manageOrgAdmins={organizationPermissions?.manageOrgAdmins}
          manageOrgMembers={organizationPermissions?.manageOrgMembers}
          readProjects={organizationPermissions?.readProjects}
          {planDisplayName}
          {organizationLogoUrl}
        />
        {#if onProjectPage && deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING}
          <ProjectTabs
            projectPermissions={runtime.projectPermissions}
            {organization}
            {pathname}
            {project}
          />
        {/if}
        {@render children()}
      </RuntimeProvider>
    {/key}
  {:else}
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
      {planDisplayName}
      {organizationLogoUrl}
    />
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
    {:else}
      <ProjectBuilding />
    {/if}
  {/if}
{/if}
