import type { Timestamp } from "@bufbuild/protobuf";
import type { TimeRangeName } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
import type { TimeSeriesTimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-control-types";
import type { MetricsViewFilter_Cond } from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  DashboardState,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { V1MetricsViewFilter } from "@rilldata/web-common/runtime-client";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/catalog_pb";

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

  const timeRange = fromTimeRangeProto(dashboard.timeRange);

  return [filters, timeRange];
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

function fromTimeRangeProto(timeRange: DashboardTimeRange) {
  const selectedTimeRange: TimeSeriesTimeRange = {};

  selectedTimeRange.interval = fromTimeGrainProto(timeRange.timeGranularity);
  selectedTimeRange.name = timeRange.name as TimeRangeName;
  if (timeRange.timeStart) {
    selectedTimeRange.start = fromTimeProto(timeRange.timeStart);
  }
  if (timeRange.timeEnd) {
    selectedTimeRange.end = fromTimeProto(timeRange.timeEnd);
  }

  return selectedTimeRange;
}

function fromTimeProto(timestamp: Timestamp) {
  return new Date(Number(timestamp.seconds)).toISOString();
}

function fromTimeGrainProto(timeGrain: TimeGrain): V1TimeGrain {
  switch (timeGrain) {
    case TimeGrain.UNSPECIFIED:
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
    case TimeGrain.YEAR:
      return V1TimeGrain.TIME_GRAIN_YEAR;
  }
}
