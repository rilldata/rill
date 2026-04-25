<script lang="ts">
  import ResourceList from "@rilldata/web-common/features/resources/ResourceList.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type {
    FilterGroup,
    SortDirection,
  } from "@rilldata/web-common/components/table-toolbar/types";
  import type { V1ReportExecution } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import { useReport } from "../selectors";
  import NoRunsYet from "./NoRunsYet.svelte";
  import ReportHistoryTableCompositeCell from "./ReportHistoryTableCompositeCell.svelte";

  export let report: string;

  const runtimeClient = useRuntimeClient();

  $: reportQuery = useReport(runtimeClient, report);

  let selectedResults: string[] = [];
  let sortDirection: SortDirection = "newest";

  $: history =
    $reportQuery.data?.resource?.report?.state?.executionHistory ?? [];

  function getResult(e: V1ReportExecution): "ok" | "error" {
    return e.errorMessage ? "error" : "ok";
  }

  $: processedHistory = history
    .filter(
      (e) =>
        selectedResults.length === 0 || selectedResults.includes(getResult(e)),
    )
    .slice()
    .sort((a, b) => {
      const aTime = a.reportTime ?? "";
      const bTime = b.reportTime ?? "";
      const cmp = aTime < bTime ? -1 : aTime > bTime ? 1 : 0;
      return sortDirection === "newest" ? -cmp : cmp;
    });

  $: filterGroups = [
    {
      label: "Result",
      key: "result",
      options: [
        { value: "ok", label: "OK" },
        { value: "error", label: "Error" },
      ],
      selected: selectedResults,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  function handleFilterChange(key: string, value: string) {
    if (key !== "result") return;
    selectedResults = selectedResults.includes(value)
      ? selectedResults.filter((v) => v !== value)
      : [...selectedResults, value];
  }

  function clearFilters() {
    selectedResults = [];
  }

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
    <h1 class="text-fg-secondary text-lg font-bold">Recent history</h1>
    <p class="text-fg-secondary text-sm">Showing up to 10 most recent runs</p>
  </div>
  {#if $reportQuery.error}
    <div class="text-red-500">
      {$reportQuery.error.message}
    </div>
  {:else if $reportQuery.isLoading}
    <div class="text-fg-secondary">Loading...</div>
  {:else if !history.length}
    <NoRunsYet />
  {:else}
    <ResourceList
      kind="report"
      {columns}
      data={processedHistory}
      fixedRowHeight={false}
    >
      <TableToolbar
        slot="toolbar"
        showSearch={false}
        {filterGroups}
        onFilterChange={handleFilterChange}
        onClearAllFilters={clearFilters}
        {sortDirection}
        onSortToggle={() =>
          (sortDirection = sortDirection === "newest" ? "oldest" : "newest")}
      />
    </ResourceList>
  {/if}
</div>
