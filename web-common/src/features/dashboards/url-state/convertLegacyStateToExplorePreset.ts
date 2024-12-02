import { FromProtoTimeGrainMap } from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";
import { convertFilterToExpression } from "@rilldata/web-common/features/dashboards/proto-state/filter-converter";
import {
  correctComparisonTimeRange,
  fromExpressionProto,
} from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import {
  createAndExpression,
  createSubQueryExpression,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  ExplorePresetDefaultChartType,
  URLStateDefaultTimezone,
} from "@rilldata/web-common/features/dashboards/url-state/defaults";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  FromLegacySortTypeMap,
  mapLegacyChartType,
} from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  FromActivePageMap,
  FromURLParamTimeDimensionMap,
  ToURLParamTimeDimensionMap,
  ToURLParamTimeGrainMapMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import type { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  type DashboardState,
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  PivotElement,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export function convertLegacyStateToExplorePreset(
  legacyState: DashboardState,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  defaultExplorePreset: V1ExplorePreset,
) {
  const preset: V1ExplorePreset = {
    ...defaultExplorePreset,
  };
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

  if (legacyState.activePage !== DashboardState_ActivePage.UNSPECIFIED) {
    preset.view = FromActivePageMap[legacyState.activePage];
  }

  if (legacyState.filters) {
    // backwards compatibility for our older filter format
    preset.where = convertFilterToExpression(legacyState.filters);
    // TODO: correct older values that would have strings for non-strings
  } else if (legacyState.where) {
    preset.where = fromExpressionProto(legacyState.where);
  }
  if (legacyState.having) {
    preset.where ??= createAndExpression([]);
    const exprs = preset.where?.cond?.exprs as V1Expression[];
    legacyState.having.forEach((h) => {
      if (!h.filter) return;
      const expr = fromExpressionProto(h.filter);
      exprs.push(
        createSubQueryExpression(h.name, getAllIdentifiers(expr), expr),
      );
    });
  }

  const { preset: trPreset, errors: trErrors } = fromLegacyTimeRangeFields(
    legacyState,
    dimensions,
  );
  Object.assign(preset, trPreset);
  errors.push(...trErrors);

  const { preset: ovPreset, errors: ovErrors } = fromLegacyOverviewFields(
    legacyState,
    measures,
    dimensions,
    explore,
  );
  Object.assign(preset, ovPreset);
  errors.push(...ovErrors);

  const { preset: tddPreset, errors: tddErrors } =
    fromLegacyTimeDimensionFields(legacyState, measures);
  Object.assign(preset, tddPreset);
  errors.push(...tddErrors);

  const { preset: pivotPreset, errors: pivotErrors } = fromLegacyPivotFields(
    legacyState,
    measures,
    dimensions,
  );
  Object.assign(preset, pivotPreset);
  errors.push(...pivotErrors);

  return { preset, errors };
}

function fromLegacyTimeRangeFields(
  legacyState: DashboardState,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (legacyState.timeRange?.name) {
    preset.timeRange = legacyState.timeRange.name;
    // TODO: custom time range
  }
  if (legacyState.timeGrain) {
    preset.timeGrain =
      ToURLParamTimeGrainMapMap[FromProtoTimeGrainMap[legacyState.timeGrain]];
  }

  if (legacyState.compareTimeRange?.name) {
    preset.compareTimeRange = correctComparisonTimeRange(
      legacyState.compareTimeRange.name,
    );
    // TODO: custom time range
  }
  if (legacyState.showTimeComparison) {
    preset.comparisonMode =
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME;
  }

  if (legacyState.comparisonDimension) {
    if (dimensions.has(legacyState.comparisonDimension)) {
      preset.comparisonDimension = legacyState.comparisonDimension;
      preset.comparisonMode =
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION;
    } else {
      errors.push(
        getSingleFieldError(
          "compare dimension",
          legacyState.comparisonDimension,
        ),
      );
    }
  } else {
    // older state would unset comparison dimension when empty
    preset.comparisonDimension = "";
  }

  if (!preset.comparisonMode) {
    // if there was no comparison in legacyState it was an unset to `None`
    preset.comparisonMode =
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE;
  }

  preset.timezone = legacyState.selectedTimezone ?? URLStateDefaultTimezone;

  // TODO: scrubRange

  return { preset, errors };
}

function fromLegacyOverviewFields(
  legacyState: DashboardState,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (legacyState.allMeasuresVisible) {
    preset.measures = explore.measures ?? [];
  } else if (legacyState.visibleMeasures?.length) {
    preset.measures = legacyState.visibleMeasures.filter((m) =>
      measures.has(m),
    );
    const missingMeasures = getMissingValues(
      legacyState.visibleMeasures,
      preset.measures,
    );
    if (missingMeasures.length) {
      errors.push(getMultiFieldError("measure", missingMeasures));
    }
  }

  if (legacyState.allDimensionsVisible) {
    preset.dimensions = explore.dimensions ?? [];
  } else if (legacyState.visibleDimensions?.length) {
    preset.dimensions = legacyState.visibleDimensions.filter((d) =>
      dimensions.has(d),
    );
    const missingDimensions = getMissingValues(
      legacyState.visibleDimensions,
      preset.dimensions,
    );
    if (missingDimensions.length) {
      errors.push(getMultiFieldError("dimension", missingDimensions));
    }
  }

  if (legacyState.leaderboardMeasure !== undefined) {
    if (measures.has(legacyState.leaderboardMeasure)) {
      preset.overviewSortBy = legacyState.leaderboardMeasure;
    } else {
      errors.push(
        getSingleFieldError("sort by measure", legacyState.leaderboardMeasure),
      );
    }
  }

  if (legacyState.leaderboardSortDirection) {
    preset.overviewSortAsc =
      legacyState.leaderboardSortDirection ===
      DashboardState_LeaderboardSortDirection.ASCENDING;
  }

  if (legacyState.leaderboardContextColumn !== undefined) {
    // TODO
  }

  if (legacyState.leaderboardSortType) {
    preset.overviewSortType =
      FromLegacySortTypeMap[legacyState.leaderboardSortType];
  }

  if (legacyState.selectedDimension) {
    if (dimensions.has(legacyState.selectedDimension)) {
      preset.overviewExpandedDimension = legacyState.selectedDimension;
    } else {
      errors.push(
        getSingleFieldError(
          "expanded dimension",
          legacyState.selectedDimension,
        ),
      );
    }
  } else if (legacyState.activePage !== DashboardState_ActivePage.UNSPECIFIED) {
    // UNSPECIFIED means it was a partial state stored to proto state
    // So anything other than that would need to unset this
    preset.overviewExpandedDimension = "";
  }

  return { preset, errors };
}

function fromLegacyTimeDimensionFields(
  legacyState: DashboardState,
  measures: Map<string, MetricsViewSpecMeasureV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (!legacyState.expandedMeasure) {
    if (legacyState.activePage) {
      preset.timeDimensionMeasure = "";
      preset.timeDimensionPin = false;
      preset.timeDimensionChartType = ExplorePresetDefaultChartType;
    }
    return { preset, errors };
  }

  if (!measures.has(legacyState.expandedMeasure)) {
    errors.push(
      getSingleFieldError("expanded measure", legacyState.expandedMeasure),
    );
    return { preset, errors };
  }

  preset.timeDimensionMeasure = legacyState.expandedMeasure;
  preset.timeDimensionPin = false; // TODO
  preset.timeDimensionChartType = mapLegacyChartType(legacyState.chartType);

  return { preset, errors };
}

function fromLegacyPivotFields(
  legacyState: DashboardState,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  const mapTimeDimension = (grain: TimeGrain) =>
    ToURLParamTimeDimensionMap[FromProtoTimeGrainMap[grain]] ?? "";
  const mapAllDimension = (dimension: PivotElement) => {
    if (dimension?.element.case === "pivotTimeDimension") {
      return mapTimeDimension(dimension?.element.value);
    } else {
      return dimension?.element.value as string;
    }
  };

  if (
    legacyState.pivotRowAllDimensions?.length ||
    legacyState.pivotColumnAllDimensions?.length
  ) {
    preset.pivotRows = legacyState.pivotRowAllDimensions.map(mapAllDimension);
    preset.pivotCols =
      legacyState.pivotColumnAllDimensions.map(mapAllDimension);
  } else if (
    // backwards compatibility for state
    legacyState.pivotRowDimensions?.length ||
    legacyState.pivotRowTimeDimensions?.length ||
    legacyState.pivotColumnDimensions?.length ||
    legacyState.pivotColumnTimeDimensions?.length
  ) {
    preset.pivotRows = [
      ...legacyState.pivotRowTimeDimensions
        .map(mapTimeDimension)
        .filter(Boolean),
      ...legacyState.pivotRowDimensions,
    ];
    preset.pivotCols = [
      ...legacyState.pivotColumnTimeDimensions
        .map(mapTimeDimension)
        .filter(Boolean),
      ...legacyState.pivotColumnDimensions,
    ];
  }

  if (legacyState.pivotColumnMeasures?.length) {
    preset.pivotCols ??= [];
    preset.pivotCols.push(...legacyState.pivotColumnMeasures);
  }

  if (preset.pivotRows?.length) {
    const allValues = preset.pivotRows;
    preset.pivotRows = preset.pivotRows.filter(
      (r) => dimensions.has(r) || r in FromURLParamTimeDimensionMap,
    );
    const missingRows = getMissingValues(allValues, preset.pivotRows);
    if (missingRows.length) {
      errors.push(getMultiFieldError("pivot row", missingRows));
    }
  }

  if (preset.pivotCols?.length) {
    const allValues = preset.pivotCols;
    preset.pivotCols = preset.pivotCols.filter(
      (c) =>
        dimensions.has(c) ||
        measures.has(c) ||
        c in FromURLParamTimeDimensionMap,
    );
    const missingCols = getMissingValues(allValues, preset.pivotCols);
    if (missingCols.length) {
      errors.push(getMultiFieldError("pivot column", missingCols));
    }
  }

  if (
    legacyState.activePage !== DashboardState_ActivePage.PIVOT &&
    // UNSPECIFIED means it was a partial state stored to proto state
    legacyState.activePage !== DashboardState_ActivePage.UNSPECIFIED
  ) {
    // legacy state would unset when active page is not pivot
    preset.pivotRows = [];
    preset.pivotCols = [];
  }

  const sortBy = legacyState.pivotSort?.[0];
  if (sortBy) {
    preset.pivotSortBy =
      sortBy.id in ToURLParamTimeDimensionMap
        ? ToURLParamTimeDimensionMap[sortBy.id]
        : sortBy.id;
    preset.pivotSortAsc = !sortBy.desc;
  }

  // TODO: other fields like expanded state and pin are not supported right now
  return { preset, errors };
}
