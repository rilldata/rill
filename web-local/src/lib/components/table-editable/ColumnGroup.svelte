<script lang="ts">
  import { onMount } from "svelte";
  import TableRow from "./TableRow.svelte";
  import TableRowWithMenu from "./TableRowWithMenu.svelte";

  import { createEventDispatcher } from "svelte";
  import EditableTableHeader from "./EditableTableHeader.svelte";

  import type { ColumnConfig } from "./ColumnConfig";

  import TableCellRenderer from "./TableCellRenderer.svelte";
  import TableHeader from "./TableHeader.svelte";

  const dispatch = createEventDispatcher();

  export let columnNames: ColumnConfig<any>[];
  export let rows: any[];

  let tableElement;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      dispatch("tableResize", tableElement.clientHeight);
    });
    observer.observe(tableElement);
    return () => observer.unobserve(tableElement);
  });
</script>

<table
  on:mouseleave
  class="relative border-collapse bg-white"
  bind:this={tableElement}
>
  <!-- headers -->
  <TableRow>
    <TableHeader position="top-left" zIndexClass="z-30">#</TableHeader>
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
    <TableRowWithMenu {index} on:delete={() => dispatch("delete", index)}>
      {#each columnNames as column (index + column.name + column.label)}
        <TableCellRenderer columnConfig={column} {row} {index} />
      {/each}
    </TableRowWithMenu>
  {/each}
</table>
