import type {
  MeasureDefinition,
  MetricsDefinitionEntity,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import * as reduxToolkit from "@reduxjs/toolkit";
import type { PayloadAction } from "@reduxjs/toolkit";
import type { DimensionDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { shallowCopy } from "$common/utils/shallowCopy";
import { createEntityAdapter } from "@reduxjs/toolkit";
const { createSlice } = reduxToolkit;

const metricsDefAdapter = createEntityAdapter<MetricsDefinitionEntity>({
  sortComparer: (a, b) => a.creationTime - b.creationTime,
});

export const metricsDefSlice = createSlice({
  name: "metricsDefinitions",
  initialState: metricsDefAdapter.getInitialState(),
  reducers: {
    bootstrapMetricsDefState: {
      reducer: (
        state,
        action: PayloadAction<Array<MetricsDefinitionEntity>>
      ) => {
        metricsDefAdapter.addMany(state, action.payload);
      },
      prepare: (metricsDefs: Array<MetricsDefinitionEntity>) => ({
        payload: metricsDefs,
      }),
    },

    addEmptyMetricsDef: {
      reducer: (state, action: PayloadAction<MetricsDefinitionEntity>) => {
        metricsDefAdapter.addOne(state, action.payload);
      },
      prepare: (metricsDef: MetricsDefinitionEntity) => ({
        payload: metricsDef,
      }),
    },

    updateMetricsDefLabel: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; label: string }>
      ) => {
        state.entities[action.payload.id].metricDefLabel = action.payload.label;
      },
      prepare: (id: string, label: string) => ({ payload: { id, label } }),
    },

    updateMetricsDefinitionModel: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; sourceModelId: string }>
      ) => {
        state.entities[action.payload.id].sourceModelId =
          action.payload.sourceModelId;
      },
      prepare: (id, sourceModelId) => ({ payload: { id, sourceModelId } }),
    },

    updateMetricsDefinitionTimestamp: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; timeDimension: string }>
      ) => {
        state.entities[action.payload.id].timeDimension =
          action.payload.timeDimension;
      },
      prepare: (id, timeDimension) => ({ payload: { id, timeDimension } }),
    },

    clearMetricsDimension: {
      reducer: (state, action: PayloadAction<{ id: string }>) => {
        state.entities[action.payload.id].dimensions = [];
        state.entities[action.payload.id].measures = [];
      },
      prepare: (id) => ({ payload: { id } }),
    },

    addNewDimension: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; dimension: DimensionDefinition }>
      ) => {
        state.entities[action.payload.id].dimensions.push(
          action.payload.dimension
        );
      },
      prepare: (id: string, dimension: DimensionDefinition) => ({
        payload: { id, dimension },
      }),
    },

    setDimensions: {
      reducer: (
        state,
        action: PayloadAction<{
          id: string;
          dimensions: Array<DimensionDefinition>;
        }>
      ) => {
        state.entities[action.payload.id].dimensions =
          action.payload.dimensions;
      },
      prepare: (id: string, dimensions: Array<DimensionDefinition>) => ({
        payload: { id, dimensions },
      }),
    },

    updateDimension: {
      reducer: (
        state,
        action: PayloadAction<{
          id: string;
          dimensionId: string;
          modifications: Partial<DimensionDefinition>;
        }>
      ) => {
        const dimension = state.entities[action.payload.id].dimensions.find(
          (dim) => dim.id === action.payload.dimensionId
        );
        shallowCopy(action.payload.modifications, dimension);
      },
      prepare: (
        id: string,
        dimensionId: string,
        modifications: Partial<DimensionDefinition>
      ) => ({
        payload: { id, dimensionId, modifications },
      }),
    },

    addNewMeasure: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; measure: MeasureDefinition }>
      ) => {
        state.entities[action.payload.id].measures.push(action.payload.measure);
      },
      prepare: (id: string, measure: MeasureDefinition) => ({
        payload: { id, measure },
      }),
    },

    setMeasures: {
      reducer: (
        state,
        action: PayloadAction<{
          id: string;
          measures: Array<MeasureDefinition>;
        }>
      ) => {
        state.entities[action.payload.id].measures = action.payload.measures;
      },
      prepare: (id: string, measures: Array<MeasureDefinition>) => ({
        payload: { id, measures },
      }),
    },

    updateMeasure: {
      reducer: (
        state,
        action: PayloadAction<{
          id: string;
          measureId: string;
          modifications: Partial<MeasureDefinition>;
        }>
      ) => {
        const measure = state.entities[action.payload.id].measures.find(
          (dim) => dim.id === action.payload.measureId
        );
        shallowCopy(action.payload.modifications, measure);
      },
      prepare: (
        id: string,
        measureId: string,
        modifications: Partial<MeasureDefinition>
      ) => ({
        payload: { id, measureId, modifications },
      }),
    },
  },
});

export const {
  bootstrapMetricsDefState,
  addEmptyMetricsDef,

  updateMetricsDefLabel,
  updateMetricsDefinitionModel,
  updateMetricsDefinitionTimestamp,
  clearMetricsDimension,

  addNewDimension,
  setDimensions,
  updateDimension,

  addNewMeasure,
  setMeasures,
  updateMeasure,
} = metricsDefSlice.actions;
export const MetricsDefSliceActions = metricsDefSlice.actions;
export type MetricsDefSliceActionTypes = typeof MetricsDefSliceActions;

export default metricsDefSlice.reducer;
