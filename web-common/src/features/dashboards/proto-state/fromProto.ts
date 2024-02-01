import { protoBase64, type Timestamp } from "@bufbuild/protobuf";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { FromProtoOperationMap } from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";
import { convertFilterToExpression } from "@rilldata/web-common/features/dashboards/proto-state/filter-converter";
import {
  createAndExpression,
  filterIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  BOOLEANS,
  INTEGERS,
  isFloat,
} from "@rilldata/web-common/lib/duckdb-data-types";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type { Expression } from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_LeaderboardContextColumn,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimensionV2,
  type StructTypeField,
  V1Expression,
  V1MetricsView,
  type V1MetricsViewSpec,
  type V1StructType,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

// TODO: make a follow up PR to use the one from the proto directly
const LeaderboardContextColumnReverseMap: Record<
  DashboardState_LeaderboardContextColumn,
  LeaderboardContextColumn
> = {
  [DashboardState_LeaderboardContextColumn.UNSPECIFIED]:
    LeaderboardContextColumn.HIDDEN,
  [DashboardState_LeaderboardContextColumn.PERCENT]:
    LeaderboardContextColumn.PERCENT,
  [DashboardState_LeaderboardContextColumn.DELTA_PERCENT]:
    LeaderboardContextColumn.DELTA_PERCENT,
  [DashboardState_LeaderboardContextColumn.DELTA_ABSOLUTE]:
    LeaderboardContextColumn.DELTA_ABSOLUTE,
  [DashboardState_LeaderboardContextColumn.HIDDEN]:
    LeaderboardContextColumn.HIDDEN,
};

export function getDashboardStateFromUrl(
  urlState: string,
  metricsView: V1MetricsView,
  schema: V1StructType,
): Partial<MetricsExplorerEntity> {
  return getDashboardStateFromProto(
    base64ToProto(decodeURIComponent(urlState)),
    metricsView,
    schema,
  );
}

export function getDashboardStateFromProto(
  binary: Uint8Array,
  metricsView: V1MetricsViewSpec,
  schema: V1StructType,
): Partial<MetricsExplorerEntity> {
  const dashboard = DashboardState.fromBinary(binary);
  const entity: Partial<MetricsExplorerEntity> = {};

  if (dashboard.filters) {
    // backwards compatibility for our older filter format
    entity.whereFilter = convertFilterToExpression(dashboard.filters);
    // older values could be strings for non-string values,
    // so we correct them using metrics view schema
    entity.whereFilter =
      correctFilterValues(
        entity.whereFilter,
        metricsView.dimensions ?? [],
        schema,
      ) ?? createAndExpression([]);
  } else if (dashboard.where) {
    entity.whereFilter = fromExpressionProto(dashboard.where);
  }
  if (dashboard.having) {
    entity.dimensionThresholdFilters = dashboard.having.map((h) => ({
      name: h.name,
      filter: fromExpressionProto(h.filter as Expression) as V1Expression,
    }));
  }
  if (dashboard.compareTimeRange) {
    entity.selectedComparisonTimeRange = fromTimeRangeProto(
      dashboard.compareTimeRange,
    );
    // backwards compatibility
    correctComparisonTimeRange(entity.selectedComparisonTimeRange);
  }
  entity.showTimeComparison = Boolean(dashboard.showTimeComparison);

  entity.selectedTimeRange = dashboard.timeRange
    ? fromTimeRangeProto(dashboard.timeRange)
    : undefined;
  if (dashboard.timeGrain && entity.selectedTimeRange) {
    entity.selectedTimeRange.interval = fromTimeGrainProto(dashboard.timeGrain);
  }

  if (dashboard.scrubRange) {
    entity.selectedScrubRange = fromTimeRangeProto(
      dashboard.scrubRange,
    ) as ScrubRange;
    entity.lastDefinedScrubRange = fromTimeRangeProto(
      dashboard.scrubRange,
    ) as ScrubRange;
  }

  if (dashboard.leaderboardMeasure) {
    entity.leaderboardMeasureName = dashboard.leaderboardMeasure;
  }
  if (dashboard.selectedDimension) {
    entity.selectedDimensionName = dashboard.selectedDimension;
  } else {
    entity.selectedDimensionName = undefined;
  }
  if (dashboard.expandedMeasure) {
    entity.expandedMeasureName = dashboard.expandedMeasure;
  } else {
    entity.expandedMeasureName = undefined;
  }
  if (dashboard.comparisonDimension) {
    entity.selectedComparisonDimension = dashboard.comparisonDimension;
  } else {
    entity.selectedComparisonDimension = undefined;
  }
  if (dashboard.pinIndex !== undefined) {
    entity.pinIndex = dashboard.pinIndex;
  }
  if (dashboard.selectedTimezone) {
    entity.selectedTimezone = dashboard.selectedTimezone;
  }

  if (dashboard.allMeasuresVisible) {
    entity.allMeasuresVisible = true;
    entity.visibleMeasureKeys = new Set(
      metricsView.measures?.map((measure) => measure.name) ?? [],
    ) as Set<string>;
  } else if (dashboard.visibleMeasures) {
    entity.allMeasuresVisible = false;
    entity.visibleMeasureKeys = new Set(dashboard.visibleMeasures);
  }

  if (dashboard.allDimensionsVisible) {
    entity.allDimensionsVisible = true;
    entity.visibleDimensionKeys = new Set(
      metricsView.dimensions?.map((measure) => measure.name) ?? [],
    ) as Set<string>;
  } else if (dashboard.visibleDimensions) {
    entity.allDimensionsVisible = false;
    entity.visibleDimensionKeys = new Set(dashboard.visibleDimensions);
  }

  if (dashboard.leaderboardContextColumn !== undefined) {
    entity.leaderboardContextColumn =
      LeaderboardContextColumnReverseMap[dashboard.leaderboardContextColumn];
  }

  if (dashboard.leaderboardSortDirection) {
    entity.sortDirection = dashboard.leaderboardSortDirection;
  }
  if (dashboard.leaderboardSortType) {
    entity.dashboardSortType = dashboard.leaderboardSortType;
  }

  return entity;
}

export function base64ToProto(message: string) {
  return protoBase64.dec(message);
}

function fromExpressionProto(expression: Expression): V1Expression | undefined {
  switch (expression.expression.case) {
    case "ident":
      return {
        ident: expression.expression.value,
      };

    case "val":
      return {
        val:
          expression.expression.value.kind.case === "nullValue"
            ? null
            : expression.expression.value.kind.value,
      };

    case "cond":
      return {
        cond: {
          op: FromProtoOperationMap[expression.expression.value.op],
          exprs: expression.expression.value.exprs
            .map((e) => fromExpressionProto(e))
            .filter((e): e is V1Expression => e !== undefined),
        },
      };
  }
}

function correctFilterValues(
  filter: V1Expression,
  dimensions: MetricsViewSpecDimensionV2[],
  schema: V1StructType,
) {
  return filterIdentifiers(filter, (e, ident) => {
    const dim = dimensions?.find((d) => d.name === ident);
    // ignore if dimension is not present anymore
    if (!dim) return false;
    const field = schema.fields?.find((f) => f.name === ident);
    // ignore if field is not found
    if (!field) return false;

    if (e.cond?.exprs) {
      e.cond.exprs =
        e.cond.exprs.map((e, i) => {
          if (i === 0) return e; // 1st expr is always the identifier
          return correctFilterValue(e, field);
        }) ?? [];
    }
    return true;
  });
}

function correctFilterValue(
  valueExpr: V1Expression,
  field: StructTypeField,
): V1Expression {
  if (valueExpr.val === null) {
    return valueExpr;
  }
  if (typeof valueExpr.val === "string") {
    // older filters were storing everything as strings
    return {
      val: correctStringFilterValue(valueExpr.val, field),
    };
  }
  return valueExpr;
}

function correctStringFilterValue(val: string, field: StructTypeField) {
  if (!field.type?.code) return val;

  if (INTEGERS.has(field.type?.code)) {
    return Number.parseInt(val);
  } else if (isFloat(field.type?.code)) {
    return Number.parseFloat(val);
  } else if (BOOLEANS.has(field.type?.code)) {
    return val === "true";
  } else {
    // TODO: other types
    return val;
  }
}

function fromTimeRangeProto(timeRange: DashboardTimeRange) {
  const selectedTimeRange: DashboardTimeControls = {
    name: timeRange.name,
  } as DashboardTimeControls;
  // backwards compatibility
  if (timeRange.name && timeRange.name in TimeRangePreset) {
    selectedTimeRange.name = TimeRangePreset[timeRange.name];
  }

  if (timeRange.timeStart) {
    selectedTimeRange.start = fromTimeProto(timeRange.timeStart);
  }
  if (timeRange.timeEnd) {
    selectedTimeRange.end = fromTimeProto(timeRange.timeEnd);
  }

  return selectedTimeRange;
}

function correctComparisonTimeRange(
  comparisonTimeRange: DashboardTimeControls,
) {
  switch (comparisonTimeRange.name as string) {
    case "CONTIGUOUS":
      comparisonTimeRange.name = TimeComparisonOption.CONTIGUOUS;
      break;
    case "P1D":
      comparisonTimeRange.name = TimeComparisonOption.DAY;
      break;
    case "P1W":
      comparisonTimeRange.name = TimeComparisonOption.WEEK;
      break;
    case "P1M":
      comparisonTimeRange.name = TimeComparisonOption.MONTH;
      break;
    case "P3M":
      comparisonTimeRange.name = TimeComparisonOption.QUARTER;
      break;
    case "P1Y":
      comparisonTimeRange.name = TimeComparisonOption.YEAR;
      break;
  }
}

function fromTimeProto(timestamp: Timestamp) {
  return new Date(Number(timestamp.seconds));
}

function fromTimeGrainProto(timeGrain: TimeGrain): V1TimeGrain {
  switch (timeGrain) {
    case TimeGrain.UNSPECIFIED:
    default:
      return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
    case TimeGrain.MILLISECOND:
      return V1TimeGrain.TIME_GRAIN_MILLISECOND;
    case TimeGrain.SECOND:
      return V1TimeGrain.TIME_GRAIN_SECOND;
    case TimeGrain.MINUTE:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case TimeGrain.HOUR:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeGrain.DAY:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case TimeGrain.WEEK:
      return V1TimeGrain.TIME_GRAIN_WEEK;
    case TimeGrain.MONTH:
      return V1TimeGrain.TIME_GRAIN_MONTH;
    case TimeGrain.QUARTER:
      return V1TimeGrain.TIME_GRAIN_QUARTER;
    case TimeGrain.YEAR:
      return V1TimeGrain.TIME_GRAIN_YEAR;
  }
}
