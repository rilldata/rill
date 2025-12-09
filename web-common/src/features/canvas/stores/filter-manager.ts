import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { type DimensionFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimension-filters";
import type { MeasureFilterItem } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  MetricsViewSpecDimension,
  V1CanvasPresetFilterExpr,
  V1Expression,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import { type MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import {
  derived,
  get,
  type Readable,
  type Writable,
  writable,
} from "svelte/store";
import { ExploreStateURLParams } from "../../dashboards/url-state/url-params";
import { goto } from "$app/navigation";
import { FilterState } from "./metrics-view-filter";
import { getDimensionDisplayName } from "../../dashboards/filters/getDisplayName";

export type UIFilters = {
  dimensionFilters: Map<string, DimensionFilterItem>;
  measureFilters: Map<string, MeasureFilterItem>;
  complexFilters: V1Expression[];
  hasFilters: boolean;
  hasClearableFilters: boolean;
};

type MetricsViewName = string;
type DimensionName = string;
type MeasureName = string;

type Lookup<T> = Map<MetricsViewName, Map<DimensionName | MeasureName, T>>;

export type DimensionLookup = Lookup<MetricsViewSpecDimension>;
export type MeasureLookup = Lookup<MetricsViewSpecMeasure>;

export type ParsedFilters = ReturnType<typeof initFilterBase>;

export function initFilterBase(metricsViewName: string) {
  return {
    where: createAndExpression([]),
    dimensionFilter: createAndExpression([]),
    string: "",
    metricsSQL: "",
    metricsViewName,
    dimensionsWithInlistFilter: <string[]>[],
    dimensionThresholdFilters: <DimensionThresholdFilter[]>[],
    measureFilters: new Map<string, MeasureFilterItem>(),
    dimensionFilters: new Map<string, DimensionFilterItem>(),
    complexFilters: [] as V1Expression[],
    hasFilters: false,
    hasClearableFilters: false,
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
      map.set(key, value);
      return map;
    });
  }
}

export class FilterManager {
  metricsViewFilters = new StoreOfStores<FilterState>();
  _pinnedFilterKeys = writable<Set<string>>(new Set());
  _defaultPinnedFilterKeys = writable<Set<string>>(new Set());
  _temporaryFilterKeys = writable<Map<string, boolean>>(new Map());
  _allDimensions = writable<DimensionLookup>(new Map());
  _allMeasures = writable<MeasureLookup>(new Map());
  _dimensionFilterKeys: Readable<string[]>;
  defaultUIFiltersStore: Readable<UIFilters>;
  _defaultExpression: Readable<V1Expression>;
  _activeUIFilters: Readable<UIFilters>;
  _activeExpression: Readable<V1Expression>;
  metricsViewNameDimensionMap: Map<
    MetricsViewName,
    Map<DimensionName, MetricsViewSpecDimension>
  > = new Map();
  metricsViewNameMeasureMap: Map<
    MetricsViewName,
    Map<MeasureName, MetricsViewSpecMeasure>
  > = new Map();
  _filterMap: Readable<Map<string, V1Expression>>;
  _scopedDimensions = writable<Map<string, DimensionLookup>>(new Map());
  _scopedMeasures = writable<Map<string, MeasureLookup>>(new Map());

  constructor(
    metricsViews: Record<string, V1MetricsView | undefined>,
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
          [this._defaultPinnedFilterKeys, ...stores],
          ([defaultPinnedFilterKeys, ...filters]) => {
            return this.convertToUIFilters(
              filters,
              new Map(),
              defaultPinnedFilterKeys,
            );
          },
        ).subscribe(set);
      },
    );

    this._activeUIFilters = derived(
      [this.metricsViewFilters],
      ([metricsViewFilters], set) => {
        const stores = Array.from(metricsViewFilters.values()).map(
          (f) => f.parsed,
        );

        derived(
          [this._pinnedFilterKeys, this._temporaryFilterKeys, ...stores],
          ([pinnedFilters, temporaryFilterKeys, ...filters]) => {
            return this.convertToUIFilters(
              filters,
              temporaryFilterKeys,
              pinnedFilters,
            );
          },
        ).subscribe(set);
      },
    );

    this._filterMap = derived(
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
    return new FilterState(metricsViewName, this);
  };

  onUrlChange = (searchParams: URLSearchParams) => {
    const legacyFilter = searchParams.get(ExploreStateURLParams.Filters);

    this.metricsViewFilters.forEach((filters, mvName) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
      const filterString = searchParams.get(paramKey) ?? legacyFilter ?? "";

      filters.onFilterStringChange(filterString);
    });
  };

  updateConfig(
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
    const dimensionNameToMetricsViewNames: Map<
      DimensionName,
      MetricsViewName[]
    > = new Map();
    const measureNameToMetricsViewNames: Map<MeasureName, MetricsViewName[]> =
      new Map();
    const dimensionLookups: Map<string, DimensionLookup> = new Map();
    const measureLookups: Map<string, MeasureLookup> = new Map();

    Object.entries(metricsViews).forEach(([metricsViewName, metricsView]) => {
      const dimensionNameToDimension: DimensionLookup = new Map();
      const measureMap: MeasureLookup = new Map();

      if (metricsView) {
        this.metricsViewNameDimensionMap.set(metricsViewName, new Map());
        this.metricsViewNameMeasureMap.set(metricsViewName, new Map());

        const { measures, dimensions } = metricsView.state?.validSpec || {};

        dimensions?.forEach((dim) => {
          const dimName = dim.name;
          if (!dimName || dim.type === "DIMENSION_TYPE_TIME") return;

          dimensionNameToDimension.set(
            dimName,
            new Map([[metricsViewName, dim]]),
          );

          const existingMetricsViews =
            dimensionNameToMetricsViewNames.get(dimName) || [];
          existingMetricsViews.push(metricsViewName);
          dimensionNameToMetricsViewNames.set(dimName, existingMetricsViews);

          this.metricsViewNameDimensionMap
            .get(metricsViewName)
            ?.set(dimName, dim);
        });

        measures?.forEach((measure) => {
          const measureName = measure.name;
          if (!measureName) return;

          measureMap.set(measureName, new Map([[metricsViewName, measure]]));

          const existingMetricsViews =
            measureNameToMetricsViewNames.get(measureName) || [];
          existingMetricsViews.push(metricsViewName);
          measureNameToMetricsViewNames.set(measureName, existingMetricsViews);

          this.metricsViewNameMeasureMap
            .get(metricsViewName)
            ?.set(measureName, measure);
        });

        dimensionLookups.set(metricsViewName, dimensionNameToDimension);
        measureLookups.set(metricsViewName, measureMap);

        let existingFilterStore = this.metricsViewFilters.get(metricsViewName);

        if (!existingFilterStore) {
          existingFilterStore = new FilterState(metricsViewName, this);
          this.metricsViewFilters.set(metricsViewName, existingFilterStore);
        }

        existingFilterStore.onDefaultExpressionChange(
          defaultFilters?.[metricsViewName],
        );
      }
    });

    const mergedDimensions = mergeFilters(
      this.metricsViewNameDimensionMap,
      dimensionNameToMetricsViewNames,
      "all",
    );

    const mergedMeasures = mergeFilters(
      this.metricsViewNameMeasureMap,
      measureNameToMetricsViewNames,
      "all",
    );

    this._allMeasures.set(mergedMeasures);

    this._allDimensions.set(mergedDimensions);

    this._scopedDimensions.set(dimensionLookups);
    this._scopedMeasures.set(measureLookups);

    if (pinnedFilters) {
      const keys = new Set<string>();

      pinnedFilters.forEach((filterName) => {
        const foundDimensionsOrMeasures = new Map();

        this.metricsViewFilters.forEach((_, name) => {
          const foundDimension = this.metricsViewNameDimensionMap
            .get(name)
            ?.get(filterName);
          if (foundDimension) {
            foundDimensionsOrMeasures.set(name, foundDimension);
          }
          const foundMeasure = this.metricsViewNameMeasureMap
            .get(name)
            ?.get(filterName);
          if (foundMeasure) {
            foundDimensionsOrMeasures.set(name, foundMeasure);
          }
        });

        const filterKey = `${Array.from(foundDimensionsOrMeasures.keys()).join("//")}::${filterName}`;

        keys.add(filterKey);
      });

      this._pinnedFilterKeys.set(new Set(keys));
      this._defaultPinnedFilterKeys.set(new Set(keys));
    }
  }

  getUIFiltersFromString = (filterString: string): UIFilters => {
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
      get(this._pinnedFilterKeys),
    );
  };

  convertToUIFilters = (
    parsedFilters: ParsedFilters[],
    temporaryFilterKeys: Map<string, boolean>,
    pinnedFilters: Set<string>,
  ): UIFilters => {
    const parsedMap = new Map<string, ParsedFilters>();

    parsedFilters.forEach((parsed) => {
      parsedMap.set(parsed.metricsViewName, parsed);
    });

    const merged = {
      dimensionFilters: new Map<string, DimensionFilterItem>(),
      measureFilters: new Map<string, MeasureFilterItem>(),
      complexFilters: [],
      hasFilters: false,
      hasClearableFilters: false,
    };

    const allMeasures = get(this._allMeasures);
    const allDimensions = get(this._allDimensions);

    const fullFilterString = parsedFilters.map((p) => p.string).join(" AND ");

    allMeasures.forEach((measures, key) => {
      const filters: MeasureFilterItem[] = [];

      const pinned = pinnedFilters.has(key);
      const temporary = temporaryFilterKeys.has(key);

      const measureMap = allMeasures.get(key);
      const metricsViewNames = measureMap ? Array.from(measureMap.keys()) : [];

      const measureSpecs = Array.from(measureMap?.values() || []);

      measures.forEach((measure, metricsViewName) => {
        const parsed = parsedMap.get(metricsViewName);
        if (!parsed) return;

        const measureFilter = parsed.measureFilters.get(measure.name as string);
        if (!measureFilter) {
          if (pinned || temporary) {
            filters.push({
              dimensionName: "",
              dimensions: undefined,
              name: key.split("::")[1],
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
          filters.push(measureFilter);
        }
      });

      if (filters.length === 0) return;

      merged.measureFilters.set(key, {
        ...filters[0],
        measures: measures,
      });
    });

    // can improve efficiency at a later date - bgh
    // iterate through all the unique dimension keys
    allDimensions.forEach((dimensions, key) => {
      const filters: DimensionFilterItem[] = [];

      const firstDimension = Array.from(dimensions.values())[0];

      const pinned = pinnedFilters.has(key);
      const temporary = temporaryFilterKeys.has(key);

      // iterate through the merged dimensions under this unique key
      dimensions.forEach((dimension, metricsViewName) => {
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
              dimensions: dimensions,
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
          dimensions: dimensions,
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

    // Temporary to get sorting to work, will revisit later - bgh
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
        tempFilters.set(measureOrDimensionKey, true);
        return tempFilters;
      });

      // Boolean controls whether the filter pill should open the dropdown automatically
      // This removes the flag after 200ms
      setTimeout(() => {
        this._temporaryFilterKeys.update((tempFilters) => {
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
      this._pinnedFilterKeys.update((pinned) => {
        const key = metricsViewNames.sort().join("//") + "::" + name;
        const deleted = pinned.delete(key);

        if (!deleted) {
          pinned.add(key);
        }

        this._temporaryFilterKeys.update((tempFilters) => {
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
    this._temporaryFilterKeys.set(new Map());
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

  const aName = aKey.split("::")[1] || aKey;
  const bName = bKey.split("::")[1] || bKey;

  const aIndex = fullFilterString.indexOf(aName);
  const bIndex = fullFilterString.indexOf(bName);

  return aIndex - bIndex;
}
