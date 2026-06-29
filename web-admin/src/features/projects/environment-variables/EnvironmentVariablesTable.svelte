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
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let data: V1ProjectVariable[];
  export let emptyText: string = "";
  export let variableNames: VariableNames = [];

  $: resolvedEmptyText = emptyText || m.env_no_variables();

  $: columns = [
    {
      accessorKey: "name",
      header: m.env_table_key_header(),
      cell: ({ row }) =>
        renderComponent(KeyCell, {
          name: row.original.name,
          environment: row.original.environment,
        }),
    },
    {
      accessorKey: "value",
      header: m.env_table_value_header(),
      enableSorting: false,
      cell: ({ row }) =>
        renderComponent(ValueCell, {
          value: row.original.value,
        }),
    },
    {
      header: m.env_table_activity_header(),
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
  ] as ColumnDef<V1ProjectVariable, any>[];
</script>

<BasicTable
  {data}
  {columns}
  emptyIcon={KeyIcon}
  emptyText={resolvedEmptyText}
  columnLayout="minmax(170px, 1.75fr) 2fr minmax(84px, 1fr) 56px"
/>
