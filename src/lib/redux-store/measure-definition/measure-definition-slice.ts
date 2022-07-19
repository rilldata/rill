import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { PayloadAction } from "@reduxjs/toolkit";

const measureDefAdapter = createEntityAdapter<MeasureDefinitionEntity>();

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

    clearMeasuresForMetricsDefId: {
      reducer: (state, action: PayloadAction<string>) => {
        measureDefAdapter.removeMany(
          state,
          state.ids.filter(
            (id) => state.entities[id].metricsDefId === action.payload
          )
        );
      },
      prepare: (id: string) => ({ payload: id }),
    },
  },
});

export const {
  addManyMeasures,
  addOneMeasure,
  updateMeasure,
  removeMeasure,
  clearMeasuresForMetricsDefId,
} = measureDefSlice.actions;

export const measureDefSliceReducer = measureDefSlice.reducer;
