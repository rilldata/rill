<script lang="ts">
  import Table from "$lib/components/table-editable/Table.svelte";
  import TableRow from "$lib/components/table-editable/TableRow.svelte";
  import TableRowWithMenu from "$lib/components/table-editable/TableRowWithMenu.svelte";

  import TableCellInput from "$lib/components/table-editable/TableCellInput.svelte";

  import { createEventDispatcher } from "svelte";
  import EditableTableHeader from "$lib/components/table-editable/EditableTableHeader.svelte";

  import {
    ColumnConfig,
    RenderType,
  } from "$lib/components/table-editable/ColumnConfig";

  import { columnIsPinned } from "$lib/components/table-editable/pinnableUtils";

  import TableCellSparkline from "$lib/components/metrics-definition/TableCellSparkline.svelte";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig[];
  export let selectedColumns: ColumnConfig[];
  export let rows: any[];
</script>

<Table>
  <!-- headers -->
  <TableRow>
    {#each columnNames as columnConfig (columnConfig.name + columnConfig.label)}
      <EditableTableHeader
        {columnConfig}
        pinned={columnIsPinned(columnConfig.name, selectedColumns)}
        on:pin={() => {
          dispatch("pin", { columnConfig });
        }}
      />
    {/each}
  </TableRow>
  <!-- values -->
  {#each rows as row, index}
    <TableRowWithMenu on:delete={() => dispatch("delete", row.id)}>
      {#each columnNames as column (index + column.name + column.label)}
        {#if column.renderType === RenderType.INPUT}
          <TableCellInput
            on:change={(evt) => dispatch("change", evt.detail)}
            {column}
            {index}
            {row}
          />
        {:else if column.renderType === RenderType.SPARKLINE}
          <TableCellSparkline measureId={row[column.name]} />
        {:else if column.renderType === RenderType.CARDINALITY}
          <td class="py-2 px-4 border border-gray-200 hover:bg-gray-200"
            >(card. placeholder)</td
          >
        {/if}
      {/each}
    </TableRowWithMenu>
  {/each}
</Table>
