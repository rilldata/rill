<script lang="ts">
  import ResourceHeader from "@rilldata/web-admin/components/table/ResourceHeader.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender, type Row } from "@tanstack/svelte-table";
  import { createEventDispatcher } from "svelte";
  import Table from "../../../components/table/Table.svelte";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import { useDashboardsV2, type DashboardResource } from "./selectors";
  import NoResourceCTA from "@rilldata/web-admin/features/projects/NoResourceCTA.svelte";
  import ResourceError from "@rilldata/web-admin/features/projects/ResourceError.svelte";

  export let isEmbedded = false;

  $: ({ instanceId } = $runtime);

  $: dashboards = useDashboardsV2(instanceId);

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
        const dashboardResource = row.original as DashboardResource;
        const resource = dashboardResource.resource;
        const refreshedOn = dashboardResource.refreshedOn;
        const name = resource.meta.name.name;

        // If not a Metrics Explorer, it's a Custom Dashboard.
        const isMetricsExplorer = !!resource?.explore;
        const title = isMetricsExplorer
          ? resource.explore.spec.displayName
          : resource.canvas.spec.displayName;
        const description = isMetricsExplorer
          ? resource.explore.spec.description
          : "";

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
      accessorFn: (row: DashboardResource) => {
        const resource = row.resource;
        const isMetricsExplorer = !!resource?.explore;
        return isMetricsExplorer
          ? resource.explore.spec.displayName
          : resource.canvas.spec.displayName;
      },
    },
    {
      id: "name",
      accessorFn: (row: DashboardResource) => row.resource.meta.name.name,
    },
    {
      id: "lastRefreshed",
      accessorFn: (row: DashboardResource) => row.refreshedOn,
    },
    {
      id: "description",
      accessorFn: (row: DashboardResource) => {
        const resource = row.resource;
        const isMetricsExplorer = !!resource?.explore;
        return isMetricsExplorer ? resource.explore.spec.description : "";
      },
    },
  ];

  const columnVisibility = {
    title: false,
    name: false,
    lastRefreshed: false,
    description: false,
  };

  const dispatch = createEventDispatcher();

  function handleClickRow(e: CustomEvent<Row<DashboardResource>>) {
    dispatch("select-resource", e.detail.original.resource.meta.name);
  }
</script>

{#if $dashboards.isLoading}
  <div class="m-auto mt-20">
    <DelayedSpinner isLoading={$dashboards.isLoading} size="24px" />
  </div>
{:else if $dashboards.isError}
  <ResourceError kind="dashboard" />
{:else if $dashboards.isSuccess}
  {#if !$dashboards.data.length}
    <NoResourceCTA kind="dashboard">
      <svelte:fragment>
        Learn how to deploy a dashboard in our
        <a href="https://docs.rilldata.com/" target="_blank">docs</a>
      </svelte:fragment>
    </NoResourceCTA>
  {:else}
    <Table
      kind="dashboard"
      data={$dashboards?.data}
      {columns}
      {columnVisibility}
      on:click-row={handleClickRow}
    >
      <ResourceHeader kind="dashboard" icon={ExploreIcon} slot="header" />
    </Table>
  {/if}
{/if}
