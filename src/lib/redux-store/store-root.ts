import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { applicationReducer } from "$lib/redux-store/application/application-slice";
import { dimensionDefSliceReducer } from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import { measureDefSliceReducer } from "$lib/redux-store/measure-definition/measure-definition-slice";
import { configureStore } from "$lib/redux-store/redux-toolkit-wrapper";
import { readable } from "svelte/store";
import { metricsDefinitionReducer } from "./metrics-definition/metrics-definition-slice";

export const store = configureStore({
  reducer: {
    application: applicationReducer,
    metricsDefinition: metricsDefinitionReducer,
    measureDefinition: measureDefSliceReducer,
    dimensionDefinition: dimensionDefSliceReducer,
  },
});

const state = store.getState();
export type RillReduxState = typeof state;
export type RillReduxStore = typeof store;
export type RillReduxEntities =
  | MetricsDefinitionEntity
  | MeasureDefinitionEntity
  | DimensionDefinitionEntity;
export type RillReduxEntityKeys =
  | "metricsDefinition"
  | "measureDefinition"
  | "dimensionDefinition";

export const reduxReadable = readable(store.getState(), (set) => {
  return store.subscribe(() => {
    set(store.getState());
  });
});
