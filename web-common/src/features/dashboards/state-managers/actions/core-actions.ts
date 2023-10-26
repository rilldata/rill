import type { MetricsExplorerEntity } from "../../stores/metrics-explorer-entity";

export const setLeaderboardMeasureName = (
  dash: MetricsExplorerEntity,
  name: string
) => {
  dash.leaderboardMeasureName = name;
};
