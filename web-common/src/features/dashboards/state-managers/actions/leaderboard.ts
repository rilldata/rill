import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardSortByMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardSortByMeasureName = name;

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
