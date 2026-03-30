import { getFiltersFromRow } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotDataRow,
  PivotDataStore,
  PivotDataStoreConfig,
  PivotFilter,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { get, writable, type Readable } from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import { dimKeyFromRow } from "../../../dashboards/pivot/pivot-click-selection";
import type { FilterManager } from "../../stores/filter-manager";
import { createPivotClickToFilter } from "./pivot-click-to-filter";

// Partial mocks: override only the filter-extraction functions while keeping
// the rest of each module's real exports (extractDimensionFiltersFromExpression,
// getActiveDimensionNames, etc.) which the factory depends on.
vi.mock(
  "@rilldata/web-common/features/dashboards/pivot/pivot-utils",
  async () => ({
    ...(await vi.importActual(
      "@rilldata/web-common/features/dashboards/pivot/pivot-utils",
    )),
    getFiltersFromRow: vi.fn(() => ({ filters: undefined, timeRange: null })),
  }),
);

vi.mock(
  "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection",
  async () => ({
    ...(await vi.importActual(
      "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection",
    )),
    getFiltersForRowData: vi.fn(() => ({
      filters: undefined,
      timeRange: null,
    })),
  }),
);

/** Build a PivotFilter with no time range (sufficient for these tests). */
function makePivotFilter(dimensionName: string, values: string[]): PivotFilter {
  return {
    filters: createAndExpression([createInExpression(dimensionName, values)]),
    timeRange: { start: undefined, end: undefined },
  };
}

/**
 * Stub for FilterManager. The factory only accesses `metricsViewFilters`
 * (via `.get`/`.set`), `checkTemporaryFilter`, and `applyFiltersToUrl`.
 * A plain Map structurally satisfies the `.get`/`.set` calls at runtime.
 */
function stubFilterManager() {
  return {
    metricsViewFilters: new Map<
      string,
      {
        addDimensionValueSelections: ReturnType<typeof vi.fn>;
        toggleDimensionValueSelections: ReturnType<typeof vi.fn>;
      }
    >(),
    checkTemporaryFilter: vi.fn(),
    applyFiltersToUrl: vi.fn(),
  } as unknown as FilterManager;
}

/** Create a FilterManager with a working filterClass stub for a metrics view */
function stubFilterManagerWithClass(metricsViewName: string) {
  const fm = stubFilterManager();
  const filterClass = {
    addDimensionValueSelections: vi.fn(() => "filter-string"),
    toggleDimensionValueSelections: vi.fn(() => "filter-string"),
  };
  (fm.metricsViewFilters as unknown as Map<string, typeof filterClass>).set(
    metricsViewName,
    filterClass,
  );
  return { fm, filterClass };
}

/** Stub config with only the fields the factory reads. */
function emptyConfig() {
  return {
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    measureNames: ["total"],
    isFlat: true,
  } as unknown as PivotDataStoreConfig;
}

function flatConfigTwoDims() {
  return {
    rowDimensionNames: ["country", "city"],
    colDimensionNames: [],
    measureNames: ["revenue"],
    isFlat: true,
  } as unknown as PivotDataStoreConfig;
}

function nestedConfig() {
  return {
    rowDimensionNames: ["country"],
    colDimensionNames: [],
    measureNames: ["revenue"],
    isFlat: false,
  } as unknown as PivotDataStoreConfig;
}

/** Build a PivotDataStore with the given data rows and optional axes. */
function stubPivotDataStore(
  data: PivotDataRow[],
  columnDimensionAxes: Record<string, string[]> = {},
): PivotDataStore {
  return writable({
    isFetching: false,
    data,
    columnDef: [],
    assembled: true,
    totalColumns: 0,
    columnDimensionAxes,
  });
}

/**
 * Build the args object for createPivotClickToFilter with sensible defaults.
 * Callers override only what they need.
 */
function createFactoryArgs(
  overrides: Partial<Parameters<typeof createPivotClickToFilter>[0]> = {},
): Parameters<typeof createPivotClickToFilter>[0] {
  return {
    pivotConfig: writable(emptyConfig()) as Readable<PivotDataStoreConfig>,
    pivotDataStore: stubPivotDataStore([]),
    filterManager: stubFilterManager(),
    metricsViewName: "mv1",
    componentId: "pivot-1",
    activeComponent: writable<string | null>(null),
    selfFilteredDimensions: writable<Set<string>>(new Set()),
    whereFilterStore: writable<V1Expression | undefined>(undefined),
    ...overrides,
  };
}

describe("pivot-click-to-filter: clearActiveComponent", () => {
  it("should clear selfFilteredDimensions when activeComponent is set to null", () => {
    const activeComponent = writable<string | null>(null);
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const onBecomeActive = vi.fn();
    const onBecomeInactive = vi.fn();

    const result = createPivotClickToFilter(
      createFactoryArgs({
        activeComponent,
        selfFilteredDimensions,
        onBecomeActive,
        onBecomeInactive,
      }),
    );

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
    const onBecomeInactive = vi.fn();

    const result = createPivotClickToFilter(
      createFactoryArgs({
        activeComponent,
        selfFilteredDimensions,
        onBecomeInactive,
      }),
    );

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

    const result = createPivotClickToFilter(
      createFactoryArgs({
        activeComponent,
        selfFilteredDimensions,
      }),
    );

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

describe("pivot-click-to-filter: flat table single-cell-per-row", () => {
  /** Flat table data: two rows, each with country and city dimensions */
  const flatTableData: PivotDataRow[] = [
    { country: "US", city: "NYC", revenue: 100 },
    { country: "UK", city: "London", revenue: 200 },
  ];

  const row0 = flatTableData[0];
  const row1 = flatTableData[1];
  const row0DimKey = dimKeyFromRow(row0, ["country", "city"]);
  const row1DimKey = dimKeyFromRow(row1, ["country", "city"]);

  function setupFlat() {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    // Configure getFiltersFromRow mock to return appropriate filters per column
    vi.mocked(getFiltersFromRow).mockImplementation(
      (_config, _rowData, colId) => {
        if (colId === "country") return makePivotFilter("country", ["US"]);
        if (colId === "city") return makePivotFilter("city", ["NYC"]);
        return makePivotFilter("country", ["US"]);
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(
          flatConfigTwoDims(),
        ) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(flatTableData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    return { result, filterClass, selfFilteredDimensions, fm };
  }

  it("should replace existing cell in the same row for flat tables", () => {
    const { result, filterClass } = setupFlat();

    // Click on country column in row 0
    result.handleCellClickToFilter("0", "country", false, row0);

    let sel = get(result.clickSelection);
    expect(sel.isCellSelected(row0DimKey, "country")).toBe(true);
    expect(sel.cellSelections.size).toBe(1);

    // Now click on city column in the same row 0; should replace, not accumulate
    result.handleCellClickToFilter("0", "city", false, row0);

    sel = get(result.clickSelection);
    expect(sel.isCellSelected(row0DimKey, "country")).toBe(false);
    expect(sel.isCellSelected(row0DimKey, "city")).toBe(true);
    expect(sel.cellSelections.size).toBe(1);

    // addDimensionValueSelections should have been called for the new cell values
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalled();

    result.destroy();
  });

  it("should still allow deselect by re-clicking the same cell", () => {
    const { result } = setupFlat();

    // Click on country in row 0
    result.handleCellClickToFilter("0", "country", false, row0);
    let sel = get(result.clickSelection);
    expect(sel.isCellSelected(row0DimKey, "country")).toBe(true);

    // Click on the same cell again to deselect
    result.handleCellClickToFilter("0", "country", false, row0);
    sel = get(result.clickSelection);
    expect(sel.isCellSelected(row0DimKey, "country")).toBe(false);
    expect(sel.cellSelections.size).toBe(0);

    result.destroy();
  });

  it("should allow selections across different rows independently", () => {
    const { result } = setupFlat();

    // Click on country in row 0
    result.handleCellClickToFilter("0", "country", false, row0);

    // Click on country in row 1 (different row; not a replacement)
    vi.mocked(getFiltersFromRow).mockImplementation(() =>
      makePivotFilter("country", ["UK"]),
    );
    result.handleCellClickToFilter("1", "country", false, row1);

    const sel = get(result.clickSelection);
    expect(sel.isCellSelected(row0DimKey, "country")).toBe(true);
    expect(sel.isCellSelected(row1DimKey, "country")).toBe(true);
    expect(sel.cellSelections.size).toBe(2);

    result.destroy();
  });
});

describe("pivot-click-to-filter: nested table multi-select", () => {
  const nestedData: PivotDataRow[] = [
    {
      country: "US",
      revenue: 100,
      subRows: [{ country: "US-East", revenue: 50 }],
    },
  ];

  const row0 = nestedData[0];
  const row0DimKey = dimKeyFromRow(row0, ["country"]);

  function setupNested() {
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    vi.mocked(getFiltersFromRow).mockImplementation(() =>
      makePivotFilter("country", ["US"]),
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(nestedConfig()) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(nestedData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
      }),
    );

    return { result, filterClass };
  }

  it("should allow multiple cells in the same row for nested tables", () => {
    const { result } = setupNested();

    // Click on two different columns in the same row
    result.handleCellClickToFilter("0", "revenue", false, row0);
    let sel = get(result.clickSelection);
    expect(sel.isCellSelected(row0DimKey, "revenue")).toBe(true);

    result.handleCellClickToFilter("0", "other_measure", false, row0);
    sel = get(result.clickSelection);

    // Both should be selected (no replacement in nested mode)
    expect(sel.isCellSelected(row0DimKey, "revenue")).toBe(true);
    expect(sel.isCellSelected(row0DimKey, "other_measure")).toBe(true);
    expect(sel.cellSelections.size).toBe(2);

    result.destroy();
  });
});

describe("pivot-click-to-filter: selection survives sorting", () => {
  it("should identify same row after data order changes (simulated sort)", () => {
    const { fm } = stubFilterManagerWithClass("mv1");

    const dataBeforeSort: PivotDataRow[] = [
      { country: "US", revenue: 100 },
      { country: "UK", revenue: 200 },
    ];

    const pivotDataStore = writable({
      isFetching: false,
      data: dataBeforeSort,
      columnDef: [],
      assembled: true,
      totalColumns: 0,
      columnDimensionAxes: {},
    });

    vi.mocked(getFiltersFromRow).mockImplementation(() =>
      makePivotFilter("country", ["US"]),
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(emptyConfig()) as Readable<PivotDataStoreConfig>,
        pivotDataStore: pivotDataStore as unknown as PivotDataStore,
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
      }),
    );

    const usRow = dataBeforeSort[0];
    const usDimKey = dimKeyFromRow(usRow, ["country"]);

    // Click on US row (index 0 before sort)
    result.handleCellClickToFilter("0", "total", false, usRow);

    let sel = get(result.clickSelection);
    expect(sel.isCellSelected(usDimKey, "total")).toBe(true);

    // Simulate sorting: UK now comes first, US second
    const dataAfterSort: PivotDataRow[] = [
      { country: "UK", revenue: 200 },
      { country: "US", revenue: 100 },
    ];
    pivotDataStore.set({
      isFetching: false,
      data: dataAfterSort,
      columnDef: [],
      assembled: true,
      totalColumns: 0,
      columnDimensionAxes: {},
    });

    // The selection should still match US, not the row at index 0 (which is now UK)
    sel = get(result.clickSelection);
    expect(sel.isCellSelected(usDimKey, "total")).toBe(true);

    // UK's dimKey should NOT be selected
    const ukDimKey = dimKeyFromRow(dataAfterSort[0], ["country"]);
    expect(sel.isCellSelected(ukDimKey, "total")).toBe(false);

    result.destroy();
  });
});
