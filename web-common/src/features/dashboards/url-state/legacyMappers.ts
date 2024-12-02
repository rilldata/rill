import { ExploreStateDefaultChartType } from "@rilldata/web-common/features/dashboards/url-state/defaults";
import { reverseMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1ExploreOverviewSortType } from "@rilldata/web-common/runtime-client";

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

// TODO: use V1ExploreOverviewSortType across the app instead
export const FromLegacySortTypeMap: Record<
  DashboardState_LeaderboardSortType,
  V1ExploreOverviewSortType
> = {
  [DashboardState_LeaderboardSortType.UNSPECIFIED]:
    V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_UNSPECIFIED,
  [DashboardState_LeaderboardSortType.VALUE]:
    V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_VALUE,
  [DashboardState_LeaderboardSortType.PERCENT]:
    V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_PERCENT,
  [DashboardState_LeaderboardSortType.DELTA_ABSOLUTE]:
    V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_DELTA_ABSOLUTE,
  [DashboardState_LeaderboardSortType.DELTA_PERCENT]:
    V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_DELTA_PERCENT,
  [DashboardState_LeaderboardSortType.DIMENSION]:
    V1ExploreOverviewSortType.EXPLORE_OVERVIEW_SORT_TYPE_DIMENSION,
};
export const ToLegacySortTypeMap = reverseMap(FromLegacySortTypeMap);
