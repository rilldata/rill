<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterRemove from "@rilldata/web-common/components/icons/FilterRemove.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { getStateManagers } from "../state-managers/state-managers";
  import FilterButton from "./FilterButton.svelte";
  import DimensionFilter from "./dimension-filters/DimensionFilter.svelte";

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
      dimensions: { allDimensions },
      dimensionFilters: { getDimensionFilterItems, getAllDimensionFilterItems },
      measures: { allMeasures },
      measureFilters: { getMeasureFilterItems, getAllMeasureFilterItems },
    },
  } = StateManagers;

  $: dimensions = $allDimensions;
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name as string,
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

<div
  class="grid items-center place-items-center"
  class:ui-copy-icon={true}
  class:ui-copy-icon-inactive={false}
  style:height={ROW_HEIGHT}
  style:width={ROW_HEIGHT}
>
  <Filter size="16px" />
</div>
<div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 items-center">
  {#if !allDimensionFilters.length && !allMeasureFilters.length}
    <div
      in:fly|local={{ duration: 200, x: 8 }}
      class="ui-copy-disabled grid items-center"
      style:min-height={ROW_HEIGHT}
    >
      No filters selected
    </div>
  {:else}
    {#each allDimensionFilters as { name, label, selectedValues } (name)}
      {@const dimension = dimensions.find((d) => d.name === name)}
      <div animate:flip={{ duration: 200 }}>
        {#if dimension?.column}
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
    {#each allMeasureFilters as { name, label, dimensionName, filter } (name)}
      <div animate:flip={{ duration: 200 }}>
        <MeasureFilter
          {name}
          {label}
          {dimensionName}
          {filter}
          on:remove={() => removeMeasureFilter(dimensionName, name)}
          on:apply={({ detail: { dimension, oldDimension, filter } }) =>
            handleMeasureFilterApply(dimension, name, oldDimension, filter)}
        />
      </div>
    {/each}
  {/if}
  <FilterButton />
  <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
  {#if hasFilters}
    <div class="ml-auto">
      <Chip on:click={clearAllFilters}>
        <span slot="icon" class="ui-copy-disabled-faint">
          <FilterRemove size="16px" />
        </span>
        <svelte:fragment slot="body">Clear filters</svelte:fragment>
      </Chip>
    </div>
  {/if}
</div>
