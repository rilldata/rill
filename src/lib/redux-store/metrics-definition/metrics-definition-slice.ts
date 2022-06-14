import {
  createSlice,
  createEntityAdapter,
  PayloadAction,
} from "@reduxjs/toolkit";

import { guidGenerator } from "../../util/guid";

import type {
  MetricsDefinition,
  UUID,
} from "$common/state-slice-types/metrics-defintion-types";

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
        summaryExpandedInNav: false,
      });
      state.defsCount++;
    },
    deleteMetricsDef: metricsDefAdapter.removeOne,
    toggleSummaryInNav: (state, action: PayloadAction<UUID>) => {
      const expanded = state.entities[action.payload].summaryExpandedInNav;
      state.entities[action.payload].summaryExpandedInNav = !expanded;
    },
    updateMetricDefLabel: (
      state,
      action: PayloadAction<{ id: UUID; label: string }>
    ) => {
      const { id, label } = action.payload;
      // const id = action.payload.id
      // const label = action.payload.label
      //       const expanded = state.entities[action.payload].summaryExpandedInNav;
      // state.entities[action.payload].summaryExpandedInNav = !expanded;
      // metricsDefAdapter.updateOne(state, action.payload);
      state.entities[id].metricDefLabel = label;
    },
    // toggleSummaryInNav: {
    //   reducer:(state, action:PayloadAction<UUID>) =>{
    //     state.entities[action.payload.id]
    //   },
    //   prepare: (id:UUID) =>  ({ payload:  {id}  }),
    // },
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
export const {
  addEmptyMetricsDef,
  deleteMetricsDef,
  toggleSummaryInNav,
  updateMetricDefLabel,
} = metricsDefSlice.actions;

export default metricsDefSlice.reducer;
