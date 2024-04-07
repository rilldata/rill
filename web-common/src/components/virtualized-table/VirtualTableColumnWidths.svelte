<script lang="ts">
  import { ROW_HEADER_WIDTH } from "./VirtualTable.svelte";

  export let rowHeaders: boolean;
  export let pinnedColumns: Map<number, number>;
  export let columnWidths: number[];
  export let paddingLeft: number;
  export let renderedColumns: number;
  export let startColumn: number;
</script>

<colgroup>
  {#if rowHeaders}
    <col style:width="{ROW_HEADER_WIDTH}px" />
  {/if}

  {#each pinnedColumns as [index] (index)}
    <col style:width="{columnWidths[index]}px" />
  {/each}

  <col style:width="{paddingLeft}px" />

  {#each { length: renderedColumns } as _, i (i)}
    {#if !pinnedColumns.has(startColumn + i)}
      <col style:width="{columnWidths[i]}px" />
    {/if}
  {/each}

  <col class="w-full" />
</colgroup>
