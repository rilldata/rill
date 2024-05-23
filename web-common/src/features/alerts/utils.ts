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

  console.log(fieldName, metricsViewSpec.measures);
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
