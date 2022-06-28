import { store } from "../store-root";

import {
  selectMetricsDefinitionById,
  selectAllMetricsDefinitions,
} from "$lib/redux-store/metrics-definition/metrics-definitioin-selectors";
import { createReadableFactoryWithSelector } from "../svelte-readables-wrapper";

export const getMetricsDefReadableById = createReadableFactoryWithSelector(
  store,
  selectMetricsDefinitionById
);

export const getAllMetricsDefinitionsReadable =
  createReadableFactoryWithSelector(store, selectAllMetricsDefinitions);
