import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";

/**
 * Returns the bookmark data to be stored.
 * Converts the dashboard to base64 protobuf string using {@link getProtoFromDashboardState}
 *
 * @param dashboard
 * @param filtersOnly Only dimension/measure filters and the selected time range is stored.
 * @param absoluteTimeRange Time ranges is treated as absolute.
 * @param timeControlState Time control state to derive the time range from.
 */
export function getBookmarkDataForDashboard(
  dashboard: MetricsExplorerEntity,
  filtersOnly?: boolean,
  absoluteTimeRange?: boolean,
  timeControlState?: TimeControlState,
): string {
  const newDashboard = structuredClone(dashboard);

  if (absoluteTimeRange && timeControlState) {
    if (
      timeControlState.selectedTimeRange?.start &&
      timeControlState.selectedTimeRange?.end
    ) {
      newDashboard.selectedTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: timeControlState.selectedTimeRange.interval,
        start: timeControlState.selectedTimeRange.start,
        end: timeControlState.selectedTimeRange.end,
      };
    }

    if (
      timeControlState.selectedComparisonTimeRange?.start &&
      timeControlState.selectedComparisonTimeRange?.end
    ) {
      newDashboard.selectedComparisonTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: timeControlState.selectedComparisonTimeRange.interval,
        start: timeControlState.selectedComparisonTimeRange.start,
        end: timeControlState.selectedComparisonTimeRange.end,
      };
    }
  }

  if (filtersOnly) {
    return getProtoFromDashboardState({
      whereFilter: newDashboard.whereFilter,
      dimensionThresholdFilters: newDashboard.dimensionThresholdFilters,
      selectedTimeRange: newDashboard.selectedTimeRange,
    } as MetricsExplorerEntity);
  } else {
    return getProtoFromDashboardState(newDashboard);
  }
}
