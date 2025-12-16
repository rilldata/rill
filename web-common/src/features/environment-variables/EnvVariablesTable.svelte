<script lang="ts">
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import KeyIcon from "@rilldata/web-common/components/icons/KeyIcon.svelte";
  import KeyCell from "./KeyCell.svelte";
  import ValueCell from "./ValueCell.svelte";
  import type { EnvVariable } from "./types";

  export let data: EnvVariable[];
  export let emptyText: string = "No environment variables";
  export let actionsColumn: ColumnDef<EnvVariable, any> | null = null;

  $: columns = [
    {
      accessorKey: "key",
      header: "Key",
      cell: ({ row }: any) =>
        flexRender(KeyCell, {
          name: row.original.key,
        }),
    },
    {
      accessorKey: "value",
      header: "Value",
      enableSorting: false,
      cell: ({ row }: any) =>
        flexRender(ValueCell, {
          value: row.original.value,
        }),
    },
    ...(actionsColumn ? [actionsColumn] : []),
  ] as ColumnDef<EnvVariable, any>[];
</script>

<BasicTable
  {data}
  {columns}
  emptyIcon={KeyIcon}
  {emptyText}
  columnLayout={actionsColumn
    ? "minmax(170px, 1.75fr) 2fr 56px"
    : "minmax(170px, 1.75fr) 2fr"}
/>
