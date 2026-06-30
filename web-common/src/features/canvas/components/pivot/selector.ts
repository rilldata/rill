import { isTimeDimension } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type { PivotSpec, TableSpec } from ".";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export function validateTableSchema(
  metricsView: V1MetricsViewSpec | undefined,
  tableSpec: PivotSpec | TableSpec,
  isLoading?: boolean,
): {
  isValid: boolean;
  error?: string;
  isLoading?: boolean;
} {
  if (isLoading) {
    return { isValid: true, isLoading: true };
  }

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

  // Filter columns to only those accessible in the metrics view, silently
  // dropping any excluded by a security policy.
  const accessibleColumns = (tableSpec?.columns || []).filter(
    (c) => allMeasures.includes(c) || allDimensions.includes(c),
  );

  if (!accessibleColumns.length) {
    return {
      isValid: false,
      error: "Select at least one measure or dimension for the table",
    };
  }

  return {
    isValid: true,
    error: undefined,
  };
}

function validatePivot(tableSpec: PivotSpec, metricsView: V1MetricsViewSpec) {
  const allMeasures = metricsView?.measures?.map((m) => m.name as string) || [];
  const allDimensions =
    metricsView?.dimensions?.map((d) => d.name || (d.column as string)) || [];

  // Filter each list to only accessible fields, silently dropping any
  // excluded by a security policy.
  const measures = (tableSpec.measures || []).filter((m) =>
    allMeasures.includes(m),
  );
  const rowDimensions = (tableSpec.row_dimensions || []).filter(
    (d) =>
      allDimensions.includes(d) ||
      (metricsView.timeDimension && isTimeDimension(d, metricsView.timeDimension)),
  );
  const colDimensions = (tableSpec.col_dimensions || []).filter(
    (d) =>
      allDimensions.includes(d) ||
      (metricsView.timeDimension && isTimeDimension(d, metricsView.timeDimension)),
  );

  if (!measures.length && !rowDimensions.length && !colDimensions.length) {
    return {
      isValid: false,
      error: "Select at least one measure or dimension for the table",
    };
  }

  return {
    isValid: true,
    error: undefined,
  };
}
