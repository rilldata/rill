import { type AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { AllMeasureFilterTypeOptions } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import type { TaintedFields } from "sveltekit-superforms";

export function isSomeFieldTainted(taintedFields: TaintedFields<any>) {
  for (const key in taintedFields) {
    if (typeof taintedFields[key] === "boolean") {
      if (taintedFields[key]) return true;
      continue;
    }
    if (typeof taintedFields[key] !== "object") continue;
    if (Array.isArray(taintedFields[key])) {
      if ((taintedFields[key] as Array<any>).some((e) => isSomeFieldTainted(e)))
        return true;
      continue;
    }
    if (isSomeFieldTainted(taintedFields[key] as Record<string, any>))
      return true;
  }
}

export function generateAlertName(
  formValues: AlertFormValues,
  selectedComparisonTimeRange: DashboardTimeControls | undefined,
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
    measure.displayName || measure.expression || (measure.name as string);

  let comparisonTitle = "";
  if (
    selectedComparisonTimeRange?.name &&
    selectedComparisonTimeRange?.name in TIME_COMPARISON
  ) {
    const label =
      TIME_COMPARISON[selectedComparisonTimeRange.name].label.toLowerCase();
    comparisonTitle = ` vs ${label}`;
  }

  return `${measureLabel} ${typeEntry.description}${comparisonTitle} alert`;
}
