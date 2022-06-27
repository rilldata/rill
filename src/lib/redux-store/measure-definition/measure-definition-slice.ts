import * as reduxToolkit from "@reduxjs/toolkit";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

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
