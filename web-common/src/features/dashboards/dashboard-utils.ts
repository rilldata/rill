import { forEachExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import type {
  QueryServiceMetricsViewComparisonBody,
  MetricsViewDimension,
  V1MetricsViewFilter,
  MetricsViewSpecMeasureV2,
  V1MetricsViewAggregationMeasure,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import type { TimeControlState } from "./time-controls/time-control-store";
import { getQuerySortType } from "./leaderboard/leaderboard-utils";
import { SortType } from "./proto-state/derived-types";

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
export function getDimensionColumn(dimension: MetricsViewDimension): string {
  return dimension?.column || dimension?.name;
}

export function prepareSortedQueryBody(
  dimensionName: string,
  measureNames: string[],
  timeControls: TimeControlState,
  // Note: sortMeasureName may be null if we are sorting by dimension values
  sortMeasureName: string | null,
  sortType: SortType,
  sortAscending: boolean,
  filterForDimension: V1Expression,
  havingFilter: V1Expression
): QueryServiceMetricsViewComparisonBody {
  let comparisonTimeRange = {
    start: timeControls.comparisonTimeStart,
    end: timeControls.comparisonTimeEnd,
  };

  // FIXME: As a temporary way of enabling sorting by dimension values,
  // Benjamin and Egor put in a patch that will allow us to use the
  // dimension name as the measure name. This will need to be updated
  // once they have stabilized the API.
  if (sortType === SortType.DIMENSION || sortMeasureName === null) {
    sortMeasureName = dimensionName;
    // note also that we need to remove the comparison time range
    // when sorting by dimension values, or the query errors
    comparisonTimeRange = undefined;
  }

  const querySortType = getQuerySortType(sortType);

  // We need to select measures for filters on them to work
  // TODO: fix this in the backend with an improved filter clause builder in future.
  //       It should switch to raw expression if dimension/measure is not in select clause.
  const selectedMeasures = new Set(measureNames);
  forEachExpression(havingFilter, (e) => {
    if (!e.ident || selectedMeasures.has(e.ident)) return;
    selectedMeasures.add(e.ident);
    measureNames.push(e.ident);
  });

  return {
    dimension: {
      name: dimensionName,
    },
    measures: measureNames.map(
      (n) =>
        <V1MetricsViewAggregationMeasure>{
          name: n,
        }
    ),
    timeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    comparisonTimeRange,
    sort: [
      {
        desc: !sortAscending,
        name: sortMeasureName,
        sortType: querySortType,
      },
    ],
    where: filterForDimension,
    having: havingFilter,
    limit: "250",
    offset: "0",
  };
}
