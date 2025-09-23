import { ExploreStateDefaultChartType } from "@rilldata/web-common/features/dashboards/url-state/defaults";

import { reverseMap } from "@rilldata/web-common/lib/map-utils.ts";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1ExploreSortType } from "@rilldata/web-common/runtime-client";

const LegacyCharTypeToPresetChartType: Record<string, string> = {
  default: ExploreStateDefaultChartType,
  grouped_bar: "bar",
  stacked_bar: "stacked_bar",
  stacked_area: "stacked_area",
};
export function mapLegacyChartType(chartType: string | undefined) {
  if (!chartType) {
    return ExploreStateDefaultChartType;
  }
  return (
    LegacyCharTypeToPresetChartType[chartType] ?? ExploreStateDefaultChartType
  );
}

// TODO: use V1ExploreSortType across the app instead
export const FromLegacySortTypeMap: Record<
  DashboardState_LeaderboardSortType,
  V1ExploreSortType
> = {
  [DashboardState_LeaderboardSortType.UNSPECIFIED]:
    V1ExploreSortType.EXPLORE_SORT_TYPE_UNSPECIFIED,
  [DashboardState_LeaderboardSortType.VALUE]:
    V1ExploreSortType.EXPLORE_SORT_TYPE_VALUE,
  [DashboardState_LeaderboardSortType.PERCENT]:
    V1ExploreSortType.EXPLORE_SORT_TYPE_PERCENT,
  [DashboardState_LeaderboardSortType.DELTA_ABSOLUTE]:
    V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_ABSOLUTE,
  [DashboardState_LeaderboardSortType.DELTA_PERCENT]:
    V1ExploreSortType.EXPLORE_SORT_TYPE_DELTA_PERCENT,
  [DashboardState_LeaderboardSortType.DIMENSION]:
    V1ExploreSortType.EXPLORE_SORT_TYPE_DIMENSION,
};
export const ToLegacySortTypeMap = reverseMap(FromLegacySortTypeMap);
