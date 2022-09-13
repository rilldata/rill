<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { VirtualizedTableColumns } from "$lib/types";
  import ColumnHeader from "../core/ColumnHeader.svelte";

  const dispatch = createEventDispatcher();

  export let columns: VirtualizedTableColumns[];
  export let pinnedColumns: VirtualizedTableColumns[] = [];
  export let virtualColumnItems;
  export let noPin = false;
  export let showDataIcon = false;
  export let selectedColumn: string = null;

  $: columnHeaders = virtualColumnItems.map((column) => {
    const name = columns[column.index]?.label || columns[column.index]?.name;
    const isEnableResizeDefined = "enableResize" in columns[column.index];
    const enableResize = isEnableResizeDefined
      ? columns[column.index].enableResize
      : true;
    return {
      name,
      key: column.key,
      index: column.index,
      header: { start: column.start, size: column.size },
      type: columns[column.index]?.type,
      pinned: pinnedColumns.some((column) => column.name === name),
      isSelected: selectedColumn === columns[column.index]?.name,
      enableResize,
    };
  });
</script>

<div class="w-full sticky relative top-0 z-10">
  {#each columnHeaders as column (column.key)}
    <ColumnHeader
      on:resize-column
      on:reset-column-size
      name={column.name}
      header={column.header}
      type={column.type}
      pinned={column.pinned}
      isSelected={column.isSelected}
      enableResize={column.enableResize}
      {noPin}
      {showDataIcon}
      on:pin={() => {
        dispatch("pin", columns[column.index]);
      }}
      on:click-column={() => {
        dispatch("click-column", columns[column.index]?.name);
      }}
    />
  {/each}
</div>
