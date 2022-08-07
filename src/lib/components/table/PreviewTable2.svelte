<script lang="ts">
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { tweened } from "svelte/motion";
  import Cell from "./Cell.svelte";
  import ColumnHeader from "./ColumnHeader.svelte";
  import Row from "./Row.svelte";
  import StickyHeader from "./StickyHeader.svelte";

  export let data;
  export let columns;

  let rowVirtualizer;
  let columnVirtualizer;
  let container;
  let columnOrder;
  let columnSizes;

  const defaultRangeExtractor = (range) => {
    const start = Math.max(range.startIndex - range.overscan, 0);
    const end = Math.min(range.endIndex + range.overscan, range.count - 1);
    const arr = [];

    for (let i = start; i <= end; i++) {
      arr.push(i);
    }

    return arr;
  };

  let stickyIndexes = [0];
  let activeStickyIndex;
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
      overscan: 60,
      paddingStart: 36,
    });
    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: columns.length,
      estimateSize: (index) => $columnSizes[index],
      overscan: 10,
      paddingStart: 60,
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
    <div
      class="relative bg-white"
      on:mouseleave={clearActiveIndex}
      on:blur={clearActiveIndex}
      style:will-change="transform, contents"
      style:width="{$columnVirtualizer.getTotalSize()}px"
      style:height="{$rowVirtualizer.getTotalSize()}px"
    >
      <div class="w-full sticky relative top-0 z-10">
        {#each $columnVirtualizer.getVirtualItems() as header, i (header.key)}
          {@const name = columnOrder[header.index]?.name}
          {@const type = columnOrder[header.index]?.type}
          <ColumnHeader {header} {name} {type} />
        {/each}
      </div>
      <div
        class="sticky left-0 top-0 z-20"
        style:height="{$rowVirtualizer.getTotalSize()}px"
        style:width="60px"
      >
        <StickyHeader header={{ size: 60, start: 0 }} position="top-left"
          >row</StickyHeader
        >
        {#each $rowVirtualizer.getVirtualItems() as row (`row-${row.key}`)}
          <div
            class="absolute left-0 z-20 bg-gray-100 grid place-items-center font-bold border-r border-gray-300 border-b"
            style:height="{row.size}px"
            style:width="60px"
            style:left={0}
            style:top={0}
            style:transform="translateY({row.start}px)"
          >
            {row.key + 1}
          </div>
        {/each}
      </div>

      {#each $columnVirtualizer.getVirtualItems() as column (column.key)}
        <Row>
          {#each $rowVirtualizer.getVirtualItems() as row (`${row.key}-${column.key}`)}
            {@const value = data[row.index][columnOrder[column.index]?.name]}
            {@const type = columnOrder[column.index]?.type}
            {@const rowActive = activeIndex === row?.index}
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
    </div>
  {/if}
</div>
