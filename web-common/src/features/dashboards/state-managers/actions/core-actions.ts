import type { DashboardMutables } from "./types";

export const setLeaderboardMeasureName = (
  { dashboard }: DashboardMutables,
  name: string,
) => {
  dashboard.leaderboardMeasureName = name;
};
