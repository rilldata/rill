import { writable } from "svelte/store";
import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
import type { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

export interface MetricsExplorerStoreType {
  entities: Record<string, MetricsExplorerEntity>;
}
export const MetricsExplorerStore = writable({
  entities: {},
} as MetricsExplorerStoreType);

const UpdateMetricsExplorer = (
  id: string,
  callback: (metricsExplorer: MetricsExplorerEntity) => void,
  absenceCallback?: () => MetricsExplorerEntity
) => {
  MetricsExplorerStore.update((state) => {
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

export const syncMetricsExplorer = (
  id: string,
  meta: MetricViewMetaResponse
) => {
  if (!id || !meta || !meta.measures) return;
  UpdateMetricsExplorer(
    id,
    (metricsExplorer) => {
      // sync measures with selected leaderboard measure.
      if (!metricsExplorer.leaderboardMeasureId && meta.measures.length) {
        metricsExplorer.leaderboardMeasureId = meta.measures[0].id;
      } else if (!meta.measures.length) {
        metricsExplorer.leaderboardMeasureId = undefined;
      }
      // TODO: update selected measure id. This is not being used right now
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
};

export const setMetricsExplorerLeaderboardMeasureId = (
  id: string,
  measureId: string
) => {
  UpdateMetricsExplorer(id, (metricsExplorer) => {
    metricsExplorer.leaderboardMeasureId = measureId;
  });
};

export const setMetricsExplorerSelectedTimeRange = (
  id: string,
  timeRange: TimeSeriesTimeRange
) => {
  UpdateMetricsExplorer(id, (metricsExplorer) => {
    metricsExplorer.selectedTimeRange = timeRange;
  });
};
