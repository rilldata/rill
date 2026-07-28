import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import type { DimensionFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters.ts";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters.ts";

export type MissingRequiredFilter = {
  key: string;
  name: string;
  label: string;
};

export function getMissingRequiredFilters(
  exploreSpec: V1ExploreSpec,
  dimensionFilters: DimensionFilterItem[],
  measureFilters: MeasureFilterItem[],
) {
  if (!exploreSpec.defaultPreset?.requiredFilters?.length) return [];

  const missing: MissingRequiredFilter[] = [];
  exploreSpec.defaultPreset.requiredFilters.forEach((name) => {
    const dimItem = dimensionFilters.find((d) => d.name === name);
    if (dimItem) {
      if (isDimensionItemEmpty(dimItem)) {
        missing.push({ key: name, name, label: dimItem.label || name });
      }
      return;
    }

    const measureItem = measureFilters.find((m) => m.name === name);
    if (measureItem) {
      if (isMeasureItemEmpty(measureItem)) {
        missing.push({ key: name, name, label: measureItem.label || name });
      }
      return;
    }

    // Required filter has no entry in the UI filter map at all => unsatisfied.
    missing.push({ key: name, name, label: name });
  });
  return missing;
}

function isDimensionItemEmpty(item: DimensionFilterItem): boolean {
  const hasSelectedValues = (item.selectedValues?.length ?? 0) > 0;
  const hasInputText = !!item.inputText && item.inputText.length > 0;
  return !hasSelectedValues && !hasInputText;
}

function isMeasureItemEmpty(item: MeasureFilterItem): boolean {
  return !item.filter;
}
