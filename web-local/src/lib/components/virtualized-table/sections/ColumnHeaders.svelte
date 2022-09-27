<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { VirtualizedTableColumns } from "../../../types";
  import ColumnHeader from "../core/ColumnHeader.svelte";

  const dispatch = createEventDispatcher();

  export let columns: VirtualizedTableColumns[];
  export let pinnedColumns: VirtualizedTableColumns[] = [];
  export let virtualColumnItems;
  export let noPin = false;
  export let showDataIcon = false;
  export let selectedColumn: string = null;

  const getColumnHeaderProps = (header) => {
    const name = columns[header.index]?.label || columns[header.index]?.name;
    const isEnableResizeDefined = "enableResize" in columns[header.index];
    const enableResize = isEnableResizeDefined
      ? columns[header.index].enableResize
      : true;
    return {
      name,
      enableResize,
      type: columns[header.index]?.type,
      pinned: pinnedColumns.some((column) => column.name === name),
      isSelected: selectedColumn === columns[header.index]?.name,
    };
  };
</script>

<div class="w-full sticky relative top-0 z-10">
  {#each virtualColumnItems as header (header.key)}
    <ColumnHeader
      on:resize-column
      on:reset-column-size
      {...getColumnHeaderProps(header)}
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
