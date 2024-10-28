import { type AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { AllMeasureFilterTypeOptions } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

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

export function generateAlertName(
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
  if (!measure || !typeEntry) return undefined;

  const measureLabel =
    measure.displayName ?? measure.expression ?? (measure.name as string);

  let comparisonTitle = "";
  if (
    formValues.comparisonTimeRange?.isoOffset &&
    formValues.comparisonTimeRange.isoOffset in TIME_COMPARISON
  ) {
    const label = getComparisonLabel(
      formValues.comparisonTimeRange,
    ).toLowerCase();
    comparisonTitle = ` vs ${label}`;
  }

  return `${measureLabel} ${typeEntry.description}${comparisonTitle} alert`;
}
