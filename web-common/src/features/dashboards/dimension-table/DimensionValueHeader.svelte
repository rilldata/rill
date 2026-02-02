<script lang="ts">
  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import { getContext } from "svelte";
  import Cell from "../../../components/virtualized-table/core/Cell.svelte";
  import type {
    VirtualizedTableColumns,
    VirtualizedTableConfig,
  } from "../../../components/virtualized-table/types";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { VirtualItem } from "@tanstack/svelte-virtual";
  import type { DimensionTableRow } from "./dimension-table-types";

  const config: VirtualizedTableConfig = getContext("config");

  export let totalHeight: number;
  export let virtualRowItems: VirtualItem[];
  export let selectedIndex: number[] = [];
  export let column: VirtualizedTableColumns;
  export let rows: DimensionTableRow[];
  export let width = config.indexWidth;
  export let horizontalScrolling: boolean;

  // Cell props
  export let scrolling: boolean;
  export let activeIndex: number;
  export let excludeMode = false;
  export let onSelectItem: (data: {
    index: number;
    meta: boolean;
  }) => void = () => {};
  export let onResizeColumn: (size: number) => void = () => {};
  export let onInspect: (rowIndex: number) => void = () => {};

  const {
    actions: {
      sorting: { sortByDimensionValue },
    },
    selectors: {
      sorting: { sortedByDimensionValue, sortedAscending },
    },
  } = getStateManagers();

  $: atLeastOneSelected = !!selectedIndex?.length;

  const getCellProps = (row: VirtualItem, selectedIndex: number[]) => {
    const value = rows[row.index]?.[column.name];
    return {
      value,
      // NOTE: for this "header" column, we don't use a
      // formatted value, we use the dimension value
      // directly. Thus, we pass `null` as the formatted.
      formattedValue: null,
      type: column?.type,
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
    enableResize={true}
    position="top-left"
    borderRight={true}
    bgClass="bg-surface-background"
    onClick={sortByDimensionValue}
    onResize={onResizeColumn}
  >
    <div class="flex items-center">
      <span class:font-bold={"px-1 " + $sortedByDimensionValue}
        >{column?.label || column?.name}</span
      >
      {#if $sortedByDimensionValue}
        <div class="text-fg-secondary">
          {#if $sortedAscending}
            <div in:fly|global={{ duration: 200, y: -8 }} style:opacity={1}>
              <ArrowDown size="12px" />
            </div>
          {:else}
            <div in:fly|global={{ duration: 200, y: 8 }} style:opacity={1}>
              <ArrowDown flip size="12px" />
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </StickyHeader>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const rowActive = activeIndex === row?.index}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: width, start: row.start }}
      borderRight={horizontalScrolling}
      bgClass="bg-surface-background"
    >
      <Cell
        label="Filter dimension value"
        positionStatic
        {row}
        column={{ start: 0, size: width }}
        {atLeastOneSelected}
        {excludeMode}
        {rowActive}
        {...getCellProps(row, selectedIndex)}
        colSelected={$sortedByDimensionValue}
        {onInspect}
        {onSelectItem}
      />
    </StickyHeader>
  {/each}
</div>
