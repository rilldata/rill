<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";
  import ModelSizeCell from "./ModelSizeCell.svelte";
  import NameCell from "./NameCell.svelte";
  import MaterializationCell from "./MaterializationCell.svelte";

  export let tables: V1OlapTableInfo[] = [];
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

{#key isView}
  <VirtualizedTable
    tableId="project-tables-table"
    data={tableData}
    {columns}
    columnLayout="minmax(60px, 0.5fr) minmax(150px, 3fr) minmax(100px, 1fr)"
  />
{/key}
