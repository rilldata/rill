import { forEachExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";
import type { MetricsViewSpecDimensionV2 } from "@rilldata/web-common/runtime-client";
import { getDisplayName } from "./getDisplayName";

export type DimensionFilter = {
  name: string;
  label: string;
  selectedValues: string[];
};

export function formatFilters(
  whereFilter: V1Expression | undefined,
  dimensionIdMap: Map<string | number, MetricsViewSpecDimensionV2>
) {
  if (!whereFilter) return [];
  const filteredDimensions = new Array<DimensionFilter>();
  const addedDimension = new Set<string>();

  forEachExpression(whereFilter, (e) => {
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
