<script lang="ts">
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { tweened } from "svelte/motion";
  import Cell from "./Cell.svelte";
  import ColumnHeader from "./ColumnHeader.svelte";
  import Row from "./Row.svelte";
  import StickyHeader from "./StickyHeader.svelte";

  export let rows;
  export let columnNames;

  let rowVirtualizer;
  let columnVirtualizer;
  let container;
  let columnOrder;
  let columnSizes;
  let pinnedColumns = [];

  $: if (rows && columnNames) {
    columnOrder = columnNames.reduce((obj, profile, i) => {
      obj[i] = profile;
      return obj;
    }, {});

    columnSizes = tweened(
      columnNames.reduce((obj, _, i) => {
        obj[i] = 200;
        return obj;
      }, {}),
      { duration: 50 }
    );

    rowVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      count: rows.length,
      estimateSize: () => 36,
      overscan: 90,
      paddingStart: 36,
    });
    columnVirtualizer = createVirtualizer({
      getScrollElement: () => container,
      horizontal: true,
      count: columnNames.length,
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

  /** pinning functionality */
  function handlePin(column) {
    if (pinnedColumns.some((p) => p.name === column.name)) {
      pinnedColumns = [...pinnedColumns.filter((c) => c.name !== column.name)];
    } else {
      pinnedColumns = [...pinnedColumns, column];
    }
  }
</script>

<div
  bind:this={container}
  style:width="100%"
  style:height="100%"
  class="overflow-auto grid"
  style:grid-template-columns="max-content auto"
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
      <!-- header-->
      <div class="w-full sticky relative top-0 z-10">
        {#each $columnVirtualizer.getVirtualItems() as header, i (header.key)}
          {@const name = columnOrder[header.index]?.name}
          {@const type = columnOrder[header.index]?.type}
          {@const pinned = pinnedColumns.some((column) => column.name === name)}
          <ColumnHeader
            {header}
            {name}
            {type}
            {pinned}
            on:pin={() => {
              handlePin(columnOrder[header.index]);
            }}
          />
        {/each}
      </div>
      <!-- body -->
      <div
        class="sticky left-0 top-0 z-20"
        style:height="{$rowVirtualizer.getTotalSize()}px"
        style:width="60px"
      >
        <StickyHeader header={{ size: 60, start: 0 }} position="top-left"
          >#</StickyHeader
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
            {@const value = rows[row.index][columnOrder[column.index]?.name]}
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
    <!-- sticker -->

    {#if pinnedColumns.length}
      <div
        style:right={0}
        class=" top-0 sticky z-40 border-l-2 border-gray-400"
        style:width="{pinnedColumns.length * 200}px"
      >
        <div class="w-full sticky relative top-0 z-10">
          {#each pinnedColumns as column, i (column.name)}
            <ColumnHeader
              header={{ start: i * 200, size: 200 }}
              name={column.name}
              type={column.type}
              on:pin={() => handlePin(column)}
              pinned={true}
            />
          {/each}
        </div>
        {#each pinnedColumns as column, i (column.name)}
          <Row>
            {#each $rowVirtualizer.getVirtualItems() as row (`${row.key}-${column.index}`)}
              {@const value = rows[row.index][column.name]}
              {@const type = column.type}
              {@const rowActive = activeIndex === row?.index}
              {@const suppressTooltip = scrolling}

              <Cell
                {suppressTooltip}
                {rowActive}
                {value}
                {row}
                column={{ start: i * 200, size: 200 }}
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
  {/if}
</div>
