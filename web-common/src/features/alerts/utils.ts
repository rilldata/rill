import type { V1MetricsViewSpec } from "../../runtime-client";

export function getLabelForFieldName(
  metricsViewSpec: V1MetricsViewSpec,
  fieldName: string,
) {
  const measureLabel = metricsViewSpec.measures?.find(
    (m) => m.name === fieldName,
  )?.label;
  const dimensionLabel = metricsViewSpec.dimensions?.find(
    (d) => d.name === fieldName,
  )?.label;

  return measureLabel || dimensionLabel || fieldName;
}
