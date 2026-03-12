<script lang="ts">
  import { page } from "$app/stores";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import CanvasBookmarks from "@rilldata/web-admin/features/bookmarks/CanvasBookmarks.svelte";
  import ExploreBookmarks from "@rilldata/web-admin/features/bookmarks/ExploreBookmarks.svelte";
  import ShareDashboardPopover from "@rilldata/web-admin/features/dashboards/share/ShareDashboardPopover.svelte";
  import ShareProjectPopover from "@rilldata/web-admin/features/projects/user-management/ShareProjectPopover.svelte";
  import { createAdminServiceGetProjectWithBearerToken } from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { V1ProjectPermissions } from "../../client";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeploymentCredentials,
  } from "../../client";
  import {
    useBreadcrumbOrgPaths,
    useBreadcrumbProjectPaths,
  } from "../navigation/breadcrumb-selectors";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../../features/view-as-user/viewAsUserStore";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import EditButton from "@rilldata/web-admin/features/edit-session/EditButton.svelte";
  import CreateAlert from "../alerts/CreateAlert.svelte";
  import { useAlerts } from "../alerts/selectors";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import BranchSelector from "../branches/BranchSelector.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import { useDashboards } from "../dashboards/listing/selectors";
  import PageTitle from "../public-urls/PageTitle.svelte";
  import { useReports } from "../scheduled-reports/selectors";
  import {
    isCanvasDashboardPage,
    isMetricsExplorerPage,
    isProjectPage,
    isPublicURLPage,
  } from "../navigation/nav-utils";

  export let organization: string;
  export let project: string;
  export let projectPermissions: V1ProjectPermissions;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;
  export let readProjects: boolean;
  export let primaryBranch: string | undefined = undefined;
  export let planDisplayName: string | undefined;
  export let organizationLogoUrl: string | undefined;
  export let editContext: boolean = false;

  const user = createAdminServiceGetCurrentUser();
  const runtimeClient = useRuntimeClient();
  const {
    alerts: alertsFlag,
    dimensionSearch,
    dashboardChat,
    stickyDashboardState,
  } = featureFlags;

  $: ({
    params: { dashboard, alert, report },
  } = $page);

  $: onAlertPage = !!alert;
  $: onReportPage = !!report;
  $: onProjectPage = isProjectPage($page);
  $: onMetricsExplorerPage = isMetricsExplorerPage($page);
  $: onCanvasDashboardPage = isCanvasDashboardPage($page);
  $: onPublicURLPage = isPublicURLPage($page);

  // When "View As" is active, fetch deployment credentials for the mocked user.
  $: mockedUserId = $viewAsUserStore?.id;
  $: activeBranch = extractBranchFromPath($page.url.pathname);

  $: mockedCredentialsQuery = createAdminServiceGetDeploymentCredentials(
    organization,
    project,
    { userId: mockedUserId, ...(activeBranch ? { branch: activeBranch } : {}) },
    {
      query: {
        enabled: !!mockedUserId && !!organization && !!project,
      },
    },
  );

  $: mockedProjectQuery = createAdminServiceGetProjectWithBearerToken(
    organization,
    project,
    $mockedCredentialsQuery.data?.accessToken ?? "",
    undefined,
    {
      query: {
        enabled: !!$mockedCredentialsQuery.data?.accessToken,
      },
    },
  );

  // Use effective permissions when "View As" is active (from server)
  $: effectiveManageProjectMembers =
    $mockedProjectQuery.data?.projectPermissions?.manageProjectMembers ??
    projectPermissions.manageProjectMembers;
  $: effectiveCreateMagicAuthTokens =
    $mockedProjectQuery.data?.projectPermissions?.createMagicAuthTokens ??
    projectPermissions.createMagicAuthTokens;

  $: loggedIn = !!$user.data?.user;
  $: rillLogoHref = !loggedIn ? "https://www.rilldata.com" : "/";

  $: orgPathsQuery = useBreadcrumbOrgPaths(
    loggedIn,
    organization,
    planDisplayName,
  );
  $: projectPathsQuery = useBreadcrumbProjectPaths(organization, readProjects);
  $: visualizationsQuery = useDashboards(runtimeClient);
  $: alertsQuery = useAlerts(runtimeClient, onAlertPage);
  $: reportsQuery = useReports(runtimeClient, onReportPage);

  $: visualizations = $visualizationsQuery.data ?? [];
  $: alerts = $alertsQuery.data?.resources ?? [];
  $: reports = $reportsQuery.data?.resources ?? [];

  $: visualizationPaths = {
    options: [...visualizations]
      .sort((a, b) => {
        const aIsCanvas = !!a?.canvas;
        const bIsCanvas = !!b?.canvas;
        if (aIsCanvas !== bIsCanvas) return aIsCanvas ? -1 : 1;
        const aName = a.meta.name.name;
        const bName = b.meta.name.name;
        return aName.localeCompare(bName);
      })
      .reduce((map, resource) => {
        const name = resource.meta.name.name;
        const isMetricsExplorer = !!resource?.explore;
        return map.set(name.toLowerCase(), {
          label:
            (isMetricsExplorer
              ? resource?.explore?.spec?.displayName
              : resource?.canvas?.spec?.displayName) || name,
          section: isMetricsExplorer ? "explore" : "canvas",
          resourceKind: isMetricsExplorer
            ? ResourceKind.Explore
            : ResourceKind.Canvas,
        });
      }, new Map<string, PathOption>()),
    carryOverSearchParams: $stickyDashboardState,
  };

  $: alertPaths = {
    options: alerts.reduce((map, alert) => {
      const name = alert.meta.name.name;
      return map.set(name.toLowerCase(), {
        label: alert.alert.spec.displayName || name,
        section: "-/alerts",
      });
    }, new Map<string, PathOption>()),
  };

  $: reportPaths = {
    options: reports.reduce((map, report) => {
      const name = report.meta.name.name;
      return map.set(name.toLowerCase(), {
        label: report.report.spec.displayName || name,
        section: "-/reports",
      });
    }, new Map<string, PathOption>()),
  };

  $: pathParts = [
    { options: $orgPathsQuery.data ?? new Map() },
    { options: $projectPathsQuery.data ?? new Map() },
    visualizationPaths,
    report ? reportPaths : alert ? alertPaths : null,
  ];

  $: exploreQuery = useExplore(runtimeClient, dashboard, {
    enabled:
      !!runtimeClient.instanceId && !!dashboard && !!onMetricsExplorerPage,
  });
  $: exploreSpec = $exploreQuery.data?.explore?.explore?.state?.validSpec;
  $: isDashboardValid = !!exploreSpec;
  $: hasUserAccess = $user.isSuccess && $user.data.user && !onPublicURLPage;

  $: publicURLDashboardTitle =
    $exploreQuery.data?.explore?.explore?.state?.validSpec?.displayName ||
    dashboard;

  $: currentPath = [organization, project, dashboard, report || alert];
</script>

<Header borderBottom={!onProjectPage}>
  <HeaderLogo href={rillLogoHref} logoUrl={organizationLogoUrl} />
  {#if onPublicURLPage}
    <PageTitle title={publicURLDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath}>
      <svelte:fragment slot="after-project">
        {#if editContext && activeBranch}
          <li class="flex items-center mr-2">
            <span
              class="flex items-center gap-x-1 px-2 py-0 rounded-2xl border bg-primary-50 border-primary-200 text-primary-800"
            >
              {activeBranch.length > 12
                ? activeBranch.slice(0, 11) + "…"
                : activeBranch}
            </span>
          </li>
        {:else if !onPublicURLPage && projectPermissions?.readDev}
          <BranchSelector {organization} {project} {primaryBranch} />
        {/if}
      </svelte:fragment>
    </Breadcrumbs>
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if editContext}
      <EditActions {organization} {project} branch={activeBranch ?? ""} />
    {:else}
      {#if $viewAsUserStore}
        <ViewAsUserChip />
      {/if}
      {#if onProjectPage && projectPermissions.manageDev}
        <EditButton {organization} {project} {activeBranch} />
      {/if}
      {#if onProjectPage && effectiveManageProjectMembers}
        <ShareProjectPopover
          {organization}
          {project}
          manageProjectAdmins={projectPermissions.manageProjectAdmins}
          {manageOrgAdmins}
          {manageOrgMembers}
        />
      {/if}
    {/if}

    {#if onMetricsExplorerPage && isDashboardValid}
      {#if exploreSpec}
        {#key dashboard}
          <StateManagersProvider
            metricsViewName={exploreSpec.metricsView}
            exploreName={dashboard}
            let:ready
          >
            <LastRefreshedDate {dashboard} />
            {#if $dimensionSearch && ready}
              <GlobalDimensionSearch />
            {/if}
            {#if $dashboardChat && !onPublicURLPage}
              <ChatToggle />
            {/if}
            {#if hasUserAccess}
              <ExploreBookmarks
                {organization}
                {project}
                metricsViewName={exploreSpec.metricsView}
                exploreName={dashboard}
              />
              {#if $alertsFlag}
                <CreateAlert />
              {/if}
              <ShareDashboardPopover
                createMagicAuthTokens={effectiveCreateMagicAuthTokens}
              />
            {/if}
          </StateManagersProvider>
        {/key}
      {/if}
    {/if}

    {#if onCanvasDashboardPage && hasUserAccess}
      {#if $dashboardChat && !onPublicURLPage}
        <ChatToggle />
      {/if}
      <CanvasBookmarks {organization} {project} canvasName={dashboard} />
      <ShareDashboardPopover createMagicAuthTokens={false} />
    {/if}

    {#if $user.isSuccess}
      {#if $user.data?.user}
        <AvatarButton {projectPermissions} />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</Header>
