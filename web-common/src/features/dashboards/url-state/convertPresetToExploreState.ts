import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  type PivotChipData,
  PivotChipType,
  type PivotTableMode,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import { ToLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  FromURLParamTDDChartMap,
  FromURLParamTimeDimensionMap,
  FromURLParamTimeGrainMap,
  ToActivePageViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortType,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  V1ExploreSortType,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { SortingState } from "@tanstack/svelte-table";

/**
 * Converts a V1ExplorePreset to our internal metrics explore state.
 */
export function convertPresetToExploreState(
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const partialExploreState: Partial<ExploreState> = {};
  const errors: Error[] = [];

  const measures = getMapFromArray(
    metricsView.measures?.filter((m) => explore.measures?.includes(m.name!)) ??
      [],
    (m) => m.name!,
  );
  const dimensions = getMapFromArray(
    metricsView.dimensions?.filter((d) =>
      explore.dimensions?.includes(d.name!),
    ) ?? [],
    (d) => d.name!,
  );

  if (preset.view) {
    partialExploreState.activePage = Number(
      ToActivePageViewMap[preset.view] ?? "0",
    );
  }

  if (preset.where) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      preset.where,
    );
    partialExploreState.whereFilter = dimensionFilters;
    partialExploreState.dimensionThresholdFilters = dimensionThresholdFilters;
  }
  if (preset.dimensionsWithInlistFilter) {
    partialExploreState.dimensionsWithInlistFilter =
      preset.dimensionsWithInlistFilter;
  }

  const { partialExploreState: trPartialState, errors: trErrors } =
    fromTimeRangesParams(preset, dimensions);
  Object.assign(partialExploreState, trPartialState);
  errors.push(...trErrors);

  const { partialExploreState: ovPartialState, errors: ovErrors } =
    fromExploreUrlParams(measures, dimensions, explore, preset);
  Object.assign(partialExploreState, ovPartialState);
  errors.push(...ovErrors);

  const { partialExploreState: tddPartialState, errors: tddErrors } =
    fromTimeDimensionUrlParams(measures, preset);
  Object.assign(partialExploreState, tddPartialState);
  errors.push(...tddErrors);

  const { partialExploreState: pivotPartialState, errors: pivotErrors } =
    fromPivotUrlParams(measures, dimensions, preset);
  Object.assign(partialExploreState, pivotPartialState);
  errors.push(...pivotErrors);

  return { partialExploreState, errors };
}

function fromTimeRangesParams(
  preset: V1ExplorePreset,
  dimensions: Map<string, MetricsViewSpecDimension>,
) {
  const partialExploreState: Partial<ExploreState> = {};
  const errors: Error[] = [];

  if (preset.timeRange) {
    partialExploreState.selectedTimeRange = fromTimeRangeUrlParam(
      preset.timeRange,
    );
  }

  if (preset.timeGrain) {
    partialExploreState.selectedTimeRange ??= {} as DashboardTimeControls;
    partialExploreState.selectedTimeRange.interval =
      FromURLParamTimeGrainMap[preset.timeGrain];
  }

  if (preset.timezone) {
    partialExploreState.selectedTimezone = preset.timezone;
  }

  let setCompareTimeRange = false;
  if (preset.compareTimeRange) {
    partialExploreState.selectedComparisonTimeRange = fromTimeRangeUrlParam(
      preset.compareTimeRange,
    );
    if (
      partialExploreState.selectedComparisonTimeRange.name ===
      TimeRangePreset.CUSTOM
    ) {
      partialExploreState.selectedComparisonTimeRange.name =
        TimeComparisonOption.CUSTOM;
    }
    partialExploreState.showTimeComparison = true;
    setCompareTimeRange = true;
    if (
      preset.comparisonMode ===
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
    ) {
      // unset compare dimension
      partialExploreState.selectedComparisonDimension = "";
    }
  } else if (
    preset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
  ) {
    partialExploreState.selectedComparisonTimeRange = undefined;
    partialExploreState.showTimeComparison = true;
  }

  if (preset.comparisonDimension) {
    if (dimensions.has(preset.comparisonDimension)) {
      partialExploreState.selectedComparisonDimension =
        preset.comparisonDimension;
      if (
        preset.comparisonMode ===
          V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION &&
        // since we are setting partial explore state we need to unset time compare settings
        !setCompareTimeRange
      ) {
        // unset compare time ranges
        partialExploreState.selectedComparisonTimeRange = undefined;
        partialExploreState.showTimeComparison = false;
      }
    } else {
      errors.push(
        getSingleFieldError("compare dimension", preset.comparisonDimension),
      );
    }
  } else if (
    preset.comparisonMode !==
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION
  ) {
    partialExploreState.selectedComparisonDimension = "";
  }

  if (preset.selectTimeRange) {
    partialExploreState.lastDefinedScrubRange =
      partialExploreState.selectedScrubRange = {
        ...fromTimeRangeUrlParam(preset.selectTimeRange),
        isScrubbing: false,
      };
  } else {
    partialExploreState.selectedScrubRange = undefined;
  }

  if (
    preset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE
  ) {
    // unset all comparison setting if mode is none
    partialExploreState.selectedComparisonTimeRange = undefined;
    partialExploreState.selectedComparisonDimension = "";
    partialExploreState.showTimeComparison = false;
  }

  return { partialExploreState, errors };
}

export const CustomTimeRangeRegex =
  /(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z),(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z)/;
export function fromTimeRangeUrlParam(tr: string) {
  const customTimeRangeMatch = CustomTimeRangeRegex.exec(tr);
  if (customTimeRangeMatch?.length) {
    const [, start, end] = customTimeRangeMatch;
    return {
      name: TimeRangePreset.CUSTOM,
      start: new Date(start),
      end: new Date(end),
    } as DashboardTimeControls;
  }

  return {
    name: tr,
  } as DashboardTimeControls;
}

function fromExploreUrlParams(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  explore: V1ExploreSpec,
  preset: V1ExplorePreset,
) {
  const partialExploreState: Partial<ExploreState> = {};
  const errors: Error[] = [];

  if (preset.measures?.length) {
    const selectedMeasures = preset.measures.filter((m) => measures.has(m));
    const missingMeasures = getMissingValues(selectedMeasures, preset.measures);
    if (missingMeasures.length) {
      errors.push(getMultiFieldError("measure", missingMeasures));
    }

    partialExploreState.allMeasuresVisible =
      selectedMeasures.length === explore.measures?.length;
    partialExploreState.visibleMeasures = [...selectedMeasures];
  }

  if (preset.dimensions?.length) {
    const selectedDimensions = preset.dimensions.filter((d) =>
      dimensions.has(d),
    );
    const missingDimensions = getMissingValues(
      selectedDimensions,
      preset.dimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }

    partialExploreState.allDimensionsVisible =
      selectedDimensions.length === explore.dimensions?.length;
    partialExploreState.visibleDimensions = [...selectedDimensions];
  }

  if (preset.exploreSortBy) {
    if (measures.has(preset.exploreSortBy)) {
      partialExploreState.leaderboardSortByMeasureName = preset.exploreSortBy;
    } else {
      errors.push(getSingleFieldError("sort by measure", preset.exploreSortBy));
    }
  }

  if (preset.exploreSortAsc !== undefined) {
    partialExploreState.sortDirection = preset.exploreSortAsc
      ? SortDirection.ASCENDING
      : SortDirection.DESCENDING;
  }

  if (
    preset.exploreSortType !== undefined &&
    preset.exploreSortType !== V1ExploreSortType.EXPLORE_SORT_TYPE_UNSPECIFIED
  ) {
    partialExploreState.dashboardSortType = Number(
      ToLegacySortTypeMap[preset.exploreSortType] ??
        DashboardState_LeaderboardSortType.UNSPECIFIED,
    );
  }

  if (preset.exploreLeaderboardMeasures !== undefined) {
    partialExploreState.leaderboardMeasureNames =
      preset.exploreLeaderboardMeasures;
  }

  if (preset.exploreExpandedDimension !== undefined) {
    if (preset.exploreExpandedDimension === "") {
      partialExploreState.selectedDimensionName = "";
      // if preset didnt have a view then this is a dimension table unset.
      if (
        preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED ||
        preset.view === undefined
      ) {
        partialExploreState.activePage = DashboardState_ActivePage.DEFAULT;
      }
    } else if (dimensions.has(preset.exploreExpandedDimension)) {
      partialExploreState.selectedDimensionName =
        preset.exploreExpandedDimension;
      if (
        preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_EXPLORE ||
        preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_UNSPECIFIED ||
        preset.view === undefined
      ) {
        partialExploreState.activePage =
          DashboardState_ActivePage.DIMENSION_TABLE;
      }
    } else {
      errors.push(
        getSingleFieldError(
          "expanded dimension",
          preset.exploreExpandedDimension,
        ),
      );
    }
  }

  return { partialExploreState, errors };
}

function fromTimeDimensionUrlParams(
  measures: Map<string, MetricsViewSpecMeasure>,
  preset: V1ExplorePreset,
): {
  partialExploreState: Partial<ExploreState>;
  errors: Error[];
} {
  if (!preset.timeDimensionMeasure) {
    return {
      partialExploreState: {
        tdd: {
          expandedMeasureName: "",
          chartType: TDDChart.DEFAULT,
          pinIndex: -1,
        },
      },
      errors: [],
    };
  }

  const errors: Error[] = [];

  let expandedMeasureName = preset.timeDimensionMeasure;
  if (expandedMeasureName && !measures.has(expandedMeasureName)) {
    expandedMeasureName = "";
    errors.push(getSingleFieldError("expanded measure", expandedMeasureName));
  }

  const partialExploreState: Partial<ExploreState> = {
    tdd: {
      expandedMeasureName,
      chartType: preset.timeDimensionChartType
        ? FromURLParamTDDChartMap[preset.timeDimensionChartType]
        : TDDChart.DEFAULT,
      pinIndex: preset.timeDimensionPin ? Number(preset.timeDimensionPin) : -1,
    },
  };

  return {
    partialExploreState,
    errors,
  };
}

function fromPivotUrlParams(
  measures: Map<string, MetricsViewSpecMeasure>,
  dimensions: Map<string, MetricsViewSpecDimension>,
  preset: V1ExplorePreset,
): {
  partialExploreState: Partial<ExploreState>;
  errors: Error[];
} {
  const errors: Error[] = [];

  const mapPivotEntry = (entry: string): PivotChipData | undefined => {
    if (entry in FromURLParamTimeDimensionMap) {
      const grain = FromURLParamTimeDimensionMap[entry];
      return {
        id: grain,
        title: TIME_GRAIN[grain]?.label,
        type: PivotChipType.Time,
      };
    }

    if (measures.has(entry)) {
      const m = measures.get(entry)!;
      return {
        id: entry,
        title: m.displayName || m.name || "Unknown",
        type: PivotChipType.Measure,
      };
    }

    if (dimensions.has(entry)) {
      const d = dimensions.get(entry)!;
      return {
        id: entry,
        title: d.displayName || d.name || "Unknown",
        type: PivotChipType.Dimension,
      };
    }

    errors.push(getSingleFieldError("pivot entry", entry));

    return undefined;
  };

  let hasSomePivotFields = false;

  const rowDimensions: PivotChipData[] = [];
  if (preset.pivotRows) {
    preset.pivotRows.forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      rowDimensions.push(chip);
    });
    hasSomePivotFields = true;
  }

  const colChips: PivotChipData[] = [];
  if (preset.pivotCols) {
    preset.pivotCols.forEach((pivotRow) => {
      const chip = mapPivotEntry(pivotRow);
      if (!chip) return;
      colChips.push(chip);
    });
    hasSomePivotFields = true;
  }

  const showPivot = preset.view === V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT;

  if (!hasSomePivotFields && !showPivot) {
    return {
      partialExploreState: {
        pivot: {
          rows: [],
          columns: [],
          sorting: [],
          expanded: {},
          columnPage: 1,
          rowPage: 1,
          enableComparison: true,
          activeCell: null,
          tableMode: "nest",
        },
      },
      errors,
    };
  }

  const sorting: SortingState = [];
  if (preset.pivotSortBy) {
    const sortById =
      preset.pivotSortBy in FromURLParamTimeDimensionMap
        ? FromURLParamTimeDimensionMap[preset.pivotSortBy]
        : preset.pivotSortBy;
    sorting.push({
      id: sortById,
      desc: !preset.pivotSortAsc,
    });
  }

  let tableMode: PivotTableMode = "nest";

  if (preset.pivotTableMode) {
    if (preset.pivotTableMode === "nest" || preset.pivotTableMode === "flat") {
      tableMode = preset.pivotTableMode;
    } else {
      errors.push(
        getSingleFieldError("pivot table mode", preset.pivotTableMode),
      );
    }
  }

  return {
    partialExploreState: {
      pivot: {
        rows: rowDimensions,
        columns: colChips,
        sorting,
        // TODO: other fields are not supported right now
        expanded: {},
        columnPage: 1,
        rowPage: 1,
        enableComparison: true,
        activeCell: null,
        tableMode,
      },
    },
    errors,
  };
}
