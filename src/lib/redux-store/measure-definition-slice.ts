import * as reduxToolkit from "@reduxjs/toolkit";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { generateApis } from "$lib/redux-store/slice-utils";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

const { createSlice, createEntityAdapter } = reduxToolkit;

const measureDefAdapter = createEntityAdapter<MeasureDefinitionEntity>({
  sortComparer: (a, b) => a.creationTime - b.creationTime,
});

export const measureDefSlice = createSlice({
  name: "measureDefinition",
  initialState: measureDefAdapter.getInitialState(),
  reducers: {
    addManyMeasures: {
      reducer: measureDefAdapter.addMany,
      prepare: (measures: Array<MeasureDefinitionEntity>) => ({
        payload: measures,
      }),
    },

    addOneMeasure: {
      reducer: measureDefAdapter.addOne,
      prepare: (measure: MeasureDefinitionEntity) => ({
        payload: measure,
      }),
    },

    updateMeasure: {
      reducer: measureDefAdapter.updateOne,
      prepare: (id: string, measure: Partial<MeasureDefinitionEntity>) => ({
        payload: { id, changes: measure },
      }),
    },

    removeMeasure: {
      reducer: measureDefAdapter.removeOne,
      prepare: (id: string) => ({ payload: id }),
    },
  },
});

export const { addManyMeasures, addOneMeasure, updateMeasure, removeMeasure } =
  measureDefSlice.actions;

export const measureDefSliceReducer = measureDefSlice.reducer;

export const {
  fetchManyApi: fetchManyMeasuresApi,
  createApi: createMeasuresApi,
  updateApi: updateMeasuresApi,
  deleteApi: deleteMeasuresApi,
} = generateApis<
  EntityType.MeasureDefinition,
  { metricsDefId: string },
  { metricsDefId: string }
>(
  EntityType.MeasureDefinition,
  addManyMeasures,
  addOneMeasure,
  updateMeasure,
  removeMeasure,
  "measures"
);

export const manyMeasuresSelector = (metricsDefId) => {
  return (state: RillReduxState) =>
    state.measureDefinition.ids
      .filter(
        (id) =>
          state.measureDefinition.entities[id].metricsDefId === metricsDefId
      )
      .map((id) => state.measureDefinition.entities[id]);
};
export const singleMeasureSelector = (measureId: number | string) => {
  return (state: RillReduxState) => state.measureDefinition.entities[measureId];
};
