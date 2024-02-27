import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";

export function getBookmarkForDashboard(
  dashboard: MetricsExplorerEntity,
  filtersOnly: boolean,
  absoluteTimeRange: boolean,
): string {
  if (filtersOnly) {
    return getProtoFromDashboardState({
      whereFilter: dashboard.whereFilter,
      dimensionThresholdFilters: dashboard.dimensionThresholdFilters,
    } as MetricsExplorerEntity);
  }

  if (absoluteTimeRange) {
    dashboard = {
      ...dashboard,
    };
    if (
      dashboard.selectedTimeRange?.start &&
      dashboard.selectedTimeRange?.end
    ) {
      dashboard.selectedTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: dashboard.selectedTimeRange.interval,
        start: dashboard.selectedTimeRange.start,
        end: dashboard.selectedTimeRange.end,
      };
    }
    if (
      dashboard.selectedComparisonTimeRange?.start &&
      dashboard.selectedComparisonTimeRange?.end
    ) {
      dashboard.selectedComparisonTimeRange = {
        name: TimeRangePreset.CUSTOM,
        interval: dashboard.selectedComparisonTimeRange.interval,
        start: dashboard.selectedComparisonTimeRange.start,
        end: dashboard.selectedComparisonTimeRange.end,
      };
    }
  }

  return getProtoFromDashboardState(dashboard);
}
