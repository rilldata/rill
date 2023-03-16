import {
  NullValue,
  PartialMessage,
  Timestamp,
  Value,
} from "@bufbuild/protobuf";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { TimeRangeName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
import type { TimeSeriesTimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
import {
  TimeGrain,
  TimeGrain as TimeGrainProto,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/catalog_pb";
import {
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  DashboardState,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export function toProto(metrics: MetricsExplorerEntity): string {
  const state: PartialMessage<DashboardState> = {};
  if (metrics.filters) {
    state.filters = toFiltersProto(metrics.filters) as any;
  }
  if (metrics.selectedTimeRange) {
    state.timeRange = toTimeRangeProto(metrics.selectedTimeRange);
    if (metrics.selectedTimeRange.interval) {
      state.timeGrain = toTimeGrainProto(metrics.selectedTimeRange.interval);
    }
  }
  if (metrics.leaderboardMeasureName) {
    state.leaderboardMeasure = metrics.leaderboardMeasureName;
  }
  if (metrics.selectedDimensionName) {
    state.selectedDimension = metrics.selectedDimensionName;
  }
  const message = new DashboardState(state);
  return protoToBase64(message.toBinary());
}

export function protoToBase64(proto: Uint8Array) {
  return btoa(String.fromCharCode.apply(null, proto));
}

function toFiltersProto(filters: V1MetricsViewFilter) {
  return new MetricsViewFilter({
    include: toFilterCondProto(filters.include) as any,
    exclude: toFilterCondProto(filters.exclude) as any,
  });
}

function toTimeRangeProto(range: TimeSeriesTimeRange) {
  const timeRangeArgs: PartialMessage<DashboardTimeRange> = {
    name: range.name,
  };
  if (range.name === TimeRangeName.Custom) {
    timeRangeArgs.timeStart = toTimeProto(range.start);
    timeRangeArgs.timeEnd = toTimeProto(range.end);
  }
  return new DashboardTimeRange(timeRangeArgs);
}

function toTimeProto(time: string) {
  return new Timestamp({
    seconds: BigInt(new Date(time).getTime()),
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
        in: include.in.map(
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
                })) as any
        ),
      })
  );
}
