<script lang="ts">
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import type { V1ProjectVariable } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import KeyIcon from "@rilldata/web-common/components/icons/KeyIcon.svelte";
  import ActivityCell from "./ActivityCell.svelte";
  import KeyCell from "./KeyCell.svelte";
  import ValueCell from "./ValueCell.svelte";
  import ActionsCell from "./ActionsCell.svelte";
  import type { VariableNames } from "./types";

  export let data: V1ProjectVariable[];
  export let emptyText: string = "No environment variables";
  export let variableNames: VariableNames = [];

  const columns: ColumnDef<V1ProjectVariable, any>[] = [
    {
      accessorKey: "name",
      header: "Key",
      cell: ({ row }) =>
        renderComponent(KeyCell, {
          name: row.original.name,
          environment: row.original.environment,
        }),
    },
    {
      accessorKey: "value",
      header: "Value",
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(ValueCell, {
          value: row.original.value,
        }),
    },
    {
      header: "Activity",
      sortDescFirst: true,
      accessorFn: (row) => row.createdOn,
      cell: ({ row }) => {
        return renderComponent(ActivityCell, {
          updatedOn: row.original.updatedOn,
        });
      },
    },
    {
      accessorKey: "actions",
      header: "",
      cell: ({ row }) =>
        renderComponent(ActionsCell, {
          id: row.original.id,
          name: row.original.name,
          value: row.original.value,
          environment: row.original.environment,
          variableNames,
        }),
      enableSorting: false,
    },
  ];
</script>

<BasicTable
  {data}
  {columns}
  emptyIcon={KeyIcon}
  {emptyText}
  columnLayout="minmax(170px, 1.75fr) 2fr minmax(84px, 1fr) 56px"
  enableSorting={false}
/>
