<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender } from "@tanstack/svelte-table";
  import Table from "../../../components/table/Table.svelte";
  import DashboardsError from "./DashboardsError.svelte";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import DashboardsTableEmpty from "./DashboardsTableEmpty.svelte";
  import DashboardsTableHeader from "./DashboardsTableHeader.svelte";
  import NoDashboardsCTA from "./NoDashboardsCTA.svelte";
  import { DashboardResource, useDashboardsV2 } from "./selectors";

  export let organization: string;
  export let project: string;

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
      cell: ({ row }) =>
        flexRender(DashboardsTableCompositeCell, {
          organization: organization,
          project: project,
          name: row.original.resource.meta.name.name,
          title: row.original.resource.metricsView.spec.title,
          lastRefreshed: row.original.refreshedOn,
          description: row.original.resource.metricsView.spec.description,
          error: row.original.resource.meta.reconcileError,
        }),
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
    <Table data={$dashboards?.data} {columns} {columnVisibility}>
      <DashboardsTableHeader slot="header" />
      <DashboardsTableEmpty slot="empty" />
    </Table>
  {/if}
{/if}
