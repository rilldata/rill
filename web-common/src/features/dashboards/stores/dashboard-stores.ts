import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getWhereFilterExpressionIndex } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import { AdvancedMeasureCorrector } from "@rilldata/web-common/features/dashboards/stores/AdvancedMeasureCorrector";
import {
  createAndExpression,
  filterExpressions,
  forEachIdentifier,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { type ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  TimeRangePreset,
  type DashboardTimeControls,
  type ScrubRange,
  type TimeRange,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1ExploreSpec,
  V1Expression,
  V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import type { ExpandedState, SortingState } from "@tanstack/svelte-table";
import { derived, writable, type Readable } from "svelte/store";
import { SortType } from "web-common/src/features/dashboards/proto-state/derived-types";
import {
  PivotChipType,
  type PivotChipData,
  type PivotTableMode,
} from "../pivot/types";

export interface MetricsExplorerStoreType {
  entities: Record<string, ExploreState>;
}
const { update, subscribe } = writable({
  entities: {},
} as MetricsExplorerStoreType);

export const updateMetricsExplorerByName = (
  name: string,
  callback: (exploreState: ExploreState) => void,
) => {
  update((state) => {
    if (!state.entities[name]) {
      return state;
    }

    callback(state.entities[name]);
    return state;
  });
};

export function includeExcludeModeFromFilters(
  filters: V1Expression | undefined,
) {
  const map = new Map<string, boolean>();
  if (!filters) return map;
  forEachIdentifier(filters, (e, ident) => {
    if (
      e.cond?.op === V1Operation.OPERATION_NIN ||
      e.cond?.op === V1Operation.OPERATION_NLIKE ||
      e.cond?.op === V1Operation.OPERATION_NEQ
    ) {
      map.set(ident, true);
    }
  });
  return map;
}

function syncMeasures(explore: V1ExploreSpec, exploreState: ExploreState) {
  const measuresSet = new Set(explore.measures ?? []);

  // sync measures with selected leaderboard measure and ensure default measure is set
  if (explore.measures?.length) {
    const defaultMeasure = explore.measures[0];
    if (!measuresSet.has(exploreState.leaderboardSortByMeasureName)) {
      exploreState.leaderboardSortByMeasureName = defaultMeasure;
    }
    if (!exploreState.leaderboardMeasureNames?.length) {
      exploreState.leaderboardMeasureNames = [defaultMeasure];
    }
  }

  if (
    exploreState.tdd.expandedMeasureName &&
    !measuresSet.has(exploreState.tdd.expandedMeasureName)
  ) {
    exploreState.tdd.expandedMeasureName = undefined;
  }

  exploreState.pivot.columns = exploreState.pivot.columns.filter((measure) =>
    measuresSet.has(measure.id),
  );

  if (exploreState.allMeasuresVisible) {
    // this makes sure that the visible keys is in sync with list of measures
    exploreState.visibleMeasures = [...measuresSet];
  } else {
    // remove any visible measures that doesn't exist anymore
    exploreState.visibleMeasures = exploreState.visibleMeasures.filter((m) =>
      measuresSet.has(m),
    );
    // If there are no visible measures, make the first measure visible
    if (explore.measures?.length && exploreState.visibleMeasures.length === 0) {
      exploreState.visibleMeasures = [explore.measures[0]];
    }
  }
}

function syncDimensions(explore: V1ExploreSpec, exploreState: ExploreState) {
  // Having a map here improves the lookup for existing dimension name
  const dimensionsSet = new Set(explore.dimensions ?? []);
  exploreState.whereFilter =
    filterExpressions(exploreState.whereFilter, (e) => {
      if (!e.cond?.exprs?.length) return true;
      return dimensionsSet.has(e.cond.exprs[0].ident!);
    }) ?? createAndExpression([]);

  if (
    exploreState.selectedDimensionName &&
    !dimensionsSet.has(exploreState.selectedDimensionName)
  ) {
    exploreState.selectedDimensionName = undefined;
    exploreState.activePage = DashboardState_ActivePage.DEFAULT;
  }

  exploreState.pivot.rows = exploreState.pivot.rows.filter(
    (dimension) =>
      dimensionsSet.has(dimension.id) || dimension.type === PivotChipType.Time,
  );

  exploreState.pivot.columns = exploreState.pivot.columns.filter(
    (dimension) =>
      dimensionsSet.has(dimension.id) || dimension.type === PivotChipType.Time,
  );

  if (exploreState.allDimensionsVisible) {
    // this makes sure that the visible keys is in sync with list of dimensions
    exploreState.visibleDimensions = [...dimensionsSet];
  } else {
    // remove any visible dimensions that doesn't exist anymore
    exploreState.visibleDimensions = exploreState.visibleDimensions
      ? exploreState.visibleDimensions.filter((d) => dimensionsSet.has(d))
      : [];
  }
}

const metricsViewReducers = {
  init(name: string, initState: ExploreState) {
    update((state) => {
      // TODO: revisit this during the url state / restore user refactor
      initState.dimensionFilterExcludeMode = includeExcludeModeFromFilters(
        initState.whereFilter,
      );
      state.entities[name] = structuredClone(initState);
      state.entities[name].name = name;

      return state;
    });
  },

  syncFromUrl(
    name: string,
    urlState: string,
    metricsView: V1MetricsViewSpec,
    explore: V1ExploreSpec,
  ) {
    if (!urlState || !metricsView) return;
    // not all data for MetricsExplorerEntity will be filled out here.
    // Hence, it is a Partial<MetricsExplorerEntity>
    const partial = getDashboardStateFromUrl(urlState, metricsView, explore);
    if (!partial) return;

    updateMetricsExplorerByName(name, (exploreState) => {
      for (const key in partial) {
        exploreState[key] = partial[key];
      }
      // this hack is needed since what is shown for comparison is not a single source
      // TODO: use an enum and get rid of this
      if (!partial.showTimeComparison) {
        exploreState.showTimeComparison = false;
      }
      exploreState.dimensionFilterExcludeMode = includeExcludeModeFromFilters(
        partial.whereFilter,
      );
      AdvancedMeasureCorrector.correct(exploreState, metricsView);
    });
  },

  mergePartialExplorerEntity(
    name: string,
    partialExploreState: Partial<ExploreState>,
    metricsView: V1MetricsViewSpec,
  ) {
    partialExploreState = structuredClone(partialExploreState);

    updateMetricsExplorerByName(name, (exploreState) => {
      for (const key in partialExploreState) {
        exploreState[key] = partialExploreState[key];
      }
      // this hack is needed since what is shown for comparison is not a single source
      // TODO: use an enum and get rid of this
      if (!partialExploreState.showTimeComparison) {
        exploreState.showTimeComparison = false;
      }
      exploreState.dimensionFilterExcludeMode = includeExcludeModeFromFilters(
        partialExploreState.whereFilter,
      );
      AdvancedMeasureCorrector.correct(exploreState, metricsView);
    });
  },

  sync(name: string, explore: V1ExploreSpec) {
    if (!name || !explore || !explore.measures) return;
    updateMetricsExplorerByName(name, (exploreState) => {
      // remove references to non existent measures
      syncMeasures(explore, exploreState);

      // remove references to non existent dimensions
      syncDimensions(explore, exploreState);
    });
  },

  setPivotMode(name: string, mode: boolean) {
    updateMetricsExplorerByName(name, (exploreState) => {
      if (mode) {
        exploreState.activePage = DashboardState_ActivePage.PIVOT;
      } else if (exploreState.selectedDimensionName) {
        exploreState.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
      } else if (exploreState.tdd.expandedMeasureName) {
        exploreState.activePage =
          DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
      } else {
        exploreState.activePage = DashboardState_ActivePage.DEFAULT;
      }
    });
  },

  setPivotRows(name: string, value: PivotChipData[]) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot.rowPage = 1;
      exploreState.pivot.activeCell = null;

      const dimensions: PivotChipData[] = [];

      value.forEach((val) => {
        if (val.type !== PivotChipType.Measure) {
          dimensions.push(val);
        }
      });

      if (exploreState.pivot.sorting.length) {
        const accessor = exploreState.pivot.sorting[0].id;
        const anchorDimension = dimensions?.[0]?.id;
        if (accessor !== anchorDimension) {
          exploreState.pivot.sorting = [];
        }
      }

      exploreState.pivot.rows = dimensions;
    });
  },

  setPivotColumns(name: string, value: PivotChipData[]) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot.rowPage = 1;
      exploreState.pivot.activeCell = null;
      exploreState.pivot.expanded = {};

      if (exploreState.pivot.sorting.length) {
        const accessor = exploreState.pivot.sorting[0].id;

        if (exploreState.pivot.tableMode === "flat") {
          const validAccessors = value.map((d) => d.id);
          if (!validAccessors.includes(accessor)) {
            exploreState.pivot.sorting = [];
          }
        } else {
          const anchorDimension = exploreState.pivot.rows?.[0]?.id;
          if (accessor !== anchorDimension) {
            exploreState.pivot.sorting = [];
          }
        }
      }
      exploreState.pivot.columns = value;
    });
  },

  addPivotField(name: string, value: PivotChipData, rows: boolean) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot.rowPage = 1;
      exploreState.pivot.activeCell = null;
      if (value.type === PivotChipType.Measure) {
        exploreState.pivot.columns.push(value);
      } else {
        if (rows) {
          exploreState.pivot.rows.push(value);
        } else {
          exploreState.pivot.columns.push(value);
        }
      }
    });
  },

  setPivotExpanded(name: string, expanded: ExpandedState) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot = { ...exploreState.pivot, expanded };
    });
  },

  setPivotComparison(name: string, enableComparison: boolean) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot = { ...exploreState.pivot, enableComparison };
    });
  },

  setPivotSort(name: string, sorting: SortingState) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot = {
        ...exploreState.pivot,
        sorting,
        rowPage: 1,
        expanded: {},
        activeCell: null,
      };
    });
  },

  setPivotColumnPage(name: string, pageNumber: number) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot = {
        ...exploreState.pivot,
        columnPage: pageNumber,
      };
    });
  },

  setPivotRowPage(name: string, pageNumber: number) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot = {
        ...exploreState.pivot,
        rowPage: pageNumber,
      };
    });
  },

  setPivotActiveCell(name: string, rowId: string, columnId: string) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot.activeCell = { rowId, columnId };
    });
  },

  removePivotActiveCell(name: string) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot.activeCell = null;
    });
  },

  createPivot(name: string, rows: PivotChipData[], columns: PivotChipData[]) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.activePage = DashboardState_ActivePage.PIVOT;
      exploreState.pivot = {
        ...exploreState.pivot,
        rows,
        columns,
        expanded: {},
        sorting: [],
        columnPage: 1,
        rowPage: 1,
        activeCell: null,
      };
    });
  },

  setExpandedMeasureName(name: string, measureName: string | undefined) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.tdd.expandedMeasureName = measureName;
      if (measureName) {
        exploreState.activePage =
          DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL;
      } else {
        exploreState.activePage = DashboardState_ActivePage.DEFAULT;
      }

      // If going into TDD view and already having a comparison dimension,
      // then set the pinIndex
      if (exploreState.selectedComparisonDimension) {
        exploreState.tdd.pinIndex = getPinIndexForDimension(
          exploreState,
          exploreState.selectedComparisonDimension,
        );
      }
    });
  },

  setPinIndex(name: string, index: number) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.tdd.pinIndex = index;
    });
  },

  setTDDChartType(name: string, type: TDDChart) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.tdd.chartType = type;
    });
  },

  setSelectedTimeRange(name: string, timeRange: DashboardTimeControls) {
    updateMetricsExplorerByName(name, (exploreState) => {
      setSelectedScrubRange(exploreState, undefined);
      exploreState.selectedTimeRange = timeRange;
    });
  },

  setSelectedScrubRange(name: string, scrubRange: ScrubRange | undefined) {
    updateMetricsExplorerByName(name, (exploreState) => {
      setSelectedScrubRange(exploreState, scrubRange);
    });
  },

  setMetricDimensionName(name: string, dimensionName: string | null) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.selectedDimensionName = dimensionName ?? undefined;
      if (dimensionName) {
        exploreState.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
      } else {
        exploreState.activePage = DashboardState_ActivePage.DEFAULT;
      }
    });
  },

  setComparisonDimension(name: string, dimensionName: string) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.selectedComparisonDimension = dimensionName;
      exploreState.tdd.pinIndex = getPinIndexForDimension(
        exploreState,
        dimensionName,
      );
    });
  },

  disableAllComparisons(name: string) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.selectedComparisonDimension = undefined;
    });
  },

  setSelectedComparisonRange(
    name: string,
    comparisonTimeRange: DashboardTimeControls,
    metricsViewSpec: V1MetricsViewSpec,
  ) {
    updateMetricsExplorerByName(name, (exploreState) => {
      if (comparisonTimeRange) {
        exploreState.showTimeComparison = true;
      }
      exploreState.selectedComparisonTimeRange = comparisonTimeRange;
      AdvancedMeasureCorrector.correct(exploreState, metricsViewSpec);
    });
  },

  setTimeZone(name: string, zoneIANA: string) {
    updateMetricsExplorerByName(name, (exploreState) => {
      // Reset scrub when timezone changes
      setSelectedScrubRange(exploreState, undefined);

      exploreState.selectedTimezone = zoneIANA;
    });
  },

  displayTimeComparison(name: string, showTimeComparison: boolean) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.showTimeComparison = showTimeComparison;
    });
  },

  selectTimeRange(
    name: string,
    timeRange: TimeRange,
    timeGrain: V1TimeGrain,
    comparisonTimeRange: DashboardTimeControls | undefined,
    metricsViewSpec: V1MetricsViewSpec,
  ) {
    updateMetricsExplorerByName(name, (exploreState) => {
      if (!timeRange.name) return;

      // Reset scrub when range changes
      setSelectedScrubRange(exploreState, undefined);

      if (timeRange.name === TimeRangePreset.ALL_TIME) {
        exploreState.showTimeComparison = false;
      }

      exploreState.selectedTimeRange = {
        ...timeRange,
        interval: timeGrain,
      };

      exploreState.selectedComparisonTimeRange = comparisonTimeRange;

      AdvancedMeasureCorrector.correct(exploreState, metricsViewSpec);
    });
  },

  setTimeGrain(name: string, timeGrain: V1TimeGrain) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.selectedTimeRange = {
        ...(exploreState.selectedTimeRange as DashboardTimeControls),
        interval: timeGrain,
      };
    });
  },

  remove(name: string) {
    update((state) => {
      delete state.entities[name];
      return state;
    });
  },

  setPivotTableMode(
    name: string,
    tableMode: PivotTableMode,
    rows: PivotChipData[],
    columns: PivotChipData[],
  ) {
    updateMetricsExplorerByName(name, (exploreState) => {
      exploreState.pivot = {
        ...exploreState.pivot,
        tableMode,
        rows,
        columns,
        sorting: [],
        expanded: {},
        activeCell: null,
      };
    });
  },
};

export const metricsExplorerStore: Readable<MetricsExplorerStoreType> &
  typeof metricsViewReducers = {
  subscribe,
  ...metricsViewReducers,
};

export function useExploreState(name: string): Readable<ExploreState> {
  return derived(metricsExplorerStore, ($store) => {
    return $store.entities[name];
  });
}

export function sortTypeForContextColumnType(
  contextColumn: LeaderboardContextColumn,
): SortType {
  const sortType = {
    [LeaderboardContextColumn.DELTA_PERCENT]: SortType.DELTA_PERCENT,
    [LeaderboardContextColumn.DELTA_ABSOLUTE]: SortType.DELTA_ABSOLUTE,
    [LeaderboardContextColumn.PERCENT]: SortType.PERCENT,
    [LeaderboardContextColumn.HIDDEN]: SortType.VALUE,
  }[contextColumn];

  // Note: the above map needs to be EXHAUSTIVE over
  // LeaderboardContextColumn variants. If we ever add a new
  // context column type, we need to add it to the map above.
  // Otherwise, we will throw an error here.
  if (!sortType) {
    throw new Error(`Invalid context column type: ${contextColumn}`);
  }
  return sortType;
}

function setSelectedScrubRange(
  exploreState: ExploreState,
  scrubRange: ScrubRange | undefined,
) {
  if (scrubRange === undefined) {
    exploreState.lastDefinedScrubRange = undefined;
  } else if (!scrubRange.isScrubbing && scrubRange?.start && scrubRange?.end) {
    exploreState.lastDefinedScrubRange = scrubRange;
  }

  exploreState.selectedScrubRange = scrubRange;
}

function getPinIndexForDimension(
  exploreState: ExploreState,
  dimensionName: string,
) {
  const dimensionEntryIndex = getWhereFilterExpressionIndex({
    dashboard: exploreState,
  })(dimensionName);
  if (dimensionEntryIndex === undefined || dimensionEntryIndex === -1)
    return -1;

  const dimExpr = exploreState.whereFilter.cond?.exprs?.[dimensionEntryIndex];
  if (!dimExpr?.cond?.exprs?.length) return -1;

  // 1st entry in the expression is the identifier. hence the -2 here.
  return dimExpr.cond.exprs.length - 2;
}

export const dimensionSearchText = writable("");
