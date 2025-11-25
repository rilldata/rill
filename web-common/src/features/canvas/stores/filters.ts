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
  getDimensionFiltersMap,
  type DimensionFilterItem,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
  forEachIdentifier,
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
  V1CanvasPresetFilterExpr,
  V1Expression,
  V1MetricsView,
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
import { goto } from "$app/navigation";

export type CanvasDimensionFilter = {
  dimensions: Map<string, MetricsViewSpecDimension>;
  mode: DimensionFilterMode;
  selectedValues: string[];
  inputText: string | undefined;
  isInclude: boolean;
  pinned: boolean;
};

type UIFilters = {
  dimensions: Map<string, CanvasDimensionFilter>;
  measures: Map<string, MeasureFilterItem>;
};

// wip - bgh
export class FilterManager {
  metricsViewFilters: Map<string, NewFilters> = new Map();
  private pinnedFilters = writable<Map<string, CanvasDimensionFilter>>(
    new Map(),
  );
  temporaryFilters = writable<Map<string, CanvasDimensionFilter>>(new Map());
  _allDimensions = writable(
    new Map<
      string,
      MetricsViewSpecDimension & { metricsViewNames: string[] }
    >(),
  );
  _dimensionFilterKeys: Readable<string[]>;
  _defaultUIFilters: Readable<UIFilters>;
  _defaultExpression: Readable<V1Expression>;
  allMetricsViewNamesPrefix = writable<string>("");
  _activeUIFilters: Readable<UIFilters>;
  _activeExpression: Readable<V1Expression>;

  updateConfig(
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
    const allMetricsViewNames = Object.keys(metricsViews);
    const allMetricsViewNamesPrefix = allMetricsViewNames.join(".");
    this.allMetricsViewNamesPrefix.set(allMetricsViewNamesPrefix);

    Object.entries(metricsViews).forEach(([name, mv]) => {
      if (mv) {
        mv.state?.validSpec?.dimensions?.forEach((dim) => {
          const dimName = dim.name;
          if (!dimName) return;
          this._allDimensions.update((allDims) => {
            allDims.set(`${allMetricsViewNamesPrefix}-${dimName}`, {
              ...dim,
              metricsViewNames: allMetricsViewNames,
            });
            return allDims;
          });
        });
        const existingFilterStore = this.metricsViewFilters.get(name);
        if (existingFilterStore) {
          existingFilterStore.update(mv, defaultFilters?.[name]);
        } else {
          this.metricsViewFilters.set(
            name,
            new NewFilters(mv, name, defaultFilters?.[name]),
          );
        }
      }
    });

    if (pinnedFilters) {
      const map = new Map();

      pinnedFilters.forEach((f) => {
        const foundDimensions = new Map();

        this.metricsViewFilters.entries().forEach(([name, filters]) => {
          const foundDimension = filters.dimensionMap.get(f);
          if (foundDimension) {
            foundDimensions.set(name, foundDimension);
          }
        });
        map.set(f, {
          mode: DimensionFilterMode.Select,
          selectedValues: [],
          isInclude: true,
          inputText: undefined,
          dimensions: foundDimensions,
          pinned: true,
        });
      });
      this.pinnedFilters.set(map);
    }
  }

  constructor(
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
    this.updateConfig(metricsViews, pinnedFilters, defaultFilters);

    this._defaultUIFilters = derived(
      [
        ...Array.from(this.metricsViewFilters.values()).map(
          (f) => f.parsedDefaultFilters,
        ),
      ],
      (expr) => {
        return expr[0];
      },
    );

    this._activeUIFilters = derived(
      [
        this.pinnedFilters,
        this.temporaryFilters,
        ...Array.from(this.metricsViewFilters.values()).map((f) => f.parsed),
      ],
      ([pinnedFilters, temporaryFilters, ...filters]) => {
        const effective = filters[0];

        console.log({ filters });

        pinnedFilters.entries().forEach(([name, filter]) => {
          const existing = effective.dimensions.get(name);
          if (!existing) {
            effective.dimensions.set(name, filter);
          } else {
            existing.pinned = true;
          }
        });

        temporaryFilters.forEach((keyedFilter, name) => {
          const existing = effective.dimensions.get(name);
          if (!existing) {
            effective.dimensions.set(name, keyedFilter);
          }
        });

        return effective;
      },
    );
  }

  applyDimensionContainsMode = async (
    dimensionName: string,
    searchText: string,
    metricsViewNames: string[],
  ) => {
    this.checkTemporaryFilter(dimensionName);
    const searchParams = new URLSearchParams();

    metricsViewNames.forEach((name) => {
      const filterClass = this.metricsViewFilters.get(name);
      if (!filterClass) return;
      const string = filterClass.applyDimensionContainsMode(
        dimensionName,
        searchText,
      );
      if (string) {
        searchParams.set(`${ExploreStateURLParams.Filters}.${name}`, string);
      }
    });
    const string = searchParams.toString();

    if (!string) {
      await goto(`?default=true`);
    } else {
      await goto(`?${string}`);
    }
  };

  removeDimensionFilter = async (
    dimensionName: string,
    metricsViewNames: string[],
  ) => {
    const searchParams = new URLSearchParams();
    metricsViewNames.forEach((name) => {
      const filterClass = this.metricsViewFilters.get(name);
      if (!filterClass) return;
      const string = filterClass.removeDimensionFilter(dimensionName);
      if (string) {
        searchParams.set(`${ExploreStateURLParams.Filters}.${name}`, string);
      }
    });
    const string = searchParams.toString();

    if (!string) {
      // Weird behavior when removing last filter
      // Need to consider other options - bgh
      await goto(`?default=true`);
    } else {
      await goto(`?${string}`);
    }
  };

  checkTemporaryFilter = (filterName: string) => {
    const tempFilters = get(this.temporaryFilters);
    if (tempFilters.has(filterName)) {
      tempFilters.delete(filterName);

      this.temporaryFilters.set(tempFilters);
    }
  };

  addTemporaryFilter = (dimensionName: string) => {
    const foundDimensions = new Map<string, MetricsViewSpecDimension>();

    this.metricsViewFilters.entries().forEach(([name, filters]) => {
      const foundDimension = filters.dimensionMap.get(dimensionName);
      if (foundDimension) {
        foundDimensions.set(name, foundDimension);
      }
    });
    this.temporaryFilters.update((tempFilters) => {
      tempFilters.set(dimensionName, {
        mode: DimensionFilterMode.Select,
        selectedValues: [],
        dimensions: foundDimensions,
        isInclude: true,
        inputText: undefined,
        pinned: false,
      });
      return tempFilters;
    });
  };

  toggleDimensionFilterMode = async (
    dimensionName: string,
    metricsViewNames: string[],
  ) => {
    this.checkTemporaryFilter(dimensionName);
    const searchParams = new URLSearchParams();

    metricsViewNames.forEach((name) => {
      const filterClass = this.metricsViewFilters.get(name);

      if (!filterClass) return;
      const string = filterClass.toggleDimensionFilterMode(dimensionName);

      if (!string) return;

      searchParams.set(`${ExploreStateURLParams.Filters}.${name}`, string);
    });

    const string = searchParams.toString();

    if (!string) {
      await goto(`?default=true`);
    } else {
      await goto(`?${string}`);
    }
  };

  // Unclear on what this actually should do - bgh
  // Go to defaults or truly clear all filters?
  clearAllFilters = async () => {
    await goto(`?default=true`);
  };

  applyDimensionInListMode = async (
    dimensionName: string,
    values: string[],
    metricsViewNames: string[],
  ) => {
    this.checkTemporaryFilter(dimensionName);
    const searchParams = new URLSearchParams();

    metricsViewNames.forEach((name) => {
      const filterClass = this.metricsViewFilters.get(name);

      if (!filterClass) return;
      const string = filterClass.applyDimensionInListMode(
        dimensionName,
        values,
      );

      if (!string) return;

      searchParams.set(`${ExploreStateURLParams.Filters}.${name}`, string);
    });

    const string = searchParams.toString();

    if (!string) {
      await goto(`?default=true`);
    } else {
      await goto(`?${string}`);
    }
  };

  toggleMultipleDimensionValueSelections = async (
    dimensionName: string,
    dimensionValues: string[],
    metricsViewNames: string[],
    keepPillVisible?: boolean,
    isExclusiveFilter?: boolean,
  ) => {
    this.checkTemporaryFilter(dimensionName);
    const searchParams = new URLSearchParams();

    metricsViewNames.forEach((name) => {
      const filterClass = this.metricsViewFilters.get(name);

      if (!filterClass) return;
      const string = filterClass.toggleMultipleDimensionValueSelections(
        dimensionName,
        dimensionValues,
        keepPillVisible,
        isExclusiveFilter,
      );

      if (!string) return;

      searchParams.set(`${ExploreStateURLParams.Filters}.${name}`, string);
    });

    const string = searchParams.toString();

    if (!string) {
      await goto(`?default=true`);
    } else {
      await goto(`?${string}`);
    }
  };
}

type ParsedFilters = {
  string: string;
  where: V1Expression;
  measures: Map<string, MeasureFilterItem>;
  dimensions: Map<string, CanvasDimensionFilter>;
  dimensionsWithInlistFilter: string[];
  dimensionThresholdFilters: DimensionThresholdFilter[];
};

const parsedFilterBase: ParsedFilters = {
  where: {
    cond: {
      op: "OPERATION_AND",
      exprs: [],
    },
  },
  string: "",
  dimensionsWithInlistFilter: [],
  dimensionThresholdFilters: [],
  measures: new Map(),
  dimensions: new Map(),
};

// wip - bgh
export class NewFilters {
  parsed = writable<ParsedFilters>(parsedFilterBase);
  parsedDefaultFilters = writable<ParsedFilters>(parsedFilterBase);
  dimensionMap: Map<string, MetricsViewSpecDimension> = new Map();

  constructor(
    metricsView: V1MetricsView,
    private metricsViewName: string,
    defaultExpression?: string,
  ) {
    this.update(metricsView, defaultExpression);
  }

  update(metricsView: V1MetricsView, defaultExpression?: string) {
    metricsView.state?.validSpec?.dimensions?.forEach((dim) => {
      if (!dim.name) return;
      this.dimensionMap.set(dim.name, dim);
    });

    this.parsedDefaultFilters.set(this.parseFilterString(defaultExpression));
  }

  parseFilterString(filterString: string = "") {
    const { expr, dimensionsWithInlistFilter } =
      getFiltersFromText(filterString);

    const { dimensionThresholdFilters } = splitWhereFilter(expr);

    console.log(this.metricsViewName, { expr, map: this.dimensionMap });

    const processed = processExpression({
      expr,
      dimensionMap: this.dimensionMap,
      metricsViewName: this.metricsViewName,
      dimensionsWithInlistFilter,
    });

    console.log({ processed });

    return {
      string: filterString,
      where: expr,
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
      ...processed,
    };
  }

  removeDimensionFilter = (dimensionName: string) => {
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    } = get(this.parsed);
    const exprIdx = wf.cond?.exprs?.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (exprIdx === undefined || exprIdx === -1) return;
    wf.cond?.exprs?.splice(exprIdx, 1);

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };

  applyDimensionContainsMode = (dimensionName: string, searchText: string) => {
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    } = get(this.parsed);

    const exprIndex = wf.cond?.exprs?.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );

    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(
        createLikeExpression(dimensionName, `%${searchText}%`, false),
      );
    } else {
      const operation = wf.cond!.exprs![exprIndex].cond!.op;
      const isExclude =
        operation === V1Operation.OPERATION_NLIKE ||
        operation === V1Operation.OPERATION_NIN;
      wf.cond!.exprs![exprIndex] = createLikeExpression(
        dimensionName,
        `%${searchText}%`,
        isExclude,
      );
    }
    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };

  onFilterStringChange(filterString: string) {
    this.parsed.set(this.parseFilterString(filterString));
  }

  toggleDimensionFilterMode = (dimensionName: string) => {
    const {
      dimensionsWithInlistFilter,
      where: wf,
      dimensionThresholdFilters,
    } = get(this.parsed);

    if (!wf.cond?.exprs) return;
    const exprIdx = wf.cond.exprs.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (exprIdx === -1) return;
    wf.cond.exprs[exprIdx] = negateExpression(wf.cond.exprs[exprIdx]);

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };

  toggleMultipleDimensionValueSelections = (
    dimensionName: string,
    dimensionValues: string[],
    keepPillVisible?: boolean,
    isExclusiveFilter?: boolean,
    exclude: boolean = false,
  ) => {
    const {
      where: wf,
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
    } = get(this.parsed);

    let exprIndex =
      wf.cond?.exprs?.findIndex(
        (e) => e.cond?.exprs?.[0].ident === dimensionName,
      ) ?? -1;
    let expr = wf.cond?.exprs?.[exprIndex];

    const wasLikeFilter =
      expr?.cond?.op === V1Operation.OPERATION_LIKE ||
      expr?.cond?.op === V1Operation.OPERATION_NLIKE;
    if (!expr?.cond?.exprs || wasLikeFilter) {
      expr = createInExpression(dimensionName, [], exclude);
      wf.cond?.exprs?.push(expr);
      exprIndex = wf.cond!.exprs!.length - 1;
    }

    const wasInListFilter = dimensionsWithInlistFilter.includes(dimensionName);
    if (wasInListFilter) {
      dimensionsWithInlistFilter.filter((d) => d !== dimensionName);
    }

    dimensionValues.forEach((dimensionValue) => {
      toggleDimensionFilterValue(expr, dimensionValue, !!isExclusiveFilter);
    });

    if (expr?.cond?.exprs?.length === 1) {
      wf.cond?.exprs?.splice(exprIndex, 1);

      if (keepPillVisible) {
        // this.setTemporaryFilterName(dimensionName);
      }
    }

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };

  applyDimensionInListMode = (dimensionName: string, values: string[]) => {
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    } = get(this.parsed);
    const isExclude = false;

    const expr = createInExpression(dimensionName, values, isExclude);

    dimensionsWithInlistFilter.push(dimensionName);

    const exprIndex =
      wf.cond?.exprs?.findIndex(
        (e) => e.cond?.exprs?.[0].ident === dimensionName,
      ) ?? -1;
    if (exprIndex === undefined || exprIndex === -1) {
      wf.cond!.exprs!.push(expr);
    } else {
      wf.cond!.exprs![exprIndex] = expr;
    }

    return getFilterParam(
      wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    );
  };
}

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
  pinnedFilters: Readable<Set<string>> = writable(new Set());

  constructor(
    metricsView: MetricsViewSelectors,
    public searchParamsStore: SearchParamsStore,
    public componentName?: string,
    public metricsViewName?: string,
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
        this.pinnedFilters,
      ],
      ([
        tempFilters,
        dimensionFilters,
        $allDimensions,
        $excludeMode,
        pinnedFilters,
      ]) => {
        const merged = structuredClone(dimensionFilters);

        [...tempFilters, ...pinnedFilters].forEach((tempFilter) => {
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
      const names = get(this.metricsView.metricViewNames);

      names.forEach((mvName) => {
        const filtersKey = mvName
          ? `${ExploreStateURLParams.Filters}.${mvName}`
          : ExploreStateURLParams.Filters;
        const tempFiltersKey = mvName
          ? `${ExploreStateURLParams.TemporaryFilters}.${mvName}`
          : ExploreStateURLParams.TemporaryFilters;
        const filterText = searchParams.get(filtersKey);
        if (!this.componentName) {
          const tempFilters = searchParams.get(tempFiltersKey);
          this.temporaryFilters.set(new Set(tempFilters?.split(",") ?? []));
        }

        this.setFiltersFromText(filterText ?? "");
      });
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
          undefined,
          undefined,
          get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
        undefined,
        undefined,
        get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
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
    this.searchParamsStore.set(
      ExploreStateURLParams.Filters,
      undefined,
      undefined,
      undefined,
      get(this.metricsView.metricViewNames),
    );
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
      this.searchParamsStore.set(
        ExploreStateURLParams.TemporaryFilters,
        name,
        undefined,
        undefined,
        get(this.metricsView.metricViewNames),
      );
    }
  };

  setFiltersFromText = (filterText: string) => {
    const { expr, dimensionsWithInlistFilter } = getFiltersFromText(filterText);

    this.setFilters(expr);
    this.dimensionsWithInlistFilter.set(dimensionsWithInlistFilter);
  };
}

export function getFilterParam(
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

function processExpression({
  expr,
  dimensionMap,
  metricsViewName,
  dimensionsWithInlistFilter,
}: {
  expr: V1Expression;
  dimensionMap: Map<string, MetricsViewSpecDimension>;
  metricsViewName: string;
  dimensionsWithInlistFilter: string[];
}): UIFilters {
  return {
    measures: new Map(),
    dimensions: getCanvasDimensionFiltersMap(
      dimensionMap,
      expr,
      dimensionsWithInlistFilter,
      metricsViewName,
    ),
  };
}

export function getCanvasDimensionFiltersMap(
  dimensionIdMap: Map<string, MetricsViewSpecDimension>,
  filter: V1Expression | undefined,
  dimensionsWithInlistFilter: string[],
  metricsViewName: string,
): Map<string, CanvasDimensionFilter> {
  if (!filter) return new Map();
  const filteredDimensions: Map<string, CanvasDimensionFilter> = new Map();
  const addedDimension = new Set<string>();

  forEachIdentifier(filter, (e, ident) => {
    console.log({ ident });
    if (addedDimension.has(ident) || !dimensionIdMap.has(ident)) return;
    const dim = dimensionIdMap.get(ident);
    console.log({ dim });
    if (!dim) {
      return;
    }
    addedDimension.add(ident);

    const op = e.cond?.op;
    if (op === V1Operation.OPERATION_IN || op === V1Operation.OPERATION_NIN) {
      const isInListMode = dimensionsWithInlistFilter.includes(ident);
      filteredDimensions.set(ident, {
        dimensions: new Map([[metricsViewName, dim]]),

        mode: isInListMode
          ? DimensionFilterMode.InList
          : DimensionFilterMode.Select,
        selectedValues: getValuesInExpression(e),
        isInclude: e.cond?.op === V1Operation.OPERATION_IN,
        inputText: undefined,
        pinned: false,
      });
    } else if (
      op === V1Operation.OPERATION_LIKE ||
      op === V1Operation.OPERATION_NLIKE
    ) {
      filteredDimensions.set(ident, {
        mode: DimensionFilterMode.Contains,
        selectedValues: [],
        inputText: e.cond?.exprs?.[1]?.val?.toString?.() ?? "",
        isInclude: e.cond?.op === V1Operation.OPERATION_LIKE,
        dimensions: new Map([[metricsViewName, dim]]),
        pinned: false,
      });
    }
  });

  return filteredDimensions;
}
