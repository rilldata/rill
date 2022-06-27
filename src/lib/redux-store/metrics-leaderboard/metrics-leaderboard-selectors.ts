import { generateBasicSelectors } from "$lib/redux-store/slice-utils";

export const {
  manySelector: manyMetricsLeaderboardSelector,
  singleSelector: singleMetricsLeaderboardSelector,
} = generateBasicSelectors("metricsLeaderboard");
