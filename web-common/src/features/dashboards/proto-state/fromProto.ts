import type { Timestamp } from "@bufbuild/protobuf";
import type { TimeSeriesTimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
import type { MetricsViewFilter_Cond } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
import { DashboardState } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type { V1MetricsViewFilter } from "@rilldata/web-common/runtime-client";
import { TimeGrain as TimeGrainProto } from "@rilldata/web-common/proto/gen/rill/runtime/v1/catalog_pb";

export function fromProto(
  binary: Uint8Array
): [filters: V1MetricsViewFilter, selectedTimeRange: TimeSeriesTimeRange] {
  const dashboard = DashboardState.fromBinary(binary);

  const filters: V1MetricsViewFilter = {
    include: [],
    exclude: [],
  };
  if (dashboard.filters) {
    filters.include = fromFiltersProto(dashboard.filters.include);
    filters.exclude = fromFiltersProto(dashboard.filters.exclude);
  }

  const selectedTimeRange: TimeSeriesTimeRange = {};
  if (dashboard.timeStart) {
    selectedTimeRange.start = fromTimeProto(dashboard.timeStart);
  }
  if (dashboard.timeEnd) {
    selectedTimeRange.end = fromTimeProto(dashboard.timeEnd);
  }
  if (dashboard.timeGranularity) {
    selectedTimeRange.interval = fromTimeGrainProto(dashboard.timeGranularity);
  }

  return [filters, selectedTimeRange];
}

export function base64ToProto(message: string) {
  return new Uint8Array(
    atob(message)
      .split("")
      .map(function (c) {
        return c.charCodeAt(0);
      })
  );
}

function fromFiltersProto(conditions: Array<MetricsViewFilter_Cond>) {
  return conditions.map((condition) => ({
    name: condition.name,
    like: condition.like,
    in: condition.in.map((v) =>
      v.kind.case === "nullValue" ? null : v.kind.value
    ),
  }));
}

function fromTimeProto(timestamp: Timestamp) {
  return new Date(Number(timestamp.seconds)).toISOString();
}

function fromTimeGrainProto(timeGrain: TimeGrainProto) {
  switch (timeGrain) {
    case TimeGrainProto.MINUTE:
      return TimeGrain.OneMinute;
    case TimeGrainProto.HOUR:
      return TimeGrain.OneHour;
    case TimeGrainProto.DAY:
      return TimeGrain.OneDay;
    case TimeGrainProto.WEEK:
      return TimeGrain.OneWeek;
    case TimeGrainProto.MONTH:
      return TimeGrain.OneMonth;
    case TimeGrainProto.YEAR:
      return TimeGrain.OneYear;
  }
}
