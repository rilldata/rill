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

export const setLeaderboardShowAllMeasures = (
  { dashboard }: DashboardMutables,
  showAllMeasures: boolean,
) => {
  dashboard.leaderboardShowAllMeasures = showAllMeasures;
};

export const leaderboardActions = {
  setLeaderboardSortByMeasureName,
  setLeaderboardMeasureNames,
  setLeaderboardShowAllMeasures,
};
