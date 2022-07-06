import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";

export const {
  manySelector: selectMetricsExplores,
  singleSelector: selectMetricsExploreById,
} = generateEntitySelectors<MetricsExploreEntity>("metricsLeaderboard");
