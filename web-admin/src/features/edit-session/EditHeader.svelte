<script lang="ts">
  import { page } from "$app/stores";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import { isEditPreviewRoute } from "@rilldata/web-admin/features/edit-session/edit-route-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import InputWithConfirm from "@rilldata/web-common/components/forms/InputWithConfirm.svelte";
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
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
  import {
    BellPlusIcon,
    BookmarkIcon,
    GitBranchIcon,
    Share2,
    Sparkles,
  } from "lucide-svelte";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import {
    createAdminServiceGetCurrentUser,
    type V1ProjectPermissions,
  } from "../../client";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import LastRefreshedDate from "../dashboards/listing/LastRefreshedDate.svelte";
  import ViewAsUserChip from "../view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../view-as-user/viewAsUserStore";

  const cloudCta = "Publish project to use this feature";

  type CloudFeature = {
    label: string;
    icon?: typeof HomeBookmark;
    compact?: boolean;
    square?: boolean;
  };

  // All disabled cloud-feature buttons are square 28x28 for visual
  // consistency. Share stays as a plain text button since "Share"
  // doesn't fit a 28x28 box.
  const exploreCloudFeatures: CloudFeature[] = [
    { label: "AI", compact: true, square: true },
    { label: "Home bookmark", icon: HomeBookmark, compact: true, square: true },
    { label: "Bookmark", icon: BookmarkIcon, compact: true, square: true },
    { label: "Alert", icon: BellPlusIcon, compact: true, square: true },
    { label: "Share" },
  ];

  const canvasCloudFeatures: CloudFeature[] = [
    { label: "AI", compact: true, square: true },
    { label: "Home bookmark", icon: HomeBookmark, compact: true, square: true },
    { label: "Bookmark", icon: BookmarkIcon, compact: true, square: true },
    { label: "Share" },
  ];

  export let organization: string;
  export let project: string;
  export let projectPermissions: V1ProjectPermissions;

  const user = createAdminServiceGetCurrentUser();
  const runtimeClient = useRuntimeClient();
  const { developerChat, dimensionSearch } = featureFlags;

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
  <nav class="flex gap-x-2 items-center">
    <ol class="flex flex-row items-center">
      <li class="flex items-center gap-x-2 px-2">
        <span class="text-fg-muted">{project}</span>
      </li>
      {#if activeBranch}
        <Slash />
        <li class="flex items-center gap-x-2 px-2">
          <span
            class="text-fg-primary font-medium flex flex-row items-center gap-x-2"
          >
            <GitBranchIcon size="14" class="text-fg-primary" />
            {activeBranch.length > 12
              ? activeBranch.slice(0, 11) + "…"
              : activeBranch}
          </span>
        </li>
      {/if}
    </ol>
  </nav>

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
    <nav class="flex gap-x-2 items-center shrink-0" data-edit-home="preview">
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
    <div data-edit-home="developer">
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
          {#each exploreCloudFeatures as { label, icon, compact, square } (label)}
            <Tooltip distance={8}>
              <Button type="secondary" {compact} {square} disabled {label}>
                {#if icon}
                  <svelte:component
                    this={icon}
                    size="16px"
                    class="flex-none"
                    className="flex-none"
                  />
                {:else}
                  {label}
                {/if}
              </Button>
              <TooltipContent slot="tooltip-content" maxWidth="240px">
                <span class="text-xs">{cloudCta}</span>
              </TooltipContent>
            </Tooltip>
          {/each}
        </StateManagersProvider>
      {/key}
    {:else if onEditCanvas}
      {#each canvasCloudFeatures as { label, icon, compact, square } (label)}
        <Tooltip distance={8}>
          <Button type="secondary" {compact} {square} disabled {label}>
            {#if icon}
              <svelte:component
                this={icon}
                size="16px"
                class="flex-none"
                className="flex-none"
              />
            {:else}
              {label}
            {/if}
          </Button>
          <TooltipContent slot="tooltip-content" maxWidth="240px">
            <span class="text-xs">{cloudCta}</span>
          </TooltipContent>
        </Tooltip>
      {/each}
    {:else if !previewMode && $developerChat}
      <ChatToggle class="!bg-surface-base" />
    {/if}
  </div>
</div>
