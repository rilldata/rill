<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import type { ColumnConfig } from "./ColumnConfig";
  import EditableTableHeader from "./EditableTableHeader.svelte";
  import TableCellRenderer from "./TableCellRenderer.svelte";
  import TableHeader from "./TableHeader.svelte";
  import TableRow from "./TableRow.svelte";
  import TableRowWithMenu from "./TableRowWithMenu.svelte";

  export let columnNames: ColumnConfig<any>[];
  export let rows: any[];
  export let label: string | undefined = undefined;

  const dispatch = createEventDispatcher();

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
  aria-label={label}
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
    <TableRowWithMenu
      {index}
      on:delete={() => dispatch("delete", index)}
      menuLabel="More"
    >
      {#each columnNames as column (index + column.name + column.label)}
        <TableCellRenderer columnConfig={column} {row} {index} />
      {/each}
    </TableRowWithMenu>
  {/each}
</table>
