<script lang="ts">
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import type { V1ReportExecution } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import { useReport } from "../selectors";
  import NoRunsYet from "./NoRunsYet.svelte";
  import ReportHistoryTableCompositeCell from "./ReportHistoryTableCompositeCell.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let report: string;

  const runtimeClient = useRuntimeClient();

  $: reportQuery = useReport(runtimeClient, report);

  /**
   * Table column definitions.
   * - "composite": Renders all dashboard data in a single cell.
   * - Others: Used for sorting and filtering but not displayed.
   */
  const columns: ColumnDef<V1ReportExecution>[] = [
    {
      id: "composite",
      cell: (info) =>
        renderComponent(ReportHistoryTableCompositeCell, {
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
    <h1 class="text-fg-secondary text-lg font-bold">{m.report_recent_history()}</h1>
    <p class="text-fg-secondary text-sm">{m.report_showing_recent_runs()}</p>
  </div>
  {#if $reportQuery.error}
    <div class="text-red-500">
      {$reportQuery.error.message}
    </div>
  {:else if $reportQuery.isLoading}
    <div class="text-fg-secondary">{m.report_loading()}</div>
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
