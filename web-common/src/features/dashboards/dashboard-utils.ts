import type {
  QueryServiceMetricsViewComparisonToplistBody,
  MetricsViewDimension,
  V1MetricsViewFilter,
  MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";
import type { TimeControlState } from "./time-controls/time-control-store";
import { getQuerySortType } from "./leaderboard/leaderboard-utils";
import type { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";

export function isSummableMeasure(measure: MetricsViewSpecMeasureV2): boolean {
  return (
    measure?.expression.toLowerCase()?.includes("count(") ||
    measure?.expression?.toLowerCase()?.includes("sum(")
  );
}

/**
 * Returns a sanitized column name appropriate for use in e.g. filters.
 *
 * Even though this is a one-liner, we externalize it as a function
 * becuase it is used in a few places and we want to make sure we
 * are consistent in how we handle this.
 */
export function getDimensionColumn(dimension: MetricsViewDimension) {
  return dimension?.column || dimension?.name;
}

export function prepareSortedQueryBody(
  dimensionName: string,
  measureNames: string[],
  timeControls: TimeControlState,
  sortMeasureName: string,
  sortType: DashboardState_LeaderboardSortType,
  sortAscending: boolean,
  filterForDimension: V1MetricsViewFilter
): QueryServiceMetricsViewComparisonToplistBody {
  const querySortType = getQuerySortType(sortType);

  return {
    dimensionName,
    measureNames,
    baseTimeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    comparisonTimeRange: {
      start: timeControls.comparisonTimeStart,
      end: timeControls.comparisonTimeEnd,
    },
    sort: [
      {
        ascending: sortAscending,
        measureName: sortMeasureName,
        type: querySortType,
      },
    ],
    filter: filterForDimension,
    limit: "250",
    offset: "0",
  };
}
