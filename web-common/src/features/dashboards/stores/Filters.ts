import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import { toggleDimensionFilterValue } from "@rilldata/web-common/features/dashboards/state-managers/actions/dimension-filters.ts";
import {
  type DimensionFilterItem,
  getDimensionFilters,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters.ts";
import { filterItemsSortFunction } from "@rilldata/web-common/features/dashboards/state-managers/selectors/filters.ts";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters.ts";
import type {
  DimensionThresholdFilter,
  ExploreState,
} from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import {
  copyFilterExpression,
  createAndExpression,
  createInExpression,
  createLikeExpression,
  matchExpressionByName,
  negateExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils.ts";
import {
  type MetricsViewSpecMeasure,
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import type { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";

export type FiltersState = Pick<
  ExploreState,
  | "whereFilter"
  | "dimensionsWithInlistFilter"
  | "dimensionThresholdFilters"
  | "dimensionFilterExcludeMode"
>;

/**
 * Filters class encapsulates all filter related selectors and actions into a single class.
 * It has individual stores for each data point.
 *
 * This is a copy of canvas filter class without canvas related stuff.
 * TODO: refactor canvas to use this
 */
export class Filters {
  // -------------------
  // STORES (writable)
  // -------------------
  public readonly whereFilter: Writable<V1Expression>;
  public readonly dimensionsWithInlistFilter: Writable<string[]>;
  public readonly dimensionThresholdFilters: Writable<
    Array<DimensionThresholdFilter>
  >;
  public readonly dimensionFilterExcludeMode: Writable<Map<string, boolean>>;
  public readonly temporaryFilterName: Writable<string | null>;

  // -------------------
  // "SELECTORS" (readable/derived)
  // -------------------
  public readonly measureFilterItems: Readable<MeasureFilterItem[]>;
  public readonly allMeasureFilterItems: Readable<MeasureFilterItem[]>;
  public readonly measureHasFilter: Readable<(measureName: string) => boolean>;

  public readonly dimensionFilterItems: Readable<DimensionFilterItem[]>;
  public readonly allDimensionFilterItems: Readable<DimensionFilterItem[]>;
  public readonly isFilterExcludeMode: Readable<(dimName: string) => boolean>;
  public readonly dimensionHasFilter: Readable<(dimName: string) => boolean>;

  public readonly hasFilters: Readable<boolean>;

  constructor(
    public readonly metricsViewMetadata: ExploreMetricsViewMetadata,
    {
      whereFilter,
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
      dimensionFilterExcludeMode,
    }: FiltersState,
  ) {
    // -----------------------------
    // Initialize writable stores
    // Lot of these are edited in place. So create a copy to avoid updating the original.
    // -----------------------------
    this.whereFilter = writable(copyFilterExpression(whereFilter));
    this.dimensionsWithInlistFilter = writable([...dimensionsWithInlistFilter]);
    this.dimensionThresholdFilters = writable(
      structuredClone(dimensionThresholdFilters),
    );
    this.dimensionFilterExcludeMode = writable(
      new Map(dimensionFilterExcludeMode),
    );
    this.temporaryFilterName = writable(null);

    // -------------------------------
    // MEASURE SELECTORS
    // -------------------------------
    this.measureFilterItems = derived(
      [this.metricsViewMetadata.measureNameMap, this.dimensionThresholdFilters],
      ([$measureNameMap, $dimensionThresholdFilters]) => {
        return this.getMeasureFilters(
          $measureNameMap,
          $dimensionThresholdFilters,
        );
      },
    );

    this.allMeasureFilterItems = derived(
      [
        this.metricsViewMetadata.measureNameMap,
        this.measureFilterItems,
        this.temporaryFilterName,
      ],
      ([$measureNameMap, $measureFilterItems, tempFilter]) => {
        const itemsCopy = [...$measureFilterItems];
        if (tempFilter && $measureNameMap.has(tempFilter)) {
          itemsCopy.push({
            dimensionName: "",
            name: tempFilter,
            label: getMeasureDisplayName($measureNameMap.get(tempFilter)),
            // dimensions, // TODO: for canvas
          });
        }
        return itemsCopy;
      },
    );

    this.measureHasFilter = derived(
      this.dimensionThresholdFilters,
      ($dimensionThresholdFilters) => {
        return (measureName: string) => {
          return $dimensionThresholdFilters.some((dtf) =>
            dtf.filters.some((f) => f.measure === measureName),
          );
        };
      },
    );

    // -------------------------------
    // DIMENSION SELECTORS
    // -------------------------------
    this.dimensionFilterItems = derived(
      [
        this.metricsViewMetadata.dimensionNameMap,
        this.whereFilter,
        this.dimensionsWithInlistFilter,
      ],
      ([$dimensionNameMap, $whereFilter, $dimensionsWithInlistFilter]) => {
        // TODO: fill in metricsViewNames for each dimension filter item when integrated into canvas
        return getDimensionFilters(
          $dimensionNameMap,
          $whereFilter,
          $dimensionsWithInlistFilter,
        );
      },
    );

    this.allDimensionFilterItems = derived(
      [
        this.metricsViewMetadata.dimensionNameMap,
        this.dimensionFilterItems,
        this.temporaryFilterName,
      ],
      ([$dimensionNameMap, $dimensionFilterItems, tempFilter]) => {
        const merged = $dimensionFilterItems.map((dfi) => ({
          ...dfi,
          metricsViewNames: [this.metricsViewMetadata.metricsViewName],
        }));
        if (tempFilter && $dimensionNameMap.has(tempFilter)) {
          merged.push({
            mode: DimensionFilterMode.Select,
            name: tempFilter,
            label: getDimensionDisplayName($dimensionNameMap.get(tempFilter)),
            selectedValues: [],
            isInclude: true,
            metricsViewNames: [this.metricsViewMetadata.metricsViewName],
          });
        }
        return merged.sort(filterItemsSortFunction);
      },
    );

    this.isFilterExcludeMode = derived(
      this.dimensionFilterExcludeMode,
      ($excludeMode) => {
        return (dimName: string) => {
          return $excludeMode.get(dimName) ?? false;
        };
      },
    );

    this.dimensionHasFilter = derived(this.whereFilter, ($whereFilter) => {
      return (dimName: string) => {
        return (
          $whereFilter.cond?.exprs?.find((e) =>
            matchExpressionByName(e, dimName),
          ) !== undefined
        );
      };
    });

    this.hasFilters = derived(
      [this.dimensionFilterItems, this.measureFilterItems],
      ([$dimensionFilterItems, $measureFilterItems]) =>
        $dimensionFilterItems.length > 0 || $measureFilterItems.length > 0,
    );
  }

  // --------------------
  // ACTIONS / MUTATORS
  // --------------------

  public setMeasureFilter = (
    dimensionName: string,
    filter: MeasureFilterEntry,
  ) => {
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
  };

  public removeMeasureFilter = (dimensionName: string, measureName: string) => {
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
  };

  toggleDimensionValueSelection = (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean,
    isExclusiveFilter?: boolean,
  ) => {
    this.toggleMultipleDimensionValueSelections(
      dimensionName,
      [dimensionValue],
      keepPillVisible,
      isExclusiveFilter,
    );
  };

  toggleMultipleDimensionValueSelections = (
    dimensionName: string,
    dimensionValues: string[],
    keepPillVisible?: boolean,
    isExclusiveFilter?: boolean,
  ) => {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter !== null) {
      this.temporaryFilterName.set(null);
    }

    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);

    // Use the derived selector:
    let exprIndex = this.getWhereFilterExpressionIndex(dimensionName) ?? -1;
    let expr = wf.cond?.exprs?.[exprIndex];

    const wasLikeFilter =
      expr?.cond?.op === V1Operation.OPERATION_LIKE ||
      expr?.cond?.op === V1Operation.OPERATION_NLIKE;
    if (!expr?.cond?.exprs || wasLikeFilter) {
      expr = createInExpression(dimensionName, [], isExclude);
      wf.cond?.exprs?.push(expr);
      exprIndex = wf.cond!.exprs!.length - 1;
    }

    const wasInListFilter = get(this.dimensionsWithInlistFilter).includes(
      dimensionName,
    );
    if (wasInListFilter) {
      this.dimensionsWithInlistFilter.update((dimensionsWithInlistFilter) =>
        dimensionsWithInlistFilter.filter((d) => d !== dimensionName),
      );
    }

    dimensionValues.forEach((dimensionValue) => {
      toggleDimensionFilterValue(expr, dimensionValue, !!isExclusiveFilter);
    });

    if (expr?.cond?.exprs?.length === 1) {
      wf.cond?.exprs?.splice(exprIndex, 1);

      if (keepPillVisible) {
        this.setTemporaryFilterName(dimensionName);
      }
    }

    this.whereFilter.set(wf);
  };

  public applyDimensionInListMode = (
    dimensionName: string,
    values: string[],
  ) => {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter !== null) {
      this.temporaryFilterName.set(null);
    }
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);

    const expr = createInExpression(dimensionName, values, isExclude);
    this.dimensionsWithInlistFilter.update((dimensionsWithInlistFilter) => {
      return [...dimensionsWithInlistFilter, dimensionName];
    });

    const exprIndex = this.getWhereFilterExpressionIndex(dimensionName);
    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(expr);
    } else {
      wf.cond!.exprs![exprIndex] = expr;
    }
    this.whereFilter.set(wf);
  };

  public applyDimensionContainsMode = (
    dimensionName: string,
    searchText: string,
  ) => {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter !== null) {
      this.temporaryFilterName.set(null);
    }
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);

    const expr = createLikeExpression(
      dimensionName,
      `%${searchText}%`,
      isExclude,
    );
    const exprIndex = this.getWhereFilterExpressionIndex(dimensionName);
    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(expr);
    } else {
      wf.cond!.exprs![exprIndex] = expr;
    }
    this.whereFilter.set(wf);
  };

  public toggleDimensionFilterMode = (dimensionName: string) => {
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
  };

  public removeDimensionFilter = (dimensionName: string) => {
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
  };

  public setTemporaryFilterName = (name: string) => {
    this.temporaryFilterName.set(name);
  };

  public toState(): FiltersState {
    return {
      whereFilter: get(this.whereFilter),
      dimensionThresholdFilters: get(this.dimensionThresholdFilters),
      dimensionsWithInlistFilter: get(this.dimensionsWithInlistFilter),
      dimensionFilterExcludeMode: get(this.dimensionFilterExcludeMode),
    };
  }

  public getStore(): Readable<FiltersState> {
    return derived(
      [
        this.whereFilter,
        this.dimensionThresholdFilters,
        this.dimensionsWithInlistFilter,
        this.dimensionFilterExcludeMode,
      ],
      ([
        whereFilter,
        dimensionThresholdFilters,
        dimensionsWithInlistFilter,
        dimensionFilterExcludeMode,
      ]) => ({
        whereFilter,
        dimensionThresholdFilters,
        dimensionsWithInlistFilter,
        dimensionFilterExcludeMode,
      }),
    );
  }

  public clearAllFilters = () => {
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
  };

  private getMeasureFilters(
    measureNameMap: Map<string, MetricsViewSpecMeasure>,
    dimensionThresholdFilters: DimensionThresholdFilter[],
  ): MeasureFilterItem[] {
    return dedupe(
      dimensionThresholdFilters
        .map((dtf) =>
          this.getMeasureFilterForDimension(
            measureNameMap,
            dtf.filters,
            dtf.name,
          ),
        )
        .flat(),
      (i) => i.name,
    );
  }

  private getMeasureFilterForDimension(
    measureNameMap: Map<string, MetricsViewSpecMeasure>,
    filters: MeasureFilterEntry[],
    name: string,
  ): MeasureFilterItem[] {
    return filters
      .map((filter) => {
        const measure = measureNameMap.get(filter.measure);
        if (!measure) return undefined;
        return <MeasureFilterItem>{
          dimensionName: name,
          name: filter.measure,
          label: measure.displayName || measure.expression || filter.measure,
          filter,
          // dimensions, // TODO: for canvas
        };
      })
      .filter(Boolean) as MeasureFilterItem[];
  }

  private getWhereFilterExpressionIndex(name: string) {
    const $whereFilter = get(this.whereFilter);
    return $whereFilter.cond?.exprs?.findIndex((e) =>
      matchExpressionByName(e, name),
    );
  }
}
