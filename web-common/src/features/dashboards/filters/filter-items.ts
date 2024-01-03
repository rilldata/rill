import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import { forEachExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
import {
  MetricsViewSpecMeasureV2,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import type {
  MetricsViewSpecDimensionV2,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import { writable } from "svelte/store";

export const potentialFilterName = writable<string | null>(null);

export type DimensionFilterItem = {
  name: string;
  label: string;
  selectedValues: string[];
};

export function getDimensionFilterItems(
  whereClause: V1Expression | undefined,
  dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
  potentialFilterName: string | null
) {
  const filterItems = getDimensionFiltersFromWhereClause(
    whereClause,
    dimensionIdMap
  );
  // potential filter could be dimension or measure
  if (potentialFilterName && dimensionIdMap.has(potentialFilterName)) {
    filterItems.push({
      name: potentialFilterName,
      label: getDimensionDisplayName(dimensionIdMap.get(potentialFilterName)),
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
  const filteredDimensions = new Array<DimensionFilterItem>();
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
      label: getDimensionDisplayName(dim),
      selectedValues: e.cond.exprs?.slice(1).map((e) => e.val) as any[],
    });
  });

  return filteredDimensions;
}

export type MeasureFilterItem = {
  name: string;
  label: string;
  expr?: V1Expression;
};

export function getMeasureFilterItems(
  havingClause: V1Expression | undefined,
  measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
  potentialFilterName: string | null
) {
  const filterItems = getMeasureFiltersFromHavingClause(
    havingClause,
    measureIdMap
  );
  // potential filter could be dimension or measure
  if (potentialFilterName && measureIdMap.has(potentialFilterName)) {
    filterItems.push({
      name: potentialFilterName,
      label: getMeasureDisplayName(measureIdMap.get(potentialFilterName)),
    });
  }
  // sort based on name to make sure toggling include/exclude is not jarring
  return filterItems.sort((a, b) => (a.name > b.name ? 1 : -1));
}

function getMeasureFiltersFromHavingClause(
  havingClause: V1Expression | undefined,
  measureIdMap: Map<string, MetricsViewSpecMeasureV2>
) {
  if (!havingClause) return [];

  const filteredMeasures = new Array<MeasureFilterItem>();
  const addedMeasure = new Set<string>();
  forEachExpression(havingClause, (e, depth) => {
    if (!e.cond?.exprs?.length) {
      return;
    }
    const ident =
      e.cond?.exprs?.[0].ident ||
      (depth === 0 && e.cond?.exprs?.[0]?.cond?.exprs?.length
        ? e.cond?.exprs?.[0]?.cond?.exprs?.[0].ident
        : undefined);
    if (
      ident === undefined ||
      addedMeasure.has(ident) ||
      !measureIdMap.has(ident)
    ) {
      return;
    }
    const measure = measureIdMap.get(ident);
    if (!measure) {
      return;
    }
    addedMeasure.add(ident);
    filteredMeasures.push({
      name: ident,
      label: getMeasureDisplayName(measure),
      expr: e,
    });
  });
  return filteredMeasures;
}
