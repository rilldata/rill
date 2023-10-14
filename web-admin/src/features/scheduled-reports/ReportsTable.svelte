<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/table-core/src/types";
  import Table from "../../components/table/Table.svelte";
  import { defaultData, Report } from "./fetch-reports";
  import ReportsTableActionCell from "./ReportsTableActionCell.svelte";
  import ReportsTableHeader from "./ReportsTableHeader.svelte";
  import ReportsTableInfoCell from "./ReportsTableInfoCell.svelte";

  export let organization: string;
  export let project: string;

  // TODO: fetch reports for a given project

  // Note: need an accessorFn in order to enable sorting and filtering

  const columns: ColumnDef<Report>[] = [
    {
      // TODO: this header should be a Search element
      id: "monocolumn",
      header: "",
      // accessorFn that returns all the info to filter on -- this will take some massauging to get right (objects and arrays don't seem to work)
      accessorFn: (row) => row.name + row.author,
      cell: (info) =>
        flexRender(ReportsTableInfoCell, {
          id: info.row.original.id,
          reportName: info.row.original.name,
          lastRun: info.row.original.lastRun,
          frequency: info.row.original.frequency,
          author: info.row.original.author,
          status: info.row.original.status,
        }),
    },
    // Hidden column to enable sorting by last run. There's probably a better way than this.
    {
      id: "lastRun",
      header: "",
      accessorFn: (row) => row.lastRun,
      cell: undefined,
    },
    // Hidden column to enable sorting by next run. There's probably a better way than this.
    // {
    //   id: "nextRun",
    //   header: "",
    //   accessorFn: (row) => row.nextRun,
    //   cell: undefined,
    // },
    {
      // TODO: this header should be a "Display" dropdown button + a total rows count
      id: "actions",
      header: "",
      cell: (info) =>
        flexRender(ReportsTableActionCell, {
          reportName: info.row.original.name,
        }),
    },
  ];

  function globalFilterFn(row: Report, filter: string) {
    return (
      row.name.toLowerCase().includes(filter.toLowerCase()) ||
      row.author.toLowerCase().includes(filter.toLowerCase())
    );
  }
</script>

<Table dataTypeName="report" {columns} data={defaultData} {globalFilterFn}>
  <ReportsTableHeader slot="header" />
</Table>
