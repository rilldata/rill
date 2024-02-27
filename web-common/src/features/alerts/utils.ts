import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

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
