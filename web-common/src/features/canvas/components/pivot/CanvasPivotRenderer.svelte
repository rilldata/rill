<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { extractDimensionFiltersFromExpression } from "@rilldata/web-common/features/dashboards/pivot/pivot-filter-extraction";
  import {
    computePivotRowSelection,
    extractSelectionDimensionFilters,
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

  // Track last clicked cell for ring indicator
  let lastClickedCell: { rowId: string; columnId: string } | null = null;

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

  $: rowDimensionNames = pivotConfig
    ? ($pivotConfig?.rowDimensionNames ?? [])
    : [];

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

  // Click-to-filter: extract dimension filters for the clicked cell,
  // then batch-toggle them on the FilterState and apply to URL once.
  function handleCellClickToFilter(
    rowId: string,
    columnId: string,
    event: MouseEvent,
  ) {
    if (!pivotDataStore || !$pivotDataStore || !pivotConfig || !$pivotConfig)
      return;

    const isExclusive = event.metaKey || event.ctrlKey;

    // Compute filters for this cell directly (row + column dimensions)
    const cellFilters = getFiltersForCell(
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

    // Track clicked cell for ring indicator
    lastClickedCell = { rowId, columnId };

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

  // Clear clicked cell when the pivot has no self-applied filters
  $: if ($selfFilteredDimensions.size === 0) {
    lastClickedCell = null;
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
        clickedCell={lastClickedCell}
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
