<script lang="ts">
  import AlertHistoryTableCompositeCell from "@rilldata/web-admin/features/alerts/history/AlertHistoryTableCompositeCell.svelte";
  import NoAlertRunsYet from "@rilldata/web-admin/features/alerts/history/NoAlertRunsYet.svelte";
  import { useAlert } from "@rilldata/web-admin/features/alerts/selectors";
  import ReportHistoryTableHeader from "@rilldata/web-admin/features/scheduled-reports/history/ReportHistoryTableHeader.svelte";
  import type { V1AlertExecution } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/table-core/src/types";
  import Table from "../../../components/table/Table.svelte";

  export let alert: string;

  $: alertQuery = useAlert($runtime.instanceId, alert);

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns: ColumnDef<V1AlertExecution>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(AlertHistoryTableCompositeCell, {
          alertTime: info.row.original.executionTime,
          timeZone:
            $alertQuery.data.resource.alert.spec.refreshSchedule.timeZone,
          result: info.row.original.result,
        }),
    },
  ];
</script>

<div class="flex flex-col gap-y-4 w-full">
  <h1 class="text-gray-600 text-lg font-bold">Recent history</h1>
  {#if $alertQuery.error}
    <div class="text-red-500">
      {$alertQuery.error.message}
    </div>
  {:else if $alertQuery.isLoading}
    <div class="text-gray-500">Loading...</div>
  {:else if !$alertQuery.data?.resource.alert.state.executionHistory.length}
    <NoAlertRunsYet />
  {:else}
    <Table
      {columns}
      data={$alertQuery.data?.resource.alert.state.executionHistory}
      maxWidthOverride="max-w-[960px]"
    >
      <ReportHistoryTableHeader
        slot="header"
        maxWidthOverride="max-w-[960px]"
      />
    </Table>
  {/if}
</div>
