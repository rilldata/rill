<script lang="ts">
  import { page } from "$app/stores";
  import ResourceError from "@rilldata/web-admin/features/projects/ResourceError.svelte";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { resourceColorMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender } from "@tanstack/svelte-table";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import { useDashboards } from "./selectors";

  export let isEmbedded = false;
  export let isPreview = false;
  export let previewLimit = 5;

  const exploreColor = resourceColorMapping[ResourceKind.Explore];

  $: ({ instanceId } = $runtime);
  $: ({
    params: { organization, project },
  } = $page);

  $: dashboards = useDashboards(instanceId);
  $: ({
    data: dashboardsData,
    isLoading,
    isError,
    isSuccess,
    error,
  } = $dashboards);

  $: displayData = isPreview
    ? (dashboardsData?.slice(0, previewLimit) ?? [])
    : (dashboardsData ?? []);
  $: hasMoreDashboards =
    isPreview && dashboardsData && dashboardsData.length > previewLimit;

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
      cell: ({ row }) => {
        const resource = row.original as V1Resource;
        const name = resource.meta.name.name;

        // If not a Metrics Explorer, it's a Custom Dashboard.
        const isMetricsExplorer = !!resource?.explore;
        const title = isMetricsExplorer
          ? resource.explore.spec.displayName
          : resource.canvas.spec.displayName;
        const description = isMetricsExplorer
          ? resource.explore.spec.description
          : "";
        const refreshedOn = isMetricsExplorer
          ? resource.explore?.state?.dataRefreshedOn
          : resource.canvas?.state?.dataRefreshedOn;

        return flexRender(DashboardsTableCompositeCell, {
          name,
          title,
          lastRefreshed: refreshedOn,
          description,
          error: resource.meta.reconcileError,
          isMetricsExplorer,
          isEmbedded,
        });
      },
    },
    {
      id: "title",
      accessorFn: (row: V1Resource) => {
        const isMetricsExplorer = !!row?.explore;
        return isMetricsExplorer
          ? row.explore.spec.displayName
          : row.canvas.spec.displayName;
      },
    },
    {
      id: "name",
      accessorFn: (row: V1Resource) => row.meta.name.name,
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
        return isMetricsExplorer ? row.explore.spec.description : "";
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
{:else if isSuccess}
  <div class="flex flex-col w-full gap-y-3">
    <ResourceList
      kind="dashboard"
      data={displayData}
      {columns}
      {columnVisibility}
      toolbar={!isPreview}
    >
      <ResourceListEmptyState
        slot="empty"
        icon={ExploreIcon}
        iconColor={exploreColor}
        message="You don't have any dashboards yet"
      >
        <span slot="action">
          <a
            href="https://docs.rilldata.com/build/dashboards"
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
{/if}
