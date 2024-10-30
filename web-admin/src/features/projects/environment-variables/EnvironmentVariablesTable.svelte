<script lang="ts">
  import type { ColumnDef } from "@tanstack/svelte-table";
  import type { V1ProjectVariable } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";

  export let data: V1ProjectVariable[];

  const columns: ColumnDef<V1ProjectVariable, any>[] = [
    {
      accessorKey: "name",
      header: "Name",
      enableSorting: false,
      cell: ({ row }) => row.original.name,
    },
  ];
</script>

<BasicTable
  {data}
  {columns}
  emptyText="No environment variables found"
  scrollable
/>
