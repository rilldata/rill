import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import { isTimeDimension } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type { PivotSpec, TableSpec } from "./";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export function validateTableSchema(
  metricsView: V1MetricsViewSpec | undefined,
  tableSpec: PivotSpec | TableSpec,
): {
  isValid: boolean;
  error?: string;
} {
  if (!metricsView) {
    return {
      isValid: false,
      error: `Metrics view not found`,
    };
  }

  if ("columns" in tableSpec) {
    return validateFlat(tableSpec, metricsView);
  } else {
    return validatePivot(tableSpec, metricsView);
  }
}

function validateFlat(tableSpec: TableSpec, metricsView: V1MetricsViewSpec) {
  const allMeasures = metricsView?.measures?.map((m) => m.name as string) || [];
  const allDimensions =
    metricsView?.dimensions?.map((d) => d.name || (d.column as string)) || [];
  const columns = tableSpec?.columns || [];

  const measures = columns.filter((c) => allMeasures.includes(c));
  const dimensions = columns.filter((c) => allDimensions.includes(c));

  if (!columns.length) {
    return {
      isValid: false,
      error: "Select at least one measure or dimension for the table",
    };
  }
  const validateMeasuresRes = validateMeasures(metricsView, measures);
  if (!validateMeasuresRes.isValid) {
    const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
    return {
      isValid: false,
      error: `Invalid measure(s) "${invalidMeasures}" selected for the table`,
    };
  }

  const validateDimensionsRes = validateDimensions(metricsView, dimensions);

  if (!validateDimensionsRes.isValid) {
    const invalidDimensions =
      validateDimensionsRes.invalidDimensions.join(", ");

    return {
      isValid: false,
      error: `Invalid dimension(s) "${invalidDimensions}" selected for the table`,
    };
  }
  return {
    isValid: true,
    error: undefined,
  };
}

function validatePivot(tableSpec: PivotSpec, metricsView: V1MetricsViewSpec) {
  const measures = tableSpec.measures || [];
  const rowDimensions = tableSpec.row_dimensions || [];
  const colDimensions = tableSpec.col_dimensions || [];

  if (!measures.length && !rowDimensions.length && !colDimensions.length) {
    return {
      isValid: false,
      error: "Select at least one measure or dimension for the table",
    };
  }
  const validateMeasuresRes = validateMeasures(metricsView, measures);
  if (!validateMeasuresRes.isValid) {
    const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
    return {
      isValid: false,
      error: `Invalid measure(s) "${invalidMeasures}" selected for the table`,
    };
  }

  const allDimensions = rowDimensions
    .concat(colDimensions)
    .filter(
      (d) =>
        !metricsView.timeDimension ||
        !isTimeDimension(d, metricsView.timeDimension),
    );

  const validateDimensionsRes = validateDimensions(metricsView, allDimensions);

  if (!validateDimensionsRes.isValid) {
    const invalidDimensions =
      validateDimensionsRes.invalidDimensions.join(", ");

    return {
      isValid: false,
      error: `Invalid dimension(s) "${invalidDimensions}" selected for the table`,
    };
  }
  return {
    isValid: true,
    error: undefined,
  };
}
