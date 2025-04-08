import { stripMeasureSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
  type PivotTableMode,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  filterIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { decompressUrlParams } from "@rilldata/web-common/features/dashboards/url-state/compression";
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
  TIME_COMPARISON,
  TIME_GRAIN,
} from "@rilldata/web-common/lib/time/config";
import { validateISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type { SortingState } from "@tanstack/svelte-table";

export function convertUrlSearchToPartialExploreState(
  searchParams: URLSearchParams,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
) {
  const partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  const measures = getMapFromArray(
    metricsViewSpec.measures?.filter((m) =>
      exploreSpec.measures?.includes(m.name!),
    ) ?? [],
    (m) => m.name!,
  );
  const dimensions = getMapFromArray(
    metricsViewSpec.dimensions?.filter((d) =>
      exploreSpec.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );

  if (searchParams.has(ExploreStateURLParams.GzippedParams)) {
    searchParams = new URLSearchParams(
      decompressUrlParams(
        searchParams.get(ExploreStateURLParams.GzippedParams)!,
      ),
    );
  }

  // TODO: legacy dashboard param.

  if (searchParams.has(ExploreStateURLParams.WebView)) {
    const view = searchParams.get(ExploreStateURLParams.WebView) as string;
    if (view in FromURLParamViewMap) {
      partialExploreState.activePage = Number(
        ToActivePageViewMap[FromURLParamViewMap[view]] ?? "0",
      );
    } else {
      errors.push(getSingleFieldError("view", view));
    }
  }

  if (searchParams.has(ExploreStateURLParams.Filters)) {
    const {
      expr,
      dimensionsWithInlistFilter,
      errors: filterErrors,
    } = fromFilterUrlParam(
      searchParams.get(ExploreStateURLParams.Filters) as string,
      measures,
      dimensions,
    );
    if (filterErrors) errors.push(...filterErrors);
    if (expr) {
      const { dimensionFilters, dimensionThresholdFilters } =
        splitWhereFilter(expr); // TODO: split in fromFilterUrlParam itself
      partialExploreState.whereFilter = dimensionFilters;
      partialExploreState.dimensionThresholdFilters = dimensionThresholdFilters;
    }
    if (dimensionsWithInlistFilter) {
      partialExploreState.dimensionsWithInlistFilter =
        dimensionsWithInlistFilter;
    }
  }

  const { partialExploreState: trPartialExploreState, errors: trErrors } =
    fromTimeRangesParams(searchParams, dimensions);
  Object.assign(partialExploreState, trPartialExploreState);
  errors.push(...trErrors);

  // only extract params if the view is explicitly set to the relevant one
  switch (partialExploreState.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.DIMENSION_TABLE: {
      const { partialExploreState: ovPartialExploreState, errors: ovErrors } =
        fromExploreUrlParams(searchParams, measures, dimensions, exploreSpec);
      Object.assign(partialExploreState, ovPartialExploreState);
      errors.push(...ovErrors);
      break;
    }

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL: {
      const { partialExploreState: tddPartialExploreState, errors: tddErrors } =
        fromTimeDimensionUrlParams(searchParams, measures);
      Object.assign(partialExploreState, tddPartialExploreState);
      errors.push(...tddErrors);
      break;
    }

    case DashboardState_ActivePage.PIVOT: {
      const {
        partialExploreState: pivotPartialExploreState,
        errors: pivotErrors,
      } = fromPivotUrlParams(searchParams, measures, dimensions);
      Object.assign(partialExploreState, pivotPartialExploreState);
      errors.push(...pivotErrors);
      break;
    }
  }

  return { partialExploreState, errors };
}

function fromFilterUrlParam(
  filter: string,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
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

function fromTimeRangesParams(
  searchParams: URLSearchParams,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.TimeRange)) {
    const tr = searchParams.get(ExploreStateURLParams.TimeRange) as string;
    if (tr === "") {
      partialExploreState.selectedTimeRange = undefined;
    } else if (
      tr in FromURLParamTimeRangePresetMap ||
      validateISODuration(tr) ||
      CustomTimeRangeRegex.test(tr)
    ) {
      partialExploreState.selectedTimeRange = fromTimeRangeUrlParam(tr);
    } else {
      errors.push(getSingleFieldError("time range", tr));
    }
  }

  if (searchParams.has(ExploreStateURLParams.TimeZone)) {
    partialExploreState.selectedTimezone = searchParams.get(
      ExploreStateURLParams.TimeZone,
    ) as string;
  }

  if (searchParams.has(ExploreStateURLParams.ComparisonTimeRange)) {
    const ctr = searchParams.get(
      ExploreStateURLParams.ComparisonTimeRange,
    ) as string;
    if (ctr in TIME_COMPARISON || CustomTimeRangeRegex.test(ctr)) {
      partialExploreState.selectedComparisonTimeRange =
        fromTimeRangeUrlParam(ctr);
      partialExploreState.showTimeComparison = true;
    } else if (ctr == "") {
      partialExploreState.selectedComparisonTimeRange = undefined;
      partialExploreState.showTimeComparison = false;
    } else {
      errors.push(getSingleFieldError("compare time range", ctr));
    }
  }

  if (
    searchParams.has(ExploreStateURLParams.TimeGrain) &&
    partialExploreState.selectedTimeRange
  ) {
    const tg = searchParams.get(ExploreStateURLParams.TimeGrain) as string;
    if (tg in FromURLParamTimeGrainMap) {
      partialExploreState.selectedTimeRange.interval =
        FromURLParamTimeGrainMap[tg];
    } else {
      errors.push(getSingleFieldError("time grain", tg));
    }
  }

  if (searchParams.has(ExploreStateURLParams.ComparisonDimension)) {
    const comparisonDimension = searchParams.get(
      ExploreStateURLParams.ComparisonDimension,
    ) as string;
    // unsetting a default from url
    if (comparisonDimension === "") {
      partialExploreState.selectedComparisonDimension = "";
    } else if (dimensions.has(comparisonDimension)) {
      partialExploreState.selectedComparisonDimension = comparisonDimension;
    } else {
      errors.push(
        getSingleFieldError("compare dimension", comparisonDimension),
      );
    }
  }

  if (searchParams.has(ExploreStateURLParams.HighlightedTimeRange)) {
    const selectTr = searchParams.get(
      ExploreStateURLParams.HighlightedTimeRange,
    ) as string;
    if (CustomTimeRangeRegex.test(selectTr) || selectTr === "") {
      partialExploreState.selectedScrubRange = {
        ...fromTimeRangeUrlParam(selectTr),
        isScrubbing: false,
      };
    } else {
      errors.push(getSingleFieldError("highlighted time range", selectTr));
    }
  }
  return { partialExploreState, errors };
}

const CustomTimeRangeRegex =
  /(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z),(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z)/;
function fromTimeRangeUrlParam(tr: string) {
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
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
  exploreSpec: V1ExploreSpec,
) {
  const partialExploreState: Partial<MetricsExplorerEntity> = {};
  const errors: Error[] = [];

  if (searchParams.has(ExploreStateURLParams.VisibleMeasures)) {
    const mes = searchParams.get(
      ExploreStateURLParams.VisibleMeasures,
    ) as string;
    if (mes === "*") {
      partialExploreState.visibleMeasures = exploreSpec.measures ?? [];
      partialExploreState.allMeasuresVisible = true;
    } else {
      const selectedMeasures = mes.split(",").filter((m) => measures.has(m));
      partialExploreState.visibleMeasures = selectedMeasures;
      partialExploreState.allMeasuresVisible = false;

      const missingMeasures = getMissingValues(
        selectedMeasures,
        mes.split(","),
      );
      if (missingMeasures.length) {
        errors.push(getMultiFieldError("measure", missingMeasures));
      }
    }
  }

  if (searchParams.has(ExploreStateURLParams.VisibleDimensions)) {
    const dims = searchParams.get(
      ExploreStateURLParams.VisibleDimensions,
    ) as string;
    if (dims === "*") {
      partialExploreState.visibleDimensions = exploreSpec.dimensions ?? [];
      partialExploreState.allDimensionsVisible = true;
    } else {
      const selectedDimensions = dims
        .split(",")
        .filter((d) => dimensions.has(d));
      partialExploreState.visibleDimensions = selectedDimensions;
      partialExploreState.allDimensionsVisible = false;

      const missingDimensions = getMissingValues(
        selectedDimensions,
        dims.split(","),
      );
      if (missingDimensions.length) {
        errors.push(getMultiFieldError("dimension", missingDimensions));
      }
    }
  }

  if (searchParams.has(ExploreStateURLParams.ExpandedDimension)) {
    const dim = searchParams.get(
      ExploreStateURLParams.ExpandedDimension,
    ) as string;
    if (dimensions.has(dim)) {
      partialExploreState.selectedDimensionName = dim;
      partialExploreState.activePage =
        DashboardState_ActivePage.DIMENSION_TABLE;
    } else if (dim === "") {
      partialExploreState.selectedDimensionName = "";
      partialExploreState.activePage = DashboardState_ActivePage.DEFAULT;
    } else {
      errors.push(getSingleFieldError("expanded dimension", dim));
    }
  }

  if (searchParams.has(ExploreStateURLParams.SortBy)) {
    const sortBy = searchParams.get(ExploreStateURLParams.SortBy) as string;
    if (measures.has(sortBy)) {
      if (
        (partialExploreState.visibleMeasures &&
          partialExploreState.visibleMeasures.includes(sortBy)) ||
        !partialExploreState.visibleMeasures
      ) {
        partialExploreState.leaderboardSortByMeasureName = sortBy;
      } else {
        partialExploreState.leaderboardSortByMeasureName =
          partialExploreState.visibleMeasures?.[0] ??
          exploreSpec.measures?.[0] ??
          "";
        errors.push(
          getSingleFieldError("sort by measure", sortBy, "It is hidden."),
        );
      }
    } else {
      errors.push(getSingleFieldError("sort by measure", sortBy));
    }
  }

  if (searchParams.has(ExploreStateURLParams.SortDirection)) {
    const sortDirection = searchParams.get(
      ExploreStateURLParams.SortDirection,
    ) as string;
    partialExploreState.sortDirection =
      sortDirection === "ASC"
        ? DashboardState_LeaderboardSortDirection.ASCENDING
        : DashboardState_LeaderboardSortDirection.DESCENDING;
  }

  if (searchParams.has(ExploreStateURLParams.SortType)) {
    const sortType = searchParams.get(ExploreStateURLParams.SortType) as string;
    if (sortType in FromURLParamsSortTypeMap) {
      partialExploreState.dashboardSortType =
        Number(
          ToLegacySortTypeMap[FromURLParamsSortTypeMap[sortType]] ?? "0",
        ) ?? DashboardState_LeaderboardSortType.UNSPECIFIED;
    } else {
      errors.push(getSingleFieldError("sort type", sortType));
    }
  }

  if (searchParams.has(ExploreStateURLParams.LeaderboardMeasureCount)) {
    const count = searchParams.get(
      ExploreStateURLParams.LeaderboardMeasureCount,
    );
    const parsedCount = parseInt(count ?? "", 10);
    if (!isNaN(parsedCount) && parsedCount > 0) {
      partialExploreState.leaderboardMeasureCount = parsedCount;
    } else {
      errors.push(
        getSingleFieldError("leaderboard measure count", count ?? ""),
      );
    }
  }

  return { partialExploreState, errors };
}

function fromTimeDimensionUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
): {
  partialExploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
  const errors: Error[] = [];

  let expandedMeasureName = "";
  let chartType = TDDChart.DEFAULT;

  if (searchParams.has(ExploreStateURLParams.ExpandedMeasure)) {
    const mes = searchParams.get(
      ExploreStateURLParams.ExpandedMeasure,
    ) as string;
    if (measures.has(mes) || mes === "") {
      expandedMeasureName = mes;
    } else {
      errors.push(getSingleFieldError("expanded measure", mes));
    }
  }

  if (searchParams.has(ExploreStateURLParams.ChartType)) {
    const urlCharType = searchParams.get(
      ExploreStateURLParams.ChartType,
    ) as string;
    if (urlCharType in FromURLParamTDDChartMap) {
      chartType = FromURLParamTDDChartMap[urlCharType];
    } else {
      errors.push(getSingleFieldError("chart type", urlCharType));
    }
  }

  // TODO: ExploreStateURLParams.Pin

  return {
    partialExploreState: {
      activePage: expandedMeasureName
        ? DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL
        : DashboardState_ActivePage.DEFAULT,
      tdd: {
        expandedMeasureName,
        chartType,
        pinIndex: -1,
      },
    },
    errors,
  };
}

function fromPivotUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
): {
  partialExploreState: Partial<MetricsExplorerEntity>;
  errors: Error[];
} {
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

  let hasSomePivotFields = false;

  const rowChips: PivotChipData[] = [];
  if (searchParams.has(ExploreStateURLParams.PivotRows)) {
    const rows = (
      searchParams.get(ExploreStateURLParams.PivotRows) as string
    ).split(",");
    rows.forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      rowChips.push(chip);
    });
    hasSomePivotFields = true;
  }

  const colChips: PivotChipData[] = [];
  if (searchParams.has(ExploreStateURLParams.PivotColumns)) {
    const cols = (
      searchParams.get(ExploreStateURLParams.PivotColumns) as string
    ).split(",");
    cols.forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      colChips.push(chip);
    });
    hasSomePivotFields = true;
  }

  if (!hasSomePivotFields) {
    return {
      partialExploreState: {
        activePage: DashboardState_ActivePage.DEFAULT,
        pivot: {
          active: false,
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

  let sortDesc = false;
  if (searchParams.has(ExploreStateURLParams.SortDirection)) {
    const sortDescParam = searchParams.get(
      ExploreStateURLParams.SortDirection,
    ) as string;
    sortDesc = sortDescParam === "DESC";
  }

  const sorting: SortingState = [];
  if (searchParams.has(ExploreStateURLParams.SortBy)) {
    const sortBy = searchParams.get(ExploreStateURLParams.SortBy) as string;

    const sortById =
      sortBy in FromURLParamTimeDimensionMap
        ? FromURLParamTimeDimensionMap[sortBy]
        : sortBy;
    sorting.push({
      id: sortById,
      desc: sortDesc,
    });
  }

  let tableMode: PivotTableMode = "nest";
  if (searchParams.has(ExploreStateURLParams.PivotTableMode)) {
    const tableModeParam = searchParams.get(
      ExploreStateURLParams.PivotTableMode,
    ) as string;
    if (tableModeParam === "nest" || tableModeParam === "flat") {
      tableMode = tableModeParam;
    } else {
      errors.push(getSingleFieldError("pivot table mode", tableModeParam));
    }
  }

  return {
    partialExploreState: {
      pivot: {
        active: true,
        rows: rowChips,
        columns: colChips,
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
