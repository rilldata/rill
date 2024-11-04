<script lang="ts">
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import type { V1ProjectVariable } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import KeyIcon from "@rilldata/web-common/components/icons/KeyIcon.svelte";
  import ActivityCell from "./ActivityCell.svelte";
  import KeyCell from "./KeyCell.svelte";
  import ValueCell from "./ValueCell.svelte";
  import ActionsCell from "./ActionsCell.svelte";

  export let data: V1ProjectVariable[];

  const columns: ColumnDef<V1ProjectVariable, any>[] = [
    {
      accessorKey: "name",
      header: "Key",
      cell: ({ row }) =>
        flexRender(KeyCell, {
          name: row.original.name,
          environment: row.original.environment,
        }),
      meta: {
        widthPercent: 30,
      },
    },
    {
      accessorKey: "value",
      header: "Value",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(ValueCell, {
          value: row.original.value,
        }),
      meta: {
        widthPercent: 30,
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
    {
      accessorKey: "actions",
      header: "",
      cell: ({ row }) =>
        flexRender(ActionsCell, {
          id: row.original.id,
        }),
      enableSorting: false,
      meta: {
        widthPercent: 0,
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
