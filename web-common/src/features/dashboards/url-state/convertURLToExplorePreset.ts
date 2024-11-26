import { base64ToProto } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertLegacyStateToExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/convertLegacyStateToExplorePreset";
import {
  getMultiFieldError,
  getSingleFieldError,
} from "@rilldata/web-common/features/dashboards/url-state/error-message-helpers";
import {
  convertFilterParamToExpression,
  stripParserError,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import {
  FromURLParamTimeDimensionMap,
  FromURLParamTimeGrainMap,
  FromURLParamTimeRangePresetMap,
  FromURLParamViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import {
  getMapFromArray,
  getMissingValues,
} from "@rilldata/web-common/lib/arrayUtils";
import {
  TIME_COMPARISON,
  TIME_GRAIN,
} from "@rilldata/web-common/lib/time/config";
import { validateISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
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

  // Support legacy dashboard param.
  // This will be applied 1st so that any newer params added can be applied as well.
  if (searchParams.has("state")) {
    const legacyState = searchParams.get("state") as string;
    const { preset: presetFromLegacyState, errors: errorsFromLegacyState } =
      fromLegacyStateUrlParam(
        legacyState,
        metricsView,
        explore,
        defaultExplorePreset,
      );
    Object.assign(preset, presetFromLegacyState);
    errors.push(...errorsFromLegacyState);
  }

  if (searchParams.has("view")) {
    const view = searchParams.get("view") as string;
    if (view in FromURLParamViewMap) {
      preset.view = FromURLParamViewMap[view];
    } else {
      errors.push(getSingleFieldError("view", view));
    }
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
  defaultExplorePreset: V1ExplorePreset,
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
      defaultExplorePreset,
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
    return {
      errors: [new Error("Selected filter is invalid: " + stripParserError(e))],
    };
  }
}

function fromTimeRangesParams(
  searchParams: URLSearchParams,
  dimensions: Map<string, MetricsViewSpecDimensionV2>,
) {
  const preset: V1ExplorePreset = {};
  const errors: Error[] = [];

  if (searchParams.has("tr")) {
    const tr = searchParams.get("tr") as string;
    if (tr in FromURLParamTimeRangePresetMap || validateISODuration(tr)) {
      preset.timeRange = tr;
    } else {
      errors.push(getSingleFieldError("time range", tr));
    }
  }

  if (searchParams.has("tz")) {
    preset.timezone = searchParams.get("tz") as string;
  }

  if (searchParams.has("compare_tr")) {
    const ctr = searchParams.get("compare_tr") as string;
    if (ctr in TIME_COMPARISON) {
      preset.compareTimeRange = ctr;
      preset.comparisonMode ??=
        V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME;
    } else {
      errors.push(getSingleFieldError("compare time range", ctr));
    }
  }

  if (searchParams.has("grain")) {
    const tg = searchParams.get("grain") as string;
    if (tg in FromURLParamTimeGrainMap) {
      preset.timeGrain = tg;
    } else {
      errors.push(getSingleFieldError("time grain", tg));
    }
  }

  if (searchParams.has("compare_dim")) {
    const comparisonDimension = searchParams.get("compare_dim") as string;
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

  // TODO: scrub range

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

  if (searchParams.has("measures")) {
    const mes = searchParams.get("measures") as string;
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

  if (searchParams.has("dims")) {
    const dims = searchParams.get("dims") as string;
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

  if (searchParams.has("sort_by")) {
    const sortBy = searchParams.get("sort_by") as string;
    if (measures.has(sortBy)) {
      preset.overviewSortBy = sortBy;
    } else {
      errors.push(getSingleFieldError("sort by measure", sortBy));
    }
  }

  if (searchParams.has("sort_dir")) {
    preset.overviewSortAsc = (searchParams.get("sort_dir") as string) === "ASC";
  }

  if (searchParams.has("expanded_dim")) {
    const dim = searchParams.get("expanded_dim") as string;
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

  if (searchParams.has("measure")) {
    const mes = searchParams.get("measure") as string;
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
  if (searchParams.has("chart_type")) {
    preset.timeDimensionChartType = searchParams.get("chart_type") as string;
  }
  if (searchParams.has("pin")) {
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

  if (searchParams.has("rows")) {
    const rows = (searchParams.get("rows") as string).split(",");
    const validRows = rows.filter(
      (r) => dimensions.has(r) || r in FromURLParamTimeDimensionMap,
    );
    preset.pivotRows = validRows;
    const missingRows = getMissingValues(validRows, rows);
    if (missingRows.length) {
      errors.push(getMultiFieldError("pivot row", missingRows));
    }
  }

  if (searchParams.has("cols")) {
    const cols = (searchParams.get("cols") as string).split(",");
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
