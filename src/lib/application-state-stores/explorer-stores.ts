import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type {
  MetricsViewMetaResponse,
  MetricsViewRequestFilter,
} from "$common/rill-developer-service/MetricsViewActions";
import { removeIfExists } from "$common/utils/arrayUtils";
import { Readable, writable } from "svelte/store";

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

export interface MetricsExplorerEntity {
  id: string;
  // selected measure IDs to be shown
  selectedMeasureIds: Array<string>;
  // this is used to show leaderboard values
  leaderboardMeasureId: string;
  filters: MetricsViewRequestFilter;
  // user selected time range
  selectedTimeRange?: TimeSeriesTimeRange;
}

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
  sync(id: string, meta: MetricsViewMetaResponse) {
    if (!id || !meta || !meta.measures) return;
    updateMetricsExplorerById(
      id,
      (metricsExplorer) => {
        // sync measures with selected leaderboard measure.
        if (
          meta.measures.length &&
          (!metricsExplorer.leaderboardMeasureId ||
            !meta.measures.find(
              (measure) => measure.id === metricsExplorer.leaderboardMeasureId
            ))
        ) {
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
      metricsExplorer.filters.include = [];
      metricsExplorer.filters.exclude = [];
    });
  },

  clearFilterForDimension(id: string, dimensionId: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      removeIfExists(
        metricsExplorer.filters.include,
        (dimensionValues) => dimensionValues.name === dimensionId
      );
      removeIfExists(
        metricsExplorer.filters.exclude,
        (dimensionValues) => dimensionValues.name === dimensionId
      );
    });
  },
};
export const metricsExplorerStore: Readable<MetricsExplorerStoreType> &
  typeof metricViewReducers = {
  subscribe,

  ...metricViewReducers,
};
