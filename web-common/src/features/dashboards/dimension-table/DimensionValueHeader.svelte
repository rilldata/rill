<script lang="ts">
  import StickyHeader from "@rilldata/web-common/components/virtualized-table/core/StickyHeader.svelte";
  import { createEventDispatcher, getContext } from "svelte";
  import Cell from "../../../components/virtualized-table/core/Cell.svelte";
  import type {
    VirtualizedTableColumns,
    VirtualizedTableConfig,
  } from "../../../components/virtualized-table/types";
  import ArrowDown from "@rilldata/web-common/components/icons/ArrowDown.svelte";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import type { ResizeEvent } from "@rilldata/web-common/components/virtualized-table/drag-table-cell";
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

  const {
    actions: {
      sorting: { sortByDimensionValue },
    },
    selectors: {
      sorting: { sortedByDimensionValue, sortedAscending },
    },
  } = getStateManagers();
  const dispatch = createEventDispatcher();

  $: atLeastOneSelected = !!selectedIndex?.length;

  const getCellProps = (row: VirtualItem) => {
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
  const handleResize = (event: ResizeEvent) => {
    dispatch("resize-column", {
      size: event.detail.size,
      name,
    });
  };
</script>

<div
  class="sticky self-start left-6 top-0 z-20"
  style:height="{totalHeight}px"
  style:width="{20}px"
>
  <StickyHeader
    header={{ size: width, start: 0 }}
    enableResize={true}
    position="top-left"
    borderRight={horizontalScrolling}
    bgClass={$sortedByDimensionValue ? `bg-gray-50` : "bg-white"}
    on:click={sortByDimensionValue}
    on:keydown={sortByDimensionValue}
    on:resize={handleResize}
  >
    <div class="flex items-center">
      <span class={"px-1 " + $sortedByDimensionValue ? "font-bold" : ""}
        >{column?.label || column?.name}</span
      >
      {#if $sortedByDimensionValue}
        <div class="ui-copy-icon">
          {#if $sortedAscending}
            <div in:fly|global={{ duration: 200, y: -8 }} style:opacity={1}>
              <ArrowDown size="12px" />
            </div>
          {:else}
            <div in:fly|global={{ duration: 200, y: 8 }} style:opacity={1}>
              <ArrowDown transform="scale(1 -1)" size="12px" />
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
      bgClass={$sortedByDimensionValue ? `bg-gray-50` : "bg-white"}
    >
      <Cell
        positionStatic
        {row}
        column={{ start: 0, size: width }}
        {atLeastOneSelected}
        {excludeMode}
        {rowActive}
        {...getCellProps(row)}
        colSelected={$sortedByDimensionValue}
        on:inspect
        on:select-item
        label="Filter dimension value"
      />
    </StickyHeader>
  {/each}
</div>
