import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  // dashboard.leaderboardMeasureName = name;
  dashboard.leaderboardMeasureNames = [name];
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};

export const setLeaderboardMeasureNames = (
  { dashboard }: DashboardMutables,
  names: string[],
) => {
  dashboard.leaderboardMeasureNames = names;
  // First measure is used as the leaderboard measure for sorting
  // dashboard.leaderboardMeasureName = names[0];
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};
