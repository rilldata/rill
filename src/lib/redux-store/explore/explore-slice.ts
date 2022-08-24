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
  measureIds?: Array<string>;
  // selected measure IDs to be shown
  selectedMeasureIds: Array<string>;
  // this is used to show leaderboard values
  leaderboardMeasureId: string;
  leaderboards?: Array<LeaderboardValues>;
  filters: MetricViewRequestFilter;
  selectedCount?: number;
  // user selected time range
  selectedTimeRange?: TimeSeriesTimeRange;
  // this marks whether anything related to this explore is stale
  // this is set to true when any measure or dimension changes.
  // this also is set to true when related model and its dependant source updates (TODO)
  isStale?: boolean;
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
  addDimensionToExplore,
  removeDimensionFromExplore,
  setExplorerIsStale,
} = exploreSlice.actions;
export const MetricsExplorerSliceActions = exploreSlice.actions;
export type MetricsExplorerSliceTypes = typeof MetricsExplorerSliceActions;

export const metricsExplorerReducer = exploreSlice.reducer;
