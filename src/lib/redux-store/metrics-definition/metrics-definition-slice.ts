import { createSlice, createEntityAdapter } from "@reduxjs/toolkit";
import type { PayloadAction } from "@reduxjs/toolkit";

import { guidGenerator } from "../../util/guid";

import type { MetricsDefinition } from "$common/state-slice-types/metrics-defintion-types";

// const initialState: MetricsDefinitionsSlice = {
//   defs: {},
//   defsCounter: 0,
// };

type updateDefLabelPayload = {
  id: string;
  label: string;
};

type setDefModelPayload = {
  id: string;
  sourceModelId: string;
};

const metricsDefAdapter = createEntityAdapter<MetricsDefinition>({
  sortComparer: (a, b) => a.creationTime - b.creationTime,
});

export const metricsDefSlice = createSlice({
  name: "metricsDefinitions",
  initialState: metricsDefAdapter.getInitialState({ defsCount: 1 }),
  reducers: {
    addEmptyMetricsDef: (state) => {
      metricsDefAdapter.addOne(state, {
        id: guidGenerator(),
        metricDefLabel: `metrics definition ${state.defsCount}`,
        sourceModelId: undefined,
        timeDimension: undefined,
        measures: [],
        dimensions: [],
        creationTime: Date.now(),
      });
      state.defsCount++;
    },
    deleteMetricsDef: metricsDefAdapter.removeOne,
    // updateLabel:metricsDefAdapter.
    // (state) => {
    // const newId = guidGenerator();
    // state.defs[newId] = {
    //   metricDefinitionId: newId,
    //   metricDefLabel: `metric definition ${state.defsCounter}`,
    //   sourceModelId: undefined,
    //   timeDimension: undefined,
    //   measures: [],
    //   dimensions: [],
    // };
    // state.defsCounter++;
    // },

    // updateLabel: {
    //   reducer: (state, action: PayloadAction<updateDefLabelPayload>) => {
    //     state.defs[action.payload.id].metricDefLabel = action.payload.label;
    //   },
    //   prepare: (id: string, label: string) => ({ payload: { id, label } }),
    // },

    // setMetricsDefModel: {
    //   // QUESTION: will the client already kno
    //   reducer: (state, action: PayloadAction<setDefModelPayload>) => {
    //     state.defs[action.payload.id].sourceModelId =
    //       action.payload.sourceModelId;
    //   },
    //   prepare: (id, sourceModelId) => ({ payload: { id, sourceModelId } }),
    // },
  },
});

// Action creators are generated for each case reducer function
export const { addEmptyMetricsDef, deleteMetricsDef } = metricsDefSlice.actions;

export default metricsDefSlice.reducer;
