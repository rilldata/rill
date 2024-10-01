import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { ToURLParamTimeDimensionMap } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  arrayOrderedEquals,
  arrayUnorderedEquals,
} from "@rilldata/web-common/lib/arrayUtils";
import type {
  V1ExplorePreset,
  V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

export function getUrlFromMetricsExplorer(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  if (!metrics) return;

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
  if (metrics.selectedTimezone !== preset.timezone) {
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

  if (
    // if sort by is defined then only set param if selected is not the same.
    (preset.overviewSortBy &&
      metrics.leaderboardMeasureName !== preset.overviewSortBy) ||
    // else the default is the 1st measure in explore, so check that next
    metrics.leaderboardMeasureName !== explore.measures?.[0]
  ) {
    searchParams.set("o.sb", metrics.leaderboardMeasureName);
  }

  const sortAsc = metrics.sortDirection === SortDirection.ASCENDING;
  if (
    preset.overviewSortAsc === undefined ||
    preset.overviewSortAsc !== sortAsc
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
    metrics.tdd.chartType !== TDDChart.DEFAULT
  ) {
    searchParams.set("tdd.p", metrics.tdd.chartType);
  }

  // TODO: pin
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
  if (arrayOrderedEquals(rows, preset.pivotRows ?? [])) {
    searchParams.set("p.r", rows.join(","));
  }

  const cols = [
    ...metrics.pivot.columns.dimension.map(mapPivotEntry),
    ...metrics.pivot.columns.measure.map(mapPivotEntry),
  ];
  if (arrayOrderedEquals(cols, preset.pivotCols ?? [])) {
    searchParams.set("p.c", cols.join(","));
  }

  // TODO: other fields
}
