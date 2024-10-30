import { FromProtoTimeGrainMap } from "@rilldata/web-common/features/dashboards/proto-state/enum-maps";
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
import { mapLegacyChartType } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  FromURLParamTimeDimensionMap,
  ToURLParamTimeDimensionMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import type { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  type DashboardState,
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
  basePreset: V1ExplorePreset,
) {
  const preset: V1ExplorePreset = {
    ...basePreset,
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

  if (legacyState.filters) {
    // TODO
  } else if (legacyState.where) {
    preset.where = fromExpressionProto(legacyState.where);
  }
  if (legacyState.having) {
    preset.where ??= createAndExpression([]);
    const exprs = preset.where?.cond?.exprs as V1Expression[];
    legacyState.having.forEach((h) => {
      const expr = fromExpressionProto(h.filter);
      exprs.push(
        createSubQueryExpression(h.name, getAllIdentifiers(expr), expr),
      );
    });
  }

  const { entity: trEntity, errors: trErrors } = fromLegacyTimeRangeFields(
    legacyState,
    dimensions,
  );
  Object.assign(entity, trEntity);
  errors.push(...trErrors);

  const { preset: ovPreset, errors: ovErrors } = fromLegacyOverviewFields(
    legacyState,
    dimensions,
    measures,
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
    // TODO
    // preset.timeGrain = legacyState.timeGrain;
  }

  if (legacyState.compareTimeRange?.name) {
    preset.compareTimeRange = correctComparisonTimeRange(
      legacyState.compareTimeRange.name,
    );
    // TODO: custom time range
  }
  if (legacyState.showTimeComparison !== undefined) {
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
        getMultiFieldError("sort by measure", legacyState.leaderboardMeasure),
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
    // TODO
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
    if (dashboard.activePage) {
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
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  const mapTimeDimension = (grain: TimeGrain) =>
    ToURLParamTimeDimensionMap[FromProtoTimeGrainMap[grain]];
  const mapAllDimension = (dimension: PivotElement) => {
    if (dimension?.element.case === "pivotTimeDimension") {
      return mapTimeDimension(dimension?.element.value);
    } else {
      return dimension?.element.value as string;
    }
  };

  if (
    legacyState.pivotRowAllDimensions?.length &&
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
      ...legacyState.pivotRowTimeDimensions.map(mapTimeDimension),
      ...legacyState.pivotRowDimensions,
    ];
    preset.pivotCols = [
      ...legacyState.pivotColumnTimeDimensions.map(mapTimeDimension),
      ...legacyState.pivotColumnDimensions,
    ];
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

  // TODO: other fields

  return { preset, errors };
}
