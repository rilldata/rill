import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardMeasureCount = (
  { dashboard }: DashboardMutables,
  count: number,
) => {
  dashboard.leaderboardMeasureCount = count;
};

export const setLeaderboardSortByMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardSortByMeasureName = name;

  // reset column widths when changing the leaderboard measure
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};
