<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import CanvasBookmarks from "@rilldata/web-admin/features/bookmarks/CanvasBookmarks.svelte";
  import ExploreBookmarks from "@rilldata/web-admin/features/bookmarks/ExploreBookmarks.svelte";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import ShareDashboardPopover from "@rilldata/web-admin/features/dashboards/share/ShareDashboardPopover.svelte";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import EditButton from "@rilldata/web-admin/features/edit-session/EditButton.svelte";
  import ShareProjectPopover from "@rilldata/web-admin/features/projects/user-management/ShareProjectPopover.svelte";
  import CloudViewAsButton from "@rilldata/web-admin/features/view-as-user/CloudViewAsButton.svelte";
  import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
  import { skipNextPlatformReset } from "@rilldata/web-common/features/preview-mode/platform-reset";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { GitBranch } from "lucide-svelte";
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import PreviewModeToggleButton from "@rilldata/web-common/layout/header/PreviewModeToggleButton.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { V1ProjectPermissions } from "../../client";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import ViewAsUserChip from "../../features/view-as-user/ViewAsUserChip.svelte";
  import ViewAsUserPopover from "../../features/view-as-user/ViewAsUserPopover.svelte";
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
    params: { dashboard, alert, report, name },
    route,
  } = $page);

  $: editPreviewKind = route.id?.includes("/-/edit/explore/")
    ? "explore"
    : route.id?.includes("/-/edit/canvas/")
      ? "canvas"
      : null;

  $: editPreviewHref =
    editPreviewKind && name
      ? `/${organization}/${project}${branchPathPrefix(activeBranch)}/${editPreviewKind}/${name}`
      : `/${organization}/${project}${branchPathPrefix(activeBranch)}/-/edit/dashboards`;

  // Cloud editor sub-routes that should adopt a "Dev preview" chrome
  // (View-as pill + Edit-back button) rather than the editor chrome
  // (split Preview button + EditActions). Mirrors local's preview mode.
  // Includes the dashboard view routes so a click into a dashboard from
  // the listing keeps the dev-preview chrome instead of falling back to
  // the in-workspace editor.
  $: inEditDevPreview = !!route.id?.match(
    /\/-\/edit\/(dashboards|status|ai|explore|canvas)(\/|$)/,
  );

  $: editBackHref = `/${organization}/${project}${branchPathPrefix(activeBranch)}/-/edit`;

  let editorViewAsOpen = false;

  // Reset session state (View-as impersonation + AI chat panel) when the
  // user explicitly swaps between editor and dev-preview chrome via the
  // Preview/Edit toggle. Mirrors local's `resetOnModeToggle`.
  function resetOnPreviewSwap() {
    sidebarActions.closeChat();
    if ($viewAsUserStore) viewAsUserStore.set(null);
  }

  // Belt-and-suspenders: also reset whenever we transition out of the
  // dev-preview chrome back into the editor (e.g., browser back button,
  // direct URL, or any path that bypasses the Edit button's onClick).
  let prevInEditDevPreview: boolean | null = null;
  $: {
    const inDevPreview = editContext && inEditDevPreview;
    if (
      prevInEditDevPreview === true &&
      !inDevPreview &&
      editContext // only reset when staying inside the editor
    ) {
      sidebarActions.closeChat();
      if ($viewAsUserStore) viewAsUserStore.set(null);
    }
    prevInEditDevPreview = inDevPreview;
  }

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
        const slug = isMetricsExplorer ? "explore" : "canvas";
        // In the dev-preview chrome, swapping resources via the breadcrumb
        // dropdown should keep the user inside `/-/edit/...`. Set an
        // explicit `href` that pins the edit prefix; otherwise fall back
        // to the production-style `${section}/${name}` path linkMaker
        // builds from `section` alone.
        const href = inEditDevPreview
          ? `/${organization}/${project}${branchPathPrefix(activeBranch)}/-/edit/${slug}/${name}`
          : undefined;
        return map.set(name.toLowerCase(), {
          label:
            (isMetricsExplorer
              ? resource?.explore?.spec?.displayName
              : resource?.canvas?.spec?.displayName) || name,
          section: slug,
          href,
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

  // In the dev-preview chrome, the project breadcrumb should keep the user
  // inside the editor session — link each project option to its
  // `${branch}/-/edit/dashboards` instead of the production project root.
  $: projectOptions = (() => {
    const raw = $projectPathsQuery.data ?? new Map<string, PathOption>();
    if (!inEditDevPreview) return raw;
    const branchPart = branchPathPrefix(activeBranch);
    return new Map(
      [...raw].map(([key, option]) => [
        key,
        {
          ...option,
          href: `/${organization}/${key}${branchPart}/-/edit/dashboards`,
        },
      ]),
    );
  })();

  $: pathParts = [
    { options: $orgPathsQuery.data ?? new Map() },
    { options: projectOptions },
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

  // On `/-/edit/{explore,canvas}/[name]`, the route param is `name`, not
  // `dashboard`; fall back so the breadcrumb still surfaces the resource.
  $: currentDashboard = dashboard || (editPreviewKind ? name : undefined);
  $: currentPath = [organization, project, currentDashboard, report || alert];
</script>

<Header
  borderBottom={!onProjectPage && !inEditDevPreview}
  tinted={inEditDevPreview}
>
  <HeaderLogo href={rillLogoHref} logoUrl={organizationLogoUrl} />
  {#if editContext}
    <Tag text={inEditDevPreview ? "Preview" : "Developer"} color="theme" />
  {/if}
  {#if onPublicURLPage}
    <PageTitle title={publicURLDashboardTitle} />
  {:else if organization}
    <Breadcrumbs {pathParts} {currentPath}>
      <svelte:fragment slot="after-project">
        {#if editContext && activeBranch}
          <Slash />
          <li
            class="flex items-center gap-x-1 px-2 text-fg-primary text-xs font-medium"
            title={activeBranch}
          >
            <GitBranch size="14" />
            <span class="truncate max-w-[200px]">{activeBranch}</span>
          </li>
        {:else if !onPublicURLPage && projectPermissions?.manageDev}
          <BranchSelector {organization} {project} {primaryBranch} />
        {/if}
      </svelte:fragment>
    </Breadcrumbs>
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if editContext && inEditDevPreview}
      {#if projectPermissions?.manageDev}
        <CloudViewAsButton />
      {/if}
      <PreviewModeToggleButton
        mode="Edit"
        href={editBackHref}
        onPreviewClick={resetOnPreviewSwap}
      />
      <EditActions
        {organization}
        {project}
        branch={activeBranch ?? ""}
        {primaryBranch}
      />
    {:else if editContext}
      {#if $developerChat}
        <ChatToggle />
      {/if}
      {#if $viewAsUserStore}
        <ViewAsUserChip />
      {/if}
      <PreviewModeToggleButton
        mode="Preview"
        href={editPreviewHref}
        showViewAs={projectPermissions?.manageProject ?? false}
        bind:dropdownOpen={editorViewAsOpen}
        onPreviewClick={resetOnPreviewSwap}
      >
        <svelte:fragment slot="dropdown">
          <ViewAsUserPopover
            {organization}
            {project}
            onSelectUser={() => {
              editorViewAsOpen = false;
              // Preserve the impersonation across the editor → cloud-prod
              // transition this navigation triggers.
              skipNextPlatformReset();
              void goto(editPreviewHref);
            }}
          />
        </svelte:fragment>
      </PreviewModeToggleButton>
      <EditActions
        {organization}
        {project}
        branch={activeBranch ?? ""}
        {primaryBranch}
      />
    {:else}
      {#if projectPermissions?.manageDev}
        <CloudViewAsButton />
      {/if}
      {#if $cloudEditing && onProjectPage && projectPermissions.manageDev}
        <EditButton {organization} {project} {activeBranch} {primaryBranch} />
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
              <EditButton
                {organization}
                {project}
                {activeBranch}
                {primaryBranch}
              />
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
        <EditButton {organization} {project} {activeBranch} {primaryBranch} />
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
