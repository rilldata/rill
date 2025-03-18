import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { isTimeDimension } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { type Readable, derived } from "svelte/store";

export function validateTableSchema(
  ctx: StateManagers,
  tableSpec: TableSpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view } = tableSpec;
  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      const measures = tableSpec.measures || [];
      const rowDimensions = tableSpec.row_dimensions || [];
      const colDimensions = tableSpec.col_dimensions || [];

      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }

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

      const validateDimensionsRes = validateDimensions(
        metricsView,
        allDimensions,
      );

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
    },
  );
}
