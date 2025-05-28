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

const ensureTimeComparisonEnabled = (
  dashboard: DashboardMutables["dashboard"],
) => {
  if (
    dashboard.leaderboardShowContextForAllMeasures &&
    !dashboard.showTimeComparison
  ) {
    dashboard.showTimeComparison = true;
  }
};

export const toggleLeaderboardShowContextForAllMeasures = ({
  dashboard,
}: DashboardMutables) => {
  dashboard.leaderboardShowContextForAllMeasures =
    !dashboard.leaderboardShowContextForAllMeasures;
  ensureTimeComparisonEnabled(dashboard);
};

export const leaderboardActions = {
  setLeaderboardSortByMeasureName,
  setLeaderboardMeasureNames,
  toggleLeaderboardShowContextForAllMeasures,
};
