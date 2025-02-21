import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

// TODO: to be removed
export const setLeaderboardMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardMeasureName = name;

  // reset column widths when changing the leaderboard measure
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};

export const setSelectedMeasureNames = (
  { dashboard }: DashboardMutables,
  names: string[],
) => {
  dashboard.selectedMeasureNames = names;

  // reset column widths when changing the leaderboard measure
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};
