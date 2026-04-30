<script lang="ts">
  import { page } from "$app/stores";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import { isEditPreviewRoute } from "@rilldata/web-admin/features/edit-session/edit-route-utils";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import GlobalDimensionSearch from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearch.svelte";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExplore } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { GitBranchIcon } from "lucide-svelte";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import {
    createAdminServiceGetCurrentUser,
    type V1ProjectPermissions,
  } from "../../client";
  import CreateAlert from "../alerts/CreateAlert.svelte";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import CanvasBookmarks from "../bookmarks/CanvasBookmarks.svelte";
  import ExploreBookmarks from "../bookmarks/ExploreBookmarks.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import ShareDashboardPopover from "../dashboards/share/ShareDashboardPopover.svelte";
  import ViewAsUserChip from "../view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../view-as-user/viewAsUserStore";

  export let organization: string;
  export let project: string;
  export let projectPermissions: V1ProjectPermissions;

  const user = createAdminServiceGetCurrentUser();
  const runtimeClient = useRuntimeClient();
  const {
    alerts: alertsFlag,
    dashboardChat,
    developerChat,
    dimensionSearch,
  } = featureFlags;

  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: previewMode = isEditPreviewRoute($page.url.pathname);
  $: editPrefix = `/${organization}/${project}${branchPathPrefix(activeBranch)}/-/edit`;
  $: previewHomeHref = `${editPrefix}/dashboards`;

  // Top header: project name as plain (non-clickable) text. Display name
  // comes from the runtime instance metadata when available.
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {});
  $: projectDisplayName =
    $instanceQuery.data?.instance?.projectDisplayName || project;

  // Secondary header breadcrumb: Home / dashboard-name (with dropdown to
  // switch between dashboards on this branch).
  $: dashboardName = $page.params.name;
  $: onEditExplore = $page.url.pathname.includes("/-/edit/explore/");
  $: onEditCanvas = $page.url.pathname.includes("/-/edit/canvas/");
  $: onDashboardPage = onEditExplore || onEditCanvas;

  $: visualizationsQuery = useDashboards(runtimeClient);
  $: visualizations = $visualizationsQuery.data ?? [];

  // Build dashboard options with explicit hrefs so the breadcrumb dropdown
  // navigates within the edit surface, not the production routes.
  $: visualizationPaths = {
    options: [...visualizations]
      .sort((a, b) => {
        const aIsCanvas = !!a?.canvas;
        const bIsCanvas = !!b?.canvas;
        if (aIsCanvas !== bIsCanvas) return aIsCanvas ? -1 : 1;
        return a.meta.name.name.localeCompare(b.meta.name.name);
      })
      .reduce((map, resource) => {
        const name = resource.meta.name.name;
        const isMetricsExplorer = !!resource?.explore;
        const section = isMetricsExplorer ? "explore" : "canvas";
        const label =
          (isMetricsExplorer
            ? resource?.explore?.spec?.displayName
            : resource?.canvas?.spec?.displayName) || name;
        return map.set(name.toLowerCase(), {
          label,
          href: `${editPrefix}/${section}/${name}`,
          resourceKind: isMetricsExplorer
            ? ResourceKind.Explore
            : ResourceKind.Canvas,
        });
      }, new Map<string, PathOption>()),
  };

  $: exploreQuery = useExplore(runtimeClient, dashboardName, {
    enabled: !!runtimeClient.instanceId && !!dashboardName && onEditExplore,
  });
  $: exploreSpec = $exploreQuery.data?.explore?.explore?.state?.validSpec;
  $: hasUserAccess = $user.isSuccess && !!$user.data?.user;

  // Editable project title (developer side). Mirrors the rill.yaml
  // display_name handling from web-local's ApplicationHeader.
  $: projectTitleQuery = useProjectTitle(runtimeClient);
  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";
  $: ({ unsavedFiles } = fileArtifacts);
  $: ({ size: unsavedFileCount } = $unsavedFiles);

  async function submitTitleChange(editedTitle: string) {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");
    let content = get(artifact.editorContent);
    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.remoteContent);
      if (!content) return;
    }
    const parsed = parseDocument(content);
    parsed.set("display_name", editedTitle);
    artifact.updateEditorContent(parsed.toString(), true);
    await artifact.saveLocalContent();
  }
</script>

<Header borderBottom tinted>
  <HeaderLogo href={`/${organization}/${project}`} />
  {#if activeBranch}
    <Tag
      text={previewMode ? "Preview" : "Developer"}
      color="gray"
      class="!bg-surface-base"
    />
  {/if}
  <div class="flex items-center gap-x-2 px-3">
    <span class="text-fg-muted text-sm">{projectDisplayName}</span>
    {#if activeBranch}
      <Slash />
      <div class="flex items-center gap-x-1.5">
        <GitBranchIcon size="14" class="text-fg-primary" />
        <span class="text-fg-primary text-sm font-medium">
          {activeBranch.length > 12
            ? activeBranch.slice(0, 11) + "…"
            : activeBranch}
        </span>
      </div>
    {/if}
  </div>

  <div class="flex gap-x-2 items-center ml-auto">
    {#if previewMode && $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    <EditActions {organization} {project} branch={activeBranch ?? ""} />
    {#if $user.isSuccess && $user.data?.user}
      <AvatarButton {projectPermissions} />
    {/if}
  </div>
</Header>

<div
  class="bg-surface-base flex items-center h-10 px-3 gap-x-2 border-b border-border"
>
  {#if previewMode}
    <nav class="flex gap-x-2 items-center">
      <ol class="flex flex-row items-center">
        <li class="flex items-center gap-x-2 px-2">
          <a
            href={previewHomeHref}
            class="text-fg-muted hover:text-fg-secondary flex flex-row items-center gap-x-2"
          >
            <span>Home</span>
          </a>
        </li>
        {#if onDashboardPage && dashboardName}
          <Slash />
          <BreadcrumbItem
            depth={0}
            pathOptions={visualizationPaths}
            current={dashboardName.toLowerCase()}
            isCurrentPage={true}
          />
        {/if}
      </ol>
    </nav>
  {:else}
    <div class="px-2">
      <InputWithConfirm
        size="md"
        bumpDown
        type="Project"
        textClass="font-medium"
        value={projectTitle}
        onConfirm={submitTitleChange}
        showIndicator={unsavedFileCount > 0}
      />
    </div>
  {/if}

  <div class="ml-auto flex gap-x-2 items-center">
    {#if onEditExplore && exploreSpec}
      {#key dashboardName}
        <StateManagersProvider
          metricsViewName={exploreSpec.metricsView}
          exploreName={dashboardName}
          let:ready
        >
          <LastRefreshedDate dashboard={dashboardName} />
          {#if $dimensionSearch && ready}
            <GlobalDimensionSearch />
          {/if}
          {#if $dashboardChat}
            <ChatToggle />
          {/if}
          {#if hasUserAccess}
            <ExploreBookmarks
              {organization}
              {project}
              metricsViewName={exploreSpec.metricsView}
              exploreName={dashboardName}
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
    {:else if onEditCanvas}
      {#if $dashboardChat}
        <ChatToggle />
      {/if}
      {#if hasUserAccess}
        <CanvasBookmarks {organization} {project} canvasName={dashboardName} />
        <ShareDashboardPopover
          createMagicAuthTokens={projectPermissions.createMagicAuthTokens}
        />
      {/if}
    {:else if $developerChat}
      <ChatToggle />
    {/if}
  </div>
</div>
