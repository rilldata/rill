<script lang="ts">
  import VirtualTooltip from "@rilldata/web-common/components/virtualized-table/VirtualTooltip.svelte";
  import FlatTable from "@rilldata/web-common/features/dashboards/pivot/FlatTable.svelte";
  import {
    getDimensionColumnProps,
    getMeasureColumnProps,
    getRowNestedLabel,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
  import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
  import {
    isElement,
    splitPivotChips,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import {
    type Cell,
    type ExpandedState,
    type SortingState,
    type TableOptions,
    createSvelteTable,
    getCoreRowModel,
    getExpandedRowModel,
  } from "@tanstack/svelte-table";
  import {
    createVirtualizer,
    defaultRangeExtractor,
  } from "@tanstack/svelte-virtual";
  import { onMount } from "svelte";
  import type { Readable } from "svelte/motion";
  import { derived } from "svelte/store";
  import NestedTable from "./NestedTable.svelte";
  import type {
    PivotDataRow,
    PivotDataStore,
    PivotDataStoreConfig,
    PivotState,
  } from "./types";

  // Distance threshold (in pixels) for triggering data fetch
  const ROW_THRESHOLD = 200;
  const OVERSCAN = 60;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;

  export let pivotDataStore: PivotDataStore;
  export let config: Readable<PivotDataStoreConfig>;
  export let pivotState: Readable<PivotState>;
  export let canShowDataViewer = false;
  export let border = true;
  export let setPivotExpanded: (expanded: ExpandedState) => void;
  export let setPivotSort: (sorting: SortingState) => void;
  export let setPivotRowPage: (page: number) => void;
  export let setPivotActiveCell:
    | ((rowId: string, columnId: string) => void)
    | undefined = undefined;

  const options: Readable<TableOptions<PivotDataRow>> = derived(
    [pivotDataStore, pivotState],
    ([pivotData, state]) => {
      let tableData = [...pivotData.data];
      if (pivotData.totalsRowData) {
        tableData = [pivotData.totalsRowData, ...pivotData.data];
      }
      return {
        data: tableData,
        columns: pivotData.columnDef,
        state: {
          expanded: state.expanded,
          sorting: state.sorting,
        },
        onExpandedChange: (updater) => {
          const expanded =
            typeof updater === "function" ? updater(state.expanded) : updater;
          setPivotExpanded(expanded);
        },
        getSubRows: (row) => row.subRows,
        onSortingChange: (updater) => {
          const sorting =
            typeof updater === "function" ? updater(state.sorting) : updater;
          setPivotSort(sorting);
        },
        getExpandedRowModel: getExpandedRowModel(),
        getCoreRowModel: getCoreRowModel(),
        enableSortingRemoval: false,
        enableExpanding: true,
      };
    },
  );

  $: table = createSvelteTable(options);

  let containerRefElement: HTMLDivElement;
  let stickyRows = [0];
  let rowScrollOffset = 0;
  let scrollLeft = 0;

  $: timeDimension = $config.time.timeDimension;
  $: hasRowDimension = $pivotState.rows.length > 0;
  $: hasColumnDimension =
    splitPivotChips($pivotState.columns).dimension.length > 0;
  $: reachedEndForRows = !!$pivotDataStore?.reachedEndForRowData;
  $: assembled = $pivotDataStore.assembled;
  $: dataRows = $pivotDataStore.data;
  $: totalsRow = $pivotDataStore.totalsRowData;
  $: isFlat = $config.isFlat;
  $: hasMeasureContextColumns = $config.enableComparison;

  $: measures = getMeasureColumnProps($config);
  $: rowDimensions = getDimensionColumnProps(
    $config.rowDimensionNames,
    $config,
  );
  $: rowDimensionLabel = getRowNestedLabel(rowDimensions);

  $: headerGroups = $table.getHeaderGroups();
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;

  $: rows = $table.getRowModel().rows;
  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLTableRowElement>({
    count: rows.length,
    getScrollElement: () => containerRefElement,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
    initialOffset: rowScrollOffset,
    rangeExtractor: (range) => {
      const next = new Set([...stickyRows, ...defaultRangeExtractor(range)]);

      return [...next].sort((a, b) => a - b);
    },
  });

  $: virtualRows = $virtualizer.getVirtualItems();
  $: totalRowSize = $virtualizer.getTotalSize();

  $: rowScrollOffset = $virtualizer?.scrollOffset || 0;

  // In this virtualization model, we create buffer rows before and after our real data
  // This maintains the "correct" scroll position when the user scrolls
  $: [before, after] = virtualRows.length
    ? [
        (virtualRows[1]?.start ?? virtualRows[0].start) - ROW_HEIGHT,
        totalRowSize - virtualRows[virtualRows.length - 1].end,
      ]
    : [0, 0];

  let customShortcuts: { description: string; shortcut: string }[] = [];
  $: if (canShowDataViewer) {
    customShortcuts = [
      { description: "View raw data for aggregated cell", shortcut: "Click" },
    ];
  }

  const handleScroll = (containerRefElement?: HTMLDivElement | null) => {
    if (containerRefElement) {
      if (hovering) hovering = null;
      const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
      const bottomEndDistance = scrollHeight - scrollTop - clientHeight;
      scrollLeft = containerRefElement.scrollLeft;

      const isReachingPageEnd = bottomEndDistance < ROW_THRESHOLD;
      const canFetchMoreData =
        !$pivotDataStore.isFetching && !reachedEndForRows;
      const hasMoreDataThanOnePage = rows.length >= NUM_ROWS_PER_PAGE;

      if (isReachingPageEnd && hasMoreDataThanOnePage && canFetchMoreData) {
        setPivotRowPage($pivotState.rowPage + 1);
      }
    }
  };

  onMount(() => {
    // wait for layout to be calculated
    requestAnimationFrame(() => {
      handleScroll(containerRefElement);
    });
  });

  let showTooltip = false;
  let hoverPosition: DOMRect;
  let hovering: HoveringData | null = null;
  let timer: ReturnType<typeof setTimeout>;

  type HoveringData = {
    value: string | number | null;
  };

  function handleCellClick(cell: Cell<PivotDataRow, unknown>) {
    if (!canShowDataViewer || !setPivotActiveCell) return;

    const value = cell.getValue();
    if (value === undefined) return;
    setPivotActiveCell(cell.row.id, cell.column.id);
  }

  function handleHover(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLElement;
    },
  ) {
    hoverPosition = e.currentTarget.getBoundingClientRect();
    const value = e.currentTarget.dataset.value;
    if (value === undefined) return;

    hovering = {
      value,
    };

    timer = setTimeout(() => {
      showTooltip = true;
    }, 250);
  }

  function handleLeave() {
    clearTimeout(timer);
    showTooltip = false;
    hovering = null;
  }

  function handleClick(e: MouseEvent) {
    if (!isElement(e.target)) return;

    const value = e.target.dataset.value;
    if (value === undefined) return;

    copyToClipboard(value);
  }
</script>

<div
  class:border
  class="table-wrapper relative"
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-height="{HEADER_HEIGHT}px"
  style:--total-header-height="{totalHeaderHeight + 1}px"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
>
  {#if isFlat}
    <FlatTable
      {headerGroups}
      {rows}
      {virtualRows}
      {measures}
      {totalsRow}
      {dataRows}
      {before}
      {after}
      {totalRowSize}
      {canShowDataViewer}
      {hasMeasureContextColumns}
      activeCell={$pivotState.activeCell}
      {assembled}
      onCellClick={handleCellClick}
      onCellHover={handleHover}
      onCellLeave={handleLeave}
      onCellCopy={handleClick}
    />
  {:else}
    <NestedTable
      {headerGroups}
      {rows}
      {virtualRows}
      {before}
      {after}
      {hasRowDimension}
      {timeDimension}
      {totalsRow}
      {totalRowSize}
      {rowDimensionLabel}
      {hasColumnDimension}
      {dataRows}
      {measures}
      {canShowDataViewer}
      activeCell={$pivotState.activeCell}
      {assembled}
      {scrollLeft}
      {containerRefElement}
      onCellClick={handleCellClick}
      onCellHover={handleHover}
      onCellLeave={handleLeave}
      onCellCopy={handleClick}
    />
  {/if}
</div>

{#if showTooltip && hovering}
  <VirtualTooltip
    sortable={true}
    {hovering}
    {hoverPosition}
    pinned={false}
    {customShortcuts}
  />
{/if}

<style lang="postcss">
  .table-wrapper {
    @apply overflow-auto h-fit max-h-full w-fit max-w-full;
    @apply rounded-md z-40;
  }
</style>
