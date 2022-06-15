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
      state.entities[id].metricDefLabel = label;
    },
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
