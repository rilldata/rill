<script lang="ts">
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { isUrlTooLong } from "@rilldata/web-common/features/dashboards/url-state/url-length-limits";
  import { getStateManagers } from "../state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import ExpressionFilters from "./ExpressionFilters.svelte";
  import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
  import { untrack } from "svelte";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";

  const StateManagers = getStateManagers();
  const {
    exploreName,
    dashboardStore,
    dashboardConfigProvider,
    expressionFilterManager,
    timeFilterManager,
  } = StateManagers;

  let { metricsViewsProvider } = $derived(dashboardConfigProvider);

  syncStoreWithSource(
    expressionFilterManager,
    syncExpressionFilters,
    () => metricsViewsProvider.ready,
    undefined,
    // URL sync is managed by DashboardStateSync
    true,
  );

  syncStoreWithSource(
    timeFilterManager,
    syncTimeFilters,
    () => metricsViewsProvider.ready && timeFilterManager.ready,
    undefined,
    // URL sync is managed by DashboardStateSync
    true,
  );

  let {
    interval,
    hasTimeSeries,
    ready: timeControlsReady,
  } = $derived(timeFilterManager);

  const dashboardStateSync = DashboardStateSync.getFromContext();

  function isUrlTooLongAfterInListFilter(
    dimensionName: string,
    values: string[],
  ) {
    if (!dashboardStateSync) return false;

    // The chip calls this from a `$derived`, and the clone below mutates its own state while it
    // applies the filter, so the whole computation has to be untracked.
    return untrack(() => {
      const tempFilterManger = expressionFilterManager.clone();
      tempFilterManger.dimensionFilterAction(dimensionName, (m) =>
        m.setInList(values, m.exclude),
      );

      // Only the filter differs from the current state, and getUrlForExploreState only reads,
      // so a shallow copy is enough.
      const exploreState: ExploreState = {
        ...$dashboardStore,
        whereFilter:
          Object.values(tempFilterManger.topLevelJoiner.expr)[0] ??
          createAndExpression([]),
        dimensionsWithInlistFilter: tempFilterManger.inList,
      };

      const url = dashboardStateSync.getUrlForExploreState(exploreState);
      return isUrlTooLong(url);
    });
  }

  function syncExpressionFilters() {
    if (!expressionFilterManager.updating) {
      metricsExplorerStore.syncExpressionFilter(
        $exploreName,
        expressionFilterManager,
      );
    }
    return Promise.resolve();
  }

  function syncTimeFilters() {
    if (!timeFilterManager.updating) {
      metricsExplorerStore.syncTimeFilters($exploreName, timeFilterManager);
    }
    return Promise.resolve();
  }
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <TimeFilters
      {timeFilterManager}
      {dashboardConfigProvider}
      config={{
        showTimeDimensionSelector: true,
        showDefaultItem: true,
        showFullRange: true,
        showWatermark: true,
      }}
      context="explore"
    />
  {/if}

  <ExpressionFilters
    {expressionFilterManager}
    {dashboardConfigProvider}
    timeStart={interval?.start?.toString()}
    timeEnd={interval?.end?.toString()}
    timeDimension={$dashboardStore.selectedTimeDimension}
    {timeControlsReady}
    {isUrlTooLongAfterInListFilter}
  />
</div>
