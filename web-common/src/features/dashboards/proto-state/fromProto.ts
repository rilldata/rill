import type { Timestamp } from "@bufbuild/protobuf";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/catalog_pb";
import type { MetricsViewFilter_Cond } from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  DashboardState,
  DashboardTimeRange,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export function getDashboardStateFromUrl(
  url: URL
): Partial<MetricsExplorerEntity> {
  const state = url.searchParams.get("state");
  if (!state) return undefined;
  return getDashboardStateFromProto(base64ToProto(decodeURIComponent(state)));
}

export function getDashboardStateFromProto(
  binary: Uint8Array
): Partial<MetricsExplorerEntity> {
  const dashboard = DashboardState.fromBinary(binary);
  const entity: Partial<MetricsExplorerEntity> = {
    filters: {
      include: [],
      exclude: [],
    },
  };

  if (dashboard.filters) {
    entity.filters.include = fromFiltersProto(dashboard.filters.include);
    entity.filters.exclude = fromFiltersProto(dashboard.filters.exclude);
  }

  entity.selectedTimeRange = dashboard.timeRange
    ? fromTimeRangeProto(dashboard.timeRange)
    : undefined;
  if (dashboard.timeGrain && dashboard.timeRange) {
    entity.selectedTimeRange.interval = fromTimeGrainProto(dashboard.timeGrain);
  }

  if (dashboard.leaderboardMeasure) {
    entity.leaderboardMeasureName = dashboard.leaderboardMeasure;
  }
  if (dashboard.selectedDimension) {
    entity.selectedDimensionName = dashboard.selectedDimension;
  }

  return entity;
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
  const selectedTimeRange: DashboardTimeControls = {
    name: timeRange.name,
  } as DashboardTimeControls;

  selectedTimeRange.name = timeRange.name;
  if (timeRange.timeStart) {
    selectedTimeRange.start = fromTimeProto(timeRange.timeStart);
  }
  if (timeRange.timeEnd) {
    selectedTimeRange.end = fromTimeProto(timeRange.timeEnd);
  }

  return selectedTimeRange;
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
    case TimeGrain.YEAR:
      return V1TimeGrain.TIME_GRAIN_YEAR;
  }
}
