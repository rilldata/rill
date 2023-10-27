<script lang="ts">
  import type { V1ReportExecution } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/table-core/src/types";
  import Table from "../../../components/table/Table.svelte";
  import { useReport } from "../selectors";
  import ReportHistoryTableCompositeCell from "./ReportHistoryTableCompositeCell.svelte";
  import ReportHistoryTableHeader from "./ReportHistoryTableHeader.svelte";

  export let report: string;

  $: reportQuery = useReport($runtime.instanceId, report);

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns: ColumnDef<V1ReportExecution>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(ReportHistoryTableCompositeCell, {
          reportTime: info.row.original.reportTime,
          errorMessage: info.row.original.errorMessage,
        }),
    },
  ];
</script>

<div class="flex flex-col gap-y-4 w-full">
  <h1 class="text-gray-800 text-base font-medium leading-none">
    Recent history
  </h1>
  {#if $reportQuery.error}
    <div class="text-red-500">
      {$reportQuery.error.message}
    </div>
  {:else if $reportQuery.isLoading}
    <div class="text-gray-500">Loading...</div>
  {:else if !$reportQuery.data?.resource.report.state.executionHistory.length}
    <div class="text-gray-500">No history available.</div>
  {:else}
    <Table
      {columns}
      data={$reportQuery.data?.resource.report.state.executionHistory}
      maxWidthOverride="max-w-[960px]"
    >
      <ReportHistoryTableHeader
        slot="header"
        maxWidthOverride="max-w-[960px]"
      />
    </Table>
  {/if}
</div>
