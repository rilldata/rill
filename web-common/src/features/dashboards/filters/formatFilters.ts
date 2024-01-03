import type { MetricsViewFilterCond } from "@rilldata/web-common/runtime-client";
import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import { getDisplayName } from "./getDisplayName";

export type DimensionFilter = {
  name: string;
  label: string;
  selectedValues: string[];
};

export function formatFilters(
  filters: MetricsViewFilterCond[] | undefined,
  dimensionIdMap: Map<string | number, MetricsViewSpecDimensionV2>
): DimensionFilter[] {
  if (!filters) return [];

  const formattedFilters: DimensionFilter[] = [];

  filters.forEach(({ name, in: selectedValues }) => {
    if (name === undefined) return;

    const formatted = {
      name,
      label: getDisplayName(dimensionIdMap.get(name)),
      selectedValues: selectedValues as string[],
    };

    formattedFilters.push(formatted);
  });

  return formattedFilters;
}
