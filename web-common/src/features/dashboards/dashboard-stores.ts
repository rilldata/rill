import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type {
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { removeIfExists } from "@rilldata/web-local/lib/util/arrayUtils";
import { Readable, Writable, derived, writable } from "svelte/store";

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

  /** 
  FIXME For now we are using the user supplied `expression` for measures
  and `name` (column name) for dimensions to determine which measures and
  dimensions are visible. These are used because they are the only fields
  that are required to exist in measures/dimensions for them to be shown
  in the dashboard

  This may lead to problems if there are ever duplicates among
  these. Hamilton has started discussions with Benjamin about
  adding unique keys that could be used to replace these temporary keys. 
  Once those become available the logic around the fields below
  should be updated.
*/
  // This array controls which measures are visible in
  // explorer on the client. Note that this will need to be
  // updated to include all measure keys upon initialization
  // or else all measure will be hidden
  visibleMeasureKeys: Set<string>;
  // This array controls which dimensions are visible in
  // explorer on the client.Note that if this is null, all
  // dimensions will be visible (this is needed to default to all visible
  // when there are not existing keys in the URL or saved on the
  // server)
  visibleDimensionKeys: Set<string>;

  // this is used to show leaderboard values
  leaderboardMeasureName: string;
  filters: V1MetricsViewFilter;
  // stores whether a dimension is in include/exclude filter mode
  // false/absence = include, true = exclude
  dimensionFilterExcludeMode: Map<string, boolean>;
  // user selected time range
  selectedTimeRange?: DashboardTimeControls;
  selectedComparisonTimeRange?: DashboardTimeControls;
  // flag to show/hide comparison based on user preference
  showComparison?: boolean;
  // user selected dimension
  selectedDimensionName?: string;

  proto?: string;
  // proto for the default set of selections
  defaultProto?: string;
  // marks that defaults have been selected
  // TODO: move default selection to a common place and avoid this
  defaultsSelected?: boolean;
}

export interface MetricsExplorerStoreType {
  entities: Record<string, MetricsExplorerEntity>;
}
const { update, subscribe } = writable({
  entities: {},
} as MetricsExplorerStoreType);

function updateMetricsExplorerProto(metricsExplorer: MetricsExplorerEntity) {
  metricsExplorer.proto = getProtoFromDashboardState(metricsExplorer);
  if (!metricsExplorer.defaultsSelected) {
    metricsExplorer.defaultProto = metricsExplorer.proto;
  }
}

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
      if (state.entities[name]) {
        updateMetricsExplorerProto(state.entities[name]);
      }
      return state;
    }

    callback(state.entities[name]);
    // every change triggers a proto update
    updateMetricsExplorerProto(state.entities[name]);
    return state;
  });
};

function includeExcludeModeFromFilters(filters: V1MetricsViewFilter) {
  const map = new Map<string, boolean>();
  filters?.exclude.forEach((cond) => map.set(cond.name, true));
  return map;
}

const metricViewReducers = {
  syncFromUrl(name: string, url: URL) {
    // not all data for MetricsExplorerEntity will be filled out here.
    // Hence, it is a Partial<MetricsExplorerEntity>
    const partial = getDashboardStateFromUrl(url);
    if (!partial) return;

    updateMetricsExplorerByName(
      name,
      (metricsExplorer) => {
        for (const key in partial) {
          metricsExplorer[key] = partial[key];
        }
        metricsExplorer.dimensionFilterExcludeMode =
          includeExcludeModeFromFilters(partial.filters);
        metricsExplorer.defaultsSelected = true;
      },
      () => ({
        name,
        selectedMeasureNames: [],
        visibleMeasureKeys: new Set(),
        visibleDimensionKeys: new Set(),
        leaderboardMeasureName: "",
        filters: {},
        dimensionFilterExcludeMode: includeExcludeModeFromFilters(
          partial.filters
        ),
        defaultsSelected: true,
        ...partial,
      })
    );
  },

  sync(name: string, metricsView: V1MetricsView) {
    if (!name || !metricsView || !metricsView.measures) return;
    updateMetricsExplorerByName(
      name,
      (metricsExplorer) => {
        // sync measures with selected leaderboard measure.
        if (
          metricsView.measures.length &&
          (!metricsExplorer.leaderboardMeasureName ||
            !metricsView.measures.find(
              (measure) =>
                measure.name === metricsExplorer.leaderboardMeasureName
            ))
        ) {
          metricsExplorer.leaderboardMeasureName = metricsView.measures[0].name;
        } else if (!metricsView.measures.length) {
          metricsExplorer.leaderboardMeasureName = undefined;
        }
        metricsExplorer.selectedMeasureNames = metricsView.measures.map(
          (measure) => measure.name
        );

        metricsExplorer.visibleMeasureKeys = new Set(
          metricsView.measures.map((measure) => measure.expression)
        );

        metricsExplorer.visibleDimensionKeys = new Set(
          metricsView.dimensions.map((dim) => dim.name)
        );
      },
      () => ({
        name,
        selectedMeasureNames: metricsView.measures.map(
          (measure) => measure.name
        ),

        visibleMeasureKeys: new Set(
          metricsView.measures.map((measure) => measure.expression)
        ),
        visibleDimensionKeys: new Set(
          metricsView.dimensions.map((dim) => dim.name)
        ),
        leaderboardMeasureName: metricsView.measures[0]?.name,
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

  toggleMeasureVisibilityByKey(name: string, key: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (metricsExplorer.visibleMeasureKeys.has(key)) {
        metricsExplorer.visibleMeasureKeys.delete(key);
      } else {
        metricsExplorer.visibleMeasureKeys.add(key);
      }
    });
  },

  hideAllMeasures(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.visibleMeasureKeys.clear();
    });
  },

  setMultipleMeasuresVisible(name: string, keys: string[]) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.visibleMeasureKeys = new Set([
        ...metricsExplorer.visibleMeasureKeys,
        ...keys,
      ]);
    });
  },

  toggleDimensionVisibilityByKey(name: string, key: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (metricsExplorer.visibleDimensionKeys.has(key)) {
        metricsExplorer.visibleDimensionKeys.delete(key);
      } else {
        metricsExplorer.visibleDimensionKeys.add(key);
      }
    });
  },

  hideAllDimensions(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.visibleDimensionKeys.clear();
    });
  },

  setMultipleDimensionsVisible(name: string, keys: string[]) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.visibleDimensionKeys = new Set([
        ...metricsExplorer.visibleDimensionKeys,
        ...keys,
      ]);
    });
  },

  clearLeaderboardMeasureName(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureName = undefined;
    });
  },

  setSelectedTimeRange(name: string, timeRange: DashboardTimeControls) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  setMetricDimensionName(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedDimensionName = dimensionName;
    });
  },

  setSelectedComparisonRange(
    name: string,
    comparisonTimeRange: DashboardTimeControls
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;
    });
  },

  toggleComparison(name: string, showComparison: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.showComparison = showComparison;
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

  allDefaultsSelected(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.defaultsSelected = true;
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
  return derived(metricsExplorerStore, ($store) => {
    return $store.entities[name];
  });
}

export const projectShareStore: Writable<boolean> = writable(false);
