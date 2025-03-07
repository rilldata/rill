import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const toggleLeaderboardMeasureNames = (
  { dashboard }: DashboardMutables,
  allMeasureNames: string[],
  name?: string,
) => {
  // If name is provided, toggle individual measure
  if (name) {
    const currentSelection = dashboard.leaderboardMeasureNames;
    // Prevent deselecting if it's the last selected measure
    if (currentSelection.length === 1 && currentSelection[0] === name) {
      return;
    }
    dashboard.leaderboardMeasureNames = currentSelection.includes(name)
      ? currentSelection.filter((n) => n !== name)
      : [...currentSelection, name];
  } else {
    // Toggle all measures
    const allSelected =
      allMeasureNames.length === dashboard.leaderboardMeasureNames.length;
    dashboard.leaderboardMeasureNames = allSelected
      ? [dashboard.leaderboardMeasureNames[0]] // Keep first selected when deselecting all
      : allMeasureNames;
  }

  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};

export const setLeaderboardMeasureCount = (
  { dashboard }: DashboardMutables,
  count: number,
) => {
  dashboard.leaderboardMeasureCount = count;
};
