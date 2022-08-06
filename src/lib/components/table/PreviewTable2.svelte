<script lang="ts">
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { tweened } from "svelte/motion";
  import Cell from "./Cell.svelte";
  import Row from "./Row.svelte";

  export let data;
  export let columns;

  let rowVirtualizer;
  let columnVirtualizer;
  let container;
  let columnOrder;
  let columnSizes;

  $: if (data && columns) {
    // get row count

    columnOrder = columns.reduce((obj, profile, i) => {
      obj[i] = profile;
      return obj;
    }, {});

    columnSizes = tweened(
      columns.reduce((obj, _, i) => {
        obj[i] = 200;
        return obj;
      }, {}),
      { duration: 50 }
    );

    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: data.length,
      estimateSize: () => 36,
      overscan: 32,
      paddingStart: 36,
    });
    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: columns.length,
      estimateSize: (index) => $columnSizes[index],
      overscan: 10,
    });
  }

  let activeIndex;
  function clearActiveIndex() {
    activeIndex = false;
  }

  /** handle scrolling tooltip suppression */
  let scrolling = false;
  let timeoutID;
  $: {
    if (scrolling) {
      if (timeoutID) clearTimeout(timeoutID);
      timeoutID = setTimeout(() => {
        scrolling = false;
      }, 200);
    }
  }
</script>

<div
  bind:this={container}
  style:width="100%"
  style:height="100%"
  class="overflow-auto"
  on:scroll={() => {
    /** capture to suppress cell tooltips. Otherwise,
     * there's quite a bit of rendering jank.
     */
    scrolling = true;
  }}
>
  {#if rowVirtualizer}
    <table
      style:height="{$rowVirtualizer.getTotalSize()}px"
      class="relative bg-white"
      on:mouseleave={clearActiveIndex}
      on:blur={clearActiveIndex}
    >
      <tr class="sticky top-0 z-10">
        {#each $columnVirtualizer.getVirtualItems() as header, i (header.key)}
          <th
            style:left={0}
            style:top={0}
            style:transform="translateX({header.start}px)"
            style:width="{header.size}px"
            style:height="36px"
            class="absolute bg-white text-left border-b border-b-4 border-r border-r-1 text-ellipsis overflow-hidden whitespace-nowrap "
          >
            {columnOrder[header.index].name}
          </th>
        {/each}
      </tr>
      {#each $rowVirtualizer.getVirtualItems() as row (row.key)}
        <Row>
          {#each $columnVirtualizer.getVirtualItems() as column (`${row.key}-${column.key}`)}
            {@const value = data[row.index][columnOrder[column.index].name]}
            {@const type = columnOrder[column.index].type}
            {@const rowActive = activeIndex === row.index}
            {@const suppressTooltip = scrolling}
            <Cell
              {suppressTooltip}
              {rowActive}
              {value}
              {row}
              {column}
              {type}
              on:inspect={(event) => {
                activeIndex = event.detail;
              }}
            />
          {/each}
        </Row>
      {/each}
    </table>
  {/if}
</div>
