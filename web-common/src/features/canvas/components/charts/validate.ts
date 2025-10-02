import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import { getFieldsByType } from "@rilldata/web-common/features/components/charts/util";
import type { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
import { derived, type Readable } from "svelte/store";

export function validateChartSchema(
  metricsView: MetricsViewSelectors,
  chartSpec: CanvasChartSpec,
): Readable<{
  isValid: boolean;
  error?: string;
  isLoading?: boolean;
}> {
  const { metrics_view } = chartSpec;

  const { measures, dimensions, timeDimensions } = getFieldsByType(chartSpec);

  return derived(
    metricsView.getMetricsViewFromName(metrics_view),
    (metricsViewQuery) => {
      if (metricsViewQuery.isLoading) {
        return {
          isValid: true,
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

      const timeDimension = metricsView.timeDimension;

      if (timeDimensions.length > 0 && timeDimension !== timeDimensions[0]) {
        return {
          isValid: false,
          error: `Invalid time dimension ${timeDimension} selected`,
        };
      }

      const validateMeasuresRes = validateMeasures(metricsView, measures);
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure ${invalidMeasures} selected`,
        };
      }

      const validateDimensionsRes = validateDimensions(metricsView, dimensions);

      if (!validateDimensionsRes.isValid) {
        const invalidDimensions =
          validateDimensionsRes.invalidDimensions.join(", ");

        return {
          isValid: false,
          error: `Invalid dimension(s) ${invalidDimensions} selected`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
