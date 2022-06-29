import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { PayloadAction } from "@reduxjs/toolkit";

const dimensionDefAdapter = createEntityAdapter<DimensionDefinitionEntity>({
  sortComparer: (a, b) => a.creationTime - b.creationTime,
});

export const dimensionDefSlice = createSlice({
  name: "dimensionDefinition",
  initialState: dimensionDefAdapter.getInitialState(),
  reducers: {
    addManyDimensions: {
      reducer: dimensionDefAdapter.addMany,
      prepare: (dimensions: Array<DimensionDefinitionEntity>) => ({
        payload: dimensions,
      }),
    },

    addOneDimension: {
      reducer: dimensionDefAdapter.addOne,
      prepare: (dimension: DimensionDefinitionEntity) => ({
        payload: dimension,
      }),
    },

    updateDimension: {
      reducer: dimensionDefAdapter.updateOne,
      prepare: (id: string, dimension: Partial<DimensionDefinitionEntity>) => ({
        payload: { id, changes: dimension },
      }),
    },

    removeDimension: {
      reducer: dimensionDefAdapter.removeOne,
      prepare: (id: string) => ({ payload: id }),
    },

    clearDimensionsForMetricsDefId: {
      reducer: (state, action: PayloadAction<string>) => {
        dimensionDefAdapter.removeMany(
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
  addManyDimensions,
  addOneDimension,
  updateDimension,
  removeDimension,
  clearDimensionsForMetricsDefId,
} = dimensionDefSlice.actions;

export const dimensionDefSliceReducer = dimensionDefSlice.reducer;
