<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";
  import ModelSizeCell from "./ModelSizeCell.svelte";
  import NameCell from "./NameCell.svelte";
  import RowCountCell from "./RowCountCell.svelte";
  import ColumnCountCell from "./ColumnCountCell.svelte";
  import MaterializationCell from "./MaterializationCell.svelte";

  export let tables: V1OlapTableInfo[] = [];
  export let columnCounts: Map<string, number> = new Map();
  export let rowCounts: Map<string, number | "loading" | "error"> = new Map();
  export let isView: Map<string, boolean> = new Map();

  const columns: ColumnDef<V1OlapTableInfo, any>[] = [
    {
      id: "materialization",
      accessorFn: (row) => isView.get(row.name ?? ""),
      header: "Type",
      cell: ({ row, getValue }) =>
        flexRender(MaterializationCell, {
          isView: getValue() as boolean | undefined,
          physicalSizeBytes: row.original.physicalSizeBytes,
        }),
    },
    {
      accessorFn: (row) => row.name,
      header: "Name",
      cell: ({ getValue }) =>
        flexRender(NameCell, {
          name: getValue() as string,
        }),
    },
    {
      id: "rowCount",
      accessorFn: (row) => rowCounts.get(row.name ?? ""),
      header: "Row Count",
      sortingFn: (rowA, rowB) => {
        const a = rowA.getValue("rowCount");
        const b = rowB.getValue("rowCount");
        if (typeof a === "number" && typeof b === "number") return b - a;
        // "loading" and "error" go to bottom
        return 0;
      },
      cell: ({ getValue }) =>
        flexRender(RowCountCell, {
          count: getValue() as number | "loading" | "error" | undefined,
        }),
    },
    {
      id: "columnCount",
      accessorFn: (row) => columnCounts.get(row.name ?? ""),
      header: "Column Count",
      cell: ({ getValue }) =>
        flexRender(ColumnCountCell, {
          count: getValue() as number | undefined,
        }),
    },
    {
      id: "size",
      accessorFn: (row) => row.physicalSizeBytes,
      header: "Database Size",
      sortDescFirst: true,
      sortingFn: (rowA, rowB) => {
        const sizeA = rowA.getValue("size") as string | number | undefined;
        const sizeB = rowB.getValue("size") as string | number | undefined;

        let numA = -1;
        if (sizeA && sizeA !== "-1") {
          numA = typeof sizeA === "number" ? sizeA : parseInt(sizeA, 10);
        }

        let numB = -1;
        if (sizeB && sizeB !== "-1") {
          numB = typeof sizeB === "number" ? sizeB : parseInt(sizeB, 10);
        }

        return numB - numA; // Descending
      },
      cell: ({ getValue }) =>
        flexRender(ModelSizeCell, {
          sizeBytes: getValue() as string | number | undefined,
        }),
    },
  ];

  $: tableData = tables;
</script>

{#key columnCounts && rowCounts}
  <VirtualizedTable
    data={tableData}
    {columns}
    columnLayout="minmax(60px, 0.5fr) minmax(150px, 3fr) minmax(100px, 1fr) minmax(100px, 1fr) minmax(100px, 1fr)"
  />
{/key}
