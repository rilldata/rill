<script lang="ts">
  import { page } from "$app/stores";
  import CanvasBookmarks from "@rilldata/web-admin/features/bookmarks/CanvasBookmarks.svelte";
  import ExploreBookmarks from "@rilldata/web-admin/features/bookmarks/ExploreBookmarks.svelte";
  import ShareDashboardPopover from "@rilldata/web-admin/features/dashboards/share/ShareDashboardPopover.svelte";
  import ShareProjectPopover from "@rilldata/web-admin/features/projects/user-management/ShareProjectPopover.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { V1ProjectPermissions } from "../../client";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../../features/view-as-user/viewAsUserStore";
  import CreateAlert from "../alerts/CreateAlert.svelte";
  import { useAlerts } from "../alerts/selectors";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import { useDashboards } from "../dashboards/listing/selectors";
  import {
    useBreadcrumbOrgPaths,
    useBreadcrumbProjectPaths,
  } from "../navigation/breadcrumb-selectors";
  import {
    isCanvasDashboardPage,
    isMetricsExplorerPage,
    isProjectPage,
    isPublicURLPage,
    isQueryPage,
  } from "../navigation/nav-utils";
  import PageTitle from "../public-urls/PageTitle.svelte";
  import { useReports } from "../scheduled-reports/selectors";

  let {
    organization,
    project,
    projectPermissions,
    manageOrgAdmins,
    manageOrgMembers,
    readProjects,
    planDisplayName,
    organizationLogoUrl,
  }: {
    organization: string;
    project: string;
    projectPermissions: V1ProjectPermissions;
    manageOrgAdmins: boolean;
    manageOrgMembers: boolean;
    readProjects: boolean;
    planDisplayName: string | undefined;
    organizationLogoUrl: string | undefined;
  } = $props();

  const user = createAdminServiceGetCurrentUser();
  const runtimeClient = useRuntimeClient();
  const {
    alerts: alertsFlag,
    dimensionSearch,
    dashboardChat,
    stickyDashboardState,
  } = featureFlags;

  let dashboard = $derived($page.params.dashboard);
  let alert = $derived($page.params.alert);
  let report = $derived($page.params.report);

  let onAlertPage = $derived(!!alert);
  let onReportPage = $derived(!!report);
  let onProjectPage = $derived(isProjectPage($page));
  let onMetricsExplorerPage = $derived(isMetricsExplorerPage($page));
  let onCanvasDashboardPage = $derived(isCanvasDashboardPage($page));
  let onPublicURLPage = $derived(isPublicURLPage($page));
  let onQueryPage = $derived(isQueryPage($page));

  let loggedIn = $derived(!!$user.data?.user);
  let rillLogoHref = $derived(!loggedIn ? "https://www.rilldata.com" : "/");

  let orgPathsQuery = $derived(
    useBreadcrumbOrgPaths(loggedIn, organization, planDisplayName),
  );
  let projectPathsQuery = $derived(
    useBreadcrumbProjectPaths(organization, readProjects),
  );
  let visualizationsQuery = $derived(useDashboards(runtimeClient));
  let alertsQuery = $derived(useAlerts(runtimeClient, onAlertPage));
  let reportsQuery = $derived(useReports(runtimeClient, onReportPage));

  let visualizations = $derived($visualizationsQuery.data ?? []);
  let alerts = $derived($alertsQuery.data?.resources ?? []);
  let reports = $derived($reportsQuery.data?.resources ?? []);

  let visualizationPaths = $derived({
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
  });

  let alertPaths = $derived({
    options: alerts.reduce((map, alert) => {
      const name = alert.meta.name.name;
      return map.set(name.toLowerCase(), {
        label: alert.alert.spec.displayName || name,
        section: "-/alerts",
      });
    }, new Map<string, PathOption>()),
  });

  let reportPaths = $derived({
    options: reports.reduce((map, report) => {
      const name = report.meta.name.name;
      return map.set(name.toLowerCase(), {
        label: report.report.spec.displayName || name,
        section: "-/reports",
      });
    }, new Map<string, PathOption>()),
  });

  let pathParts = $derived([
    { options: $orgPathsQuery.data ?? new Map() },
    { options: $projectPathsQuery.data ?? new Map() },
    visualizationPaths,
    report ? reportPaths : alert ? alertPaths : null,
  ]);

  let exploreQuery = $derived(
    useExplore(runtimeClient, dashboard, {
      enabled:
        !!runtimeClient.instanceId && !!dashboard && !!onMetricsExplorerPage,
    }),
  );
  let exploreSpec = $derived(
    $exploreQuery.data?.explore?.explore?.state?.validSpec,
  );
  let isDashboardValid = $derived(!!exploreSpec);
  let hasUserAccess = $derived(
    $user.isSuccess && $user.data.user && !onPublicURLPage,
  );

  let canvasQuery = $derived(
    useCanvas(runtimeClient, dashboard, {
      enabled:
        !!runtimeClient.instanceId &&
        !!dashboard &&
        !!onCanvasDashboardPage &&
        !!onPublicURLPage,
    }),
  );

  let publicURLDashboardTitle = $derived(
    onCanvasDashboardPage
      ? $canvasQuery.data?.canvas?.displayName || dashboard
      : $exploreQuery.data?.explore?.explore?.state?.validSpec?.displayName ||
          dashboard,
  );

  let currentPath = $derived([
    organization,
    project,
    dashboard,
    report || alert,
  ]);
</script>

<Header borderBottom={!onProjectPage}>
  <HeaderLogo href={rillLogoHref} logoUrl={organizationLogoUrl} />
  {#if onPublicURLPage}
    <PageTitle title={publicURLDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath} />
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if onProjectPage && projectPermissions.manageProjectMembers}
      <ShareProjectPopover
        {organization}
        {project}
        manageProjectAdmins={projectPermissions.manageProjectAdmins}
        {manageOrgAdmins}
        {manageOrgMembers}
      />
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
                createMagicAuthTokens={projectPermissions.createMagicAuthTokens}
              />
            {/if}
          </StateManagersProvider>
        {/key}
      {/if}
    {/if}

    {#if onCanvasDashboardPage}
      {#if $dashboardChat && !onPublicURLPage}
        <ChatToggle />
      {/if}
      {#if hasUserAccess}
        <CanvasBookmarks {organization} {project} canvasName={dashboard} />
        <ShareDashboardPopover
          createMagicAuthTokens={projectPermissions.createMagicAuthTokens}
        />
      {/if}
    {/if}

    {#if onQueryPage && $dashboardChat}
      <ChatToggle />
    {/if}

    {#if $user.isSuccess}
      {#if $user.data?.user}
        <AvatarButton />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</Header>
