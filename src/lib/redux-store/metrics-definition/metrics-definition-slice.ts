import { createSlice, PayloadAction } from "@reduxjs/toolkit";

import { guidGenerator } from "../../util/guid";

import type { MetricsDefinitionsSlice } from "$common/state-slice-types/metrics-defintion-types";

const initialState: MetricsDefinitionsSlice = {
  defs: {},
  defsCounter: 0,
};

type updateDefLabelPayload = {
  id: string;
  label: string;
};

type setDefModelPayload = {
  id: string;
  sourceModelId: string;
};

export const metricsDefSlice = createSlice({
  name: "metricsDefinitions",
  initialState,
  reducers: {
    addEmptyMetricsDef: (state) => {
      const newId = guidGenerator();
      state.defs[newId] = {
        metricDefinitionId: newId,
        metricDefLabel: `metric definition ${state.defsCounter}`,
        sourceModelId: undefined,
        timeDimension: undefined,
        measures: [],
        dimensions: [],
      };
      state.defsCounter++;
    },

    updateMetricsDefLabel: {
      reducer: (state, action: PayloadAction<updateDefLabelPayload>) => {
        state.defs[action.payload.id].metricDefLabel = action.payload.label;
      },
      prepare: (id: string, label: string) => ({ payload: { id, label } }),
    },

    setMetricsDefModel: {
      // QUESTION: will the client already kno
      reducer: (state, action: PayloadAction<setDefModelPayload>) => {
        state.defs[action.payload.id].sourceModelId =
          action.payload.sourceModelId;
      },
      prepare: (id, sourceModelId) => ({ payload: { id, sourceModelId } }),
    },
  },
});

// Action creators are generated for each case reducer function
export const { addEmptyMetricsDef, setMetricsDefModel } =
  metricsDefSlice.actions;
setMetricsDefModel(24, 46);

export default metricsDefSlice.reducer;
