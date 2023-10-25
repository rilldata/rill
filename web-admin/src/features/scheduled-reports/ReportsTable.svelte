<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import Table from "../../components/table/Table.svelte";
  import { defaultData } from "./fetch-reports";
  import ReportsTableCompositeCell from "./ReportsTableCompositeCell.svelte";
  import ReportsTableHeader from "./ReportsTableHeader.svelte";

  export let organization: string;
  export let project: string;

  // TODO: fetch reports for a given project

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
      cell: (info) =>
        flexRender(ReportsTableCompositeCell, {
          id: info.row.original.id,
          reportName: info.row.original.name,
          lastRun: info.row.original.lastRun,
          frequency: info.row.original.frequency,
          author: info.row.original.author,
          status: info.row.original.status,
        }),
    },
    {
      id: "name",
      accessorFn: (row) => row.name,
    },
    {
      id: "lastRun",
      accessorFn: (row) => row.lastRun,
    },
    // {
    //   id: "nextRun",
    //   accessorFn: (row) => row.nextRun,
    // },
    // {
    //   id: "actions",
    //   cell: ({ row }) =>
    //     flexRender(ReportsTableActionCell, {
    //       reportName: row.original.name,
    //     }),
    // },
  ];

  const columnVisibility = {
    name: false,
    lastRun: false,
  };
</script>

<Table {columns} data={defaultData} {columnVisibility}>
  <ReportsTableHeader slot="header" />
</Table>
