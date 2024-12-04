import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
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
import {
  arrayOrderedEquals,
  arrayUnorderedEquals,
} from "@rilldata/web-common/lib/arrayUtils";
import { inferCompareTimeRange } from "@rilldata/web-common/lib/time/comparisons";
import {
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { mergeSearchParams } from "@rilldata/web-common/lib/url-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

export function convertExploreStateToURLSearchParams(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (!exploreState) return searchParams;

  const currentView = FromActivePageMap[exploreState.activePage];
  if (shouldSetParam(preset.view, currentView)) {
    searchParams.set(
      ExploreStateURLParams.WebView,
      ToURLParamViewMap[currentView] as string,
    );
  }

  mergeSearchParams(
    toTimeRangesUrl(exploreState, exploreSpec, preset),
    searchParams,
  );

  const expr = mergeMeasureFilters(exploreState);
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
      mergeSearchParams(
        toExploreUrl(exploreState, exploreSpec, preset),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL:
      mergeSearchParams(
        toTimeDimensionUrlParams(exploreState, preset),
        searchParams,
      );
      break;

    case DashboardState_ActivePage.PIVOT:
      mergeSearchParams(toPivotUrlParams(exploreState, preset), searchParams);
      break;
  }

  return searchParams;
}

function toTimeRangesUrl(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (shouldSetParam(preset.timeRange, exploreState.selectedTimeRange?.name)) {
    searchParams.set(
      ExploreStateURLParams.TimeRange,
      toTimeRangeParam(exploreState.selectedTimeRange),
    );
  }

  if (shouldSetParam(preset.timezone, exploreState.selectedTimezone)) {
    searchParams.set(
      ExploreStateURLParams.TimeZone,
      exploreState.selectedTimezone,
    );
  }

  if (exploreState.showTimeComparison) {
    if (
      (preset.compareTimeRange !== undefined &&
        exploreState.selectedComparisonTimeRange !== undefined &&
        exploreState.selectedComparisonTimeRange.name !==
          preset.compareTimeRange) ||
      preset.compareTimeRange === undefined
    ) {
      searchParams.set(
        ExploreStateURLParams.ComparisonTimeRange,
        toTimeRangeParam(exploreState.selectedComparisonTimeRange),
      );
    } else if (
      !exploreState.selectedComparisonTimeRange?.name &&
      exploreState.selectedTimeRange?.name
    ) {
      // we infer compare time range if the user has not explicitly selected one but has enabled comparison
      const inferredCompareTimeRange = inferCompareTimeRange(
        exploreSpec.timeRanges,
        exploreState.selectedTimeRange.name,
      );
      if (inferredCompareTimeRange)
        searchParams.set(
          ExploreStateURLParams.ComparisonTimeRange,
          inferredCompareTimeRange,
        );
    }
  }

  const mappedTimeGrain =
    ToURLParamTimeGrainMapMap[exploreState.selectedTimeRange?.interval ?? ""] ??
    "";
  if (shouldSetParam(preset.timeGrain, mappedTimeGrain)) {
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

function toTimeRangeParam(timeRange: TimeRange | undefined) {
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
  if (!presetValue && !exploreStateValue) {
    return false;
  }

  return presetValue !== exploreStateValue;
}
