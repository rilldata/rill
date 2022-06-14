import type {
  MetricsDefinitionEntity,
  UUID,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import type { DimensionDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export type MetricsDefinitionsSlice = {
  defs: { [id: UUID]: MetricsDefinitionEntity };
  selectedDefId?: UUID;
};

const initialState: MetricsDefinitionsSlice = {
  defs: {},
};

type updateDefLabelPayload = {
  id: string;
  label: string;
};

type addNewDimensionPayload = {
  id: string;
  dimension: DimensionDefinition;
};

type setDefModelPayload = {
  id: string;
  sourceModelId: string;
};

export const metricsDefSlice = createSlice({
  name: "metricsDefinitions",
  initialState,
  reducers: {
    addEmptyMetricsDef: (
      state,
      action: PayloadAction<MetricsDefinitionEntity>
    ) => {
      state.defs[action.payload.id] = action.payload;
    },

    updateMetricsDefLabel: {
      reducer: (state, action: PayloadAction<updateDefLabelPayload>) => {
        state.defs[action.payload.id].metricDefLabel = action.payload.label;
      },
      prepare: (id: string, label: string) => ({ payload: { id, label } }),
    },

    setMetricsDefModel: {
      reducer: (state, action: PayloadAction<setDefModelPayload>) => {
        state.defs[action.payload.id].sourceModelId =
          action.payload.sourceModelId;
      },
      prepare: (id, sourceModelId) => ({ payload: { id, sourceModelId } }),
    },

    addNewDimension: {
      reducer: (state, action: PayloadAction<addNewDimensionPayload>) => {
        state.defs[action.payload.id].dimensions.push(action.payload.dimension);
      },
      prepare: (id: string, dimension: DimensionDefinition) => ({
        payload: { id, dimension },
      }),
    },
  },
});

// Action creators are generated for each case reducer function
export const { addEmptyMetricsDef, setMetricsDefModel } =
  metricsDefSlice.actions;

export default metricsDefSlice.reducer;
