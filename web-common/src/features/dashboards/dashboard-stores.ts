import type {
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { removeIfExists } from "@rilldata/web-local/lib/util/arrayUtils";
import { derived, Readable, Writable, writable } from "svelte/store";
import type {
  TimeSeriesTimeRange,
  ComparisonWithTimeRange,
} from "./time-controls/time-control-types";

export interface LeaderboardValue {
  value: number;
  label: string;
}

export interface LeaderboardValues {
  values: Array<LeaderboardValue>;
  dimensionName: string;
}

export type ActiveValues = Record<string, Array<[unknown, boolean]>>;

export interface MetricsExplorerEntity {
  name: string;
  // selected measure names to be shown
  selectedMeasureNames: Array<string>;
  // this is used to show leaderboard values
  leaderboardMeasureName: string;
  filters: V1MetricsViewFilter;
  // stores whether a dimension is in include/exclude filter mode
  // false/absence = include, true = exclude
  dimensionFilterExcludeMode: Map<string, boolean>;
  // user selected time range
  selectedTimeRange?: TimeSeriesTimeRange;
  // user selected dimension
  selectedDimensionName?: string;
  // comparison time range
  selectedComparisonTimeRange?: ComparisonWithTimeRange;
  // flag to show/hide comparison based on time range
  showComparison?: boolean;
}

export interface MetricsExplorerStoreType {
  entities: Record<string, MetricsExplorerEntity>;
}
const { update, subscribe } = writable({
  entities: {},
} as MetricsExplorerStoreType);

const updateMetricsExplorerByName = (
  name: string,
  callback: (metricsExplorer: MetricsExplorerEntity) => void,
  absenceCallback?: () => MetricsExplorerEntity
) => {
  update((state) => {
    if (!state.entities[name]) {
      if (absenceCallback) {
        state.entities[name] = absenceCallback();
      }
      return state;
    }
    callback(state.entities[name]);
    return state;
  });
};

const metricViewReducers = {
  create(
    name: string,
    filters: V1MetricsViewFilter,
    selectedTimeRange: TimeSeriesTimeRange
  ) {
    updateMetricsExplorerByName(
      name,
      (metricsExplorer) => {
        metricsExplorer.filters = filters;
        metricsExplorer.selectedTimeRange = selectedTimeRange;
      },
      () => ({
        name,
        selectedMeasureNames: [],
        leaderboardMeasureName: "",
        filters,
        dimensionFilterExcludeMode: new Map(),
        selectedTimeRange,
      })
    );
  },

  sync(name: string, meta: V1MetricsView) {
    if (!name || !meta || !meta.measures) return;
    updateMetricsExplorerByName(
      name,
      (metricsExplorer) => {
        // sync measures with selected leaderboard measure.
        if (
          meta.measures.length &&
          (!metricsExplorer.leaderboardMeasureName ||
            !meta.measures.find(
              (measure) =>
                measure.name === metricsExplorer.leaderboardMeasureName
            ))
        ) {
          metricsExplorer.leaderboardMeasureName = meta.measures[0].name;
        } else if (!meta.measures.length) {
          metricsExplorer.leaderboardMeasureName = undefined;
        }
        metricsExplorer.selectedMeasureNames = meta.measures.map(
          (measure) => measure.name
        );
      },
      () => ({
        name,
        selectedMeasureNames: meta.measures.map((measure) => measure.name),
        leaderboardMeasureName: meta.measures[0]?.name,
        filters: {
          include: [],
          exclude: [],
        },
        dimensionFilterExcludeMode: new Map(),
      })
    );
  },

  setLeaderboardMeasureName(name: string, measureName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureName = measureName;
    });
  },

  clearLeaderboardMeasureName(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureName = undefined;
    });
  },

  setSelectedTimeRange(name: string, timeRange: TimeSeriesTimeRange) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  setMetricDimensionName(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedDimensionName = dimensionName;
    });
  },

  setShowComparison(name: string, flag: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.showComparison = flag;
    });
  },

  setSelectedComparisonRange(
    name: string,
    comparisonTimeRange: ComparisonWithTimeRange
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;
    });
  },

  toggleFilter(name: string, dimensionName: string, dimensionValue: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      const relevantFilterKey = metricsExplorer.dimensionFilterExcludeMode.get(
        dimensionName
      )
        ? "exclude"
        : "include";

      const dimensionEntryIndex = metricsExplorer.filters[
        relevantFilterKey
      ].findIndex((filter) => filter.name === dimensionName);

      if (dimensionEntryIndex >= 0) {
        if (
          removeIfExists(
            metricsExplorer.filters[relevantFilterKey][dimensionEntryIndex].in,
            (value) => value === dimensionValue
          )
        ) {
          if (
            metricsExplorer.filters[relevantFilterKey][dimensionEntryIndex].in
              .length === 0
          ) {
            metricsExplorer.filters[relevantFilterKey].splice(
              dimensionEntryIndex,
              1
            );
          }
          return;
        }

        metricsExplorer.filters[relevantFilterKey][dimensionEntryIndex].in.push(
          dimensionValue
        );
      } else {
        metricsExplorer.filters[relevantFilterKey].push({
          name: dimensionName,
          in: [dimensionValue],
        });
      }
    });
  },

  clearFilters(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.filters.include = [];
      metricsExplorer.filters.exclude = [];
      metricsExplorer.dimensionFilterExcludeMode.clear();
    });
  },

  clearFilterForDimension(
    name: string,
    dimensionName: string,
    include: boolean
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (include) {
        removeIfExists(
          metricsExplorer.filters.include,
          (dimensionValues) => dimensionValues.name === dimensionName
        );
      } else {
        removeIfExists(
          metricsExplorer.filters.exclude,
          (dimensionValues) => dimensionValues.name === dimensionName
        );
      }
    });
  },

  /**
   * Toggle a dimension filter between include/exclude modes
   */
  toggleFilterMode(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      const exclude =
        metricsExplorer.dimensionFilterExcludeMode.get(dimensionName);
      metricsExplorer.dimensionFilterExcludeMode.set(dimensionName, !exclude);

      const relevantFilterKey = exclude ? "exclude" : "include";
      const otherFilterKey = exclude ? "include" : "exclude";

      const otherFilterEntryIndex = metricsExplorer.filters[
        relevantFilterKey
      ].findIndex((filter) => filter.name === dimensionName);
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
};
export const metricsExplorerStore: Readable<MetricsExplorerStoreType> &
  typeof metricViewReducers = {
  subscribe,

  ...metricViewReducers,
};

export function useDashboardStore(
  name: string
): Readable<MetricsExplorerEntity> {
  const derivedStore = derived(metricsExplorerStore, ($store) => {
    return $store.entities[name];
  });
  return derivedStore;
}

export const calendlyModalStore: Writable<string> = writable("");
