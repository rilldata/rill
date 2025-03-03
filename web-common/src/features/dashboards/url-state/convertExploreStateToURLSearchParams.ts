import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { FromLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  FromActivePageMap,
  FromURLParamViewMap,
  ToURLParamSortTypeMap,
  ToURLParamTDDChartMap,
  ToURLParamTimeDimensionMap,
  ToURLParamTimeGrainMapMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  arrayOrderedEquals,
  arrayUnorderedEquals,
} from "@rilldata/web-common/lib/arrayUtils";
import {
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  copyParamsToTarget,
  mergeParamsWithOverwrite,
} from "@rilldata/web-common/lib/url-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Sometimes data is loaded from sources other than the url.
 * In that case update the URL to make sure the state matches the current url.
 */
export function getUpdatedUrlForExploreState(
  exploreSpec: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
  defaultExplorePreset: V1ExplorePreset,
  partialExploreState: Partial<MetricsExplorerEntity>,
  curSearchParams: URLSearchParams,
): string {
  // Create params from the explore state
  const stateParams = convertExploreStateToURLSearchParams(
    partialExploreState as MetricsExplorerEntity,
    exploreSpec,
    timeControlsState,
    defaultExplorePreset,
  );

  // Filter out the default view parameter if needed
  const filteredCurrentParams = new URLSearchParams();
  curSearchParams.forEach((value, key) => {
    if (
      key === ExploreStateURLParams.WebView &&
      FromURLParamViewMap[value] === defaultExplorePreset.view
    ) {
      return; // Skip this parameter
    }
    filteredCurrentParams.set(key, value);
  });

  // Merge with current params overwriting the state params
  const mergedParams = mergeParamsWithOverwrite(
    filteredCurrentParams,
    stateParams,
  );
  return mergedParams.toString();
}

export function convertExploreStateToURLSearchParams(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  // We have quite a bit of logic in TimeControlState to validate selections and update them
  // Eg: if a selected grain is not applicable for the current grain then we change it
  // But it is only available in TimeControlState and not MetricsExplorerEntity
  timeControlsState: TimeControlState | undefined,
  preset: V1ExplorePreset,
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
      toTimeRangesUrl(exploreState, exploreSpec, timeControlsState, preset),
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
      convertExpressionToFilterParam(expr),
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
      break;
  }

  return searchParams;
}

function toTimeRangesUrl(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
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

  if (
    exploreSpec.timeZones?.length &&
    shouldSetParam(preset.timezone, exploreState.selectedTimezone)
  ) {
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
    shouldSetParam(preset.exploreSortBy, exploreState.leaderboardMeasureName)
  ) {
    searchParams.set(
      ExploreStateURLParams.SortBy,
      exploreState.leaderboardMeasureName,
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

  return searchParams;
}

function toVisibleMeasuresUrlParam(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!exploreState.visibleMeasureKeys) return undefined;

  const measures = [...exploreState.visibleMeasureKeys];
  const presetMeasures = preset.measures ?? exploreSpec.measures ?? [];
  if (arrayUnorderedEquals(measures, presetMeasures)) {
    return undefined;
  }
  if (exploreState.allMeasuresVisible) {
    return "*";
  }
  return measures.join(",");
}

function toVisibleDimensionsUrlParam(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!exploreState.visibleDimensionKeys) return undefined;

  const dimensions = [...exploreState.visibleDimensionKeys];
  const presetDimensions = preset.dimensions ?? exploreSpec.dimensions ?? [];
  if (arrayUnorderedEquals(dimensions, presetDimensions)) {
    return undefined;
  }
  if (exploreState.allDimensionsVisible) {
    return "*";
  }
  return dimensions.join(",");
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

  const rows = exploreState.pivot.rows.dimension.map(mapPivotEntry);
  if (!arrayOrderedEquals(rows, preset.pivotRows ?? [])) {
    searchParams.set(ExploreStateURLParams.PivotRows, rows.join(","));
  }

  const cols = [
    ...exploreState.pivot.columns.dimension.map(mapPivotEntry),
    ...exploreState.pivot.columns.measure.map(mapPivotEntry),
  ];
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
