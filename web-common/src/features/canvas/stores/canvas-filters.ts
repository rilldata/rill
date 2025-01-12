import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { DimensionFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import { filterItemsSortFunction } from "@rilldata/web-common/features/dashboards/state-managers/selectors/filters";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import {
  createAndExpression,
  createInExpression,
  forEachIdentifier,
  getValueIndexInExpression,
  getValuesInExpression,
  isExpressionUnsupported,
  matchExpressionByName,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type {
  MetricsViewSpecDimensionV2,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  V1Operation,
  type MetricsViewSpecMeasureV2,
} from "@rilldata/web-common/runtime-client";
import { get, writable, type Writable } from "svelte/store";

export class CanvasFilters {
  whereFilter: Writable<V1Expression>;
  dimensionThresholdFilters: Writable<Array<DimensionThresholdFilter>>;
  dimensionFilterExcludeMode: Writable<Map<string, boolean>>;
  temporaryFilterName: Writable<string | null>;

  constructor() {
    this.dimensionFilterExcludeMode = writable(new Map<string, boolean>());
    this.temporaryFilterName = writable(null);
    this.whereFilter = writable({});
    this.dimensionThresholdFilters = writable([]);
  }

  // ----- MEASURE FILTER SELECTORS -----
  measureHasFilter(measureName: string) {
    const dtfs = get(this.dimensionThresholdFilters);
    return dtfs.some((dtf) =>
      dtf.filters.some((f) => f.measure === measureName),
    );
  }

  getMeasureFilterItems(measureIdMap: Map<string, MetricsViewSpecMeasureV2>) {
    return this.getMeasureFilters(
      measureIdMap,
      get(this.dimensionThresholdFilters),
    );
  }

  private getMeasureFilters(
    measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
    dimensionThresholdFilters: DimensionThresholdFilter[],
  ): MeasureFilterItem[] {
    const filteredMeasures: MeasureFilterItem[] = [];
    const addedMeasure = new Set<string>();
    for (const dtf of dimensionThresholdFilters) {
      filteredMeasures.push(
        ...this.getMeasureFilterForDimension(
          measureIdMap,
          dtf.filters,
          dtf.name,
          addedMeasure,
        ),
      );
    }
    return filteredMeasures;
  }

  private getMeasureFilterForDimension(
    measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
    filters: MeasureFilterEntry[],
    name: string,
    addedMeasure: Set<string>,
  ): MeasureFilterItem[] {
    if (!filters.length) return [];
    const filteredMeasures: MeasureFilterItem[] = [];
    filters.forEach((filter) => {
      if (addedMeasure.has(filter.measure)) return;
      const measure = measureIdMap.get(filter.measure);
      if (!measure) return;
      addedMeasure.add(filter.measure);
      filteredMeasures.push({
        dimensionName: name,
        name: filter.measure,
        label: measure.displayName || measure.expression || filter.measure,
        filter,
      });
    });
    return filteredMeasures;
  }

  getAllMeasureFilterItems(
    measureFilterItems: MeasureFilterItem[],
    measureIdMap: Map<string, MetricsViewSpecMeasureV2>,
  ) {
    const itemsCopy = [...measureFilterItems];
    const tempFilterName = get(this.temporaryFilterName);
    if (tempFilterName && measureIdMap.has(tempFilterName)) {
      itemsCopy.push({
        dimensionName: "",
        name: tempFilterName,
        label: getMeasureDisplayName(measureIdMap.get(tempFilterName)),
      });
    }
    return itemsCopy;
  }

  // ----- MEASURE FILTER ACTIONS -----
  setMeasureFilter(dimensionName: string, filter: MeasureFilterEntry) {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter !== null) {
      this.temporaryFilterName.set(null);
    }

    const dtfs = get(this.dimensionThresholdFilters);
    let dimThresholdFilter = dtfs.find((dtf) => dtf.name === dimensionName);
    if (!dimThresholdFilter) {
      dimThresholdFilter = { name: dimensionName, filters: [] };
      dtfs.push(dimThresholdFilter);
    }
    const exprIdx = dimThresholdFilter.filters.findIndex(
      (f) => f.measure === filter.measure,
    );
    if (exprIdx === -1) {
      dimThresholdFilter.filters.push(filter);
    } else {
      dimThresholdFilter.filters.splice(exprIdx, 1, filter);
    }
    this.dimensionThresholdFilters.set(dtfs);
  }

  removeMeasureFilter(dimensionName: string, measureName: string) {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter === measureName) {
      this.temporaryFilterName.set(null);
      return;
    }
    const dtfs = get(this.dimensionThresholdFilters);
    const dimIdx = dtfs.findIndex((dtf) => dtf.name === dimensionName);
    if (dimIdx === -1) return;
    const filters = dtfs[dimIdx].filters;
    const exprIdx = filters.findIndex((f) => f.measure === measureName);
    if (exprIdx === -1) return;
    filters.splice(exprIdx, 1);
    if (!filters.length) {
      dtfs.splice(dimIdx, 1);
    }
    this.dimensionThresholdFilters.set(dtfs);
  }

  // ----- DIMENSION FILTER SELECTORS -----
  selectedDimensionValues(dimName: string): string[] {
    const wf = get(this.whereFilter);
    if (isExpressionUnsupported(wf)) return [];
    return [
      ...new Set(
        getValuesInExpression(
          this.getWhereFilterExpression(dimName),
        ) as string[],
      ),
    ];
  }

  atLeastOneSelection(dimName: string) {
    return this.selectedDimensionValues(dimName).length > 0;
  }

  isFilterExcludeMode(dimName: string) {
    const excludeMode = get(this.dimensionFilterExcludeMode);
    return excludeMode.get(dimName) ?? false;
  }

  dimensionHasFilter(dimName: string) {
    return this.getWhereFilterExpression(dimName) !== undefined;
  }

  getWhereFilterExpression(name: string) {
    const wf = get(this.whereFilter);
    return wf.cond?.exprs?.find((e) => matchExpressionByName(e, name));
  }

  getWhereFilterExpressionIndex(name: string) {
    const wf = get(this.whereFilter);
    return wf.cond?.exprs?.findIndex((e) => matchExpressionByName(e, name));
  }

  getDimensionFilterItems(
    dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
  ) {
    return this.getDimensionFilters(dimensionIdMap, get(this.whereFilter));
  }

  private getDimensionFilters(
    dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
    filter: V1Expression | undefined,
  ) {
    if (!filter) return [];
    const filteredDimensions: DimensionFilterItem[] = [];
    const addedDimension = new Set<string>();
    forEachIdentifier(filter, (e, ident) => {
      if (addedDimension.has(ident) || !dimensionIdMap.has(ident)) return;
      const dim = dimensionIdMap.get(ident);
      if (!dim) return;
      addedDimension.add(ident);
      filteredDimensions.push({
        name: ident,
        label: getDimensionDisplayName(dim),
        selectedValues: getValuesInExpression(e),
        isInclude: e.cond?.op === V1Operation.OPERATION_IN,
      });
    });
    return filteredDimensions.sort(filterItemsSortFunction);
  }

  getAllDimensionFilterItems(
    dimensionFilterItems: DimensionFilterItem[],
    dimensionIdMap: Map<string, MetricsViewSpecDimensionV2>,
  ) {
    const tempFilter = get(this.temporaryFilterName);
    const merged = [...dimensionFilterItems];
    if (tempFilter && dimensionIdMap.has(tempFilter)) {
      merged.push({
        name: tempFilter,
        label: getDimensionDisplayName(dimensionIdMap.get(tempFilter)),
        selectedValues: [],
        isInclude: true,
      });
    }
    return merged.sort(filterItemsSortFunction);
  }

  unselectedDimensionValues(
    dimensionName: string,
    values: unknown[],
  ): unknown[] {
    const expr = this.getWhereFilterExpression(dimensionName);
    if (!expr) return values;
    return values.filter(
      (v) => expr.cond?.exprs?.findIndex((e) => e.val === v) === -1,
    );
  }

  includedDimensionValues(dimensionName: string): unknown[] {
    const expr = this.getWhereFilterExpression(dimensionName);
    if (!expr || expr.cond?.op !== V1Operation.OPERATION_IN) {
      return [];
    }
    return getValuesInExpression(expr);
  }

  hasAtLeastOneDimensionFilter(): boolean {
    const wf = get(this.whereFilter);
    return !!(wf.cond?.exprs?.length && wf.cond.exprs.length > 0);
  }

  // ----- DIMENSION FILTER ACTIONS -----
  toggleDimensionValueSelection(
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean,
    isExclusiveFilter?: boolean,
  ) {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter !== null) {
      this.temporaryFilterName.set(null);
    }
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isInclude = !excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);
    const exprIdx = this.getWhereFilterExpressionIndex(dimensionName);
    if (exprIdx === undefined || exprIdx === -1) {
      wf.cond?.exprs?.push(
        createInExpression(dimensionName, [dimensionValue], !isInclude),
      );
      this.whereFilter.set(wf);
      return;
    }
    const expr = wf.cond?.exprs?.[exprIdx];
    if (!expr?.cond?.exprs) return;
    const inIdx = getValueIndexInExpression(expr, dimensionValue) as number;
    if (inIdx === -1) {
      if (isExclusiveFilter) {
        expr.cond.exprs.splice(1, expr.cond.exprs.length - 1, {
          val: dimensionValue,
        });
      } else {
        expr.cond.exprs.push({ val: dimensionValue });
      }
    } else {
      expr.cond.exprs.splice(inIdx, 1);
      if (expr.cond.exprs.length === 1) {
        wf.cond?.exprs?.splice(exprIdx, 1);
        if (keepPillVisible) {
          this.temporaryFilterName.set(dimensionName);
        }
      }
    }
    this.whereFilter.set(wf);
  }

  toggleDimensionFilterMode(dimensionName: string) {
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const newExclude = !excludeMode.get(dimensionName);
    excludeMode.set(dimensionName, newExclude);
    this.dimensionFilterExcludeMode.set(excludeMode);

    const wf = get(this.whereFilter);
    if (!wf.cond?.exprs) return;
    const exprIdx = wf.cond.exprs.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (exprIdx === -1) return;
    wf.cond.exprs[exprIdx] = negateExpression(wf.cond.exprs[exprIdx]);
    this.whereFilter.set(wf);
  }

  removeDimensionFilter(dimensionName: string) {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter === dimensionName) {
      this.temporaryFilterName.set(null);
      return;
    }
    const wf = get(this.whereFilter);
    const exprIdx = this.getWhereFilterExpressionIndex(dimensionName);
    if (exprIdx === undefined || exprIdx === -1) return;
    wf.cond?.exprs?.splice(exprIdx, 1);
    this.whereFilter.set(wf);
  }

  selectItemsInFilter(dimensionName: string, values: (string | null)[]) {
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isInclude = !excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);
    const exprIdx = this.getWhereFilterExpressionIndex(dimensionName);
    if (exprIdx === undefined || exprIdx === -1) {
      wf.cond?.exprs?.push(
        createInExpression(dimensionName, values, !isInclude),
      );
      this.whereFilter.set(wf);
      return;
    }
    const expr = wf.cond?.exprs?.[exprIdx];
    if (!expr?.cond?.exprs) return;
    const oldValues = getValuesInExpression(expr);
    const newValues = values.filter((v) => !oldValues.includes(v));
    expr.cond.exprs.push(...newValues.map((v) => ({ val: v })));
    this.whereFilter.set(wf);
  }

  deselectItemsInFilter(dimensionName: string, values: (string | null)[]) {
    const wf = get(this.whereFilter);
    const exprIdx = this.getWhereFilterExpressionIndex(dimensionName);
    if (exprIdx === undefined || exprIdx === -1) return;
    const expr = wf.cond?.exprs?.[exprIdx];
    if (!expr?.cond?.exprs) return;
    const oldValues = getValuesInExpression(expr);
    const newValues = oldValues.filter((v) => !values.includes(v));
    if (newValues.length) {
      expr.cond.exprs.splice(
        1,
        expr.cond.exprs.length - 1,
        ...newValues.map((v) => ({ val: v })),
      );
    } else {
      wf.cond?.exprs?.splice(exprIdx, 1);
    }
    this.whereFilter.set(wf);
  }

  setFilters(filter: V1Expression) {
    const { dimensionFilters, dimensionThresholdFilters } =
      splitWhereFilter(filter);
    this.whereFilter.set(dimensionFilters);
    this.dimensionThresholdFilters.set(dimensionThresholdFilters);
  }

  // ----- FILTER ACTIONS -----
  clearAllFilters() {
    const wf = get(this.whereFilter);
    const dtfs = get(this.dimensionThresholdFilters);
    const hasFilters = wf.cond?.exprs?.length || dtfs.length;
    if (!hasFilters) return;
    this.whereFilter.set(createAndExpression([]));
    this.dimensionThresholdFilters.set([]);
    this.temporaryFilterName.set(null);
    const excludeMode = get(this.dimensionFilterExcludeMode);
    excludeMode.clear();
    this.dimensionFilterExcludeMode.set(excludeMode);
    // ...existing code...
  }

  setTemporaryFilterName(name: string) {
    this.temporaryFilterName.set(name);
  }
}
