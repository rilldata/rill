import * as reduxToolkit from "@reduxjs/toolkit";
import type { PayloadAction } from "@reduxjs/toolkit";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

const { createSlice, createEntityAdapter } = reduxToolkit;

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
}

const metricsLeaderboardAdapter =
  createEntityAdapter<MetricsLeaderboardEntity>();

export const metricsLeaderboardSlice = createSlice({
  name: "metricsLeaderboard",
  initialState: metricsLeaderboardAdapter.getInitialState(),
  reducers: {
    initMetricsLeaderboard: {
      reducer: (state, action: PayloadAction<MetricsDefinitionEntity>) => {
        // metricsLeaderboardAdapter.addOne(state, {
        //   id: action.payload.id,
        //   measureId: action.payload.measures?.[0]?.id ?? "",
        //   bigNumber: 0,
        //   referenceValue: 0,
        //   leaderboards: action.payload.dimensions.map((column) => ({
        //     values: [],
        //     displayName: column.dimensionColumn,
        //   })),
        //   activeValues: {},
        // });
      },
      prepare: (metricsDef: MetricsDefinitionEntity) => ({
        payload: metricsDef,
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
          } else {
            // else toggle the 'include' of the value
            metricsLeaderboard.activeValues[payload.dimensionName][
              existingIndex
            ] = [existing[0], payload.include];
          }
        } else {
          // add the value if not present
          metricsLeaderboard.activeValues[payload.dimensionName] = [
            ...metricsLeaderboard.activeValues[payload.dimensionName],
            [payload.dimensionValue, payload.include],
          ];
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
      },
      prepare: (id: string, bigNumber: number) => ({
        payload: { id, bigNumber },
      }),
    },

    setReferenceValue: {
      reducer: (
        state,
        action: PayloadAction<{ id: string; referenceValue: number }>
      ) => {
        if (!state.entities[action.payload.id]) return;
        state.entities[action.payload.id].referenceValue =
          action.payload.referenceValue;
      },
      prepare: (id: string, referenceValue: number) => ({
        payload: { id, referenceValue },
      }),
    },
  },
});

export const {
  initMetricsLeaderboard,
  setMeasureId,
  toggleLeaderboardActiveValue,
  setDimensionLeaderboard,
  setBigNumber,
  setReferenceValue,
} = metricsLeaderboardSlice.actions;
export const MetricsLeaderboardSliceActions = metricsLeaderboardSlice.actions;
export type MetricsLeaderboardSliceTypes =
  typeof MetricsLeaderboardSliceActions;

export const metricsLeaderboardReducer = metricsLeaderboardSlice.reducer;
