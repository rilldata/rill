import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import * as reduxToolkit from "@reduxjs/toolkit";

const { createSlice, createEntityAdapter } = reduxToolkit;

const metricsDefAdapter = createEntityAdapter<MetricsDefinitionEntity>({
  sortComparer: (a, b) => a.creationTime - b.creationTime,
});

export const metricsDefSlice = createSlice({
  name: "metricsDefinitions",
  initialState: metricsDefAdapter.getInitialState(),
  reducers: {},
});

export const {} = metricsDefSlice.actions;
export const MetricsDefSliceActions = metricsDefSlice.actions;
export type MetricsDefSliceActionTypes = typeof MetricsDefSliceActions;

export const metricsDefinitionReducer = metricsDefSlice.reducer;
