import { base64ToProto } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertLegacyStateToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertLegacyStateToExplorePreset";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import {
  FromURLParamTimeDimensionMap,
  FromURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import { DashboardState } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimensionV2,
  type MetricsViewSpecMeasureV2,
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  type V1ExploreSpec,
  type V1Expression,
  type V1MetricsViewSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";

export function convertURLToExplorePreset(
  searchParams: URLSearchParams,
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

  // Support legacy dashboard param.
  // This will be applied 1st so that any newer params added can be applied as well.
  if (searchParams.has("state")) {
    const legacyState = searchParams.get("state") as string;
    const { preset: presetFromLegacyState, errors: errorsFromLegacyState } =
      fromLegacyStateUrlParam(legacyState, metricsView, explore, basePreset);
    Object.assign(preset, presetFromLegacyState);
    errors.push(...errorsFromLegacyState);
  }

  if (searchParams.has("vw")) {
    preset.view = FromURLParamViewMap[searchParams.get("vw") as string];
  }

  if (searchParams.has("f")) {
    const { expr, errors: filterErrors } = fromFilterUrlParam(
      searchParams.get("f") as string,
    );
    if (filterErrors) errors.push(...filterErrors);
    if (expr) preset.where = expr;
  }

  const { preset: trPreset, errors: trErrors } = fromTimeRangesParams(
    searchParams,
    dimensions,
  );
  Object.assign(preset, trPreset);
  errors.push(...trErrors);

  const { preset: ovPreset, errors: ovErrors } = fromOverviewUrlParams(
    searchParams,
    measures,
    dimensions,
    explore,
  );
  Object.assign(preset, ovPreset);
  errors.push(...ovErrors);

  const { preset: tddPreset, errors: tddErrors } = fromTimeDimensionUrlParams(
    searchParams,
    measures,
  );
  Object.assign(preset, tddPreset);
  errors.push(...tddErrors);

  const { preset: pivotPreset, errors: pivotErrors } = fromPivotUrlParams(
    searchParams,
    measures,
    dimensions,
  );
  Object.assign(preset, pivotPreset);
  errors.push(...pivotErrors);

  return { preset, errors };
}

function fromLegacyStateUrlParam(
  legacyState: string,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  basePreset: V1ExplorePreset,
) {
  try {
    legacyState = legacyState.includes("%")
      ? decodeURIComponent(legacyState)
      : legacyState;
    const legacyDashboardState = DashboardState.fromBinary(
      base64ToProto(legacyState),
    );

    return convertLegacyStateToExplorePreset(
      legacyDashboardState,
      metricsView,
      explore,
      basePreset,
    );
  } catch (e) {
    return {
      preset: {},
      errors: [e], // TODO: parse and show meaningful error
    };
  }
}

function fromFilterUrlParam(filter: string): {
  expr?: V1Expression;
  errors?: Error[];
} {
  try {
    let expr = convertFilterParamToExpression(filter);
    // if root is not AND/OR then add AND
    if (
      expr?.cond?.op !== V1Operation.OPERATION_AND &&
      expr?.cond?.op !== V1Operation.OPERATION_OR
    ) {
      expr = createAndExpression([expr]);
    }
    return { expr };
  } catch (e) {
    return { errors: [e] };
  }
}

function fromTimeRangesParams(
  searchParams: URLSearchParams,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has("tr")) {
    preset.timeRange = searchParams.get("tr") as string;
    // TODO: parse and return errors
  }
  if (searchParams.has("tz")) {
    preset.timezone = searchParams.get("tz") as string;
  }

  if (searchParams.has("ctr")) {
    preset.compareTimeRange = searchParams.get("ctr") as string;
    preset.comparisonMode ??=
      V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME;
    // TODO: parse and return errors
  }
  if (searchParams.has("cd")) {
    const comparisonDimension = searchParams.get("cd") as string;
    // unsetting a default from url
    if (comparisonDimension === "") {
      preset.comparisonDimension = "";
      preset.comparisonMode ??=
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_NONE;
    } else if (dimensions.has(comparisonDimension)) {
      preset.comparisonDimension = comparisonDimension;
      preset.comparisonMode ??=
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_DIMENSION;
    } else {
      errors.push(
        getSingleFieldError("compare dimension", comparisonDimension),
      );
    }
  }

  // TODO: grain from time range ?

  return { preset, errors };
}

function fromOverviewUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
  explore: V1ExploreSpec,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has("o.m")) {
    const mes = searchParams.get("o.m") as string;
    if (mes === "*") {
      preset.measures = explore.measures ?? [];
    } else {
      const selectedMeasures = mes.split(",").filter((m) => measures.has(m));
      preset.measures = selectedMeasures;
      const missingMeasures = getMissingValues(
        mes.split(","),
        selectedMeasures,
      );
      if (missingMeasures.length) {
        errors.push(getMultiFieldError("measure", missingMeasures));
      }
    }
  }

  if (searchParams.has("o.d")) {
    const dims = searchParams.get("o.d") as string;
    if (dims === "*") {
      preset.dimensions = explore.dimensions ?? [];
    } else {
      const selectedDimensions = dims
        .split(",")
        .filter((d) => dimensions.has(d));
      preset.dimensions = selectedDimensions;
      const missingDimensions = getMissingValues(
        dims.split(","),
        selectedDimensions,
      );
      if (missingDimensions.length) {
        errors.push(getMultiFieldError("dimension", missingDimensions));
      }
    }
  }

  if (searchParams.has("o.sb")) {
    const sortBy = searchParams.get("o.sb") as string;
    if (measures.has(sortBy)) {
      preset.overviewSortBy = sortBy;
    } else {
      errors.push(getSingleFieldError("sort by measure", sortBy));
    }
  }

  if (searchParams.has("o.sd")) {
    preset.overviewSortAsc = (searchParams.get("o.sd") as string) === "ASC";
  }

  if (searchParams.has("o.ed")) {
    const dim = searchParams.get("o.ed") as string;
    if (
      dimensions.has(dim) ||
      // we are unsetting from a default preset
      dim === ""
    ) {
      preset.overviewExpandedDimension = dim;
    } else {
      errors.push(getSingleFieldError("expanded dimension", dim));
    }
  }

  return { preset, errors };
}

function fromTimeDimensionUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has("tdd.m")) {
    const mes = searchParams.get("tdd.m") as string;
    if (
      measures.has(mes) ||
      // we are unsetting from a default preset
      mes === ""
    ) {
      preset.timeDimensionMeasure = mes;
    } else {
      errors.push(getSingleFieldError("expanded measure", mes));
    }
  }
  if (searchParams.has("tdd.ct")) {
    preset.timeDimensionChartType = searchParams.get("tdd.ct") as string;
  }
  if (searchParams.has("tdd.p")) {
    preset.timeDimensionPin = true;
  }

  return {
    preset,
    errors,
  };
}

function fromPivotUrlParams(
  searchParams: URLSearchParams,
  measures: Map<string, MetricsViewSpecMeasureV2>,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has("p.r")) {
    const rows = (searchParams.get("p.r") as string).split(",");
    const validRows = rows.filter(
      (r) => dimensions.has(r) || r in FromURLParamTimeDimensionMap,
    );
    preset.pivotRows = validRows;
    const missingRows = getMissingValues(validRows, rows);
    if (missingRows.length) {
      errors.push(getMultiFieldError("pivot row", missingRows));
    }
  }

  if (searchParams.has("p.c")) {
    const cols = (searchParams.get("p.c") as string).split(",");
    const validCols = cols.filter(
      (c) =>
        dimensions.has(c) ||
        measures.has(c) ||
        c in FromURLParamTimeDimensionMap,
    );
    preset.pivotCols = validCols;
    const missingCols = getMissingValues(validCols, cols);
    if (missingCols.length) {
      errors.push(getMultiFieldError("pivot column", missingCols));
    }
  }

  // TODO: other fields

  return { preset, errors };
}
