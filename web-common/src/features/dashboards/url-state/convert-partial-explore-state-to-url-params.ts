import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params";
import {
  compressUrlParams,
  shouldCompressParams,
} from "@rilldata/web-common/features/dashboards/url-state/compression";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { FromLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  ExploreUrlWebView,
  FromActivePageMap,
  ToURLParamSortTypeMap,
  ToURLParamTDDChartMap,
  ToURLParamTimeDimensionMap,
  ToURLParamTimeGrainMapMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  ExploreStateKeyToURLParamMap,
  ExploreStateURLParams,
} from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { arrayOrderedEquals } from "@rilldata/web-common/lib/arrayUtils";
import {
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { type V1ExploreSpec } from "@rilldata/web-common/runtime-client";

/**
 * getCleanedUrlParamsForGoto returns url params with defaults removed.
 * If the params size is greater than a threshold it is compressed as well.
 */
export function getCleanedUrlParamsForGoto(
  exploreSpec: V1ExploreSpec,
  partialExploreState: Partial<ExploreState>,
  timeControlsState: TimeControlState | undefined,
  defaultExploreUrlParams: URLSearchParams,
  urlForCompressionCheck?: URL,
) {
  // Create params from the explore state
  const stateParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    partialExploreState,
    timeControlsState,
  );

  // Clean the url params of any default or empty values.
  const cleanedUrlParams = cleanUrlParams(stateParams, defaultExploreUrlParams);

  if (!urlForCompressionCheck) return cleanedUrlParams;

  // compression
  const urlCopy = new URL(urlForCompressionCheck);
  urlCopy.search = cleanedUrlParams.toString();
  const shouldCompress = shouldCompressParams(urlCopy);
  if (!shouldCompress) return cleanedUrlParams;

  const compressedUrlParams = new URLSearchParams();
  compressedUrlParams.set(
    ExploreStateURLParams.GzippedParams,
    compressUrlParams(cleanedUrlParams.toString()),
  );
  return compressedUrlParams;
}

export function convertPartialExploreStateToUrlParams(
  exploreSpec: V1ExploreSpec,
  partialExploreState: Partial<ExploreState>,
  // We have quite a bit of logic in TimeControlState to validate selections and update them
  // Eg: if a selected grain is not applicable for the current grain then we change it
  // But it is only available in TimeControlState and not MetricsExplorerEntity
  timeControlsState: TimeControlState | undefined,
) {
  const searchParams = new URLSearchParams();

  maybeSetParam(
    searchParams,
    partialExploreState,
    "activePage",
    (ap) =>
      ToURLParamViewMap[
        FromActivePageMap[ap ?? DashboardState_ActivePage.DEFAULT]
      ] ?? ExploreUrlWebView.Explore,
  );

  // timeControlsState will be undefined for dashboards without timeseries
  if (timeControlsState?.selectedTimeRange) {
    copyParamsToTarget(
      toTimeRangesUrl(partialExploreState, timeControlsState),
      searchParams,
    );
  }

  if ("whereFilter" in partialExploreState) {
    const expr = mergeDimensionAndMeasureFilters(
      partialExploreState.whereFilter,
      partialExploreState.dimensionThresholdFilters ?? [],
    );
    let filterParam = "";
    if (expr && expr?.cond?.exprs?.length) {
      filterParam = convertExpressionToFilterParam(
        expr,
        partialExploreState.dimensionsWithInlistFilter,
      );
    }

    searchParams.set(ExploreStateURLParams.Filters, filterParam);
  }

  switch (partialExploreState.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.DIMENSION_TABLE:
    case undefined:
      copyParamsToTarget(
        toExploreUrlParams(partialExploreState, exploreSpec),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      copyParamsToTarget(
        toTimeDimensionUrlParams(partialExploreState),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.PIVOT:
      copyParamsToTarget(toPivotUrlParams(partialExploreState), searchParams);
      // Since we do a shallow merge, we cannot remove time grain from the state for pivot as it is a deeper key.
      // So this is a patch to remove it from the final url.
      searchParams.delete(ExploreStateURLParams.TimeGrain);
      // TODO: fix the need for this once we move out of V1ExplorePreset in converting url to explore state
      searchParams.delete(ExploreStateURLParams.ComparisonDimension);
      break;
  }

  return searchParams;
}

function toTimeRangesUrl(
  partialExploreState: Partial<ExploreState>,
  timeControlsState: TimeControlState,
) {
  const searchParams = new URLSearchParams();

  const timeRangeParam = toTimeRangeParam(timeControlsState.selectedTimeRange);
  searchParams.set(ExploreStateURLParams.TimeRange, timeRangeParam);

  maybeSetParam(searchParams, partialExploreState, "selectedTimezone");

  if ("selectedComparisonTimeRange" in partialExploreState) {
    const compareTimeRangeParam = partialExploreState.showTimeComparison
      ? toTimeRangeParam(timeControlsState.selectedComparisonTimeRange)
      : undefined;
    searchParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      compareTimeRangeParam ?? "",
    );
  }

  if (
    timeControlsState.selectedTimeRange &&
    "interval" in timeControlsState.selectedTimeRange
  ) {
    const mappedTimeGrain =
      ToURLParamTimeGrainMapMap[
        timeControlsState.selectedTimeRange?.interval ?? ""
      ] ?? "";
    searchParams.set(ExploreStateURLParams.TimeGrain, mappedTimeGrain);
  }

  maybeSetParam(
    searchParams,
    partialExploreState,
    "selectedComparisonDimension",
  );

  if (
    partialExploreState.selectedScrubRange &&
    !partialExploreState.selectedScrubRange.isScrubbing
  ) {
    const scrubbingTimeRange = toTimeRangeParam(
      partialExploreState.selectedScrubRange,
    );
    searchParams.set(
      ExploreStateURLParams.HighlightedTimeRange,
      scrubbingTimeRange,
    );
  }

  return searchParams;
}

export function toTimeRangeParam(timeRange: TimeRange | undefined) {
  if (!timeRange) return "";
  if (
    timeRange.name &&
    timeRange.name !== TimeRangePreset.CUSTOM &&
    timeRange.name !== TimeComparisonOption.CUSTOM
  ) {
    return timeRange.name;
  }

  if (!timeRange.start || !timeRange.end) return "";

  return `${timeRange.start.toISOString()},${timeRange.end.toISOString()}`;
}

function toExploreUrlParams(
  partialExploreState: Partial<ExploreState>,
  exploreSpec: V1ExploreSpec,
) {
  const searchParams = new URLSearchParams();

  const visibleMeasuresParam = toVisibleMeasuresUrlParam(
    partialExploreState,
    exploreSpec,
  );
  if (visibleMeasuresParam) {
    searchParams.set(
      ExploreStateURLParams.VisibleMeasures,
      visibleMeasuresParam,
    );
  }

  const visibleDimensionsParam = toVisibleDimensionsUrlParam(
    partialExploreState,
    exploreSpec,
  );
  if (visibleDimensionsParam) {
    searchParams.set(
      ExploreStateURLParams.VisibleDimensions,
      visibleDimensionsParam,
    );
  }

  maybeSetParam(searchParams, partialExploreState, "selectedDimensionName");

  maybeSetParam(
    searchParams,
    partialExploreState,
    "leaderboardSortByMeasureName",
  );

  maybeSetParam(
    searchParams,
    partialExploreState,
    "dashboardSortType",
    // TODO: update mappers
    (st) => ToURLParamSortTypeMap[FromLegacySortTypeMap[st ?? ""]],
  );

  maybeSetParam(searchParams, partialExploreState, "sortDirection", (sd) =>
    sd === SortDirection.ASCENDING ? "ASC" : "DESC",
  );

  maybeSetParam(
    searchParams,
    partialExploreState,
    "leaderboardMeasureNames",
    (names) => names?.join(","),
  );

  maybeSetParam(
    searchParams,
    partialExploreState,
    "leaderboardShowContextForAllMeasures",
    (value) => (value ? "true" : "false"),
  );

  return searchParams;
}

function toVisibleMeasuresUrlParam(
  partialExploreState: Partial<ExploreState>,
  exploreSpec: V1ExploreSpec,
) {
  if (!partialExploreState.visibleMeasures) return undefined;

  if (
    // if the measures are exactly equal to measures from explore then show "*"
    // else the measures are re-ordered, so retain them in url param
    arrayOrderedEquals(
      partialExploreState.visibleMeasures,
      exploreSpec.measures ?? [],
    )
  ) {
    return "*";
  }

  return partialExploreState.visibleMeasures.join(",");
}

function toVisibleDimensionsUrlParam(
  partialExploreState: Partial<ExploreState>,
  exploreSpec: V1ExploreSpec,
) {
  if (!partialExploreState.visibleDimensions) return undefined;

  if (
    // if the dimensions are exactly equal to dimensions from explore then show "*"
    // else the dimensions are re-ordered, so retain them in url param
    arrayOrderedEquals(
      partialExploreState.visibleDimensions,
      exploreSpec.dimensions ?? [],
    )
  ) {
    return "*";
  }

  return partialExploreState.visibleDimensions.join(",");
}

function toTimeDimensionUrlParams(partialExploreState: Partial<ExploreState>) {
  const searchParams = new URLSearchParams();
  if (!partialExploreState.tdd) return searchParams;

  searchParams.set(
    ExploreStateURLParams.ExpandedMeasure,
    partialExploreState.tdd.expandedMeasureName ?? "",
  );

  const chartType = ToURLParamTDDChartMap[partialExploreState.tdd.chartType];
  searchParams.set(ExploreStateURLParams.ChartType, chartType ?? "");

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not
  return searchParams;
}

function toPivotUrlParams(partialExploreState: Partial<ExploreState>) {
  const searchParams = new URLSearchParams();
  if (
    !partialExploreState.pivot ||
    partialExploreState.activePage !== DashboardState_ActivePage.PIVOT
  ) {
    return searchParams;
  }

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  const rows = partialExploreState.pivot.rows.map(mapPivotEntry);
  const rowsParams = rows.join(",");
  searchParams.set(ExploreStateURLParams.PivotRows, rowsParams);

  const cols = partialExploreState.pivot.columns.map(mapPivotEntry);
  const colsParams = cols.join(",");

  searchParams.set(ExploreStateURLParams.PivotColumns, colsParams);

  const sort = partialExploreState.pivot.sorting?.[0];
  const sortId =
    sort?.id in ToURLParamTimeDimensionMap
      ? ToURLParamTimeDimensionMap[sort?.id]
      : sort?.id;

  searchParams.set(ExploreStateURLParams.SortBy, sortId ?? "");

  if (sort) {
    const sortDirParam = sort.desc ? "DESC" : "ASC";
    searchParams.set(ExploreStateURLParams.SortDirection, sortDirParam);
  }

  const tableModeParam = partialExploreState.pivot?.tableMode ?? "nest";
  searchParams.set(ExploreStateURLParams.PivotTableMode, tableModeParam);

  // Only encode rowLimit if it's defined
  if (partialExploreState.pivot?.rowLimit !== undefined) {
    searchParams.set(
      ExploreStateURLParams.PivotRowLimit,
      partialExploreState.pivot.rowLimit.toString(),
    );
  }

  // TODO: other fields like expanded state and pin are not supported right now
  return searchParams;
}

function maybeSetParam<K extends keyof ExploreState>(
  searchParams: URLSearchParams,
  partialExploreState: Partial<ExploreState>,
  key: K,
  mapper: (value: Partial<ExploreState>[K]) => string | undefined = (x) =>
    x?.toString(),
) {
  const param = ExploreStateKeyToURLParamMap[key];
  if (
    // Do not set param if the key was not in explore state. This allows for sparse state conversion.
    !(key in partialExploreState) ||
    !param
  ) {
    return;
  }

  const mappedValue = mapper(partialExploreState[key]);
  searchParams.set(param, mappedValue ?? "");
}
