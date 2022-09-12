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
  export let selectedColumn: string;
</script>

<div class="w-full sticky relative top-0 z-10">
  {#each virtualColumnItems as header (header.key)}
    {@const name = columns[header.index]?.label || columns[header.index]?.name}
    {@const type = columns[header.index]?.type}
    {@const pinned = pinnedColumns.some((column) => column.name === name)}
    {@const isSelected = selectedColumn === columns[header.index]?.name}

    <ColumnHeader
      on:resize-column
      on:reset-column-size
      {header}
      {name}
      {type}
      {noPin}
      {pinned}
      {showDataIcon}
      {isSelected}
      on:pin={() => {
        dispatch("pin", columns[header.index]);
      }}
      on:click-column={() => {
        dispatch("click-column", columns[header.index]?.name);
      }}
    />
  {/each}
</div>
