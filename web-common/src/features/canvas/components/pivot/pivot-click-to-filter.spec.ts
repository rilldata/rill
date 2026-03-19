import { describe, expect, it, vi } from "vitest";
import { get, writable, type Readable } from "svelte/store";
import { createPivotClickToFilter } from "./pivot-click-to-filter";
import type { PivotDataStoreConfig } from "@rilldata/web-common/features/dashboards/pivot/types";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import type { FilterManager } from "../../stores/filter-manager";

/**
 * Minimal stub for FilterManager; only the fields touched by the factory
 * are provided. Everything else is left as `undefined` / `never`.
 */
function stubFilterManager(): FilterManager {
  return {
    metricsViewFilters: new Map(),
    checkTemporaryFilter: vi.fn(),
    applyFiltersToUrl: vi.fn(),
  } as unknown as FilterManager;
}

function emptyConfig(): PivotDataStoreConfig {
  return {
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    measureNames: ["total"],
    isFlat: true,
  } as unknown as PivotDataStoreConfig;
}

describe("pivot-click-to-filter: clearActiveComponent", () => {
  it("should clear selfFilteredDimensions when activeComponent is set to null", () => {
    const activeComponent = writable<string | null>(null);
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const whereFilter = writable<V1Expression | undefined>(undefined);
    const pivotConfig = writable(emptyConfig());
    const pivotDataStore = writable({ data: null }) as any;

    const onBecomeActive = vi.fn();
    const onBecomeInactive = vi.fn();

    const result = createPivotClickToFilter({
      pivotConfig: pivotConfig as Readable<PivotDataStoreConfig>,
      pivotDataStore,
      filterManager: stubFilterManager(),
      metricsViewName: "mv1",
      componentId: "pivot-1",
      activeComponent,
      selfFilteredDimensions,
      whereFilterStore: whereFilter,
      onBecomeActive,
      onBecomeInactive,
    });

    // Simulate the pivot becoming active with some self-filtered dimensions
    activeComponent.set("pivot-1");
    selfFilteredDimensions.set(new Set(["country"]));
    onBecomeInactive.mockClear();
    onBecomeActive.mockClear();

    // Now simulate clearActiveComponent: set activeComponent to null
    activeComponent.set(null);

    // The self-filtered dimensions should be cleared
    expect(get(selfFilteredDimensions).size).toBe(0);

    // onBecomeInactive should have been called
    expect(onBecomeInactive).toHaveBeenCalled();

    result.destroy();
  });

  it("should clear selfFilteredDimensions when another component becomes active", () => {
    const activeComponent = writable<string | null>(null);
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const whereFilter = writable<V1Expression | undefined>(undefined);
    const pivotConfig = writable(emptyConfig());
    const pivotDataStore = writable({ data: null }) as any;

    const onBecomeInactive = vi.fn();

    const result = createPivotClickToFilter({
      pivotConfig: pivotConfig as Readable<PivotDataStoreConfig>,
      pivotDataStore,
      filterManager: stubFilterManager(),
      metricsViewName: "mv1",
      componentId: "pivot-1",
      activeComponent,
      selfFilteredDimensions,
      whereFilterStore: whereFilter,
      onBecomeInactive,
    });

    // Simulate active pivot with self-filtered dimensions
    activeComponent.set("pivot-1");
    selfFilteredDimensions.set(new Set(["country"]));
    onBecomeInactive.mockClear();

    // Another component becomes active
    activeComponent.set("pivot-2");

    expect(get(selfFilteredDimensions).size).toBe(0);
    expect(onBecomeInactive).toHaveBeenCalled();

    result.destroy();
  });

  it("should NOT clear selfFilteredDimensions when this component is set as active", () => {
    const activeComponent = writable<string | null>(null);
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const whereFilter = writable<V1Expression | undefined>(undefined);
    const pivotConfig = writable(emptyConfig());
    const pivotDataStore = writable({ data: null }) as any;

    const result = createPivotClickToFilter({
      pivotConfig: pivotConfig as Readable<PivotDataStoreConfig>,
      pivotDataStore,
      filterManager: stubFilterManager(),
      metricsViewName: "mv1",
      componentId: "pivot-1",
      activeComponent,
      selfFilteredDimensions,
      whereFilterStore: whereFilter,
    });

    // Simulate active pivot with self-filtered dimensions
    selfFilteredDimensions.set(new Set(["country"]));

    // Set this component as active
    activeComponent.set("pivot-1");

    // selfFilteredDimensions should remain unchanged
    expect(get(selfFilteredDimensions).size).toBe(1);
    expect(get(selfFilteredDimensions).has("country")).toBe(true);

    result.destroy();
  });
});
