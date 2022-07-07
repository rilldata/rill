import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";
import { store } from "$lib/redux-store/store-root";
import {
  selectMetricsExploreById,
  selectMetricsExplores,
} from "$lib/redux-store/explore/explore-selectors";

export const getMetricsExplores = createReadableFactoryWithSelector(
  store,
  selectMetricsExplores
);

export const getMetricsExploreById = createReadableFactoryWithSelector(
  store,
  selectMetricsExploreById
);
