import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  MeasureFilterBaseTypeOptions,
  MeasureFilterComparisonTypeOptions,
  MeasureFilterPercentOfTotalOption,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config.ts";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

export function getTypeOptions(
  formValues: AlertFormValues,
  selectedComparisonTimeRange: DashboardTimeControls | undefined,
  selectedMeasure: MetricsViewSpecMeasure | undefined,
) {
  const options: {
    value: string;
    label: string;
    disabled?: boolean;
    tooltip?: string;
  }[] = [...MeasureFilterBaseTypeOptions];

  if (
    selectedComparisonTimeRange?.name &&
    selectedComparisonTimeRange?.name in TIME_COMPARISON
  ) {
    const comparisonLabel =
      TIME_COMPARISON[selectedComparisonTimeRange.name].label.toLowerCase();
    options.push(
      ...MeasureFilterComparisonTypeOptions.map((o) => {
        return {
          ...o,
          label: `${o.label} ${comparisonLabel}`,
        };
      }),
    );
  } else {
    options.push(
      ...MeasureFilterComparisonTypeOptions.map((o) => {
        return {
          ...o,
          label: o.shortLabel,
          tooltip: "Available when comparing time periods.",
          disabled: true,
        };
      }),
    );
  }

  if (selectedMeasure?.validPercentOfTotal && formValues.splitByDimension) {
    options.push(MeasureFilterPercentOfTotalOption);
  } else {
    options.push({
      ...MeasureFilterPercentOfTotalOption,
      tooltip: "Measure does not support percent-of-total.",
      disabled: true,
    });
  }

  return options;
}
