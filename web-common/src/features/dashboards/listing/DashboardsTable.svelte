<script lang="ts">
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceTable from "@rilldata/web-common/features/resources/ResourceTable.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import NameCell from "@rilldata/web-common/features/resources/cells/NameCell.svelte";
  import RelativeTimeCell from "@rilldata/web-common/features/resources/cells/RelativeTimeCell.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    applyTableFilters,
    TableToolbar,
  } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { renderComponent, type ColumnDef } from "tanstack-table-8-svelte-5";
  import DashboardActionsCell from "./DashboardActionsCell.svelte";
  import DashboardStatusCell from "./DashboardStatusCell.svelte";
  import DashboardTypeCell from "./DashboardTypeCell.svelte";

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

  function getCreatedOn(r: V1Resource): string | undefined {
    return r.meta?.createdOn;
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

  $: processedData = applyTableFilters({
    data: data ?? [],
    searchText,
    matchesSearch,
    filterPredicates: [(r) => matchesType(r, selectedTypes)],
    sortDirection,
    getSortKey: getCreatedOn,
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

  function handleFilterChange(key: string, selected: string | string[]) {
    if (key !== "type") return;
    selectedTypes = Array.isArray(selected) ? selected : [selected];
  }

  function clearFilters() {
    selectedTypes = [];
    searchText = "";
  }

  $: getRowHref = (row: unknown) => {
    const r = row as V1Resource;
    const name = r.meta?.name?.name ?? "";
    const isMetricsExplorer = !!r.explore;
    return getHref(name, isMetricsExplorer);
  };

  const columns: ColumnDef<V1Resource, string>[] = [
    {
      id: "type",
      header: "Type",
      accessorFn: (row) => (row.explore ? "Explore" : "Canvas"),
      cell: (info) =>
        renderComponent(DashboardTypeCell, { resource: info.row.original }),
      meta: { width: "85px" },
    },
    {
      id: "name",
      header: "Dashboard name",
      accessorFn: (row) =>
        row.explore?.spec?.displayName ??
        row.canvas?.spec?.displayName ??
        row.meta?.name?.name ??
        "",
      cell: (info) =>
        renderComponent(NameCell, { name: info.getValue() as string }),
    },
    {
      id: "status",
      header: "Status",
      accessorFn: (row) =>
        row.meta?.reconcileError ? "Error" : (row.meta?.reconcileStatus ?? ""),
      cell: (info) =>
        renderComponent(DashboardStatusCell, { resource: info.row.original }),
      meta: { width: "100px" },
    },
    // TODO(#9283): add the Tags column once resource-level tags are exposed
    // by the API. https://github.com/rilldata/rill/issues/9283
    {
      id: "lastRefreshed",
      header: "Last refreshed",
      // Prefer `state.dataRefreshedOn` (when the underlying metrics data was
      // last refreshed). It can be empty when the data refresh time is
      // unknown (e.g. externally-managed tables), so fall back to
      // `meta.stateUpdatedOn` (when the resource state was last updated).
      accessorFn: (row) =>
        row.explore?.state?.dataRefreshedOn ??
        row.canvas?.state?.dataRefreshedOn ??
        row.meta?.stateUpdatedOn ??
        "",
      cell: (info) =>
        renderComponent(RelativeTimeCell, {
          value: info.getValue() as string,
        }),
      meta: { width: "140px" },
    },
    {
      id: "actions",
      header: "",
      cell: (info) => {
        const r = info.row.original;
        const name = r.meta?.name?.name ?? "";
        const isMetricsExplorer = !!r.explore;
        const title =
          r.explore?.spec?.displayName ?? r.canvas?.spec?.displayName ?? name;
        return renderComponent(DashboardActionsCell, {
          dashboardHref: getHref(name, isMetricsExplorer),
          title,
          isMetricsExplorer,
        });
      },
      enableSorting: false,
      meta: { width: "48px", align: "right" },
    },
  ];
</script>

{#if isLoading}
  <div class="m-auto mt-20">
    <DelayedSpinner {isLoading} size="24px" />
  </div>
{:else if isError}
  <ResourceError kind="dashboard" {error} />
{:else}
  <div class="flex flex-col w-full gap-y-3">
    <ResourceTable
      kind="dashboard"
      data={displayData}
      {columns}
      toolbar={resolvedToolbar}
      isFiltered={searchText !== "" || selectedTypes.length > 0}
      {getRowHref}
    >
      <TableToolbar
        slot="toolbar"
        bind:searchText
        {filterGroups}
        onFilterChange={handleFilterChange}
        onClearAllFilters={clearFilters}
        bind:sortDirection
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
    </ResourceTable>
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
