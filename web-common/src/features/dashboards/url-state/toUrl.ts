import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  URLStateDefaultSortDirection,
  URLStateDefaultTDDChartType,
  URLStateDefaultTimezone,
} from "@rilldata/web-common/features/dashboards/url-state/defaults";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import {
  FromActivePageMap,
  ToURLParamTDDChartMap,
  ToURLParamTimeDimensionMap,
  ToURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  arrayOrderedEquals,
  arrayUnorderedEquals,
} from "@rilldata/web-common/lib/arrayUtils";
import {
  V1ExplorePreset,
  V1ExploreSpec,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export function getUrlFromMetricsExplorer(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!metrics) return;

  const currentView = FromActivePageMap[metrics.activePage];
  if (
    (preset.view !== undefined && preset.view !== currentView) ||
    currentView !== V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW
  ) {
    searchParams.set("vw", ToURLParamViewMap[currentView] as string);
  }

  const expr = mergeMeasureFilters(metrics);
  if (expr && expr?.cond?.exprs?.length) {
    searchParams.set("f", convertExpressionToFilterParam(expr));
  }

  toTimeRangesUrl(metrics, searchParams, preset);

  toOverviewUrl(metrics, searchParams, explore, preset);

  toTimeDimensionUrlParams(metrics, searchParams, preset);

  toPivotUrlParams(metrics, searchParams, preset);
}

function toTimeRangesUrl(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  preset: V1ExplorePreset,
) {
  if (
    metrics.selectedTimeRange?.name &&
    metrics.selectedTimeRange?.name !== preset.timeRange
  ) {
    searchParams.set("tr", metrics.selectedTimeRange.name);
  }
  if (
    (preset.timezone !== undefined &&
      metrics.selectedTimezone !== preset.timezone) ||
    metrics.selectedTimezone !== URLStateDefaultTimezone
  ) {
    searchParams.set("tz", metrics.selectedTimezone);
  }

  if (metrics.selectedComparisonTimeRange?.name !== preset.compareTimeRange) {
    searchParams.set("ctr", metrics.selectedComparisonTimeRange?.name ?? "");
  }
  if (metrics.selectedComparisonDimension !== preset.comparisonDimension) {
    searchParams.set("cd", metrics.selectedComparisonDimension ?? "");
  }
}

function toOverviewUrl(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const measures = [...metrics.visibleMeasureKeys];
  const presetMeasures = preset.measures ?? explore.measures ?? [];
  if (!arrayUnorderedEquals(measures, presetMeasures)) {
    if (metrics.allMeasuresVisible) {
      searchParams.set("o.m", "*");
    } else {
      searchParams.set("o.m", measures.join(","));
    }
  }

  const dimensions = [...metrics.visibleDimensionKeys];
  const presetDimensions = preset.dimensions ?? explore.dimensions ?? [];
  if (!arrayUnorderedEquals(dimensions, presetDimensions)) {
    if (metrics.allDimensionsVisible) {
      searchParams.set("o.d", "*");
    } else {
      searchParams.set("o.d", dimensions.join(","));
    }
  }

  const defaultLeaderboardMeasure =
    preset.measures?.[0] ?? explore.measures?.[0];
  if (
    // if sort by is defined then only set param if selected is not the same.
    (preset.overviewSortBy &&
      metrics.leaderboardMeasureName !== preset.overviewSortBy) ||
    // else the default is the 1st measure in preset or explore, so check that next
    metrics.leaderboardMeasureName !== defaultLeaderboardMeasure
  ) {
    searchParams.set("o.sb", metrics.leaderboardMeasureName);
  }

  const sortAsc = metrics.sortDirection === SortDirection.ASCENDING;
  if (
    // if preset has a sort direction then only set if not the same
    (preset.overviewSortAsc !== undefined &&
      preset.overviewSortAsc !== sortAsc) ||
    // else if the direction is not the default then set the param
    metrics.sortDirection !== URLStateDefaultSortDirection
  ) {
    searchParams.set("o.sd", sortAsc ? "ASC" : "DESC");
  }

  if (
    metrics.selectedDimensionName &&
    metrics.selectedDimensionName !== preset.overviewExpandedDimension
  ) {
    searchParams.set("o.ed", metrics.selectedDimensionName);
  }
}

function toTimeDimensionUrlParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  preset: V1ExplorePreset,
) {
  if (
    metrics.tdd.expandedMeasureName &&
    metrics.tdd.expandedMeasureName !== preset.timeDimensionMeasure
  ) {
    searchParams.set("tdd.m", metrics.tdd.expandedMeasureName);
  }

  if (
    (preset.timeDimensionChartType !== undefined &&
      metrics.tdd.chartType !== preset.timeDimensionChartType) ||
    metrics.tdd.chartType !== URLStateDefaultTDDChartType
  ) {
    searchParams.set(
      "tdd.ct",
      ToURLParamTDDChartMap[metrics.tdd.chartType] ?? "",
    );
  }

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not st
}

function toPivotUrlParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  preset: V1ExplorePreset,
) {
  if (!metrics.pivot.active) return;

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  const rows = metrics.pivot.rows.dimension.map(mapPivotEntry);
  if (!arrayOrderedEquals(rows, preset.pivotRows ?? [])) {
    searchParams.set("p.r", rows.join(","));
  }

  const cols = [
    ...metrics.pivot.columns.dimension.map(mapPivotEntry),
    ...metrics.pivot.columns.measure.map(mapPivotEntry),
  ];
  if (!arrayOrderedEquals(cols, preset.pivotCols ?? [])) {
    searchParams.set("p.c", cols.join(","));
  }

  // TODO: other fields
}
