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

type UIFilters = {
  dimensions: Map<string, DimensionFilterItem>;
  measures: Map<string, MeasureFilterItem>;
  complexFilters: V1Expression[];
  hasFilters: boolean;
  hasClearableFilters: boolean;
};

type MetricsViewName = string;
type DimensionName = string;

type Lookup<T> = Map<MetricsViewName, Map<DimensionName, T>>;

export type DimensionLookup = Lookup<MetricsViewSpecDimension>;
export type MeasureLookup = Lookup<MetricsViewSpecMeasure>;

type ParsedFilters = ReturnType<typeof initFilterBase>;

function initFilterBase() {
  return {
    where: createAndExpression([]),
    dimensionFilter: createAndExpression([]),
    string: "",
    metricsViewName: "",
    dimensionsWithInlistFilter: <string[]>[],
    dimensionThresholdFilters: <DimensionThresholdFilter[]>[],
    measures: new Map<string, MeasureFilterItem>(),
    dimensions: new Map<string, DimensionFilterItem>(),
    complex: false,
  };
}

// wip - bgh
function mergeFilters<T>(
  metricsViewItems: Map<string, Map<string, T>>,
  locations: Map<string, string[]>,
  mergeStrategy: "all" = "all",
): Lookup<T> {
  const merged = new Map<string, Map<string, T>>();

  if (mergeStrategy === "all") {
    locations.forEach((mvNames, name) => {
      const key = `${mvNames.sort().join("//")}::${name}`;
      const dimMap = new Map<string, T>();
      mvNames.forEach((mvName) => {
        const dim = metricsViewItems.get(mvName)?.get(name);
        if (dim) {
          dimMap.set(mvName, dim);
        }
      });
      merged.set(key, dimMap);
    });
  }

  return merged;
}

// wip - bgh
export class FilterManager {
  metricsViewFilters: Map<string, NewFilters> = new Map();
  _pinnedFilterKeys = writable<Set<string>>(new Set());
  _temporaryFilterKeys = writable<Set<string>>(new Set());
  _allDimensions = writable<DimensionLookup>(new Map());
  _allMeasures = writable<MeasureLookup>(new Map());
  _dimensionFilterKeys: Readable<string[]>;
  _defaultUIFilters: Readable<UIFilters>;
  _defaultExpression: Readable<V1Expression>;
  allMetricsViewNamesPrefix = writable<string>("");
  _activeUIFilters: Readable<UIFilters>;
  _activeExpression: Readable<V1Expression>;
  metricsViewNameDimensionMap: Map<
    string,
    Map<string, MetricsViewSpecDimension>
  > = new Map();
  metricsViewNameMeasureMap: Map<string, Map<string, MetricsViewSpecMeasure>> =
    new Map();

  updateConfig(
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
    const allMetricsViewNames = Object.keys(metricsViews);
    const allMetricsViewNamesPrefix = allMetricsViewNames.join(".");
    this.allMetricsViewNamesPrefix.set(allMetricsViewNamesPrefix);

    const dimensionIdMap: Map<string, string[]> = new Map();
    const measureIdMap: Map<string, string[]> = new Map();

    Object.entries(metricsViews).forEach(([name, mv]) => {
      if (mv) {
        this.metricsViewNameDimensionMap.set(name, new Map());
        this.metricsViewNameMeasureMap.set(name, new Map());

        mv.state?.validSpec?.dimensions?.forEach((dim) => {
          const dimName = dim.name;
          if (!dimName) return;

          const array = dimensionIdMap.get(dimName) || [];
          array.push(name);
          dimensionIdMap.set(dimName, array);
          this.metricsViewNameDimensionMap.get(name)?.set(dimName, dim);
        });

        mv.state?.validSpec?.measures?.forEach((measure) => {
          const measureName = measure.name;
          if (!measureName) return;

          const array = measureIdMap.get(measureName) || [];
          array.push(name);
          measureIdMap.set(measureName, array);
          this.metricsViewNameMeasureMap.get(name)?.set(measureName, measure);
        });

        const existingFilterStore = this.metricsViewFilters.get(name);

        if (existingFilterStore) {
          existingFilterStore.update(mv, defaultFilters?.[name]);
        } else {
          this.metricsViewFilters.set(
            name,
            new NewFilters(mv, name, defaultFilters?.[name], this),
          );
        }
      }
    });

    const mergedDimensions = mergeFilters(
      this.metricsViewNameDimensionMap,
      dimensionIdMap,
      "all",
    );

    const mergedMeasures = mergeFilters(
      this.metricsViewNameMeasureMap,
      measureIdMap,
      "all",
    );

    this._allMeasures.set(mergedMeasures);

    this._allDimensions.set(mergedDimensions);

    if (pinnedFilters) {
      const keys = new Set<string>();

      pinnedFilters.forEach((filterName) => {
        const foundDimensions = new Map();

        this.metricsViewFilters.forEach((filters, name) => {
          const foundDimension = this.metricsViewNameDimensionMap
            .get(name)
            ?.get(filterName);
          if (foundDimension) {
            foundDimensions.set(name, foundDimension);
          }
        });

        const filterKey = `${Array.from(foundDimensions.keys()).join("//")}::${filterName}`;

        keys.add(filterKey);
      });
      this._pinnedFilterKeys.set(keys);
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
        this._pinnedFilterKeys,
        this._temporaryFilterKeys,
        ...Array.from(this.metricsViewFilters.values()).map((f) => f.parsed),
      ],
      ([pinnedFilters, temporaryFilterKeys, ...parsedFilters]) => {
        const parsedMap = new Map<string, ParsedFilters>();
        parsedFilters.forEach((parsed) => {
          parsedMap.set(parsed.metricsViewName, parsed);
        });

        const merged = {
          dimensions: new Map<string, DimensionFilterItem>(),
          measures: new Map<string, MeasureFilterItem>(),
          complexFilters: [],
          hasFilters: false,
          hasClearableFilters: false,
        };

        const allMeasures = get(this._allMeasures);

        allMeasures.forEach((measures, key) => {
          const filters: MeasureFilterItem[] = [];

          const measureMap = allMeasures.get(key);
          const metricsViewNames = measureMap
            ? Array.from(measureMap.keys())
            : [];

          const measureSpecs = Array.from(measureMap?.values() || []);

          // Needs work - bgh
          if (temporaryFilterKeys.has(key)) {
            merged.measures.set(key, {
              dimensionName: "",
              dimensions: undefined,
              name: key.split("::")[1],
              label: measureSpecs[0].displayName ?? "",
              pinned: false,
              measures: measureMap,
              metricsViewNames: metricsViewNames,
            });
            return;
          }

          const pinned = pinnedFilters.has(key);

          measures.forEach((measure, metricsViewName) => {
            const parsed = parsedMap.get(metricsViewName);
            if (!parsed) return;

            const dimFilter = parsed.measures.get(measure.name as string);
            if (!dimFilter) {
              if (pinned) {
                filters.push({
                  dimensionName: "",
                  dimensions: undefined,
                  name: key.split("::")[1],
                  label: measureSpecs[0].displayName ?? "",
                  pinned: true,
                  measures: measureMap,
                  metricsViewNames: metricsViewNames,
                });
              }
            } else {
              if (pinned) {
                dimFilter.pinned = true;
              }
              filters.push(dimFilter);
            }
          });

          if (filters.length === 0) return;
          // if (
          //   filters.every(
          //     (f) =>
          //       f.measure === filters[0].measure &&
          //       f.operation === filters[0].operation &&
          //       f.type === filters[0].type,
          //   )
          // ) {
          merged.measures.set(key, {
            ...filters[0],
            measures: measures,
          });
          // } else {
          //   // mixed filters - need to resolve
          // }
        });

        const allDimensions = get(this._allDimensions);

        // can improve efficiency at a later date - bgh
        allDimensions.forEach((dimensions, key) => {
          const filters: DimensionFilterItem[] = [];

          if (temporaryFilterKeys.has(key)) {
            filters.push({
              mode: DimensionFilterMode.Select,
              selectedValues: [],
              dimensions: dimensions,
              isInclude: true,
              inputText: undefined,
              pinned: false,
            });

            merged.dimensions.set(key, {
              ...filters[0],
              dimensions: dimensions,
            });
            return;
          }

          const pinned = pinnedFilters.has(key);

          dimensions.forEach((dimension, metricsViewName) => {
            const parsed = parsedMap.get(metricsViewName);
            if (!parsed) return;

            const dimFilter = parsed.dimensions.get(dimension.name as string);
            if (!dimFilter) {
              if (pinned) {
                filters.push({
                  mode: DimensionFilterMode.Select,
                  selectedValues: [],
                  dimensions: dimensions,
                  isInclude: true,
                  inputText: undefined,
                  pinned: true,
                });
              }
            } else {
              if (pinned) {
                dimFilter.pinned = true;
              }
              filters.push(dimFilter);
            }
          });

          if (filters.length === 0) return;
          if (
            filters.every(
              (f) =>
                f.isInclude === filters[0].isInclude &&
                f.mode === filters[0].mode,
            )
          ) {
            merged.dimensions.set(key, {
              ...filters[0],
              dimensions: dimensions,
            });
          } else {
            // mixed filters - need to resolve
          }
        });

        merged.hasClearableFilters = parsedFilters.some(
          (p) => p.dimensions.size > 0 || p.measures.size > 0,
        );

        merged.hasFilters =
          merged.hasClearableFilters ||
          pinnedFilters.size > 0 ||
          temporaryFilterKeys.size > 0;

        return merged;
      },
    );
  }

  actions = {
    applyDimensionContainsMode: async (
      dimensionName: string,
      searchText: string,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);
      const map = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);
        if (!filterClass) return;
        const string = filterClass.applyDimensionContainsMode(
          dimensionName,
          searchText,
        );

        map.set(name, string || null);
      });
      await this.applyFiltersToUrl(map);
    },
    removeDimensionFilter: async (
      dimensionName: string,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);

      const map = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);
        if (!filterClass) return;
        const string = filterClass.removeDimensionFilter(dimensionName);
        map.set(name, string || null);
      });

      await this.applyFiltersToUrl(map, true);
    },
    addTemporaryFilter: (measureOrDimensionKey: string) => {
      this._temporaryFilterKeys.update((tempFilters) => {
        tempFilters.add(measureOrDimensionKey);
        return tempFilters;
      });
    },
    toggleDimensionFilterMode: async (
      dimensionName: string,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);

      const map = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);

        if (!filterClass) return;
        const string = filterClass.toggleDimensionFilterMode(dimensionName);

        if (!string) return;

        map.set(name, string);
      });

      await this.applyFiltersToUrl(map);
    },
    applyDimensionInListMode: async (
      dimensionName: string,
      values: string[],
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);
      const map = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);

        if (!filterClass) return;
        const string = filterClass.applyDimensionInListMode(
          dimensionName,
          values,
        );

        map.set(name, string || null);
      });

      await this.applyFiltersToUrl(map);
    },
    toggleDimensionValueSelections: async (
      dimensionName: string,
      dimensionValues: string[],
      metricsViewNames: string[],
      keepPillVisible?: boolean,
      isExclusiveFilter?: boolean,
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);

      const newFilters = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);

        if (!filterClass) return;

        const string = filterClass.toggleDimensionValueSelections(
          dimensionName,
          dimensionValues,
          keepPillVisible,
          isExclusiveFilter,
        );

        newFilters.set(name, string || null);
      });

      await this.applyFiltersToUrl(newFilters);
    },
    setMeasureFilter: async (
      dimensionName: string,
      filter: MeasureFilterEntry,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);

      const newFilters = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);

        if (!filterClass) return;

        const string = filterClass.setMeasureFilter(dimensionName, filter);

        newFilters.set(name, string || null);
      });

      await this.applyFiltersToUrl(newFilters);
    },
    removeMeasureFilter: async (
      dimensionName: string,
      measureName: string,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(dimensionName, metricsViewNames);

      const newFilters = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);

        if (!filterClass) return;

        const string = filterClass.removeMeasureFilter(
          dimensionName,
          measureName,
        );

        newFilters.set(name, string || null);
      });

      await this.applyFiltersToUrl(newFilters, true);
    },
    toggleFilterPin: (
      name: string,

      metricsViewNames: string[],
    ) => {
      this._pinnedFilterKeys.update((pinned) => {
        const key = metricsViewNames.sort().join("//") + "::" + name;
        if (pinned.has(key)) {
          pinned.delete(key);
        } else {
          pinned.add(key);
        }
        return pinned;
      });
    },
  };

  checkTemporaryFilter = (
    measureOrDimensionName: string,
    metricsViewNames: string[],
  ) => {
    const key =
      metricsViewNames.sort().join("//") + "::" + measureOrDimensionName;
    const tempFilters = get(this._temporaryFilterKeys);
    const test = tempFilters.delete(key);
    if (test) {
      this._temporaryFilterKeys.set(tempFilters);
    }
  };

  // Unclear on what this actually should do - bgh
  // Go to defaults or truly clear all filters?
  clearAllFilters = async () => {
    await goto(`?clear=true`);
  };

  applyFiltersToUrl = async (
    filters: Map<string, string | null>,
    allowFilterClear = false,
  ) => {
    const existingParams = new URLSearchParams(window.location.search);

    existingParams.delete("default");
    existingParams.delete("clear");
    existingParams.delete(ExploreStateURLParams.Filters);

    filters.forEach((filterString, mvName) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
      if (filterString === null) {
        existingParams.delete(paramKey);
      } else {
        existingParams.set(paramKey, filterString);
      }
    });

    const string = existingParams.toString();

    if (!string) {
      if (allowFilterClear) {
        await goto(`?clear=true`);
        return;
      } else {
        await goto(`?default=true`);
        return;
      }
    } else {
      await goto(`?${string}`);
    }
  };
}

// wip - bgh
export class NewFilters {
  parsed = writable(initFilterBase());
  parsedDefaultFilters = writable<ParsedFilters>(initFilterBase());
  // dimensionMap: Map<string, MetricsViewSpecDimension> = new Map();

  constructor(
    metricsView: V1MetricsView,
    private metricsViewName: string,
    defaultExpression: string | undefined,
    private manager: FilterManager,
  ) {
    this.update(metricsView, defaultExpression);
  }

  update(metricsView: V1MetricsView, defaultExpression?: string) {
    this.parsedDefaultFilters.set(this.parseFilterString(defaultExpression));
  }

  parseFilterString(filterString: string = ""): ParsedFilters {
    const { expr, dimensionsWithInlistFilter } =
      getFiltersFromText(filterString);

    const { dimensionThresholdFilters, dimensionFilters } =
      splitWhereFilter(expr);

    const isComplexFilter = isExpressionUnsupported(expr);

    if (isComplexFilter) {
      return {
        string: filterString,
        where: expr,
        dimensionFilter: dimensionFilters,
        metricsViewName: this.metricsViewName,
        dimensionsWithInlistFilter,
        dimensionThresholdFilters,
        dimensions: new Map(),
        measures: new Map(),
        complex: true,
      };
    }

    const dimensionMap =
      this.manager.metricsViewNameDimensionMap.get(this.metricsViewName) ??
      new Map<string, MetricsViewSpecDimension>();
    const measureMap =
      this.manager.metricsViewNameMeasureMap.get(this.metricsViewName) ??
      new Map<string, MetricsViewSpecMeasure>();

    const processed = processExpression({
      expr: dimensionFilters,
      measureMap,
      dimensionMap,
      metricsViewName: this.metricsViewName,
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
    });

    return {
      string: filterString,
      where: expr,
      dimensionFilter: dimensionFilters,
      metricsViewName: this.metricsViewName,
      dimensionsWithInlistFilter,
      dimensionThresholdFilters,
      ...processed,
      complex: false,
    };
  }

  removeDimensionFilter = (dimensionName: string) => {
    console.log("new filter remove", { dimensionName });
    const {
      where: wf,
      dimensionThresholdFilters,
      dimensionsWithInlistFilter,
    } = get(this.parsed);
    const exprIdx = wf.cond?.exprs?.findIndex(
      (e) => e.cond?.exprs?.[0].ident === dimensionName,
    );
    if (!(exprIdx === undefined || exprIdx === -1)) {
      wf.cond?.exprs?.splice(exprIdx, 1);
    }

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

  toggleDimensionValueSelections = (
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

  setMeasureFilter = (dimensionName: string, filter: MeasureFilterEntry) => {
    const {
      where: wf,
      dimensionThresholdFilters: dtfs,
      dimensionsWithInlistFilter,
      dimensionFilter,
    } = get(this.parsed);

    const dimIdx = dtfs.findIndex((dtf) => dtf.name === dimensionName);
    let dimThresholdFilter = dtfs[dimIdx];

    if (!dimThresholdFilter) {
      dimThresholdFilter = { name: dimensionName, filters: [] };
      dtfs.push(dimThresholdFilter);
    } else {
      const filters = dimThresholdFilter.filters;
      const exprIdx = filters.findIndex((f) => f.measure === filter.measure);
      if (exprIdx !== -1) {
        filters.splice(exprIdx, 1);
      }
    }

    const exprIdx = dimThresholdFilter.filters.findIndex(
      (f) => f.measure === filter.measure,
    );
    if (exprIdx === -1) {
      dimThresholdFilter.filters.push(filter);
    } else {
      dimThresholdFilter.filters.splice(exprIdx, 1, filter);
    }

    return getFilterParam(dimensionFilter, dtfs, dimensionsWithInlistFilter);
  };
  removeMeasureFilter = (dimensionName: string, measureName: string) => {
    const {
      dimensionThresholdFilters: dtfs,
      dimensionsWithInlistFilter,
      dimensionFilter,
    } = get(this.parsed);

    const dimIdx = dtfs.findIndex((dtf) => dtf.name === dimensionName);
    const dimThresholdFilter = dtfs[dimIdx];

    if (dimThresholdFilter) {
      const filters = dimThresholdFilter.filters;
      const exprIdx = filters.findIndex((f) => f.measure === measureName);
      if (exprIdx !== -1) {
        filters.splice(exprIdx, 1);
      }
      if (filters.length === 0) {
        dtfs.splice(dimIdx, 1);
      }
    }

    return getFilterParam(dimensionFilter, dtfs, dimensionsWithInlistFilter);
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
              inputText: undefined,
              dimensions: new Map<string, MetricsViewSpecDimension>(),
              pinned: false,
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
  measureMap,
  metricsViewName,
  dimensionsWithInlistFilter,
  dimensionThresholdFilters,
}: {
  expr: V1Expression;
  measureMap: Map<string, MetricsViewSpecMeasure>;
  dimensionMap: Map<string, MetricsViewSpecDimension>;
  metricsViewName: string;
  dimensionsWithInlistFilter: string[];
  dimensionThresholdFilters: DimensionThresholdFilter[];
}): UIFilters {
  const isComplex = isExpressionUnsupported(expr);
  const dimensions = getDimensionFilterItemsMap(
    dimensionMap,
    expr,
    dimensionsWithInlistFilter,
    metricsViewName,
  );
  const measures = getCanvasMeasureFiltersMap(
    measureMap,
    dimensionThresholdFilters,
    metricsViewName,
  );
  return {
    complexFilters: isComplex ? [expr] : [],
    measures: measures,
    dimensions: dimensions,
    hasFilters: dimensions.size > 0,
    hasClearableFilters: dimensions.size > 0,
  };
}

export function getCanvasMeasureFiltersMap(
  measureIdMap: Map<string, MetricsViewSpecMeasure>,
  dimensionThresholdFilters: DimensionThresholdFilter[],
  metricsViewName: string,
): Map<string, MeasureFilterItem> {
  const map = new Map();

  dimensionThresholdFilters.forEach((dtf) => {
    const filter = dtf.filters[0];
    const measureName = filter.measure;
    const measure = measureIdMap.get(measureName);
    if (!measure) return;

    const entry: MeasureFilterItem = {
      dimensionName: dtf.name,
      name: measureName,
      label: measure.displayName || measure.expression || filter.measure,
      filter: filter,

      // dimensions,
    };

    map.set(measureName, entry);
  });

  return map;
}

export function getDimensionFilterItemsMap(
  dimensionIdMap: Map<string, MetricsViewSpecDimension>,
  filter: V1Expression | undefined,
  dimensionsWithInlistFilter: string[],
  metricsViewName: string,
): Map<string, DimensionFilterItem> {
  if (!filter) return new Map();
  const filteredDimensions: Map<string, DimensionFilterItem> = new Map();
  const addedDimension = new Set<string>();

  forEachIdentifier(filter, (e, ident) => {
    if (addedDimension.has(ident) || !dimensionIdMap.has(ident)) return;
    const dim = dimensionIdMap.get(ident);

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
        name: ident,
      });
    } else if (
      op === V1Operation.OPERATION_LIKE ||
      op === V1Operation.OPERATION_NLIKE
    ) {
      filteredDimensions.set(ident, {
        name: ident,
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
