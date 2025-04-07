import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

// @deprecated
// NOTE: this is a temporary action to set the leaderboard measure count.
// It will be removed in the multiple measures second pass anyway.
export const setLeaderboardMeasureCount = (
  { dashboard }: DashboardMutables,
  count: number,
) => {
  dashboard.leaderboardMeasureCount = count;

  // // If the current leaderboard measure is not in the first N visible measures,
  // // set it to the first visible measure
  // const visibleMeasures = dashboard.visibleMeasures.slice(0, count);
  // if (!visibleMeasures.includes(dashboard.leaderboardSortByMeasureName)) {
  //   dashboard.leaderboardSortByMeasureName = visibleMeasures[0];
  // }
};

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
  setLeaderboardMeasureCount,
  setLeaderboardSortByMeasureName,
  setLeaderboardMeasureNames,
  setLeaderboardShowAllMeasures,
};
