import { CriteriaOperationOptions } from "@rilldata/web-common/features/alerts/criteria-tab/operations";
import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaPreviousSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  AllMeasureFilterTypeOptions,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
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

export function getTitleAndDescription(
  formValues: AlertFormValues,
  metricsViewSpec: V1MetricsViewSpec,
) {
  const firstCriteria = formValues.criteria[0];
  if (!firstCriteria) return undefined;

  const measure = metricsViewSpec.measures?.find(
    (m) => m.name === firstCriteria.measure,
  );
  const typeEntry = AllMeasureFilterTypeOptions.find(
    (o) => o.value === firstCriteria.type,
  );
  const operationEntry = CriteriaOperationOptions.find(
    (o) => o.value === firstCriteria.operation,
  );
  if (!measure || !typeEntry || !operationEntry) return undefined;

  const measureLabel =
    measure.label ?? measure.expression ?? (measure.name as string);

  let value = firstCriteria.value1;
  if (firstCriteria.type === MeasureFilterType.PercentChange) {
    value = `${value}%`;
  }

  let comparisonTitle = "";
  let comparisonDescription = "";
  if (
    formValues.comparisonTimeRange?.isoOffset &&
    formValues.comparisonTimeRange.isoOffset in TIME_COMPARISON
  ) {
    const label =
      TIME_COMPARISON[
        formValues.comparisonTimeRange.isoOffset
      ].label.toLowerCase();
    comparisonTitle = ` vs ${label}`;
    comparisonDescription = ` compared to the ${label}`;
  }

  return {
    title: `${measureLabel} ${typeEntry.description}${comparisonTitle} alert`,
    description: `Triggers when ${measureLabel} ${typeEntry.description}${comparisonDescription} is ${operationEntry.description} ${value}`,
  };
}
