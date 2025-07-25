import { getAggregationDimensionFromFieldName } from "@rilldata/web-common/features/dashboards/aggregation-request/dimension-utils.ts";
import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import type { FiltersState } from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers.ts";
import type {
  V1ExploreSpec,
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";

export type AggregationRequestUpdater = (
  aggregationRequest: V1MetricsViewAggregationRequest,
) => V1MetricsViewAggregationRequest;

export function buildAggregationRequest(
  baseAggregationRequest: V1MetricsViewAggregationRequest,
  updaters: AggregationRequestUpdater[],
) {
  let aggregationRequest = baseAggregationRequest;
  for (const updater of updaters) {
    aggregationRequest = updater(aggregationRequest);
  }
  return aggregationRequest;
}

export const aggregationRequestWithTimeRange = (
  exploreSpec: V1ExploreSpec,
  timeControlArgs: TimeControlState,
) => {
  return (aggregationRequest: V1MetricsViewAggregationRequest) => {
    const timeRange = mapSelectedTimeRangeToV1TimeRange(
      timeControlArgs.selectedTimeRange,
      timeControlArgs.selectedTimezone,
      exploreSpec,
    );
    const comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
      timeControlArgs.selectedComparisonTimeRange,
      timeControlArgs.showTimeComparison,
      timeRange,
    );
    return {
      ...aggregationRequest,
      timeRange,
      comparisonTimeRange,
    };
  };
};

export const aggregationRequestWithFilters = (filtersState: FiltersState) => {
  return (aggregationRequest: V1MetricsViewAggregationRequest) => {
    const whereFilter = sanitiseExpression(
      mergeDimensionAndMeasureFilters(
        filtersState.whereFilter,
        filtersState.dimensionThresholdFilters,
      ),
      undefined,
    );
    return {
      ...aggregationRequest,
      where: whereFilter,
    };
  };
};

export const aggregationRequestWithRowsAndColumns = ({
  exploreSpec,
  rows,
  columns,
  showTimeComparison,
  selectedTimezone,
}: {
  exploreSpec: V1ExploreSpec;
  rows: string[];
  columns: string[];
  showTimeComparison: boolean;
  selectedTimezone: string;
}) => {
  return (aggregationRequest: V1MetricsViewAggregationRequest) => {
    const allFields = new Set<string>([...rows, ...columns]);
    const isFlat = rows.length === 0;

    // Get measures defined as columns. We do allow adding measures as rows so need to check it.
    const measures = columns
      .filter((col) => exploreSpec.measures?.includes(col))
      .flatMap((measureName) => {
        const group = [{ name: measureName }];

        if (showTimeComparison) {
          group.push(
            { name: `${measureName}${COMPARISON_DELTA}` },
            { name: `${measureName}${COMPARISON_PERCENT}` },
          );
        }

        return group;
      });

    // Get dimensions defined as rows
    const dimensionsFromRows: V1MetricsViewAggregationDimension[] = rows.map(
      (row) => getAggregationDimensionFromFieldName(row, selectedTimezone),
    );

    // Get dimensions defined as columns
    const dimensionsFromColumns: V1MetricsViewAggregationDimension[] = columns
      .filter((col) => !exploreSpec.measures?.includes(col))
      .map((col) =>
        getAggregationDimensionFromFieldName(col, selectedTimezone),
      );

    // only add column dimensions as pivot if it is a non-flat view
    const pivotOn = !isFlat
      ? dimensionsFromColumns.map((d) => d.alias ?? d.name!)
      : [];

    // Get the full list of dimensions
    const dimensions = [...dimensionsFromRows, ...dimensionsFromColumns];

    // Get the updated sort based on the new measures and dimensions
    const updatedAggregationSort = getUpdatedAggregationSort(
      aggregationRequest,
      measures,
      dimensions,
      pivotOn,
      allFields,
    );

    return {
      ...aggregationRequest,
      measures,
      dimensions,
      pivotOn: !pivotOn.length ? undefined : pivotOn,
      sort: updatedAggregationSort,
    };
  };
};

function getUpdatedAggregationSort(
  aggregationRequest: V1MetricsViewAggregationRequest,
  measures: V1MetricsViewAggregationMeasure[],
  dimensions: V1MetricsViewAggregationDimension[],
  pivotOn: string[],
  allFields: Set<string>,
) {
  const hasPivot = pivotOn.length > 0;
  const sort: V1MetricsViewAggregationSort[] =
    aggregationRequest.sort?.filter((s) => {
      if (!allFields.has(s.name!)) return false;
      if (!hasPivot) return true;
      // When there is a pivot we cannot sort by measure or the pivoted dimension
      return (
        !measures.find((m) => m.name === s.name) && !pivotOn.includes(s.name!)
      );
    }) ?? [];

  // Old sort is still valid. So retain it
  if (sort.length > 0) {
    return sort;
  }

  // Get the sort from the 1st measure
  let sortField: string | undefined = measures?.[0]?.name;
  let sortFieldIsMeasure = !!sortField;
  // If there is no measure or if we are pivoting the get the 1st non-pivoted dimension
  if (!sortField || hasPivot) {
    const nonPivotedDimension = dimensions.find(
      (d) => !pivotOn.includes(d.alias!),
    );
    sortField = nonPivotedDimension?.alias || nonPivotedDimension?.name;
    sortFieldIsMeasure = false;
  }

  if (!sortField) {
    return [];
  }

  return [
    {
      desc: sortFieldIsMeasure,
      name: sortField,
    },
  ];
}
