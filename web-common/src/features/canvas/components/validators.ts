import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export const isMeasureValid = (
  metricsViewSpec: V1MetricsViewSpec,
  measureName: string,
): boolean => {
  return (
    metricsViewSpec?.measures?.some((m) => m.name === measureName) || false
  );
};

export const isDimensionValid = (
  metricsViewSpec: V1MetricsViewSpec,
  dimensionName: string,
): boolean => {
  return (
    metricsViewSpec?.dimensions?.some((d) => d.name === dimensionName) || false
  );
};

export const validateMeasures = (
  metricsSpecQueryResult: V1MetricsViewSpec,
  measureNames: string[],
): { isValid: boolean; invalidMeasures: string[] } => {
  const invalidMeasures = measureNames.filter(
    (m) => !isMeasureValid(metricsSpecQueryResult, m),
  );
  return {
    isValid: invalidMeasures.length === 0,
    invalidMeasures,
  };
};

export const validateDimensions = (
  metricsSpecQueryResult: V1MetricsViewSpec,
  dimensionNames: string[],
): { isValid: boolean; invalidDimensions: string[] } => {
  const invalidDimensions = dimensionNames.filter(
    (d) => !isDimensionValid(metricsSpecQueryResult, d),
  );
  return {
    isValid: invalidDimensions.length === 0,
    invalidDimensions,
  };
};
