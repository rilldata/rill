<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils.ts";
  import { UrlParamsState } from "@rilldata/web-common/lib/store-utils/url-params-state.svelte.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/expression-filter-manager.svelte.ts";
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

  const manager = new ExpressionFilterManager();
  $effect(() =>
    manager.setExprParam(
      filterUrlStore.value ?? "",
      dimensionIdMap,
      metricsViewName,
    ),
  );
  let hasFilters = $derived(manager.dimensionFilterManagers.length > 0);

  $effect(() => setFilter(manager.expr));

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
    {#each manager.dimensionFilterManagers as dimensionManager (dimensionManager.name)}
      <DimensionFilter
        {manager}
        {dimensionManager}
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
        manager.addTemporaryDimensionFilter(
          name,
          dimensionIdMap,
          metricsViewName,
        )}
    />
    <!-- if filters are present, place a chip at the end of the flex container
    that enables clearing all filters -->
    {#if hasFilters}
      <Button type="text" onClick={() => manager.clear()}>
        {m.dashboard_clear_filters()}
      </Button>
    {/if}
  </div>
</div>
