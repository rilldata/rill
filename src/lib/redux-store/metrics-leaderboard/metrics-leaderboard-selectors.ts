import { generateBasicSelectors } from "$lib/redux-store/utils/selector-utils";

export const {
  manySelector: manyMetricsLeaderboardSelector,
  singleSelector: singleMetricsLeaderboardSelector,
} = generateBasicSelectors("metricsLeaderboard");
