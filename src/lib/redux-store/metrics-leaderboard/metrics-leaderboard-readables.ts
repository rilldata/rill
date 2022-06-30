import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";
import { store } from "$lib/redux-store/store-root";
import {
  selectMetricsLeaderboardById,
  selectMetricsLeaderboards,
} from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-selectors";

export const getMetricsLeaderboards = createReadableFactoryWithSelector(
  store,
  selectMetricsLeaderboards
);

export const getMetricsLeaderboardById = createReadableFactoryWithSelector(
  store,
  selectMetricsLeaderboardById
);
