import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getWhereFilterExpressionIndex } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  createAndExpression,
  filterExpressions,
  forEachExpression,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type {
  DashboardTimeControls,
  ScrubRange,
  TimeRange,
} from "@rilldata/web-common/lib/time/types";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import type {
  V1Expression,
  V1MetricsView,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { Readable, derived, writable } from "svelte/store";
import {
  SortDirection,
  SortType,
} from "web-common/src/features/dashboards/proto-state/derived-types";

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
  callback: (metricsExplorer: MetricsExplorerEntity) => void,
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

function includeExcludeModeFromFilters(filters: V1Expression | undefined) {
  const map = new Map<string, boolean>();
  if (!filters) return map;
  forEachIdentifier(filters, (e, ident) => {
    if (e.cond?.op === V1Operation.OPERATION_NIN) {
      map.set(ident, true);
    }
  });
  return map;
}

function syncMeasures(
  metricsView: V1MetricsView,
  metricsExplorer: MetricsExplorerEntity,
) {
  const measuresMap = getMapFromArray(
    metricsView.measures,
    (measure) => measure.name,
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

  if (metricsExplorer.allMeasuresVisible) {
    // this makes sure that the visible keys is in sync with list of measures
    metricsExplorer.visibleMeasureKeys = new Set(
      metricsView.measures.map((measure) => measure.name),
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
        metricsExplorer.leaderboardMeasureName,
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
  metricsExplorer: MetricsExplorerEntity,
) {
  // Having a map here improves the lookup for existing dimension name
  const dimensionsMap = getMapFromArray(
    metricsView.dimensions,
    (dimension) => dimension.name,
  );
  metricsExplorer.whereFilter =
    filterExpressions(metricsExplorer.whereFilter, (e) => {
      if (!e.cond?.exprs?.length) return true;
      return dimensionsMap.has(e.cond.exprs[0].ident);
    }) ?? createAndExpression([]);

  if (
    metricsExplorer.selectedDimensionName &&
    !dimensionsMap.has(metricsExplorer.selectedDimensionName)
  ) {
    metricsExplorer.selectedDimensionName = undefined;
  }

  if (metricsExplorer.allDimensionsVisible) {
    // this makes sure that the visible keys is in sync with list of dimensions
    metricsExplorer.visibleDimensionKeys = new Set(
      metricsView.dimensions.map((dimension) => dimension.name),
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
    metricsView: V1MetricsViewSpec,
    fullTimeRange: V1MetricsViewTimeRangeResponse | undefined,
  ) {
    update((state) => {
      if (state.entities[name]) return state;

      state.entities[name] = getDefaultMetricsExplorerEntity(
        name,
        metricsView,
        fullTimeRange,
      );

      updateMetricsExplorerProto(state.entities[name]);
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
      // this hack is needed since what is shown for comparison is not a single source
      // TODO: use an enum and get rid of this
      if (!partial.showTimeComparison) {
        metricsExplorer.showTimeComparison = false;
      }
      metricsExplorer.dimensionFilterExcludeMode =
        includeExcludeModeFromFilters(partial.whereFilter);
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

  /**
   * DEPRECATED!!!
   * use setLeaderboardMeasureName via:
   * getStateManagers().actions.setLeaderboardMeasureName
   *
   * Still used in tests, so we can't remove it yet, but don't use
   * it in production code.
   */
  setLeaderboardMeasureName(name: string, measureName: string) {
    console.warn(
      "setLeaderboardMeasureName is deprecated. Use setLeaderboardMeasureName via `getStateManagers().actions.setLeaderboardMeasureName`. Still used in tests, so we can't remove it yet, but don't use it in production code.",
    );
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.leaderboardMeasureName = measureName;
    });
  },

  setExpandedMeasureName(name: string, measureName: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.expandedMeasureName = measureName;

      // If going into TDD view and already having a comparison dimension,
      // then set the pinIndex
      if (metricsExplorer.selectedComparisonDimension) {
        metricsExplorer.pinIndex = getPinIndexForDimension(
          metricsExplorer,
          metricsExplorer.selectedComparisonDimension,
        );
      }
    });
  },

  setPinIndex(name: string, index: number) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.pinIndex = index;
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

  setMetricDimensionName(name: string, dimensionName: string | null) {
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
      metricsExplorer.pinIndex = getPinIndexForDimension(
        metricsExplorer,
        dimensionName,
      );
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
    comparisonTimeRange: DashboardTimeControls,
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

  setSearchText(name: string, searchText: string) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      metricsExplorer.dimensionSearchText = searchText;
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
  ) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      if (!timeRange.name) return;

      // Reset scrub when range changes
      setSelectedScrubRange(metricsExplorer, undefined);

      metricsExplorer.selectedTimeRange = {
        ...timeRange,
        interval: timeGrain,
      };

      metricsExplorer.selectedComparisonTimeRange = comparisonTimeRange;

      setDisplayComparison(
        metricsExplorer,
        metricsExplorer.selectedComparisonTimeRange !== undefined &&
          metricsExplorer.selectedComparisonDimension === undefined,
      );
    });
  },

  setContextColumn(name: string, contextColumn: LeaderboardContextColumn) {
    updateMetricsExplorerByName(name, (metricsExplorer) => {
      const initialSort = sortTypeForContextColumnType(
        metricsExplorer.leaderboardContextColumn,
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
  name: string,
): Readable<MetricsExplorerEntity> {
  return derived(metricsExplorerStore, ($store) => {
    return $store.entities[name];
  });
}

export function setDisplayComparison(
  metricsExplorer: MetricsExplorerEntity,
  showTimeComparison: boolean,
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

export function sortTypeForContextColumnType(
  contextCol: LeaderboardContextColumn,
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
  scrubRange: ScrubRange,
) {
  if (scrubRange === undefined) {
    metricsExplorer.lastDefinedScrubRange = undefined;
  } else if (!scrubRange.isScrubbing && scrubRange?.start && scrubRange?.end) {
    metricsExplorer.lastDefinedScrubRange = scrubRange;
  }

  metricsExplorer.selectedScrubRange = scrubRange;
}

function getPinIndexForDimension(
  metricsExplorer: MetricsExplorerEntity,
  dimensionName: string,
) {
  const dimensionEntryIndex = getWhereFilterExpressionIndex({
    dashboard: metricsExplorer,
  })(dimensionName);
  if (dimensionEntryIndex === undefined || dimensionEntryIndex === -1)
    return -1;

  const dimExpr =
    metricsExplorer.whereFilter.cond?.exprs?.[dimensionEntryIndex];
  if (!dimExpr?.cond?.exprs?.length) return -1;

  // 1st entry in the expression is the identifier. hence the -2 here.
  return dimExpr.cond.exprs.length - 2;
}
