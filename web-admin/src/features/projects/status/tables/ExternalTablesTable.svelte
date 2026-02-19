<script lang="ts">
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type { V1OlapTableInfo } from "@rilldata/web-common/runtime-client";
  import { compareSizes } from "./utils";
  import ModelSizeCell from "./ModelSizeCell.svelte";
  import NameCell from "../resource-table/NameCell.svelte";
  import MaterializationCell from "./MaterializationCell.svelte";

  export let tables: V1OlapTableInfo[] = [];
  export let isView: Map<string, boolean> = new Map();

  const columns: ColumnDef<V1OlapTableInfo, unknown>[] = [
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
        return compareSizes(sizeA, sizeB);
      },
      cell: ({ getValue }) =>
        flexRender(ModelSizeCell, {
          sizeBytes: getValue() as string | number | undefined,
        }),
    },
  ];
</script>

<VirtualizedTable
  tableId="external-tables-table"
  data={tables}
  {columns}
  columnLayout="minmax(80px, 0.5fr) minmax(150px, 2fr) minmax(100px, 1fr)"
/>
