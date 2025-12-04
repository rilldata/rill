import {
  type PivotChipData,
  PivotChipType,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import {
  filterIdentifiers,
  maybeConvertEqualityToInExpressions,
  flattenInExpressionValues,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { parseTimeRangeFromFilters } from "@rilldata/web-common/features/explore-mappers/parse-time-range-from-filters.ts";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types.ts";
import { DateTimeUnitToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains.ts";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types.ts";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import {
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  V1Operation,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type {
  Dimension,
  Expression,
  Measure,
  Schema as MetricsResolverQuery,
  Sort,
  TimeRange,
} from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import type { SortingState } from "@tanstack/svelte-table";

export function mapMetricsResolverQueryToDashboard(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
  query: MetricsResolverQuery,
) {
  // Build partial ExploreState directly from Query
  const partialExploreState: Partial<ExploreState> = {};

  const { exploreDimensions, timeDimensions } = getValidDimensions(
    metricsViewSpec,
    exploreSpec,
    query.dimensions ?? [],
  );

  if (exploreDimensions.length > 0) {
    partialExploreState.visibleDimensions = exploreDimensions;
    partialExploreState.allDimensionsVisible =
      exploreDimensions.length === exploreSpec.dimensions?.length;
  }

  // Convert measures
  if (query.measures) {
    const measureNames = query.measures.map((m) => m.name).filter(Boolean);

    // Validate measures exist in the metrics view
    const validMeasures = measureNames.filter(
      (name) =>
        metricsViewSpec.measures?.some((m) => m.name === name) &&
        exploreSpec.measures?.includes(name),
    );

    if (validMeasures.length > 0) {
      partialExploreState.visibleMeasures = validMeasures;
      partialExploreState.allMeasuresVisible =
        validMeasures.length === exploreSpec.measures?.length;
    }
  }

  // Convert time ranges
  partialExploreState.selectedTimeRange =
    mapResolverTimeRangeToDashboardControls(query.time_range);
  if (query.comparison_time_range) {
    partialExploreState.selectedComparisonTimeRange =
      mapResolverTimeRangeToDashboardControls(query.comparison_time_range);
    partialExploreState.showTimeComparison = true;
  }

  // Convert where filter
  if (query.where) {
    partialExploreState.whereFilter = mapResolverExpressionToV1Expression(
      query.where,
    );
  }

  maybeGetTimeRangeFromFilter(
    partialExploreState,
    metricsViewSpec,
    timeRangeSummary,
  );

  // Convert sort
  if (query.sort) {
    mapSort(query.measures ?? [], query.sort, partialExploreState);
  }

  // Set default timezone if not specified
  if (query.time_zone) {
    partialExploreState.selectedTimezone = query.time_zone;
  }

  mapActivePage(query, partialExploreState, timeDimensions);

  return partialExploreState;
}

const OperationMap: Record<string, V1Operation> = {
  "": V1Operation.OPERATION_UNSPECIFIED,
  eq: V1Operation.OPERATION_EQ,
  neq: V1Operation.OPERATION_NEQ,
  lt: V1Operation.OPERATION_LT,
  lte: V1Operation.OPERATION_LTE,
  gt: V1Operation.OPERATION_GT,
  gte: V1Operation.OPERATION_GTE,
  in: V1Operation.OPERATION_IN,
  nin: V1Operation.OPERATION_NIN,
  ilike: V1Operation.OPERATION_LIKE,
  nilike: V1Operation.OPERATION_NLIKE,
  or: V1Operation.OPERATION_OR,
  and: V1Operation.OPERATION_AND,
};
export function mapResolverExpressionToV1Expression(
  expr: Expression | undefined,
): V1Expression | undefined {
  if (!expr) return undefined;

  if (expr.name) {
    return { ident: expr.name };
  }

  if (expr.val) {
    return { val: expr.val };
  }

  if (expr.cond) {
    const condExpr = maybeConvertEqualityToInExpressions({
      cond: {
        op: OperationMap[expr.cond.op] || V1Operation.OPERATION_UNSPECIFIED,
        exprs: expr.cond.exprs
          ?.map(mapResolverExpressionToV1Expression)
          .filter(Boolean) as V1Expression[] | undefined,
      },
    });
    return flattenInExpressionValues(condExpr);
  }

  if (expr.subquery) {
    return {
      subquery: {
        dimension: expr.subquery.dimension.name,
        measures: expr.subquery.measures.map((m) => m.name),
        where: mapResolverExpressionToV1Expression(expr.subquery.where),
        having: mapResolverExpressionToV1Expression(expr.subquery.having),
      },
    };
  }

  return {};
}

function getValidDimensions(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  dimensions: Dimension[],
) {
  const mvDimensions = new Set<string>(
    metricsViewSpec.dimensions?.map((d) => d.name!) ?? [],
  );
  const exploreDimensions = dimensions
    .filter(
      (d) =>
        mvDimensions.has(d.name) && exploreSpec.dimensions?.includes(d.name),
    )
    .map((d) => d.name);
  // TODO: handle multiple timestamp metrics views
  const timeDimensions = dimensions.filter(
    (d) => d.compute?.time_floor?.dimension === metricsViewSpec.timeDimension,
  );

  return {
    exploreDimensions,
    timeDimensions,
  };
}

function mapResolverTimeRangeToDashboardControls(
  timeRange: TimeRange | undefined,
): DashboardTimeControls {
  // Default to "All Time" when no time range is specified
  if (!timeRange)
    return { name: TimeRangePreset.ALL_TIME } as DashboardTimeControls;

  if (timeRange.start && timeRange.end) {
    return {
      name: TimeRangePreset.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (timeRange.expression) {
    return {
      name: timeRange.expression,
    } as DashboardTimeControls;
  } else if (timeRange.iso_duration) {
    return {
      name: timeRange.iso_duration,
    } as DashboardTimeControls;
  }

  // Fallback to all-time
  return { name: TimeRangePreset.ALL_TIME } as DashboardTimeControls;
}

function maybeGetTimeRangeFromFilter(
  partialExploreState: Partial<ExploreState>,
  metricsViewSpec: V1MetricsViewSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  if (
    !partialExploreState.whereFilter ||
    !metricsViewSpec.timeDimension ||
    !timeRangeSummary
  ) {
    return;
  }
  const tr = parseTimeRangeFromFilters(
    partialExploreState.whereFilter,
    metricsViewSpec.timeDimension,
    partialExploreState.selectedTimezone ?? "UTC",
    timeRangeSummary,
  );
  if (!tr) return;
  partialExploreState.selectedTimeRange = tr;
  // Remove any filter that apply on time dimension
  partialExploreState.whereFilter = filterIdentifiers(
    partialExploreState.whereFilter,
    (_, i) => i !== metricsViewSpec.timeDimension,
  );
}

function mapSort(
  measures: Measure[],
  sort: Sort[] | undefined,
  partialExploreState: Partial<ExploreState>,
) {
  if (!sort?.length) return;
  const sortField = sort[0];

  const measure = measures.find((m) => m.name === sortField.name);
  if (!measure) return;
  const { name, type } = getMeasureNameAndType(measure);

  partialExploreState.leaderboardSortByMeasureName = name;
  partialExploreState.sortDirection = sortField.desc
    ? SortDirection.DESCENDING
    : SortDirection.ASCENDING;
  partialExploreState.dashboardSortType = type;
}

function getMeasureNameAndType(measure: Measure) {
  if (measure.compute?.comparison_delta?.measure) {
    return {
      name: measure.compute.comparison_delta.measure,
      type: DashboardState_LeaderboardSortType.DELTA_ABSOLUTE,
    };
  }

  if (measure.compute?.comparison_ratio?.measure) {
    return {
      name: measure.compute.comparison_ratio.measure,
      type: DashboardState_LeaderboardSortType.DELTA_PERCENT,
    };
  }

  if (measure.compute?.percent_of_total?.measure) {
    return {
      name: measure.compute.percent_of_total.measure,
      type: DashboardState_LeaderboardSortType.PERCENT,
    };
  }

  return {
    name: measure.name,
    type: DashboardState_LeaderboardSortType.VALUE,
  };
}

function mapActivePage(
  query: MetricsResolverQuery,
  partialExploreState: Partial<ExploreState>,
  timeDimensions: Dimension[],
) {
  const hasExactlyOneMeasure =
    (partialExploreState.visibleMeasures?.length ?? 0) === 1;
  const visibleDimensions = partialExploreState.visibleDimensions ?? [];
  const showTDD =
    timeDimensions.length === 1 &&
    hasExactlyOneMeasure &&
    visibleDimensions.length <= 1;
  const showDimensionTable = !showTDD && visibleDimensions.length === 1;
  const showPivot = timeDimensions.length > 1 || visibleDimensions.length > 1;

  if (showTDD) {
    partialExploreState.tdd = {
      expandedMeasureName: partialExploreState.visibleMeasures?.[0],
      chartType: TDDChart.DEFAULT,
      pinIndex: -1,
    };
    partialExploreState.selectedComparisonDimension = visibleDimensions[0];
    partialExploreState.activePage =
      DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;

    const timeDimension = timeDimensions[0];
    const shouldUpdateTimeGrain =
      partialExploreState.selectedTimeRange &&
      timeDimension?.compute?.time_floor?.grain;
    if (shouldUpdateTimeGrain) {
      // Selected time grain is used in TDD's pivot table at the bottom.
      partialExploreState.selectedTimeRange!.interval =
        DateTimeUnitToV1TimeGrain[timeDimension.compute!.time_floor!.grain];
    }
  } else if (showDimensionTable) {
    partialExploreState.selectedDimensionName = visibleDimensions[0];
    partialExploreState.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
  } else if (showPivot) {
    partialExploreState.pivot = mapPivot(
      query,
      partialExploreState,
      timeDimensions,
    );
    partialExploreState.activePage = DashboardState_ActivePage.PIVOT;
  }
}

function mapPivot(
  query: MetricsResolverQuery,
  partialExploreState: Partial<ExploreState>,
  timeDimensions: Dimension[],
): PivotState {
  const columns: PivotChipData[] = timeDimensions
    .map((d) => {
      if (!d.compute?.time_floor?.grain) return undefined;
      const grain = DateTimeUnitToV1TimeGrain[d.compute.time_floor.grain];
      return {
        id: grain,
        title: d.compute.time_floor.grain,
        type: PivotChipType.Time,
      };
    })
    .filter(Boolean) as PivotChipData[];

  if (partialExploreState.visibleDimensions) {
    columns.push(
      ...partialExploreState.visibleDimensions.map((d) => ({
        id: d,
        title: d,
        type: PivotChipType.Dimension,
      })),
    );
  }

  if (partialExploreState.visibleMeasures) {
    columns.push(
      ...partialExploreState.visibleMeasures.map((m) => ({
        id: m,
        title: m,
        type: PivotChipType.Measure,
      })),
    );
  }

  const sorting: SortingState =
    query.sort?.map((s) => ({ id: s.name, desc: !!s.desc })) ?? [];

  return {
    rows: [],
    columns,
    expanded: {},
    sorting,
    columnPage: 0,
    rowPage: 0,
    enableComparison: false,
    tableMode: "flat",
    activeCell: null,
  };
}
