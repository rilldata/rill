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
  FromActivePageMap,
  ToURLParamSortTypeMap,
  ToURLParamTDDChartMap,
  ToURLParamTimeDimensionMap,
  ToURLParamTimeGrainMapMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import { arrayOrderedEquals } from "@rilldata/web-common/lib/arrayUtils";
import {
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

export function convertExploreStateToURLSearchParams(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  // We have quite a bit of logic in TimeControlState to validate selections and update them
  // Eg: if a selected grain is not applicable for the current grain then we change it
  // But it is only available in TimeControlState and not MetricsExplorerEntity
  timeControlsState: TimeControlState | undefined,
  preset: V1ExplorePreset,
  // Used to decide whether to compress or not based on the full url length
  urlForCompressionCheck?: URL,
): URLSearchParams {
  const searchParams = new URLSearchParams();

  if (!exploreState) return searchParams;

  const currentView = FromActivePageMap[exploreState.activePage];
  if (shouldSetParam(preset.view, currentView)) {
    searchParams.set(
      ExploreStateURLParams.WebView,
      ToURLParamViewMap[currentView] as string,
    );
  }

  // timeControlsState will be undefined for dashboards without timeseries
  if (timeControlsState) {
    copyParamsToTarget(
      toTimeRangesUrl(exploreState, timeControlsState, preset),
      searchParams,
    );
  }

  const expr = mergeDimensionAndMeasureFilters(
    exploreState.whereFilter,
    exploreState.dimensionThresholdFilters,
  );
  if (expr && expr?.cond?.exprs?.length) {
    searchParams.set(
      ExploreStateURLParams.Filters,
      convertExpressionToFilterParam(
        expr,
        exploreState.dimensionsWithInlistFilter,
      ),
    );
  }

  switch (exploreState.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.DIMENSION_TABLE:
      copyParamsToTarget(
        toExploreUrl(exploreState, exploreSpec, preset),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      copyParamsToTarget(
        toTimeDimensionUrlParams(exploreState, preset),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.PIVOT:
      copyParamsToTarget(toPivotUrlParams(exploreState, preset), searchParams);
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
  exploreState: MetricsExplorerEntity,
  timeControlsState: TimeControlState,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (
    shouldSetParam(preset.timeRange, timeControlsState.selectedTimeRange?.name)
  ) {
    searchParams.set(
      ExploreStateURLParams.TimeRange,
      toTimeRangeParam(timeControlsState.selectedTimeRange),
    );
  }

  if (shouldSetParam(preset.timezone, exploreState.selectedTimezone)) {
    searchParams.set(
      ExploreStateURLParams.TimeZone,
      exploreState.selectedTimezone,
    );
  }

  if (
    exploreState.showTimeComparison &&
    ((preset.compareTimeRange !== undefined &&
      timeControlsState.selectedComparisonTimeRange !== undefined &&
      timeControlsState.selectedComparisonTimeRange.name !==
        preset.compareTimeRange) ||
      preset.compareTimeRange === undefined)
  ) {
    searchParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      toTimeRangeParam(timeControlsState.selectedComparisonTimeRange),
    );
  } else if (
    !exploreState.showTimeComparison &&
    preset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
  ) {
    searchParams.set(ExploreStateURLParams.ComparisonTimeRange, "");
  }

  const mappedTimeGrain =
    ToURLParamTimeGrainMapMap[
      timeControlsState.selectedTimeRange?.interval ?? ""
    ] ?? "";

  if (mappedTimeGrain && shouldSetParam(preset.timeGrain, mappedTimeGrain)) {
    searchParams.set(ExploreStateURLParams.TimeGrain, mappedTimeGrain);
  }

  if (
    shouldSetParam(
      preset.comparisonDimension,
      exploreState.selectedComparisonDimension,
    )
  ) {
    // TODO: move this based on expected param sequence
    searchParams.set(
      ExploreStateURLParams.ComparisonDimension,
      exploreState.selectedComparisonDimension ?? "",
    );
  }

  if (
    exploreState.selectedScrubRange &&
    !exploreState.selectedScrubRange.isScrubbing
  ) {
    searchParams.set(
      ExploreStateURLParams.HighlightedTimeRange,
      toTimeRangeParam(exploreState.selectedScrubRange),
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
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  const visibleMeasuresParam = toVisibleMeasuresUrlParam(
    exploreState,
    exploreSpec,
    preset,
  );
  if (visibleMeasuresParam) {
    searchParams.set(
      ExploreStateURLParams.VisibleMeasures,
      visibleMeasuresParam,
    );
  }

  const visibleDimensionsParam = toVisibleDimensionsUrlParam(
    exploreState,
    exploreSpec,
    preset,
  );
  if (visibleDimensionsParam) {
    searchParams.set(
      ExploreStateURLParams.VisibleDimensions,
      visibleDimensionsParam,
    );
  }

  if (
    shouldSetParam(
      preset.exploreExpandedDimension,
      exploreState.selectedDimensionName,
    )
  ) {
    searchParams.set(
      ExploreStateURLParams.ExpandedDimension,
      exploreState.selectedDimensionName ?? "",
    );
  }

  if (
    shouldSetParam(
      preset.exploreSortBy,
      exploreState.leaderboardSortByMeasureName,
    )
  ) {
    searchParams.set(
      ExploreStateURLParams.SortBy,
      exploreState.leaderboardSortByMeasureName,
    );
  }

  const sortType = FromLegacySortTypeMap[exploreState.dashboardSortType];
  if (shouldSetParam(preset.exploreSortType, sortType)) {
    searchParams.set(
      ExploreStateURLParams.SortType,
      ToURLParamSortTypeMap[sortType] ?? "",
    );
  }

  const sortAsc = exploreState.sortDirection === SortDirection.ASCENDING;
  if (shouldSetParam(preset.exploreSortAsc, sortAsc)) {
    searchParams.set(
      ExploreStateURLParams.SortDirection,
      sortAsc ? "ASC" : "DESC",
    );
  }

  if (
    shouldSetParam(
      preset.exploreLeaderboardMeasureCount,
      exploreState.leaderboardMeasureCount,
    )
  ) {
    searchParams.set(
      ExploreStateURLParams.LeaderboardMeasureCount,
      exploreState.leaderboardMeasureCount?.toString(),
    );
  }

  return searchParams;
}

function toVisibleMeasuresUrlParam(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!exploreState.visibleMeasures) return undefined;

  const presetMeasures = preset.measures ?? exploreSpec.measures ?? [];
  if (arrayOrderedEquals(exploreState.visibleMeasures, presetMeasures)) {
    return undefined;
  }
  if (
    // if the measures are exactly equal to measures from explore then show "*"
    // else the measures are re-ordered, so retain them in url param
    arrayOrderedEquals(exploreState.visibleMeasures, exploreSpec.measures ?? [])
  ) {
    return "*";
  }
  return exploreState.visibleMeasures.join(",");
}

function toVisibleDimensionsUrlParam(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!exploreState.visibleDimensions) return undefined;

  const presetDimensions = preset.dimensions ?? exploreSpec.dimensions ?? [];
  if (arrayOrderedEquals(exploreState.visibleDimensions, presetDimensions)) {
    return undefined;
  }
  if (
    // if the dimensions are exactly equal to dimensions from explore then show "*"
    // else the dimensions are re-ordered, so retain them in url param
    arrayOrderedEquals(
      exploreState.visibleDimensions,
      exploreSpec.dimensions ?? [],
    )
  ) {
    return "*";
  }
  return exploreState.visibleDimensions.join(",");
}

function toTimeDimensionUrlParams(
  exploreState: MetricsExplorerEntity,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();
  if (!exploreState.tdd) return searchParams;

  if (
    shouldSetParam(
      preset.timeDimensionMeasure,
      exploreState.tdd.expandedMeasureName,
    )
  ) {
    searchParams.set(
      ExploreStateURLParams.ExpandedMeasure,
      exploreState.tdd.expandedMeasureName ?? "",
    );
  }

  const chartType = ToURLParamTDDChartMap[exploreState.tdd.chartType];
  if (shouldSetParam(preset.timeDimensionChartType, chartType)) {
    searchParams.set(ExploreStateURLParams.ChartType, chartType ?? "");
  }

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not
  return searchParams;
}

function toPivotUrlParams(
  exploreState: MetricsExplorerEntity,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();
  if (!exploreState.pivot?.active) return searchParams;

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  const rows = exploreState.pivot.rows.map(mapPivotEntry);
  if (!arrayOrderedEquals(rows, preset.pivotRows ?? [])) {
    searchParams.set(ExploreStateURLParams.PivotRows, rows.join(","));
  }

  const cols = exploreState.pivot.columns.map(mapPivotEntry);
  if (!arrayOrderedEquals(cols, preset.pivotCols ?? [])) {
    searchParams.set(ExploreStateURLParams.PivotColumns, cols.join(","));
  }

  const sort = exploreState.pivot.sorting?.[0];
  const sortId =
    sort?.id in ToURLParamTimeDimensionMap
      ? ToURLParamTimeDimensionMap[sort?.id]
      : sort?.id;

  if (shouldSetParam(preset.pivotSortBy, sortId)) {
    searchParams.set(ExploreStateURLParams.SortBy, sortId ?? "");
  }
  if (sort && !!preset.pivotSortAsc !== !sort?.desc) {
    searchParams.set(
      ExploreStateURLParams.SortDirection,
      sort?.desc ? "DESC" : "ASC",
    );
  }

  const tableMode = exploreState.pivot?.tableMode;
  if (shouldSetParam(preset.pivotTableMode, tableMode)) {
    searchParams.set(ExploreStateURLParams.PivotTableMode, tableMode ?? "nest");
  }

  // TODO: other fields like expanded state and pin are not supported right now
  return searchParams;
}

function shouldSetParam<T>(
  presetValue: T | undefined,
  exploreStateValue: T | undefined,
) {
  // there is no value in preset, set param only if state has a value
  if (presetValue === undefined) {
    return !!exploreStateValue;
  }

  // both preset and state value is non-truthy.
  // EG: one is "" and another is undefined then we should not set param as empty string
  if ((!presetValue && !exploreStateValue) || exploreStateValue === undefined) {
    return false;
  }

  return presetValue !== exploreStateValue;
}
