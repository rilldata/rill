import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import {
  createEntityAdapter,
  createSlice,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { PayloadAction } from "@reduxjs/toolkit";

const metricsDefAdapter = createEntityAdapter<MetricsDefinitionEntity>({
  sortComparer: (a, b) => a.creationTime - b.creationTime,
});

export const metricsDefSlice = createSlice({
  name: "metricsDefinitions",
  initialState: metricsDefAdapter.getInitialState(),
  reducers: {
    addManyMetricsDefs: {
      reducer: metricsDefAdapter.addMany,
      prepare: (metricsDefs: Array<MetricsDefinitionEntity>) => ({
        payload: metricsDefs,
      }),
    },

    addOneMetricsDef: {
      reducer: metricsDefAdapter.addOne,
      prepare: (metricsDef: MetricsDefinitionEntity) => ({
        payload: metricsDef,
      }),
    },

    updateMetricsDef: {
      reducer: metricsDefAdapter.updateOne,
      prepare: (id: string, metricsDef: Partial<MetricsDefinitionEntity>) => ({
        payload: { id, changes: metricsDef },
      }),
    },

    removeMetricsDef: {
      reducer: metricsDefAdapter.removeOne,
      prepare: (id: string) => ({ payload: id }),
    },

    toggleMetricsDefSummaryInNav: {
      reducer: (state, action: PayloadAction<string>) => {
        state.entities[action.payload].summaryExpandedInNav =
          !state.entities[action.payload].summaryExpandedInNav;
      },
      prepare: (id: string) => ({ payload: id }),
    },
  },
});

export const {
  addManyMetricsDefs,
  addOneMetricsDef,
  updateMetricsDef,
  removeMetricsDef,
  toggleMetricsDefSummaryInNav,
} = metricsDefSlice.actions;
export const MetricsDefSliceActions = metricsDefSlice.actions;
export type MetricsDefSliceActionTypes = typeof MetricsDefSliceActions;

export const metricsDefinitionReducer = metricsDefSlice.reducer;
