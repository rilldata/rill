import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import {
  MeasureFilterBaseTypeOptions,
  MeasureFilterComparisonTypeOptions,
  MeasureFilterPercentOfTotalOption,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
import { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

export function getTypeOptions(
  formValues: AlertFormValues,
  selectedMeasure: MetricsViewSpecMeasureV2 | undefined,
) {
  const options: typeof MeasureFilterComparisonTypeOptions = [];

  if (
    formValues.comparisonTimeRange?.isoDuration ||
    formValues.comparisonTimeRange?.isoOffset
  ) {
    const comparisonLabel = getComparisonLabel(
      formValues.comparisonTimeRange,
    ).toLowerCase();
    options.push(
      ...MeasureFilterComparisonTypeOptions.map((o) => {
        if (
          o.value !== MeasureFilterType.AbsoluteChange &&
          o.value !== MeasureFilterType.PercentChange
        )
          return o;
        return {
          ...o,
          label: `${o.label} ${comparisonLabel}`,
        };
      }),
    );
  } else {
    options.push(...MeasureFilterBaseTypeOptions);
  }

  if (selectedMeasure?.validPercentOfTotal && formValues.splitByDimension) {
    options.push(MeasureFilterPercentOfTotalOption);
  }

  return options;
}
