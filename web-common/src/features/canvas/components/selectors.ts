import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";
import { MetricsViewSpecMeasureType } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export const useAllDimensionFromMetrics = (
  instanceId: string,
  metricsViewNames: string[],
) => {
  const dimensionsByViewStores = metricsViewNames.map((viewName) =>
    useAllDimensionFromMetric(instanceId, viewName),
  );
  return derived(dimensionsByViewStores, ($dimensionsByViewStores) =>
    $dimensionsByViewStores
      .filter((dimensions) => dimensions?.data)
      .map((dimensions) => dimensions.data)
      .flat(),
  );
};

export const useAllDimensionFromMetric = (
  instanceId: string,
  metricsViewName: string,
) =>
  useMetricsViewValidSpec(
    instanceId,
    metricsViewName,
    (meta) => meta?.dimensions ?? [],
  );

export const useAllMeasuresFromMetric = (
  instanceId: string,
  metricsViewName: string,
) =>
  useMetricsViewValidSpec(
    instanceId,
    metricsViewName,
    (meta) => meta?.measures ?? [],
  );

export const useAllSimpleMeasureFromMetric = (
  instanceId: string,
  metricsViewName: string,
) =>
  useMetricsViewValidSpec(
    instanceId,
    metricsViewName,
    (meta) =>
      meta?.measures?.filter(
        (m) =>
          !m.window &&
          m.type !== MetricsViewSpecMeasureType.MEASURE_TYPE_TIME_COMPARISON,
      ) ?? [],
  );

export const useMetricsViewSpecMeasure = (
  instanceId: string,
  metricsViewName: string,
  measureName: string,
) =>
  useMetricsViewValidSpec(instanceId, metricsViewName, (meta) =>
    meta?.measures?.find((measure) => measure.name === measureName),
  );

export const useMetricsViewSpecDimension = (
  instanceId: string,
  metricsViewName: string,
  dimensionName: string,
) =>
  useMetricsViewValidSpec(instanceId, metricsViewName, (meta) =>
    meta?.dimensions?.find(
      (dimension) =>
        dimension.name === dimensionName || dimension.column === dimensionName,
    ),
  );

export function useMeasureDimensionSpec(
  instanceId: string,
  metricsViewName: string,
  fields: { name: string; type: "dimension" | "measure" }[],
) {
  return fields.map((field) => {
    if (field.type === "measure") {
      return useMetricsViewSpecMeasure(instanceId, metricsViewName, field.name);
    } else {
      return useMetricsViewSpecDimension(
        instanceId,
        metricsViewName,
        field.name,
      );
    }
  });
}
