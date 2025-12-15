import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { type DimensionFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import type {
  MetricsViewSpecDimension,
  V1CanvasPresetFilterExpr,
  V1Expression,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import {
  V1Operation,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import {
  derived,
  get,
  type Readable,
  type Writable,
  writable,
} from "svelte/store";
import { ExploreStateURLParams } from "../../dashboards/url-state/url-params";
import { goto } from "$app/navigation";
import { FilterState } from "./filter-state";
import { getDimensionDisplayName } from "../../dashboards/filters/getDisplayName";
import type { ParsedFilters } from "./filter-state";
import { createAndExpression } from "../../dashboards/stores/filter-utils";

export type UIFilters = {
  dimensionFilters: Map<string, DimensionFilterItem>;
  measureFilters: Map<string, MeasureFilterItem>;
  complexFilters: V1Expression[];
  hasFilters: boolean;
  hasClearableFilters: boolean;
};

export type MetricsViewName = string;
type DimensionName = string;
type MeasureName = string;

type LookupKey = `${string}${typeof NAME_SEPARATOR}${string}`;

// // A Lookup allows you to take a unique dimension/measure key (after namespace merging)
// // And find all the corresponding measures/dimensions across all metrics views
type UniqueLookup<T> = Map<LookupKey, Map<MetricsViewName, T>>;

export type DimensionLookup = UniqueLookup<MetricsViewSpecDimension>;
export type MeasureLookup = UniqueLookup<MetricsViewSpecMeasure>;

const MV_NAME_SEPARATOR = "//";
const NAME_SEPARATOR = "::";

function getLookupKey(mvNames: string[], name: string): LookupKey {
  return `${mvNames.sort().join(MV_NAME_SEPARATOR)}${NAME_SEPARATOR}${name}`;
}

export class FilterManager {
  metricsViewFilters = new StoreOfStores<FilterState>();
  pinnedFilterKeysStore = writable<Set<string>>(new Set());
  defaultPinnedFilterKeysStore = writable<Set<string>>(new Set());
  temporaryFilterKeysStore = writable<Map<string, boolean>>(new Map());

  allDimensionsStore = writable<DimensionLookup>(new Map());
  allMeasuresStore = writable<MeasureLookup>(new Map());

  activeUIFiltersStore: Readable<UIFilters>;
  defaultUIFiltersStore: Readable<UIFilters>;

  filterMapStore: Readable<Map<MetricsViewName, V1Expression>>;

  dimensionsForMetricsView = writable(
    new Map<MetricsViewName, Map<DimensionName, MetricsViewSpecDimension>>(),
  );
  measuresForMetricsView = writable(
    new Map<MetricsViewName, Map<MeasureName, MetricsViewSpecMeasure>>(),
  );

  // Look up list of dimensions based on which combination of metrics views a measure is applicable for
  // {{metrics_view_1}}//{{metrics_view_2}} will get you all the dimensions in metrics_view_1 and metrics_view_2
  uniqueMetricsViewGroupingToDimensionArrayStore = writable(
    new Map<string, MetricsViewSpecDimension[]>(),
  );

  constructor(
    metricsViews: Record<string, V1MetricsView | undefined>,
    public instanceId: string,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
    this.updateConfig(metricsViews, pinnedFilters, defaultFilters);

    this.defaultUIFiltersStore = derived(
      [this.metricsViewFilters],
      ([metricsViewFilters], set) => {
        const stores = Array.from(metricsViewFilters.values()).map(
          (f) => f.parsedDefaultFilters,
        );
        derived(
          [
            this.defaultPinnedFilterKeysStore,
            this.allMeasuresStore,
            this.allDimensionsStore,
            ...stores,
          ],
          ([
            defaultPinnedFilterKeys,
            allMeasures,
            allDimensions,
            ...filters
          ]) => {
            return this.convertToUIFilters(
              filters,
              new Map(),
              defaultPinnedFilterKeys,
              allMeasures,
              allDimensions,
            );
          },
        ).subscribe(set);
      },
    );

    this.activeUIFiltersStore = derived(
      [this.metricsViewFilters],
      ([metricsViewFilters], set) => {
        const stores = Array.from(metricsViewFilters.values()).map(
          (f) => f.parsed,
        );

        derived(
          [
            this.pinnedFilterKeysStore,
            this.temporaryFilterKeysStore,
            this.allMeasuresStore,
            this.allDimensionsStore,
            ...stores,
          ],
          ([
            pinnedFilters,
            temporaryFilterKeys,
            allMeasures,
            allDimensions,
            ...filters
          ]) => {
            return this.convertToUIFilters(
              filters,
              temporaryFilterKeys,
              pinnedFilters,
              allMeasures,
              allDimensions,
            );
          },
        ).subscribe(set);
      },
    );

    this.filterMapStore = derived(
      [this.metricsViewFilters],
      ([metricsViewFilters], set) => {
        const stores = Array.from(metricsViewFilters.values()).map(
          (f) => f.parsed,
        );

        derived(stores, (filters) => {
          const map = new Map<string, V1Expression>();
          filters.forEach((expr, i) => {
            const mvName = Array.from(metricsViewFilters.keys())[i];
            map.set(mvName, expr.where);
          });
          return map;
        }).subscribe(set);
      },
    );
  }

  createLocalFilterStore = (metricsViewName: string) => {
    return new FilterState(metricsViewName, this, this.instanceId);
  };

  onUrlChange = (searchParams: URLSearchParams) => {
    const legacyFilter = searchParams.get(ExploreStateURLParams.Filters);

    this.metricsViewFilters.forEach((filters, mvName) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
      const filterString = searchParams.get(paramKey) ?? legacyFilter ?? "";

      filters.onFilterStringChange(filterString);
    });
  };

  updateConfig = (
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ): void => {
    const dimensionNameToMetricsViewNames: Map<
      DimensionName,
      MetricsViewName[]
    > = new Map();
    const measureNameToMetricsViewNames: Map<MeasureName, MetricsViewName[]> =
      new Map();

    const exhaustiveMeasures: Map<
      MetricsViewName,
      Map<DimensionName, MetricsViewSpecMeasure>
    > = new Map();
    const exhaustiveDimensions: Map<
      MetricsViewName,
      Map<DimensionName, MetricsViewSpecDimension>
    > = new Map();

    Object.entries(metricsViews).forEach(([metricsViewName, metricsView]) => {
      if (!metricsView) return;

      const { measures, dimensions } = metricsView.state?.validSpec || {};
      const dimensionsForMetricsView = new Map<
        DimensionName,
        MetricsViewSpecDimension
      >();
      updateNameToMetricsViews(
        dimensions?.filter((d) => d.type !== "DIMENSION_TYPE_TIME"),
        (d) => d.name,
        dimensionNameToMetricsViewNames,
        dimensionsForMetricsView,
        metricsViewName,
      );

      exhaustiveDimensions.set(metricsViewName, dimensionsForMetricsView);

      const measuresForMetricsView = new Map<
        MeasureName,
        MetricsViewSpecMeasure
      >();
      updateNameToMetricsViews(
        measures,
        (m) => m.name,
        measureNameToMetricsViewNames,
        measuresForMetricsView,
        metricsViewName,
      );

      exhaustiveMeasures.set(metricsViewName, measuresForMetricsView);

      let filterStore = this.metricsViewFilters.get(metricsViewName);
      if (!filterStore) {
        filterStore = new FilterState(metricsViewName, this, this.instanceId);
        this.metricsViewFilters.set(metricsViewName, filterStore);
      }

      const filter = defaultFilters?.[metricsViewName];
      filterStore.onDefaultExpressionChange(
        flattenExpression(filter?.expression),
      );
    });

    const mergedDimensions = mergeFilters(
      exhaustiveDimensions,
      dimensionNameToMetricsViewNames,
      "all",
    );

    const mergedMeasures = mergeFilters(
      exhaustiveMeasures,
      measureNameToMetricsViewNames,
      "all",
    );

    if (pinnedFilters) {
      const pinnedKeys = new Set<string>();

      pinnedFilters.forEach((filterName) => {
        const metricsViewNames = new Set<MetricsViewName>();

        this.metricsViewFilters.forEach((_, metricsViewName) => {
          const dimensionsForView = exhaustiveDimensions.get(metricsViewName);
          const measuresForView = exhaustiveMeasures.get(metricsViewName);

          if (dimensionsForView?.has(filterName)) {
            metricsViewNames.add(metricsViewName);
          }

          if (measuresForView?.has(filterName)) {
            metricsViewNames.add(metricsViewName);
          }
        });

        if (metricsViewNames.size > 0) {
          const filterKey = getLookupKey(
            Array.from(metricsViewNames),
            filterName,
          );
          pinnedKeys.add(filterKey);
        }
      });

      this.pinnedFilterKeysStore.set(pinnedKeys);
      this.defaultPinnedFilterKeysStore.set(new Set(pinnedKeys));
    }

    // Update stores
    this.dimensionsForMetricsView.set(exhaustiveDimensions);
    this.measuresForMetricsView.set(exhaustiveMeasures);

    this.allMeasuresStore.set(mergedMeasures);

    this.allDimensionsStore.set(mergedDimensions);

    this.uniqueMetricsViewGroupingToDimensionArrayStore.set(
      createDimensionsForUniqueCombinationsOfMetricsViews(exhaustiveDimensions),
    );
  };

  getUIFiltersFromString = (filterString: string) => {
    const searchParams = new URLSearchParams(filterString);

    const parsedFilters: ParsedFilters[] = [];
    const legacyFilter = searchParams.get(ExploreStateURLParams.Filters);

    this.metricsViewFilters.forEach((filters, mvName) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
      const filterString = searchParams.get(paramKey) ?? legacyFilter ?? "";

      const parsed = filters.parseFilterString(filterString);
      parsedFilters.push(parsed);
    });

    return this.convertToUIFilters(
      parsedFilters,
      new Map(),
      get(this.pinnedFilterKeysStore),
      get(this.allMeasuresStore),
      get(this.allDimensionsStore),
    );
  };

  convertToUIFilters = (
    parsedFilters: ParsedFilters[],
    temporaryFilterKeys: Map<string, boolean>,
    pinnedFilters: Set<string>,
    allMeasures: MeasureLookup,
    allDimensions: DimensionLookup,
  ): UIFilters => {
    const parsedMap = new Map<string, ParsedFilters>(
      parsedFilters.map((p) => [p.metricsViewName, p]),
    );

    // Used for sorting
    const fullFilterString = parsedFilters
      .map((p) => p.urlFormat)
      .join(" AND ");

    const merged = {
      dimensionFilters: new Map<string, DimensionFilterItem>(),
      measureFilters: new Map<string, MeasureFilterItem>(),
      complexFilters: [],
      hasFilters: false,
      hasClearableFilters: false,
    };

    const metricsViewCombinationToDimensionMap = get(
      this.uniqueMetricsViewGroupingToDimensionArrayStore,
    );

    allMeasures.forEach((measureMap, key) => {
      const filters: MeasureFilterItem[] = [];

      const pinned = pinnedFilters.has(key);
      const temporary = temporaryFilterKeys.has(key);

      const metricsViewNames = measureMap ? Array.from(measureMap.keys()) : [];
      const measureSpecs = Array.from(measureMap?.values() || []);

      measureMap.forEach((measure, metricsViewName) => {
        const parsed = parsedMap.get(metricsViewName);
        if (!parsed) return;

        const [metricsViewGroup, measureName] = key.split(NAME_SEPARATOR);

        const measureFilter = parsed.measureFilters.get(measure.name as string);
        if (!measureFilter) {
          if (pinned || temporary) {
            filters.push({
              dimensionName: "",
              dimensions:
                metricsViewCombinationToDimensionMap.get(metricsViewGroup),
              name: measureName,
              label: measureSpecs[0].displayName ?? "",
              pinned: pinned,
              measures: measureMap,
              metricsViewNames: metricsViewNames,
            });
          }
        } else {
          if (pinned) {
            measureFilter.pinned = true;
          } else {
            measureFilter.pinned = false;
          }
          filters.push({
            ...measureFilter,
            dimensions:
              metricsViewCombinationToDimensionMap.get(metricsViewGroup),
          });
        }
      });

      if (filters.length === 0) return;

      merged.measureFilters.set(key, {
        ...filters[0],
        measures: measureMap,
      });
    });

    // can improve efficiency at a later date - bgh
    // iterate through all the unique dimension keys
    allDimensions.forEach((dimensionMap, key) => {
      const filters: DimensionFilterItem[] = [];

      const firstDimension = Array.from(dimensionMap.values())[0];

      const pinned = pinnedFilters.has(key);
      const temporary = temporaryFilterKeys.has(key);

      // iterate through the merged dimensions under this unique key
      dimensionMap.forEach((dimension, metricsViewName) => {
        const parsed = parsedMap.get(metricsViewName);

        if (!parsed) return;

        const dimFilter = parsed.dimensionFilters.get(dimension.name as string);

        if (!dimFilter) {
          if (pinned || temporary) {
            const tempData = {
              name: firstDimension.name || "",
              label: getDimensionDisplayName(firstDimension),
              mode: DimensionFilterMode.Select,
              selectedValues: [],
              dimensions: dimensionMap,
              isInclude: true,
              inputText: undefined,
              pinned: pinned,
            };

            filters.push(tempData);
          }
        } else {
          if (pinned) {
            dimFilter.pinned = true;
          } else {
            dimFilter.pinned = false;
          }
          filters.push(dimFilter);
        }
      });

      if (filters.length === 0) return;
      if (
        filters.every(
          (f) =>
            f.isInclude === filters[0].isInclude && f.mode === filters[0].mode,
        )
      ) {
        merged.dimensionFilters.set(key, {
          ...filters[0],
          dimensions: dimensionMap,
        });
      } else {
        // mixed filters - need to resolve
      }
    });

    merged.hasClearableFilters = parsedFilters.some(
      (p) => p.dimensionFilters.size > 0 || p.measureFilters.size > 0,
    );

    merged.hasFilters =
      merged.hasClearableFilters ||
      pinnedFilters.size > 0 ||
      temporaryFilterKeys.size > 0;

    // Sorting to ensure that pills don't jump around unnecessarily in the UI
    // Can be optimized - bgh
    const sortedDimensionMap = new Map(
      Array.from(merged.dimensionFilters.entries()).sort((a, b) => {
        return sortMeasuresOrDimensions(
          a[0],
          b[0],
          Array.from(pinnedFilters),
          Array.from(temporaryFilterKeys.keys()),
          fullFilterString,
        );
      }),
    );

    const sortedMeasureMap = new Map(
      Array.from(merged.measureFilters.entries()).sort((a, b) => {
        return sortMeasuresOrDimensions(
          a[0],
          b[0],
          Array.from(pinnedFilters),
          Array.from(temporaryFilterKeys.keys()),
          fullFilterString,
        );
      }),
    );

    merged.measureFilters = sortedMeasureMap;
    merged.dimensionFilters = sortedDimensionMap;

    return merged;
  };

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
      this.checkPinnedFilter(dimensionName, metricsViewNames);

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
      this.temporaryFilterKeysStore.update((tempFilters) => {
        tempFilters.set(measureOrDimensionKey, true);
        return tempFilters;
      });

      // Boolean controls whether the filter pill should open the dropdown automatically
      // This removes the flag after 200ms
      setTimeout(() => {
        this.temporaryFilterKeysStore.update((tempFilters) => {
          tempFilters.set(measureOrDimensionKey, false);
          return tempFilters;
        });
      }, 200);
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
      oldDimension: string,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(filter.measure, metricsViewNames);

      const newFilters = new Map<string, string | null>();

      metricsViewNames.forEach((name) => {
        const filterClass = this.metricsViewFilters.get(name);

        if (!filterClass) return;

        const string = filterClass.setMeasureFilter(
          dimensionName,
          filter,
          oldDimension,
        );

        newFilters.set(name, string || null);
      });

      await this.applyFiltersToUrl(newFilters);
    },
    removeMeasureFilter: async (
      dimensionName: string,
      measureName: string,
      metricsViewNames: string[],
    ) => {
      this.checkTemporaryFilter(measureName, metricsViewNames);
      this.checkPinnedFilter(measureName, metricsViewNames);

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
    toggleFilterPin: (name: string, metricsViewNames: string[]) => {
      this.pinnedFilterKeysStore.update((pinned) => {
        const key = getLookupKey(metricsViewNames, name);
        const deleted = pinned.delete(key);

        if (!deleted) {
          pinned.add(key);
        }

        this.temporaryFilterKeysStore.update((tempFilters) => {
          if (deleted) {
            tempFilters.set(key, false);
          } else {
            tempFilters.delete(key);
          }

          return tempFilters;
        });
        return pinned;
      });
    },
  };

  checkTemporaryFilter = (
    measureOrDimensionName: string,
    metricsViewNames: string[],
  ) => {
    const key = getLookupKey(metricsViewNames, measureOrDimensionName);
    const tempFilters = get(this.temporaryFilterKeysStore);

    const deleted = tempFilters.delete(key);
    if (deleted) {
      this.temporaryFilterKeysStore.set(tempFilters);
    }
  };

  checkPinnedFilter = (
    measureOrDimensionName: string,
    metricsViewNames: string[],
  ) => {
    const key = getLookupKey(metricsViewNames, measureOrDimensionName);
    const pinnedFilters = get(this.pinnedFilterKeysStore);

    const deleted = pinnedFilters.delete(key);
    if (deleted) {
      this.pinnedFilterKeysStore.set(pinnedFilters);
    }
  };

  // Unclear on what this actually should do - bgh
  // Go to defaults or truly clear all filters?
  clearAllFilters = async () => {
    this.temporaryFilterKeysStore.set(new Map());
    const existingParams = new URLSearchParams(window.location.search);
    const filterParamsToDelete = Array.from(existingParams.keys()).filter(
      (key) => key.startsWith(ExploreStateURLParams.Filters),
    );
    filterParamsToDelete.forEach((key) => {
      existingParams.delete(key);
    });
    const string = existingParams.toString();
    if (string) {
      await goto(`?${string}`);
      return;
    } else {
      await goto(`?clear=true`);
    }
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

/**
 * Sorts filter items with the following priority:
 * 1. Pinned items, following the order of the yaml
 * 2. Regular filter items, following their appearance in the full filter string
 * 3. Temporary items
 */
function sortMeasuresOrDimensions(
  aKey: string,
  bKey: string,
  pinnedFilters: string[],
  temporaryFilterKeys: string[],
  fullFilterString: string,
): number {
  const isAPinned = pinnedFilters.includes(aKey);
  const isBPinned = pinnedFilters.includes(bKey);
  const isATemporary = temporaryFilterKeys.includes(aKey);
  const isBTemporary = temporaryFilterKeys.includes(bKey);

  if (isAPinned && isBPinned) {
    return pinnedFilters.indexOf(aKey) - pinnedFilters.indexOf(bKey);
  }
  if (isAPinned !== isBPinned) {
    return isAPinned ? -1 : 1;
  }

  if (isATemporary && isBTemporary) {
    return (
      temporaryFilterKeys.indexOf(aKey) - temporaryFilterKeys.indexOf(bKey)
    );
  }
  if (isATemporary !== isBTemporary) {
    return isATemporary ? 1 : -1;
  }

  const aName = aKey.split(NAME_SEPARATOR)[1] || aKey;
  const bName = bKey.split(NAME_SEPARATOR)[1] || bKey;

  const aIndex = fullFilterString.indexOf(aName);
  const bIndex = fullFilterString.indexOf(bName);

  return aIndex - bIndex;
}

// This should be deprecated eventually in favor of better support for variously formatted expressions
export function flattenExpression(
  expression: V1Expression | undefined,
): V1Expression {
  if (!expression) {
    return createAndExpression([]);
  }

  let root: V1Expression;

  // Ensure top level is an OPERATION_AND
  if (!expression.cond || expression.cond.op !== V1Operation.OPERATION_AND) {
    root = createAndExpression([expression]);
  } else {
    root = expression;
  }

  const rootCond = root.cond;
  if (
    !rootCond ||
    rootCond.op !== V1Operation.OPERATION_AND ||
    !Array.isArray(rootCond.exprs)
  ) {
    return root;
  }

  // Recursively flatten all nested ANDs, preserving order
  rootCond.exprs = flattenAndExprs(rootCond.exprs);

  return root;
}

function flattenAndExprs(exprs: V1Expression[]): V1Expression[] {
  const result: V1Expression[] = [];

  for (const expr of exprs) {
    const cond = expr.cond;
    if (
      cond &&
      cond.op === V1Operation.OPERATION_AND &&
      Array.isArray(cond.exprs)
    ) {
      // Inline children in order
      result.push(...flattenAndExprs(cond.exprs));
    } else {
      result.push(expr);
    }
  }

  return result;
}

// wip - bgh
function mergeFilters<T>(
  metricsViewItems: Map<string, Map<string, T>>,
  locations: Map<string, string[]>,
  mergeStrategy: "all" = "all",
): UniqueLookup<T> {
  const merged = new Map<LookupKey, Map<string, T>>();

  if (mergeStrategy === "all") {
    locations.forEach((mvNames, name) => {
      const key = getLookupKey(mvNames, name);
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

class StoreOfStores<T> {
  private store: Writable<Map<string, T>>;
  subscribe: Readable<Map<string, T>>["subscribe"];

  constructor() {
    this.store = writable<Map<string, T>>(new Map());
    this.subscribe = this.store.subscribe;
  }

  forEach = (
    callback: (value: T, key: string, map: Map<string, T>) => void,
  ) => {
    const map = get(this.store);
    map.forEach(callback);
  };

  get(key: string): T | undefined {
    const map = get(this.store);
    return map.get(key);
  }

  set(key: string, value: T) {
    this.store.update((map) => {
      const next = new Map(map);
      next.set(key, value);
      return next;
    });
  }
}

function getAllCombinations(items: string[]): string[][] {
  const results: string[][] = [];
  const n = items.length;

  function dfs(start: number, path: string[]) {
    if (path.length > 0) {
      results.push(structuredClone(path));
    }

    for (let i = start; i < n; i++) {
      path.push(items[i]);
      dfs(i + 1, path);
      path.pop();
    }
  }

  dfs(0, []);

  return results;
}

function createDimensionsForUniqueCombinationsOfMetricsViews(
  metricsViewMap: Map<
    MetricsViewName,
    Map<DimensionName, MetricsViewSpecDimension>
  >,
) {
  const metricsViewNames = Array.from(metricsViewMap.keys()).sort();

  const allSortedCombinationsOfMetricsViewNames: string[][] =
    getAllCombinations(metricsViewNames);

  const map = new Map<string, MetricsViewSpecDimension[]>();

  allSortedCombinationsOfMetricsViewNames.forEach((group) => {
    const dimensions: Map<string, MetricsViewSpecDimension> = new Map();
    const key = group.join(MV_NAME_SEPARATOR);

    group.forEach((mvName) => {
      const dims = metricsViewMap.get(mvName);

      if (dims) {
        dims.forEach((dim, dimName) => {
          dimensions.set(dimName, dim);
        });
      }
    });

    map.set(key, Array.from(dimensions.values()));
  });

  return map;
}

function updateNameToMetricsViews<Name, Spec>(
  specs: Spec[] | undefined,
  getName: (spec: Spec) => Name | undefined,
  nameToMetricsViewNames: Map<Name, MetricsViewName[]>,
  nameToSpecForView: Map<Name, Spec>,
  metricsViewName: MetricsViewName,
) {
  specs?.forEach((spec) => {
    const name = getName(spec);
    if (!name) return;

    const existing = nameToMetricsViewNames.get(name) ?? [];
    nameToMetricsViewNames.set(name, [...existing, metricsViewName]);
    nameToSpecForView.set(name, spec);
  });
}
