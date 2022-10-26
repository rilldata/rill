<script lang="ts">
  import { getContext } from "svelte";
  import type { VirtualizedTableColumns } from "../../types";
  import Cell from "../virtualized-table/core/Cell.svelte";
  import StickyHeader from "../virtualized-table/core/StickyHeader.svelte";
  import type { VirtualizedTableConfig } from "../virtualized-table/types";

  const config: VirtualizedTableConfig = getContext("config");
  export let totalHeight: number;
  export let virtualRowItems;
  export let selectedIndex = [];
  export let column: VirtualizedTableColumns;
  export let rows;
  export let width = config.indexWidth;

  // Cell props
  export let scrolling;
  export let activeIndex;
  export let excludeMode = false;

  $: atLeastOneSelected = !!selectedIndex?.length;

  const getCellProps = (row) => {
    const value = rows[row.index][column.name];
    return {
      value,
      formattedValue: value,
      type: column.type,
      suppressTooltip: scrolling,
      barValue: 0,
      rowSelected: selectedIndex.findIndex((tgt) => row?.index === tgt) >= 0,
    };
  };
</script>

<div
  class="sticky self-start left-6 top-0 z-20"
  style:height="{totalHeight}px"
  style:width="{width}px"
>
  <StickyHeader
    header={{ size: width, start: 0 }}
    enableResize={false}
    position="top-left"
  >
    <span class="px-1">{column.label || column.name}</span>
  </StickyHeader>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const rowActive = activeIndex === row?.index}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: width, start: row.start }}
    >
      <Cell
        positionStatic
        {row}
        column={{ start: 0, size: width }}
        {atLeastOneSelected}
        {excludeMode}
        {rowActive}
        {...getCellProps(row)}
        on:inspect
        on:select-item
      />
    </StickyHeader>
  {/each}
</div>
