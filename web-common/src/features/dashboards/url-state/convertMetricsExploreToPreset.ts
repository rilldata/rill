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
  ToURLParamTimeGrainMapMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { DashboardState_LeaderboardSortDirection } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";

export function convertMetricsExploreToPreset(
  exploreState: Partial<MetricsExplorerEntity>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};

  if (exploreState.activePage) {
    preset.view = FromActivePageMap[exploreState.activePage];
  }

  if (exploreState.whereFilter || exploreState.dimensionThresholdFilters) {
    preset.where = mergeDimensionAndMeasureFilter(
      exploreState.whereFilter ?? createAndExpression([]),
      exploreState.dimensionThresholdFilters ?? [],
    );
  }

  Object.assign(preset, fromMetricsExploreTimeRangeFields(exploreState));

  Object.assign(
    preset,
    fromMetricsExploreOverviewFields(exploreState, explore),
  );

  Object.assign(preset, fromMetricsExploreTimeDimensionFields(exploreState));

  Object.assign(preset, fromMetricsExplorePivotFields(exploreState));

  return preset;
}

function fromMetricsExploreTimeRangeFields(
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const preset: V1ExplorePreset = {};

  if (exploreState.selectedTimeRange?.name) {
    preset.timeRange = exploreState.selectedTimeRange?.name;
    // TODO: custom time range
  }
  if (exploreState.selectedTimeRange?.interval) {
    preset.timeGrain =
      ToURLParamTimeGrainMapMap[exploreState.selectedTimeRange.interval];
  }

  if (exploreState.selectedComparisonTimeRange?.name) {
    preset.compareTimeRange = exploreState.selectedComparisonTimeRange.name;
    // TODO: custom time range
  }
  if (exploreState.showTimeComparison) {
    preset.comparisonMode =
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME;
  }

  if (exploreState.selectedComparisonDimension !== undefined) {
    preset.comparisonDimension = exploreState.selectedComparisonDimension;
    if (exploreState.selectedComparisonDimension) {
      preset.comparisonMode =
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION;
    }
  }

  if (exploreState.selectedTimezone) {
    preset.timezone = exploreState.selectedTimezone;
  }

  // TODO: scrubRange

  return preset;
}

function fromMetricsExploreOverviewFields(
  exploreState: Partial<MetricsExplorerEntity>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};

  if (exploreState.allMeasuresVisible) {
    preset.measures = explore.measures ?? [];
  } else if (exploreState.visibleMeasureKeys) {
    preset.measures = [...exploreState.visibleMeasureKeys];
  }

  if (exploreState.allDimensionsVisible) {
    preset.dimensions = explore.dimensions ?? [];
  } else if (exploreState.visibleDimensionKeys) {
    preset.dimensions = [...exploreState.visibleDimensionKeys];
  }

  if (exploreState.leaderboardMeasureName !== undefined) {
    preset.overviewSortBy = exploreState.leaderboardMeasureName;
  }

  if (exploreState.sortDirection) {
    preset.overviewSortAsc =
      exploreState.sortDirection ===
      DashboardState_LeaderboardSortDirection.ASCENDING;
  }

  if (exploreState.leaderboardContextColumn !== undefined) {
    // TODO
  }

  if (exploreState.dashboardSortType) {
    // TODO
  }

  if (exploreState.selectedDimensionName !== undefined) {
    preset.overviewExpandedDimension = exploreState.selectedDimensionName;
  }

  return preset;
}

function fromMetricsExploreTimeDimensionFields(
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const preset: V1ExplorePreset = {};

  if (!exploreState.tdd) {
    return preset;
  }

  preset.timeDimensionMeasure = exploreState.tdd.expandedMeasureName;
  preset.timeDimensionPin = false; // TODO
  preset.timeDimensionChartType =
    ToURLParamTDDChartMap[exploreState.tdd.chartType];

  return preset;
}

function fromMetricsExplorePivotFields(
  exploreState: Partial<MetricsExplorerEntity>,
) {
  const preset: V1ExplorePreset = {};

  if (!exploreState.pivot) {
    return preset;
  }

  const mapPivotEntry = (data: PivotChipData) => {
    if (data.type === PivotChipType.Time)
      return ToURLParamTimeDimensionMap[data.id] as string;
    return data.id;
  };

  preset.pivotRows = exploreState.pivot.rows.dimension.map(mapPivotEntry);
  preset.pivotCols = [
    ...exploreState.pivot.columns.dimension.map(mapPivotEntry),
    ...exploreState.pivot.columns.measure.map(mapPivotEntry),
  ];

  // TODO: other fields

  return preset;
}
