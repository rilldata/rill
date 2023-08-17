import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
import { getLocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  getMapFromArray,
  removeIfExists,
} from "@rilldata/web-common/lib/arrayUtils";
import { getComparionRangeForScrub } from "@rilldata/web-common/lib/time/comparisons";
import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
import { convertTimeRangePreset } from "@rilldata/web-common/lib/time/ranges";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  ScrubRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_LeaderboardSortDirection as SortDirection,
  DashboardState_LeaderboardSortType as SortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1ColumnTimeRangeResponse,
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { derived, get, Readable, Writable, writable } from "svelte/store";

export { SortType as SortType };
export { SortDirection as SortDirection };

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

  // This array controls which measures are visible in
  // explorer on the client. Note that this will need to be
  // updated to include all measure keys upon initialization
  // or else all measure will be hidden
  visibleMeasureKeys: Set<string>;
  // While the `visibleMeasureKeys` has the list of visible measures,
  // this is explicitly needed to fill the state.
  // TODO: clean this up when we refactor how url state is synced
  allMeasuresVisible: boolean;

  // This array controls which dimensions are visible in
  // explorer on the client.Note that if this is null, all
  // dimensions will be visible (this is needed to default to all visible
  // when there are not existing keys in the URL or saved on the
  // server)
  visibleDimensionKeys: Set<string>;
  // While the `visibleDimensionKeys` has the list of all visible dimensions,
  // this is explicitly needed to fill the state.
  // TODO: clean this up when we refactor how url state is synced
  allDimensionsVisible: boolean;

  // This is the name of the primary active measure in the dashboard.
  // This is the measure that will be shown in leaderboards, and
  // will be used for sorting the leaderboard and dimension
  // detail table.
  // This "name" is the internal name of the measure from the YAML,
  // not the human readable name.
  leaderboardMeasureName: string;

  // This is the sort type that will be used for the leaderboard
  // and dimension detail table. See SortType for more details.
  dashboardSortType: SortType;
  // This is the sort direction that will be used for the leaderboard
  // and dimension detail table.
  sortDirection: SortDirection;

  filters: V1MetricsViewFilter;
  // stores whether a dimension is in include/exclude filter mode
  // false/absence = include, true = exclude
  dimensionFilterExcludeMode: Map<string, boolean>;
  // user selected time range
  selectedTimeRange?: DashboardTimeControls;

  // user selected scrub range
  selectedScrubRange?: ScrubRange;
  lastDefinedScrubRange?: ScrubRange;

  selectedComparisonTimeRange?: DashboardTimeControls;

  // user selected timezone
  selectedTimezone?: string;

  // flag to show/hide time comparison based on user preference.
  // This controls whether a time comparison is shown in e.g.
  // the line charts and bignums.
  // It does NOT affect the leaderboard context column.
  showComparison?: boolean;

  // state of context column in the leaderboard
  leaderboardContextColumn: LeaderboardContextColumn;

  // user selected dimension
  selectedDimensionName?: string;

  proto?: string;
  // proto for the default set of selections
  defaultProto?: string;
}

export interface MetricsExplorerStoreType {
  entities: Record<string, MetricsExplorerEntity>;
}
const { update, subscribe } = writable({
  entities: {},
} as MetricsExplorerStoreType);

function updateMetricsExplorerProto(metricsExplorer: MetricsExplorerEntity) {
  metricsExplorer.proto = getProtoFromDashboardState(metricsExplorer);
}

export const updateMetricsExplorerByName = (
  name: string,
  callback: (metricsExplorer: MetricsExplorerEntity) => void
) => {
  update((state) => {
    if (!state.entities[name]) {
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

function syncMeasures(
  metricsView: V1MetricsView,
  metricsExplorer: MetricsExplorerEntity
) {
  const measuresMap = getMapFromArray(
    metricsView.measures,
    (measure) => measure.name
  );

  // sync measures with selected leaderboard measure.
  if (
    metricsView.measures.length &&
    (!metricsExplorer.leaderboardMeasureName ||
      !measuresMap.has(metricsExplorer.leaderboardMeasureName))
  ) {
    metricsExplorer.leaderboardMeasureName = metricsView.measures[0].name;
  } else if (!metricsView.measures.length) {
    metricsExplorer.leaderboardMeasureName = undefined;
  }
  // TODO: how does this differ from visibleMeasureKeys?
  metricsExplorer.selectedMeasureNames = metricsView.measures.map(
    (measure) => measure.name
  );

  if (metricsExplorer.allMeasuresVisible) {
    // this makes sure that the visible keys is in sync with list of measures
    metricsExplorer.visibleMeasureKeys = new Set(
      metricsView.measures.map((measure) => measure.name)
    );
  } else {
    // remove any keys from visible measure if it doesn't exist anymore
    for (const measureKey of metricsExplorer.visibleMeasureKeys) {
      if (!measuresMap.has(measureKey)) {
        metricsExplorer.visibleMeasureKeys.delete(measureKey);
      }
    }
    // If there are no visible measures, make the first measure visible
    if (
      metricsView.measures.length &&
      metricsExplorer.visibleMeasureKeys.size === 0
    ) {
      metricsExplorer.visibleMeasureKeys = new Set([
        metricsView.measures[0].name,
      ]);
    }

    // check if current leaderboard measure is visible,
    // if not set it to first visible measure
    if (
      metricsExplorer.visibleMeasureKeys.size &&
      !metricsExplorer.visibleMeasureKeys.has(
        metricsExplorer.leaderboardMeasureName
      )
    ) {
      const firstVisibleMeasure = metricsView.measures
        .map((measure) => measure.name)
        .find((key) => metricsExplorer.visibleMeasureKeys.has(key));
      metricsExplorer.leaderboardMeasureName = firstVisibleMeasure;
    }
  }
}

function syncDimensions(
  metricsView: V1MetricsView,
  metricsExplorer: MetricsExplorerEntity
) {
  // Having a map here improves the lookup for existing dimension name
  const dimensionsMap = getMapFromArray(
    metricsView.dimensions,
    (dimension) => dimension.name
  );
  metricsExplorer.filters.include = metricsExplorer.filters.include.filter(
    (filter) => dimensionsMap.has(filter.name)
  );
  metricsExplorer.filters.exclude = metricsExplorer.filters.exclude.filter(
    (filter) => dimensionsMap.has(filter.name)
  );

  if (
    metricsExplorer.selectedDimensionName &&
    !dimensionsMap.has(metricsExplorer.selectedDimensionName)
  ) {
    metricsExplorer.selectedDimensionName = undefined;
  }

  if (metricsExplorer.allDimensionsVisible) {
    // this makes sure that the visible keys is in sync with list of dimensions
    metricsExplorer.visibleDimensionKeys = new Set(
      metricsView.dimensions.map((dimension) => dimension.name)
    );
  } else {
    // remove any keys from visible dimension if it doesn't exist anymore
    for (const dimensionKey of metricsExplorer.visibleDimensionKeys) {
      if (!dimensionsMap.has(dimensionKey)) {
        metricsExplorer.visibleDimensionKeys.delete(dimensionKey);
      }
    }
  }
}

const metricViewReducers = {
  init(
    name: string,
    metricsView: V1MetricsView,
    fullTimeRange: V1ColumnTimeRangeResponse | undefined
  ) {
    update((state) => {
      if (state.entities[name]) return state;

      const timeSelections: Partial<MetricsExplorerEntity> = {};
      if (fullTimeRange) {
        const timeZone = get(getLocalUserPreferences()).timeZone;
        const timeRange = convertTimeRangePreset(
          TimeRangePreset.ALL_TIME,
          new Date(fullTimeRange.timeRangeSummary.min),
          new Date(fullTimeRange.timeRangeSummary.max),
          timeZone
        );
        const timeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end);
        timeSelections.selectedTimezone = timeZone;
        timeSelections.selectedTimeRange = {
          ...timeRange,
          interval: timeGrain.grain,
        };
      }

      state.entities[name] = {
        name,
        selectedMeasureNames: metricsView.measures.map(
          (measure) => measure.name
        ),

        visibleMeasureKeys: new Set(
          metricsView.measures.map((measure) => measure.name)
        ),
        allMeasuresVisible: true,
        visibleDimensionKeys: new Set(
          metricsView.dimensions.map((dim) => dim.name)
        ),
        allDimensionsVisible: true,
        leaderboardMeasureName: metricsView.measures[0]?.name,
        filters: {
          include: [],
          exclude: [],
        },
        dimensionFilterExcludeMode: new Map(),
        leaderboardContextColumn: LeaderboardContextColumn.HIDDEN,
        dashboardSortType: SortType.VALUE,
        sortDirection: SortDirection.DESCENDING,

        ...timeSelections,
        showComparison: false,
      };

      updateMetricsExplorerProto(state.entities[name]);
      state.entities[name].defaultProto = state.entities[name].proto;
      return state;
    });
  },

  syncFromUrl(name: string, urlState: string, metricsView: V1MetricsView) {
    if (!urlState || !metricsView) return;
    // not all data for MetricsExplorerEntity will be filled out here.
    // Hence, it is a Partial<MetricsExplorerEntity>
    const partial = getDashboardStateFromUrl(urlState, metricsView);
    if (!partial) return;

    updateMetricsExplorerByName(name, (metricsExplorer) => {
      for (const key in partial) {
        metricsExplorer[key] = partial[key];
      }
      metricsExplorer.dimensionFilterExcludeMode =
        includeExcludeModeFromFilters(partial.filters);
    });
  },

  sync(name: string, metricsView: V1MetricsView) {
    if (!name || !metricsView || !metricsView.measures) return;
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // remove references to non existent measures
      syncMeasures(metricsView, metricsExplorer);

      // remove references to non existent dimensions
      syncDimensions(metricsView, metricsExplorer);
    });
  },

  setLeaderboardMeasureName(name: string, measureName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureName = measureName;
    });
  },

  setSortDescending(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.sortDirection = SortDirection.DESCENDING;
    });
  },

  setSortAscending(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.sortDirection = SortDirection.ASCENDING;
    });
  },

  toggleSortDirection(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.sortDirection =
        metricsExplorer.sortDirection === SortDirection.ASCENDING
          ? SortDirection.DESCENDING
          : SortDirection.ASCENDING;
    });
  },

  setSortDirection(name: string, direction: SortDirection) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.sortDirection = direction;
    });
  },

  setSelectedTimeRange(name: string, timeRange: DashboardTimeControls) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  setSelectedScrubRange(name: string, scrubRange: ScrubRange) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (scrubRange === undefined) {
        metricsExplorer.lastDefinedScrubRange = undefined;
      } else if (
        !scrubRange.isScrubbing &&
        scrubRange?.start &&
        scrubRange?.end
      ) {
        metricsExplorer.lastDefinedScrubRange = scrubRange;
      }

      metricsExplorer.selectedScrubRange = scrubRange;
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

  setTimeZone(name: string, zoneIANA: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedTimezone = zoneIANA;
    });
  },

  displayComparison(name: string, showComparison: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.showComparison = showComparison;
      // if setting showComparison===true and not currently
      //  showing any context column, then show DELTA_CHANGE
      if (
        showComparison &&
        metricsExplorer.leaderboardContextColumn ===
          LeaderboardContextColumn.HIDDEN
      ) {
        metricsExplorer.leaderboardContextColumn =
          LeaderboardContextColumn.DELTA_CHANGE;
      }

      // if setting showComparison===false and currently
      //  showing DELTA_CHANGE, then hide context column
      if (
        !showComparison &&
        metricsExplorer.leaderboardContextColumn ===
          LeaderboardContextColumn.DELTA_CHANGE
      ) {
        metricsExplorer.leaderboardContextColumn =
          LeaderboardContextColumn.HIDDEN;
      }
    });
  },

  displayDeltaChange(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // NOTE: only show delta change if comparison is enabled
      if (metricsExplorer.showComparison === false) return;

      metricsExplorer.leaderboardContextColumn =
        LeaderboardContextColumn.DELTA_CHANGE;
    });
  },

  displayPercentOfTotal(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardContextColumn =
        LeaderboardContextColumn.PERCENT;
    });
  },

  hideContextColumn(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardContextColumn =
        LeaderboardContextColumn.HIDDEN;
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

  remove(name: string) {
    update((state) => {
      delete state.entities[name];
      return state;
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

/***
 * Dervied stores to get time range and comparison range to be
 * used for fetching data. If we have a scrub range and
 * isScrubbing is false, use that, otherwise use the selected
 * time range
 */

export function useFetchTimeRange(name: string) {
  return derived(metricsExplorerStore, ($store) => {
    const entity = $store.entities[name];
    if (entity?.lastDefinedScrubRange) {
      // Use last scrub range before scrubbing started
      const { start, end } = getOrderedStartEnd(
        entity.lastDefinedScrubRange?.start,
        entity.lastDefinedScrubRange?.end
      );

      return { start, end };
    } else {
      return {
        start: entity.selectedTimeRange?.start,
        end: entity.selectedTimeRange?.end,
      };
    }
  });
}

export function useComparisonRange(name: string) {
  return derived(metricsExplorerStore, ($store) => {
    const entity = $store.entities[name];

    if (
      !entity?.showComparison ||
      !entity.selectedComparisonTimeRange?.start ||
      !entity.selectedComparisonTimeRange?.end
    ) {
      return {
        start: undefined,
        end: undefined,
      };
    } else if (entity?.lastDefinedScrubRange) {
      const { start, end } = getOrderedStartEnd(
        entity.lastDefinedScrubRange?.start,
        entity.lastDefinedScrubRange?.end
      );

      const comparisonRange = getComparionRangeForScrub(
        entity.selectedTimeRange?.start,
        entity.selectedTimeRange?.end,
        entity.selectedComparisonTimeRange?.start,
        entity.selectedComparisonTimeRange?.end,
        start,
        end
      );

      return {
        start: comparisonRange?.start?.toISOString(),
        end: comparisonRange?.end?.toISOString(),
      };
    } else {
      return {
        start: entity.selectedComparisonTimeRange?.start?.toISOString(),
        end: entity.selectedComparisonTimeRange?.end?.toISOString(),
      };
    }
  });
}

export const projectShareStore: Writable<boolean> = writable(false);
