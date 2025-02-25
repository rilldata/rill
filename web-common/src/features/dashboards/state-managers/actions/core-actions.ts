import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardMeasureNames = (
  { dashboard }: DashboardMutables,
  names: string[],
) => {
  dashboard.leaderboardMeasureNames = names;
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};

export const toggleLeaderboardMeasureNames = (
  { dashboard }: DashboardMutables,
  allMeasureNames: string[],
) => {
  const allSelected =
    allMeasureNames.length === dashboard.leaderboardMeasureNames.length;
  dashboard.leaderboardMeasureNames = allSelected
    ? [allMeasureNames[0]]
    : allMeasureNames;
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};
