<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Calendar from "@rilldata/web-common/components/icons/Calendar.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import {
    useAllDimensionFromMetric,
    useAllMeasuresFromMetric,
    useAllSimpleMeasureFromMetric,
  } from "@rilldata/web-common/features/canvas/components/selectors";
  import CanvasGrainSelector from "@rilldata/web-common/features/canvas/filters/CanvasGrainSelector.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import CanvasComparisonPill from "./CanvasComparisonPill.svelte";
  import CanvasSuperPill from "./CanvasSuperPill.svelte";

  export let readOnly = false;
  export let metricsViewName = "bids";

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const { instanceId } = $runtime;
  const { canvasEntity } = getCanvasStateManagers();
  const { timeControls, filters } = canvasEntity;

  $: selectedTimeRange = timeControls.selectedTimeRange;
  $: selectedComparisonTimeRange = timeControls?.selectedComparisonTimeRange;
  $: activeTimeZone = timeControls.selectedTimezone;

  $: whereFilter = filters.whereFilter;

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(0),
    end: new Date(),
  };

  const {
    toggleDimensionValueSelection,
    removeDimensionFilter,
    toggleDimensionFilterMode,
    setMeasureFilter,
    removeMeasureFilter,
    setTemporaryFilterName,
    clearAllFilters,
    dimensionHasFilter,
    getDimensionFilterItems,
    getAllDimensionFilterItems,
    isFilterExcludeMode,
    getMeasureFilterItems,
    getAllMeasureFilterItems,
    measureHasFilter,
  } = filters;

  $: dimensionsQuery = useAllDimensionFromMetric(instanceId, metricsViewName);
  $: measuresQuery = useAllMeasuresFromMetric(instanceId, metricsViewName);
  $: simpleMeasuresQuery = useAllSimpleMeasureFromMetric(
    instanceId,
    metricsViewName,
  );

  // $: alllDimensions = useAllDimensionFromMetrics(instanceId, [
  //   "nyc_311_latest_metrics",
  //   "bids",
  //   "auction",
  // ]);

  // $: console.log($alllDimensions);

  $: dimensions = $dimensionsQuery?.data || [];
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: simpleMeasures = $simpleMeasuresQuery?.data || [];
  $: measures = $measuresQuery?.data || [];
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

  $: isComplexFilter = isExpressionUnsupported($whereFilter);

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
</script>

<div class="flex flex-col gap-y-2 w-full h-20 justify-center ml-2">
  <div class="flex flex-row flex-wrap gap-x-2 gap-y-1.5 items-center">
    <Calendar size="16px" />
    <CanvasSuperPill
      {allTimeRange}
      selectedTimeRange={$selectedTimeRange}
      activeTimeZone={$activeTimeZone}
    />
    <CanvasComparisonPill
      {allTimeRange}
      selectedTimeRange={$selectedTimeRange}
      selectedComparisonTimeRange={$selectedComparisonTimeRange}
    />
    <CanvasGrainSelector
      selectedTimeRange={$selectedTimeRange}
      selectedComparisonTimeRange={$selectedComparisonTimeRange}
    />
  </div>

  <div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
    {#if !readOnly}
      <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
    {/if}
    <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
      {#if isComplexFilter}
        <AdvancedFilter advancedFilter={$whereFilter} />
      {:else if !allDimensionFilters.length && !allMeasureFilters.length}
        <div
          in:fly={{ duration: 200, x: 8 }}
          class="ui-copy-disabled grid ml-1 items-center"
          style:min-height={ROW_HEIGHT}
        >
          No filters selected
        </div>
      {:else}
        {#each allDimensionFilters as { name, label, selectedValues } (name)}
          {@const dimension = dimensions.find(
            (d) => d.name === name || d.column === name,
          )}
          {@const dimensionName = dimension?.name || dimension?.column}
          <div animate:flip={{ duration: 200 }}>
            {#if dimensionName}
              <DimensionFilter
                {metricsViewName}
                {readOnly}
                {name}
                {label}
                {selectedValues}
                timeStart={$selectedTimeRange.start.toISOString()}
                timeEnd={$selectedTimeRange.end.toISOString()}
                timeControlsReady
                excludeMode={$isFilterExcludeMode(name)}
                onRemove={() => removeDimensionFilter(name)}
                onToggleFilterMode={() => toggleDimensionFilterMode(name)}
                onSelect={(value) =>
                  toggleDimensionValueSelection(name, value, true)}
              />
            {/if}
          </div>
        {/each}
        {#each allMeasureFilters as { name, label, dimensionName, filter } (name)}
          <div animate:flip={{ duration: 200 }}>
            <MeasureFilter
              allDimensions={dimensions}
              {name}
              {label}
              {dimensionName}
              {filter}
              onRemove={() => removeMeasureFilter(dimensionName, name)}
              onApply={({ dimension, oldDimension, filter }) =>
                handleMeasureFilterApply(dimension, name, oldDimension, filter)}
            />
          </div>
        {/each}
      {/if}

      {#if !readOnly}
        <FilterButton
          allDimensions={dimensions}
          filteredSimpleMeasures={simpleMeasures}
          dimensionHasFilter={$dimensionHasFilter}
          measureHasFilter={$measureHasFilter}
          {setTemporaryFilterName}
        />
        <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
        {#if hasFilters}
          <Button type="text" on:click={clearAllFilters}>Clear filters</Button>
        {/if}
      {/if}
    </div>
  </div>
</div>
