import { protoBase64, type Timestamp } from "@bufbuild/protobuf";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { FromProtoOperationMap } from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";
import { convertFilterToExpression } from "@rilldata/web-common/features/dashboards/proto-state/filter-converter";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type { Expression } from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import type { MetricsViewFilter_Cond } from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_LeaderboardContextColumn,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1Expression,
  V1MetricsView,
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
  metricsView: V1MetricsView
): Partial<MetricsExplorerEntity> {
  return getDashboardStateFromProto(
    base64ToProto(decodeURIComponent(urlState)),
    metricsView
  );
}

export function getDashboardStateFromProto(
  binary: Uint8Array,
  metricsView: V1MetricsView
): Partial<MetricsExplorerEntity> {
  const dashboard = DashboardState.fromBinary(binary);
  const entity: Partial<MetricsExplorerEntity> = {
    filters: {
      include: [],
      exclude: [],
    },
  };

  if (dashboard.filters) {
    entity.whereFilter = convertFilterToExpression(dashboard.filters);
  } else if (dashboard.where) {
    entity.whereFilter = fromExpressionProto(dashboard.where);
  }
  if (dashboard.having) {
    entity.havingFilter = fromExpressionProto(dashboard.having);
  }
  if (dashboard.compareTimeRange) {
    entity.selectedComparisonTimeRange = fromTimeRangeProto(
      dashboard.compareTimeRange
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
      dashboard.scrubRange
    ) as ScrubRange;
    entity.lastDefinedScrubRange = fromTimeRangeProto(
      dashboard.scrubRange
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
      metricsView.measures?.map((measure) => measure.name) ?? []
    ) as Set<string>;
  } else if (dashboard.visibleMeasures) {
    entity.allMeasuresVisible = false;
    entity.visibleMeasureKeys = new Set(dashboard.visibleMeasures);
  }

  if (dashboard.allDimensionsVisible) {
    entity.allDimensionsVisible = true;
    entity.visibleDimensionKeys = new Set(
      metricsView.dimensions?.map((measure) => measure.name) ?? []
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

function fromFiltersProto(conditions: Array<MetricsViewFilter_Cond>) {
  return conditions.map((condition) => {
    return {
      name: condition.name,
      ...(condition.like?.length ? { like: condition.like } : {}),
      ...(condition.in?.length
        ? {
            in: condition.in.map((v) =>
              v.kind.case === "nullValue" ? null : v.kind.value
            ),
          }
        : {}),
    };
  });
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
  comparisonTimeRange: DashboardTimeControls
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

function fromExpressionProto(expression: Expression) {
  switch (expression.expression.case) {
    case "ident":
      return {
        ident: expression.expression.value,
      } as V1Expression;

    case "val":
      return {
        val:
          expression.expression.value.kind.case === "nullValue"
            ? null
            : expression.expression.value.kind.value,
      } as V1Expression;

    case "cond":
      return {
        cond: {
          op: FromProtoOperationMap[expression.expression.value.op],
          exprs: expression.expression.value.exprs.map((e) =>
            fromExpressionProto(e)
          ),
        },
      } as V1Expression;
  }
}
