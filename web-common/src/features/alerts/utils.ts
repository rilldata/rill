import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
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

const IgnoredKeys: Partial<Record<keyof AlertFormValues, boolean>> = {
  whereFilter: true,
  timeRange: true,
};
export function getTouched(
  touchedRecord: Record<string, boolean | Array<any> | Record<string, any>>,
) {
  for (const key in touchedRecord) {
    if (key in IgnoredKeys) continue;
    if (typeof touchedRecord[key] === "boolean") {
      if (touchedRecord[key]) return true;
      continue;
    }
    if (typeof touchedRecord[key] !== "object") continue;
    if (Array.isArray(touchedRecord[key])) {
      if ((touchedRecord[key] as Array<any>).some((e) => getTouched(e)))
        return true;
      continue;
    }
    if (getTouched(touchedRecord[key] as Record<string, any>)) return true;
  }
  return false;
}
