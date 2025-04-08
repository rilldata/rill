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
  blankExploreUrlParams: URLSearchParams,
  // Used to decide whether to compress or not based on the full url length
  urlForCompressionCheck?: URL,
) {
  const searchParams = new URLSearchParams();

  setParam(
    searchParams,
    partialExploreState,
    "activePage",
    blankExploreUrlParams,
    (ap) =>
      ToURLParamViewMap[
        FromActivePageMap[ap ?? DashboardState_ActivePage.DEFAULT]
      ] ?? ExploreUrlWebView.Explore,
  );

  // timeControlsState will be undefined for dashboards without timeseries
  if (timeControlsState) {
    copyParamsToTarget(
      toTimeRangesUrl(
        partialExploreState,
        timeControlsState,
        blankExploreUrlParams,
      ),
      searchParams,
    );
  }

  if ("whereFilter" in partialExploreState) {
    const expr = mergeDimensionAndMeasureFilters(
      partialExploreState.whereFilter,
      partialExploreState.dimensionThresholdFilters ?? [],
    );
    if (expr && expr?.cond?.exprs?.length) {
      searchParams.set(
        ExploreStateURLParams.Filters,
        convertExpressionToFilterParam(
          expr,
          partialExploreState.dimensionsWithInlistFilter,
        ),
      );
    } else {
      searchParams.set(ExploreStateURLParams.Filters, "");
    }
  }

  switch (partialExploreState.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.DIMENSION_TABLE:
    case undefined:
      copyParamsToTarget(
        toExploreUrl(partialExploreState, blankExploreUrlParams, exploreSpec),
        searchParams,
      );
      break;

    // We dont really need to check blankExploreState for non-explore views since we land on explore.

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
  blankExploreUrlParams: URLSearchParams,
) {
  const searchParams = new URLSearchParams();

  setTimeRangeParam(
    searchParams,
    ExploreStateURLParams.TimeRange,
    timeControlsState.selectedTimeRange,
    blankExploreUrlParams.get(ExploreStateURLParams.TimeRange),
  );

  setParam(
    searchParams,
    partialExploreState,
    "selectedTimezone",
    blankExploreUrlParams,
  );

  if ("selectedComparisonTimeRange" in partialExploreState) {
    // TODO: check showComparison
    setTimeRangeParam(
      searchParams,
      ExploreStateURLParams.ComparisonTimeRange,
      partialExploreState.showTimeComparison
        ? timeControlsState.selectedComparisonTimeRange
        : undefined,
      blankExploreUrlParams.get(ExploreStateURLParams.ComparisonTimeRange),
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
    if (
      shouldSetParamValue(
        mappedTimeGrain,
        blankExploreUrlParams.get(ExploreStateURLParams.TimeGrain),
      )
    ) {
      searchParams.set(ExploreStateURLParams.TimeGrain, mappedTimeGrain);
    }
  }

  setParam(
    searchParams,
    partialExploreState,
    "selectedComparisonDimension",
    blankExploreUrlParams,
  );

  if (
    partialExploreState.selectedScrubRange &&
    !partialExploreState.selectedScrubRange.isScrubbing
  ) {
    setTimeRangeParam(
      searchParams,
      ExploreStateURLParams.HighlightedTimeRange,
      partialExploreState.selectedScrubRange,
      null,
    );
  }

  return searchParams;
}

function setTimeRangeParam(
  searchParams: URLSearchParams,
  param: ExploreStateURLParams,
  timeRange: TimeRange | undefined,
  defaultTimeRange: string | null,
) {
  if (!shouldSetParamValue(timeRange?.name, defaultTimeRange)) return;

  if (
    timeRange?.name &&
    timeRange.name !== TimeRangePreset.CUSTOM &&
    timeRange.name !== TimeComparisonOption.CUSTOM
  ) {
    searchParams.set(param, timeRange.name);
  } else if (!timeRange?.start || !timeRange?.end) {
    searchParams.set(param, "");
  } else {
    searchParams.set(
      param,
      `${timeRange.start.toISOString()},${timeRange.end.toISOString()}`,
    );
  }
}

function toExploreUrl(
  partialExploreState: Partial<MetricsExplorerEntity>,
  blankExploreUrlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  const searchParams = new URLSearchParams();

  const visibleMeasuresParam = toVisibleMeasuresUrlParam(
    partialExploreState,
    blankExploreUrlParams,
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
    blankExploreUrlParams,
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
    blankExploreUrlParams,
  );

  setParam(
    searchParams,
    partialExploreState,
    "leaderboardSortByMeasureName",
    blankExploreUrlParams,
  );

  setParam(
    searchParams,
    partialExploreState,
    "dashboardSortType",
    blankExploreUrlParams,
    // TODO: update mappers
    (st) => ToURLParamSortTypeMap[FromLegacySortTypeMap[st ?? ""]],
  );

  setParam(
    searchParams,
    partialExploreState,
    "sortDirection",
    blankExploreUrlParams,
    (sd) => (sd === SortDirection.ASCENDING ? "ASC" : "DESC"),
  );

  setParam(
    searchParams,
    partialExploreState,
    "leaderboardMeasureCount",
    blankExploreUrlParams,
  );

  return searchParams;
}

function toVisibleMeasuresUrlParam(
  partialExploreState: Partial<MetricsExplorerEntity>,
  blankExploreUrlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  if (!partialExploreState.visibleMeasures) return undefined;

  const defaultVisibleMeasuresParam = blankExploreUrlParams.get(
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
  blankExploreUrlParams: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  if (!partialExploreState.visibleDimensions) return undefined;

  const defaultVisibleDimensionsParam = blankExploreUrlParams.get(
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
) {
  const searchParams = new URLSearchParams();
  if (!partialExploreState.tdd) return searchParams;

  if (partialExploreState.tdd.expandedMeasureName) {
    searchParams.set(
      ExploreStateURLParams.ExpandedMeasure,
      partialExploreState.tdd.expandedMeasureName,
    );
  }

  const chartType = ToURLParamTDDChartMap[partialExploreState.tdd.chartType];
  searchParams.set(ExploreStateURLParams.ChartType, chartType ?? "");

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not
  return searchParams;
}

function toPivotUrlParams(partialExploreState: Partial<MetricsExplorerEntity>) {
  const searchParams = new URLSearchParams();
  if (!partialExploreState.pivot?.active) return searchParams;

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  const rows = partialExploreState.pivot.rows.map(mapPivotEntry);
  searchParams.set(ExploreStateURLParams.PivotRows, rows.join(","));

  const cols = partialExploreState.pivot.columns.map(mapPivotEntry);
  searchParams.set(ExploreStateURLParams.PivotColumns, cols.join(","));

  const sort = partialExploreState.pivot.sorting?.[0];
  const sortId =
    sort?.id in ToURLParamTimeDimensionMap
      ? ToURLParamTimeDimensionMap[sort?.id]
      : sort?.id;

  searchParams.set(ExploreStateURLParams.SortBy, sortId ?? "");
  if (sort) {
    searchParams.set(
      ExploreStateURLParams.SortDirection,
      sort.desc ? "DESC" : "ASC",
    );
  }

  const tableMode = partialExploreState.pivot?.tableMode;
  searchParams.set(ExploreStateURLParams.PivotTableMode, tableMode ?? "nest");

  // TODO: other fields like expanded state and pin are not supported right now
  return searchParams;
}

function setParam<K extends keyof MetricsExplorerEntity>(
  searchParams: URLSearchParams,
  partialExploreState: Partial<MetricsExplorerEntity>,
  key: K,
  blankExploreUrlParams: URLSearchParams,
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
  if (!shouldSetParamValue(mappedValue, blankExploreUrlParams.get(param))) {
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
