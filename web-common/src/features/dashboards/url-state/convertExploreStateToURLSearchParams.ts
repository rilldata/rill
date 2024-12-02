import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  ExploreStateDefaultSortDirection,
  ExploreStateDefaultTDDChartType,
  ExploreStateDefaultTimeRange,
  ExploreStateDefaultTimezone,
} from "@rilldata/web-common/features/dashboards/url-state/defaults";
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
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export function convertExploreStateToURLSearchParams(
  exploreState: MetricsExplorerEntity,
  exploreSpec: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const searchParams = new URLSearchParams();

  if (!exploreState) return searchParams;

  const currentView = FromActivePageMap[exploreState.activePage];
  if (
    shouldSetParamWithDefault(
      preset.view,
      currentView,
      V1ExploreWebView.EXPLORE_WEB_VIEW_OVERVIEW,
    )
  ) {
    searchParams.set("view", ToURLParamViewMap[currentView] as string);
  }

  mergeSearchParams(
    toTimeRangesUrl(exploreState, exploreSpec, preset),
    searchParams,
  );

  const expr = mergeMeasureFilters(exploreState);
  if (expr && expr?.cond?.exprs?.length) {
    searchParams.set("f", convertExpressionToFilterParam(expr));
  }

  switch (exploreState.activePage) {
    case DashboardState_ActivePage.UNSPECIFIED:
    case DashboardState_ActivePage.DEFAULT:
    case DashboardState_ActivePage.DIMENSION_TABLE:
      mergeSearchParams(
        toOverviewUrl(exploreState, exploreSpec, preset),
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

  if (
    (preset.timeRange !== undefined &&
      exploreState.selectedTimeRange !== undefined &&
      exploreState.selectedTimeRange.name !== preset.timeRange) ||
    (preset.timeRange === undefined &&
      exploreState.selectedTimeRange?.name !== ExploreStateDefaultTimeRange)
  ) {
    searchParams.set("tr", toTimeRangeParam(exploreState.selectedTimeRange));
  }

  if (
    shouldSetParamWithDefault(
      preset.timezone,
      exploreState.selectedTimezone,
      ExploreStateDefaultTimezone,
    )
  ) {
    searchParams.set("tz", exploreState.selectedTimezone);
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
        "compare_tr",
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
        searchParams.set("compare_tr", inferredCompareTimeRange);
    }
  }

  const mappedTimeGrain =
    ToURLParamTimeGrainMapMap[exploreState.selectedTimeRange?.interval ?? ""] ??
    "";
  if (shouldSetParam(preset.timeGrain, mappedTimeGrain)) {
    searchParams.set("grain", mappedTimeGrain);
  }

  if (
    shouldSetParam(
      preset.comparisonDimension,
      exploreState.selectedComparisonDimension,
    )
  ) {
    // TODO: move this based on expected param sequence
    searchParams.set(
      "compare_dim",
      exploreState.selectedComparisonDimension ?? "",
    );
  }

  if (
    exploreState.selectedScrubRange &&
    !exploreState.selectedScrubRange.isScrubbing
  ) {
    searchParams.set(
      "highlighted_tr",
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

function toOverviewUrl(
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
    searchParams.set("measures", visibleMeasuresParam);
  }

  const visibleDimensionsParam = toVisibleDimensionsUrlParam(
    exploreState,
    exploreSpec,
    preset,
  );
  if (visibleDimensionsParam) {
    searchParams.set("dims", visibleDimensionsParam);
  }

  if (
    shouldSetParam(
      preset.overviewExpandedDimension,
      exploreState.selectedDimensionName,
    )
  ) {
    searchParams.set("expand_dim", exploreState.selectedDimensionName ?? "");
  }

  const defaultLeaderboardMeasure =
    preset.measures?.[0] ?? exploreSpec.measures?.[0];
  if (
    shouldSetParamWithDefault(
      preset.overviewSortBy,
      exploreState.leaderboardMeasureName,
      defaultLeaderboardMeasure,
    )
  ) {
    searchParams.set("sort_by", exploreState.leaderboardMeasureName);
  }

  const sortType = FromLegacySortTypeMap[exploreState.dashboardSortType];
  if (shouldSetParam(preset.overviewSortType, sortType)) {
    searchParams.set("sort_type", ToURLParamSortTypeMap[sortType] ?? "");
  }

  const sortAsc = exploreState.sortDirection === SortDirection.ASCENDING;
  if (
    // if preset has a sort direction then only set if not the same
    (preset.overviewSortAsc !== undefined &&
      preset.overviewSortAsc !== sortAsc) ||
    // else if the direction is not the default then set the param
    (preset.overviewSortAsc === undefined &&
      exploreState.sortDirection !== ExploreStateDefaultSortDirection)
  ) {
    searchParams.set("sort_dir", sortAsc ? "ASC" : "DESC");
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
    searchParams.set("measure", exploreState.tdd.expandedMeasureName ?? "");
  }

  const chartType = ToURLParamTDDChartMap[exploreState.tdd.chartType];
  if (
    shouldSetParamWithDefault(
      preset.timeDimensionChartType,
      chartType,
      ExploreStateDefaultTDDChartType,
    )
  ) {
    searchParams.set("chart_type", chartType ?? "");
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
    searchParams.set("rows", rows.join(","));
  }

  const cols = [
    ...exploreState.pivot.columns.dimension.map(mapPivotEntry),
    ...exploreState.pivot.columns.measure.map(mapPivotEntry),
  ];
  if (!arrayOrderedEquals(cols, preset.pivotCols ?? [])) {
    searchParams.set("cols", cols.join(","));
  }

  const sort = exploreState.pivot.sorting?.[0];
  const sortId =
    sort?.id in ToURLParamTimeDimensionMap
      ? ToURLParamTimeDimensionMap[sort?.id]
      : sort?.id;
  if (shouldSetParam(preset.pivotSortBy, sortId)) {
    searchParams.set("sort_by", sortId ?? "");
  }
  if (sort && !!preset.pivotSortAsc !== !sort?.desc) {
    searchParams.set("sort_dir", sort?.desc ? "DESC" : "ASC");
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

function shouldSetParamWithDefault<T>(
  presetValue: T | undefined,
  exploreStateValue: T | undefined,
  defaultValue: T,
) {
  // there is no value in preset, set param only if state has a value
  if (presetValue === undefined) {
    return exploreStateValue != defaultValue;
  }

  // both preset and state value is non-truthy.
  // EG: one is "" and another is undefined then we should not set param as empty string
  if (!presetValue && !exploreStateValue) {
    return false;
  }

  return presetValue !== exploreStateValue;
}
