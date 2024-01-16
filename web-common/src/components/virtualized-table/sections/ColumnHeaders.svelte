<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import ColumnHeader from "../core/ColumnHeader.svelte";
  import type { VirtualizedTableColumns } from "../types";

  const dispatch = createEventDispatcher();

  export let columns: VirtualizedTableColumns[];
  export let pinnedColumns: VirtualizedTableColumns[] = [];
  export let virtualColumnItems;
  export let noPin = false;
  export let showDataIcon = false;
  export let selectedColumn: string | null = null;

  const getColumnHeaderProps = (header) => {
    const column = columns[header.index];
    const name = column.label || column.name;
    const isEnableResizeDefined = "enableResize" in column;
    const enableResize = isEnableResizeDefined ? column.enableResize : true;
    return {
      name,
      enableResize,
      type: column.type,
      description: column.description || "",
      pinned: pinnedColumns.some((pinCol) => pinCol.name === column.name),
      isSelected: selectedColumn === column.name,
      highlight: column.highlight,
      sorted: column.sorted,
    };
  };
</script>

<div class="w-full sticky relative top-0 z-10">
  {#each virtualColumnItems as header (header.key)}
    {@const props = getColumnHeaderProps(header)}
    <ColumnHeader
      on:resize-column
      on:reset-column-width
      {...props}
      {header}
      {noPin}
      {showDataIcon}
      on:pin={() => {
        dispatch("pin", columns[header.index]);
      }}
      on:click-column={() => {
        dispatch("click-column", columns[header.index]?.name);
      }}
    />
  {/each}
</div>
