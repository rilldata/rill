import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";
import {
  selectDimensionById,
  selectDimensionsByMetricsId,
} from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { store } from "$lib/redux-store/store-root";

export const getDimensionById = createReadableFactoryWithSelector(
  store,
  selectDimensionById
);

export const getDimensionsByMetricsId = createReadableFactoryWithSelector(
  store,
  selectDimensionsByMetricsId
);
