<script lang="ts" context="module">
  import { writable } from "svelte/store";
  const measureLengths = writable(new Map<string, number>());
</script>

<script lang="ts">
  import VirtualTooltip from "@rilldata/web-common/components/virtualized-table/VirtualTooltip.svelte";
  import FlatTable from "@rilldata/web-common/features/dashboards/pivot/FlatTable.svelte";
  import { getMeasureColumnProps } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
  import {
    calculateColumnWidth,
    calculateMeasureWidth,
    COLUMN_WIDTH_CONSTANTS as WIDTHS,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
  import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
  import { isElement } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
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

  // Import isMeasureColumn from FlatTable
  function isMeasureColumn(header: any, colNumber: number) {
    return !header.column.columns && header.column.columnDef.id;
  }

  // Distance threshold (in pixels) for triggering data fetch
  const ROW_THRESHOLD = 200;
  const OVERSCAN = 60;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;

  export let pivotDataStore: PivotDataStore;
  export let config: Readable<PivotDataStoreConfig>;
  export let pivotState: Readable<PivotState>;
  export let canShowDataViewer = false;
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
  let initialMeasureIndexOnResize = 0;
  let initLengthOnResize = 0;
  let initScrollOnResize = 0;
  let percentOfChangeDuringResize = 0;
  let resizingMeasure = false;
  let resizing = false;

  $: timeDimension = $config.time.timeDimension;
  $: hasDimension = $pivotState.rows.dimension.length > 0;
  $: hasColumnDimension = $pivotState.columns.dimension.length > 0;
  $: reachedEndForRows = !!$pivotDataStore?.reachedEndForRowData;
  $: assembled = $pivotDataStore.assembled;
  $: dataRows = $pivotDataStore.data;
  $: totalsRow = $pivotDataStore.totalsRowData;
  $: isFlat = $config.isFlat;

  $: measures = getMeasureColumnProps($config);
  $: measureCount = measures.length;
  $: measures.forEach(({ name, label, formatter }) => {
    if (!$measureLengths.has(name)) {
      const estimatedWidth = calculateMeasureWidth(
        name,
        label,
        formatter,
        totalsRow,
        dataRows,
      );
      measureLengths.update((measureLengths) => {
        return measureLengths.set(name, estimatedWidth);
      });
    }
  });

  // For flat tables, calculate widths for all columns including dimensions
  $: if (isFlat && headerGroups.length > 0) {
    headerGroups[0].headers.forEach((header) => {
      const columnDef = header.column.columnDef;
      const name = String(columnDef.header);

      if (!$measureLengths.has(name)) {
        let estimatedWidth;
        if (isMeasureColumn(header, 0)) {
          // For measures, use measure width calculation
          const measure = measures.find((m) => m.name === columnDef.id);
          if (measure) {
            estimatedWidth = calculateMeasureWidth(
              measure.name,
              measure.label,
              measure.formatter,
              totalsRow,
              dataRows,
            );
          }
        } else {
          // For dimensions, use column width calculation
          estimatedWidth = calculateColumnWidth(name, timeDimension, dataRows);
        }

        if (estimatedWidth) {
          measureLengths.update((lengths) => lengths.set(name, estimatedWidth));
        }
      }
    });
  }

  $: subHeaders = [
    {
      subHeaders: measures.map((m) => ({
        column: { columnDef: { name: m.name } },
      })),
    },
  ];

  let measureGroups: {
    subHeaders: { column: { columnDef: { name: string } } }[];
  }[];
  // @ts-expect-error - I have manually added the name property in pivot-column-definition.ts
  $: measureGroups =
    headerGroups[headerGroups.length - 2]?.headers?.slice(
      hasDimension ? 1 : 0,
    ) ?? subHeaders;

  $: measureGroupsLength = measureGroups.length;
  $: totalMeasureWidth = measures.reduce(
    (acc, { name }) => acc + ($measureLengths.get(name) ?? 0),
    0,
  );
  $: totalLength = measureGroupsLength * totalMeasureWidth;

  $: headerGroups = $table.getHeaderGroups();
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;
  $: headers = headerGroups[0].headers;
  $: firstColumnName = hasDimension
    ? String(headers[0]?.column.columnDef.header)
    : null;
  $: firstColumnWidth =
    hasDimension && firstColumnName
      ? calculateColumnWidth(firstColumnName, timeDimension, dataRows)
      : 0;

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

  $: if (resizingMeasure && containerRefElement && measureLengths) {
    containerRefElement.scrollTo({
      left:
        initScrollOnResize +
        percentOfChangeDuringResize * (totalLength - initLengthOnResize),
    });
  }

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
        console.log("setPivotRowPage", $pivotState.rowPage + 1);
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

  function onResizeStart(e: MouseEvent) {
    initLengthOnResize = totalLength;
    initScrollOnResize = scrollLeft;

    const offset =
      e.clientX -
      containerRefElement.getBoundingClientRect().left -
      firstColumnWidth -
      measures.reduce((rollingSum, { name }, i) => {
        return i <= initialMeasureIndexOnResize
          ? rollingSum + ($measureLengths.get(name) ?? 0)
          : rollingSum;
      }, 0) +
      4;

    percentOfChangeDuringResize = (scrollLeft + offset) / totalLength;
  }

  let showTooltip = false;
  let hoverPosition: DOMRect;
  let hovering: HoveringData | null = null;
  let timer: ReturnType<typeof setTimeout>;

  type HoveringData = {
    value: string | number | null;
  };

  function handleCellClick(cell: Cell<PivotDataRow, unknown>) {
    if (!canShowDataViewer || !setPivotActiveCell) return;
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
  class="table-wrapper relative"
  class:with-row-dimension={!isFlat && hasDimension}
  class:with-col-dimension={hasColumnDimension}
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-height="{HEADER_HEIGHT}px"
  style:--total-header-height="{totalHeaderHeight + headerGroups.length}px"
  bind:this={containerRefElement}
  on:scroll={() => handleScroll(containerRefElement)}
  class:pointer-events-none={resizing}
>
  <div
    class="w-full absolute top-0 z-50 flex pointer-events-none"
    style:width="{totalLength + firstColumnWidth}px"
    style:height="{totalRowSize + totalHeaderHeight + headerGroups.length}px"
  >
    <div
      style:width="{firstColumnWidth}px"
      class="sticky left-0 flex-none flex"
    >
      <Resizer
        side="right"
        direction="EW"
        min={WIDTHS.MIN_COL_WIDTH}
        max={WIDTHS.MAX_COL_WIDTH}
        dimension={firstColumnWidth}
        onUpdate={(d) => (firstColumnWidth = d)}
        onMouseDown={(e) => {
          resizingMeasure = false;
          resizing = true;
          onResizeStart(e);
        }}
        onMouseUp={() => {
          resizing = false;
          resizingMeasure = false;
        }}
      >
        <div class="resize-bar" />
      </Resizer>
    </div>

    {#each measureGroups as { subHeaders }, groupIndex (groupIndex)}
      <div class="h-full z-50 flex" style:width="{totalMeasureWidth}px">
        {#each subHeaders as { column: { columnDef: { name } } }, i (name)}
          {@const length =
            $measureLengths.get(name) ?? WIDTHS.INIT_MEASURE_WIDTH}
          {@const last =
            i === subHeaders.length - 1 &&
            groupIndex === measureGroups.length - 1}
          <div style:width="{length}px" class="h-full relative">
            <Resizer
              side="right"
              direction="EW"
              min={WIDTHS.MIN_MEASURE_WIDTH}
              max={WIDTHS.MAX_MEASURE_WIDTH}
              dimension={length}
              justify={last ? "end" : "center"}
              hang={!last}
              onUpdate={(d) => {
                measureLengths.update((measureLengths) => {
                  return measureLengths.set(name, d);
                });
              }}
              onMouseDown={(e) => {
                resizingMeasure = true;
                resizing = true;
                initialMeasureIndexOnResize = i;
                onResizeStart(e);
              }}
              onMouseUp={() => {
                resizing = false;
                resizingMeasure = false;
              }}
            >
              <div class="resize-bar" />
            </Resizer>
          </div>
        {/each}
      </div>
    {/each}
  </div>

  {#if isFlat}
    <FlatTable
      {headerGroups}
      {rows}
      {virtualRows}
      {before}
      {after}
      {measureCount}
      {canShowDataViewer}
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
      {firstColumnWidth}
      {firstColumnName}
      {totalLength}
      {measureCount}
      {measureGroups}
      measureLengths={$measureLengths}
      {canShowDataViewer}
      activeCell={$pivotState.activeCell}
      {assembled}
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
    @apply border rounded-md z-40;
  }

  .resize-bar {
    @apply bg-primary-500 w-1 h-full;
  }
</style>
