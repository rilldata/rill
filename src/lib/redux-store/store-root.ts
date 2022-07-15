import { configureStore } from "$lib/redux-store/redux-toolkit-wrapper";
import { metricsDefinitionReducer } from "./metrics-definition/metrics-definition-slice";
import {
  MetricsExploreEntity,
  metricsLeaderboardReducer,
} from "./explore/explore-slice";
import { readable } from "svelte/store";
import { measureDefSliceReducer } from "$lib/redux-store/measure-definition/measure-definition-slice";
import { dimensionDefSliceReducer } from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import {
  TimeSeriesEntity,
  timeSeriesReducer,
} from "$lib/redux-store/timeseries/timeseries-slice";
import {
  BigNumberEntity,
  bigNumberReducer,
} from "$lib/redux-store/big-number/big-number-slice";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { applicationReducer } from "$lib/redux-store/application/application-slice";

export const store = configureStore({
  reducer: {
    application: applicationReducer,
    metricsDefinition: metricsDefinitionReducer,
    measureDefinition: measureDefSliceReducer,
    dimensionDefinition: dimensionDefSliceReducer,
    metricsLeaderboard: metricsLeaderboardReducer,
    timeSeries: timeSeriesReducer,
    bigNumber: bigNumberReducer,
  },
});

const state = store.getState();
export type RillReduxState = typeof state;
export type RillReduxStore = typeof store;
export type RillReduxEntities =
  | MetricsDefinitionEntity
  | MeasureDefinitionEntity
  | DimensionDefinitionEntity
  | MetricsExploreEntity
  | TimeSeriesEntity
  | BigNumberEntity;
export type RillReduxEntityKeys =
  | "metricsDefinition"
  | "measureDefinition"
  | "dimensionDefinition"
  | "metricsLeaderboard"
  | "timeSeries"
  | "bigNumber";

export const reduxReadable = readable(store.getState(), (set) => {
  return store.subscribe(() => {
    set(store.getState());
  });
});
