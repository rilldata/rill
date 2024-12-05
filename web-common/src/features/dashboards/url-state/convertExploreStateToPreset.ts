import { mergeDimensionAndMeasureFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { toTimeRangeParam } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { FromLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
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

export function convertExploreStateToPreset(
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

  Object.assign(preset, getTimeRangeFields(exploreState));

  Object.assign(preset, getExploreFields(exploreState, explore));

  Object.assign(preset, getTimeDimensionFields(exploreState));

  Object.assign(preset, getPivotFields(exploreState));

  return preset;
}

function getTimeRangeFields(exploreState: Partial<MetricsExplorerEntity>) {
  const preset: V1ExplorePreset = {};

  if (exploreState.selectedTimeRange?.name) {
    preset.timeRange = toTimeRangeParam(exploreState.selectedTimeRange);
  }
  if (exploreState.selectedTimeRange?.interval) {
    preset.timeGrain =
      ToURLParamTimeGrainMapMap[exploreState.selectedTimeRange.interval];
  }

  if (
    exploreState.showTimeComparison &&
    exploreState.selectedComparisonTimeRange?.name
  ) {
    preset.compareTimeRange = toTimeRangeParam(
      exploreState.selectedComparisonTimeRange,
    );
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

  if (exploreState.selectedScrubRange) {
    preset.selectTimeRange = toTimeRangeParam(exploreState.selectedScrubRange);
  }

  return preset;
}

function getExploreFields(
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
    preset.exploreSortBy = exploreState.leaderboardMeasureName;
  }

  if (exploreState.sortDirection) {
    preset.exploreSortAsc =
      exploreState.sortDirection ===
      DashboardState_LeaderboardSortDirection.ASCENDING;
  }

  if (exploreState.leaderboardContextColumn !== undefined) {
    // TODO: is this still used?
  }

  if (exploreState.dashboardSortType) {
    preset.exploreSortType =
      FromLegacySortTypeMap[exploreState.dashboardSortType];
  }

  if (exploreState.selectedDimensionName !== undefined) {
    preset.exploreExpandedDimension = exploreState.selectedDimensionName;
  }

  return preset;
}

function getTimeDimensionFields(exploreState: Partial<MetricsExplorerEntity>) {
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

function getPivotFields(exploreState: Partial<MetricsExplorerEntity>) {
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
  const sort = exploreState.pivot.sorting?.[0];
  if (sort) {
    preset.pivotSortBy = sort.id;
    preset.pivotSortAsc = !sort.desc;
  }

  // TODO: other fields like expanded state and pin are not supported right now
  return preset;
}
