<script lang="ts">
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";

  // --- Data props (caller provides query results) ---
  export let data: V1Resource[] = [];
  export let isLoading = false;
  export let isError = false;
  export let error: unknown = null;

  // --- Display props ---
  export let isPreview = false;
  export let previewLimit = 5;

  // --- Customization props ---
  /** Function to construct href for a dashboard row */
  export let getHref: (name: string, isMetricsExplorer: boolean) => string;
  /** "See all" link target when in preview mode */
  export let seeAllHref = "";
  /** Whether to show the toolbar. Defaults to !isPreview. */
  export let toolbar: boolean | undefined = undefined;

  $: resolvedToolbar = toolbar ?? !isPreview;

  let searchText = "";
  let selectedTypes: string[] = [];
  let sortDirection: SortDirection = "newest";

  function getTitle(r: V1Resource): string {
    return r.explore
      ? (r.explore.spec?.displayName ?? r.meta?.name?.name ?? "")
      : (r.canvas?.spec?.displayName ?? r.meta?.name?.name ?? "");
  }

  function getDescription(r: V1Resource): string {
    return r.explore?.spec?.description ?? "";
  }

  function getLastRefreshed(r: V1Resource): string | undefined {
    return r.explore
      ? r.explore.state?.dataRefreshedOn
      : r.canvas?.state?.dataRefreshedOn;
  }

  function matchesSearch(r: V1Resource, q: string): boolean {
    if (!q) return true;
    const needle = q.toLowerCase();
    return (
      getTitle(r).toLowerCase().includes(needle) ||
      (r.meta?.name?.name ?? "").toLowerCase().includes(needle) ||
      getDescription(r).toLowerCase().includes(needle)
    );
  }

  function matchesType(r: V1Resource, types: string[]): boolean {
    if (types.length === 0) return true;
    if (types.includes("explore") && r.explore) return true;
    if (types.includes("canvas") && r.canvas) return true;
    return false;
  }

  $: processedData = (data ?? [])
    .filter(
      (r) => matchesType(r, selectedTypes) && matchesSearch(r, searchText),
    )
    .slice()
    .sort((a, b) => {
      const aTime = getLastRefreshed(a) ?? "";
      const bTime = getLastRefreshed(b) ?? "";
      const cmp = aTime < bTime ? -1 : aTime > bTime ? 1 : 0;
      return sortDirection === "newest" ? -cmp : cmp;
    });

  $: displayData = isPreview
    ? processedData.slice(0, previewLimit)
    : processedData;
  $: hasMoreDashboards = isPreview && (data?.length ?? 0) > previewLimit;

  $: filterGroups = [
    {
      label: "Type",
      key: "type",
      options: [
        { value: "explore", label: "Explore" },
        { value: "canvas", label: "Canvas" },
      ],
      selected: selectedTypes,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  function handleFilterChange(key: string, value: string) {
    if (key !== "type") return;
    selectedTypes = selectedTypes.includes(value)
      ? selectedTypes.filter((v) => v !== value)
      : [...selectedTypes, value];
  }

  function clearFilters() {
    selectedTypes = [];
    searchText = "";
  }

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns = [
    {
      id: "composite",
      cell: ({ row }) => {
        const resource = row.original as V1Resource;
        const name = resource.meta?.name?.name ?? "";

        const isMetricsExplorer = !!resource?.explore;
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
          href: getHref(name, isMetricsExplorer),
        });
      },
    },
    {
      id: "title",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row?.explore;
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
        const isMetricsExplorer = !!row?.explore;
        return isMetricsExplorer
          ? row.explore?.state?.dataRefreshedOn
          : row.canvas?.state?.dataRefreshedOn;
      },
    },
    {
      id: "description",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row?.explore;
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
</script>

{#if isLoading}
  <div class="m-auto mt-20">
    <DelayedSpinner {isLoading} size="24px" />
  </div>
{:else if isError}
  <ResourceError kind="dashboard" {error} />
{:else}
  <div class="flex flex-col w-full gap-y-3">
    <ResourceList
      kind="dashboard"
      data={displayData}
      {columns}
      {columnVisibility}
      toolbar={resolvedToolbar}
      isFiltered={searchText !== "" || selectedTypes.length > 0}
    >
      <TableToolbar
        slot="toolbar"
        {searchText}
        onSearchChange={(t) => (searchText = t)}
        {filterGroups}
        onFilterChange={handleFilterChange}
        onClearAllFilters={clearFilters}
        {sortDirection}
        onSortToggle={() =>
          (sortDirection = sortDirection === "newest" ? "oldest" : "newest")}
        disabled={(data?.length ?? 0) === 0}
      />
      <svelte:fragment slot="empty">
        <slot name="empty">
          <ResourceListEmptyState
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
        </slot>
      </svelte:fragment>
    </ResourceList>
    {#if hasMoreDashboards && seeAllHref}
      <div class="pl-4 py-1">
        <a
          href={seeAllHref}
          class="text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors inline-block"
        >
          See all dashboards &rarr;
        </a>
      </div>
    {/if}
  </div>
{/if}
