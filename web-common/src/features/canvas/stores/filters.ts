import type { CanvasResolvedSpec } from "@rilldata/web-common/features/canvas/stores/spec";
import {
  DimensionFilterMode,
  dimensionFilterModeMap,
} from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
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
import type {
  MetricsViewSpecDimension,
  V1CanvasPreset,
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
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type { CanvasSpecResponseStore } from "../types";
import {
  OperationShortHandMap,
  MeasureFilterType,
} from "../../dashboards/filters/measure-filters/measure-filter-options";
import type { CanvasResponse } from "../selector";

type FilterProperties = {
  hidden?: boolean;
  locked?: boolean;
  unremovable?: boolean;
  limit?: number;
};

export class Filters {
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
  allDimensions: Readable<Map<string, MetricsViewSpecDimension>>;
  allMeasures: Readable<Map<string, MetricsViewSpecMeasure>>;
  temporaryFilters = writable<Set<string>>(new Set());
  defaultFilterProperties = writable<Map<string, FilterProperties>>(new Map());
  firstPass = true;
  searchParamsStore = writable<URLSearchParams>(new URLSearchParams());
  private filterText = writable("");

  constructor(
    private spec: CanvasResolvedSpec,
    public searchParamsCallback: (
      key: string,
      value: string | undefined,
      checkIfSet?: boolean,
    ) => boolean,
    private specStore?: CanvasSpecResponseStore,
    public componentName?: string,
  ) {
    // -----------------------------
    // Initialize writable stores
    // -----------------------------
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
      [this.spec.allSimpleMeasures],
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
            const dimensions = spec.getDimensionsFromMeasure(tempFilter);
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
      [this.spec.allDimensions],
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
            spec.getMetricsViewNamesForDimension(dimensionFilter.name);
        });
        return dimensionFilters;
      },
    );

    this.allDimensionFilterItems = derived(
      [this.temporaryFilters, this.dimensionFilterItems, this.allDimensions],
      ([tempFilters, dimensionFilters, $allDimensions]) => {
        const merged = structuredClone(dimensionFilters);

        tempFilters.forEach((tempFilter) => {
          const hasFilter = merged.has(tempFilter);

          if (tempFilter && $allDimensions.has(tempFilter) && !hasFilter) {
            const metricsViewNames =
              spec.getMetricsViewNamesForDimension(tempFilter);
            merged.set(tempFilter, {
              mode: DimensionFilterMode.Select,
              name: tempFilter,
              label: getDimensionDisplayName($allDimensions.get(tempFilter)),
              selectedValues: [],
              isInclude: true,
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
  }

  onSpecChange(response: CanvasResponse) {
    const defaultPreset = response.canvas?.defaultPreset;

    if (!defaultPreset) return;

    const {
      dimensionMap,
      measureMap,
      defaultFilterProperties,
      temporaryFilters,
    } = this.processDefaults(defaultPreset);

    this.setDefaults(
      dimensionMap,
      measureMap,
      defaultFilterProperties,
      temporaryFilters,
    );

    this.firstPass = false;
  }

  onUrlChange = (searchParams: URLSearchParams) => {
    const filterText = searchParams.get(ExploreStateURLParams.Filters);

    if (!this.componentName) {
      const tempFilters = searchParams.get(
        ExploreStateURLParams.TemporaryFilters,
      );
      this.temporaryFilters.set(new Set(tempFilters?.split(",") ?? []));
    }

    this.filterText.set(filterText ?? "");

    this.setFiltersFromText(filterText ?? "");
  };

  setDefaults = (
    dimensionMap: Map<string, DimensionFilterItem>,
    measureMap: Map<string, MeasureFilterItem>,
    defaultFilterProperties: Map<string, FilterProperties>,
    temporaryFilters: Set<string>,
    maintainUrlState = true,
  ) => {
    const finalDimensions = new Map<string, DimensionFilterItem>();
    const finalMeasures = new Map<string, MeasureFilterItem>();
    const filterText = get(this.filterText);

    const { expr, dimensionsWithInlistFilter } = getFiltersFromText(filterText);

    const allDimensions = get(this.allDimensions);
    const allMeasures = get(this.allMeasures);

    const urlMeasureFilters = this.getMeasureFilters(
      allMeasures,
      splitWhereFilter(expr).dimensionThresholdFilters,
    );

    const urlDimensionFilters = getDimensionFiltersMap(
      allDimensions,
      splitWhereFilter(expr).dimensionFilters,
      dimensionsWithInlistFilter,
    );

    urlDimensionFilters.forEach((dimensionFilter) => {
      dimensionFilter.metricsViewNames =
        this.spec.getMetricsViewNamesForDimension(dimensionFilter.name);
    });

    dimensionMap.forEach((defaultDimensionItem, dimension) => {
      const urlFilter = urlDimensionFilters.get(dimension);

      if (
        !urlFilter ||
        (urlFilter &&
          (defaultFilterProperties.get(dimension)?.locked || !this.firstPass))
      ) {
        finalDimensions.set(dimension, defaultDimensionItem);
      }
    });

    measureMap.forEach((defaultMeasureItem, measure) => {
      const urlFilter = urlMeasureFilters.find((mf) => mf.name === measure);
      if (
        !urlFilter ||
        (urlFilter &&
          (defaultFilterProperties.get(measure)?.locked || !this.firstPass))
      ) {
        finalMeasures.set(measure, defaultMeasureItem);
      }
    });

    if (maintainUrlState) {
      urlDimensionFilters.forEach((urlDimensionItem, dimension) => {
        if (!finalDimensions.has(dimension)) {
          finalDimensions.set(dimension, urlDimensionItem);
        }
      });

      urlMeasureFilters.forEach((urlMeasureItem) => {
        if (!finalMeasures.has(urlMeasureItem.name)) {
          finalMeasures.set(urlMeasureItem.name, urlMeasureItem);
        }
      });
    }

    const { whereFilter, dimensionsWithInListFilter } =
      constructWhereFilterFromMeasureAndDimensionFilters(
        Array.from(finalDimensions.values()),
        Array.from(finalMeasures.values()),
      );

    const { dimensionFilters, dimensionThresholdFilters } =
      splitWhereFilter(whereFilter);

    this.searchParamsCallback(
      ExploreStateURLParams.Filters,
      getFilterParam(
        dimensionFilters,
        dimensionThresholdFilters,
        dimensionsWithInListFilter,
      ),
    );

    this.searchParamsCallback(
      ExploreStateURLParams.TemporaryFilters,
      temporaryFilters.size
        ? Array.from(temporaryFilters).join(",")
        : undefined,
    );
  };

  private processDefaults(defaultPreset: V1CanvasPreset | undefined) {
    const allDimensions = get(this.allDimensions);
    const allMeasures = get(this.allMeasures);

    const defaultFilterProperties = new Map<string, FilterProperties>();
    const dimensionMap: Map<string, DimensionFilterItem> = new Map();
    const measureMap: Map<string, MeasureFilterItem> = new Map();
    const temporaryFilters = new Set<string>();

    if (!defaultPreset)
      return {
        dimensionMap,
        measureMap,
        defaultFilterProperties,
        temporaryFilters,
      };

    defaultPreset?.filters?.dimensions?.forEach(
      ({
        dimension,
        hidden,
        removable,
        locked,
        values,
        limit,
        mode,
        exclude,
      }) => {
        if (!dimension || !allDimensions.has(dimension)) return;

        const properties: FilterProperties = {
          hidden,
          locked,
          unremovable: removable === false,
          limit,
        };

        defaultFilterProperties.set(dimension, properties);

        if (!values?.length) {
          temporaryFilters.add(dimension);
        } else {
          dimensionMap.set(dimension, {
            mode: dimensionFilterModeMap[mode ?? "select"],
            name: dimension,
            label: dimension,
            inputText: mode === "contains" ? values[0] : undefined,
            selectedValues: values?.length ? values : [],
            isInclude: !exclude,
          });
        }
      },
    );

    defaultPreset?.filters?.measures?.forEach(
      ({
        measure,
        hidden,
        locked,
        removable,
        byDimension,
        operator,
        values,
      }) => {
        if (!measure || !allMeasures.has(measure)) return;

        const properties = {
          hidden,
          locked,
          unremovable: removable === false,
        };

        defaultFilterProperties.set(measure, properties);

        const operation = OperationShortHandMap.get(operator ?? "");

        if (!byDimension || !operation) {
          temporaryFilters.add(measure);
          return;
        }

        measureMap.set(measure, {
          dimensionName: byDimension,
          name: measure,
          label: measure,
          filter: {
            measure,
            operation,
            type: MeasureFilterType.Value,
            value1: values?.[0]?.toString() ?? "",
            value2: values?.[1]?.toString() ?? "",
          },
        });
      },
    );

    this.defaultFilterProperties.set(defaultFilterProperties);

    return {
      dimensionMap,
      measureMap,
      defaultFilterProperties,
      temporaryFilters,
    };
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
      const dimensions = this.spec.getDimensionsFromMeasure(filter.measure);
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

    this.searchParamsCallback(
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
    this.searchParamsCallback(
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
        this.searchParamsCallback(
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
    skipToggling?: boolean,
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
      toggleDimensionFilterValue(
        expr,
        dimensionValue,
        !!isExclusiveFilter,
        skipToggling,
      );
    });

    if (expr?.cond?.exprs?.length === 1) {
      wf.cond?.exprs?.splice(exprIndex, 1);

      if (keepPillVisible) {
        this.setTemporaryFilterName(dimensionName);
      }
    }

    this.searchParamsCallback(
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
    this.searchParamsCallback(
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
    this.searchParamsCallback(
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

    this.searchParamsCallback(
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

    this.searchParamsCallback(
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
      this.searchParamsCallback(
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
    this.searchParamsCallback(
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
    this.searchParamsCallback(
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

    if (!this.specStore) return;
    const defaultPreset = get(this.specStore).data?.canvas?.defaultPreset;

    const {
      dimensionMap,
      measureMap,
      defaultFilterProperties,
      temporaryFilters,
    } = this.processDefaults(defaultPreset);

    this.setDefaults(
      dimensionMap,
      measureMap,
      defaultFilterProperties,
      temporaryFilters,
      false,
    );
  };

  setTemporaryFilterName = (name: string, overwrite = false) => {
    const tempFilters = get(this.temporaryFilters);

    if (this.componentName) {
      this.temporaryFilters.update((tempFilters) => {
        if (tempFilters.has(name)) {
          tempFilters.delete(name);
        }
        return tempFilters.add(name);
      });
    } else {
      if (overwrite) {
        const defaultFilterProperties = get(this.defaultFilterProperties);

        const filtered = Array.from(tempFilters).filter((t) => {
          return defaultFilterProperties.get(t)?.unremovable;
        });

        filtered.push(name);

        this.searchParamsCallback(
          ExploreStateURLParams.TemporaryFilters,
          filtered.join(","),
        );
      } else {
        tempFilters.add(name);

        this.searchParamsCallback(
          ExploreStateURLParams.TemporaryFilters,
          Array.from(tempFilters).join(","),
        );
      }
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

function constructWhereFilterFromMeasureAndDimensionFilters(
  dimensionFilters: DimensionFilterItem[],
  measureFilters: MeasureFilterItem[],
): { whereFilter: V1Expression; dimensionsWithInListFilter: string[] } {
  const dimensionExprs: V1Expression[] = [];
  const dimensionsWithInListFilter: string[] = [];

  for (const dimensionFilter of dimensionFilters) {
    if (
      dimensionFilter.mode === DimensionFilterMode.Select &&
      dimensionFilter.selectedValues.length > 0
    ) {
      dimensionExprs.push(
        createInExpression(
          dimensionFilter.name,
          dimensionFilter.selectedValues,
          !dimensionFilter.isInclude,
        ),
      );
    } else if (
      dimensionFilter.mode === DimensionFilterMode.Contains &&
      (dimensionFilter.selectedValues.length > 0 || dimensionFilter.inputText)
    ) {
      dimensionExprs.push(
        createLikeExpression(
          dimensionFilter.name,
          dimensionFilter.inputText
            ? `${dimensionFilter.inputText}`
            : `${dimensionFilter.selectedValues[0]}`,
          !dimensionFilter.isInclude,
        ),
      );
    } else if (
      dimensionFilter.mode === DimensionFilterMode.InList &&
      dimensionFilter.selectedValues.length > 0
    ) {
      dimensionsWithInListFilter.push(dimensionFilter.name);
      dimensionExprs.push(
        createInExpression(
          dimensionFilter.name,
          dimensionFilter.selectedValues,
          !dimensionFilter.isInclude,
        ),
      );
    }
  }

  const dimensionFilterExpression = createAndExpression(dimensionExprs);

  const dimensionThresholdFilters: DimensionThresholdFilter[] = [];
  const measureFiltersByDimension = new Map<string, MeasureFilterEntry[]>();

  for (const measureFilter of measureFilters) {
    if (measureFilter.filter && measureFilter.dimensionName) {
      if (!measureFiltersByDimension.has(measureFilter.dimensionName)) {
        measureFiltersByDimension.set(measureFilter.dimensionName, []);
      }
      measureFiltersByDimension
        .get(measureFilter.dimensionName)!
        .push(measureFilter.filter);
    }
  }

  for (const [dimensionName, filters] of measureFiltersByDimension) {
    dimensionThresholdFilters.push({
      name: dimensionName,
      filters,
    });
  }

  return {
    whereFilter: mergeDimensionAndMeasureFilters(
      dimensionFilterExpression,
      dimensionThresholdFilters,
    ),
    dimensionsWithInListFilter,
  };
}
