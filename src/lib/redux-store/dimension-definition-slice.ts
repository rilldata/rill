import * as reduxToolkit from "@reduxjs/toolkit";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { generateApis } from "$lib/redux-store/slice-utils";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { measureDefSlice } from "$lib/redux-store/measure-definition-slice";

const { createSlice, createEntityAdapter } = reduxToolkit;

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
  },
});

export const {
  addManyDimensions,
  addOneDimension,
  updateDimension,
  removeDimension,
} = dimensionDefSlice.actions;

export const dimensionDefSliceReducer = dimensionDefSlice.reducer;

export const {
  fetchManyApi: fetchManyDimensionsApi,
  createApi: createDimensionsApi,
  updateApi: updateDimensionsApi,
  deleteApi: deleteDimensionsApi,
} = generateApis<
  EntityType.DimensionDefinition,
  { metricsDefId: string },
  { metricsDefId: string }
>(
  EntityType.DimensionDefinition,
  addManyDimensions,
  addOneDimension,
  updateDimension,
  removeDimension,
  "dimensions"
);

export const manyDimensionsSelector = (metricsDefId) => {
  return (state: RillReduxState) =>
    state.dimensionDefinition.ids
      .filter(
        (id) =>
          state.dimensionDefinition.entities[id].metricsDefId === metricsDefId
      )
      .map((id) => state.dimensionDefinition.entities[id]);
};
export const singleDimensionSelector = (dimensionId: number | string) => {
  return (state: RillReduxState) =>
    state.dimensionDefinition.entities[dimensionId];
};
