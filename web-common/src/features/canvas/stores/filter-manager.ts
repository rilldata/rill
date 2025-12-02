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
import { MetricsViewFilter } from "./metrics-view-filter";
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

type Lookup<T> = Map<MetricsViewName, Map<DimensionName, T>>;

export type DimensionLookup = Lookup<MetricsViewSpecDimension>;
export type MeasureLookup = Lookup<MetricsViewSpecMeasure>;

export type ParsedFilters = ReturnType<typeof initFilterBase>;

export function initFilterBase() {
  return {
    where: createAndExpression([]),
    dimensionFilter: createAndExpression([]),
    string: "",
    metricsViewName: "",
    dimensionsWithInlistFilter: <string[]>[],
    dimensionThresholdFilters: <DimensionThresholdFilter[]>[],
    measureFilters: new Map<string, MeasureFilterItem>(),
    dimensionFilters: new Map<string, DimensionFilterItem>(),
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

// wip - bgh
export class FilterManager {
  metricsViewFilters = new StoreOfStores<MetricsViewFilter>();
  _pinnedFilterKeys = writable<Set<string>>(new Set());
  _defaultPinnedFilterKeys = writable<Set<string>>(new Set());
  _temporaryFilterKeys = writable<Set<string>>(new Set());
  _allDimensions = writable<DimensionLookup>(new Map());
  _allMeasures = writable<MeasureLookup>(new Map());
  _dimensionFilterKeys: Readable<string[]>;
  _defaultUIFilters: Readable<UIFilters>;
  _defaultExpression: Readable<V1Expression>;
  // allMetricsViewNamesPrefix = writable<string>("");
  _activeUIFilters: Readable<UIFilters>;
  _activeExpression: Readable<V1Expression>;
  metricsViewNameDimensionMap: Map<
    string,
    Map<string, MetricsViewSpecDimension>
  > = new Map();
  metricsViewNameMeasureMap: Map<string, Map<string, MetricsViewSpecMeasure>> =
    new Map();
  _viewingDefaults: Readable<boolean>;
  _filterMap: Readable<Map<string, V1Expression>>;

  constructor(
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
    this.updateConfig(metricsViews, pinnedFilters, defaultFilters);

    this._defaultUIFilters = derived(
      [this.metricsViewFilters, this._defaultPinnedFilterKeys],
      ([metricsViewFilters, defaultPinnedFilterKeys], set) => {
        const stores = Array.from(metricsViewFilters.values()).map(
          (f) => f.parsedDefaultFilters,
        );
        derived(stores, (filters) => {
          return this.convertToUIFilters(
            filters,
            new Set(),
            defaultPinnedFilterKeys,
          );
        }).subscribe(set);
      },
    );

    this._activeUIFilters = derived(
      [
        this._pinnedFilterKeys,
        this._temporaryFilterKeys,
        this.metricsViewFilters,
      ],
      ([pinnedFilters, temporaryFilterKeys, metricsViewFilters], set) => {
        const stores = Array.from(metricsViewFilters.values()).map(
          (f) => f.parsed,
        );

        derived(stores, (filters) => {
          return this.convertToUIFilters(
            filters,
            temporaryFilterKeys,
            pinnedFilters,
          );
        }).subscribe(set);
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

    this._viewingDefaults = derived(
      [this._activeUIFilters, this._defaultUIFilters],
      ([active, defaults]) => {
        const activeDimensionKeys = Array.from(
          active.dimensionFilters.keys(),
        ).sort();
        const defaultDimensionKeys = Array.from(
          defaults.dimensionFilters.keys(),
        ).sort();

        const activeMeasureKeys = Array.from(
          active.measureFilters.keys(),
        ).sort();
        const defaultMeasureKeys = Array.from(
          defaults.measureFilters.keys(),
        ).sort();

        return (
          JSON.stringify(activeDimensionKeys) ===
            JSON.stringify(defaultDimensionKeys) &&
          JSON.stringify(activeMeasureKeys) ===
            JSON.stringify(defaultMeasureKeys) &&
          !active.hasClearableFilters
        );
      },
    );
  }

  updateConfig(
    metricsViews: Record<string, V1MetricsView | undefined>,
    pinnedFilters?: string[],
    defaultFilters?: V1CanvasPresetFilterExpr,
  ) {
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
            new MetricsViewFilter(mv, name, defaultFilters?.[name], this),
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
      this._pinnedFilterKeys.set(keys);
      this._defaultPinnedFilterKeys.set(keys);
    }
  }

  convertToUIFilters = (
    parsedFilters: ParsedFilters[],
    temporaryFilterKeys: Set<string>,
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

      const measureMap = allMeasures.get(key);
      const metricsViewNames = measureMap ? Array.from(measureMap.keys()) : [];

      const measureSpecs = Array.from(measureMap?.values() || []);

      // Needs work - bgh
      if (temporaryFilterKeys.has(key)) {
        merged.measureFilters.set(key, {
          dimensionName: "",
          dimensions: undefined,
          name: key.split("::")[1],
          label: measureSpecs[0].displayName ?? "",
          pinned: pinned,
          measures: measureMap,
          metricsViewNames: metricsViewNames,
        });
        return;
      }

      measures.forEach((measure, metricsViewName) => {
        const parsed = parsedMap.get(metricsViewName);
        if (!parsed) return;

        const dimFilter = parsed.measureFilters.get(measure.name as string);
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
          } else {
            dimFilter.pinned = false;
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
      merged.measureFilters.set(key, {
        ...filters[0],
        measures: measures,
      });
      // } else {
      //   // mixed filters - need to resolve
      // }
    });

    // can improve efficiency at a later date - bgh
    // iterate through all the unique dimension keys
    allDimensions.forEach((dimensions, key) => {
      const filters: DimensionFilterItem[] = [];

      const firstDimension = Array.from(dimensions.values())[0];

      const pinned = pinnedFilters.has(key);

      if (temporaryFilterKeys.has(key)) {
        filters.push({
          name: firstDimension.name || "",
          label: getDimensionDisplayName(firstDimension),
          mode: DimensionFilterMode.Select,
          selectedValues: [],
          dimensions: dimensions,
          isInclude: true,
          inputText: undefined,
          pinned: pinned,
        });

        merged.dimensionFilters.set(key, {
          ...filters[0],
          dimensions: dimensions,
        });
        return;
      }

      // iterate through the merged dimensions under this unique key
      dimensions.forEach((dimension, metricsViewName) => {
        const parsed = parsedMap.get(metricsViewName);
        if (!parsed) return;

        const dimFilter = parsed.dimensionFilters.get(dimension.name as string);
        if (!dimFilter) {
          if (pinned) {
            filters.push({
              name: firstDimension.name || "",
              label: getDimensionDisplayName(firstDimension),
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
          a,
          b,
          pinnedFilters,
          temporaryFilterKeys,
          fullFilterString,
        );
      }),
    );

    const sortedMeasureMap = new Map(
      Array.from(merged.measureFilters.entries()).sort((a, b) => {
        return sortMeasuresOrDimensions(
          a,
          b,
          pinnedFilters,
          temporaryFilterKeys,
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
      this.checkTemporaryFilter(filter.measure, metricsViewNames);

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
        const deleted = pinned.delete(key);

        if (!deleted) {
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
    this._temporaryFilterKeys.set(new Set());
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

function sortMeasuresOrDimensions(
  a: [string, MeasureFilterItem | DimensionFilterItem],
  b: [string, MeasureFilterItem | DimensionFilterItem],
  pinnedFilters: Set<string>,
  temporaryFilterKeys: Set<string>,
  fullFilterString: string,
) {
  const aKey = a[0];
  const bKey = b[0];

  const aPinned = pinnedFilters.has(aKey) ? 1 : 0;
  const bPinned = pinnedFilters.has(bKey) ? 1 : 0;

  if (aPinned !== bPinned) {
    return bPinned - aPinned;
  }

  const aTemporary = temporaryFilterKeys.has(aKey) ? 1 : 0;
  const bTemporary = temporaryFilterKeys.has(bKey) ? 1 : 0;

  if (aTemporary !== bTemporary) {
    return aTemporary - bTemporary;
  }

  const aName = aKey.split("::")[1];
  const bName = bKey.split("::")[1];

  const aIndex = fullFilterString.indexOf(aName);
  const bIndex = fullFilterString.indexOf(bName);

  return aIndex - bIndex;
}
