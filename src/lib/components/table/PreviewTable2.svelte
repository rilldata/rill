<script lang="ts">
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { onMount } from "svelte";
  import { tweened } from "svelte/motion";
  import Cell from "./Cell.svelte";
  import ColumnHeader from "./ColumnHeader.svelte";
  import Row from "./Row.svelte";

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

  /** we're going to hard-position the row number column due to
   * eccentricities in the virtual table.
   */
  let containerLeft = 0;
  let containerTop = 0;
  let scrollLeft = 0;
  let scrollTop = 0;

  function place() {
    const rect = container.getBoundingClientRect();
    containerLeft = rect.left;
    containerTop = rect.top;
  }
  onMount(() => {
    const config = { attributes: true };

    const observer = new ResizeObserver(() => {
      place();
    });
    place();
    observer.observe(container, config);
  });
</script>

translateX({scrollLeft}px), translateY({scrollTop}px)
<div
  bind:this={container}
  style:width="100%"
  style:height="100%"
  class="overflow-auto"
  on:scroll={(event) => {
    /** capture to suppress cell tooltips. Otherwise,
     * there's quite a bit of rendering jank.
     */
    scrollLeft = event?.target?.scrollLeft || 32;
    scrollTop = event?.target?.scrollTop || 0;
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
      <tr class="w-full sticky relative top-0 z-10">
        <!-- <StickyHeader position="top-left" header={{ size: 100, start: 0 }}>
          "row"
        </StickyHeader> -->
        <th
          class="fixed z-10"
          style:left="{containerLeft}px"
          style:top="{containerTop}px"
        >
          row
        </th>
        <!-- <StickyHeader position="top-left" header={{ size: 60, start: 0 }}>
          row
        </StickyHeader> -->
        {#each $columnVirtualizer.getVirtualItems() as header, i (header.key)}
          {@const name = columnOrder[header.index]?.name}
          {@const type = columnOrder[header.index]?.type}
          <ColumnHeader {header} {name} {type} />
        {/each}
      </tr>
      {#each $rowVirtualizer.getVirtualItems() as row (row.key)}
        <Row {row} width={$columnVirtualizer.getTotalSize()}>
          <th
            style:height="36px"
            style:width="60px"
            style:left={0}
            style:top="{row.start}px"
            class="sticky z-20"
          >
            <div style:left={0} class="absolute z-12">
              {row.index + 1}
            </div>
          </th>
          {#each $columnVirtualizer.getVirtualItems() as column (`${row.key}-${column.key}`)}
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
    </table>
  {/if}
</div>
