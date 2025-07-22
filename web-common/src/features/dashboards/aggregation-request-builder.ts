import { getAggregationDimensionFromTimeDimension } from "@rilldata/web-common/features/dashboards/aggregation-request/dimension-utils.ts";
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
    const pivotOn: string[] = [];

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
    const dimensions: V1MetricsViewAggregationDimension[] = rows.map((d) =>
      getAggregationDimensionFromTimeDimension(d, selectedTimezone),
    );
    columns
      .filter((col) => !exploreSpec.measures?.includes(col))
      .forEach((col) => {
        if (exploreSpec.dimensions?.includes(col)) {
          dimensions.push({ name: col });
          if (!isFlat) pivotOn.push(col);
          return;
        }

        const dimension = getAggregationDimensionFromTimeDimension(
          col,
          selectedTimezone,
        );
        dimensions.push(dimension);
        if (!isFlat) pivotOn.push(dimension.alias ?? dimension.name!);
      });

    const sort = getUpdatedAggregationSort(
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
      sort,
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
  if (sort.length === 0) {
    let sortField: string | undefined = measures?.[0]?.name;
    let sortFieldIsMeasure = !!sortField;
    if (!sortField || hasPivot) {
      const sortDimension = dimensions.find((d) => !pivotOn.includes(d.alias!));
      sortField = sortDimension?.alias || sortDimension?.name;
      sortFieldIsMeasure = false;
    }
    if (sortField) {
      sort.push({
        desc: sortFieldIsMeasure,
        name: sortField,
      });
    }
  }

  return sort;
}
