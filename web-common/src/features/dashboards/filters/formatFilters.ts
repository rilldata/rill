import type { MetricsViewFilterCond } from "@rilldata/web-common/runtime-client";
import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import { getDisplayName } from "./getDisplayName";

export type DimensionFilter = {
  name: string;
  label: string;
  selectedValues: any[];
  filterType: string;
};

export function formatFilters(
  filters: MetricsViewFilterCond[] | undefined,
  exclude: boolean,
  dimensionIdMap: Map<string | number, MetricsViewSpecDimensionV2>
): DimensionFilter[] {
  if (!filters) return [];

  const formatted: DimensionFilter[] = [];

  filters.forEach(({ name, in: selectedValues }) => {
    if (name === undefined) return;
    formatted.push({
      name,
      label: getDisplayName(
        dimensionIdMap.get(name) as MetricsViewSpecDimensionV2
      ),
      selectedValues: selectedValues as any[],
      filterType: exclude ? "exclude" : "include",
    });
  });

  return formatted;
}
