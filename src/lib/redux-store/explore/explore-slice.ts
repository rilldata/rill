import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { PayloadAction } from "@reduxjs/toolkit";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

export interface LeaderboardValues {
  values: Array<unknown>;
  displayName: string;
}

export type ActiveValues = Record<string, Array<[unknown, boolean]>>;

export interface MetricsExploreEntity {
  id: string;
  measureIds: Array<string>;
  // this is used to show leaderboard values
  leaderboardMeasureId: string;
  leaderboards: Array<LeaderboardValues>;
  activeValues: ActiveValues;
  selectedCount: number;
  // time range of the selected timestamp column
  timeRange?: TimeSeriesTimeRange;
}

const metricsExploreAdapter = createEntityAdapter<MetricsExploreEntity>();

export const exploreSlice = createSlice({
  name: "metricsLeaderboard",
  initialState: metricsExploreAdapter.getInitialState(),
  reducers: {
    initMetricsExplore: {
      reducer: (
        state,
        {
          payload: { id, dimensions, measures },
        }: PayloadAction<{
          id: string;
          dimensions: Array<DimensionDefinitionEntity>;
          measures: Array<MeasureDefinitionEntity>;
        }>
      ) => {
        if (state.entities[id]) return;
        const metricsExplore: MetricsExploreEntity = {
          id,
          measureIds: measures.map((measure) => measure.id),
          leaderboardMeasureId: measures[0]?.id,
          leaderboards: dimensions.map((column) => ({
            values: [],
            displayName: column.dimensionColumn,
          })),
          activeValues: {},
          selectedCount: 0,
        };
        dimensions.forEach((column) => {
          metricsExplore.activeValues[column.dimensionColumn] = [];
        });
        metricsExploreAdapter.addOne(state, metricsExplore);
      },
      prepare: (
        id: string,
        dimensions: Array<DimensionDefinitionEntity>,
        measures: Array<MeasureDefinitionEntity>
      ) => ({
        payload: { id, dimensions, measures },
      }),
    },

    toggleExploreMeasure: {
      reducer: (
        state,
        {
          payload: { id, measureId },
        }: PayloadAction<{ id: string; measureId: string }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        const existingIndex = metricsExplore.measureIds.indexOf(measureId);

        if (existingIndex >= 0) {
          metricsExplore.measureIds = metricsExplore.measureIds.filter(
            (selectedMeasureId) => selectedMeasureId === measureId
          );
        } else {
          metricsExplore.measureIds = [...metricsExplore.measureIds, measureId];
        }
      },
      prepare: (id: string, measureId: string) => ({
        payload: { id, measureId },
      }),
    },

    setMeasureId: {
      reducer: (
        state,
        {
          payload: { id, leaderboardMeasureId },
        }: PayloadAction<{ id: string; leaderboardMeasureId: string }>
      ) => {
        if (!state.entities[id]) return;
        state.entities[id].leaderboardMeasureId = leaderboardMeasureId;
      },
      prepare: (id: string, leaderboardMeasureId: string) => ({
        payload: { id, leaderboardMeasureId },
      }),
    },

    toggleLeaderboardActiveValue: {
      reducer: (
        state,
        {
          payload: { id, dimensionName, dimensionValue, include },
        }: PayloadAction<{
          id: string;
          dimensionName: string;
          dimensionValue: unknown;
          include: boolean;
        }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        const existingIndex = metricsExplore.activeValues[
          dimensionName
        ]?.findIndex(([value]) => value === dimensionValue);
        const existing =
          metricsExplore.activeValues[dimensionName]?.[existingIndex];

        if (existing) {
          if (existing[1] === include) {
            // if existing value is an 'include' then remove the value
            metricsExplore.activeValues[dimensionName] =
              metricsExplore.activeValues[dimensionName].filter(
                (activeValue) => activeValue !== dimensionValue
              );
            metricsExplore.selectedCount--;
          } else {
            // else toggle the 'include' of the value
            metricsExplore.activeValues[dimensionName][existingIndex] = [
              existing[0],
              include,
            ];
          }
        } else {
          // add the value if not present
          metricsExplore.activeValues[dimensionName] = [
            ...(metricsExplore.activeValues[dimensionName] ?? []),
            [dimensionValue, include],
          ];
          metricsExplore.selectedCount++;
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

    setLeaderboardDimensionValues: {
      reducer: (
        state,
        {
          payload: { id, dimensionName, values },
        }: PayloadAction<{
          id: string;
          dimensionName: string;
          values: Array<unknown>;
        }>
      ) => {
        if (!state.entities[id]) return;
        const existing = state.entities[id].leaderboards.find(
          (leaderboard) => leaderboard.displayName === dimensionName
        );
        if (existing) {
          existing.displayName = dimensionName;
          existing.values = values;
        } else {
          state.entities[id].leaderboards = [
            ...state.entities[id].leaderboards,
            {
              displayName: dimensionName,
              values: values,
            },
          ];
        }
      },
      prepare: (id: string, dimensionName: string, values: Array<unknown>) => ({
        payload: { id, dimensionName, values },
      }),
    },

    clearSelectedLeaderboardValues: {
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

    setExploreTimeRange: {
      reducer: (
        state,
        {
          payload: { id, timeRange },
        }: PayloadAction<{ id: string; timeRange: TimeSeriesTimeRange }>
      ) => {
        if (!state.entities[id]) return;
        state.entities[id].timeRange = timeRange;
      },
      prepare: (id: string, timeRange: TimeSeriesTimeRange) => ({
        payload: { id, timeRange },
      }),
    },
  },
});

export const {
  initMetricsExplore,
  toggleExploreMeasure,
  setMeasureId,
  toggleLeaderboardActiveValue,
  setLeaderboardDimensionValues,
  clearSelectedLeaderboardValues,
  setExploreTimeRange,
} = exploreSlice.actions;
export const MetricsLeaderboardSliceActions = exploreSlice.actions;
export type MetricsLeaderboardSliceTypes =
  typeof MetricsLeaderboardSliceActions;

export const metricsLeaderboardReducer = exploreSlice.reducer;
