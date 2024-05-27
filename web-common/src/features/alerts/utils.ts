import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export function getLabelForFieldName(
  metricsViewSpec: V1MetricsViewSpec,
  fieldName: string,
) {
  if (metricsViewSpec.dimensions) {
    for (const dimension of metricsViewSpec.dimensions) {
      if (dimension.name === fieldName) return dimension.label ?? fieldName;
    }
  }

  if (metricsViewSpec.measures) {
    for (const measure of metricsViewSpec.measures) {
      const label = measure.label ?? fieldName;
      switch (true) {
        case measure.name === fieldName:
          return label;
        case measure.name + ComparisonDeltaPreviousSuffix === fieldName:
          return `${label} (prev)`;
        case measure.name + ComparisonDeltaAbsoluteSuffix === fieldName:
          return `${label} (Δ)`;
        case measure.name + ComparisonDeltaRelativeSuffix === fieldName:
          return `${label} (Δ%)`;
      }
    }
  }

  return fieldName;
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
