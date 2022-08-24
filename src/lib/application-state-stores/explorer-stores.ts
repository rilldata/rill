import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
import { Readable, writable } from "svelte/store";

export interface MetricsExplorerStoreType {
  entities: Record<string, MetricsExplorerEntity>;
}
const { update, subscribe } = writable({
  entities: {},
} as MetricsExplorerStoreType);

const updateMetricsExplorerById = (
  id: string,
  callback: (metricsExplorer: MetricsExplorerEntity) => void,
  absenceCallback?: () => MetricsExplorerEntity
) => {
  update((state) => {
    if (!state.entities[id]) {
      if (absenceCallback) {
        state.entities[id] = absenceCallback();
      }
      return state;
    }
    callback(state.entities[id]);
    return state;
  });
};

const metricViewReducers = {
  sync(id: string, meta: MetricViewMetaResponse) {
    if (!id || !meta || !meta.measures) return;
    updateMetricsExplorerById(
      id,
      (metricsExplorer) => {
        // sync measures with selected leaderboard measure.
        if (!metricsExplorer.leaderboardMeasureId && meta.measures.length) {
          metricsExplorer.leaderboardMeasureId = meta.measures[0].id;
        } else if (!meta.measures.length) {
          metricsExplorer.leaderboardMeasureId = undefined;
        }
        metricsExplorer.selectedMeasureIds = meta.measures.map(
          (measure) => measure.id
        );
      },
      () => ({
        id,
        selectedMeasureIds: meta.measures.map((measure) => measure.id),
        leaderboardMeasureId: meta.measures[0]?.id,
        filters: {
          include: [],
          exclude: [],
        },
      })
    );
  },

  setLeaderboardMeasureId(id: string, measureId: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureId = measureId;
    });
  },

  setSelectedTimeRange(id: string, timeRange: TimeSeriesTimeRange) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  toggleFilter(id: string, dimensionId: string, dimensionValue: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      const existingDimensionIndex = metricsExplorer.filters.include.findIndex(
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
        metricsExplorer.filters.include[existingDimensionIndex].values.indexOf(
          dimensionValue
        ) ?? -1;

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
    });
  },

  clearFilters(id: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.filters = {
        include: [],
        exclude: [],
      };
    });
  },
};
export const metricsExplorerStore: Readable<MetricsExplorerStoreType> &
  typeof metricViewReducers = {
  subscribe,

  ...metricViewReducers,
};
