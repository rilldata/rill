<script lang="ts">
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import type { V1ProjectVariable } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import KeyIcon from "@rilldata/web-common/components/icons/KeyIcon.svelte";
  import ActivityCell from "./ActivityCell.svelte";
  import KeyCell from "./KeyCell.svelte";
  import ValueCell from "./ValueCell.svelte";

  export let data: V1ProjectVariable[];

  const columns: ColumnDef<V1ProjectVariable, any>[] = [
    {
      accessorKey: "name",
      header: "Key",
      cell: ({ row }) => {
        return flexRender(KeyCell, {
          name: row.original.name,
          environment: row.original.environment,
        });
      },
      meta: {
        widthPercent: 20,
      },
    },
    {
      accessorKey: "value",
      header: "Value",
      enableSorting: false,
      cell: ({ row }) => {
        return flexRender(ValueCell, {
          value: row.original.value,
        });
      },
      meta: {
        widthPercent: 20,
      },
    },
    {
      header: "Activity",
      accessorFn: (row) => row.createdOn,
      enableSorting: false,
      cell: ({ row }) => {
        return flexRender(ActivityCell, {
          createdOn: row.original.createdOn,
          updatedOn: row.original.updatedOn,
        });
      },
    },
  ];
</script>

<BasicTable
  {data}
  {columns}
  emptyIcon={KeyIcon}
  emptyText="No environment variables"
  scrollable
/>
