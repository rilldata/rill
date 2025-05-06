import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";

/**
 * Returns the bookmark data to be stored.
 * Converts the dashboard to base64 protobuf string using {@link getProtoFromDashboardState}
 *
 * @param dashboard
 * @param exploreSpec
 * @param filtersOnly Only dimension/measure filters and the selected time range is stored.
 * @param absoluteTimeRange Time ranges is treated as absolute.
 * @param timeControlState Time control state to derive the time range from.
 */
export function getBookmarkDataForDashboard(
  dashboard: ExploreState,
  exploreSpec: V1ExploreSpec,
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
    return getProtoFromDashboardState(
      {
        whereFilter: newDashboard.whereFilter,
        dimensionThresholdFilters: newDashboard.dimensionThresholdFilters,
        selectedTimeRange: newDashboard.selectedTimeRange,
      } as ExploreState,
      exploreSpec,
    );
  } else {
    return getProtoFromDashboardState(newDashboard, exploreSpec);
  }
}
