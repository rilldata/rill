import type { TimeSeriesTimeRange } from "@rilldata/web-local/common/database-service/DatabaseTimeSeriesActions";
import type {
  MetricsViewMetaResponse,
  MetricsViewRequestFilter,
} from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";
import { removeIfExists } from "@rilldata/web-local/common/utils/arrayUtils";
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
  // stores whether a dimension is in include/exclude filter mode
  // false/absence = include, true = exclude
  dimensionFilterExcludeMode: Map<string, boolean>;
  // user selected time range
  selectedTimeRange?: TimeSeriesTimeRange;
  // user selected dimension
  selectedDimensionId?: string;
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
        dimensionFilterExcludeMode: new Map(),
      })
    );
  },

  setLeaderboardMeasureId(id: string, measureId: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureId = measureId;
    });
  },

  clearLeaderboardMeasureId(id: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureId = undefined;
    });
  },

  setSelectedTimeRange(id: string, timeRange: TimeSeriesTimeRange) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  setMetricDimensionId(id: string, dimensionId: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.selectedDimensionId = dimensionId;
    });
  },

  toggleFilter(id: string, dimensionId: string, dimensionValue: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      const relevantFilterKey = metricsExplorer.dimensionFilterExcludeMode.get(
        dimensionId
      )
        ? "exclude"
        : "include";

      const dimensionEntryIndex = metricsExplorer.filters[
        relevantFilterKey
      ].findIndex((filter) => filter.name === dimensionId);

      if (dimensionEntryIndex >= 0) {
        if (
          removeIfExists(
            metricsExplorer.filters[relevantFilterKey][dimensionEntryIndex]
              .values,
            (value) => value === dimensionValue
          )
        ) {
          if (
            metricsExplorer.filters[relevantFilterKey][dimensionEntryIndex]
              .values.length === 0
          ) {
            metricsExplorer.filters[relevantFilterKey].splice(
              dimensionEntryIndex,
              1
            );
          }
          return;
        }

        metricsExplorer.filters[relevantFilterKey][
          dimensionEntryIndex
        ].values.push(dimensionValue);
      } else {
        metricsExplorer.filters[relevantFilterKey].push({
          name: dimensionId,
          values: [dimensionValue],
        });
      }
    });
  },

  /**
   * Toggle a dimension filter between include/exclude modes
   */
  toggleFilterExcludeMode(id: string, dimensionId: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      const exclude =
        metricsExplorer.dimensionFilterExcludeMode.get(dimensionId);
      metricsExplorer.dimensionFilterExcludeMode.set(dimensionId, !exclude);

      const relevantFilterKey = exclude ? "exclude" : "include";
      const otherFilterKey = exclude ? "include" : "exclude";

      const otherFilterEntryIndex = metricsExplorer.filters[
        relevantFilterKey
      ].findIndex((filter) => filter.name === dimensionId);
      // if relevant filter is not present then return
      if (otherFilterEntryIndex === -1) return;

      // push relevant filters to other filter
      metricsExplorer.filters[otherFilterKey].push(
        metricsExplorer.filters[relevantFilterKey][otherFilterEntryIndex]
      );
      // remove entry from relevant filter
      metricsExplorer.filters[relevantFilterKey].splice(
        otherFilterEntryIndex,
        1
      );
    });
  },

  clearFilters(id: string) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      metricsExplorer.filters.include = [];
      metricsExplorer.filters.exclude = [];
      metricsExplorer.dimensionFilterExcludeMode.clear();
    });
  },

  clearFilterForDimension(id: string, dimensionId: string, include: boolean) {
    updateMetricsExplorerById(id, (metricsExplorer) => {
      if (include) {
        removeIfExists(
          metricsExplorer.filters.include,
          (dimensionValues) => dimensionValues.name === dimensionId
        );
      } else {
        removeIfExists(
          metricsExplorer.filters.exclude,
          (dimensionValues) => dimensionValues.name === dimensionId
        );
      }
    });
  },
};
export const metricsExplorerStore: Readable<MetricsExplorerStoreType> &
  typeof metricViewReducers = {
  subscribe,

  ...metricViewReducers,
};
