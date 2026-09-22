import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { type DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { ExpressionState } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

/**
 * Transforms time and filter store data into partial explore state
 */
export function transformTimeAndFiltersToExploreState(
  expressionState: ExpressionState,
  timeControlState: TimeControlState,
): Partial<ExploreState> {
  const exploreState: Partial<ExploreState> = {};

  if (expressionState.expr) {
    exploreState.whereFilter = expressionState.expr;
  }

  if (timeControlState.timeRange && timeControlState.apiTimeRange) {
    exploreState.selectedTimeRange = {
      name: timeControlState.timeRange,
      start: timeControlState.apiTimeRange.start
        ? new Date(timeControlState.apiTimeRange.start)
        : new Date(),
      end: timeControlState.apiTimeRange.end
        ? new Date(timeControlState.apiTimeRange.end)
        : new Date(),
      interval: timeControlState.timeGrain,
    };

    exploreState.selectedTimezone = timeControlState.timeZone;

    if (
      timeControlState.comparisonTimeRange &&
      timeControlState.apiComparisonTimeRange
    ) {
      exploreState.showTimeComparison = true;
      exploreState.selectedComparisonTimeRange = {
        name: timeControlState.comparisonTimeRange,
      } as DashboardTimeControls;
    } else {
      exploreState.showTimeComparison = false;
      exploreState.selectedComparisonTimeRange = undefined;
    }
  }

  return exploreState;
}
