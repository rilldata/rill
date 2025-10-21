import {
  validateDimensions,
  validateMeasures,
} from "@rilldata/web-common/features/canvas/components/validators";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import type { MapSpec } from "./";

export function validateMapSchema(
  metricsView: V1MetricsViewSpec | undefined,
  mapSpec: MapSpec,
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

  const dimension = mapSpec?.dimension;
  const measure = mapSpec?.measure;

  if (!dimension || !measure) {
    return {
      isValid: false,
      error: "Select both a dimension and a measure for the map",
    };
  }

  const validateMeasuresRes = validateMeasures(metricsView, [measure]);
  if (!validateMeasuresRes.isValid) {
    return {
      isValid: false,
      error: `Invalid measure "${measure}" selected for the map`,
    };
  }

  const validateDimensionsRes = validateDimensions(metricsView, [dimension]);
  if (!validateDimensionsRes.isValid) {
    return {
      isValid: false,
      error: `Invalid dimension "${dimension}" selected for the map`,
    };
  }

  return {
    isValid: true,
    error: undefined,
  };
}

