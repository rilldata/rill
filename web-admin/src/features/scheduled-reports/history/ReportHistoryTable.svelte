<script lang="ts">
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import type { V1ReportExecution } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import { useReport } from "../selectors";
  import NoRunsYet from "./NoRunsYet.svelte";
  import ReportHistoryTableCompositeCell from "./ReportHistoryTableCompositeCell.svelte";

  export let report: string;

  $: ({ instanceId } = $runtime);

  $: reportQuery = useReport(instanceId, report);

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
          timeZone:
            $reportQuery.data.resource.report.spec.refreshSchedule.timeZone,
          adhoc: info.row.original.adhoc,
          errorMessage: info.row.original.errorMessage,
        }),
    },
  ];
</script>

<div class="flex flex-col gap-y-4 w-full">
  <div class="flex flex-col gap-y-1">
    <h1 class="text-fg-secondary text-lg font-bold">Recent history</h1>
    <p class="text-fg-secondary text-sm">Showing up to 10 most recent runs</p>
  </div>
  {#if $reportQuery.error}
    <div class="text-red-500">
      {$reportQuery.error.message}
    </div>
  {:else if $reportQuery.isLoading}
    <div class="text-fg-secondary">Loading...</div>
  {:else if !$reportQuery.data?.resource.report.state.executionHistory.length}
    <NoRunsYet />
  {:else}
    <ResourceList
      kind="report"
      {columns}
      data={$reportQuery.data?.resource.report.state.executionHistory}
      toolbar={false}
      fixedRowHeight={false}
    />
  {/if}
</div>
