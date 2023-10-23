<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/table-core/src/types";
  import Table from "../../../components/table/Table.svelte";
  import { useDashboardsV2 } from "./dashboards";
  import DashboardsError from "./DashboardsError.svelte";
  import DashboardsTableCompositeCell from "./DashboardsTableCompositeCell.svelte";
  import DashboardsTableHeader from "./DashboardsTableHeader.svelte";
  import NoDashboardsCTA from "./NoDashboardsCTA.svelte";

  export let organization: string;
  export let project: string;

  $: dashboards = useDashboardsV2($runtime.instanceId);

  interface DashboardResource {
    resource: V1Resource;
    refreshedOn: string;
  }

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns: ColumnDef<DashboardResource, string>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(DashboardsTableCompositeCell, {
          organization: organization,
          project: project,
          name: info.row.original.resource.meta.name.name,
          title: info.row.original.resource.metricsView.spec.title,
          lastRefreshed: info.row.original.refreshedOn,
          description: info.row.original.resource.metricsView.spec.description,
          error: info.row.original.resource.meta.reconcileError,
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
      dataTypeName="dashboard"
      {columns}
      data={$dashboards?.data}
      columnVisibility={{
        title: false,
        name: false,
        lastRefreshed: false,
        description: false,
      }}
    >
      <DashboardsTableHeader slot="header" />
    </Table>
  {/if}
{/if}
