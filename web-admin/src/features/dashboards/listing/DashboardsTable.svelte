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

  interface DashboardResource {
    resource: V1Resource;
    refreshedOn: string;
  }

  const columns: ColumnDef<DashboardResource, string>[] = [
    {
      id: "monocolumn",
      // The accessorFn enables sorting and filtering. It contains all the data that will be filtered.
      accessorFn: (row: DashboardResource) =>
        row.resource.metricsView.spec.title +
        row.resource.metricsView.spec.description +
        row.refreshedOn.toString(),
      cell: (info) =>
        flexRender(DashboardsTableInfoCell, {
          organization: organization,
          project: project,
          name: info.row.original.resource.meta.name.name,
          title: info.row.original.resource.metricsView.spec.title,
          lastRefreshed: info.row.original.refreshedOn,
          description: info.row.original.resource.metricsView.spec.description,
          error: info.row.original.resource.meta.reconcileError,
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
  {#if $dashboards.data.length === 0}
    <NoDashboardsCTA />
  {:else}
    <Table dataTypeName="dashboard" {columns} data={$dashboards?.data}>
      <DashboardsTableHeader slot="header" />
    </Table>
  {/if}
{/if}
