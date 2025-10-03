import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  mergeDimensionAndMeasureFilters,
  splitWhereFilter,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { toggleDimensionFilterValue } from "@rilldata/web-common/features/dashboards/state-managers/actions/dimension-filters.ts";
import {
  type DimensionFilterItem,
  getDimensionFiltersMap,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
  getValuesInExpression,
  isExpressionUnsupported,
  matchExpressionByName,
  negateExpression,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import type { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type {
  MetricsViewSpecDimension,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  type MetricsViewSpecMeasure,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import {
  derived,
  get,
  type Readable,
  writable,
  type Writable,
} from "svelte/store";
import { ExploreStateURLParams } from "../../dashboards/url-state/url-params";
import type { SearchParamsStore } from "./canvas-entity";

export class Filters {
  private metricsView: MetricsViewSelectors;
  // -------------------
  // STORES (writable)
  // -------------------
  whereFilter: Writable<V1Expression>;
  dimensionsWithInlistFilter: Writable<string[]>;
  dimensionThresholdFilters: Writable<Array<DimensionThresholdFilter>>;
  dimensionFilterExcludeMode: Writable<Map<string, boolean>>;

  // -------------------
  // "SELECTORS" (readable/derived)
  // -------------------
  measureHasFilter: Readable<(measureName: string) => boolean>;
  allMeasureFilterItems: Readable<MeasureFilterItem[]>;
  measureFilterItems: Readable<MeasureFilterItem[]>;

  allDimensionFilterItems: Readable<Map<string, DimensionFilterItem>>;
  selectedDimensionValues: Readable<(dimName: string) => string[]>;
  atLeastOneSelection: Readable<(dimName: string) => boolean>;
  isFilterExcludeMode: Readable<(dimName: string) => boolean>;
  dimensionHasFilter: Readable<(dimName: string) => boolean>;
  getWhereFilterExpression: Readable<
    (name: string) => V1Expression | undefined
  >;
  getWhereFilterExpressionIndex: Readable<(name: string) => number | undefined>;
  dimensionFilterItems: Readable<Map<string, DimensionFilterItem>>;
  unselectedDimensionValues: Readable<
    (dimensionName: string, values: unknown[]) => unknown[]
  >;
  includedDimensionValues: Readable<(dimensionName: string) => unknown[]>;
  hasAtLeastOneDimensionFilter: Readable<() => boolean>;
  filterText: Readable<string>;
  allDimensions: Readable<Map<string, MetricsViewSpecDimension>>;
  allMeasures: Readable<Map<string, MetricsViewSpecMeasure>>;
  temporaryFilters = writable<Set<string>>(new Set());

  constructor(
    metricsView: MetricsViewSelectors,
    public searchParamsStore: SearchParamsStore,
    public componentName?: string,
  ) {
    // -----------------------------
    // Initialize writable stores
    // -----------------------------
    this.metricsView = metricsView;
    this.dimensionFilterExcludeMode = writable(new Map<string, boolean>());

    this.whereFilter = writable({
      cond: {
        op: "OPERATION_AND",
        exprs: [],
      },
    });
    this.dimensionsWithInlistFilter = writable([]);
    this.dimensionThresholdFilters = writable([]);

    // -------------------------------
    // MEASURE SELECTORS
    // -------------------------------

    this.allMeasures = derived(
      [this.metricsView.allSimpleMeasures],
      ([$allSimpleMeasures]) =>
        getMapFromArray($allSimpleMeasures, (m) => m.name as string),
    );

    this.measureFilterItems = derived(
      [this.dimensionThresholdFilters, this.allMeasures],
      ([$dimensionThresholdFilters, measureIdMap]) => {
        return this.getMeasureFilters(measureIdMap, $dimensionThresholdFilters);
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

    this.allMeasureFilterItems = derived(
      [this.temporaryFilters, this.measureFilterItems, this.allMeasures],
      ([tempFilters, measureFilterItems, measureIdMap]) => {
        const itemsCopy = [...measureFilterItems];

        tempFilters.forEach((tempFilter) => {
          if (
            tempFilter &&
            measureIdMap.has(tempFilter) &&
            !itemsCopy.some((i) => i.name === tempFilter)
          ) {
            const dimensions = metricsView.getDimensionsFromMeasure(tempFilter);
            itemsCopy.push({
              dimensionName: "",
              name: tempFilter,
              label: getMeasureDisplayName(measureIdMap.get(tempFilter)),
              dimensions: dimensions,
            });
          }
        });

        return itemsCopy;
      },
    );

    // -------------------------------
    // DIMENSION SELECTORS
    // -------------------------------
    this.allDimensions = derived(
      [this.metricsView.allDimensions],
      ([$allDimensions]) =>
        getMapFromArray(
          $allDimensions,
          (dimension) => (dimension.name || dimension.column) as string,
        ),
    );

    this.dimensionFilterItems = derived(
      [this.whereFilter, this.dimensionsWithInlistFilter, this.allDimensions],
      ([$whereFilter, $dimensionsWithInlistFilter, $allDimensions]) => {
        const dimensionFilters = getDimensionFiltersMap(
          $allDimensions,
          $whereFilter,
          $dimensionsWithInlistFilter,
        );
        dimensionFilters.forEach((dimensionFilter) => {
          dimensionFilter.metricsViewNames =
            metricsView.getMetricsViewNamesForDimension(dimensionFilter.name);
        });
        return dimensionFilters;
      },
    );

    this.allDimensionFilterItems = derived(
      [
        this.temporaryFilters,
        this.dimensionFilterItems,
        this.allDimensions,
        this.dimensionFilterExcludeMode,
      ],
      ([tempFilters, dimensionFilters, $allDimensions, $excludeMode]) => {
        const merged = structuredClone(dimensionFilters);

        tempFilters.forEach((tempFilter) => {
          const hasFilter = merged.has(tempFilter);

          if (tempFilter && $allDimensions.has(tempFilter) && !hasFilter) {
            const metricsViewNames =
              metricsView.getMetricsViewNamesForDimension(tempFilter);
            merged.set(tempFilter, {
              mode: DimensionFilterMode.Select,
              name: tempFilter,
              label: getDimensionDisplayName($allDimensions.get(tempFilter)),
              selectedValues: [],
              isInclude: !$excludeMode.get(tempFilter),
              metricsViewNames,
            });
          }
        });

        return merged;
      },
    );

    this.selectedDimensionValues = derived(this.whereFilter, ($whereFilter) => {
      return (dimName: string) => {
        if (isExpressionUnsupported($whereFilter)) return [];
        // find the filter expression for this dimension
        const expr = $whereFilter.cond?.exprs?.find((e) =>
          matchExpressionByName(e, dimName),
        );
        return [...new Set(getValuesInExpression(expr) as string[])];
      };
    });

    this.atLeastOneSelection = derived(
      this.selectedDimensionValues,
      (fnSelectedDimensionValues) => {
        return (dimName: string) => {
          return fnSelectedDimensionValues(dimName).length > 0;
        };
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

    this.getWhereFilterExpression = derived(
      this.whereFilter,
      ($whereFilter) => {
        return (name: string) => {
          return $whereFilter.cond?.exprs?.find((e) =>
            matchExpressionByName(e, name),
          );
        };
      },
    );

    this.getWhereFilterExpressionIndex = derived(
      this.whereFilter,
      ($whereFilter) => {
        return (name: string) => {
          return $whereFilter.cond?.exprs?.findIndex((e) =>
            matchExpressionByName(e, name),
          );
        };
      },
    );

    this.unselectedDimensionValues = derived(
      this.whereFilter,
      ($whereFilter) => {
        return (dimensionName: string, values: unknown[]) => {
          const expr = $whereFilter.cond?.exprs?.find((e) =>
            matchExpressionByName(e, dimensionName),
          );
          if (!expr) return values;
          return values.filter(
            (v) => expr.cond?.exprs?.findIndex((e) => e.val === v) === -1,
          );
        };
      },
    );

    this.includedDimensionValues = derived(this.whereFilter, ($whereFilter) => {
      return (dimensionName: string) => {
        const expr = $whereFilter.cond?.exprs?.find((e) =>
          matchExpressionByName(e, dimensionName),
        );
        if (!expr || expr.cond?.op !== V1Operation.OPERATION_IN) {
          return [];
        }
        return getValuesInExpression(expr);
      };
    });

    this.hasAtLeastOneDimensionFilter = derived(
      this.whereFilter,
      ($whereFilter) => {
        return () => {
          return !!(
            $whereFilter.cond?.exprs?.length &&
            $whereFilter.cond.exprs.length > 0
          );
        };
      },
    );

    this.searchParamsStore.subscribe((searchParams) => {
      const filterText = searchParams.get(ExploreStateURLParams.Filters);
      if (!this.componentName) {
        const tempFilters = searchParams.get(
          ExploreStateURLParams.TemporaryFilters,
        );
        this.temporaryFilters.set(new Set(tempFilters?.split(",") ?? []));
      }

      this.setFiltersFromText(filterText ?? "");
    });
  }

  private getMeasureFilters = (
    measureIdMap: Map<string, MetricsViewSpecMeasure>,
    dimensionThresholdFilters: DimensionThresholdFilter[],
  ): MeasureFilterItem[] => {
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
  };

  private getMeasureFilterForDimension = (
    measureIdMap: Map<string, MetricsViewSpecMeasure>,
    filters: MeasureFilterEntry[],
    name: string,
    addedMeasure: Set<string>,
  ): MeasureFilterItem[] => {
    if (!filters.length) return [];
    const filteredMeasures: MeasureFilterItem[] = [];
    filters.forEach((filter) => {
      if (addedMeasure.has(filter.measure)) return;
      const measure = measureIdMap.get(filter.measure);
      if (!measure) return;
      const dimensions = this.metricsView.getDimensionsFromMeasure(
        filter.measure,
      );
      addedMeasure.add(filter.measure);
      filteredMeasures.push({
        dimensionName: name,
        name: filter.measure,
        label: measure.displayName || measure.expression || filter.measure,
        filter,
        dimensions,
      });
    });
    return filteredMeasures;
  };

  // --------------------
  // ACTIONS / MUTATORS
  // --------------------

  setMeasureFilter = (dimensionName: string, filter: MeasureFilterEntry) => {
    this.checkTemporaryFilter(filter.measure);

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

    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        get(this.whereFilter),
        dtfs,
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  removeMeasureFilter = (dimensionName: string, measureName: string) => {
    this.checkTemporaryFilter(measureName);
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
    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        get(this.whereFilter),
        dtfs,
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  checkTemporaryFilter = (filterName: string) => {
    const tempFilters = get(this.temporaryFilters);
    if (tempFilters.has(filterName)) {
      tempFilters.delete(filterName);

      if (this.componentName) {
        this.temporaryFilters.set(tempFilters);
      } else {
        this.searchParamsStore.set(
          ExploreStateURLParams.TemporaryFilters,
          Array.from(tempFilters).join(","),
        );
      }
    }
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
    this.checkTemporaryFilter(dimensionName);

    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);

    const wf = get(this.whereFilter);

    // Use the derived selector:
    let exprIndex =
      get(this.getWhereFilterExpressionIndex)(dimensionName) ?? -1;
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

    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  applyDimensionInListMode = (dimensionName: string, values: string[]) => {
    this.checkTemporaryFilter(dimensionName);
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);

    const expr = createInExpression(dimensionName, values, isExclude);
    this.dimensionsWithInlistFilter.update((dimensionsWithInlistFilter) => {
      return [...dimensionsWithInlistFilter, dimensionName];
    });

    const exprIndex = get(this.getWhereFilterExpressionIndex)(dimensionName);
    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(expr);
    } else {
      wf.cond!.exprs![exprIndex] = expr;
    }
    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  applyDimensionContainsMode = (dimensionName: string, searchText: string) => {
    this.checkTemporaryFilter(dimensionName);
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);

    const expr = createLikeExpression(
      dimensionName,
      `%${searchText}%`,
      isExclude,
    );
    const exprIndex = get(this.getWhereFilterExpressionIndex)(dimensionName);
    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(expr);
    } else {
      wf.cond!.exprs![exprIndex] = expr;
    }
    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  toggleDimensionFilterMode = (dimensionName: string) => {
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

    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  removeDimensionFilter = (dimensionName: string) => {
    this.checkTemporaryFilter(dimensionName);
    const wf = get(this.whereFilter);
    const exprIdx = get(this.getWhereFilterExpressionIndex)(dimensionName);
    if (exprIdx === undefined || exprIdx === -1) return;
    wf.cond?.exprs?.splice(exprIdx, 1);

    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  selectItemsInFilter = (dimensionName: string, values: (string | null)[]) => {
    const excludeMode = get(this.dimensionFilterExcludeMode);
    const isExclude = !!excludeMode.get(dimensionName);
    const wf = get(this.whereFilter);
    const exprIdx = get(this.getWhereFilterExpressionIndex)(dimensionName);
    if (exprIdx === undefined || exprIdx === -1) {
      wf.cond?.exprs?.push(
        createInExpression(dimensionName, values, isExclude),
      );
      this.searchParamsStore.set(
        ExploreStateURLParams.Filters,
        getFilterParam(
          wf,
          get(this.dimensionThresholdFilters),
          get(this.dimensionsWithInlistFilter),
        ),
      );
      return;
    }
    const expr = wf.cond?.exprs?.[exprIdx];
    if (!expr?.cond?.exprs) return;
    const oldValues = getValuesInExpression(expr);
    const newValues = values.filter((v) => !oldValues.includes(v));
    expr.cond.exprs.push(...newValues.map((v) => ({ val: v })));
    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  deselectItemsInFilter = (
    dimensionName: string,
    values: (string | null)[],
  ) => {
    const wf = get(this.whereFilter);
    const exprIdx = get(this.getWhereFilterExpressionIndex)(dimensionName);
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
    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      getFilterParam(
        wf,
        get(this.dimensionThresholdFilters),
        get(this.dimensionsWithInlistFilter),
      ),
    );
  };

  setFilters = (filter: V1Expression) => {
    const { dimensionFilters, dimensionThresholdFilters } =
      splitWhereFilter(filter);

    this.whereFilter.set(dimensionFilters);
    this.dimensionThresholdFilters.set(dimensionThresholdFilters);
  };

  clearAllFilters = () => {
    this.temporaryFilters.set(new Set());
    this.searchParamsStore.set(ExploreStateURLParams.Filters, undefined);
  };

  setTemporaryFilterName = (name: string) => {
    if (this.componentName) {
      this.temporaryFilters.update((tempFilters) => {
        if (tempFilters.has(name)) {
          tempFilters.delete(name);
        }
        return tempFilters.add(name);
      });
    } else {
      this.searchParamsStore.set(ExploreStateURLParams.TemporaryFilters, name);
    }
  };

  setFiltersFromText = (filterText: string) => {
    const { expr, dimensionsWithInlistFilter } = getFiltersFromText(filterText);

    this.setFilters(expr);
    this.dimensionsWithInlistFilter.set(dimensionsWithInlistFilter);
  };
}

function getFilterParam(
  whereFilter: V1Expression,
  dtf: DimensionThresholdFilter[],
  dimensionsWithInlistFilter: string[],
) {
  const mergedFilters =
    sanitiseExpression(
      mergeDimensionAndMeasureFilters(
        whereFilter ?? createAndExpression([]),
        dtf,
      ),
      undefined,
    ) ?? createAndExpression([]);

  return convertExpressionToFilterParam(
    mergedFilters,
    dimensionsWithInlistFilter,
  );
}
