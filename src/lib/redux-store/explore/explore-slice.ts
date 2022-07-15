import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { PayloadAction } from "@reduxjs/toolkit";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

export interface LeaderboardValue {
  value: number;
  label: string;
}
export interface LeaderboardValues {
  values: Array<LeaderboardValue>;
  dimensionId: string;
  dimensionName?: string;
}

export type ActiveValues = Record<string, Array<[unknown, boolean]>>;

export interface MetricsExploreEntity {
  id: string;
  // full list of measure IDs available to explore
  measureIds: Array<string>;
  // selected measure IDs to be shown
  selectedMeasureIds: Array<string>;
  // this is used to show leaderboard values
  leaderboardMeasureId: string;
  leaderboards: Array<LeaderboardValues>;
  activeValues: ActiveValues;
  selectedCount: number;
  // time range of the selected timestamp column
  timeRange?: TimeSeriesTimeRange;
  // user selected time range
  selectedTimeRange?: TimeSeriesTimeRange;
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
          selectedMeasureIds: measures.map((measure) => measure.id),
          leaderboardMeasureId: measures[0]?.id,
          leaderboards: dimensions.map((column) => ({
            values: [],
            dimensionId: column.id,
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

    addMeasureToExplore: {
      reducer: (
        state,
        {
          payload: { id, measureId },
        }: PayloadAction<{ id: string; measureId: string }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        if (metricsExplore.measureIds.indexOf(measureId) !== -1) return;
        metricsExplore.measureIds = [...metricsExplore.measureIds, measureId];
        // this makes it so that new measure gets selected by default.
        metricsExplore.selectedMeasureIds = [
          ...metricsExplore.selectedMeasureIds,
          measureId,
        ];
        if (!metricsExplore.leaderboardMeasureId) {
          metricsExplore.leaderboardMeasureId = measureId;
        }
      },
      prepare: (id: string, measureId: string) => ({
        payload: { id, measureId },
      }),
    },

    removeMeasureFromExplore: {
      reducer: (
        state,
        {
          payload: { id, measureId },
        }: PayloadAction<{ id: string; measureId: string }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        if (metricsExplore.measureIds.indexOf(measureId) === -1) return;
        metricsExplore.measureIds = metricsExplore.measureIds.filter(
          (existingMeasureId) => existingMeasureId !== measureId
        );
        metricsExplore.selectedMeasureIds =
          metricsExplore.selectedMeasureIds.filter(
            (existingMeasureId) => existingMeasureId !== measureId
          );
        if (metricsExplore.leaderboardMeasureId === measureId) {
          metricsExplore.leaderboardMeasureId =
            metricsExplore.measureIds[0] ?? "";
        }
      },
      prepare: (id: string, measureId: string) => ({
        payload: { id, measureId },
      }),
    },

    toggleExploreMeasure: {
      reducer: (
        state,
        {
          payload: { id, selectedMeasureId },
        }: PayloadAction<{ id: string; selectedMeasureId: string }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        const existingIndex =
          metricsExplore.selectedMeasureIds.indexOf(selectedMeasureId);

        if (existingIndex >= 0) {
          metricsExplore.selectedMeasureIds =
            metricsExplore.selectedMeasureIds.filter(
              (selectedMeasureId) => selectedMeasureId === selectedMeasureId
            );
        } else {
          metricsExplore.selectedMeasureIds = [
            ...metricsExplore.selectedMeasureIds,
            selectedMeasureId,
          ];
        }
      },
      prepare: (id: string, selectedMeasureId: string) => ({
        payload: { id, selectedMeasureId },
      }),
    },

    setLeaderboardMeasureId: {
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

    addDimensionToExplore: {
      reducer: (
        state,
        {
          payload: { id, dimensionId },
        }: PayloadAction<{
          id: string;
          dimensionId: string;
        }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        if (
          metricsExplore.leaderboards.findIndex(
            (leaderboard) => leaderboard.dimensionId === dimensionId
          ) !== -1
        )
          return;
        metricsExplore.leaderboards = [
          ...metricsExplore.leaderboards,
          { dimensionId, values: [] },
        ];
        metricsExplore.activeValues[dimensionId] = [];
      },
      prepare: (id: string, dimensionId: string) => ({
        payload: { id, dimensionId },
      }),
    },

    removeDimensionFromExplore: {
      reducer: (
        state,
        {
          payload: { id, dimensionId },
        }: PayloadAction<{
          id: string;
          dimensionId: string;
        }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        if (
          metricsExplore.leaderboards.findIndex(
            (leaderboard) => leaderboard.dimensionId === dimensionId
          ) === -1
        )
          return;
        metricsExplore.leaderboards = metricsExplore.leaderboards.filter(
          (leaderboard) => leaderboard.dimensionId !== dimensionId
        );
        delete metricsExplore.activeValues[dimensionId];
      },
      prepare: (id: string, dimensionId: string) => ({
        payload: { id, dimensionId },
      }),
    },

    toggleLeaderboardActiveValue: {
      reducer: (
        state,
        {
          payload: { id, dimensionId, dimensionValue, include },
        }: PayloadAction<{
          id: string;
          dimensionId: string;
          dimensionValue: unknown;
          include: boolean;
        }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplore = state.entities[id];
        const existingIndex = metricsExplore.activeValues[
          dimensionId
        ]?.findIndex(([value]) => value === dimensionValue);
        const existing =
          metricsExplore.activeValues[dimensionId]?.[existingIndex];

        if (existing) {
          if (existing[1] === include) {
            // if existing value is an 'include' then remove the value
            metricsExplore.activeValues[dimensionId] =
              metricsExplore.activeValues[dimensionId].filter(
                (activeValue) => activeValue[0] !== dimensionValue
              );
            metricsExplore.selectedCount--;
          } else {
            // else toggle the 'include' of the value
            metricsExplore.activeValues[dimensionId][existingIndex] = [
              existing[0],
              include,
            ];
          }
        } else {
          // add the value if not present
          metricsExplore.activeValues[dimensionId] = [
            ...(metricsExplore.activeValues[dimensionId] ?? []),
            [dimensionValue, include],
          ];
          metricsExplore.selectedCount++;
        }
      },
      prepare: (
        id: string,
        dimensionId: string,
        dimensionValue: unknown,
        include = true
      ) => ({
        payload: { id, dimensionId, dimensionValue, include },
      }),
    },

    setLeaderboardDimensionValues: {
      reducer: (
        state,
        {
          payload: { id, dimensionId, values },
        }: PayloadAction<{
          id: string;
          values: Array<LeaderboardValue>;
          dimensionId: string;
        }>
      ) => {
        if (!state.entities[id]) return;
        const existing = state.entities[id].leaderboards.find(
          (leaderboard) => leaderboard.dimensionId === dimensionId
        );
        if (existing) {
          existing.dimensionId = dimensionId;
          existing.values = values;
        } else {
          state.entities[id].leaderboards = [
            ...state.entities[id].leaderboards,
            {
              dimensionId,
              values,
            },
          ];
        }
      },
      prepare: (
        id: string,
        dimensionId: string,
        values: Array<LeaderboardValue>
      ) => ({
        payload: { id, dimensionId, values },
      }),
    },

    clearSelectedLeaderboardValues: {
      reducer: (state, { payload: id }: PayloadAction<string>) => {
        if (!state.entities[id]) return;
        state.entities[id].activeValues = {};
        state.entities[id].leaderboards = state.entities[id].leaderboards.map(
          (leaderboard) => ({
            dimensionId: leaderboard.dimensionId,
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

    setExploreSelectedTimeRange: {
      reducer: (
        state,
        {
          payload: { id, selectedTimeRange },
        }: PayloadAction<{
          id: string;
          selectedTimeRange: Partial<TimeSeriesTimeRange>;
        }>
      ) => {
        if (!state.entities[id]) return;
        // overrides only the ones passed
        state.entities[id].selectedTimeRange = {
          ...(state.entities[id].selectedTimeRange ?? {}),
          ...selectedTimeRange,
        };
      },
      prepare: (
        id: string,
        selectedTimeRange: Partial<TimeSeriesTimeRange>
      ) => ({
        payload: { id, selectedTimeRange },
      }),
    },
  },
});

export const {
  initMetricsExplore,
  addMeasureToExplore,
  removeMeasureFromExplore,
  toggleExploreMeasure,
  setLeaderboardMeasureId,
  addDimensionToExplore,
  removeDimensionFromExplore,
  toggleLeaderboardActiveValue,
  setLeaderboardDimensionValues,
  clearSelectedLeaderboardValues,
  setExploreTimeRange,
  setExploreSelectedTimeRange,
} = exploreSlice.actions;
export const MetricsLeaderboardSliceActions = exploreSlice.actions;
export type MetricsLeaderboardSliceTypes =
  typeof MetricsLeaderboardSliceActions;

export const metricsLeaderboardReducer = exploreSlice.reducer;
