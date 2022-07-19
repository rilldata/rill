<script lang="ts">
  import Table from "$lib/components/table-editable/Table.svelte";
  import TableRow from "$lib/components/table-editable/TableRow.svelte";
  import TableRowWithMenu from "$lib/components/table-editable/TableRowWithMenu.svelte";

  import { createEventDispatcher } from "svelte";
  import EditableTableHeader from "$lib/components/table-editable/EditableTableHeader.svelte";

  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";

  import TableCellRenderer from "./TableCellRenderer.svelte";
  import TableHeader from "./TableHeader.svelte";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig<any>[];
  export let rows: any[];
</script>

<Table>
  <!-- headers -->
  <TableRow>
    <TableHeader position="top-left">#</TableHeader>
    {#each columnNames as columnConfig (columnConfig.name + columnConfig.label)}
      <EditableTableHeader
        {columnConfig}
        on:pin={() => {
          dispatch("pin", { columnConfig });
        }}
      />
    {/each}
  </TableRow>
  <!-- values -->
  {#each rows as row, index}
    <TableRowWithMenu {index} on:delete={() => dispatch("delete", row.id)}>
      {#each columnNames as column (index + column.name + column.label)}
        <TableCellRenderer columnConfig={column} {row} {index} />
      {/each}
    </TableRowWithMenu>
  {/each}
</Table>
