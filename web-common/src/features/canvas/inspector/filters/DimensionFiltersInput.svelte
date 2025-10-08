<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import { isExpressionUnsupported } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flip } from "svelte/animate";
  import type { Filters } from "../../stores/filters";

  export let id: string;
  export let canvasName: string;
  export let metricsView: string;
  export let localFilters: Filters;
  export let excludedDimensions: string[];

  let localFiltersEnabled = false;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      metricsView: {
        getDimensionsForMetricView,
        getSimpleMeasuresForMetricView,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({
    whereFilter,
    toggleMultipleDimensionValueSelections,
    applyDimensionInListMode,
    applyDimensionContainsMode,
    removeDimensionFilter,
    toggleDimensionFilterMode,
    setMeasureFilter,
    removeMeasureFilter,
    setTemporaryFilterName,
    clearAllFilters,
    dimensionHasFilter,
    temporaryFilters,
    allDimensionFilterItems,
    allMeasureFilterItems,
    measureHasFilter,
    searchParamsStore,
  } = localFilters);

  $: if ($searchParamsStore.has(ExploreStateURLParams.Filters)) {
    localFiltersEnabled = true;
  }

  $: allDimensions = getDimensionsForMetricView(metricsView);
  $: allValidDimensions = $allDimensions.filter(
    (d) => !excludedDimensions.includes(d.name || (d.column as string)),
  );
  $: allSimpleMeasures = getSimpleMeasuresForMetricView(metricsView);

  $: allDimensionFilters = $allDimensionFilterItems;
  $: allMeasureFilters = $allMeasureFilterItems;

  // hasFilter only checks for complete filters and excludes temporary ones
  $: hasFilters =
    allDimensionFilters.size + allMeasureFilters.length >
    $temporaryFilters.size;

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

<div class="flex flex-col gap-y-2 pt-1">
  <div class="flex justify-between">
    <InputLabel
      capitalize={false}
      small
      label="Local filters"
      {id}
      faint={!localFiltersEnabled}
    />
    <Switch
      checked={localFiltersEnabled}
      on:click={() => {
        if (localFiltersEnabled) {
          localFiltersEnabled = false;
          clearAllFilters();
        } else {
          localFiltersEnabled = true;
        }
      }}
      small
    />
  </div>
  <div class="text-gray-500">
    {#if localFiltersEnabled}
      Overriding inherited filters from canvas.
    {:else}
      Overrides inherited filters from canvas when ON.
    {/if}
  </div>
  {#if localFiltersEnabled}
    <div class="flex justify-between gap-x-2">
      <InputLabel small label="Filters" {id} />

      <FilterButton
        allDimensions={allValidDimensions}
        filteredSimpleMeasures={$allSimpleMeasures}
        dimensionHasFilter={$dimensionHasFilter}
        measureHasFilter={$measureHasFilter}
        {setTemporaryFilterName}
        addBorder={false}
      />
    </div>

    <div class="relative flex flex-col gap-x-2 gap-y-2 items-start">
      <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
        {#if isComplexFilter}
          <AdvancedFilter advancedFilter={$whereFilter} />
        {:else if allDimensionFilters.size || allMeasureFilters.length}
          {#each allDimensionFilters as [name, { label, isInclude, mode, selectedValues, inputText }] (name)}
            {@const dimension = allValidDimensions.find(
              (d) => d.name === name || d.column === name,
            )}
            {@const dimensionName = dimension?.name || dimension?.column}
            <div animate:flip={{ duration: 200 }}>
              {#if dimensionName}
                <DimensionFilter
                  metricsViewNames={[metricsView]}
                  readOnly={false}
                  smallChip
                  {name}
                  {label}
                  {mode}
                  {selectedValues}
                  {inputText}
                  timeStart={new Date(0).toISOString()}
                  timeEnd={new Date().toISOString()}
                  timeControlsReady
                  excludeMode={!isInclude}
                  whereFilter={$whereFilter}
                  onRemove={() => removeDimensionFilter(name)}
                  onToggleFilterMode={() => toggleDimensionFilterMode(name)}
                  onSelect={(value) =>
                    toggleMultipleDimensionValueSelections(name, [value], true)}
                  onMultiSelect={(values) =>
                    toggleMultipleDimensionValueSelections(name, values, true)}
                  onApplyInList={(values) =>
                    applyDimensionInListMode(name, values)}
                  onApplyContainsMode={(searchText) =>
                    applyDimensionContainsMode(name, searchText)}
                />
              {/if}
            </div>
          {/each}
          {#each allMeasureFilters as { name, label, dimensionName, filter } (name)}
            <div animate:flip={{ duration: 200 }}>
              <MeasureFilter
                allDimensions={allValidDimensions}
                {name}
                {label}
                {dimensionName}
                {filter}
                onRemove={() => removeMeasureFilter(dimensionName, name)}
                onApply={({ dimension, oldDimension, filter }) =>
                  handleMeasureFilterApply(
                    dimension,
                    name,
                    oldDimension,
                    filter,
                  )}
              />
            </div>
          {/each}
        {/if}
      </div>
      <div class="ml-auto">
        {#if hasFilters}
          <Button type="text" onClick={clearAllFilters}>Clear filters</Button>
        {/if}
      </div>
    </div>
  {/if}
</div>
