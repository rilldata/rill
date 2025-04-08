import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
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

export function convertPartialExploreStateToUrlSearch(
  partialExploreState: Partial<MetricsExplorerEntity>,
  exploreSpec: V1ExploreSpec,
  // We have quite a bit of logic in TimeControlState to validate selections and update them
  // Eg: if a selected grain is not applicable for the current grain then we change it
  // But it is only available in TimeControlState and not MetricsExplorerEntity
  timeControlsState: TimeControlState | undefined,
  defaultExploreUrlParams: URLSearchParams,
  // Used to decide whether to compress or not based on the full url length
  urlForCompressionCheck?: URL,
) {
  const searchParams = new URLSearchParams();

  setParam(
    searchParams,
    partialExploreState,
    "activePage",
    defaultExploreUrlParams,
    (ap) =>
      ToURLParamViewMap[
        FromActivePageMap[ap ?? DashboardState_ActivePage.DEFAULT]
      ] ?? ExploreUrlWebView.Explore,
  );

  // timeControlsState will be undefined for dashboards without timeseries
  if (timeControlsState?.selectedTimeRange) {
    copyParamsToTarget(
      toTimeRangesUrl(
        partialExploreState,
        timeControlsState,
        defaultExploreUrlParams,
      ),
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

    if (
      shouldSetParamValue(
        filterParam,
        defaultExploreUrlParams.get(ExploreStateURLParams.Filters),
      )
    ) {
      searchParams.set(ExploreStateURLParams.Filters, filterParam);
    }
  }

  switch (partialExploreState.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.DIMENSION_TABLE:
    case undefined:
      copyParamsToTarget(
        toExploreUrl(partialExploreState, defaultExploreUrlParams, exploreSpec),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      copyParamsToTarget(
        toTimeDimensionUrlParams(partialExploreState, defaultExploreUrlParams),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.PIVOT:
      copyParamsToTarget(
        toPivotUrlParams(partialExploreState, defaultExploreUrlParams),
        searchParams,
      );
      // Since we do a shallow merge, we cannot remove time grain from the state for pivot as it is a deeper key.
      // So this is a patch to remove it from the final url.
      searchParams.delete(ExploreStateURLParams.TimeGrain);
      break;
  }

  if (!urlForCompressionCheck) return searchParams;

  const urlCopy = new URL(urlForCompressionCheck);
  urlCopy.search = searchParams.toString();
  const shouldCompress = shouldCompressParams(urlCopy);
  if (!shouldCompress) return searchParams;

  const compressedUrlParams = new URLSearchParams();
  compressedUrlParams.set(
    ExploreStateURLParams.GzippedParams,
    compressUrlParams(searchParams.toString()),
  );
  return compressedUrlParams;
}

function toTimeRangesUrl(
  partialExploreState: Partial<MetricsExplorerEntity>,
  timeControlsState: TimeControlState,
  defaultExploreUrlParams: URLSearchParams,
) {
  const searchParams = new URLSearchParams();

  const timeRangeParam = toTimeRangeParam(timeControlsState.selectedTimeRange);
  if (
    shouldSetParamValue(
      timeRangeParam,
      defaultExploreUrlParams.get(ExploreStateURLParams.TimeRange),
    )
  ) {
    searchParams.set(ExploreStateURLParams.TimeRange, timeRangeParam);
  }

  setParam(
    searchParams,
    partialExploreState,
    "selectedTimezone",
    defaultExploreUrlParams,
  );

  if ("selectedComparisonTimeRange" in partialExploreState) {
    const compareTimeRangeParam = partialExploreState.showTimeComparison
      ? toTimeRangeParam(timeControlsState.selectedComparisonTimeRange)
      : undefined;
    if (
      shouldSetParamValue(
        compareTimeRangeParam,
        defaultExploreUrlParams.get(ExploreStateURLParams.ComparisonTimeRange),
      )
    ) {
      searchParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        compareTimeRangeParam ?? "",
      );
    }
  }

  if (
    timeControlsState.selectedTimeRange &&
    "interval" in timeControlsState.selectedTimeRange
  ) {
    const mappedTimeGrain =
      ToURLParamTimeGrainMapMap[
        timeControlsState.selectedTimeRange?.interval ?? ""
      ] ?? "";
    if (
      shouldSetParamValue(
        mappedTimeGrain,
        defaultExploreUrlParams.get(ExploreStateURLParams.TimeGrain),
      )
    ) {
      searchParams.set(ExploreStateURLParams.TimeGrain, mappedTimeGrain);
    }
  }

  setParam(
    searchParams,
    partialExploreState,
    "selectedComparisonDimension",
    defaultExploreUrlParams,
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

function toExploreUrl(
  partialExploreState: Partial<MetricsExplorerEntity>,
  defaultExploreUrlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  const searchParams = new URLSearchParams();

  const visibleMeasuresParam = toVisibleMeasuresUrlParam(
    partialExploreState,
    defaultExploreUrlParams,
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
    defaultExploreUrlParams,
    exploreSpec,
  );
  if (visibleDimensionsParam) {
    searchParams.set(
      ExploreStateURLParams.VisibleDimensions,
      visibleDimensionsParam,
    );
  }

  setParam(
    searchParams,
    partialExploreState,
    "selectedDimensionName",
    defaultExploreUrlParams,
  );

  setParam(
    searchParams,
    partialExploreState,
    "leaderboardSortByMeasureName",
    defaultExploreUrlParams,
  );

  setParam(
    searchParams,
    partialExploreState,
    "dashboardSortType",
    defaultExploreUrlParams,
    // TODO: update mappers
    (st) => ToURLParamSortTypeMap[FromLegacySortTypeMap[st ?? ""]],
  );

  setParam(
    searchParams,
    partialExploreState,
    "sortDirection",
    defaultExploreUrlParams,
    (sd) => (sd === SortDirection.ASCENDING ? "ASC" : "DESC"),
  );

  setParam(
    searchParams,
    partialExploreState,
    "leaderboardMeasureCount",
    defaultExploreUrlParams,
  );

  return searchParams;
}

function toVisibleMeasuresUrlParam(
  partialExploreState: Partial<MetricsExplorerEntity>,
  defaultExploreUrlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  if (!partialExploreState.visibleMeasures) return undefined;

  const defaultVisibleMeasuresParam = defaultExploreUrlParams.get(
    ExploreStateURLParams.VisibleMeasures,
  );
  const defaultVisibleMeasures =
    defaultVisibleMeasuresParam === "*"
      ? (exploreSpec.measures ?? [])
      : (defaultVisibleMeasuresParam?.split(",") ?? []);
  if (
    // if the measures are exactly equal to defaults then do not add a param
    arrayOrderedEquals(
      partialExploreState.visibleMeasures,
      defaultVisibleMeasures,
    )
  ) {
    return undefined;
  }

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
  partialExploreState: Partial<MetricsExplorerEntity>,
  defaultExploreUrlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  if (!partialExploreState.visibleDimensions) return undefined;

  const defaultVisibleDimensionsParam = defaultExploreUrlParams.get(
    ExploreStateURLParams.VisibleDimensions,
  );
  const defaultVisibleDimensions =
    defaultVisibleDimensionsParam === "*"
      ? (exploreSpec.dimensions ?? [])
      : (defaultVisibleDimensionsParam?.split(",") ?? []);
  if (
    // if the dimensions are exactly equal to defaults then do not add a param
    arrayOrderedEquals(
      partialExploreState.visibleDimensions,
      defaultVisibleDimensions,
    )
  ) {
    return undefined;
  }

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

function toTimeDimensionUrlParams(
  partialExploreState: Partial<MetricsExplorerEntity>,
  defaultExploreUrlParams: URLSearchParams,
) {
  const searchParams = new URLSearchParams();
  if (!partialExploreState.tdd) return searchParams;

  if (
    shouldSetParamValue(
      partialExploreState.tdd.expandedMeasureName,
      defaultExploreUrlParams.get(ExploreStateURLParams.ExpandedMeasure),
    )
  ) {
    searchParams.set(
      ExploreStateURLParams.ExpandedMeasure,
      partialExploreState.tdd.expandedMeasureName ?? "",
    );
  }

  const chartType = ToURLParamTDDChartMap[partialExploreState.tdd.chartType];
  if (
    shouldSetParamValue(
      chartType,
      defaultExploreUrlParams.get(ExploreStateURLParams.ChartType),
    )
  ) {
    searchParams.set(ExploreStateURLParams.ChartType, chartType ?? "");
  }

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not
  return searchParams;
}

function toPivotUrlParams(
  partialExploreState: Partial<MetricsExplorerEntity>,
  defaultExploreUrlParams: URLSearchParams,
) {
  const searchParams = new URLSearchParams();
  if (!partialExploreState.pivot?.active) return searchParams;

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  const rows = partialExploreState.pivot.rows.map(mapPivotEntry);
  const rowsParams = rows.join(",");
  if (
    shouldSetParamValue(
      rowsParams,
      defaultExploreUrlParams.get(ExploreStateURLParams.PivotRows),
    )
  ) {
    searchParams.set(ExploreStateURLParams.PivotRows, rowsParams);
  }

  const cols = partialExploreState.pivot.columns.map(mapPivotEntry);
  const colsParams = cols.join(",");
  if (
    shouldSetParamValue(
      colsParams,
      defaultExploreUrlParams.get(ExploreStateURLParams.PivotColumns),
    )
  ) {
    searchParams.set(ExploreStateURLParams.PivotColumns, colsParams);
  }

  const sort = partialExploreState.pivot.sorting?.[0];
  const sortId =
    sort?.id in ToURLParamTimeDimensionMap
      ? ToURLParamTimeDimensionMap[sort?.id]
      : sort?.id;

  if (
    shouldSetParamValue(
      sortId,
      defaultExploreUrlParams.get(ExploreStateURLParams.SortBy),
    )
  ) {
    searchParams.set(ExploreStateURLParams.SortBy, sortId ?? "");
  }

  if (sort) {
    const sortDirParam = sort.desc ? "DESC" : "ASC";
    if (
      shouldSetParamValue(
        sortDirParam,
        defaultExploreUrlParams.get(ExploreStateURLParams.SortDirection),
      )
    ) {
      searchParams.set(ExploreStateURLParams.SortDirection, sortDirParam);
    }
  }

  const tableModeParam = partialExploreState.pivot?.tableMode ?? "nest";
  if (
    shouldSetParamValue(
      tableModeParam,
      defaultExploreUrlParams.get(ExploreStateURLParams.PivotTableMode),
    )
  ) {
    searchParams.set(ExploreStateURLParams.PivotTableMode, tableModeParam);
  }

  // TODO: other fields like expanded state and pin are not supported right now
  return searchParams;
}

function setParam<K extends keyof MetricsExplorerEntity>(
  searchParams: URLSearchParams,
  partialExploreState: Partial<MetricsExplorerEntity>,
  key: K,
  defaultExploreUrlParams: URLSearchParams,
  mapper: (value: Partial<MetricsExplorerEntity>[K]) => string | undefined = (
    x,
  ) => x?.toString(),
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
  if (!shouldSetParamValue(mappedValue, defaultExploreUrlParams.get(param))) {
    return;
  }

  searchParams.set(param, mappedValue ?? "");
}

function shouldSetParamValue<T>(
  exploreStateValue: T | undefined,
  defaultValue: string | null,
) {
  // Always set the param if there is no default.
  if (defaultValue === null) {
    return true;
  }

  // If exploreStateValue is falsy and defaultValue is "" then do not set a param.
  if (defaultValue === "" && !exploreStateValue) return false;

  const exploreStateValueStr = exploreStateValue?.toString();
  return defaultValue !== exploreStateValueStr;
}
