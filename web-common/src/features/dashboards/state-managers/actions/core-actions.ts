import type { DashboardMutatorFnGeneralArgs } from "./types";

export const setLeaderboardMeasureName = (
  { dashboard }: DashboardMutatorFnGeneralArgs,
  name: string
) => {
  dashboard.leaderboardMeasureName = name;
};
