<script lang="ts">
  import { page } from "$app/stores";
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import TagFilterBanner from "@rilldata/web-common/components/menu/TagFilterBanner.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import DashboardsTagSidebar from "./DashboardsTagSidebar.svelte";
  import { useDashboards, useIsInitialBuild } from "./selectors";
  import {
    dashboardsTagSidebarWidth,
    MAX_TAG_SIDEBAR_WIDTH,
    MIN_TAG_SIDEBAR_WIDTH,
    DEFAULT_TAG_SIDEBAR_WIDTH,
  } from "./tag-sidebar-store";

  export let isEmbedded = false;
  export let isPreview = false;
  export let previewLimit = 5;

  type DashboardCellContext = {
    row: {
      original: V1Resource;
    };
  };

  // The tag selected in the left sidebar, or null when showing all dashboards.
  let selectedTag: string | null = null;

  // Resizable divider between the tag sidebar and the dashboards list. We use a
  // self-contained handle (positioned with top/bottom insets) rather than the
  // shared Resizer, whose `height: 100%` only resolves inside a fixed-height
  // ancestor; this listing page is content-height.
  let resizing = false;

  function startResize(e: MouseEvent) {
    e.preventDefault();
    resizing = true;
    const startX = e.clientX;
    const startWidth = $dashboardsTagSidebarWidth;

    function onMove(ev: MouseEvent) {
      const next = Math.min(
        MAX_TAG_SIDEBAR_WIDTH,
        Math.max(MIN_TAG_SIDEBAR_WIDTH, startWidth + ev.clientX - startX),
      );
      dashboardsTagSidebarWidth.set(next);
    }

    function onUp() {
      resizing = false;
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
    }

    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);
  }

  const runtimeClient = useRuntimeClient();
  $: ({
    params: { organization, project },
  } = $page);

  $: dashboards = useDashboards(runtimeClient);
  $: ({
    data: dashboardsData,
    isLoading,
    isError,
    isSuccess,
    error,
  } = $dashboards);

  $: initialBuild = useIsInitialBuild(runtimeClient);
  $: isBuilding = $initialBuild.data === true;

  // Show the tag sidebar only when at least one dashboard has tags, and never in
  // the embedded surface.
  $: hasTags = (dashboardsData ?? []).some((r) =>
    (r.meta?.tags ?? []).some((rawTag) => rawTag.trim() !== ""),
  );
  $: showTagSidebar = hasTags && !isEmbedded;

  // Drop the selection if the active tag disappears, even after the sidebar is
  // hidden because all tags were removed.
  $: {
    const activeTag = selectedTag;
    if (
      activeTag &&
      dashboardsData &&
      !dashboardsData.some((r) => resourceHasTag(r, activeTag))
    ) {
      selectedTag = null;
    }
  }

  // Filter by the selected tag before applying the preview limit, so the tag
  // filter and the top search box compose.
  $: tagFilteredData = filterByTag(dashboardsData ?? [], selectedTag);

  function filterByTag(resources: V1Resource[], tag: string | null) {
    if (!tag) return resources;
    return resources.filter((r) => resourceHasTag(r, tag));
  }

  function resourceHasTag(resource: V1Resource, tag: string) {
    return (resource.meta?.tags ?? []).some((rawTag) => rawTag.trim() === tag);
  }

  $: displayData = isPreview
    ? tagFilteredData.slice(0, previewLimit)
    : tagFilteredData;
  $: hasMoreDashboards = isPreview && tagFilteredData.length > previewLimit;

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   *
   * Note: TypeScript error prevents using `ColumnDef<DashboardResource, string>[]`.
   * Relevant issues:
   * - https://github.com/TanStack/table/issues/4241
   * - https://github.com/TanStack/table/issues/4302
   */
  const columns = [
    {
      id: "composite",
      cell: (ctx): unknown => {
        const { row } = ctx as DashboardCellContext;
        const resource = row.original;
        const name = resource.meta?.name?.name ?? "";

        // If not a Metrics Explorer, it's a Custom Dashboard.
        const isMetricsExplorer = !!resource.explore?.spec;
        const title = isMetricsExplorer
          ? (resource.explore?.spec?.displayName ?? "")
          : (resource.canvas?.spec?.displayName ?? "");
        const description = isMetricsExplorer
          ? (resource.explore?.spec?.description ?? "")
          : "";
        const refreshedOn = isMetricsExplorer
          ? resource.explore?.state?.dataRefreshedOn
          : resource.canvas?.state?.dataRefreshedOn;

        return renderComponent(DashboardsTableCompositeCell, {
          name,
          title,
          lastRefreshed: refreshedOn ?? "",
          description,
          error: resource.meta?.reconcileError ?? "",
          isMetricsExplorer,
          isEmbedded,
          organization,
          project,
        });
      },
    },
    {
      id: "title",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row.explore?.spec;
        return isMetricsExplorer
          ? (row.explore?.spec?.displayName ?? "")
          : (row.canvas?.spec?.displayName ?? "");
      },
    },
    {
      id: "name",
      accessorFn: (row: V1Resource) => row.meta?.name?.name ?? "",
    },
    {
      id: "lastRefreshed",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row.explore?.spec;
        return isMetricsExplorer
          ? row.explore?.state?.dataRefreshedOn
          : row.canvas?.state?.dataRefreshedOn;
      },
    },
    {
      id: "description",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row.explore?.spec;
        return isMetricsExplorer ? (row.explore?.spec?.description ?? "") : "";
      },
    },
  ];

  const columnVisibility = {
    title: false,
    name: false,
    lastRefreshed: false,
    description: false,
  };

  const initialSorting = [{ id: "name", desc: false }];
</script>

{#if isLoading || isBuilding}
  <div class="m-auto mt-20">
    <DelayedSpinner isLoading={true} size="24px" />
  </div>
{:else if isError}
  <ResourceError kind="dashboard" {error} />
{:else if isSuccess}
  <div class="flex w-full gap-x-4 items-stretch">
    {#if showTagSidebar}
      <div
        class="relative flex-none"
        style:width="{$dashboardsTagSidebarWidth}px"
      >
        <DashboardsTagSidebar
          resources={dashboardsData ?? []}
          bind:selectedTag
        />
        <button
          type="button"
          class="resize-handle"
          class:resizing
          aria-label="Resize tags sidebar"
          onmousedown={startResize}
          ondblclick={() =>
            dashboardsTagSidebarWidth.set(DEFAULT_TAG_SIDEBAR_WIDTH)}
        ></button>
      </div>
    {/if}
    <div class="flex flex-col flex-1 min-w-0 gap-y-3">
      {#if selectedTag}
        <!-- The banner supplies its own bottom border; the wrapper adds the
             remaining sides to form a single rounded card. -->
        <div class="rounded-lg overflow-hidden border-x border-t">
          <TagFilterBanner
            tagName={selectedTag}
            onClear={() => (selectedTag = null)}
          />
        </div>
      {/if}
      <ResourceList
        kind="dashboard"
        data={displayData}
        {columns}
        {columnVisibility}
        {initialSorting}
        toolbar={!isPreview}
      >
        <ResourceListEmptyState
          slot="empty"
          icon={ExploreIcon}
          message="You don't have any dashboards yet"
        >
          <span slot="action">
            <a
              href="https://docs.rilldata.com/developers/build/dashboards"
              target="_blank"
              rel="noopener noreferrer"
            >
              Create a dashboard</a
            > to get started
          </span>
        </ResourceListEmptyState>
      </ResourceList>
      {#if hasMoreDashboards}
        <div class="pl-4 py-1">
          <a
            href={`/${organization}/${project}/-/dashboards`}
            class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors inline-block"
          >
            See all dashboards →
          </a>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style lang="postcss">
  /* An invisible full-height strip centered on the sidebar's right border.
     Uses inset-y-0 (top/bottom) so it stretches on this content-height page,
     where a percentage height would collapse to zero. The visible 1px rule
     appears on hover/drag. */
  .resize-handle {
    @apply absolute inset-y-0 -right-1 z-50 w-2 cursor-col-resize;
  }

  .resize-handle::after {
    content: "";
    @apply absolute inset-y-0 left-1/2 w-px -translate-x-1/2;
    @apply bg-transparent transition-colors;
  }

  .resize-handle:hover::after,
  .resize-handle.resizing::after {
    @apply bg-primary-300;
  }
</style>
