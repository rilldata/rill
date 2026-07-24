<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/manager/DimensionFilter.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers.ts";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/manager/MeasureFilter.svelte";
  import { fly } from "svelte/transition";

  let {
    isUrlTooLongAfterInListFilter,
  }: {
    isUrlTooLongAfterInListFilter?: (name: string, values: string[]) => boolean;
  } = $props();

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const StateManagers = getStateManagers();
  const {
    actions: {
      filters: { setFilter },
    },
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: { dimensionHasFilter },
      measures: { allMeasures, filteredSimpleMeasures },
      measureFilters: { measureHasFilter },
    },
    validSpecStore,
    dashboardStore,
    expressionFilterManager,
  } = StateManagers;

  let validExplore = $derived($validSpecStore.data?.explore ?? {});
  let metricsViewName = $derived(validExplore.metricsView ?? "");

  let measureIdMap = $derived(
    getMapFromArray($allMeasures, (measure) => measure.name as string),
  );

  let dimensionIdMap = $derived(
    getMapFromArray(
      $allDimensions,
      (dimension) => (dimension.name || dimension.column) as string,
    ),
  );

  $effect(() =>
    expressionFilterManager.syncSpec(
      metricsViewName,
      measureIdMap,
      dimensionIdMap,
    ),
  );

  let hasFilters = $derived(
    expressionFilterManager.dimensionFilterManagers.length > 0 ||
      expressionFilterManager.measureFilterManagers.length > 0,
  );

  $effect(() => setFilter(expressionFilterManager.expr));

  let { selectedTimeDimension } = $derived($dashboardStore);
  const timeControlsStore = useTimeControlStore(StateManagers);
  let {
    timeStart,
    timeEnd,
    ready: timeControlsReady,
  } = $derived($timeControlsStore);

  function handleAddNew(name: string) {
    if (measureIdMap.has(name)) {
      expressionFilterManager.addTemporaryMeasureFilter(name);
    } else {
      expressionFilterManager.addTemporaryDimensionFilter(name);
    }
  }
</script>

<div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
    {#if !hasFilters}
      <div
        in:fly={{ duration: 200, x: 8 }}
        class="text-fg-muted grid ml-1 items-center"
        style:min-height={ROW_HEIGHT}
      >
        {m.dashboard_no_filters_selected()}
      </div>
    {:else}
      {#each expressionFilterManager.dimensionFilterManagers as dimensionManager (dimensionManager.name)}
        <DimensionFilter
          manager={expressionFilterManager}
          {dimensionManager}
          {timeStart}
          {timeEnd}
          {timeControlsReady}
          timeDimension={selectedTimeDimension}
          openOnMount={expressionFilterManager.temporaryFilter ===
            dimensionManager}
          isUrlTooLongAfterInListFilter={isUrlTooLongAfterInListFilter
            ? (values) =>
                isUrlTooLongAfterInListFilter(dimensionManager.name, values)
            : undefined}
        />
      {/each}

      {#each expressionFilterManager.measureFilterManagers as measureManager (measureManager.name)}
        <MeasureFilter
          {measureManager}
          allDimensions={$allDimensions}
          openOnMount={expressionFilterManager.temporaryFilter ===
            measureManager}
        />
      {/each}
    {/if}

    <FilterButton
      allDimensions={$allDimensions}
      filteredSimpleMeasures={$filteredSimpleMeasures()}
      dimensionHasFilter={$dimensionHasFilter}
      measureHasFilter={$measureHasFilter}
      setTemporaryFilterName={handleAddNew}
    />
    <!-- if filters are present, place a chip at the end of the flex container
    that enables clearing all filters -->
    {#if hasFilters}
      <Button type="text" onClick={() => expressionFilterManager.clear()}>
        {m.dashboard_clear_filters()}
      </Button>
    {/if}
  </div>
</div>
