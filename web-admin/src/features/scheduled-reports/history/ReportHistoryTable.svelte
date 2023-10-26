<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/table-core/src/types";
  import Table from "../../../components/table/Table.svelte";
  import { defaultData, ReportRun } from "./fetch-report-history";
  import ReportHistoryTableCompositeCell from "./ReportHistoryTableCompositeCell.svelte";
  import ReportHistoryTableHeader from "./ReportHistoryTableHeader.svelte";

  export let organization: string;
  export let project: string;

  // TODO: fetch report

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
  const columns: ColumnDef<ReportRun>[] = [
    {
      id: "composite",
      cell: (info) =>
        flexRender(ReportHistoryTableCompositeCell, {
          id: info.row.original.id,
          timestamp: info.row.original.timestamp,
          exportSize: info.row.original.exportSize,
          status: info.row.original.status,
        }),
    },
    {
      id: "timestamp",
      header: "",
      accessorFn: (row) => row.timestamp,
      cell: undefined,
    },
  ];

  const columnVisibility = {
    timestamp: false,
  };
</script>

<div class="flex flex-col gap-y-4 w-full">
  <h1 class="text-gray-800 text-base font-medium leading-none">
    Report history
  </h1>
  <Table
    {columns}
    data={defaultData}
    {columnVisibility}
    maxWidthOverride="max-w-[960px]"
  >
    <ReportHistoryTableHeader slot="header" maxWidthOverride="max-w-[960px]" />
  </Table>
</div>
