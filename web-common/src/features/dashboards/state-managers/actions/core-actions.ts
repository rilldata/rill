import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardMeasureName = name;

  // reset column widths when changing the leaderboard measure
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};

export const setLeaderboardMeasureCount = (
  { dashboard }: DashboardMutables,
  count: number,
) => {
  dashboard.leaderboardMeasureCount = count;
};
