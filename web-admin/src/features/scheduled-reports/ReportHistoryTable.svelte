<script lang="ts">
  import type { ColumnDef } from "@tanstack/table-core/src/types";
  import Table from "../../components/table/Table.svelte";
  import { defaultData, ReportRun } from "./fetch-report-history";

  export let organization: string;
  export let project: string;

  // TODO: fetch reports for a given project

  // Note: need an accessorFn in order to enable sorting and filtering

  const columns: ColumnDef<ReportRun>[] = [
    {
      // TODO: this header should be a Search element
      id: "monocolumn",
      header: "",
      // accessorFn that returns all the info to filter on -- this will take some massauging to get right (objects and arrays don't seem to work)
      accessorFn: (row) => row.timestamp.toLocaleString(),
      // cell: (info) =>
      // flexRender(ReportsTableInfoCell, {
      //   id: info.row.original.id,
      //   reportName: info.row.original.name,
      //   lastRun: info.row.original.lastRun,
      //   frequency: info.row.original.frequency,
      //   author: info.row.original.author,
      //   status: info.row.original.status,
      // }),
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
      // cell: (info) =>
      //   flexRender(ReportsTableActionCell, {
      //     reportName: info.row.original.name,
      //   }),
    },
  ];

  function globalFilterFn(row: ReportRun, filter: string) {
    return row.timestamp.toLocaleString().includes(filter.toLowerCase());
  }
</script>

<div class="flex flex-col gap-y-4">
  <h1 class="text-gray-800 text-base font-medium leading-none">
    Report history
  </h1>
  <Table {columns} data={defaultData} {globalFilterFn} />
</div>
