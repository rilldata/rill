<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import VirtualTooltip from "@rilldata/web-common/components/virtualized-table/VirtualTooltip.svelte";
  import FlatTable from "@rilldata/web-common/features/dashboards/pivot/FlatTable.svelte";
  import type { PivotClickSelectionState } from "@rilldata/web-common/features/dashboards/pivot/pivot-click-selection";
  import {
    getDimensionColumnProps,
    getMeasureColumnProps,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
  import {
    getNextRowLimit,
    SHOW_MORE_BUTTON,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
  import { NUM_ROWS_PER_PAGE } from "@rilldata/web-common/features/dashboards/pivot/pivot-infinite-scroll";
  import type { PivotRowSelectionState } from "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection";
  import {
    isElement,
    isShowMoreRow,
    splitPivotChips,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import {
    createVirtualizer,
    defaultRangeExtractor,
  } from "@tanstack/svelte-virtual";
  import { onMount } from "svelte";
  import type { Readable } from "svelte/store";
  import { derived } from "svelte/store";
  import {
    type ExpandedState,
    type Row,
    type SortingState,
    type TableOptions,
    createSvelteTable,
    getCoreRowModel,
    getExpandedRowModel,
  } from "tanstack-table-8-svelte-5";
  import NestedTable from "./NestedTable.svelte";
  import {
    type CellFormatter,
    computeMeasureDomains,
    makeCellFormatter,
  } from "./pivot-conditional-formatting";
  import type {
    PivotDataRow,
    PivotDataStore,
    PivotDataStoreConfig,
    PivotState,
  } from "./types";

  // Distance threshold (in pixels) for triggering data fetch
  const ROW_THRESHOLD = 200;
  const ROW_HEIGHT = 24;
  const HEADER_HEIGHT = 30;

  export let pivotDataStore: PivotDataStore;
  export let widthScopeKey: string;
  export let config: Readable<PivotDataStoreConfig>;
  export let pivotState: Readable<PivotState>;
  export let canShowDataViewer = false;
  export let border = true;
  export let overscan = 20;
  export let rounded = true;
  export let fillWidth = false;
  export let setPivotExpanded: (expanded: ExpandedState) => void;
  export let setPivotSort: (sorting: SortingState) => void;
  export let setPivotRowPage: (page: number) => void;
  export let setPivotActiveCell:
    | ((rowId: string, columnId: string) => void)
    | undefined = undefined;
  export let setPivotOutermostRowLimit: ((limit: number) => void) | undefined =
    undefined;
  export let setPivotRowLimitForExpanded:
    | ((expandIndex: string, limit: number) => void)
    | undefined = undefined;
  export let onCellClickToFilter:
    | ((
        rowId: string,
        columnId: string,
        isRowHeader: boolean,
        rowData: PivotDataRow,
      ) => void)
    | undefined = undefined;
  export let onColumnHeaderClick:
    | ((dimensionPath: Record<string, string>) => void)
    | undefined = undefined;
  export let enableClickToFilter = false;
  export let rowSelectionState: PivotRowSelectionState | undefined = undefined;
  export let clickSelection: PivotClickSelectionState | undefined = undefined;

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
  let containerWidth = 0;
  let stickyRows: number[] = [];
  let rowScrollOffset = 0;
  let scrollLeft = 0;
  let timeout: ReturnType<typeof setTimeout>;
  let leftCell = true;
  let ignoreInitialTimeout = false;

  $: timeDimension = $config.time.timeDimension;
  $: hasColumnDimension =
    splitPivotChips($pivotState.columns).dimension.length > 0;
  $: reachedEndForRows = !!$pivotDataStore?.reachedEndForRowData;
  $: assembled = $pivotDataStore.assembled;
  $: dataRows = $pivotDataStore.data;
  $: totalsRow = $pivotDataStore.totalsRowData;
  $: stickyRows = totalsRow ? [0] : [];
  $: isFlat = $config.isFlat;
  $: hasMeasureContextColumns = $config.enableComparison;

  $: measures = getMeasureColumnProps($config);
  $: rowDimensions = getDimensionColumnProps(
    $config.rowDimensionNames,
    $config,
  );

  // Per-measure conditional formatting. Domains prefer leaf data cells so
  // aggregate magnitudes don't dominate the gradient, falling back to nested
  // parent rows when no leaves are present (e.g. collapsed nested rows). The
  // grand-totals row and the row-totals column are always excluded. Recomputes
  // when the row model or the formatting config changes.
  $: cellFormatters = buildCellFormatters(
    $table.getRowModel().flatRows,
    $config.pivot.measureFormatting,
    !!totalsRow,
    $config.allMeasures,
  );

  $: headerGroups = $table.getHeaderGroups();
  $: totalHeaderHeight = headerGroups.length * HEADER_HEIGHT;

  $: rows = $table.getRowModel().rows;
  $: virtualizer = createVirtualizer<HTMLDivElement, HTMLTableRowElement>({
    count: rows.length,
    getScrollElement: () => containerRefElement,
    estimateSize: () => ROW_HEIGHT,
    overscan,
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

  $: clickToFilterEnabled = enableClickToFilter && !!onCellClickToFilter;
  function getCustomShortcuts(rowHeader: boolean) {
    if (clickToFilterEnabled) {
      return [
        { description: m.dashboard_filter_by_value(), shortcut: "Click" },
      ];
    }
    if (canShowDataViewer && !rowHeader) {
      return [
        {
          description: m.dashboard_view_raw_aggregated_data(),
          shortcut: "Click",
        },
      ];
    }
    return [];
  }

  function buildCellFormatters(
    flatRows: Row<PivotDataRow>[],
    measureFormatting: PivotState["measureFormatting"],
    hasTotalsRow: boolean,
    allMeasures: PivotDataStoreConfig["allMeasures"],
  ): Map<string, CellFormatter> {
    const formatters = new Map<string, CellFormatter>();
    if (!measureFormatting || Object.keys(measureFormatting).length === 0) {
      return formatters;
    }

    // Rules formatters don't depend on the value domain, so only scan the row
    // model when a scale-mode (heatmap/data bar) measure needs one.
    const needsDomains = Object.values(measureFormatting).some(
      (fmt) => fmt.mode !== "rules",
    );
    // Collect values from leaf rows and, separately, from nested parent rows.
    // Leaf values are preferred for the domain so aggregate magnitudes don't
    // dominate the gradient. But nested parent rows also render formatted
    // measure cells, and when they're collapsed the leaf rows aren't in the row
    // model at all: without a parent fallback the domain would be empty and no
    // heatmap/data-bar formatter would be built until the user expands a row.
    const leafValues: { measureName: string; value: number }[] = [];
    const parentValues: { measureName: string; value: number }[] = [];
    if (needsDomains) {
      for (const row of flatRows) {
        // Always skip the prepended grand-totals row.
        if (hasTotalsRow && row.id === "0") continue;
        const target = row.subRows.length > 0 ? parentValues : leafValues;
        for (const cell of row.getAllCells()) {
          const meta = cell.column.columnDef.meta;
          if (
            !meta?.conditionalFormat ||
            meta.isRowTotal ||
            !meta.measureName
          ) {
            continue;
          }
          const value = cell.getValue();
          if (typeof value === "number") {
            target.push({ measureName: meta.measureName, value });
          }
        }
      }
    }

    const leafDomains = computeMeasureDomains(leafValues);
    const parentDomains = computeMeasureDomains(parentValues);
    for (const [measureName, fmt] of Object.entries(measureFormatting)) {
      if (fmt.mode === "rules") {
        formatters.set(measureName, makeCellFormatter(fmt));
        continue;
      }
      const domain =
        leafDomains.get(measureName) ?? parentDomains.get(measureName);
      if (domain) {
        // Heatmap gradients flip for lower-is-better measures so low values
        // get the "good" end of the scheme.
        const lowerIsBetter =
          allMeasures.find((m) => m.name === measureName)?.lowerIsBetter ??
          false;
        formatters.set(
          measureName,
          makeCellFormatter(fmt, domain, lowerIsBetter),
        );
      }
    }
    return formatters;
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

  let hoverPosition: DOMRect;
  let hovering: HoveringData | null = null;

  type HoveringData = {
    value: string | number | null;
    rowHeader: boolean;
  };

  function onCellClick(e: MouseEvent) {
    if (!(e.target instanceof HTMLElement)) return;

    // Find the closest td element that has the data attributes
    const td = e.target.closest("td");
    if (!td || !(td instanceof HTMLElement)) return;

    const rowId = td.dataset.rowid;
    const columnId = td.dataset.columnid;
    const rowHeader = td.dataset.rowheader === "true";

    if (rowId === undefined || columnId === undefined) return;

    const row = $table.getRow(rowId);
    if (!row) return;

    // "Show More" rows only respond to a row-header click, which increases the
    // limit. All other cells in that row are inert (no filtering / active cell).
    if (isShowMoreRow(row)) {
      if (!rowHeader) return;

      const rowData = row.original;
      const currentLimit = rowData.__currentLimit as number;
      const nextLimit = getNextRowLimit(currentLimit);

      if (!nextLimit) return;

      // Check if this is the outermost dimension or a nested dimension
      // Outermost dimension has rowId like "0", "1", etc. (no dots)
      // Nested dimensions have rowId like "0.1", "0.1.2", etc.
      const isOutermostDimension = !rowId.includes(".");

      if (isOutermostDimension) {
        // Handle outermost dimension "Show more" click
        if (setPivotOutermostRowLimit) {
          setPivotOutermostRowLimit(nextLimit);
        }
      } else {
        // Handle nested dimension "Show more" click
        const expandIndex = rowId.split(".").slice(0, -1).join(".");
        if (expandIndex && setPivotRowLimitForExpanded) {
          setPivotRowLimitForExpanded(expandIndex, nextLimit);
        }
      }
      return;
    }

    if (rowHeader && !onCellClickToFilter) {
      // Legacy expand behavior: when no filter callback, row headers expand/collapse.
      // When onCellClickToFilter is present, expand is handled by the chevron icon
      // in PivotExpandableCell (stopPropagation), so row header clicks filter instead.
      if (row.getCanExpand()) row.getToggleExpandedHandler()();
    } else {
      // Skip totals row for filtering
      const isTotalsRow = totalsRow && rowId === "0";
      if (isTotalsRow && onCellClickToFilter) {
        return;
      }

      // Set active cell for Explore data viewer (if enabled)
      if (setPivotActiveCell && canShowDataViewer) {
        setPivotActiveCell(rowId, columnId);
      }

      // Apply click-to-filter (Canvas or future Explore)
      if (onCellClickToFilter) {
        onCellClickToFilter(rowId, columnId, rowHeader, row.original);
      }
    }
  }

  function onTableLeave() {
    clearTimeout(timeout);

    hovering = null;
    ignoreInitialTimeout = false;
  }

  function handleClick(e: MouseEvent) {
    if (!isElement(e.target)) return;

    // Row dimension cells render an inner component (PivotExpandableCell), so the
    // event target is a child element. Resolve the cell that carries the data attributes.
    const td = e.target.closest("td");
    const value = td?.dataset.value;
    if (value === undefined || value === SHOW_MORE_BUTTON) return;

    copyToClipboard(value);
  }

  function onMouseMove(
    e: MouseEvent & {
      target: EventTarget & Window;
    },
  ) {
    clearTimeout(timeout);

    if (ignoreInitialTimeout) {
      handleTooltip(e);
      return;
    } else {
      timeout = setTimeout(() => {
        handleTooltip(e);
      }, 400);
    }
  }

  function handleTooltip(e: MouseEvent) {
    // Element is not a cell or we haven't left the cell for the current tooltip
    if (!leftCell || !(e.target instanceof HTMLElement)) return;

    // Row dimension cells render an inner component (PivotExpandableCell), so the
    // event target is a child element. Resolve the cell that carries the data attributes.
    const td = e.target.closest("td");
    const value = td?.dataset.value;

    if (!td || value === undefined) return;

    if (value === SHOW_MORE_BUTTON) return;

    leftCell = false;
    td.addEventListener("mouseleave", () => (leftCell = true), {
      once: true,
    });

    ignoreInitialTimeout = true;

    hovering = {
      value,
      rowHeader: td.dataset.rowheader === "true",
    };
    hoverPosition = td.getBoundingClientRect();
  }
</script>

<div
  class:border
  class:rounded-sm={rounded}
  class="table-wrapper relative"
  style:--row-height="{ROW_HEIGHT}px"
  style:--header-height="{HEADER_HEIGHT}px"
  style:--total-header-height="{totalHeaderHeight + 1}px"
  bind:this={containerRefElement}
  bind:clientWidth={containerWidth}
  class:w-full={fillWidth}
  class:w-fit={!fillWidth}
  onscroll={() => handleScroll(containerRefElement)}
>
  {#if isFlat}
    <FlatTable
      {headerGroups}
      {rows}
      {virtualRows}
      {measures}
      {cellFormatters}
      {totalsRow}
      {dataRows}
      {before}
      {after}
      {totalRowSize}
      {canShowDataViewer}
      {enableClickToFilter}
      {hasMeasureContextColumns}
      {rowSelectionState}
      {clickSelection}
      config={$config}
      activeCell={$pivotState.activeCell}
      {assembled}
      {onMouseMove}
      {onCellClick}
      {onTableLeave}
      {fillWidth}
      {containerWidth}
      onCellCopy={handleClick}
    />
  {:else}
    <NestedTable
      {headerGroups}
      {rows}
      {virtualRows}
      {widthScopeKey}
      {before}
      {after}
      {timeDimension}
      {totalsRow}
      {totalRowSize}
      {rowDimensions}
      {hasColumnDimension}
      {dataRows}
      {measures}
      {cellFormatters}
      {canShowDataViewer}
      {enableClickToFilter}
      {rowSelectionState}
      {clickSelection}
      {onColumnHeaderClick}
      activeCell={$pivotState.activeCell}
      {assembled}
      {scrollLeft}
      {containerRefElement}
      {onMouseMove}
      {onCellClick}
      {onTableLeave}
      {fillWidth}
      {containerWidth}
      onCellCopy={handleClick}
    />
  {/if}
</div>

{#if hovering}
  <VirtualTooltip
    sortable
    {hovering}
    {hoverPosition}
    pinned={false}
    customShortcuts={getCustomShortcuts(hovering.rowHeader)}
  />
{/if}

<style lang="postcss">
  .table-wrapper {
    @apply overflow-auto h-fit max-h-full max-w-full;
    @apply z-40 select-none;
  }
</style>
