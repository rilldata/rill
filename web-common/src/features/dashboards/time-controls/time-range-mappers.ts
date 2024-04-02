import { PreviousCompleteRangeMap } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type {
  V1MetricsViewSpec,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";

/**
 * Maps selectedTimeRange to V1TimeRange.
 */
export function mapTimeRange(
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec,
) {
  if (!timeControlState.selectedTimeRange?.name) return undefined;

  const timeRange: V1TimeRange = {};
  switch (timeControlState.selectedTimeRange.name) {
    case TimeRangePreset.DEFAULT:
      timeRange.isoDuration = metricsView.defaultTimeRange;
      break;

    case TimeRangePreset.CUSTOM:
      timeRange.start = timeControlState.timeStart;
      timeRange.end = timeControlState.timeEnd;
      break;

    default:
      if (timeControlState.selectedTimeRange.name in PreviousCompleteRangeMap) {
        const prevCompleteTimeRange: V1TimeRange | undefined =
          PreviousCompleteRangeMap[timeControlState.selectedTimeRange.name];
        // Backend doesn't support previous complete ranges since it has offset built in.
        // We add the offset manually as a workaround for now
        timeRange.isoDuration = prevCompleteTimeRange?.isoDuration;
        timeRange.isoOffset = prevCompleteTimeRange?.isoOffset;
        timeRange.roundToGrain = prevCompleteTimeRange?.roundToGrain;
      } else {
        timeRange.isoDuration = timeControlState.selectedTimeRange.name;
      }
      break;
  }

  return timeRange;
}

/**
 * Maps selectedComparisonTimeRange to V1TimeRange if time comparison is enabled.
 */
export function mapComparisonTimeRange(
  dashboardState: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  timeRange: V1TimeRange | undefined,
) {
  if (
    !timeRange ||
    dashboardState.selectedComparisonDimension ||
    !timeControlState.showComparison ||
    !timeControlState.selectedComparisonTimeRange?.name
  ) {
    return undefined;
  }

  const comparisonTimeRange: V1TimeRange = {};
  switch (timeControlState.selectedComparisonTimeRange.name) {
    default:
      comparisonTimeRange.isoOffset =
        timeControlState.selectedComparisonTimeRange.name;
    // eslint-disable-next-line no-fallthrough
    case TimeComparisonOption.CONTIGUOUS:
      comparisonTimeRange.isoDuration = timeRange.isoDuration;
      break;

    case TimeComparisonOption.CUSTOM:
      comparisonTimeRange.start = timeControlState.comparisonTimeStart;
      comparisonTimeRange.end = timeControlState.comparisonTimeEnd;
      break;
  }
  return comparisonTimeRange;
}
