<script lang="ts">
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { FilterSpecStore } from "@rilldata/web-common/features/filters/filter-spec-store";
  import type { FilterStore } from "@rilldata/web-common/features/filters/filter-store";
  import Button from "web-common/src/components/button/Button.svelte";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";

  export let filtersStore: FilterStore;
  export let specStore: FilterSpecStore;
  export let metricsViewNames: string[];
  export let readOnly = false;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let timeControlsReady: boolean | undefined;

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  $: ({
    getAllDimensionFilterItems,
    dimensionHasFilter,
    getAllMeasureFilterItems,
    measureHasFilter,
    hasFilters,

    toggleDimensionValueSelection,
    toggleDimensionFilterMode,
    removeDimensionFilter,
    clearAllFilters,
    setTemporaryFilterName,
  } = filtersStore);

  $: ({ dimensions, measures } = specStore);

  $: allDimensionFilters = $getAllDimensionFilterItems;

  $: allMeasureFilters = $getAllMeasureFilterItems;
</script>

<div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
  {#if !readOnly}
    <Filter size="16px" className="ui-copy-icon flex-none mt-[5px]" />
  {/if}
  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
    {#if !allDimensionFilters.length && !allMeasureFilters.length}
      <div
        in:fly={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid ml-1 items-center"
        style:min-height={ROW_HEIGHT}
      >
        No filters selected
      </div>
    {:else}
      {#each allDimensionFilters as { name, label, exclude, values } (name)}
        {@const dimension = $dimensions.find(
          (d) => d.name === name || d.column === name,
        )}
        {@const dimensionName = dimension?.name || dimension?.column}
        <div animate:flip={{ duration: 200 }}>
          {#if dimensionName}
            <DimensionFilter
              {metricsViewNames}
              {readOnly}
              {name}
              {label}
              selectedValues={values ?? []}
              {timeStart}
              {timeEnd}
              {timeControlsReady}
              excludeMode={exclude}
              onRemove={() => removeDimensionFilter(name)}
              onToggleFilterMode={() => toggleDimensionFilterMode(name)}
              onSelect={(value) =>
                toggleDimensionValueSelection(name, value, true)}
            />
          {/if}
        </div>
      {/each}
      {#each allMeasureFilters as { name, label, dimensionName, filter } (name)}
        <!--        <div animate:flip={{ duration: 200 }}>-->
        <!--          <MeasureFilter-->
        <!--            allDimensions={$dimensions}-->
        <!--            {name}-->
        <!--            {label}-->
        <!--            {dimensionName}-->
        <!--            {filter}-->
        <!--            onRemove={() => removeMeasureFilter(dimensionName, name)}-->
        <!--            onApply={({ dimension, oldDimension, filter }) =>-->
        <!--              handleMeasureFilterApply(dimension, name, oldDimension, filter)}-->
        <!--          />-->
        <!--        </div>-->
      {/each}
    {/if}

    {#if !readOnly}
      <FilterButton
        allDimensions={$dimensions}
        filteredSimpleMeasures={$measures}
        dimensionHasFilter={$dimensionHasFilter}
        measureHasFilter={$measureHasFilter}
        {setTemporaryFilterName}
      />
      <!-- if filters are present, place a chip at the end of the flex container
    that enables clearing all filters -->
      {#if $hasFilters}
        <Button type="text" on:click={clearAllFilters}>Clear filters</Button>
      {/if}
    {/if}
  </div>
</div>
