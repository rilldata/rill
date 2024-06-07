import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1MetricsViewAggregationMeasure,
  V1Expression,
  QueryServiceMetricsViewAggregationBody,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { TimeControlState } from "./time-controls/time-control-store";
import { SortType } from "./proto-state/derived-types";

const countRegex = /count(?=[^(]*\()/i;
const sumRegex = /sum(?=[^(]*\()/i;

export function isSummableMeasure(measure: MetricsViewSpecMeasureV2): boolean {
  const expression = measure.expression?.toLowerCase();
  return !!(expression?.match(countRegex) || expression?.match(sumRegex));
}

/**
 * Returns a sanitized column name appropriate for use in e.g. filters.
 *
 * Even though this is a one-liner, we externalize it as a function
 * becuase it is used in a few places and we want to make sure we
 * are consistent in how we handle this.
 */
export function getDimensionColumn(
  dimension: MetricsViewSpecDimensionV2,
): string {
  return (dimension?.column || dimension?.name) as string;
}

export function prepareSortedQueryBody(
  dimensionName: string,
  measureNames: string[],
  timeControls: TimeControlState,
  // Note: sortMeasureName may be null if we are sorting by dimension values
  sortMeasureName: string | null,
  sortType: SortType,
  sortAscending: boolean,
  whereFilterForDimension: V1Expression,
  limit: number,
): QueryServiceMetricsViewAggregationBody {
  let comparisonTimeRange: V1TimeRange | undefined = {
    start: timeControls.comparisonTimeStart,
    end: timeControls.comparisonTimeEnd,
  };

  const measures = measureNames.map(
    (n) =>
      <V1MetricsViewAggregationMeasure>{
        name: n,
      },
  );

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

  if (
    comparisonTimeRange?.start &&
    comparisonTimeRange?.end &&
    sortMeasureName
  ) {
    measures.push(
      {
        name: sortMeasureName + ComparisonDeltaPreviousSuffix,
        comparisonValue: {
          measure: sortMeasureName,
        },
      },
      {
        name: sortMeasureName + ComparisonDeltaAbsoluteSuffix,
        comparisonDelta: {
          measure: sortMeasureName,
        },
      },
      {
        name: sortMeasureName + ComparisonDeltaRelativeSuffix,
        comparisonRatio: {
          measure: sortMeasureName,
        },
      },
    );
  }

  return {
    dimensions: [
      {
        name: dimensionName,
      },
    ],
    measures,
    timeRange: {
      start: timeControls.timeStart,
      end: timeControls.timeEnd,
    },
    comparisonTimeRange,
    sort: [
      {
        desc: !sortAscending,
        name: sortMeasureName,
      },
    ],
    where: sanitiseExpression(whereFilterForDimension, undefined),
    limit: limit.toString(),
    offset: "0",
  };
}
