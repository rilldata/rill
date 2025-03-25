import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { type Readable, derived } from "svelte/store";
import type { TableSpec } from "./";

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
      const allMeasures =
        metricsView?.measures?.map((m) => m.name as string) || [];
      const allDimensions =
        metricsView?.dimensions?.map((d) => d.name || (d.column as string)) ||
        [];

      const columns = tableSpec?.columns || [];

      const measures = columns.filter((c) => allMeasures.includes(c));
      const dimensions = columns.filter((c) => allDimensions.includes(c));

      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }

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
    },
  );
}
