import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { QueryObserverResult } from "@tanstack/svelte-query";

export const isMeasureValid = (
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, HTTPError>,
  measureName: string,
): boolean => {
  return (
    metricsSpecQueryResult.data?.measures?.some(
      (m) => m.name === measureName,
    ) || false
  );
};

export const isDimensionValid = (
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, HTTPError>,
  dimensionName: string,
): boolean => {
  return (
    metricsSpecQueryResult.data?.dimensions?.some(
      (d) => d.name === dimensionName,
    ) || false
  );
};

export const validateMetricsView = (
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, HTTPError>,
) => {
  if (metricsSpecQueryResult.isError) {
    const errorMsg = metricsSpecQueryResult.error?.response?.data?.message;

    return {
      isValid: false,
      error: `Error: ${errorMsg}`,
    };
  }
  return {
    isValid: true,
    error: undefined,
  };
};

export const validateMeasures = (
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, HTTPError>,
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
  metricsSpecQueryResult: QueryObserverResult<V1MetricsViewSpec, HTTPError>,
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
