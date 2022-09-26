import { createReadableFactoryWithSelector } from "../svelte-readables-wrapper";
import {
  selectDimensionById,
  selectDimensionsByMetricsId,
} from "./dimension-definition-selectors";
import { store } from "../store-root";

export const getDimensionById = createReadableFactoryWithSelector(
  store,
  selectDimensionById
);

export const getDimensionsByMetricsId = createReadableFactoryWithSelector(
  store,
  selectDimensionsByMetricsId
);
