<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { isUrlTooLong } from "@rilldata/web-common/features/dashboards/url-state/url-length-limits";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import { applyDimensionInListMode as applyDimensionInListModeDirectly } from "../state-managers/actions/dimension-filters";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import FilterButton from "./FilterButton.svelte";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";
  import { createRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";

  export let readOnly = false;
  export let metricsViewName: string;
  export let hasTimeSeries: boolean;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const StateManagers = getStateManagers();
  const {
    exploreName,
    actions: {
      dimensionsFilter: {
        toggleMultipleDimensionValueSelections,
        applyDimensionInListMode,
        applyDimensionContainsMode,
        removeDimensionFilter,
        toggleDimensionFilterMode,
      },

      measuresFilter: { setMeasureFilter, removeMeasureFilter },
      filters: { clearAllFilters, setTemporaryFilterName },
    },
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: {
        dimensionHasFilter,
        getDimensionFilterItems,
        getAllDimensionFilterItems,
      },
      measures: { allMeasures, filteredSimpleMeasures },
      measureFilters: {
        getMeasureFilterItems,
        getAllMeasureFilterItems,
        measureHasFilter,
      },
    },
    dashboardStore,
    validSpecStore,
    timeRangeSummaryStore,
    dashboardConfigProvider,
    timeFilterManager,
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);

  const dashboardStateSync = DashboardStateSync.getFromContext();

  $: ({ timeStart, timeEnd, ready: timeControlsReady } = $timeControlsStore);

  $: dimensions = $allDimensions;
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: measures = $allMeasures;
  $: measureIdMap = getMapFromArray(measures, (m) => m.name as string);

  $: currentDimensionFilters = $getDimensionFilterItems(dimensionIdMap);
  $: allDimensionFilters = $getAllDimensionFilterItems(
    currentDimensionFilters,
    dimensionIdMap,
  );

  $: currentMeasureFilters = $getMeasureFilterItems(measureIdMap);
  $: allMeasureFilters = $getAllMeasureFilterItems(
    currentMeasureFilters,
    measureIdMap,
  );

  // hasFilter only checks for complete filters and excludes temporary ones
  $: hasFilters =
    currentDimensionFilters.length > 0 || currentMeasureFilters.length > 0;

  $: ({ whereFilter, selectedTimeDimension } = $dashboardStore);

  $: isComplexFilter = isExpressionUnsupported(whereFilter);

  const defaultUrlParamsStore = createRillDefaultExploreUrlParams(
    validSpecStore as any,
    timeRangeSummaryStore as any,
  );
  $: defaultUrlParams = $defaultUrlParamsStore.data;

  function handleMeasureFilterApply(
    dimension: string,
    measureName: string,
    oldDimension: string,
    filter: MeasureFilterEntry,
  ) {
    if (oldDimension && oldDimension !== dimension) {
      removeMeasureFilter(oldDimension, measureName);
    }
    setMeasureFilter(dimension, filter);
  }

  function isUrlTooLongAfterInListFilter(
    dimensionName: string,
    values: string[],
  ) {
    if (!dashboardStateSync) return false;

    const exploreState = structuredClone($dashboardStore);
    applyDimensionInListModeDirectly(
      { dashboard: exploreState },
      dimensionName,
      values,
    );
    const url = dashboardStateSync.getUrlForExploreState(exploreState);
    return isUrlTooLong(url);
  }

  function syncTimeFilters() {
    metricsExplorerStore.syncTimeFilters($exploreName, timeFilterManager);
    return Promise.resolve();
  }
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <TimeFilters
      {timeFilterManager}
      {dashboardConfigProvider}
      {defaultUrlParams}
      config={{
        showTimeDimensionSelector: true,
        showDefaultItem: true,
        showFullRange: true,
        showWatermark: true,
      }}
      context="explore"
      {syncTimeFilters}
    />
  {/if}

  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
    {#if !readOnly}
      <Filter size="16px" className="text-fg-secondary flex-none mt-[5px]" />
    {/if}
    <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
      {#if isComplexFilter}
        <AdvancedFilter advancedFilter={whereFilter} />
      {:else if !allDimensionFilters.length && !allMeasureFilters.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="text-fg-muted grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          {m.dashboard_no_filters_selected()}
        </div>
      {:else}
        {#each allDimensionFilters as filterData (filterData.name)}
          <div animate:flip={{ duration: 200 }}>
            <DimensionFilter
              expressionMap={new Map([
                [metricsViewName, $dashboardStore.whereFilter],
              ])}
              {filterData}
              {readOnly}
              {timeStart}
              {timeEnd}
              timeDimension={selectedTimeDimension}
              {timeControlsReady}
              removeDimensionFilter={async (name) =>
                removeDimensionFilter(name)}
              toggleDimensionFilterMode={async (name) => {
                toggleDimensionFilterMode(name);
              }}
              toggleDimensionValueSelections={async (
                name,
                values,
                _metricsViewNames,
                keepPillVisible,
                isExclusiveFilter,
                exclude,
              ) =>
                toggleMultipleDimensionValueSelections(
                  name,
                  values,
                  keepPillVisible ?? true,
                  isExclusiveFilter,
                  exclude,
                )}
              applyDimensionInListMode={async (name, values) =>
                applyDimensionInListMode(name, values)}
              applyDimensionContainsMode={async (name, searchText) =>
                applyDimensionContainsMode(name, searchText)}
              isUrlTooLongAfterInListFilter={(values) =>
                isUrlTooLongAfterInListFilter(filterData.name, values)}
            />
          </div>
        {/each}
        {#each allMeasureFilters as filterData (filterData.name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              {filterData}
              allDimensions={dimensions}
              onRemove={() =>
                removeMeasureFilter(filterData.dimensionName, filterData.name)}
              onApply={({ dimension, oldDimension, filter }) =>
                handleMeasureFilterApply(
                  dimension,
                  filterData.name,
                  oldDimension,
                  filter,
                )}
            />
          </div>
        {/each}
      {/if}

      {#if !readOnly}
        <FilterButton
          allDimensions={dimensions}
          filteredSimpleMeasures={$filteredSimpleMeasures()}
          dimensionHasFilter={$dimensionHasFilter}
          measureHasFilter={$measureHasFilter}
          {setTemporaryFilterName}
        />
        <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
        {#if hasFilters}
          <Button type="text" onClick={clearAllFilters}
            >{m.dashboard_clear_filters()}</Button
          >
        {/if}
      {/if}
    </div>
  </div>
</div>
