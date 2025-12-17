import type { GaugeSpec } from "@rilldata/web-common/features/canvas/components/gauge";
import { validateMeasures } from "@rilldata/web-common/features/canvas/components/validators";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { derived, type Readable } from "svelte/store";

export function validateGaugeSchema(
  ctx: CanvasStore,
  gaugeSpec: GaugeSpec,
): Readable<{
  isValid: boolean;
  error?: string;
  isLoading?: boolean;
}> {
  const { metrics_view } = gaugeSpec;
  return derived(
    ctx.canvasEntity.metricsView.getMetricsViewFromName(metrics_view),
    (metricsViewQuery) => {
      const measure = gaugeSpec.measure;
      if (metricsViewQuery.isLoading) {
        return {
          isValid: true,
          error: undefined,
          isLoading: true,
        };
      }
      const metricsView = metricsViewQuery.metricsView;
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

