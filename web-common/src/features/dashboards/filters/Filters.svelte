<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { useMetricsView } from "../selectors/index";
  import { getStateManagers } from "../state-managers/state-managers";
  import FilterButton from "./FilterButton.svelte";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";
  import SuperPill from "../time-controls/super-pill/SuperPill.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";

  export let readOnly = false;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const StateManagers = getStateManagers();
  const {
    actions: {
      dimensionsFilter: {
        toggleDimensionValueSelection,
        removeDimensionFilter,
      },
      measuresFilter: { setMeasureFilter, removeMeasureFilter },
      filters: { clearAllFilters },
    },
    selectors: {
      dimensionFilters: { getDimensionFilterItems, getAllDimensionFilterItems },
      measureFilters: { getMeasureFilterItems, getAllMeasureFilterItems },
    },
  } = StateManagers;

  const timeControlsStore = useTimeControlStore(StateManagers);
  $: ({ selectedTimeRange, allTimeRange, showComparison } = $timeControlsStore);

  const metricsView = useMetricsView(StateManagers);

  $: dimensions = $metricsView.data?.dimensions ?? [];
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => (dimension.name || dimension.column) as string,
  );

  $: measures = $metricsView.data?.measures ?? [];
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

  function handleMeasureFilterApply(
    dimension: string,
    measureName: string,
    oldDimension: string,
    expr: V1Expression,
  ) {
    if (oldDimension && oldDimension !== dimension) {
      removeMeasureFilter(oldDimension, measureName);
    }
    setMeasureFilter(dimension, measureName, expr);
  }
</script>

<div class="flex gap-x-1 pb-2 px-2 flex-grow-0">
  {#if !readOnly}
    <div
      class:ui-copy-icon={true}
      class:ui-copy-icon-inactive={false}
      class="flex items-center text-center justify-center flex-shrink-0"
      style:height={ROW_HEIGHT}
      style:width={ROW_HEIGHT}
    >
      <Filter size="16px" />
    </div>
  {/if}

  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 items-center">
    {#if allTimeRange?.start && allTimeRange?.end}
      <SuperPill {allTimeRange} {selectedTimeRange} {showComparison} />
    {/if}
    {#each allDimensionFilters as { name, label, selectedValues } (name)}
      {@const dimension = dimensions.find(
        (d) => d.name === name || d.column === name,
      )}
      {@const dimensionName = dimension?.name || dimension?.column}
      <div animate:flip={{ duration: 200 }}>
        {#if dimensionName}
          <DimensionFilter
            {name}
            {label}
            {selectedValues}
            on:remove={() => removeDimensionFilter(name)}
            on:apply={(event) =>
              toggleDimensionValueSelection(name, event.detail, true)}
          />
        {/if}
      </div>
    {/each}
    {#each allMeasureFilters as { name, label, dimensionName, expr } (name)}
      <div animate:flip={{ duration: 200 }}>
        <MeasureFilter
          {name}
          {label}
          {dimensionName}
          {expr}
          on:remove={() => removeMeasureFilter(dimensionName, name)}
          on:apply={({ detail: { dimension, oldDimension, expr } }) =>
            handleMeasureFilterApply(dimension, name, oldDimension, expr)}
        />
      </div>
    {/each}

    {#if !readOnly}
      <FilterButton />
      <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
      {#if hasFilters}
        <Button type="text" on:click={clearAllFilters}>Clear filters</Button>
      {/if}
    {/if}
  </div>
</div>
