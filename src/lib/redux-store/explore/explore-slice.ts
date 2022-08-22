import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { MetricViewRequestFilter } from "$common/rill-developer-service/MetricViewActions";
import { removeIfExists } from "$common/utils/arrayUtils";
import {
  createEntityAdapter,
  createSlice,
} from "$lib/redux-store/redux-toolkit-wrapper";
import {
  setFieldPrepare,
  setFieldReducer,
} from "$lib/redux-store/utils/slice-utils";
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
  filters: MetricViewRequestFilter;
  selectedCount: number;
  // user selected time range
  selectedTimeRange?: TimeSeriesTimeRange;
  // this marks whether anything related to this explore is stale
  // this is set to true when any measure or dimension changes.
  // this also is set to true when related model and its dependant source updates (TODO)
  isStale: boolean;
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
          leaderboards: dimensions.map((column) => ({
            values: [],
            dimensionId: column.id,
            status: EntityStatus.Idle,
          })),
          filters: {
            include: [],
            exclude: [],
          },
          selectedCount: 0,
          isStale: false,
        };
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
      reducer: setFieldReducer("leaderboardMeasureId"),
      prepare: setFieldPrepare<MetricsExplorerEntity, "leaderboardMeasureId">(
        "leaderboardMeasureId"
      ),
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
        removeIfExists(
          metricsExplorer.filters.include,
          (d) => d.name === dimensionId
        );
        removeIfExists(
          metricsExplorer.filters.include,
          (d) => d.name === dimensionId
        );
      },
      prepare: (id: string, dimensionId: string) => ({
        payload: { id, dimensionId },
      }),
    },

    // TODO: create a separate reducer for exclude value
    toggleLeaderboardActiveValue: {
      reducer: (
        state,
        {
          payload: { id, dimensionId, dimensionValue },
        }: PayloadAction<{
          id: string;
          dimensionId: string;
          dimensionValue: unknown;
        }>
      ) => {
        if (!state.entities[id]) return;
        const metricsExplorer = state.entities[id];
        const existingDimensionIndex =
          metricsExplorer.filters.include.findIndex(
            (dimensionValues) => dimensionValues.name === dimensionId
          );

        // if entry for dimension doesnt exist, add it
        if (existingDimensionIndex === -1) {
          metricsExplorer.filters.include.push({
            name: dimensionId,
            values: [dimensionValue],
          });
          return;
        }

        const existingIncludeIndex =
          metricsExplorer.filters.include[
            existingDimensionIndex
          ].values.indexOf(dimensionValue) ?? -1;

        // add the value if it doesn't exist, remove the value if it does exist
        if (existingIncludeIndex === -1) {
          metricsExplorer.filters.include[existingDimensionIndex].values.push(
            dimensionValue
          );
        } else {
          metricsExplorer.filters.include[existingDimensionIndex].values.splice(
            existingIncludeIndex,
            1
          );
          // remove the entry for dimension if no values are selected.
          if (
            metricsExplorer.filters.include[existingDimensionIndex].values
              .length === 0
          ) {
            metricsExplorer.filters.include.splice(existingDimensionIndex, 1);
          }
        }
      },
      prepare: (id: string, dimensionId: string, dimensionValue: unknown) => ({
        payload: { id, dimensionId, dimensionValue },
      }),
    },

    clearSelectedLeaderboardValues: {
      reducer: (state, { payload: id }: PayloadAction<string>) => {
        if (!state.entities[id]) return;
        state.entities[id].filters = {
          include: [],
          exclude: [],
        };
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

    setExplorerIsStale: {
      reducer: setFieldReducer("isStale"),
      prepare: setFieldPrepare<MetricsExplorerEntity, "isStale">("isStale"),
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
  clearSelectedLeaderboardValues,
  setExploreSelectedTimeRange,
  setExplorerIsStale,
} = exploreSlice.actions;
export const MetricsExplorerSliceActions = exploreSlice.actions;
export type MetricsExplorerSliceTypes = typeof MetricsExplorerSliceActions;

export const metricsExplorerReducer = exploreSlice.reducer;
