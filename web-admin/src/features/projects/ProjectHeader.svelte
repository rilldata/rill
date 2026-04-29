<script lang="ts">
  import { page } from "$app/stores";
  import CanvasBookmarks from "@rilldata/web-admin/features/bookmarks/CanvasBookmarks.svelte";
  import ExploreBookmarks from "@rilldata/web-admin/features/bookmarks/ExploreBookmarks.svelte";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import ShareDashboardPopover from "@rilldata/web-admin/features/dashboards/share/ShareDashboardPopover.svelte";
  import { isEditPreviewRoute } from "@rilldata/web-admin/features/edit-session/edit-route-utils";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import EditButton from "@rilldata/web-admin/features/edit-session/EditButton.svelte";
  import ShareProjectPopover from "@rilldata/web-admin/features/projects/user-management/ShareProjectPopover.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { GitBranchIcon } from "lucide-svelte";
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
  import BranchSelector from "../branches/BranchSelector.svelte";
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
  } from "../navigation/nav-utils";
  import PageTitle from "../public-urls/PageTitle.svelte";
  import { useReports } from "../scheduled-reports/selectors";

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
    cloudEditing,
    developerChat,
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

  $: activeBranch = extractBranchFromPath($page.url.pathname);

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

  $: canvasQuery = useCanvas(runtimeClient, dashboard, {
    enabled:
      !!runtimeClient.instanceId &&
      !!dashboard &&
      !!onCanvasDashboardPage &&
      !!onPublicURLPage,
  });

  $: publicURLDashboardTitle = onCanvasDashboardPage
    ? $canvasQuery.data?.canvas?.displayName || dashboard
    : $exploreQuery.data?.explore?.explore?.state?.validSpec?.displayName ||
      dashboard;

  // Cloud editor hides the org breadcrumb (and its trial pill) so the bar
  // collapses to "Rill / Project / Branch". Outside the editor, all
  // breadcrumb levels render.
  $: currentPath = editContext
    ? [undefined, project, dashboard, report || alert]
    : [organization, project, dashboard, report || alert];

  // Inside /-/edit, mode is a status indicator only — Preview vs Developer
  // is reflected by the URL prefix and the pill just labels it. Toggling
  // happens via the EditActions Preview/Edit button in the action bar.
  $: editPreviewMode = editContext && isEditPreviewRoute($page.url.pathname);
</script>

<Header borderBottom={!onProjectPage}>
  <HeaderLogo
    href={editContext ? `/${organization}/${project}` : rillLogoHref}
    logoUrl={editContext ? undefined : organizationLogoUrl}
  />
  {#if editContext && activeBranch}
    <Tag text={editPreviewMode ? "Preview" : "Developer"} color="gray" />
  {/if}
  {#if onPublicURLPage}
    <PageTitle title={publicURLDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath}>
      <svelte:fragment slot="after-project">
        {#if editContext && activeBranch}
          <Slash />
          <li class="flex items-center gap-x-1.5 px-2">
            <GitBranchIcon size="14" class="text-fg-muted" />
            <span class="text-fg-muted">
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
      {#if editPreviewMode && $viewAsUserStore}
        <ViewAsUserChip />
      {/if}
      {#if $developerChat}
        <ChatToggle />
      {/if}
      <EditActions {organization} {project} branch={activeBranch ?? ""} />
    {:else}
      {#if $viewAsUserStore}
        <ViewAsUserChip />
      {/if}
      {#if $cloudEditing && onProjectPage && projectPermissions.manageDev}
        <EditButton {organization} {project} {activeBranch} />
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
            {#if $cloudEditing && (onMetricsExplorerPage || onCanvasDashboardPage) && projectPermissions.manageDev}
              <EditButton {organization} {project} {activeBranch} />
            {/if}
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
      {#if $cloudEditing && projectPermissions.manageDev}
        <EditButton {organization} {project} {activeBranch} />
      {/if}
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

    {#if $user.isSuccess}
      {#if $user.data?.user}
        <AvatarButton {projectPermissions} />
      {:else}
        <SignIn />
      {/if}
    {/if}
  </div>
</Header>
