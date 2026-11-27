import { injectEphemeralMeasuresIntoMap } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-state";
import { parseAndValidateEphemeralParam } from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-state";
import { narrowSingleSelectFilters } from "@rilldata/web-common/features/dashboards/filters/expr-utils";
import { PIVOT_ROW_LIMIT_OPTIONS } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import { dedupePivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-chip-utils";
import { fromPivotFormattingParam } from "@rilldata/web-common/features/dashboards/pivot/pivot-formatting-param";
import { pivotHasSortField } from "@rilldata/web-common/features/dashboards/pivot/pivot-sorting";
import {
  PIVOT_TOTALS_ROW_POSITIONS,
  type PivotChipData,
  PivotChipType,
  type PivotState,
  type PivotTableMode,
  type PivotTotalsRowPosition,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { dropUnknownFieldsFromFilter } from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import {
  type TDDState,
  TDDChart,
} from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  convertFilterParamToExpression,
  stripParserError,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { ToLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  type ExploreUrlWebView,
  FromURLParamsSortTypeMap,
  FromURLParamTDDChartMap,
  FromURLParamTimeDimensionMap,
  FromURLParamViewMap,
  ToActivePageViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { normalizeExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/merge-explore-url-params";
import { validateRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  dedupe,
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import {
  TIME_COMPARISON,
  TIME_GRAIN,
} from "@rilldata/web-common/lib/time/config";
import { TokenToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1Expression,
  type V1MetricsViewSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { SortingState } from "tanstack-table-8-svelte-5";

/**
 * Converts explore url params to a partial explore state.
 * Params that are not valid for the specs are dropped and reported in `errors`.
 *
 * Only the params of the url's view are read; the others are ignored.
 * The tdd and pivot states, the comparison dimension and the chart y-axis are always set,
 * so a missing param resets them instead of keeping a value from another source.
 */
export function convertURLSearchParamsToExploreState(
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
) {
  searchParams = normalizeExploreUrlParams(searchParams);
  const partialExploreState: Partial<ExploreState> = {};
  let errors: Error[] = [];

  const measures = getMapFromArray(
    metricsView.measures?.filter((m) => explore.measures?.includes(m.name!)) ??
      [],
    (m) => m.name!,
  );
  const allDimensions = getMapFromArray(
    metricsView.dimensions?.filter((d) =>
      explore.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );
  // Leaderboards, filters and pivot do not take time dimensions.
  const dimensions = new Map(
    [...allDimensions].filter(([, d]) => d.type !== "DIMENSION_TYPE_TIME"),
  );

  // normalizeExploreUrlParams keeps a legacy state that does not decode.
  if (searchParams.has(ExploreStateURLParams.LegacyProtoState)) {
    errors.push(new Error("Failed to parse the legacy dashboard state."));
  }

  // Parse ephemeral measures before any other param: their names are valid
  // measure names everywhere (visible measures, leaderboards, sort, pivot
  // columns, formatting, ...), which is achieved by injecting synthetic spec
  // measures into the `measures` map used by all validations below.
  const ephemeralParam = searchParams.get(
    ExploreStateURLParams.EphemeralMeasures,
  );
  if (ephemeralParam !== null) {
    const { valid, invalidEntries } = parseAndValidateEphemeralParam(
      ephemeralParam,
      measures,
      metricsView,
    );
    partialExploreState.ephemeralMeasures = valid.length ? valid : undefined;
    injectEphemeralMeasuresIntoMap(measures, valid);
    if (invalidEntries.length) {
      errors.push(getMultiFieldError("adhoc measure", invalidEntries));
    }
  }

  let view: V1ExploreWebView | undefined;
  const viewParam = searchParams.get(ExploreStateURLParams.WebView);
  if (viewParam !== null) {
    if (viewParam in FromURLParamViewMap) {
      const parsedView = FromURLParamViewMap[viewParam as ExploreUrlWebView];
      view = parsedView;
      partialExploreState.activePage = Number(
        ToActivePageViewMap[parsedView] ?? "0",
      );
    } else {
      errors.push(getSingleFieldError("view", viewParam));
    }
  }

  const filterParam = searchParams.get(ExploreStateURLParams.Filters);
  if (filterParam !== null) {
    const {
      expr,
      dimensionsWithInlistFilter,
      errors: filterErrors,
    } = fromFilterUrlParam(filterParam, measures, dimensions);
    if (filterErrors) errors.push(...filterErrors);

    // A link or bookmark can hold several values for a single-select filter. Narrow them here,
    // so the explore state and the url it is written back to match the chip.
    const singleSelectFilters = new Set(explore.singleSelectFilters ?? []);
    const isSingleSelect = (name: string) => singleSelectFilters.has(name);
    if (expr) {
      partialExploreState.whereFilter = singleSelectFilters.size
        ? (narrowSingleSelectFilters(expr, isSingleSelect) ?? expr)
        : expr;
      partialExploreState.dimensionThresholdFilters = [];
    }
    if (dimensionsWithInlistFilter) {
      partialExploreState.dimensionsWithInlistFilter = expr
        ? dimensionsWithInlistFilter.filter((name) => !isSingleSelect(name))
        : dimensionsWithInlistFilter;
    }
  }

  const { partialExploreState: trPartialState, errors: trErrors } =
    fromTimeRangesParams(searchParams, allDimensions);
  Object.assign(partialExploreState, trPartialState);
  errors.push(...trErrors);

  // Visible measures named by the url; leaderboard measures have to be among them.
  let urlVisibleMeasures: string[] | undefined;
  if (
    view === undefined ||
    view === V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE
  ) {
    const {
      partialExploreState: ovPartialState,
      errors: ovErrors,
      urlVisibleMeasures: visibleMeasures,
    } = fromExploreUrlParams(
      searchParams,
      measures,
      dimensions,
      allDimensions,
      explore,
    );
    Object.assign(partialExploreState, ovPartialState);
    errors.push(...ovErrors);
    urlVisibleMeasures = visibleMeasures;
  }

  const { tdd, errors: tddErrors } = fromTimeDimensionUrlParams(
    searchParams,
    measures,
    view,
  );
  partialExploreState.tdd = tdd;
  errors.push(...tddErrors);

  const { pivot, errors: pivotErrors } = fromPivotUrlParams(
    searchParams,
    measures,
    dimensions,
    view,
    metricsView.timeDimension,
  );
  partialExploreState.pivot = pivot;
  errors.push(...pivotErrors);

  const leaderboardMeasuresParam = searchParams.get(
    ExploreStateURLParams.LeaderboardMeasures,
  );
  if (leaderboardMeasuresParam !== null) {
    const leaderboardMeasures = leaderboardMeasuresParam.split(",");
    // Unset all of them if any is unknown or hidden.
    const allMeasuresValid = leaderboardMeasures.every(
      (measure) =>
        measures.has(measure) &&
        (!urlVisibleMeasures || urlVisibleMeasures.includes(measure)),
    );
    partialExploreState.leaderboardMeasureNames = allMeasuresValid
      ? leaderboardMeasures
      : [];
  }

  const showContextParam = searchParams.get(
    ExploreStateURLParams.LeaderboardShowContextForAllMeasures,
  );
  if (showContextParam !== null) {
    const showContext = parseBooleanParam(showContextParam);
    if (showContext !== undefined) {
      partialExploreState.leaderboardShowContextForAllMeasures = showContext;
    } else {
      errors.push(
        getSingleFieldError("leaderboard context toggle", showContextParam),
      );
    }
  }

  partialExploreState.dynamicYAxisScale = false;
  const dynamicYAxisParam = searchParams.get(
    ExploreStateURLParams.DynamicYAxisScale,
  );
  if (dynamicYAxisParam !== null) {
    const dynamicYAxis = parseBooleanParam(dynamicYAxisParam);
    if (dynamicYAxis !== undefined) {
      partialExploreState.dynamicYAxisScale = dynamicYAxis;
    } else {
      errors.push(
        getSingleFieldError(
          "dynamic y-axis scale",
          dynamicYAxisParam.toLowerCase(),
        ),
      );
    }
  }

  if (searchParams.has(ExploreStateURLParams.IgnoreErrors)) {
    // Ignore all errors. TODO: better handling in the future
    errors = [];
  }

  return { partialExploreState, errors };
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
  // An empty param is no filter, e.g. one filled in from defaults to keep a default filter from applying.
  if (filter === "") return { expr: createAndExpression([]) };

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
    const { expr: cleanedExpr, errors } = dropUnknownFieldsFromFilter(
      expr,
      measures,
      dimensions,
    );
    return { expr: cleanedExpr, dimensionsWithInlistFilter, errors };
  } catch (e) {
    return {
      errors: [new Error("Selected filter is invalid: " + stripParserError(e))],
    };
  }
}

/**
 * Converts the time range and comparison params.
 * The comparison dimension, scrub range and time dimension are always set, so a missing param resets them.
 */
export function fromTimeRangesParams(
  searchParams: URLSearchParams,
  dimensions: Map<string, MetricsViewSpecDimension>,
) {
  const partialExploreState: Partial<ExploreState> = {};
  const errors: Error[] = [];

  const timeRangeParam = searchParams.get(ExploreStateURLParams.TimeRange);
  if (timeRangeParam !== null) {
    if (validateRillTime(timeRangeParam)) {
      errors.push(getSingleFieldError("time range", timeRangeParam));
    } else if (timeRangeParam) {
      partialExploreState.selectedTimeRange =
        fromTimeRangeUrlParam(timeRangeParam);
    }
  }

  const timezone = searchParams.get(ExploreStateURLParams.TimeZone);
  if (timezone) {
    partialExploreState.selectedTimezone = timezone;
  }

  const compareTimeRangeParam = searchParams.get(
    ExploreStateURLParams.ComparisonTimeRange,
  );
  let compareTimeRange: string | undefined;
  if (compareTimeRangeParam) {
    if (
      compareTimeRangeParam in TIME_COMPARISON ||
      CustomTimeRangeRegex.test(compareTimeRangeParam)
    ) {
      compareTimeRange = compareTimeRangeParam;
    } else {
      errors.push(
        getSingleFieldError("compare time range", compareTimeRangeParam),
      );
    }
  }

  // A missing grain means adaptive mode; the grain is derived from the time range later.
  const timeGrainParam = searchParams.get(ExploreStateURLParams.TimeGrain);
  if (timeGrainParam && timeGrainParam in TokenToV1TimeGrain) {
    partialExploreState.selectedTimeRange ??= {} as DashboardTimeControls;
    partialExploreState.selectedTimeRange.interval =
      TokenToV1TimeGrain[timeGrainParam];
    partialExploreState.adaptiveTimeGrain = false;
  } else {
    if (timeGrainParam !== null) {
      errors.push(getSingleFieldError("time grain", timeGrainParam));
    }
    if (partialExploreState.selectedTimeRange) {
      // A time range without a grain is adaptive: `correctExploreState` derives the grain from the range.
      partialExploreState.adaptiveTimeGrain = true;
    }
  }

  const comparisonDimensionParam = searchParams.get(
    ExploreStateURLParams.ComparisonDimension,
  );
  let comparisonDimension: string | undefined;
  if (comparisonDimensionParam) {
    if (dimensions.has(comparisonDimensionParam)) {
      comparisonDimension = comparisonDimensionParam;
    } else {
      errors.push(
        getSingleFieldError("compare dimension", comparisonDimensionParam),
      );
    }
  }

  if (compareTimeRange) {
    const comparisonTimeRange = fromTimeRangeUrlParam(compareTimeRange);
    if (comparisonTimeRange.name === TimeRangePreset.CUSTOM) {
      comparisonTimeRange.name = TimeComparisonOption.CUSTOM;
    }
    partialExploreState.selectedComparisonTimeRange = comparisonTimeRange;
    partialExploreState.showTimeComparison = true;
  }
  // A comparison time range takes precedence over a comparison dimension.
  partialExploreState.selectedComparisonDimension = comparisonDimension ?? "";
  if (
    !compareTimeRange &&
    // `compare_tr=` or `compare_dim=` unsets a comparison from the defaults.
    (comparisonDimension ||
      compareTimeRangeParam === "" ||
      comparisonDimensionParam === "")
  ) {
    partialExploreState.selectedComparisonTimeRange = undefined;
    partialExploreState.showTimeComparison = false;
  }

  const highlightedTimeRange = searchParams.get(
    ExploreStateURLParams.HighlightedTimeRange,
  );
  partialExploreState.selectedScrubRange = undefined;
  if (highlightedTimeRange) {
    if (CustomTimeRangeRegex.test(highlightedTimeRange)) {
      partialExploreState.lastDefinedScrubRange =
        partialExploreState.selectedScrubRange = {
          ...fromTimeRangeUrlParam(highlightedTimeRange),
          isScrubbing: false,
        };
    } else {
      errors.push(
        getSingleFieldError("highlighted time range", highlightedTimeRange),
      );
    }
  }

  // An unknown time dimension is dropped, falling back to the primary one.
  const timeDimension = searchParams.get(ExploreStateURLParams.TimeDimension);
  partialExploreState.selectedTimeDimension =
    timeDimension && dimensions.has(timeDimension) ? timeDimension : undefined;

  return { partialExploreState, errors };
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

function fromExploreUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  allDimensions: Map<string, MetricsViewSpecDimension>,
  explore: V1ExploreSpec,
) {
  const partialExploreState: Partial<ExploreState> = {};
  const errors: Error[] = [];

  const measuresParam = searchParams.get(ExploreStateURLParams.VisibleMeasures);
  let urlVisibleMeasures: string[] | undefined;
  if (measuresParam !== null) {
    const requestedMeasures =
      measuresParam === "*"
        ? (explore.measures ?? [])
        : measuresParam.split(",");
    urlVisibleMeasures = requestedMeasures.filter((m) => measures.has(m));
    const missingMeasures = getMissingValues(
      urlVisibleMeasures,
      requestedMeasures,
    );
    if (missingMeasures.length) {
      errors.push(getMultiFieldError("measure", missingMeasures));
    }

    if (urlVisibleMeasures.length) {
      // A url can name a field twice; the lists that render the fields are keyed by name.
      partialExploreState.visibleMeasures = dedupe(urlVisibleMeasures);
      // `measures` includes ephemeral measures, so a hidden spec measure can
      // never be masked by a visible ephemeral one.
      partialExploreState.allMeasuresVisible =
        partialExploreState.visibleMeasures.length === measures.size;
    }
  }

  const dimensionsParam = searchParams.get(
    ExploreStateURLParams.VisibleDimensions,
  );
  if (dimensionsParam !== null) {
    // `*` keeps the explore's time dimensions; a list of names does not take them.
    const requestedDimensions =
      dimensionsParam === "*"
        ? (explore.dimensions ?? [])
        : dimensionsParam.split(",");
    const validDimensions =
      dimensionsParam === "*" ? allDimensions : dimensions;
    const selectedDimensions = dedupe(requestedDimensions).filter((d) =>
      validDimensions.has(d),
    );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      requestedDimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }

    if (selectedDimensions.length) {
      // Compared against the explore's dimensions that the metrics view still defines,
      // without time dimensions since leaderboards do not show them.
      partialExploreState.allDimensionsVisible =
        selectedDimensions.filter((d) => dimensions.has(d)).length ===
        dimensions.size;
      partialExploreState.visibleDimensions = selectedDimensions;
    }
  }

  const expandedDimension = searchParams.get(
    ExploreStateURLParams.ExpandedDimension,
  );
  if (expandedDimension === "") {
    // Unsets an expanded dimension from the defaults.
    partialExploreState.selectedDimensionName = "";
    partialExploreState.activePage = DashboardState_ActivePage.DEFAULT;
  } else if (expandedDimension !== null) {
    if (dimensions.has(expandedDimension)) {
      partialExploreState.selectedDimensionName = expandedDimension;
      partialExploreState.activePage =
        DashboardState_ActivePage.DIMENSION_TABLE;
    } else {
      errors.push(getSingleFieldError("expanded dimension", expandedDimension));
    }
  }

  const sortBy = searchParams.get(ExploreStateURLParams.SortBy);
  if (sortBy !== null) {
    if (!measures.has(sortBy)) {
      errors.push(getSingleFieldError("sort by measure", sortBy));
    } else if (urlVisibleMeasures && !urlVisibleMeasures.includes(sortBy)) {
      errors.push(
        getSingleFieldError("sort by measure", sortBy, "It is hidden."),
      );
    } else {
      partialExploreState.leaderboardSortByMeasureName = sortBy;
    }
  }

  const sortDirection = searchParams.get(ExploreStateURLParams.SortDirection);
  if (sortDirection !== null) {
    partialExploreState.sortDirection =
      sortDirection === "ASC"
        ? SortDirection.ASCENDING
        : SortDirection.DESCENDING;
  }

  const sortType = searchParams.get(ExploreStateURLParams.SortType);
  if (sortType !== null) {
    if (sortType in FromURLParamsSortTypeMap) {
      partialExploreState.dashboardSortType = Number(
        ToLegacySortTypeMap[FromURLParamsSortTypeMap[sortType]] ??
          DashboardState_LeaderboardSortType.UNSPECIFIED,
      );
    } else {
      errors.push(getSingleFieldError("sort type", sortType));
    }
  }

  return { partialExploreState, errors, urlVisibleMeasures };
}

function fromTimeDimensionUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  view: V1ExploreWebView | undefined,
): { tdd: TDDState; errors: Error[] } {
  const errors: Error[] = [];
  const isTDDView = view === V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION;

  // The chart type is shared by the explore and tdd views.
  const chartTypeParam =
    view === V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT
      ? null
      : searchParams.get(ExploreStateURLParams.ChartType);
  const chartType =
    (chartTypeParam && FromURLParamTDDChartMap[chartTypeParam]) ||
    TDDChart.DEFAULT;

  let expandedMeasureName = "";
  const expandedMeasure = isTDDView
    ? searchParams.get(ExploreStateURLParams.ExpandedMeasure)
    : null;
  if (expandedMeasure) {
    if (measures.has(expandedMeasure)) {
      expandedMeasureName = expandedMeasure;
    } else {
      errors.push(getSingleFieldError("expanded measure", expandedMeasure));
    }
  }

  return {
    tdd: {
      expandedMeasureName,
      chartType,
      pinIndex:
        expandedMeasureName && searchParams.has(ExploreStateURLParams.Pin)
          ? 1
          : -1,
    },
    errors,
  };
}

function fromPivotUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  view: V1ExploreWebView | undefined,
  timeDimension: string | undefined,
): { pivot: PivotState; errors: Error[] } {
  const pivot: PivotState = {
    rows: [],
    columns: [],
    sorting: [],
    // TODO: other fields are not supported right now
    expanded: {},
    columnPage: 1,
    rowPage: 1,
    enableComparison: true,
    activeCell: null,
    showTotalsColumn: true,
    showTotalsRow: true,
    totalsRowPosition: "top",
    tableMode: "nest",
  };
  const errors: Error[] = [];
  if (view !== V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT) {
    return { pivot, errors };
  }

  const toChip = (entry: string): PivotChipData | undefined => {
    if (entry in FromURLParamTimeDimensionMap) {
      const grain = FromURLParamTimeDimensionMap[entry];
      return {
        id: grain,
        title: TIME_GRAIN[grain]?.label,
        type: PivotChipType.Time,
      };
    }

    const measure = measures.get(entry);
    if (measure) {
      return {
        id: entry,
        title: measure.displayName || measure.name || "Unknown",
        type: PivotChipType.Measure,
      };
    }

    const dimension = dimensions.get(entry);
    if (dimension) {
      return {
        id: entry,
        title: dimension.displayName || dimension.name || "Unknown",
        type: PivotChipType.Dimension,
      };
    }

    return undefined;
  };

  const getChips = (
    param: ExploreStateURLParams,
    fieldLabel: string,
    isValid: (entry: string) => boolean,
  ) => {
    // A leading or trailing comma leaves an empty entry: not a field, so skip it.
    const entries =
      searchParams
        .get(param)
        ?.split(",")
        .filter((e) => e !== "") ?? [];
    const validEntries = entries.filter(isValid);
    const missingEntries = getMissingValues(validEntries, entries);
    if (missingEntries.length) {
      errors.push(getMultiFieldError(fieldLabel, missingEntries));
    }
    return validEntries.map(toChip).filter((chip) => !!chip);
  };

  const isTimeEntry = (entry: string) => entry in FromURLParamTimeDimensionMap;
  const rowChips = getChips(
    ExploreStateURLParams.PivotRows,
    "pivot row",
    (r) => dimensions.has(r) || isTimeEntry(r),
  );
  const colChips = getChips(
    ExploreStateURLParams.PivotColumns,
    "pivot column",
    (c) => dimensions.has(c) || measures.has(c) || isTimeEntry(c),
  );
  // A field listed twice, or in both zones, is dropped: the drop zones key
  // their chips by id, so a repeat is a keyed-each error that breaks the view.
  pivot.rows = dedupePivotChips(rowChips);
  pivot.columns = dedupePivotChips(
    colChips,
    pivot.rows.map((chip) => chip.id),
  );

  const tableMode = searchParams.get(ExploreStateURLParams.PivotTableMode);
  if (tableMode) {
    if (tableMode === "nest" || tableMode === "flat") {
      pivot.tableMode = tableMode satisfies PivotTableMode;
    } else {
      errors.push(getSingleFieldError("pivot table mode", tableMode));
    }
  }

  const sortBy = searchParams.get(ExploreStateURLParams.SortBy);
  if (sortBy) {
    let sortById = sortBy;
    if (sortBy in FromURLParamTimeDimensionMap) {
      // Reconstruct TanStack Table format: {timeDimension}_rill_{grain}
      const grain = FromURLParamTimeDimensionMap[sortBy];
      sortById = timeDimension ? `${timeDimension}_rill_${grain}` : grain;
    }
    // Any column id goes through (measures, dimensions, time columns and
    // minimized accessors like "c0v2m0") as long as its field is still in
    // the pivot; the pivot's queries fail on a sort by a removed field.
    if (pivotHasSortField(sortById, pivot)) {
      pivot.sorting = [
        {
          id: sortById,
          desc: searchParams.get(ExploreStateURLParams.SortDirection) !== "ASC",
        },
      ] satisfies SortingState;
    } else {
      errors.push(getSingleFieldError("pivot sort field", sortBy));
    }
  }

  const rowLimit = Number.parseInt(
    searchParams.get(ExploreStateURLParams.PivotRowLimit) ?? "",
    10,
  );
  pivot.rowLimit = (PIVOT_ROW_LIMIT_OPTIONS as readonly number[]).includes(
    rowLimit,
  )
    ? rowLimit
    : undefined;

  const showTotalsColumn = searchParams.get(
    ExploreStateURLParams.PivotShowTotalsColumn,
  );
  if (showTotalsColumn !== null) {
    pivot.showTotalsColumn = showTotalsColumn !== "false";
  }
  const showTotalsRow = searchParams.get(
    ExploreStateURLParams.PivotShowTotalsRow,
  );
  if (showTotalsRow !== null) {
    pivot.showTotalsRow = showTotalsRow !== "false";
  }

  const totalsRowPosition = searchParams.get(
    ExploreStateURLParams.PivotTotalsRowPosition,
  );
  if (totalsRowPosition) {
    if (
      PIVOT_TOTALS_ROW_POSITIONS.includes(
        totalsRowPosition as PivotTotalsRowPosition,
      )
    ) {
      pivot.totalsRowPosition = totalsRowPosition as PivotTotalsRowPosition;
    } else {
      errors.push(
        getSingleFieldError("pivot totals row position", totalsRowPosition),
      );
    }
  }

  const formattingParam = searchParams.get(
    ExploreStateURLParams.PivotFormatting,
  );
  if (formattingParam !== null) {
    const { measureFormatting, invalidEntries } =
      fromPivotFormattingParam(formattingParam);
    // Drop entries for unknown measures; keep the rest.
    const validFormatting = Object.fromEntries(
      Object.entries(measureFormatting ?? {}).filter(([measureName]) => {
        if (measures.has(measureName)) return true;
        invalidEntries.push(measureName);
        return false;
      }),
    );
    if (Object.keys(validFormatting).length) {
      pivot.measureFormatting = validFormatting;
    }
    if (invalidEntries.length) {
      errors.push(getMultiFieldError("pivot formatting", invalidEntries));
    }
  }

  return { pivot, errors };
}

function parseBooleanParam(param: string) {
  switch (param.toLowerCase()) {
    case "true":
    case "1":
      return true;
    case "false":
    case "0":
    case "":
      return false;
    default:
      return undefined;
  }
}
