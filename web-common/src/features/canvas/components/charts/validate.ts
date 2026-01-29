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
          error: `Invalid time dimension ${timeDimension} in metrics view ${metrics_view}`,
        };
      }

      // Align field roles (measure vs dimension) with the metrics view spec.
      // This corrects cases where a field like `ad_size` is typed as a measure
      // in the chart spec but is actually a dimension in the metrics view.
      const measureNamesFromSpec = new Set(measures);
      const dimensionNamesFromSpec = new Set(dimensions);

      const metricMeasureNames = new Set(
        (metricsView.measures || []).map((m) => m.name),
      );
      const metricDimensionNames = new Set(
        (metricsView.dimensions || []).map((d) => d.name),
      );

      const correctedMeasures = new Set<string>();
      const correctedDimensions = new Set<string>(dimensions);

      // Re-map measures based on the metrics view definition
      for (const name of measureNamesFromSpec) {
        if (metricMeasureNames.has(name)) {
          correctedMeasures.add(name);
        } else if (metricDimensionNames.has(name)) {
          // Field is actually a dimension – treat it as such
          correctedDimensions.add(name);
        } else {
          // Unknown field: keep it as a measure so validation surfaces it
          correctedMeasures.add(name);
        }
      }

      // Also remap any dimensions that are actually measures
      for (const name of dimensionNamesFromSpec) {
        if (metricMeasureNames.has(name)) {
          correctedMeasures.add(name);
          correctedDimensions.delete(name);
        }
      }

      const correctedMeasuresArr = Array.from(correctedMeasures);
      const correctedDimensionsArr = Array.from(correctedDimensions);

      const validateMeasuresRes = validateMeasures(
        metricsView,
        correctedMeasuresArr,
      );
      if (!validateMeasuresRes.isValid) {
        const invalidMeasures = validateMeasuresRes.invalidMeasures.join(", ");
        return {
          isValid: false,
          error: `Invalid measure ${invalidMeasures} in metrics view ${metrics_view}`,
        };
      }

      const validateDimensionsRes = validateDimensions(
        metricsView,
        correctedDimensionsArr,
      );

      if (!validateDimensionsRes.isValid) {
        const invalidDimensions =
          validateDimensionsRes.invalidDimensions.join(", ");

        return {
          isValid: false,
          error: `Invalid dimension(s) ${invalidDimensions} in metrics view ${metrics_view}`,
        };
      }
      return {
        isValid: true,
        error: undefined,
      };
    },
  );
}
