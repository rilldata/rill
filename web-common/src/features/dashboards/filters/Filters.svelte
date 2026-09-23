<script lang="ts">
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { isUrlTooLong } from "@rilldata/web-common/features/dashboards/url-state/url-length-limits";
  import { getStateManagers } from "../state-managers/state-managers";
  import ExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ExpressionFilters.svelte";
  import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
  import { untrack } from "svelte";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import TimeFilters from "@rilldata/web-common/features/dashboards/time-controls/TimeFilters.svelte";

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    dashboardConfigProvider,
    expressionFilterManager,
    timeFilterManager,
  } = StateManagers;

  let { metricsViewsProvider, yamlConfigProvider } = $derived(
    dashboardConfigProvider,
  );

  let {
    hasTimeSeries,
    timeStart,
    timeEnd,
    timeDimension,
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
</script>

<div class="flex flex-col gap-y-2 size-full">
  {#if hasTimeSeries}
    <TimeFilters
      {timeFilterManager}
      {metricsViewsProvider}
      {yamlConfigProvider}
      config={{
        showTimeDimensionSelector: true,
        showFullRange: true,
        showWatermark: true,
      }}
      context="explore"
    />
  {/if}

  <ExpressionFilters
    {expressionFilterManager}
    {dashboardConfigProvider}
    {timeStart}
    {timeEnd}
    {timeDimension}
    {timeControlsReady}
    {isUrlTooLongAfterInListFilter}
  />
</div>
