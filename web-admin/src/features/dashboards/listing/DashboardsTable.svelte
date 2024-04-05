<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { type Row, flexRender } from "@tanstack/svelte-table";
  import { createEventDispatcher } from "svelte";
  import Table from "../../../components/table/Table.svelte";
  import DashboardsError from "./DashboardsError.svelte";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import DashboardsTableEmpty from "./DashboardsTableEmpty.svelte";
  import DashboardsTableHeader from "./DashboardsTableHeader.svelte";
  import NoDashboardsCTA from "./NoDashboardsCTA.svelte";
  import { type DashboardResource, useDashboardsV2 } from "./selectors";

  export let isEmbedded = false;

  $: dashboards = useDashboardsV2($runtime.instanceId);

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
        const dashboard = row.original as DashboardResource;
        return flexRender(DashboardsTableCompositeCell, {
          name: dashboard.resource.meta.name.name,
          title: dashboard.resource.metricsView.spec.title,
          lastRefreshed: dashboard.refreshedOn,
          description: dashboard.resource.metricsView.spec.description,
          error: dashboard.resource.meta.reconcileError,
          isEmbedded,
        });
      },
    },
    {
      id: "title",
      accessorFn: (row: DashboardResource) =>
        row.resource.metricsView.spec.title,
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
      accessorFn: (row: DashboardResource) =>
        row.resource.metricsView.spec.description,
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
    dispatch("select-dashboard", e.detail.original.resource.meta.name.name);
  }
</script>

{#if $dashboards.isLoading}
  <div class="m-auto mt-20">
    <Spinner status={EntityStatus.Running} size="24px" />
  </div>
{:else if $dashboards.isError}
  <DashboardsError />
{:else if $dashboards.isSuccess}
  {#if $dashboards.data.length === 0}
    <NoDashboardsCTA />
  {:else}
    <Table
      data={$dashboards?.data}
      {columns}
      {columnVisibility}
      on:click-row={handleClickRow}
    >
      <DashboardsTableHeader slot="header" />
      <DashboardsTableEmpty slot="empty" />
    </Table>
  {/if}
{/if}
