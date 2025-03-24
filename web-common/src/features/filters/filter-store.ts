import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  getMeasureFilters,
  type MeasureFilterItem,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
import {
  createAndExpression,
  forEachExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import {
  type DimensionFilterDisplayEntry,
  type DimensionFilterEntry,
  DimensionFilterMode,
  mapExprToDimensionFilter,
} from "@rilldata/web-common/features/filters/dimension-filter";
import type { FilterSpecStore } from "@rilldata/web-common/features/filters/filter-spec-store";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import {
  derived,
  get,
  type Readable,
  writable,
  type Writable,
} from "svelte/store";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "../dashboards/filters/getDisplayName";

export class FilterStore {
  // -------------------
  // STORES (writable)
  // -------------------
  dimensionFilters: Writable<DimensionFilterEntry[]>;
  dimensionThresholdFilters: Writable<Array<DimensionThresholdFilter>>;
  temporaryFilterName: Writable<string | null>;
  // TODO: advanced filters

  // -------------------
  // "SELECTORS" (readable/derived)
  // -------------------
  measureHasFilter: Readable<(measureName: string) => boolean>;
  getAllMeasureFilterItems: Readable<MeasureFilterItem[]>;
  private getMeasureFilterItems: Readable<MeasureFilterItem[]>;

  getAllDimensionFilterItems: Readable<DimensionFilterDisplayEntry[]>;
  private getDimensionFilterItems: Readable<DimensionFilterDisplayEntry[]>;

  dimensionHasFilter: Readable<(dimName: string) => boolean>;
  selectedDimensionValues: Readable<(dimName: string) => unknown[]>;
  getDimensionFilterItem: Readable<
    (dimName: string) => DimensionFilterEntry | undefined
  >;
  unselectedDimensionValues: Readable<
    (dimensionName: string, values: unknown[]) => unknown[]
  >;
  includedDimensionValues: Readable<(dimensionName: string) => unknown[]>;
  hasAtLeastOneDimensionFilter: Readable<() => boolean>;

  hasFilters: Readable<boolean>;

  public constructor(public readonly specStore: FilterSpecStore) {
    // -----------------------------
    // Initialize writable stores
    // -----------------------------
    this.dimensionFilters = writable([]);
    this.dimensionThresholdFilters = writable([]);
    this.temporaryFilterName = writable(null);

    // -------------------------------
    // MEASURE SELECTORS
    // -------------------------------
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

    this.getMeasureFilterItems = derived(
      [this.specStore.measures, this.dimensionThresholdFilters],
      ([measuresSpecs, dimensionThresholdFilters]) => {
        // TODO: add dimensions
        return getMeasureFilters(
          getMapFromArray(measuresSpecs, (m) => m.name!),
          dimensionThresholdFilters,
        );
      },
    );

    this.getAllMeasureFilterItems = derived(
      [
        this.specStore.measures,
        this.getMeasureFilterItems,
        this.temporaryFilterName,
      ],
      ([measuresSpecs, measureFilterItems, tempFilter]) => {
        if (!tempFilter) return measureFilterItems;
        const measureSpec = measuresSpecs.find((m) => m.name === tempFilter);
        if (!measureSpec) return measureFilterItems;

        const itemsCopy = [...measureFilterItems];
        itemsCopy.push({
          dimensionName: "",
          name: tempFilter,
          label: getMeasureDisplayName(measureSpec),
          // TODO
          // dimensions: dimensions,
        });
        return itemsCopy;
      },
    );

    // -------------------------------
    // DIMENSION SELECTORS
    // -------------------------------
    this.dimensionHasFilter = derived(
      this.dimensionFilters,
      (dimensionFilters) => {
        return (dimName) => !!dimensionFilters.find((d) => d.name === dimName);
      },
    );

    this.getDimensionFilterItems = derived(
      [this.specStore.dimensions, this.dimensionFilters],
      ([dimensionSpecs, dimensionFilters]) =>
        dimensionFilters
          .map((dfe) => {
            const dimension = dimensionSpecs.find((d) => d.name === dfe.name);
            if (!dimension) return undefined;
            return <DimensionFilterDisplayEntry>{
              ...dfe,
              label: "",
            };
          })
          .filter(Boolean) as DimensionFilterDisplayEntry[],
    );

    this.getAllDimensionFilterItems = derived(
      [
        this.specStore.dimensions,
        this.getDimensionFilterItems,
        this.temporaryFilterName,
      ],
      ([dimensionSpecs, dimensionFilters, tempFilter]) => {
        if (!tempFilter) return dimensionFilters;
        const dimensionSpec = dimensionSpecs.find((m) => m.name === tempFilter);
        if (!dimensionSpec) return dimensionFilters;

        const itemsCopy = [...dimensionFilters];
        itemsCopy.push({
          name: tempFilter,
          label: getDimensionDisplayName(dimensionSpec),
          mode: DimensionFilterMode.Select,
          values: [],
          exclude: false,
        });
        return itemsCopy;
      },
    );

    this.selectedDimensionValues = derived(
      this.dimensionFilters,
      (dimensionFilters) => {
        return (dimName: string) => {
          const dimensionFilter = dimensionFilters.find(
            (d) => d.name === dimName,
          );
          if (!dimensionFilter) return [];

          if (dimensionFilter.mode === DimensionFilterMode.Contains) return []; // TODO

          return [...new Set(dimensionFilter.values)];
        };
      },
    );

    this.getDimensionFilterItem = derived(
      this.dimensionFilters,
      (dimensionFilters) => {
        return (dimName: string) =>
          dimensionFilters.find((d) => d.name === dimName);
      },
    );

    this.unselectedDimensionValues = derived(
      this.dimensionFilters,
      (dimensionFilters) => {
        return (dimensionName: string, values: unknown[]) => {
          const dimensionFilter = dimensionFilters.find(
            (df) => df.name === dimensionName,
          );
          if (
            !dimensionFilter ||
            dimensionFilter.mode === DimensionFilterMode.Contains
          ) {
            return values;
          }
          return values.filter((v) => !dimensionFilter.values.includes(v));
        };
      },
    );

    this.includedDimensionValues = derived(
      this.dimensionFilters,
      (dimensionFilters) => {
        return (dimensionName: string) => {
          const dimensionFilter = dimensionFilters.find(
            (df) => df.name === dimensionName,
          );
          if (
            !dimensionFilter ||
            dimensionFilter.exclude ||
            dimensionFilter.mode === DimensionFilterMode.Contains
          ) {
            return [];
          }
          return dimensionFilter.values;
        };
      },
    );

    this.hasAtLeastOneDimensionFilter = derived(
      this.dimensionFilters,
      (dimensionFilters) => () => dimensionFilters.length > 0,
    );

    this.hasFilters = derived(
      [this.dimensionFilters, this.dimensionThresholdFilters],
      ([dimensionFilters, dimensionThresholdFilters]) =>
        dimensionFilters.length > 0 || dimensionThresholdFilters.length > 0,
    );
  }

  // -------------------
  // LOADERS
  // -------------------

  public loadFromFilter(filter: string) {
    this.loadFromExpression(convertFilterParamToExpression(filter));
  }

  public loadFromExpression(expr: V1Expression) {
    const {
      dimensionFilters: dimensionFiltersExpr,
      dimensionThresholdFilters,
    } = splitWhereFilter(expr);
    const dimensionFilters: DimensionFilterEntry[] = [];
    forEachExpression(dimensionFiltersExpr, (e) => {
      dimensionFilters.push(mapExprToDimensionFilter(e)!);
    });
    this.dimensionThresholdFilters.set(dimensionThresholdFilters);
  }

  // --------------------
  // ACTIONS / MUTATORS
  // --------------------

  setMeasureFilter = (dimensionName: string, filter: MeasureFilterEntry) => {
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

  removeMeasureFilter = (dimensionName: string, measureName: string) => {
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
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter !== null) {
      this.temporaryFilterName.set(null);
    }
    const dimensionFilters = get(this.dimensionFilters);

    const filterIndex = this.getDimensionFilterIndex(dimensionName);

    if (filterIndex === -1) {
      dimensionFilters.push({
        name: dimensionName,
        mode: DimensionFilterMode.Select,
        values: [dimensionValue],
        exclude: false,
      });
      this.dimensionFilters.set(dimensionFilters);
      return;
    }

    const dimensionFilter = dimensionFilters[filterIndex];
    if (dimensionFilter.mode === DimensionFilterMode.Contains) return; // TODO

    const valueIndex = dimensionFilter.values.indexOf(dimensionValue);
    if (valueIndex === -1) {
      if (isExclusiveFilter) {
        dimensionFilter.values = [dimensionValue];
      } else {
        dimensionFilter.values.push(dimensionValue);
      }
    } else {
      dimensionFilter.values.splice(valueIndex, 1);
      if (dimensionFilter.values.length === 0) {
        dimensionFilters.splice(filterIndex, 1);
        if (keepPillVisible) {
          this.temporaryFilterName.set(dimensionName);
        }
      }
    }
    this.dimensionFilters.set(dimensionFilters);
  };

  toggleDimensionFilterMode = (dimensionName: string) => {
    const dimensionFilters = get(this.dimensionFilters);
    const filterIndex = this.getDimensionFilterIndex(dimensionName);
    if (filterIndex === -1) return;

    const dimensionFilter = dimensionFilters[filterIndex];
    dimensionFilter.exclude = !dimensionFilter.exclude;
    this.dimensionFilters.set(dimensionFilters);
  };

  removeDimensionFilter = (dimensionName: string) => {
    const tempFilter = get(this.temporaryFilterName);
    if (tempFilter === dimensionName) {
      this.temporaryFilterName.set(null);
      return;
    }

    const dimensionFilters = get(this.dimensionFilters);
    const filterIndex = this.getDimensionFilterIndex(dimensionName);
    if (filterIndex === -1) return;

    dimensionFilters.splice(filterIndex, 1);
    this.dimensionFilters.set(dimensionFilters);
  };

  clearAllFilters = () => {
    const dfs = get(this.dimensionFilters);
    const dtfs = get(this.dimensionThresholdFilters);
    const hasFilters = dfs.length || dtfs.length;
    if (!hasFilters) return;
    this.dimensionFilters.set([]);
    this.dimensionThresholdFilters.set([]);
    this.temporaryFilterName.set(null);
  };

  getFiltersFromText = (filterText: string) => {
    let expr = convertFilterParamToExpression(filterText);
    if (
      expr?.cond?.op !== V1Operation.OPERATION_AND &&
      expr?.cond?.op !== V1Operation.OPERATION_OR
    ) {
      expr = createAndExpression([expr]);
    }
    return expr;
  };

  setTemporaryFilterName = (name: string) => {
    this.temporaryFilterName.set(name);
  };

  private getDimensionFilterIndex(dimName: string) {
    return get(this.dimensionFilters).findIndex((df) => df.name === dimName);
  }
}
