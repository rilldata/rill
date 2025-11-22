<script lang="ts">
  import { page } from "$app/stores";
  import CanvasBookmarks from "@rilldata/web-admin/features/bookmarks/CanvasBookmarks.svelte";
  import ExploreBookmarks from "@rilldata/web-admin/features/bookmarks/ExploreBookmarks.svelte";
  import ShareDashboardPopover from "@rilldata/web-admin/features/dashboards/share/ShareDashboardPopover.svelte";
  import ShareProjectPopover from "@rilldata/web-admin/features/projects/user-management/ShareProjectPopover.svelte";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations as listOrgs,
    createAdminServiceListProjectsForOrganization as listProjects,
    type V1Organization,
  } from "../../client";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../../features/view-as-user/viewAsUserStore";
  import CreateAlert from "../alerts/CreateAlert.svelte";
  import { useAlerts } from "../alerts/selectors";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import SignIn from "../authentication/SignIn.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import { useDashboards } from "../dashboards/listing/selectors";
  import PageTitle from "../public-urls/PageTitle.svelte";
  import { useReports } from "../scheduled-reports/selectors";
  import {
    isCanvasDashboardPage,
    isMetricsExplorerPage,
    isOrganizationPage,
    isProjectPage,
    isPublicURLPage,
  } from "./nav-utils";

  export let createMagicAuthTokens: boolean;
  export let manageProjectAdmins: boolean;
  export let manageProjectMembers: boolean;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;
  export let readProjects: boolean;
  export let organizationLogoUrl: string | undefined = undefined;
  export let planDisplayName: string | undefined;

  const user = createAdminServiceGetCurrentUser();
  const { alerts: alertsFlag, dimensionSearch, dashboardChat } = featureFlags;

  $: ({ instanceId } = $runtime);

  // These can be undefined
  $: ({
    params: { organization, project, dashboard, alert, report },
  } = $page);

  $: onProjectPage = isProjectPage($page);
  $: onAlertPage = !!alert;
  $: onReportPage = !!report;
  $: onMetricsExplorerPage = isMetricsExplorerPage($page);
  $: onCanvasDashboardPage = isCanvasDashboardPage($page);
  $: onPublicURLPage = isPublicURLPage($page);
  $: onOrgPage = isOrganizationPage($page);

  $: loggedIn = !!$user.data?.user;
  $: rillLogoHref = !loggedIn ? "https://www.rilldata.com" : "/";

  $: organizationQuery = listOrgs(
    { pageSize: 100 },
    {
      query: {
        enabled: !!$user.data?.user,
        retry: 2,
        refetchOnMount: true,
      },
    },
  );

  $: projectsQuery = listProjects(
    organization,
    {
      pageSize: 100,
    },
    {
      query: {
        enabled: !!organization && readProjects,
        retry: 2,
        refetchOnMount: true,
      },
    },
  );

  $: visualizationsQuery = useDashboards(instanceId);

  $: alertsQuery = useAlerts(instanceId, onAlertPage);
  $: reportsQuery = useReports(instanceId, onReportPage);

  $: organizations = $organizationQuery.data?.organizations ?? [];
  $: projects = $projectsQuery.data?.projects ?? [];
  $: visualizations = $visualizationsQuery.data ?? [];
  $: alerts = $alertsQuery.data?.resources ?? [];
  $: reports = $reportsQuery.data?.resources ?? [];

  $: organizationPaths = createOrgPaths(
    organizations,
    organization,
    planDisplayName,
  );

  function createOrgPaths(
    organizations: V1Organization[],
    viewingOrg: string | undefined,
    planDisplayName: string,
  ) {
    const pathMap = new Map<string, PathOption>();

    organizations.forEach(({ name, displayName }) => {
      pathMap.set(name.toLowerCase(), {
        label: displayName || name,
        pill: planDisplayName,
      });
    });

    if (!viewingOrg) return pathMap;

    if (!pathMap.has(viewingOrg.toLowerCase())) {
      pathMap.set(viewingOrg.toLowerCase(), {
        label: viewingOrg,
        pill: planDisplayName,
      });
    }

    return pathMap;
  }

  $: projectPaths = projects.reduce(
    (map, { name }) =>
      map.set(name.toLowerCase(), { label: name, preloadData: false }),
    new Map<string, PathOption>(),
  );

  $: visualizationPaths = visualizations.reduce((map, resource) => {
    const name = resource.meta.name.name;
    const isMetricsExplorer = !!resource?.explore;
    return map.set(name.toLowerCase(), {
      label:
        (isMetricsExplorer
          ? resource?.explore?.spec?.displayName
          : resource?.canvas?.spec?.displayName) || name,
      section: isMetricsExplorer ? "explore" : "canvas",
    });
  }, new Map<string, PathOption>());

  $: alertPaths = alerts.reduce((map, alert) => {
    const name = alert.meta.name.name;
    return map.set(name.toLowerCase(), {
      label: alert.alert.spec.displayName || name,
      section: "-/alerts",
    });
  }, new Map<string, PathOption>());

  $: reportPaths = reports.reduce((map, report) => {
    const name = report.meta.name.name;
    return map.set(name.toLowerCase(), {
      label: report.report.spec.displayName || name,
      section: "-/reports",
    });
  }, new Map<string, PathOption>());

  $: pathParts = [
    organizationPaths,
    projectPaths,
    visualizationPaths,
    report ? reportPaths : alert ? alertPaths : null,
  ];

  $: exploreQuery = useExplore(instanceId, dashboard, {
    enabled: !!instanceId && !!dashboard && !!onMetricsExplorerPage,
  });
  $: exploreSpec = $exploreQuery.data?.explore?.explore?.state?.validSpec;
  $: isDashboardValid = !!exploreSpec;

  $: publicURLDashboardTitle =
    $exploreQuery.data?.explore?.explore?.state?.validSpec?.displayName ||
    dashboard;

  $: currentPath = [organization, project, dashboard, report || alert];
</script>

<div
  class="flex items-center w-full pr-4 pl-2 py-1"
  class:border-b={!onProjectPage && !onOrgPage}
>
  <!-- Left side -->
  <a
    href={rillLogoHref}
    class="grid place-content-center rounded {organizationLogoUrl
      ? 'pl-2 pr-2'
      : 'p-2'}"
  >
    {#if organizationLogoUrl}
      <img src={organizationLogoUrl} alt="logo" class="h-7" />
    {:else}
      <Rill />
    {/if}
  </a>
  {#if onPublicURLPage}
    <PageTitle title={publicURLDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath} />
  {/if}

  <!-- Right side -->
  <div class="flex gap-x-2 items-center ml-auto">
    {#if $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    <!-- NOTE: only project admin and editor can manage project members -->
    <!-- https://docs.rilldata.com/manage/roles-permissions#project-level-permissions -->
    {#if onProjectPage && manageProjectMembers}
      <ShareProjectPopover
        {organization}
        {project}
        {manageProjectAdmins}
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
          >
            <LastRefreshedDate {dashboard} />
            {#if $dimensionSearch}
              <GlobalDimensionSearch />
            {/if}
            {#if $dashboardChat}
              <ChatToggle />
            {/if}
            {#if $user.isSuccess && $user.data.user && !onPublicURLPage}
              <ExploreBookmarks
                {organization}
                {project}
                metricsViewName={exploreSpec.metricsView}
                exploreName={dashboard}
              />
              {#if $alertsFlag}
                <CreateAlert />
              {/if}
              <ShareDashboardPopover {createMagicAuthTokens} />
            {/if}
          </StateManagersProvider>
        {/key}
      {/if}
    {/if}

    {#if onCanvasDashboardPage}
      <CanvasBookmarks {organization} {project} canvasName={dashboard} />
      <ShareDashboardPopover createMagicAuthTokens={false} />
    {/if}
    {#if $user.isSuccess}
      {#if $user.data && $user.data.user}
        <AvatarButton />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</div>
