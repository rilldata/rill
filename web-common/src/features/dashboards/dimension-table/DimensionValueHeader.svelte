<script lang="ts">
  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
  import { getContext } from "svelte";
  import Cell from "../../../components/virtualized-table/core/Cell.svelte";
  import type { VirtualizedTableConfig } from "../../../components/virtualized-table/types";

  const config: VirtualizedTableConfig = getContext("config");

  export let totalHeight: number;
  export let virtualRowItems;
  export let selectedIndex = [];
  export let column: VirtualizedTableColumns;
  export let rows;
  export let width = config.indexWidth;
  export let horizontalScrolling;

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
    borderRight={horizontalScrolling}
  >
    <span class="px-1">{column.label || column.name}</span>
  </StickyHeader>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const rowActive = activeIndex === row?.index}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: width, start: row.start }}
      borderRight={horizontalScrolling}
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
