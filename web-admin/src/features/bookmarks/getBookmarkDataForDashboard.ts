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
 */
export function getBookmarkDataForDashboard(
  dashboard: MetricsExplorerEntity,
  filtersOnly: boolean,
  absoluteTimeRange: boolean,
  timeControlState: TimeControlState,
): string {
  if (absoluteTimeRange) {
    dashboard = {
      ...dashboard,
    };

    dashboard.selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      interval: timeControlState.selectedTimeRange.interval,
      start: timeControlState.selectedTimeRange.start,
      end: timeControlState.selectedTimeRange.end,
    };

    if (
      timeControlState.selectedComparisonTimeRange?.start &&
      timeControlState.selectedComparisonTimeRange?.end
    ) {
      dashboard.selectedComparisonTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: timeControlState.selectedComparisonTimeRange.interval,
        start: timeControlState.selectedComparisonTimeRange.start,
        end: timeControlState.selectedComparisonTimeRange.end,
      };
    }
  }

  if (filtersOnly) {
    return getProtoFromDashboardState({
      whereFilter: dashboard.whereFilter,
      dimensionThresholdFilters: dashboard.dimensionThresholdFilters,
      selectedTimeRange: timeControlState.selectedTimeRange,
    } as MetricsExplorerEntity);
  }

  return getProtoFromDashboardState(dashboard);
}
