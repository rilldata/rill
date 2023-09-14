import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getLocalUserPreferences } from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  getMapFromArray,
  removeIfExists,
} from "@rilldata/web-common/lib/arrayUtils";
import { getTimeComparisonParametersForComponent } from "@rilldata/web-common/lib/time/comparisons";
import type {
  ScrubRange,
  TimeRange,
} from "@rilldata/web-common/lib/time/types";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import { getDefaultTimeGrain } from "@rilldata/web-common/lib/time/grains";
import {
  convertTimeRangePreset,
  ISODurationToTimePreset,
} from "@rilldata/web-common/lib/time/ranges";
import type { TimeComparisonOption } from "@rilldata/web-common/lib/time/types";
import type {
  V1ColumnTimeRangeResponse,
  V1MetricsView,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import { derived, get, Readable, Writable, writable } from "svelte/store";
import { SortDirection, SortType } from "./proto-state/derived-types";

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
  selectedComparisonDimension?: string;

  // user selected timezone
  selectedTimezone?: string;

  // flag to show/hide time comparison based on user preference.
  // This controls whether a time comparison is shown in e.g.
  // the line charts and bignums.
  // It does NOT affect the leaderboard context column.
  showTimeComparison?: boolean;

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
        const fullTimeStart = new Date(fullTimeRange.timeRangeSummary.min);
        const fullTimeEnd = new Date(fullTimeRange.timeRangeSummary.max);
        const preset = ISODurationToTimePreset(
          metricsView.defaultTimeRange,
          true
        );

        const timeRange = convertTimeRangePreset(
          preset,
          fullTimeStart,
          fullTimeEnd,
          timeZone
        );
        const timeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end);
        timeSelections.selectedTimezone = timeZone;
        timeSelections.selectedTimeRange = {
          ...timeRange,
          interval: timeGrain.grain,
        };
        timeSelections.lastDefinedScrubRange = undefined;

        const comparisonOption = DEFAULT_TIME_RANGES[preset]
          ?.defaultComparison as TimeComparisonOption;
        if (comparisonOption) {
          const comparisonRange = getTimeComparisonParametersForComponent(
            comparisonOption,
            fullTimeStart,
            fullTimeEnd,
            timeRange.start,
            timeRange.end
          );
          if (comparisonRange.isComparisonRangeAvailable) {
            timeSelections.selectedComparisonTimeRange = {
              name: comparisonOption,
              start: comparisonRange.start,
              end: comparisonRange.end,
            };
            timeSelections.showTimeComparison = true;
            timeSelections.leaderboardContextColumn =
              LeaderboardContextColumn.DELTA_PERCENT;
          }
        }
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

        showTimeComparison: false,
        ...timeSelections,
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

  toggleSort(name: string, sortType?: SortType) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // if sortType is not provided,  or if it is provided
      // and is the same as the current sort type,
      // then just toggle the current sort direction
      if (
        sortType === undefined ||
        metricsExplorer.dashboardSortType === sortType
      ) {
        metricsExplorer.sortDirection =
          metricsExplorer.sortDirection === SortDirection.ASCENDING
            ? SortDirection.DESCENDING
            : SortDirection.ASCENDING;
      } else {
        // if the sortType is different from the current sort type,
        //  then update the sort type and set the sort direction
        // to descending
        metricsExplorer.dashboardSortType = sortType;
        metricsExplorer.sortDirection = SortDirection.DESCENDING;
      }
    });
  },

  setSortDirection(name: string, direction: SortDirection) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.sortDirection = direction;
    });
  },

  setSelectedTimeRange(name: string, timeRange: DashboardTimeControls) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      setSelectedScrubRange(metricsExplorer, undefined);
      metricsExplorer.selectedTimeRange = timeRange;
    });
  },

  setSelectedScrubRange(name: string, scrubRange: ScrubRange) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      setSelectedScrubRange(metricsExplorer, scrubRange);
    });
  },

  setMetricDimensionName(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedDimensionName = dimensionName;
    });
  },

  setComparisonDimension(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (dimensionName === undefined) {
        setDisplayComparison(metricsExplorer, true);
      } else {
        setDisplayComparison(metricsExplorer, false);
      }
      metricsExplorer.selectedComparisonDimension = dimensionName;
    });
  },

  disableAllComparisons(name: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.selectedComparisonDimension = undefined;
      setDisplayComparison(metricsExplorer, false);
    });
  },

  setSelectedComparisonRange(
    name: string,
    comparisonTimeRange: DashboardTimeControls
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      setDisplayComparison(metricsExplorer, true);
      metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;
    });
  },

  setTimeZone(name: string, zoneIANA: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // Reset scrub when timezone changes
      setSelectedScrubRange(metricsExplorer, undefined);

      metricsExplorer.selectedTimezone = zoneIANA;
    });
  },

  displayTimeComparison(name: string, showTimeComparison: boolean) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      setDisplayComparison(metricsExplorer, showTimeComparison);
    });
  },

  selectTimeRange(
    name: string,
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
    allTimeRange: TimeRange
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      // Reset scrub when range changes
      setSelectedScrubRange(metricsExplorer, undefined);

      metricsExplorer.selectedTimeRange = {
        ...timeRange,
        interval: timeGrain,
      };

      if (!comparisonTimeRange) {
        // when switching time range we reset comparison time range
        // get the default for the new time range and set it only if is valid
        const comparisonOption = DEFAULT_TIME_RANGES[timeRange.name]
          ?.defaultComparison as TimeComparisonOption;
        const range = getTimeComparisonParametersForComponent(
          comparisonOption,
          allTimeRange.start,
          allTimeRange.end,
          timeRange.start,
          timeRange.end
        );

        if (range.isComparisonRangeAvailable) {
          metricsExplorer.selectedComparisonTimeRange = {
            start: range.start,
            end: range.end,
            name: comparisonOption,
          };
        } else {
          metricsExplorer.selectedComparisonTimeRange = undefined;
        }
      } else {
        metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;
      }

      setDisplayComparison(
        metricsExplorer,
        metricsExplorer.selectedComparisonTimeRange !== undefined &&
          metricsExplorer.selectedComparisonDimension === undefined
      );
    });
  },

  setContextColumn(name: string, contextColumn: LeaderboardContextColumn) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      const initialSort = sortTypeForContextColumnType(
        metricsExplorer.leaderboardContextColumn
      );
      switch (contextColumn) {
        case LeaderboardContextColumn.DELTA_ABSOLUTE:
        case LeaderboardContextColumn.DELTA_PERCENT: {
          // if there is no time comparison, then we can't show
          // these context columns, so return with no change
          if (metricsExplorer.showTimeComparison === false) return;

          metricsExplorer.leaderboardContextColumn = contextColumn;
          break;
        }
        default:
          metricsExplorer.leaderboardContextColumn = contextColumn;
      }

      // if we have changed the context column, and the leaderboard is
      // sorted by the context column from before we made the change,
      // then we also need to change
      // the sort type to match the new context column
      if (metricsExplorer.dashboardSortType === initialSort) {
        metricsExplorer.dashboardSortType =
          sortTypeForContextColumnType(contextColumn);
      }
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

function setDisplayComparison(
  metricsExplorer: MetricsExplorerEntity,
  showTimeComparison: boolean
) {
  metricsExplorer.showTimeComparison = showTimeComparison;

  if (showTimeComparison) {
    metricsExplorer.selectedComparisonDimension = undefined;
  }

  // if setting showTimeComparison===true and not currently
  //  showing any context column, then show DELTA_PERCENT
  if (
    showTimeComparison &&
    metricsExplorer.leaderboardContextColumn === LeaderboardContextColumn.HIDDEN
  ) {
    metricsExplorer.leaderboardContextColumn =
      LeaderboardContextColumn.DELTA_PERCENT;
  }

  // if setting showTimeComparison===false and currently
  //  showing DELTA_PERCENT, then hide context column
  if (
    !showTimeComparison &&
    metricsExplorer.leaderboardContextColumn ===
      LeaderboardContextColumn.DELTA_PERCENT
  ) {
    metricsExplorer.leaderboardContextColumn = LeaderboardContextColumn.HIDDEN;
  }
}

function sortTypeForContextColumnType(
  contextCol: LeaderboardContextColumn
): SortType {
  const sortType = {
    [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
    [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
    [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
    [LeaderboardContextColumn.HIDDEN]: SortType.VALUE,
  }[contextCol];

  // Note: the above map needs to be EXHAUSTIVE over
  // LeaderboardContextColumn variants. If we ever add a new
  // context column type, we need to add it to the map above.
  // Otherwise, we will throw an error here.
  if (!sortType) {
    throw new Error(`Invalid context column type: ${contextCol}`);
  }
  return sortType;
}

function setSelectedScrubRange(
  metricsExplorer: MetricsExplorerEntity,
  scrubRange: ScrubRange
) {
  if (scrubRange === undefined) {
    metricsExplorer.lastDefinedScrubRange = undefined;
  } else if (!scrubRange.isScrubbing && scrubRange?.start && scrubRange?.end) {
    metricsExplorer.lastDefinedScrubRange = scrubRange;
  }

  metricsExplorer.selectedScrubRange = scrubRange;
}

export const projectShareStore: Writable<boolean> = writable(false);
