import type { KPISpec } from "@rilldata/web-common/features/canvas/components/kpi";
import { validateMeasures } from "@rilldata/web-common/features/canvas/components/validators";
import type { StateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { derived, type Readable } from "svelte/store";

export function validateKPISchema(
  ctx: StateManagers,
  kpiSpec: KPISpec,
): Readable<{
  isValid: boolean;
  error?: string;
}> {
  const { metrics_view } = kpiSpec;
  return derived(
    ctx.canvasEntity.spec.getMetricsViewFromName(metrics_view),
    (metricsView) => {
      const measure = kpiSpec.measure;
      if (!metricsView) {
        return {
          isValid: false,
          error: `Metrics view ${metrics_view} not found`,
        };
      }
      const validateMeasuresRes = validateMeasures(metricsView, [measure]);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure "${invalidMeasures}" selected`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
