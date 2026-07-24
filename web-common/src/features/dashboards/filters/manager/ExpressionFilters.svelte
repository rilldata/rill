<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/manager/DimensionFilter.svelte";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MeasureFilter from "@rilldata/web-common/features/dashboards/filters/manager/MeasureFilter.svelte";
  import { fly } from "svelte/transition";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/expression-filter-manager.svelte.ts";
  import { getFilteredSimpleMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let {
    expressionFilterManager,

    timeEnd,
    timeStart,
    timeControlsReady,
    timeDimension,

    isUrlTooLongAfterInListFilter,
  }: {
    expressionFilterManager: ExpressionFilterManager;

    timeStart: string | undefined;
    timeEnd: string | undefined;
    timeControlsReady: boolean | undefined;
    timeDimension: string | undefined;

    isUrlTooLongAfterInListFilter?: (name: string, values: string[]) => boolean;
  } = $props();

  /** the height of a row of chips */
  const ROW_HEIGHT = "26px";

  const runtimeClient = useRuntimeClient();

  let validSpecQuery = $derived(
    useExploreValidSpec(runtimeClient, expressionFilterManager.exploreName),
  );
  let exploreSpec = $derived($validSpecQuery.data?.explore ?? {});

  let allMeasures = $derived([
    ...expressionFilterManager.measureIdMap.values(),
  ]);
  let allDimensions = $derived([
    ...expressionFilterManager.dimensionIdMap.values(),
  ]);

  let filteredSimpleMeasures = $derived(
    getFilteredSimpleMeasures(allMeasures, exploreSpec.measures),
  );

  let hasFilters = $derived(
    expressionFilterManager.dimensionFilterManagers.length > 0 ||
      expressionFilterManager.measureFilterManagers.length > 0,
  );

  let measureHasFilter = $derived((name: string) =>
    expressionFilterManager.measureFilterManagers.some(
      (mfm) => mfm.name === name,
    ),
  );
  let dimensionHasFilter = $derived((name: string) =>
    expressionFilterManager.dimensionFilterManagers.some(
      (dfm) => dfm.name === name,
    ),
  );

  function handleAddNew(name: string) {
    if (expressionFilterManager.measureIdMap.has(name)) {
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
          {timeDimension}
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
          {allDimensions}
          openOnMount={expressionFilterManager.temporaryFilter ===
            measureManager}
        />
      {/each}
    {/if}

    <FilterButton
      {allDimensions}
      {filteredSimpleMeasures}
      {dimensionHasFilter}
      {measureHasFilter}
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
