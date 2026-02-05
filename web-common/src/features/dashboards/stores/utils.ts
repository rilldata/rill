import { getMetricsViewTimeRangeFromExploreQueryOptions } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { getExploreValidSpecQueryOptions } from "@rilldata/web-common/features/explores/selectors.ts";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

export function getFilterStateFromNameStore(
  exploreNameStore: Readable<string>,
): Readable<FiltersState> {
  return derived(
    [metricsExplorerStore, exploreNameStore],
    ([metricsExplorerState, exploreName]) => {
      const exploreState = metricsExplorerState.entities[exploreName];
      const filtersState: FiltersState = {
        whereFilter: exploreState?.whereFilter ?? createAndExpression([]),
        dimensionThresholdFilters:
          exploreState?.dimensionThresholdFilters ?? [],
        dimensionsWithInlistFilter:
          exploreState?.dimensionsWithInlistFilter ?? [],
        dimensionFilterExcludeMode:
          exploreState?.dimensionFilterExcludeMode ?? new Map(),
      };

      return filtersState;
    },
  );
}

export function getTimeControlsStateFromNameStore(
  exploreNameStore: Readable<string>,
) {
  const validSpecQuery = createQuery(
    getExploreValidSpecQueryOptions(exploreNameStore),
  );
  const metricsViewTimeRangeQuery = createQuery(
    getMetricsViewTimeRangeFromExploreQueryOptions(exploreNameStore),
  );

  return derived(
    [
      metricsExplorerStore,
      exploreNameStore,
      validSpecQuery,
      metricsViewTimeRangeQuery,
    ],
    ([metricsExplorerState, exploreName, validSpecResp, timeRangeResp]) => {
      const exploreState = metricsExplorerState.entities[exploreName];
      const metricsViewSpec = validSpecResp.data?.metricsViewSpec ?? {};
      const exploreSpec = validSpecResp.data?.exploreSpec ?? {};
      const timeRangeSummary = timeRangeResp.data?.timeRangeSummary;

      const exploreTimeControlState = exploreState
        ? getTimeControlState(
            metricsViewSpec,
            exploreSpec,
            timeRangeSummary,
            exploreState,
          )
        : undefined;
      const timeControlState = <TimeControlState>{
        selectedTimeRange: exploreTimeControlState?.selectedTimeRange,
        selectedComparisonTimeRange:
          exploreTimeControlState?.selectedComparisonTimeRange,
        showTimeComparison: exploreTimeControlState?.showTimeComparison,
        selectedTimezone: exploreState?.selectedTimezone,
      };

      return timeControlState;
    },
  );
}
