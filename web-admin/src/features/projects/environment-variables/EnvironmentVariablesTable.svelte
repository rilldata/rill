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
    {
      accessorKey: "value",
      header: "Value",
      enableSorting: false,
      cell: ({ row }) => row.original.value,
    },
    // TODO
    // Environment the variable is set for.
    // If empty, the variable is shared for all environments.
    {
      accessorKey: "environment",
      header: "Environment",
      enableSorting: false,
      cell: ({ row }) => row.original.environment,
    },
    {
      accessorKey: "updatedByUserId",
      header: "Updated By",
      enableSorting: false,
      cell: ({ row }) => row.original.updatedByUserId,
    },
    {
      accessorKey: "createdOn",
      header: "Created On",
      enableSorting: false,
      cell: ({ row }) => row.original.createdOn,
    },
    {
      accessorKey: "updatedOn",
      header: "Updated On",
      enableSorting: false,
      cell: ({ row }) => row.original.updatedOn,
    },
  ];
</script>

<BasicTable
  {data}
  {columns}
  emptyText="No environment variables found"
  scrollable
/>
