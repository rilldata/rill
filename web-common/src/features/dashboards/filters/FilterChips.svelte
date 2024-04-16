<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterRemove from "@rilldata/web-common/components/icons/FilterRemove.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";
  import { useMetricsView } from "../selectors/index";
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
      dimensionFilters: { getDimensionFilterItems, getAllDimensionFilterItems },
      measureFilters: { getMeasureFilterItems, getAllMeasureFilterItems },
    },
  } = StateManagers;

  const metricsView = useMetricsView(StateManagers);

  $: dimensions = $metricsView.data?.dimensions ?? [];
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name as string,
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

<div
  class:ui-copy-icon={true}
  class:ui-copy-icon-inactive={false}
  class="grid items-center place-items-center"
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
            column={dimension.column}
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
  {/if}
  <FilterButton />
  <!-- if filters are present, place a chip at the end of the flex container 
      that enables clearing all filters -->
  {#if hasFilters}
    <div class="ml-auto">
      <Chip
        bgBaseClass="surface"
        bgHoverClass="hover:bg-gray-100 hover:dark:bg-gray-700"
        textClass="ui-copy-disabled-faint hover:text-gray-500 dark:text-gray-500"
        bgActiveClass="bg-gray-200 dark:bg-gray-600"
        outlineBaseClass="outline-gray-400"
        on:click={clearAllFilters}
      >
        <span slot="icon" class="ui-copy-disabled-faint">
          <FilterRemove size="16px" />
        </span>
        <svelte:fragment slot="body">Clear filters</svelte:fragment>
      </Chip>
    </div>
  {/if}
</div>
