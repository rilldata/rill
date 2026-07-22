<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import { FilterManagerManager } from "@rilldata/web-common/features/dashboards/filters/manager/expression-filter-manager.svelte.ts";
  import DimensionFilter from "@rilldata/web-common/features/dashboards/filters/manager/DimensionFilter.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers.ts";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
  import FilterButton from "@rilldata/web-common/features/dashboards/filters/FilterButton.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  const StateManagers = getStateManagers();
  const {
    actions: {
      filters: { setFilter },
    },
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: { dimensionHasFilter },
      measures: { filteredSimpleMeasures },
      measureFilters: { measureHasFilter },
    },
    validSpecStore,
    dashboardStore,
  } = StateManagers;

  let validExplore = $derived($validSpecStore.data?.explore ?? {});
  let metricsViewName = $derived(validExplore.metricsView ?? "");

  let dimensionIdMap = $derived(
    getMapFromArray(
      $allDimensions,
      (dimension) => (dimension.name || dimension.column) as string,
    ),
  );

  const filterUrlStore = UrlParamsState.createStringParam(
    ExploreStateURLParams.Filters,
  );

  const expressionFilterManager = new FilterManagerManager();
  $effect(() =>
    expressionFilterManager.setExprParam(
      filterUrlStore.value ?? "",
      dimensionIdMap,
      metricsViewName,
    ),
  );
  let hasFilters = $derived(
    expressionFilterManager.dimensionFilterManagers.length > 0,
  );

  $effect(() => setFilter(expressionFilterManager.expr));

  let expressionMap = $derived(
    new Map([[metricsViewName, expressionFilterManager.expr]]),
  );

  let { selectedTimeDimension } = $derived($dashboardStore);
  const timeControlsStore = useTimeControlStore(StateManagers);
  let {
    timeStart,
    timeEnd,
    ready: timeControlsReady,
  } = $derived($timeControlsStore);
</script>

<div class="relative flex flex-row gap-x-2 gap-y-2 items-start">
  <div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2">
    {#each expressionFilterManager.dimensionFilterManagers as dimensionFilterManager (dimensionFilterManager.name)}
      <DimensionFilter
        filterData={dimensionFilterManager}
        {expressionMap}
        {timeStart}
        {timeEnd}
        {timeControlsReady}
        timeDimension={selectedTimeDimension}
      />
    {/each}

    <FilterButton
      allDimensions={$allDimensions}
      filteredSimpleMeasures={$filteredSimpleMeasures()}
      dimensionHasFilter={$dimensionHasFilter}
      measureHasFilter={$measureHasFilter}
      setTemporaryFilterName={(name) =>
        expressionFilterManager.addTemporaryDimensionFilter(
          name,
          dimensionIdMap,
          metricsViewName,
        )}
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
