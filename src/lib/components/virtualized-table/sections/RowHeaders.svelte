<script lang="ts">
  import { getContext } from "svelte";
  import Cell from "../core/Cell.svelte";
  import StickyHeader from "../core/StickyHeader.svelte";
  import type { VirtualizedTableConfig } from "../types";

  const config: VirtualizedTableConfig = getContext("config");

  export let totalHeight: number;
  export let virtualRowItems;

  export let virtualColumnItems = null;
  export let rows = [];
  export let columnName: string = "#";
  export let scrolling = false;
  export let activeIndex: number = null;

  let showRowNumbers = true;
  let rowHeaderWidth = config.indexWidth;

  if (rows.length) {
    showRowNumbers = false;
    rowHeaderWidth = virtualColumnItems[0].size;
  }
</script>

<div
  class="sticky left-0 top-0 z-20"
  style:height="{totalHeight}px"
  style:width="{rowHeaderWidth}px"
>
  <StickyHeader
    header={{ size: rowHeaderWidth, start: 0 }}
    enableResize={false}
    position="top-left">{columnName}</StickyHeader
  >
  {#each virtualRowItems as row (`row-${row.key}`)}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: rowHeaderWidth, start: row.start }}
    >
      {#if showRowNumbers}
        {row.key + 1}
      {:else}
        {@const value = rows[row?.index][columnName]}
        {@const type = "VARCHAR"}
        {@const rowActive = activeIndex === row?.index}
        {@const suppressTooltip = scrolling}
        <Cell
          {suppressTooltip}
          {rowActive}
          {value}
          {row}
          {type}
          column={virtualColumnItems[0]}
          on:inspect
        />
      {/if}
    </StickyHeader>
  {/each}
</div>
