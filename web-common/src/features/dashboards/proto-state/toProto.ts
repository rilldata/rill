import {
  NullValue,
  PartialMessage,
  Timestamp,
  Value,
} from "@bufbuild/protobuf";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import { TimeGrain } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import {
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "../../../../../proto/gen/rill/runtime/v1/api_pb";
import { TimeGrain as TimeGrainProto } from "../../../../../proto/gen/rill/runtime/v1/catalog_pb";
import { DashboardState } from "../../../../../proto/gen/rill/ui/v1/dashboard_pb";

export function toProto(metrics: MetricsExplorerEntity) {
  const data: PartialMessage<DashboardState> = {};
  if (metrics.filters) {
    data.filters = toFiltersProto(metrics.filters) as any;
  }
  if (metrics.selectedTimeRange?.start) {
    data.timeStart = toTimeProto(metrics.selectedTimeRange?.start);
  }
  if (metrics.selectedTimeRange?.end) {
    data.timeEnd = toTimeProto(metrics.selectedTimeRange?.end);
  }
  if (metrics.selectedTimeRange?.interval) {
    data.timeGranularity = toTimeGrainProto(
      metrics.selectedTimeRange?.interval
    );
  }
  return new DashboardState(data);
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

function toTimeProto(time: string) {
  return new Timestamp({
    seconds: BigInt(new Date(time).getTime()),
  });
}

function toTimeGrainProto(timeGrain: TimeGrain) {
  switch (timeGrain) {
    case TimeGrain.OneMinute:
      return TimeGrainProto.MINUTE;
    case TimeGrain.OneHour:
      return TimeGrainProto.HOUR;
    case TimeGrain.OneDay:
      return TimeGrainProto.DAY;
    case TimeGrain.OneWeek:
      return TimeGrainProto.WEEK;
    case TimeGrain.OneMonth:
      return TimeGrainProto.MONTH;
    case TimeGrain.OneYear:
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
