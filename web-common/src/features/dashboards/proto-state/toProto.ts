import {
  NullValue,
  PartialMessage,
  Timestamp,
  Value,
  protoBase64,
} from "@bufbuild/protobuf";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  DashboardTimeControls,
  ScrubRange,
} from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  Condition,
  Expression,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/expression_pb";
import {
  TimeGrain,
  TimeGrain as TimeGrainProto,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_LeaderboardContextColumn,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { V1Operation, V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { ToProtoOperationMap } from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";

// TODO: make a follow up PR to use the one from the proto directly
const LeaderboardContextColumnMap: Record<
  LeaderboardContextColumn,
  DashboardState_LeaderboardContextColumn
> = {
  [LeaderboardContextColumn.PERCENT]:
    DashboardState_LeaderboardContextColumn.PERCENT,
  [LeaderboardContextColumn.DELTA_PERCENT]:
    DashboardState_LeaderboardContextColumn.DELTA_PERCENT,
  [LeaderboardContextColumn.DELTA_ABSOLUTE]:
    DashboardState_LeaderboardContextColumn.DELTA_ABSOLUTE,
  [LeaderboardContextColumn.HIDDEN]:
    DashboardState_LeaderboardContextColumn.HIDDEN,
};

export function getProtoFromDashboardState(
  metrics: MetricsExplorerEntity,
): string {
  if (!metrics) return "";

  const state: PartialMessage<DashboardState> = {};
  if (metrics.whereFilter) {
    state.where = toExpressionProto(metrics.whereFilter);
  }
  if (metrics.havingFilter) {
    state.having = toExpressionProto(metrics.havingFilter);
  }
  if (metrics.selectedTimeRange) {
    state.timeRange = toTimeRangeProto(metrics.selectedTimeRange);
    if (metrics.selectedTimeRange.interval) {
      state.timeGrain = toTimeGrainProto(metrics.selectedTimeRange.interval);
    }
  }
  if (metrics.selectedComparisonTimeRange) {
    state.compareTimeRange = toTimeRangeProto(
      metrics.selectedComparisonTimeRange,
    );
  }
  if (metrics.lastDefinedScrubRange) {
    state.scrubRange = toScrubProto(metrics.lastDefinedScrubRange);
  }
  state.showTimeComparison = Boolean(metrics.showTimeComparison);
  if (metrics.selectedComparisonDimension) {
    state.comparisonDimension = metrics.selectedComparisonDimension;
  }
  if (metrics.selectedTimezone) {
    state.selectedTimezone = metrics.selectedTimezone;
  }
  if (metrics.leaderboardMeasureName) {
    state.leaderboardMeasure = metrics.leaderboardMeasureName;
  }
  if (metrics.expandedMeasureName) {
    state.expandedMeasure = metrics.expandedMeasureName;
  }
  if (metrics.pinIndex !== undefined) {
    state.pinIndex = metrics.pinIndex;
  }
  if (metrics.selectedDimensionName) {
    state.selectedDimension = metrics.selectedDimensionName;
  }

  if (metrics.allMeasuresVisible) {
    state.allMeasuresVisible = true;
  } else if (metrics.visibleMeasureKeys) {
    state.visibleMeasures = [...metrics.visibleMeasureKeys];
  }

  if (metrics.allDimensionsVisible) {
    state.allDimensionsVisible = true;
  } else if (metrics.visibleDimensionKeys) {
    state.visibleDimensions = [...metrics.visibleDimensionKeys];
  }

  if (metrics.leaderboardContextColumn) {
    state.leaderboardContextColumn =
      LeaderboardContextColumnMap[metrics.leaderboardContextColumn];
  }

  if (metrics.sortDirection) {
    state.leaderboardSortDirection = metrics.sortDirection;
  }
  if (metrics.dashboardSortType) {
    state.leaderboardSortType = metrics.dashboardSortType;
  }

  const message = new DashboardState(state);
  return protoToBase64(message.toBinary());
}

function protoToBase64(proto: Uint8Array) {
  return protoBase64.enc(proto);
}

function toTimeRangeProto(range: DashboardTimeControls) {
  const timeRangeArgs: PartialMessage<DashboardTimeRange> = {
    name: range.name,
  };
  if (
    range.name === TimeRangePreset.CUSTOM ||
    range.name === TimeComparisonOption.CUSTOM
  ) {
    if (range.start) timeRangeArgs.timeStart = toTimeProto(range.start);
    if (range.end) timeRangeArgs.timeEnd = toTimeProto(range.end);
  }
  return new DashboardTimeRange(timeRangeArgs);
}

function toScrubProto(range: ScrubRange) {
  const timeRangeArgs: PartialMessage<DashboardTimeRange> = {
    name: TimeRangePreset.CUSTOM,
  };
  timeRangeArgs.timeStart = toTimeProto(range.start);
  timeRangeArgs.timeEnd = toTimeProto(range.end);

  return new DashboardTimeRange(timeRangeArgs);
}

function toTimeProto(date: Date) {
  return new Timestamp({
    seconds: BigInt(date.getTime()),
  });
}

function toTimeGrainProto(timeGrain: V1TimeGrain) {
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_UNSPECIFIED:
    default:
      return TimeGrain.UNSPECIFIED;
    case V1TimeGrain.TIME_GRAIN_MILLISECOND:
      return TimeGrain.MILLISECOND;
    case V1TimeGrain.TIME_GRAIN_SECOND:
      return TimeGrain.SECOND;
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return TimeGrainProto.MINUTE;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return TimeGrainProto.HOUR;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return TimeGrainProto.DAY;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return TimeGrainProto.WEEK;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return TimeGrainProto.MONTH;
    case V1TimeGrain.TIME_GRAIN_QUARTER:
      return TimeGrainProto.QUARTER;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return TimeGrainProto.YEAR;
  }
}

function toExpressionProto(expression: V1Expression | undefined) {
  if (!expression) return undefined;
  if ("ident" in expression) {
    return new Expression({
      expression: {
        case: "ident",
        value: expression.ident as string,
      },
    });
  }
  if ("val" in expression) {
    return new Expression({
      expression: {
        case: "val",
        value: toPbValue(expression.val),
      },
    });
  }
  if (expression.cond) {
    return new Expression({
      expression: {
        case: "cond",
        value: new Condition({
          op: ToProtoOperationMap[
            expression.cond.op ?? V1Operation.OPERATION_UNSPECIFIED
          ],
          exprs: expression.cond.exprs?.map((e) => toExpressionProto(e)) ?? [],
        }),
      },
    });
  }
  return undefined;
}

function toPbValue(val: unknown) {
  if (val === null) {
    return new Value({
      kind: {
        case: "nullValue",
        value: NullValue.NULL_VALUE,
      },
    });
  }
  switch (typeof val) {
    case "string":
      return new Value({
        kind: {
          case: "stringValue",
          value: val,
        },
      });
    case "number":
      return new Value({
        kind: {
          case: "numberValue",
          value: val,
        },
      });
    case "boolean":
      return new Value({
        kind: {
          case: "boolValue",
          value: val,
        },
      });
    // TODO: other options are not currently in a filter. but we might need them in future
    default:
      // force as string for unknown types. this is the older behaviour
      return new Value({
        kind: {
          case: "stringValue",
          value: JSON.stringify(val),
        },
      });
  }
}
