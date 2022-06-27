import * as reduxToolkit from "@reduxjs/toolkit";
import type { PayloadAction } from "@reduxjs/toolkit";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { generateBasicSelectors } from "$lib/redux-store/slice-utils";
import { fetchWrapper, streamingFetchWrapper } from "$lib/util/fetchWrapper";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../routes/_surfaces/workspace/leaderboard/utils";

const { createSlice, createEntityAdapter, createAsyncThunk } = reduxToolkit;

export interface LeaderboardValues {
  values: Array<unknown>;
  displayName: string;
}

export type ActiveValues = Record<string, Array<[unknown, boolean]>>;

export interface MetricsLeaderboardEntity {
  id: string;
  measureId: string;
  bigNumber: number;
  referenceValue: number;
  leaderboards: Array<LeaderboardValues>;
  activeValues: ActiveValues;
  selectedCount: number;
}

const metricsLeaderboardAdapter =
  createEntityAdapter<MetricsLeaderboardEntity>();

export const metricsLeaderboardSlice = createSlice({
  name: "metricsLeaderboard",
  initialState: metricsLeaderboardAdapter.getInitialState(),
  reducers: {
    initMetricsLeaderboard: {
      reducer: (
        state,
        {
          payload: { id, dimensions },
        }: PayloadAction<{
          id: string;
          dimensions: Array<DimensionDefinitionEntity>;
        }>
      ) => {
        if (state.entities[id]) return;
        const metricsLeaderboard = {
          id,
          measureId: "",
          bigNumber: 0,
          referenceValue: 0,
          leaderboards: dimensions.map((column) => ({
            values: [],
            displayName: column.dimensionColumn,
          })),
          activeValues: {},
          selectedCount: 0,
        };
        dimensions.forEach((column) => {
          state.entities[id].activeValues[column.dimensionColumn] = [];
        });
        metricsLeaderboardAdapter.addOne(state, metricsLeaderboard);
      },
      prepare: (id: string, dimensions: Array<DimensionDefinitionEntity>) => ({
        payload: { id, dimensions },
      }),
    },

    setMeasureId: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; measureId: string }>
      ) => {
        if (!state.entities[action.payload.id]) return;
        state.entities[action.payload.id].measureId = action.payload.measureId;
      },
      prepare: (id: string, measureId: string) => ({
        payload: { id, measureId },
      }),
    },

    toggleLeaderboardActiveValue: {
      reducer: (
        state,
        {
          payload,
        }: PayloadAction<{
          id: string;
          dimensionName: string;
          dimensionValue: unknown;
          include: boolean;
        }>
      ) => {
        if (!state.entities[payload.id]) return;
        const metricsLeaderboard = state.entities[payload.id];
        const existingIndex = metricsLeaderboard.activeValues[
          payload.dimensionName
        ]?.findIndex(([value]) => value === payload.dimensionValue);
        const existing =
          metricsLeaderboard.activeValues[payload.dimensionName]?.[
            existingIndex
          ];

        if (existing) {
          if (existing[1] === payload.include) {
            // if existing value is an 'include' then remove the value
            metricsLeaderboard.activeValues[payload.dimensionName] =
              metricsLeaderboard.activeValues[payload.dimensionName].filter(
                (activeValue) => activeValue !== payload.dimensionValue
              );
            metricsLeaderboard.selectedCount--;
          } else {
            // else toggle the 'include' of the value
            metricsLeaderboard.activeValues[payload.dimensionName][
              existingIndex
            ] = [existing[0], payload.include];
          }
        } else {
          // add the value if not present
          metricsLeaderboard.activeValues[payload.dimensionName] = [
            ...(metricsLeaderboard.activeValues[payload.dimensionName] ?? []),
            [payload.dimensionValue, payload.include],
          ];
          metricsLeaderboard.selectedCount++;
        }
      },
      prepare: (
        id: string,
        dimensionName: string,
        dimensionValue: unknown,
        include = true
      ) => ({
        payload: { id, dimensionName, dimensionValue, include },
      }),
    },

    setDimensionLeaderboard: {
      reducer: (
        state,
        action: PayloadAction<{
          id: string;
          dimensionName: string;
          values: Array<unknown>;
        }>
      ) => {
        if (!state.entities[action.payload.id]) return;
        const existing = state.entities[action.payload.id].leaderboards.find(
          (leaderboard) =>
            leaderboard.displayName === action.payload.dimensionName
        );
        if (existing) {
          existing.displayName = action.payload.dimensionName;
          existing.values = action.payload.values;
        } else {
          state.entities[action.payload.id].leaderboards = [
            ...state.entities[action.payload.id].leaderboards,
            {
              displayName: action.payload.dimensionName,
              values: action.payload.values,
            },
          ];
        }
      },
      prepare: (id: string, dimensionName: string, values: Array<unknown>) => ({
        payload: { id, dimensionName, values },
      }),
    },

    setBigNumber: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; bigNumber: number }>
      ) => {
        if (!state.entities[action.payload.id]) return;
        state.entities[action.payload.id].bigNumber = action.payload.bigNumber;
        if (state.entities[action.payload.id].selectedCount > 0) {
          state.entities[action.payload.id].referenceValue =
            action.payload.bigNumber;
        }
      },
      prepare: (id: string, bigNumber: number) => ({
        payload: { id, bigNumber },
      }),
    },

    clearLeaderboard: {
      reducer: (state, { payload: id }: PayloadAction<string>) => {
        if (!state.entities[id]) return;
        state.entities[id].activeValues = {};
        state.entities[id].leaderboards = state.entities[id].leaderboards.map(
          (leaderboard) => ({
            displayName: leaderboard.displayName,
            values: [],
          })
        );
      },
      prepare: (id) => ({ payload: id }),
    },
  },
});

export const {
  initMetricsLeaderboard,
  setMeasureId,
  toggleLeaderboardActiveValue,
  setDimensionLeaderboard,
  setBigNumber,
  clearLeaderboard,
} = metricsLeaderboardSlice.actions;
export const MetricsLeaderboardSliceActions = metricsLeaderboardSlice.actions;
export type MetricsLeaderboardSliceTypes =
  typeof MetricsLeaderboardSliceActions;

export const metricsLeaderboardReducer = metricsLeaderboardSlice.reducer;

export const updateLeaderboardMeasure = (
  dispatch,
  id: string,
  measureId: string
) => {
  dispatch(setMeasureId(id, measureId));
  dispatch(updateLeaderboardApi(id));
};

export const toggleValueAndUpdateLeaderboard = (
  dispatch,
  id: string,
  dimensionName: string,
  dimensionValue: unknown,
  include = true
) => {
  dispatch(
    toggleLeaderboardActiveValue(id, dimensionName, dimensionValue, include)
  );
  dispatch(updateLeaderboardApi(id));
};

export const clearLeaderboardAndUpdate = (dispatch, id: string) => {
  dispatch(clearLeaderboard(id));
  dispatch(updateLeaderboardApi(id));
};

export const updateLeaderboardApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/updateLeaderboard`,
  async (id: string, thunkAPI) => {
    const metricsLeaderboard: MetricsLeaderboardEntity = (
      thunkAPI.getState() as RillReduxState
    ).metricsLeaderboard.entities[id];
    const filters = prune(metricsLeaderboard.activeValues);
    const requestBody = {
      measureId: metricsLeaderboard.measureId,
      filters,
    };

    thunkAPI.dispatch(
      setBigNumber(
        metricsLeaderboard.id,
        await fetchWrapper(
          `metrics/${metricsLeaderboard.id}/bigNumber`,
          "POST",
          requestBody
        )
      )
    );
    const stream = streamingFetchWrapper<{
      dimensionName: string;
      values: Array<unknown>;
    }>(`metrics/${metricsLeaderboard.id}/leaderboards`, "POST", requestBody);
    for await (const dimensionData of stream) {
      thunkAPI.dispatch(
        setDimensionLeaderboard(
          metricsLeaderboard.id,
          dimensionData.dimensionName,
          dimensionData.values
        )
      );
    }
  },
  {
    condition: (id: string, { getState }) => {
      const metricsLeaderboard: MetricsLeaderboardEntity = (
        getState() as RillReduxState
      ).metricsLeaderboard.entities[id];
      return metricsLeaderboard.measureId !== "";
    },
  }
);

export const {
  manySelector: manyMetricsLeaderboardSelector,
  singleSelector: singleMetricsLeaderboardSelector,
} = generateBasicSelectors("metricsLeaderboard");
