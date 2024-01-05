import { forEachExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import { getDisplayName } from "./getDisplayName";

export type DimensionFilter = {
  name: string;
  label: string;
  selectedValues: string[];
};

export function formatFilters(
  whereClause: V1Expression | undefined,
  dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
  potentialFilterName: string | null
): DimensionFilter[] {
  const filterItems = getDimensionFiltersFromWhereClause(
    whereClause,
    dimensionIdMap
  );
  // potential filter could be dimension or measure
  if (potentialFilterName && dimensionIdMap.has(potentialFilterName)) {
    filterItems.push({
      name: potentialFilterName,
      label: getDisplayName(dimensionIdMap.get(potentialFilterName)),
      selectedValues: [],
    });
  }
  // sort based on name to make sure toggling include/exclude is not jarring
  return filterItems.sort((a, b) => (a.name > b.name ? 1 : -1));
}

function getDimensionFiltersFromWhereClause(
  whereClause: V1Expression | undefined,
  dimensionIdMap: Map<string | number, MetricsViewSpecDimensionV2>
) {
  if (!whereClause) return [];
  const filteredDimensions = new Array<DimensionFilter>();
  const addedDimension = new Set<string>();

  forEachExpression(whereClause, (e) => {
    if (
      e.cond?.op !== V1Operation.OPERATION_IN &&
      e.cond?.op !== V1Operation.OPERATION_NIN
    ) {
      return;
    }
    const ident = e.cond?.exprs?.[0].ident;
    if (
      ident === undefined ||
      addedDimension.has(ident) ||
      !dimensionIdMap.has(ident)
    ) {
      return;
    }
    const dim = dimensionIdMap.get(ident);
    if (!dim) {
      return;
    }
    addedDimension.add(ident);
    filteredDimensions.push({
      name: ident,
      label: getDisplayName(dim),
      selectedValues: e.cond.exprs?.slice(1).map((e) => e.val) as any[],
    });
  });

  return filteredDimensions;
}
