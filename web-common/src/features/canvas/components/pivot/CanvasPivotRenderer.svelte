<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { extractDimensionFiltersFromExpression } from "@rilldata/web-common/features/dashboards/pivot/pivot-filter-extraction";
  import {
    cellKey,
    computePivotRowSelection,
    createEmptyClickSelectionState,
    extractSelectionDimensionFilters,
    getFiltersForRowHeader,
    type PivotClickSelectionState,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection";
  import {
    getFiltersForCell,
    splitPivotChips,
  } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import PivotTable from "@rilldata/web-common/features/dashboards/pivot/PivotTable.svelte";
  import {
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { PivotCanvasComponent } from "./index";

  import { derived, get, type Readable, type Writable } from "svelte/store";

  export let schema: {
    isValid: boolean;
    error?: string;
  };
  export let pivotDataStore: PivotDataStore | undefined;
  export let pivotConfig: Readable<PivotDataStoreConfig> | undefined;
  export let pivotState: Writable<PivotState>;
  export let hasHeader = false;
  export let component: PivotCanvasComponent;

  // Track click selections: row-header clicks and data-cell clicks
  let clickSelection: PivotClickSelectionState =
    createEmptyClickSelectionState();

  $: pivotColumns = splitPivotChips($pivotState.columns);

  $: hasColumnAndNoMeasure =
    pivotColumns.dimension.length > 0 && pivotColumns.measure.length === 0;

  // FilterManager and metrics view for filter application
  $: filterManager = component.parent.filterManager;
  $: spec = component.specStore;
  $: metricsViewName = $spec?.metrics_view;
  $: selfFilteredDimensions = component.selfFilteredDimensions;

  // Reactively compute row selection state from the current filters.
  // filterMapStore gives us Map<metricsViewName, V1Expression> that updates on URL changes.
  $: whereFilterStore = derived(filterManager.filterMapStore, (filterMap) => {
    return metricsViewName ? filterMap.get(metricsViewName) : undefined;
  });

  // Prune selfFilteredDimensions: remove any dimension whose filter was cleared
  // (e.g., by toggling off all values, or clearing the filter pill externally).
  // Uses get() to read selfFilteredDimensions non-reactively so this block
  // only re-runs when whereFilterStore changes (not when we add dimensions).
  $: {
    const currentFilter = $whereFilterStore;
    const activeDims = new Set(
      currentFilter?.cond?.exprs
        ?.map((e) => e.cond?.exprs?.[0]?.ident)
        .filter(Boolean) ?? [],
    );
    const dims = get(selfFilteredDimensions);
    let changed = false;
    for (const dim of dims) {
      if (!activeDims.has(dim)) {
        dims.delete(dim);
        changed = true;
      }
    }
    if (changed) {
      selfFilteredDimensions.set(new Set(dims));
    }
  }

  // Only compute selection state for dimensions the pivot itself filtered
  $: dimensionFilterMap = extractSelectionDimensionFilters($whereFilterStore, [
    ...$selfFilteredDimensions,
  ]);

  $: rowSelectionState =
    pivotDataStore && $pivotDataStore?.data && pivotConfig && $pivotConfig
      ? computePivotRowSelection(
          $pivotConfig,
          $pivotDataStore.data,
          dimensionFilterMap,
        )
      : undefined;

  // Click-to-filter: extract dimension filters for the clicked cell or row header,
  // then batch-toggle them on the FilterState and apply to URL once.
  function handleCellClickToFilter(
    rowId: string,
    columnId: string,
    isRowHeader: boolean,
    event: MouseEvent,
  ) {
    if (!pivotDataStore || !$pivotDataStore || !pivotConfig || !$pivotConfig)
      return;

    const isExclusive = event.metaKey || event.ctrlKey;

    // Row header click: only row dimension filters
    // Data cell click: row + column dimension filters
    const cellFilters = isRowHeader
      ? getFiltersForRowHeader($pivotConfig, rowId, $pivotDataStore.data)
      : getFiltersForCell(
          $pivotConfig,
          rowId,
          columnId,
          $pivotDataStore.columnDimensionAxes ?? {},
          $pivotDataStore.data,
        );

    if (!cellFilters.filters) return;

    // Extract dimension name/value pairs from the filter expression
    const dimensionFilters = extractDimensionFiltersFromExpression(
      cellFilters.filters,
    );

    if (dimensionFilters.length === 0) return;

    // Get the FilterState for this metrics view to batch updates
    const filterClass = filterManager.metricsViewFilters.get(metricsViewName);
    if (!filterClass) return;

    // Clear temporary filter status for all dimensions being toggled
    dimensionFilters.forEach(({ dimensionName }) => {
      filterManager.checkTemporaryFilter(dimensionName, [metricsViewName]);
    });

    // Toggle each dimension's values on FilterState (no URL update per call)
    let filterString: string | null = null;
    dimensionFilters.forEach(({ dimensionName, values }) => {
      filterString = filterClass.toggleDimensionValueSelections(
        dimensionName,
        values,
        false, // keepPillVisible
        isExclusive,
      );
    });

    // Track click selection for visual highlighting.
    // Exclusive mode clears all previous selections; normal mode accumulates.
    const nextRowHeaders = isExclusive
      ? new Set<string>()
      : new Set(clickSelection.rowHeaderSelections);
    const nextCells = isExclusive
      ? new Set<string>()
      : new Set(clickSelection.cellSelections);

    if (isRowHeader) {
      // Toggle: if already selected, the filter toggle above removed it
      if (nextRowHeaders.has(rowId)) {
        nextRowHeaders.delete(rowId);
      } else {
        nextRowHeaders.add(rowId);
      }
    } else {
      const key = cellKey(rowId, columnId);
      if (nextCells.has(key)) {
        nextCells.delete(key);
      } else {
        nextCells.add(key);
      }
    }

    const hasAny = nextRowHeaders.size > 0 || nextCells.size > 0;
    clickSelection = {
      rowHeaderSelections: nextRowHeaders,
      cellSelections: nextCells,
      hasAnySelection: hasAny,
      isRowHeaderSelected: (rid) => nextRowHeaders.has(rid),
      isCellSelected: (rid, cid) => nextCells.has(cellKey(rid, cid)),
    };

    // Mark these dimensions as self-filtered by the pivot
    selfFilteredDimensions.update((dims) => {
      const next = new Set(dims);
      dimensionFilters.forEach(({ dimensionName }) => next.add(dimensionName));
      return next;
    });

    // Single batch URL update with the final filter string
    if (filterString !== null) {
      filterManager.applyFiltersToUrl(
        new Map([[metricsViewName, filterString]]),
      );
    }
  }

  // Clear selection state when the pivot has no self-applied filters
  $: if ($selfFilteredDimensions.size === 0) {
    clickSelection = createEmptyClickSelectionState();
  }
</script>

<div
  class="size-full overflow-hidden"
  style:max-height="inherit"
  class:p-4={hasHeader}
  class:pt-1={hasHeader}
>
  {#if !schema.isValid}
    <ComponentError error={schema.error} />
  {:else if pivotDataStore && $pivotDataStore && pivotConfig && $pivotConfig}
    {#if $pivotDataStore?.error?.length}
      <PivotError errors={$pivotDataStore.error} />
    {:else if !$pivotDataStore?.data || $pivotDataStore?.data?.length === 0}
      <PivotEmpty
        assembled={$pivotDataStore.assembled}
        isFetching={$pivotDataStore.isFetching}
        {hasColumnAndNoMeasure}
      />
    {:else}
      <PivotTable
        border={hasHeader}
        rounded={hasHeader}
        {pivotDataStore}
        config={pivotConfig}
        {pivotState}
        {rowSelectionState}
        {clickSelection}
        enableClickToFilter
        setPivotExpanded={(expanded) => {
          pivotState.update((state) => ({
            ...state,
            expanded,
          }));
        }}
        setPivotSort={(sorting) => {
          pivotState.update((state) => ({
            ...state,
            sorting,
            rowPage: 1,
            expanded: {},
          }));
        }}
        setPivotRowPage={(page) => {
          pivotState.update((state) => ({
            ...state,
            rowPage: page,
          }));
        }}
        onCellClickToFilter={handleCellClickToFilter}
      />
    {/if}
  {/if}
</div>
