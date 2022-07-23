import {
  selectMetricsExplorerById,
  selectMetricsExplorers,
} from "$lib/redux-store/explore/explore-selectors";
import { store } from "$lib/redux-store/store-root";
import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";

export const getMetricsExplorers = createReadableFactoryWithSelector(
  store,
  selectMetricsExplorers
);

export const getMetricsExplorerById = createReadableFactoryWithSelector(
  store,
  selectMetricsExplorerById
);
