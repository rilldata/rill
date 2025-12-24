<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilter.svelte";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilter.svelte";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import CanvasFilterButton from "@rilldata/web-common/features/dashboards/filters/CanvasFilterButton.svelte";
  import type { FilterState } from "../../stores/filter-state";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import AdvancedFilter from "@rilldata/web-common/features/dashboards/filters/AdvancedFilter.svelte";

  export let id: string;
  export let canvasName: string;
  export let metricsView: string;
  export let localFilters: FilterState;
  export let excludedDimensions: Set<string>;
  export let updateLocalFilterString: (newFilterString: string) => void;

  let localFiltersEnabledOverride = false;

  const instanceId = httpClient.getInstanceId();

  $: ({
    canvasEntity: {
      filterManager: { dimensionsForMetricsView, measuresForMetricsView },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({
    parsed,
    temporaryFilterKeys,
    removeDimensionFilter,
    toggleDimensionFilterMode,
    toggleDimensionValueSelections,
    applyDimensionContainsMode,
    applyDimensionInListMode,
    setTemporaryFilterName,
    removeMeasureFilter,
    setMeasureFilter,
    clearAllFilters,
  } = localFilters);

  $: ({
    dimensionFilters,
    where,
    urlFormat,
    measureFilters,
    complexFilters,
    hasFilters,
  } = $parsed);

  $: localFiltersEnabled = !!urlFormat?.length || localFiltersEnabledOverride;

  $: allDimensions = $dimensionsForMetricsView.get(metricsView);
  $: allSimpleMeasures = $measuresForMetricsView.get(metricsView);

  $: dimensionArray = Array.from(allDimensions?.values() ?? []);

  $: remappedDimensions = new Map(
    Array.from(allDimensions?.entries() ?? []).map(([id, dim]) => {
      return [id, new Map([[metricsView, dim]])];
    }),
  );

  $: remappedMeasures = new Map(
    Array.from(allSimpleMeasures?.entries() ?? []).map(([id, measure]) => {
      return [id, new Map([[metricsView, measure]])];
    }),
  );

  function dimensionHasFilter(dimensionName: string): boolean {
    return (
      dimensionFilters.has(dimensionName) ||
      excludedDimensions.has(dimensionName)
    );
  }
  function measureHasFilter(measureName: string): boolean {
    return measureFilters.has(measureName);
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
          localFiltersEnabledOverride = false;
          updateLocalFilterString("");
        } else {
          localFiltersEnabledOverride = true;
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

      <CanvasFilterButton
        allDimensions={remappedDimensions}
        filteredSimpleMeasures={remappedMeasures}
        addBorder={false}
        {dimensionHasFilter}
        {measureHasFilter}
        {setTemporaryFilterName}
      />
    </div>

    <div class="relative flex flex-col gap-x-2 gap-y-2 items-start">
      <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
        {#each complexFilters as filter, i (i)}
          <AdvancedFilter advancedFilter={filter} />
        {/each}

        {#each dimensionFilters as [id, filterData] (id)}
          <DimensionFilter
            {filterData}
            timeStart={new Date(0).toISOString()}
            timeEnd={new Date().toISOString()}
            openOnMount={$temporaryFilterKeys.has(id)}
            timeControlsReady
            expressionMap={new Map([[metricsView, where]])}
            removeDimensionFilter={async () => {
              const newParam = removeDimensionFilter(id);
              updateLocalFilterString(newParam);
            }}
            toggleDimensionFilterMode={async () => {
              const newParam = toggleDimensionFilterMode(id);
              if (newParam) updateLocalFilterString(newParam);
            }}
            toggleDimensionValueSelections={async (_, values) => {
              const newParam = toggleDimensionValueSelections(id, values);

              if (newParam) updateLocalFilterString(newParam);
            }}
            applyDimensionInListMode={async (_, values) => {
              const newParam = applyDimensionInListMode(id, values);
              if (newParam) updateLocalFilterString(newParam);
            }}
            applyDimensionContainsMode={async (searchText) => {
              const newParam = applyDimensionContainsMode(id, searchText);
              if (newParam) updateLocalFilterString(newParam);
            }}
          />
        {/each}

        {#each measureFilters as [id, filterData] (id)}
          <MeasureFilter
            {filterData}
            allDimensions={dimensionArray}
            onRemove={() => {
              const newParam = removeMeasureFilter(
                filterData.dimensionName,
                filterData.name,
              );

              if (newParam) updateLocalFilterString(newParam);
            }}
            onApply={({ dimension, filter, oldDimension }) => {
              const newParam = setMeasureFilter(
                dimension,
                filter,
                oldDimension,
              );
              if (newParam) updateLocalFilterString(newParam);
            }}
            toggleFilterPin={undefined}
          />
        {/each}
      </div>
      <div class="ml-auto">
        {#if hasFilters}
          <Button
            type="text"
            onClick={() => {
              const newParam = clearAllFilters();
              updateLocalFilterString(newParam);
            }}
          >
            Clear filters
          </Button>
        {/if}
      </div>
    </div>
  {/if}
</div>
