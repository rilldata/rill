import { mergeDimensionAndMeasureFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  FromActivePageMap,
  ToURLParamTDDChartMap,
  ToURLParamTimeDimensionMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { DashboardState_LeaderboardSortDirection } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

export function convertMetricsExploreToPreset(
  metrics: Partial<MetricsExplorerEntity>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};

  if (metrics.activePage) {
    preset.view = FromActivePageMap[metrics.activePage];
  }

  if (metrics.whereFilter || metrics.dimensionThresholdFilters) {
    preset.where = mergeDimensionAndMeasureFilter(
      metrics.whereFilter ?? createAndExpression([]),
      metrics.dimensionThresholdFilters ?? [],
    );
  }

  Object.assign(preset, fromMetricsExploreTimeRangeFields(metrics));

  Object.assign(preset, fromMetricsExploreOverviewFields(metrics, explore));

  Object.assign(preset, fromMetricsExploreTimeDimensionFields(metrics));

  Object.assign(preset, fromMetricsExplorePivotFields(metrics));

  return preset;
}

function fromMetricsExploreTimeRangeFields(
  metrics: Partial<MetricsExplorerEntity>,
) {
  const preset: V1ExplorePreset = {};

  if (metrics.selectedTimeRange?.name) {
    preset.timeRange = metrics.selectedTimeRange?.name;
    // TODO: custom time range
  }
  // TODO: grain

  if (metrics.selectedComparisonTimeRange?.name) {
    preset.compareTimeRange = metrics.selectedComparisonTimeRange.name;
    // TODO: custom time range
  }
  if (metrics.showTimeComparison) {
    preset.comparisonMode =
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME;
  }

  if (metrics.selectedComparisonDimension !== undefined) {
    preset.comparisonDimension = metrics.selectedComparisonDimension;
    if (metrics.selectedComparisonDimension) {
      preset.comparisonMode =
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION;
    }
  }

  if (metrics.selectedTimezone) {
    preset.timezone = metrics.selectedTimezone;
  }

  // TODO: scrubRange

  return preset;
}

function fromMetricsExploreOverviewFields(
  metrics: Partial<MetricsExplorerEntity>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};

  if (metrics.allMeasuresVisible) {
    preset.measures = explore.measures ?? [];
  } else if (metrics.visibleMeasureKeys) {
    preset.measures = [...metrics.visibleMeasureKeys];
  }

  if (metrics.allDimensionsVisible) {
    preset.dimensions = explore.dimensions ?? [];
  } else if (metrics.visibleDimensionKeys) {
    preset.dimensions = [...metrics.visibleDimensionKeys];
  }

  if (metrics.leaderboardMeasureName !== undefined) {
    preset.overviewSortBy = metrics.leaderboardMeasureName;
  }

  if (metrics.sortDirection) {
    preset.overviewSortAsc =
      metrics.sortDirection ===
      DashboardState_LeaderboardSortDirection.ASCENDING;
  }

  if (metrics.leaderboardContextColumn !== undefined) {
    // TODO
  }

  if (metrics.dashboardSortType) {
    // TODO
  }

  if (metrics.selectedDimensionName !== undefined) {
    preset.overviewExpandedDimension = metrics.selectedDimensionName;
  }

  return preset;
}

function fromMetricsExploreTimeDimensionFields(
  metrics: Partial<MetricsExplorerEntity>,
) {
  const preset: V1ExplorePreset = {};

  if (!metrics.tdd) {
    return preset;
  }

  preset.timeDimensionMeasure = metrics.tdd.expandedMeasureName;
  preset.timeDimensionPin = false; // TODO
  preset.timeDimensionChartType = ToURLParamTDDChartMap[metrics.tdd.chartType];

  return preset;
}

function fromMetricsExplorePivotFields(
  metrics: Partial<MetricsExplorerEntity>,
) {
  const preset: V1ExplorePreset = {};

  if (!metrics.pivot) {
    return preset;
  }

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  preset.pivotRows = metrics.pivot.rows.dimension.map(mapPivotEntry);
  preset.pivotCols = [
    ...metrics.pivot.columns.dimension.map(mapPivotEntry),
    ...metrics.pivot.columns.measure.map(mapPivotEntry),
  ];

  // TODO: other fields

  return preset;
}
