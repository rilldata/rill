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
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  TimeGrain,
  TimeGrain as TimeGrainProto,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_LeaderboardContextColumn,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

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
  metrics: MetricsExplorerEntity
): string {
  if (!metrics) return "";

  const state: PartialMessage<DashboardState> = {};
  if (metrics.filters) {
    state.filters = toFiltersProto(metrics.filters);
  }
  if (metrics.selectedTimeRange) {
    state.timeRange = toTimeRangeProto(metrics.selectedTimeRange);
    if (metrics.selectedTimeRange.interval) {
      state.timeGrain = toTimeGrainProto(metrics.selectedTimeRange.interval);
    }
  }
  if (metrics.selectedComparisonTimeRange) {
    state.compareTimeRange = toTimeRangeProto(
      metrics.selectedComparisonTimeRange
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

function toFiltersProto(filters: V1MetricsViewFilter) {
  return new MetricsViewFilter({
    include: toFilterCondProto(filters.include ?? []),
    exclude: toFilterCondProto(filters.exclude ?? []),
  });
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

function toFilterCondProto(conds: Array<MetricsViewFilterCond>) {
  return conds.map(
    (include) =>
      new MetricsViewFilter_Cond({
        name: include.name,
        like: include.like,
        in: include.in?.map(
          (v) =>
            (v === null
              ? new Value({
                  kind: {
                    case: "nullValue",
                    value: NullValue.NULL_VALUE,
                  },
                })
              : new Value({
                  kind: {
                    case: "stringValue",
                    value: v as string,
                  },
                })) ?? []
        ),
      })
  );
}
