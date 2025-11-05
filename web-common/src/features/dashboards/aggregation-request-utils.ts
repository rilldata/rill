import {
  getAggregationDimensionFromFieldName,
  getDimensionForTimeField,
} from "@rilldata/web-common/features/dashboards/aggregation-request/dimension-utils.ts";
import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils.ts";
import { MeasureModifierSuffixRegex } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import {
  prepareMeasureForComparison,
  splitPivotChips,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  ComparisonModifierSuffixRegex,
  PivotChipType,
  type PivotState,
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
  V1MetricsViewSpec,
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
    const isFlat = rows.length === 0;

    // Get measures defined as columns. We do allow adding measures as rows so need to check it.
    const measures = columns
      .filter((col) => exploreSpec.measures?.includes(col))
      .flatMap((measureName) => {
        const group: V1MetricsViewAggregationMeasure[] = [
          { name: measureName },
        ];

        if (showTimeComparison && aggregationRequest.comparisonTimeRange) {
          group.push(...getComparisonRequestMeasures(measureName));
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
    const updatedAggregationSort = getUpdatedAggregationSort({
      aggregationRequest,
      measures,
      dimensions,
      pivotOn,
      selectedTimezone,
    });

    return {
      ...aggregationRequest,
      measures,
      dimensions,
      pivotOn: !pivotOn.length ? undefined : pivotOn,
      sort: updatedAggregationSort,
    };
  };
};

/**
 * Splits dimensions and measures from a {@link V1MetricsViewAggregationRequest} into
 * logical rows and columns based on the request's pivot configuration.
 *
 * @returns An object containing three arrays:
 *   - `rows`: Dimensions that are not part of the pivot (specified by `pivotedOn` field in the request)
 *   - `dimensionColumns`: Dimensions that are part of the pivot (specified by `pivotedOn` field in the request)
 *   - `measureColumns`: Measure fields that represent the values in the table, includes just the base measures.
 */
export function splitDimensionsAndMeasuresAsRowsAndColumns(
  aggregationRequest: V1MetricsViewAggregationRequest,
) {
  const pivotedOn = new Set<string>(aggregationRequest.pivotOn ?? []);
  const isFlat = aggregationRequest.pivotOn === undefined;

  const rows =
    aggregationRequest.dimensions?.filter(
      (dimension) =>
        !isFlat &&
        !pivotedOn.has(dimension.alias!) &&
        !pivotedOn.has(dimension.name!),
    ) ?? [];

  const dimensionColumns =
    aggregationRequest.dimensions?.filter(
      (dimension) =>
        isFlat ||
        pivotedOn.has(dimension.alias!) ||
        pivotedOn.has(dimension.name!),
    ) ?? [];

  const measureColumns =
    aggregationRequest.measures?.filter(
      (measure) =>
        !MeasureModifierSuffixRegex.test(measure.name!) &&
        !ComparisonModifierSuffixRegex.test(measure.name!),
    ) ?? [];

  return {
    rows,
    dimensionColumns,
    measureColumns,
  };
}

export function aggregationRequestWithPivotState(
  pivotState: PivotState,
  metricsViewSpec: V1MetricsViewSpec,
) {
  return (aggregationRequest: V1MetricsViewAggregationRequest) => {
    const isFlat = pivotState.tableMode === "flat";
    const columns = splitPivotChips(pivotState.columns);
    const timeDimension = metricsViewSpec.timeDimension ?? "";
    const timeZone = aggregationRequest.timeRange?.timeZone ?? "UTC";
    const enableComparison = !!aggregationRequest.comparisonTimeRange;

    const measures = columns.measure.flatMap((m) => {
      const measureName = m.id;
      const group = [{ name: measureName }];

      if (enableComparison) {
        group.push(
          { name: `${measureName}${COMPARISON_DELTA}` },
          { name: `${measureName}${COMPARISON_PERCENT}` },
        );
      }

      return group;
    });

    const allDimensions = [...pivotState.rows, ...columns.dimension].map((d) =>
      d.type === PivotChipType.Time
        ? getDimensionForTimeField(timeDimension, timeZone, d, !isFlat)
        : {
            name: d.id,
          },
    );

    const pivotOn = isFlat
      ? undefined
      : columns.dimension.map((d) =>
          d.type === PivotChipType.Time ? `Time ${d.title}` : d.id,
        );

    const rowDimensions = pivotState.rows.map((d) =>
      d.type === PivotChipType.Time
        ? getDimensionForTimeField(timeDimension, timeZone, d, true)
        : {
            name: d.id,
          },
    );

    let sort: V1MetricsViewAggregationSort[] = [];

    if (isFlat) {
      if (pivotState.sorting.length > 0) {
        sort = [
          {
            name: pivotState.sorting[0].id,
            desc: pivotState.sorting[0].desc,
          },
        ];
      } else {
        sort = [
          {
            desc: measures?.[0] ? true : false,
            name: measures?.[0]?.name || allDimensions?.[0]?.name,
          },
        ];
      }
    } else {
      // Sort by the dimensions in the pivot's rows
      sort = rowDimensions.map((d) => {
        return {
          name: d.alias ? d.alias : d.name,
          desc: pivotState.sorting.find((s) => s.id === d.name)?.desc ?? false,
        };
      });
    }

    return {
      ...aggregationRequest,
      measures: enableComparison
        ? prepareMeasureForComparison(measures)
        : measures,
      dimensions: allDimensions,
      pivotOn,
      sort,
    };
  };
}

function getUpdatedAggregationSort({
  aggregationRequest,
  measures,
  dimensions,
  pivotOn,
  selectedTimezone,
}: {
  aggregationRequest: V1MetricsViewAggregationRequest;
  measures: V1MetricsViewAggregationMeasure[];
  dimensions: V1MetricsViewAggregationDimension[];
  pivotOn: string[];
  selectedTimezone: string;
}) {
  const hasPivot = pivotOn.length > 0;
  const sort: V1MetricsViewAggregationSort[] =
    (aggregationRequest.sort
      ?.map((s) => {
        const isMeasure = measures.find((m) => m.name === s.name);
        // We cannot sort by measure when pivoting.
        if (isMeasure) return hasPivot ? undefined : s;

        const dim = getAggregationDimensionFromFieldName(
          s.name!,
          selectedTimezone,
        );
        const field = dim.alias ?? dim.name!;
        const isDimension = dimensions.find(
          (d) => (!!d.alias && d.alias === dim.alias) || d.name === dim.name,
        );
        return isDimension && !pivotOn.includes(field)
          ? {
              ...s,
              name: field,
            }
          : undefined;
      })
      .filter(Boolean) as V1MetricsViewAggregationSort[]) ?? [];

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
