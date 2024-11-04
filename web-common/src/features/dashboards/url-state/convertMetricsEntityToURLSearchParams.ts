import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  URLStateDefaultSortDirection,
  URLStateDefaultTDDChartType,
  URLStateDefaultTimeRange,
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
  type V1ExplorePreset,
  type V1ExploreSpec,
  V1ExploreWebView,
} from "@rilldata/web-common/runtime-client";

export function convertMetricsEntityToURLSearchParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!metrics) return;

  const currentView = FromActivePageMap[metrics.activePage];
  if (
    (preset.view !== undefined && preset.view !== currentView) ||
    (preset.view === undefined &&
      currentView !== V1ExploreWebView.EXPLORE_ACTIVE_PAGE_OVERVIEW)
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
    (preset.timeRange !== undefined &&
      metrics.selectedTimeRange?.name !== preset.timeRange) ||
    (preset.timeRange === undefined &&
      metrics.selectedTimeRange?.name !== URLStateDefaultTimeRange)
  ) {
    searchParams.set("tr", metrics.selectedTimeRange?.name ?? "");
  }
  if (
    // if preset has timezone then only set if selected is not the same
    (preset.timezone !== undefined &&
      metrics.selectedTimezone !== preset.timezone) ||
    // else if the timezone is not the default then set the param
    (preset.timezone === undefined &&
      metrics.selectedTimezone !== URLStateDefaultTimezone)
  ) {
    searchParams.set("tz", metrics.selectedTimezone);
  }

  if (
    (preset.compareTimeRange !== undefined &&
      metrics.selectedComparisonTimeRange?.name !== preset.compareTimeRange) ||
    (preset.compareTimeRange === undefined &&
      !!metrics.selectedComparisonTimeRange?.name)
  ) {
    searchParams.set("ctr", metrics.selectedComparisonTimeRange?.name ?? "");
  }
  if (
    // if preset has a compare dimension then only set if selected is not the same
    (preset.comparisonDimension !== undefined &&
      metrics.selectedComparisonDimension !== preset.comparisonDimension) ||
    // else if there is no default then set if there was a selected compare dimension
    (preset.comparisonDimension === undefined &&
      !!metrics.selectedComparisonDimension)
  ) {
    searchParams.set("cd", metrics.selectedComparisonDimension ?? "");
  }

  // TODO: grain
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
    // if sort by is defined in preset then only set param if selected is not the same.
    (preset.overviewSortBy &&
      metrics.leaderboardMeasureName !== preset.overviewSortBy) ||
    // else the default is the 1st measure in preset or explore, so check that next
    (!preset.overviewSortBy &&
      metrics.leaderboardMeasureName !== defaultLeaderboardMeasure)
  ) {
    searchParams.set("o.sb", metrics.leaderboardMeasureName);
  }

  const sortAsc = metrics.sortDirection === SortDirection.ASCENDING;
  if (
    // if preset has a sort direction then only set if not the same
    (preset.overviewSortAsc !== undefined &&
      preset.overviewSortAsc !== sortAsc) ||
    // else if the direction is not the default then set the param
    (preset.overviewSortAsc === undefined &&
      metrics.sortDirection !== URLStateDefaultSortDirection)
  ) {
    searchParams.set("o.sd", sortAsc ? "ASC" : "DESC");
  }

  if (
    (preset.overviewExpandedDimension !== undefined &&
      metrics.selectedDimensionName !== preset.overviewExpandedDimension) ||
    (preset.overviewExpandedDimension === undefined &&
      metrics.selectedDimensionName)
  ) {
    searchParams.set("o.ed", metrics.selectedDimensionName ?? "");
  }
}

function toTimeDimensionUrlParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  preset: V1ExplorePreset,
) {
  if (
    (preset.timeDimensionMeasure !== undefined &&
      metrics.tdd.expandedMeasureName !== preset.timeDimensionMeasure) ||
    (preset.timeDimensionMeasure === undefined &&
      metrics.tdd.expandedMeasureName)
  ) {
    searchParams.set("tdd.m", metrics.tdd.expandedMeasureName ?? "");
  }

  if (
    (preset.timeDimensionChartType !== undefined &&
      ToURLParamTDDChartMap[metrics.tdd.chartType] !==
        preset.timeDimensionChartType) ||
    (preset.timeDimensionChartType === undefined &&
      metrics.tdd.chartType !== URLStateDefaultTDDChartType)
  ) {
    searchParams.set(
      "tdd.ct",
      ToURLParamTDDChartMap[metrics.tdd.chartType] ?? "",
    );
  }

  // TODO: pin
  // TODO: what should be done when chartType is set but expandedMeasureName is not
}

function toPivotUrlParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  preset: V1ExplorePreset,
) {
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
