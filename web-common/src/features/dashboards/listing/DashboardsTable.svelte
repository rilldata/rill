<script lang="ts">
  import ResourceError from "@rilldata/web-common/features/resources/ResourceError.svelte";
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { flexRender } from "@tanstack/svelte-table";
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
  /** Whether to show the search toolbar. Defaults to !isPreview. */
  export let toolbar: boolean | undefined = undefined;

  $: resolvedToolbar = toolbar ?? !isPreview;

  $: displayData = isPreview
    ? (data?.slice(0, previewLimit) ?? [])
    : (data ?? []);
  $: hasMoreDashboards = isPreview && data && data.length > previewLimit;

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

        return flexRender(DashboardsTableCompositeCell, {
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

  const initialSorting = [{ id: "name", desc: false }];
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
      {initialSorting}
      toolbar={resolvedToolbar}
    >
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
