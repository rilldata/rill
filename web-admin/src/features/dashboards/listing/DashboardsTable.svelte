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
  import DashboardsTableHeader from "./DashboardsTableHeader.svelte";
  import DashboardsTableInfoCell from "./DashboardsTableInfoCell.svelte";
  import NoDashboardsCTA from "./NoDashboardsCTA.svelte";

  export let organization: string;
  export let project: string;

  $: dashboards = useDashboardsV2($runtime.instanceId);

  const columns: ColumnDef<V1Resource>[] = [
    {
      id: "monocolumn",
      // The accessorFn enables sorting and filtering. It contains all the data that will be filtered.
      accessorFn: (row) =>
        row.metricsView.spec.title + row.metricsView.spec.description,
      cell: (info) =>
        flexRender(DashboardsTableInfoCell, {
          organization: organization,
          project: project,
          name: info.row.original.meta.name.name,
          title: info.row.original.metricsView.spec.title,
          // TODO: it'd be more accurate to use the `state.refreshedOn` field of the `meta.refs[0]` resource
          lastRefreshed: new Date(info.row.original.meta.stateUpdatedOn),
          description: info.row.original.metricsView.spec.description,
          error: info.row.original.meta.reconcileError,
        }),
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
  {#if $dashboards.data.resources.length === 0}
    <NoDashboardsCTA />
  {:else}
    <Table
      dataTypeName="dashboard"
      {columns}
      data={$dashboards?.data?.resources}
    >
      <DashboardsTableHeader slot="header" />
    </Table>
  {/if}
{/if}
