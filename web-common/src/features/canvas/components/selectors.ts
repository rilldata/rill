import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors";

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
