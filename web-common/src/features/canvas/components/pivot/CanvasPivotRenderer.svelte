<script lang="ts">
  import ComponentError from "@rilldata/web-common/features/components/ComponentError.svelte";
  import { extractDimensionFiltersFromExpression } from "@rilldata/web-common/features/dashboards/pivot/pivot-filter-extraction";
  import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
  import PivotEmpty from "@rilldata/web-common/features/dashboards/pivot/PivotEmpty.svelte";
  import PivotError from "@rilldata/web-common/features/dashboards/pivot/PivotError.svelte";
  import PivotTable from "@rilldata/web-common/features/dashboards/pivot/PivotTable.svelte";
  import {
    type PivotDataStore,
    type PivotDataStoreConfig,
    type PivotState,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import type { PivotCanvasComponent } from "./index";

  import { tick } from "svelte";
  import type { Readable, Writable } from "svelte/store";

  export let schema: {
    isValid: boolean;
    error?: string;
  };
  export let pivotDataStore: PivotDataStore | undefined;
  export let pivotConfig: Readable<PivotDataStoreConfig> | undefined;
  export let pivotState: Writable<PivotState>;
  export let hasHeader = false;
  export let component: PivotCanvasComponent;

  $: pivotColumns = splitPivotChips($pivotState.columns);

  $: hasColumnAndNoMeasure =
    pivotColumns.dimension.length > 0 && pivotColumns.measure.length === 0;

  // Extract FilterManager and metrics view for filter application
  $: filterManager = component.parent.filterManager;
  $: spec = component.specStore;
  $: metricsViewName = $spec?.metrics_view;

  // Click-to-filter callback
  async function handleCellClickToFilter(rowId: string, columnId: string) {
    if (!$pivotDataStore) return;

    try {
      // Set the active cell so the store computes activeCellFilters
      pivotState.update((state) => ({
        ...state,
        activeCell: { rowId, columnId },
      }));

      // Wait for the next tick to allow the store to react and compute activeCellFilters
      await tick();

      // Get the computed activeCellFilters from the store
      const activeCellFilters = $pivotDataStore.activeCellFilters;

      if (!activeCellFilters) return;

      // Extract dimension filters
      const dimensionFilters = extractDimensionFiltersFromExpression(
        activeCellFilters.filters,
      );

      if (dimensionFilters.length === 0) return;

      // Apply all dimension filters at once by toggling them sequentially
      // on the FilterState (which doesn't trigger URL updates), then apply to URL once
      const filterClass = filterManager.metricsViewFilters.get(metricsViewName);
      if (!filterClass) return;

      // Remove temporary filter status for all dimensions
      dimensionFilters.forEach(({ dimensionName }) => {
        filterManager.checkTemporaryFilter(dimensionName, [metricsViewName]);
      });

      // Toggle all dimension values in the filter state
      // Each call to toggleDimensionValueSelections returns the updated filter string
      // but doesn't apply it to URL - we just need the final result
      let filterString: string | null = null;
      dimensionFilters.forEach(({ dimensionName, values }) => {
        filterString = filterClass.toggleDimensionValueSelections(
          dimensionName,
          values,
          false, // keepPillVisible
          false, // isExclusiveFilter
        );
      });

      // Apply to URL once with the final filter string
      if (filterString !== null) {
        await filterManager.applyFiltersToUrl(
          new Map([[metricsViewName, filterString]]),
        );
      }

      // TODO: Apply time range if present
      // For now, time range application is deferred as it requires
      // investigation into the proper mechanism (component.parent.timeState)
      // if (activeCellFilters.timeRange) {
      //   // Apply time range...
      // }
    } catch (error) {
      console.error("Error applying cell filters:", error);
    }
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
