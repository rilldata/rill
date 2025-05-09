import {
  PivotChipType,
  type PivotChipData,
  type PivotTableMode,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import { ToLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  FromURLParamsSortTypeMap,
  FromURLParamTDDChartMap,
  FromURLParamTimeDimensionMap,
  FromURLParamTimeGrainMap,
  FromURLParamTimeRangePresetMap,
  FromURLParamViewMap,
  ToActivePageViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import {
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState,
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreWebView,
  V1Operation,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { SortingState } from "@tanstack/svelte-table";
import { TIME_COMPARISON, TIME_GRAIN } from "../../../lib/time/config";
import { validateISODuration } from "../../../lib/time/ranges/iso-ranges";
import { stripMeasureSuffix } from "../filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "../filters/measure-filters/measure-filter-utils";
import { SortDirection } from "../proto-state/derived-types";
import { base64ToProto } from "../proto-state/fromProto";
import { createAndExpression, filterIdentifiers } from "../stores/filter-utils";
import type { MetricsExplorerEntity } from "../stores/metrics-explorer-entity";
import { decompressUrlParams } from "./compression";
import { convertLegacyStateToExplorePreset } from "./convertLegacyStateToExplorePreset";
import { convertPresetToExploreState } from "./convertPresetToExploreState";
import {
  convertFilterParamToExpression,
  stripParserError,
} from "./filters/converters";

export function convertURLSearchParamsToExploreState(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
) {
  // Get the measures and dimensions to help with some conversions
  const measures = getMapFromArray(
    metricsView.measures?.filter((m) => explore.measures?.includes(m.name!)) ??
      [],
    (m) => m.name!,
  );
  const dimensions = getMapFromArray(
    metricsView.dimensions?.filter((d) =>
      explore.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );

  // Decompress URL params if necessary
  if (searchParams.has(ExploreStateURLParams.GzippedParams)) {
    searchParams = new URLSearchParams(
      decompressUrlParams(
        searchParams.get(ExploreStateURLParams.GzippedParams)!,
      ),
    );
  }

  // URLSearchParams -> Partial<ExploreState> conversions
  const { exploreState: legacyStateState, errors: legacyStateErrors } =
    getExploreStateFromLegacyStateUrlParam(searchParams, metricsView, explore);
  const { exploreState: activePageState, errors: activePageErrors } =
    getActivePageFromURLParams(searchParams);
  const { exploreState: filtersState, errors: filtersErrors } =
    getFiltersFromURLParams(searchParams, measures, dimensions);
  const { exploreState: timeControlsState, errors: timeControlsErrors } =
    getTimeControlsFromURLParams(searchParams, dimensions);
  const { exploreState: viewState, errors: viewErrors } =
    getViewSpecificStateFromURLParams(
      searchParams,
      measures,
      dimensions,
      explore,
    );

  // Combine all the states
  const exploreState: Partial<MetricsExplorerEntity> = {
    ...legacyStateState, // First in this list so that it can be overridden by additional explicit URL state
    ...activePageState,
    ...filtersState,
    ...timeControlsState,
    ...viewState,
  };
  const errors: Error[] = [
    ...legacyStateErrors,
    ...activePageErrors,
    ...filtersErrors,
    ...timeControlsErrors,
    ...viewErrors,
  ];

  return { exploreState, errors };
}

function getActivePageFromURLParams(searchParams: URLSearchParams): {
  exploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  const exploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.WebView)) {
    const view = searchParams.get(ExploreStateURLParams.WebView) as string;

    if (view in FromURLParamViewMap) {
      const internalView = FromURLParamViewMap[view];
      exploreState.activePage = Number(
        ToActivePageViewMap[internalView] ?? "0",
      );
    } else {
      errors.push(getSingleFieldError("view", view));
    }
  }

  return { exploreState, errors };
}

function getFiltersFromURLParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
): {
  exploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  const exploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.Filters)) {
    const filter = searchParams.get(ExploreStateURLParams.Filters) as string;

    const {
      expr,
      dimensionsWithInlistFilter,
      errors: filterErrors,
    } = fromFilterUrlParam(filter, measures, dimensions);

    if (filterErrors) errors.push(...filterErrors);

    if (expr) {
      const { dimensionFilters, dimensionThresholdFilters } =
        splitWhereFilter(expr);
      exploreState.whereFilter = dimensionFilters;
      exploreState.dimensionThresholdFilters = dimensionThresholdFilters;
    }

    if (dimensionsWithInlistFilter) {
      exploreState.dimensionsWithInlistFilter = dimensionsWithInlistFilter;
    }
  }

  return { exploreState, errors };
}

function fromFilterUrlParam(
  filter: string,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
): {
  expr?: V1Expression;
  dimensionsWithInlistFilter?: string[];
  errors?: Error[];
} {
  try {
    const { expr: exprFromFilter, dimensionsWithInlistFilter } =
      convertFilterParamToExpression(filter);
    let expr = exprFromFilter;
    if (!expr) {
      return {
        expr: createAndExpression([]),
        errors: [new Error("Failed to parse filter: " + filter)],
      };
    }

    // if root is not AND/OR then add AND
    if (
      expr?.cond?.op !== V1Operation.OPERATION_AND &&
      expr?.cond?.op !== V1Operation.OPERATION_OR
    ) {
      expr = createAndExpression([expr]);
    }
    const errors: Error[] = [];
    const missingDims: string[] = [];
    const missingFields: string[] = [];
    expr =
      filterIdentifiers(expr, (e, ident) => {
        if (
          // these we are sure are dimensions so add errors as "missing dimension"
          e.cond?.op === V1Operation.OPERATION_IN ||
          e.cond?.op === V1Operation.OPERATION_NIN ||
          !!e.subquery
        ) {
          if (dimensions.has(ident)) {
            return true;
          }
          missingDims.push(ident);
          return false;
        }

        if (
          measures.has(ident) ||
          measures.has(stripMeasureSuffix(ident)) ||
          dimensions.has(ident)
        ) {
          return true;
        }
        missingFields.push(ident);

        return false;
      }) ?? createAndExpression([]);
    if (missingDims.length) {
      errors.push(getMultiFieldError("filter dimension", missingDims));
    }
    if (missingFields.length) {
      errors.push(getMultiFieldError("filter field", missingFields));
    }
    return { expr, dimensionsWithInlistFilter, errors };
  } catch (e) {
    return {
      errors: [new Error("Selected filter is invalid: " + stripParserError(e))],
    };
  }
}

export function getTimeControlsFromURLParams(
  searchParams: URLSearchParams,
  dimensions: Map<string, MetricsViewSpecDimension>,
): {
  exploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  const exploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  // Time range
  const timeRange = searchParams.get(ExploreStateURLParams.TimeRange);
  if (
    timeRange &&
    (timeRange in FromURLParamTimeRangePresetMap ||
      validateISODuration(timeRange) ||
      CustomTimeRangeRegex.test(timeRange))
  ) {
    exploreState.selectedTimeRange = fromTimeRangeUrlParam(timeRange);
  } else if (timeRange) {
    errors.push(getSingleFieldError("time range", timeRange));
  }

  // Time grain
  const timeGrain = searchParams.get(ExploreStateURLParams.TimeGrain);
  if (timeGrain && timeGrain in FromURLParamTimeGrainMap) {
    if (exploreState.selectedTimeRange) {
      exploreState.selectedTimeRange.interval =
        FromURLParamTimeGrainMap[timeGrain];
    }
  } else if (timeGrain) {
    errors.push(getSingleFieldError("time grain", timeGrain));
  }

  // Timezone
  const timezone = searchParams.get(ExploreStateURLParams.TimeZone);
  if (timezone) {
    exploreState.selectedTimezone = timezone;
  }

  // Comparison time range
  const ctr = searchParams.get(ExploreStateURLParams.ComparisonTimeRange);
  let setCompareTimeRange = false;

  if (ctr && (ctr in TIME_COMPARISON || CustomTimeRangeRegex.test(ctr))) {
    exploreState.selectedComparisonTimeRange = fromTimeRangeUrlParam(ctr);
    exploreState.showTimeComparison = true;
    setCompareTimeRange = true;
  } else if (ctr === "") {
    exploreState.selectedComparisonTimeRange = undefined;
    exploreState.showTimeComparison = false;
  } else if (ctr) {
    errors.push(getSingleFieldError("compare time range", ctr));
  }

  // Comparison dimension
  const compDim = searchParams.get(ExploreStateURLParams.ComparisonDimension);
  if (compDim === "") {
    exploreState.selectedComparisonDimension = "";
    exploreState.selectedComparisonTimeRange = undefined;
    exploreState.showTimeComparison = false;
  } else if (compDim && dimensions.has(compDim)) {
    exploreState.selectedComparisonDimension = compDim;
    if (!setCompareTimeRange) {
      exploreState.selectedComparisonTimeRange = undefined;
      exploreState.showTimeComparison = false;
    }
  } else if (compDim) {
    errors.push(getSingleFieldError("compare dimension", compDim));
  }

  // Highlighted time range
  const selectTr = searchParams.get(ExploreStateURLParams.HighlightedTimeRange);
  if (selectTr === "") {
    exploreState.selectedScrubRange = undefined;
  } else if (selectTr && CustomTimeRangeRegex.test(selectTr)) {
    const scrubRange = fromTimeRangeUrlParam(selectTr);
    exploreState.lastDefinedScrubRange = exploreState.selectedScrubRange = {
      ...scrubRange,
      isScrubbing: false,
    };
  } else if (selectTr) {
    errors.push(getSingleFieldError("highlighted time range", selectTr));
  }

  // comparisonMode override: explicitly NONE
  const cm = searchParams.get(ExploreStateURLParams.ComparisonTimeRange);
  if (
    cm === "" ||
    searchParams.get(ExploreStateURLParams.ComparisonDimension) === ""
  ) {
    exploreState.selectedComparisonTimeRange = undefined;
    exploreState.selectedComparisonDimension = "";
    exploreState.showTimeComparison = false;
  }

  return { exploreState, errors };
}

export const CustomTimeRangeRegex =
  /(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z),(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z)/;
export function fromTimeRangeUrlParam(tr: string) {
  const customTimeRangeMatch = CustomTimeRangeRegex.exec(tr);
  if (customTimeRangeMatch?.length) {
    const [, start, end] = customTimeRangeMatch;
    return {
      name: TimeRangePreset.CUSTOM,
      start: new Date(start),
      end: new Date(end),
    } as DashboardTimeControls;
  }
  return {
    name: tr,
  } as DashboardTimeControls;
}

function getExploreViewStateFromURLParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  explore: V1ExploreSpec,
): { exploreState: Partial<MetricsExplorerEntity>; errors: Error[] } {
  const exploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  // Measures
  const mes = searchParams.get(ExploreStateURLParams.VisibleMeasures);
  if (mes !== null) {
    const selected =
      mes === "*"
        ? (explore.measures ?? [])
        : mes.split(",").filter((m) => measures.has(m));
    const missing =
      mes === "*" ? [] : getMissingValues(selected, mes.split(","));
    if (missing.length) errors.push(getMultiFieldError("measure", missing));
    exploreState.visibleMeasures = [...selected];
    exploreState.allMeasuresVisible =
      selected.length === (explore.measures?.length ?? 0);
  }

  // Dimensions
  const dims = searchParams.get(ExploreStateURLParams.VisibleDimensions);
  if (dims !== null) {
    const selected =
      dims === "*"
        ? (explore.dimensions ?? [])
        : dims.split(",").filter((d) => dimensions.has(d));
    const missing =
      dims === "*" ? [] : getMissingValues(selected, dims.split(","));
    if (missing.length) errors.push(getMultiFieldError("dimension", missing));
    exploreState.visibleDimensions = [...selected];
    exploreState.allDimensionsVisible =
      selected.length === (explore.dimensions?.length ?? 0);
  }

  // Expanded Dimension
  const dim = searchParams.get(ExploreStateURLParams.ExpandedDimension);
  const view = searchParams.get(ExploreStateURLParams.WebView); // used for activePage logic

  if (dim !== null) {
    if (dim === "") {
      exploreState.selectedDimensionName = "";
      if (
        view === null ||
        view === V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED.toString()
      ) {
        exploreState.activePage = DashboardState_ActivePage.DEFAULT;
      }
    } else if (dimensions.has(dim)) {
      exploreState.selectedDimensionName = dim;
      if (
        view === null ||
        view === V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED.toString() ||
        view === V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE.toString()
      ) {
        exploreState.activePage = DashboardState_ActivePage.DIMENSION_TABLE;
      }
    } else {
      errors.push(getSingleFieldError("expanded dimension", dim));
    }
  }

  // Sort By
  const sortBy = searchParams.get(ExploreStateURLParams.SortBy);
  if (sortBy) {
    if (measures.has(sortBy)) {
      // Validate it's visible (or all measures visible)
      if (
        mes === "*" ||
        !mes ||
        exploreState.visibleMeasures?.includes(sortBy)
      ) {
        exploreState.leaderboardSortByMeasureName = sortBy;
      } else {
        errors.push(
          getSingleFieldError("sort by measure", sortBy, "It is hidden."),
        );
      }
    } else {
      errors.push(getSingleFieldError("sort by measure", sortBy));
    }
  }

  // Sort Direction
  const sortDir = searchParams.get(ExploreStateURLParams.SortDirection);
  if (sortDir === "ASC") {
    // TODO: maybe normalize these to lowercase?
    exploreState.sortDirection = SortDirection.ASCENDING;
  } else if (sortDir === "DESC") {
    exploreState.sortDirection = SortDirection.DESCENDING;
  }

  // Sort Type
  const sortType = searchParams.get(ExploreStateURLParams.SortType);
  if (sortType) {
    const mapped = FromURLParamsSortTypeMap[sortType];
    if (mapped !== undefined) {
      exploreState.dashboardSortType =
        Number(ToLegacySortTypeMap[mapped]) ??
        DashboardState_LeaderboardSortType.UNSPECIFIED;
    } else {
      errors.push(getSingleFieldError("sort type", sortType));
    }
  }

  // Leaderboard Measure Count
  const count = searchParams.get(ExploreStateURLParams.LeaderboardMeasureCount);
  if (count !== null) {
    const parsed = parseInt(count, 10);
    if (!isNaN(parsed) && parsed > 0) {
      exploreState.leaderboardMeasureCount = parsed;
    } else {
      errors.push(getSingleFieldError("leaderboard measure count", count));
    }
  }

  return { exploreState, errors };
}

function getTDDViewStateFromURLParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
): { exploreState: Partial<MetricsExplorerEntity>; errors: Error[] } {
  const exploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  let expandedMeasureName = "";
  if (searchParams.has(ExploreStateURLParams.ExpandedMeasure)) {
    const mes = searchParams.get(
      ExploreStateURLParams.ExpandedMeasure,
    ) as string;
    if (mes === "" || measures.has(mes)) {
      expandedMeasureName = mes;
    } else {
      errors.push(getSingleFieldError("expanded measure", mes));
    }
  }

  const chartTypeRaw = searchParams.get(ExploreStateURLParams.ChartType);
  const chartType = chartTypeRaw
    ? (FromURLParamTDDChartMap[chartTypeRaw] ?? TDDChart.DEFAULT)
    : TDDChart.DEFAULT;

  const pin = searchParams.has(ExploreStateURLParams.Pin) ? 0 : -1;

  exploreState.tdd = {
    expandedMeasureName,
    chartType,
    pinIndex: pin,
  };

  return { exploreState, errors };
}

function getPivotViewStateFromURLParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
): { exploreState: Partial<MetricsExplorerEntity>; errors: Error[] } {
  const errors: Error[] = [];

  const mapPivotEntry = (entry: string): PivotChipData | undefined => {
    if (entry in FromURLParamTimeDimensionMap) {
      const grain = FromURLParamTimeDimensionMap[entry];
      return {
        id: grain,
        title: TIME_GRAIN[grain]?.label,
        type: PivotChipType.Time,
      };
    }

    if (measures.has(entry)) {
      const m = measures.get(entry)!;
      return {
        id: entry,
        title: m.displayName || m.name || "Unknown",
        type: PivotChipType.Measure,
      };
    }

    if (dimensions.has(entry)) {
      const d = dimensions.get(entry)!;
      return {
        id: entry,
        title: d.displayName || d.name || "Unknown",
        type: PivotChipType.Dimension,
      };
    }

    errors.push(getSingleFieldError("pivot entry", entry));
    return undefined;
  };

  const rowParam = searchParams.get(ExploreStateURLParams.PivotRows);
  const colParam = searchParams.get(ExploreStateURLParams.PivotColumns);

  const rowDimensions = rowParam
    ? rowParam
        .split(",")
        .map(mapPivotEntry)
        .filter((x): x is PivotChipData => !!x)
    : [];

  const colDimensions = colParam
    ? colParam
        .split(",")
        .map(mapPivotEntry)
        .filter((x): x is PivotChipData => !!x)
    : [];

  const pivotIsActive =
    searchParams.get(ExploreStateURLParams.WebView) ===
    V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT.toString();

  if (
    !pivotIsActive &&
    rowDimensions.length === 0 &&
    colDimensions.length === 0
  ) {
    return {
      exploreState: {
        pivot: {
          rows: [],
          columns: [],
          sorting: [],
          expanded: {},
          columnPage: 1,
          rowPage: 1,
          enableComparison: true,
          activeCell: null,
          tableMode: "nest",
        },
      },
      errors,
    };
  }

  // Sort
  const sorting: SortingState = [];
  const sortBy = searchParams.get(ExploreStateURLParams.SortBy);
  if (sortBy) {
    const sortId =
      sortBy in FromURLParamTimeDimensionMap
        ? FromURLParamTimeDimensionMap[sortBy]
        : sortBy;
    sorting.push({
      id: sortId,
      desc: searchParams.get(ExploreStateURLParams.SortDirection) !== "ASC", // TODO: maybe normalize these to lowercase?
    });
  }

  // Table Mode
  let tableMode: PivotTableMode = "nest";
  const tableModeParam = searchParams.get(ExploreStateURLParams.PivotTableMode);
  if (tableModeParam) {
    if (tableModeParam === "flat" || tableModeParam === "nest") {
      tableMode = tableModeParam;
    } else {
      errors.push(getSingleFieldError("pivot table mode", tableModeParam));
    }
  }

  return {
    exploreState: {
      pivot: {
        rows: rowDimensions,
        columns: colDimensions,
        sorting,
        expanded: {},
        columnPage: 1,
        rowPage: 1,
        enableComparison: true,
        activeCell: null,
        tableMode,
      },
    },
    errors,
  };
}

function getViewSpecificStateFromURLParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  explore: V1ExploreSpec,
): { exploreState: Partial<MetricsExplorerEntity>; errors: Error[] } {
  const viewParamValue = searchParams.get(ExploreStateURLParams.WebView);

  // If no view parameter, default to explore view
  if (!viewParamValue) {
    return getExploreViewStateFromURLParams(
      searchParams,
      measures,
      dimensions,
      explore,
    );
  }

  // If view parameter exists but isn't valid, return error
  if (!(viewParamValue in FromURLParamViewMap)) {
    return {
      exploreState: {},
      errors: [getSingleFieldError("view", viewParamValue)],
    };
  }

  // At this point we know we have a valid view
  const view = FromURLParamViewMap[viewParamValue];

  // Get the view-specific state
  switch (view) {
    case V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE:
      return getExploreViewStateFromURLParams(
        searchParams,
        measures,
        dimensions,
        explore,
      );
    case V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION:
      return getTDDViewStateFromURLParams(searchParams, measures);
    case V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT:
      return getPivotViewStateFromURLParams(searchParams, measures, dimensions);
    default:
      return {
        exploreState: {},
        errors: [getSingleFieldError("view", viewParamValue)],
      };
  }
}

// Convert legacy protobuf `state` param
function getExploreStateFromLegacyStateUrlParam(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
): {
  exploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  if (!searchParams.has(ExploreStateURLParams.LegacyProtoState)) {
    return {
      exploreState: {},
      errors: [],
    };
  }

  let legacyState = searchParams.get(
    ExploreStateURLParams.LegacyProtoState,
  ) as string;

  try {
    legacyState = legacyState.includes("%")
      ? decodeURIComponent(legacyState)
      : legacyState;

    const legacyDashboardState = DashboardState.fromBinary(
      base64ToProto(legacyState),
    );

    // NOTE: In the future, we should convert the legacy state to the explore state directly,
    // without going through the explore preset.
    const { preset, errors: presetErrors } = convertLegacyStateToExplorePreset(
      legacyDashboardState,
      metricsView,
      explore,
    );
    const { partialExploreState, errors: exploreStateErrors } =
      convertPresetToExploreState(metricsView, explore, preset);

    return {
      exploreState: { ...partialExploreState },
      errors: [...presetErrors, ...exploreStateErrors],
    };
  } catch (e) {
    return {
      exploreState: {},
      errors: [e], // TODO: parse and show meaningful error
    };
  }
}
