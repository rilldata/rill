import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import {
  createEntityAdapter,
  createSlice,
} from "$lib/redux-store/redux-toolkit-wrapper";
import { setStatusPrepare } from "$lib/redux-store/utils/loading-utils";
import type { PayloadAction } from "@reduxjs/toolkit";

export interface LeaderboardValue {
  value: number;
  label: string;
}
export interface LeaderboardValues {
  values: Array<LeaderboardValue>;
  dimensionId: string;
  dimensionName?: string;
  status: EntityStatus;
}

export type ActiveValues = Record<string, Array<[unknown, boolean]>>;

export interface MetricsExplorerEntity {
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

const metricsExplorerAdapter = createEntityAdapter<MetricsExplorerEntity>();

export const exploreSlice = createSlice({
  name: "metricsExplorer",
  initialState: metricsExplorerAdapter.getInitialState(),
  reducers: {
    initMetricsExplorer: {
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
        const metricsExplorer: MetricsExplorerEntity = {
          id,
          measureIds: measures.map((measure) => measure.id),
          selectedMeasureIds: measures.map((measure) => measure.id),
          leaderboardMeasureId: measures[0]?.id,
          leaderboardFormatScale: "none",
          leaderboards: dimensions.map((column) => ({
            values: [],
            dimensionId: column.id,
            status: EntityStatus.Idle,
          })),
          activeValues: {},
          selectedCount: 0,
        };
        dimensions.forEach((column) => {
          metricsExplorer.activeValues[column.dimensionColumn] = [];
        });
        metricsExplorerAdapter.addOne(state, metricsExplorer);
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
        const metricsExplorer = state.entities[id];
        if (metricsExplorer.measureIds.indexOf(measureId) !== -1) return;
        metricsExplorer.measureIds = [...metricsExplorer.measureIds, measureId];
        // this makes it so that new measure gets selected by default.
        metricsExplorer.selectedMeasureIds = [
          ...metricsExplorer.selectedMeasureIds,
          measureId,
        ];
        if (!metricsExplorer.leaderboardMeasureId) {
          metricsExplorer.leaderboardMeasureId = measureId;
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
        const metricsExplorer = state.entities[id];
        if (metricsExplorer.measureIds.indexOf(measureId) === -1) return;
        metricsExplorer.measureIds = metricsExplorer.measureIds.filter(
          (existingMeasureId) => existingMeasureId !== measureId
        );
        metricsExplorer.selectedMeasureIds =
          metricsExplorer.selectedMeasureIds.filter(
            (existingMeasureId) => existingMeasureId !== measureId
          );
        if (metricsExplorer.leaderboardMeasureId === measureId) {
          metricsExplorer.leaderboardMeasureId =
            metricsExplorer.measureIds[0] ?? "";
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
        const metricsExplorer = state.entities[id];
        const existingIndex =
          metricsExplorer.selectedMeasureIds.indexOf(selectedMeasureId);

        if (existingIndex >= 0) {
          metricsExplorer.selectedMeasureIds =
            metricsExplorer.selectedMeasureIds.filter(
              (selectedMeasureId) => selectedMeasureId === selectedMeasureId
            );
        } else {
          metricsExplorer.selectedMeasureIds = [
            ...metricsExplorer.selectedMeasureIds,
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
        const metricsExplorer = state.entities[id];
        if (
          metricsExplorer.leaderboards.findIndex(
            (leaderboard) => leaderboard.dimensionId === dimensionId
          ) !== -1
        )
          return;
        metricsExplorer.leaderboards = [
          ...metricsExplorer.leaderboards,
          { dimensionId, values: [], status: EntityStatus.Idle },
        ];
        metricsExplorer.activeValues[dimensionId] = [];
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
        const metricsExplorer = state.entities[id];
        if (
          metricsExplorer.leaderboards.findIndex(
            (leaderboard) => leaderboard.dimensionId === dimensionId
          ) === -1
        )
          return;
        metricsExplorer.leaderboards = metricsExplorer.leaderboards.filter(
          (leaderboard) => leaderboard.dimensionId !== dimensionId
        );
        delete metricsExplorer.activeValues[dimensionId];
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
        const metricsExplorer = state.entities[id];
        const existingIndex = metricsExplorer.activeValues[
          dimensionId
        ]?.findIndex(([value]) => value === dimensionValue);
        const existing =
          metricsExplorer.activeValues[dimensionId]?.[existingIndex];

        if (existing) {
          if (existing[1] === include) {
            // if existing value is an 'include' then remove the value
            metricsExplorer.activeValues[dimensionId] =
              metricsExplorer.activeValues[dimensionId].filter(
                (activeValue) => activeValue[0] !== dimensionValue
              );
            metricsExplorer.selectedCount--;
          } else {
            // else toggle the 'include' of the value
            metricsExplorer.activeValues[dimensionId][existingIndex] = [
              existing[0],
              include,
            ];
          }
        } else {
          // add the value if not present
          metricsExplorer.activeValues[dimensionId] = [
            ...(metricsExplorer.activeValues[dimensionId] ?? []),
            [dimensionValue, include],
          ];
          metricsExplorer.selectedCount++;
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
          existing.status = EntityStatus.Idle;
        } else {
          state.entities[id].leaderboards = [
            ...state.entities[id].leaderboards,
            {
              dimensionId,
              values,
              status: EntityStatus.Idle,
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

    setLeaderboardValuesStatus: {
      reducer: (
        state,
        {
          payload: { id, status },
        }: PayloadAction<{ id: string; status: EntityStatus }>
      ) => {
        if (!state.entities[id]) return;
        state.entities[id].leaderboards = state.entities[id].leaderboards.map(
          (leaderboard) => ({
            dimensionId: leaderboard.dimensionId,
            values: leaderboard.values,
            status,
          })
        );
      },
      prepare: setStatusPrepare,
    },

    setLeaderboardValuesErrorStatus: {
      reducer: (state, { payload: id }: PayloadAction<string>) => {
        if (!state.entities[id]) return;
        state.entities[id].leaderboards = state.entities[id].leaderboards.map(
          (leaderboard) => {
            if (leaderboard.status === EntityStatus.Idle) return leaderboard;
            return {
              dimensionId: leaderboard.dimensionId,
              values: [],
              status: EntityStatus.Error,
            };
          }
        );
      },
      prepare: (id: string) => ({ payload: id }),
    },

    clearSelectedLeaderboardValues: {
      reducer: (state, { payload: id }: PayloadAction<string>) => {
        if (!state.entities[id]) return;
        state.entities[id].activeValues = {};
        state.entities[id].leaderboards = state.entities[id].leaderboards.map(
          (leaderboard) => ({
            dimensionId: leaderboard.dimensionId,
            values: [],
            status: EntityStatus.Idle,
          })
        );
      },
      prepare: (id: string) => ({ payload: id }),
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
  initMetricsExplorer,
  addMeasureToExplore,
  removeMeasureFromExplore,
  toggleExploreMeasure,
  setLeaderboardMeasureId,
  addDimensionToExplore,
  removeDimensionFromExplore,
  toggleLeaderboardActiveValue,
  setLeaderboardDimensionValues,
  setLeaderboardFormatScale,
  setLeaderboardValuesStatus,
  setLeaderboardValuesErrorStatus,
  clearSelectedLeaderboardValues,
  setExploreTimeRange,
  setExploreSelectedTimeRange,
} = exploreSlice.actions;
export const MetricsExplorerSliceActions = exploreSlice.actions;
export type MetricsExplorerSliceTypes = typeof MetricsExplorerSliceActions;

export const metricsExplorerReducer = exploreSlice.reducer;
