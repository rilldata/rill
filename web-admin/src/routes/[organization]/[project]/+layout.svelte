<!--
  Project layout: connects to the project's runtime when the deployment is running,
  or shows a status page (building, error, stopped, hibernating) when it isn't.

  Five dimensions converge here:
    1. Auth mode: cookie (logged-in), bearer token (public URL), or mock (View As)
    2. Branch: production vs. feature branch (extracted from URL's @branch segment)
    3. Deployment lifecycle: running, pending, errored, stopped, hibernating
    4. Runtime connection: host, instanceId, JWT passed to RuntimeProvider
    5. Navigation chrome: full header + tabs when running, slim header otherwise
-->
<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/state";
  import { onDestroy, untrack } from "svelte";
  import type { Snippet } from "svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    branchPathPrefix,
    extractBranchFromPath,
    handleBranchNavigation,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import {
    V1DeploymentStatus,
    type V1Organization,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
    createAdminServiceGetProject,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    getResourceFromPage,
    getScreenNameFromPage,
    isEditPage,
    isProjectInvitePage,
    isProjectPage,
    isPublicAlertPage,
    isPublicReportPage,
    isPublicURLPage,
    isProjectWelcomePage,
  } from "@rilldata/web-admin/features/navigation/nav-utils";
  import BranchDeploymentStopped from "@rilldata/web-admin/features/branches/BranchDeploymentStopped.svelte";
  import ProjectBuilding from "@rilldata/web-admin/features/projects/ProjectBuilding.svelte";
  import ProjectHeader from "../../../features/projects/header/ProjectHeader.svelte";
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
  import {
    behaviourEvent,
    metricsService,
  } from "@rilldata/web-common/metrics/initMetrics";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import type { HTTPError } from "@rilldata/web-common/lib/errors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { Throttler } from "@rilldata/web-common/lib/throttler";

  const PAGE_VIEW_THROTTLE_TIMEOUT = 250;

  let { children }: { children: Snippet } = $props();

  // --- Page state ---

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);

  let activeBranch = $derived(extractBranchFromPath(page.url.pathname));
  let branchPrefix = $derived(branchPathPrefix(activeBranch));

  // Inject the active branch segment into intra-project navigations
  beforeNavigate((nav) =>
    handleBranchNavigation(nav, activeBranch, organization, project, goto),
  );

  // Token: from route params, or from search params on report/alert pages
  let token = $derived.by(() => {
    if (isPublicReportPage(page) || isPublicAlertPage(page)) {
      return page.url.searchParams.get("token") ?? page.params.token;
    }
    return page.params.token;
  });

  let onProjectPage = $derived(isProjectPage(page));
  let onEditPage = $derived(isEditPage(page));
  let onInvitePage = $derived(isProjectInvitePage(page));
  let onPublicURLPage = $derived(isPublicURLPage(page));
  let onWelcomePage = $derived(isProjectWelcomePage(page));

  // From root layout; passed through to header components
  let organizationPermissions = $derived(
    page.data?.organizationPermissions ?? {},
  );
  let planDisplayName = $derived(page.data?.planDisplayName);
  let organizationLogoUrl = $derived(
    getThemedLogoUrl(
      $themeControl,
      page.data?.organization as V1Organization | undefined,
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
   * When `activeBranch` is set, the branch param is passed so the API
   * returns the branch deployment instead of production.
   */
  let cookieProjectQuery = $derived(
    createAdminServiceGetProject(
      organization,
      project,
      activeBranch ? { branch: activeBranch } : undefined,
      { query: baseGetProjectQueryOptions },
    ),
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
      {
        userId: mockedUserId,
        ...(activeBranch ? { branch: activeBranch } : {}),
      },
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
  let error = $derived($projectQuery.error as HTTPError);

  let deploymentStatus = $derived(projectData?.deployment?.status);
  let isProjectAvailable = $derived(
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING,
  );

  let mockUser = $derived(
    mockedUserId && mockedUserDeploymentCredentials
      ? {
          credentials: mockedUserDeploymentCredentials,
          permissions: $mockedUserProjectQuery.data?.projectPermissions,
        }
      : undefined,
  );

  let runtime = $derived(
    resolveRuntimeConnection(projectData, mockUser, onPublicURLPage),
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

    // Keep the BranchesSection's ListDeployments query in sync
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceListDeploymentsQueryKey(organization, project),
    });
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

  // Fire a page view for every dashboard, canvas and project page the user opens.
  // This runs off `page` rather than `afterNavigate` because the first page view has to wait for the
  // current user query above to resolve, which happens after the initial navigation has completed.
  //
  // Views are deduped on the path plus the explore view mode. The view mode is a query param rather
  // than a path segment, so keying on the path alone would never record a user switching to pivot or
  // time dimension detail. The rest of the query string is deliberately excluded: filters, time range
  // and sort change on nearly every interaction, and keying on them would emit an event per click.
  //
  // Events are throttled so that pages the user only passes through, like clicking into a dashboard
  // and immediately hitting back, don't each record a view. The throttler keeps the most recent
  // callback, so the page the user actually landed on is the one reported.
  const pageViewThrottler = new Throttler(
    PAGE_VIEW_THROTTLE_TIMEOUT,
    PAGE_VIEW_THROTTLE_TIMEOUT,
  );
  let lastTrackedView: string | undefined;
  $effect(() => {
    const trackedView = `${page.url.pathname}?view=${page.url.searchParams.get("view") ?? ""}`;
    // Events are dropped until loadCloudFields has run, so wait on the same inputs it needs.
    if (!project || !$user.data?.user?.id || trackedView === lastTrackedView)
      return;
    lastTrackedView = trackedView;
    // Capture the page's fields now; `page` will have moved on by the time the throttler fires.
    const screenName = getScreenNameFromPage(page);
    const resource = getResourceFromPage(page);
    pageViewThrottler.throttle(() =>
      behaviourEvent?.firePageViewEvent(
        screenName,
        resource.type,
        resource.name,
      ),
    );
  });

  // Navigating out of the project within the throttle window destroys this layout with a callback
  // still pending. Left alone it would fire from the page the user has already left, and the event's
  // `page_url` is read when it fires, so it would pair this page's resource with the next page's URL.
  onDestroy(() => pageViewThrottler.cancel());
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
    header={m.error_fetching_deployment()}
    body={error.response.data?.message}
  />
{:else if projectData}
  {#if onEditPage}
    <!-- Edit layout manages its own runtime and header -->
    {@render children()}
  {:else if onInvitePage}
    <!-- Invite is admin-server-only and doesn't depend on the runtime, so we
         render it immediately and let users invite teammates while the
         deployment provisions. Otherwise the layout would briefly show
         `ProjectBuilding` here while the just-created prod deployment is
         still PENDING. -->
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
      {planDisplayName}
      {organizationLogoUrl}
    />
    {@render children()}
  {:else if isProjectAvailable && runtime.host != null && runtime.instanceId}
    <!-- Re-key on host::instanceId to force RuntimeProvider to tear down and
         reconnect when the deployment changes (e.g. branch switch, View As). -->
    {#key `${runtime.host}::${runtime.instanceId}`}
      <RuntimeProvider
        host={runtime.host}
        instanceId={runtime.instanceId}
        jwt={runtime.jwt}
        authContext={runtime.authContext}
      >
        {#if !onWelcomePage}
          <ProjectHeader
            {organization}
            {project}
            projectPermissions={runtime.projectPermissions}
            manageOrgAdmins={organizationPermissions?.manageOrgAdmins}
            manageOrgMembers={organizationPermissions?.manageOrgMembers}
            readProjects={organizationPermissions?.readProjects}
            primaryBranch={projectData?.project?.primaryBranch}
            {planDisplayName}
            {organizationLogoUrl}
          />
          {#if onProjectPage && deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING}
            <ProjectTabs
              projectPermissions={runtime.projectPermissions}
              {organization}
              pathname={page.url.pathname}
              {project}
              {branchPrefix}
            />
          {/if}
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
      <ProjectBuilding branch={activeBranch} />
    {:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED}
      <ErrorPage
        statusCode={500}
        header={m.error_deployment_error()}
        body={projectData.deployment.statusMessage !== ""
          ? projectData.deployment.statusMessage
          : m.error_deploying_project()}
      />
    {:else if deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED || deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING}
      <BranchDeploymentStopped
        {organization}
        {project}
        deploymentId={projectData.deployment.id}
        status={deploymentStatus}
        canManage={!!runtime.projectPermissions?.manageDev}
        branch={activeBranch}
      />
    {:else}
      <ProjectBuilding branch={activeBranch} />
    {/if}
  {/if}
{/if}
