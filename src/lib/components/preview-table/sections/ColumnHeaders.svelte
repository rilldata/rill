<script lang="ts">
  import type { ProfileColumn } from "$lib/types";

  import { createEventDispatcher } from "svelte";

  import ColumnHeader from "../core/ColumnHeader.svelte";
  const dispatch = createEventDispatcher();
  export let columns: ProfileColumn[];
  export let pinnedColumns: ProfileColumn[];
  export let virtualColumnItems;
</script>

<div class="w-full sticky relative top-0 z-10">
  {#each virtualColumnItems as header (header.key)}
    {@const name = columns[header.index]?.name}
    {@const type = columns[header.index]?.type}
    {@const pinned = pinnedColumns.some((column) => column.name === name)}
    <ColumnHeader
      on:resize-column
      on:reset-column-size
      {header}
      {name}
      {type}
      {pinned}
      on:pin={() => {
        dispatch("pin", columns[header.index]);
      }}
    />
  {/each}
</div>
