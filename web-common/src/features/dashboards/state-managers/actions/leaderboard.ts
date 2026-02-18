import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardSortByMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardSortByMeasureName = name;

  // Ensure leaderboardMeasureNames stays in sync with sort measure
  // This prevents blank values when sorting by a measure that's not displayed
  if (dashboard.leaderboardMeasureNames?.length > 0) {
    const isMultiSelect = dashboard.leaderboardMeasureNames.length > 1;

    if (isMultiSelect) {
      // In multi-select mode: add sort measure if not already in the list
      if (!dashboard.leaderboardMeasureNames.includes(name)) {
        dashboard.leaderboardMeasureNames = [
          name,
          ...dashboard.leaderboardMeasureNames,
        ];
      }
    } else {
      // In single-select mode: replace with just the sort measure
      dashboard.leaderboardMeasureNames = [name];
    }
  }

  // reset column widths when changing the leaderboard measure
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};

export const setLeaderboardMeasureNames = (
  { dashboard }: DashboardMutables,
  names: string[],
) => {
  dashboard.leaderboardMeasureNames = names;
};

export const toggleLeaderboardShowContextForAllMeasures = ({
  dashboard,
}: DashboardMutables) => {
  dashboard.leaderboardShowContextForAllMeasures =
    !dashboard.leaderboardShowContextForAllMeasures;
};

export const leaderboardActions = {
  setLeaderboardSortByMeasureName,
  setLeaderboardMeasureNames,
  toggleLeaderboardShowContextForAllMeasures,
};
