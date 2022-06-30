import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { MetricsLeaderboardEntity } from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";

export const {
  manySelector: selectMetricsLeaderboards,
  singleSelector: selectMetricsLeaderboardById,
} = generateEntitySelectors<MetricsLeaderboardEntity>("metricsLeaderboard");
