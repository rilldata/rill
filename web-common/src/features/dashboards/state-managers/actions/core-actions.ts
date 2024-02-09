import { getLocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import { resetAllContextColumnWidths } from "./context-columns";
import type { DashboardMutables } from "./types";

export const setLeaderboardMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardMeasureName = name;
  getLocalUserPreferences()?.updateLeaderboardMeasureName(name);

  // reset column widths when changing the leaderboard measure
  resetAllContextColumnWidths(dashboard.contextColumnWidths);
};
