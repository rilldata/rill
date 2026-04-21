import {
  getFiltersForColumnHeader,
  getFiltersForRowHeader,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection";
import {
  getFiltersForCell,
  getFiltersFromRow,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
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
import {
  dimKeyFromDimValues,
  dimKeyFromRow,
} from "../../../dashboards/pivot/pivot-click-selection";
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
    getFiltersFromRow: vi.fn(
      (): PivotFilter => ({
        filters: undefined,
        timeRange: { start: undefined, end: undefined },
      }),
    ),
    getFiltersForCell: vi.fn(
      (): PivotFilter => ({
        filters: undefined,
        timeRange: { start: undefined, end: undefined },
      }),
    ),
  }),
);

vi.mock(
  "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection",
  async () => ({
    ...(await vi.importActual(
      "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection",
    )),
    getFiltersForRowData: vi.fn(
      (): PivotFilter => ({
        filters: undefined,
        timeRange: { start: undefined, end: undefined },
      }),
    ),
    getFiltersForRowHeader: vi.fn(
      (): PivotFilter => ({
        filters: undefined,
        timeRange: { start: undefined, end: undefined },
      }),
    ),
    getFiltersForColumnHeader: vi.fn(
      (): PivotFilter => ({
        filters: undefined,
        timeRange: { start: undefined, end: undefined },
      }),
    ),
  }),
);

/** Build a PivotFilter with no time range (sufficient for these tests). */
function makePivotFilter(
  dimensionName: string,
  values: (string | null)[],
): PivotFilter {
  return {
    filters: createAndExpression([createInExpression(dimensionName, values)]),
    timeRange: { start: undefined, end: undefined },
  };
}

/** Build a PivotFilter with multiple dimensions (for column header tests). */
function makeMultiDimPivotFilter(
  dims: Array<{ name: string; values: (string | null)[] }>,
): PivotFilter {
  return {
    filters: createAndExpression(
      dims.map(({ name, values }) => createInExpression(name, values)),
    ),
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

function nestedConfigTwoDims() {
  return {
    rowDimensionNames: ["outer", "inner"],
    colDimensionNames: [],
    measureNames: ["revenue"],
    isFlat: false,
    time: {
      timeDimension: "",
      timeStart: undefined,
      timeEnd: undefined,
    },
  } as unknown as PivotDataStoreConfig;
}

function nestedConfigWithColDims() {
  return {
    rowDimensionNames: ["country"],
    colDimensionNames: ["region", "category", "product"],
    measureNames: ["revenue"],
    isFlat: false,
    whereFilter: createAndExpression([]),
    time: {
      timeDimension: "",
      timeStart: undefined,
      timeEnd: undefined,
    },
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

    vi.mocked(getFiltersForCell).mockImplementation(() =>
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

describe("pivot-click-to-filter: nested table cross-parent selection isolation", () => {
  // Two outer rows A and B, each with inner row X.
  // Clicking on X under A should NOT highlight X under B.
  //
  // pivotDataStore.data does NOT include the totals row (TanStack prepends it).
  // getValuesForExpandedKey adjusts indices[0] by -1 to map TanStack rowIds
  // back to this array.
  //
  // TanStack rowIds: "0" → totals, "1" → A, "2" → B, "1.0" → X-under-A, "2.0" → X-under-B
  // After adjustment: "1" maps to data[0]=A, "2" maps to data[1]=B
  const nestedTwoDimData: PivotDataRow[] = [
    {
      outer: "A",
      revenue: 100,
      subRows: [{ outer: "X", inner: "X", revenue: 50 }],
    },
    {
      outer: "B",
      revenue: 200,
      subRows: [{ outer: "X", inner: "X", revenue: 75 }],
    },
  ];

  // The inner row data as the component sees it (value mapped to first dim)
  const innerRowXUnderA = nestedTwoDimData[0].subRows![0];

  function setupNestedTwoDims() {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    // For nested cell clicks, getFiltersForCell is called
    vi.mocked(getFiltersForCell).mockImplementation(
      (_config, rowId, _colId, _axes, _data) => {
        // Return filters based on which inner row was clicked
        if (rowId === "1.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["A"] },
            { name: "inner", values: ["X"] },
          ]);
        }
        if (rowId === "2.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["B"] },
            { name: "inner", values: ["X"] },
          ]);
        }
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    // For nested row header clicks
    vi.mocked(getFiltersForRowHeader).mockImplementation(
      (_config, rowId, _data) => {
        if (rowId === "1.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["A"] },
            { name: "inner", values: ["X"] },
          ]);
        }
        if (rowId === "2.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["B"] },
            { name: "inner", values: ["X"] },
          ]);
        }
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(
          nestedConfigTwoDims(),
        ) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(nestedTwoDimData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    return { result, filterClass, selfFilteredDimensions };
  }

  it("should produce distinct dimKeys for same inner value under different parents", () => {
    // Verify that dimKeyFromDimValues produces different keys
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    const dkB = dimKeyFromDimValues({ outer: "B", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(dkA).not.toBe(dkB);
    expect(dkA).toBe("A\0X");
    expect(dkB).toBe("B\0X");
  });

  it("should NOT select X under B when clicking X under A", () => {
    const { result } = setupNestedTwoDims();

    // Click on cell in inner row X under parent A (rowId "1.0")
    result.handleCellClickToFilter("1.0", "revenue", false, innerRowXUnderA);

    const sel = get(result.clickSelection);

    // X under A should be selected
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkA, "revenue")).toBe(true);

    // X under B should NOT be selected
    const dkB = dimKeyFromDimValues({ outer: "B", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkB, "revenue")).toBe(false);

    expect(sel.cellSelections.size).toBe(1);

    result.destroy();
  });

  it("should NOT select row header X under B when clicking row header X under A", () => {
    const { result } = setupNestedTwoDims();

    // Click on row header for inner row X under parent A
    result.handleCellClickToFilter("1.0", "outer", true, innerRowXUnderA);

    const sel = get(result.clickSelection);

    const dkA = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    const dkB = dimKeyFromDimValues({ outer: "B", inner: "X" }, [
      "outer",
      "inner",
    ]);

    expect(sel.isRowHeaderSelected(dkA)).toBe(true);
    expect(sel.isRowHeaderSelected(dkB)).toBe(false);
    expect(sel.rowHeaderSelections.size).toBe(1);

    result.destroy();
  });
});

describe("pivot-click-to-filter: null dimension values", () => {
  const nullData: PivotDataRow[] = [
    { country: null, revenue: 100 },
    { country: "US", revenue: 200 },
  ];

  const nullRow = nullData[0];
  const nullDimKey = dimKeyFromRow(nullRow, ["country"]);

  function setupNull() {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    // Mock getFiltersFromRow to return a filter with null value
    vi.mocked(getFiltersFromRow).mockImplementation((_config, rowData) => {
      const value = rowData["country"];
      return makePivotFilter("country", [value as string]);
    });

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(emptyConfig()) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(nullData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    return { result, filterClass, selfFilteredDimensions };
  }

  it("should select a cell with a null dimension value", () => {
    const { result, filterClass } = setupNull();

    // Click on a row where country is null
    result.handleCellClickToFilter("0", "total", false, nullRow);

    const sel = get(result.clickSelection);
    expect(sel.isCellSelected(nullDimKey, "total")).toBe(true);
    expect(sel.cellSelections.size).toBe(1);

    // The filter should have been applied
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "country",
      [null],
    );

    result.destroy();
  });

  it("should deselect a cell with a null dimension value", () => {
    const { result, filterClass } = setupNull();

    // Select then deselect
    result.handleCellClickToFilter("0", "total", false, nullRow);
    result.handleCellClickToFilter("0", "total", false, nullRow);

    const sel = get(result.clickSelection);
    expect(sel.isCellSelected(nullDimKey, "total")).toBe(false);
    expect(sel.cellSelections.size).toBe(0);

    // Toggle should have been called to remove the null value
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalled();

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

describe("pivot-click-to-filter: column header level selection constraint", () => {
  // Level 0 (1 key): { region: "NA" }
  // Level 1 (2 keys): { region: "NA", category: "Electronics" }
  // Level 2 (3 keys): { region: "NA", category: "Electronics", product: "Laptop" }

  function setupColumnHeaders() {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    vi.mocked(getFiltersForColumnHeader).mockImplementation(
      (_config, dimensionPath) => {
        const dims = Object.entries(dimensionPath).map(([name, value]) => ({
          name,
          values: [value],
        }));
        return makeMultiDimPivotFilter(dims);
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(
          nestedConfigWithColDims(),
        ) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore([]),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    return { result, filterClass, selfFilteredDimensions, fm };
  }

  it("should allow multiple selections at the same level", () => {
    const { result } = setupColumnHeaders();

    // Select two different level-0 headers
    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });

    const sel = get(result.clickSelection);
    expect(sel.isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel.isColumnHeaderSelected({ region: "EU" })).toBe(true);
    expect(sel.columnHeaderSelections.size).toBe(2);

    result.destroy();
  });

  it("should replace column header selections when clicking a different level", () => {
    const { result, filterClass } = setupColumnHeaders();

    // Select a level-0 header
    result.handleColumnHeaderClick({ region: "NA" });

    let sel = get(result.clickSelection);
    expect(sel.isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel.columnHeaderSelections.size).toBe(1);

    // Now click a level-1 header (different level); should replace
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    sel = get(result.clickSelection);
    expect(sel.isColumnHeaderSelected({ region: "NA" })).toBe(false);
    expect(
      sel.isColumnHeaderSelected({ region: "NA", category: "Electronics" }),
    ).toBe(true);
    expect(sel.columnHeaderSelections.size).toBe(1);

    // New values should have been added
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "category",
      ["Electronics"],
    );

    result.destroy();
  });

  it("should remove orphaned values when switching to a level with different dimensions", () => {
    const { result, filterClass } = setupColumnHeaders();

    // Select level-0: { region: "NA" }
    result.handleColumnHeaderClick({ region: "NA" });

    // Switch to level-0 with different value, then switch to level-1
    // First, let's select { region: "EU" } at level 0 (same level, accumulates)
    result.handleColumnHeaderClick({ region: "EU" });

    // Now switch to level-1: { region: "NA", category: "Electronics" }
    // "EU" is no longer needed (only "NA" is in the new selection)
    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    // "EU" should have been toggled off as orphaned
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "region",
      ["EU"],
      false,
      false,
    );

    result.destroy();
  });

  it("should replace multiple same-level selections when switching levels", () => {
    const { result } = setupColumnHeaders();

    // Select two level-0 headers
    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });

    let sel = get(result.clickSelection);
    expect(sel.columnHeaderSelections.size).toBe(2);

    // Now click a level-1 header; both level-0 selections should be removed
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    sel = get(result.clickSelection);
    expect(sel.isColumnHeaderSelected({ region: "NA" })).toBe(false);
    expect(sel.isColumnHeaderSelected({ region: "EU" })).toBe(false);
    expect(
      sel.isColumnHeaderSelected({ region: "NA", category: "Electronics" }),
    ).toBe(true);
    expect(sel.columnHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("should still allow deselect by re-clicking the same header", () => {
    const { result } = setupColumnHeaders();

    // Select a level-1 header
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    let sel = get(result.clickSelection);
    expect(
      sel.isColumnHeaderSelected({ region: "NA", category: "Electronics" }),
    ).toBe(true);

    // Click it again to deselect
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    sel = get(result.clickSelection);
    expect(
      sel.isColumnHeaderSelected({ region: "NA", category: "Electronics" }),
    ).toBe(false);
    expect(sel.columnHeaderSelections.size).toBe(0);

    result.destroy();
  });

  it("should allow fresh selection at any level after all headers are deselected", () => {
    const { result } = setupColumnHeaders();

    // Select and deselect a level-0 header
    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA" });

    let sel = get(result.clickSelection);
    expect(sel.columnHeaderSelections.size).toBe(0);

    // Now select a level-1 header; should work as a fresh add
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    sel = get(result.clickSelection);
    expect(
      sel.isColumnHeaderSelected({ region: "NA", category: "Electronics" }),
    ).toBe(true);
    expect(sel.columnHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("should not remove shared dimension values when switching levels", () => {
    const { result, filterClass } = setupColumnHeaders();

    // Select level-0: { region: "NA" }
    result.handleColumnHeaderClick({ region: "NA" });

    // Switch to level-1: { region: "NA", category: "Electronics" }
    // "region: NA" is shared; it should NOT be orphaned/removed
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    // toggleDimensionValueSelections should NOT have been called with "region"
    // because "NA" is still needed by the new selection
    const toggleCalls = filterClass.toggleDimensionValueSelections.mock.calls;
    const regionToggleCalls = toggleCalls.filter(
      (call: unknown[]) => call[0] === "region",
    );
    expect(regionToggleCalls.length).toBe(0);

    // addDimensionValueSelections should have been called for the new values
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "category",
      ["Electronics"],
    );

    result.destroy();
  });
});

describe("pivot-click-to-filter: nesting level must match", () => {
  // Data: outer=A has child inner=X; outer=B has child inner=Y
  const nestedData: PivotDataRow[] = [
    {
      outer: "A",
      revenue: 100,
      subRows: [{ outer: "X", inner: "X", revenue: 50 }],
    },
    {
      outer: "B",
      revenue: 200,
      subRows: [{ outer: "Y", inner: "Y", revenue: 75 }],
    },
  ];

  const parentRowA: PivotDataRow = nestedData[0];
  const childRowXUnderA: PivotDataRow = nestedData[0].subRows![0];
  const parentRowB: PivotDataRow = nestedData[1];

  function setupNestingLevelTest() {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    // Parent row header clicks: only "outer" filter
    vi.mocked(getFiltersForRowHeader).mockImplementation(
      (_config, rowId, _data) => {
        if (rowId === "1") return makePivotFilter("outer", ["A"]);
        if (rowId === "2") return makePivotFilter("outer", ["B"]);
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    // Child cell clicks: "outer" + "inner" filter
    vi.mocked(getFiltersForCell).mockImplementation(
      (_config, rowId, _colId) => {
        if (rowId === "1.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["A"] },
            { name: "inner", values: ["X"] },
          ]);
        }
        if (rowId === "2.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["B"] },
            { name: "inner", values: ["Y"] },
          ]);
        }
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(
          nestedConfigTwoDims(),
        ) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(nestedData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    return { result, filterClass, selfFilteredDimensions };
  }

  it("should reset when parent is selected then child cell is clicked", () => {
    const { result, filterClass } = setupNestingLevelTest();

    // Click parent row header A (depth 0)
    result.handleCellClickToFilter("1", "outer", true, parentRowA);

    let sel = get(result.clickSelection);
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isRowHeaderSelected(dkA)).toBe(true);
    expect(sel.currentRowDepth).toBe(0);

    // Now click child cell X under A (depth 1) — should reset parent selection
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    result.handleCellClickToFilter("1.0", "revenue", false, childRowXUnderA);

    sel = get(result.clickSelection);

    // Parent selection should be gone
    expect(sel.isRowHeaderSelected(dkA)).toBe(false);
    // Child cell should be selected
    const dkChild = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkChild, "revenue")).toBe(true);
    expect(sel.currentRowDepth).toBe(1);

    // Old parent value should have been removed via toggleDimensionValueSelections
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "outer",
      ["A"],
      false,
      false,
    );

    result.destroy();
  });

  it("should reset when child is selected then parent header is clicked", () => {
    const { result, filterClass } = setupNestingLevelTest();

    // Click child cell X under A (depth 1)
    result.handleCellClickToFilter("1.0", "revenue", false, childRowXUnderA);

    let sel = get(result.clickSelection);
    expect(sel.currentRowDepth).toBe(1);

    // Now click parent header B (depth 0) — should reset child selection
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    result.handleCellClickToFilter("2", "outer", true, parentRowB);

    sel = get(result.clickSelection);

    // Child cell should be gone
    const dkChild = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkChild, "revenue")).toBe(false);

    // Parent B should be selected
    const dkB = dimKeyFromDimValues({ outer: "B", inner: "" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isRowHeaderSelected(dkB)).toBe(true);
    expect(sel.currentRowDepth).toBe(0);

    result.destroy();
  });

  it("should NOT reset when clicks are at the same nesting level", () => {
    const { result } = setupNestingLevelTest();

    // Click parent header A (depth 0)
    result.handleCellClickToFilter("1", "outer", true, parentRowA);

    // Click parent header B (depth 0) — same level, should accumulate
    result.handleCellClickToFilter("2", "outer", true, parentRowB);

    const sel = get(result.clickSelection);
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "" }, [
      "outer",
      "inner",
    ]);
    const dkB = dimKeyFromDimValues({ outer: "B", inner: "" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isRowHeaderSelected(dkA)).toBe(true);
    expect(sel.isRowHeaderSelected(dkB)).toBe(true);
    expect(sel.currentRowDepth).toBe(0);

    result.destroy();
  });
});

describe("pivot-click-to-filter: bordered-header anchoring", () => {
  // Same data as nesting level tests
  const nestedData: PivotDataRow[] = [
    {
      outer: "A",
      revenue: 100,
      subRows: [{ outer: "X", inner: "X", revenue: 50 }],
    },
    {
      outer: "B",
      revenue: 200,
      subRows: [{ outer: "Y", inner: "Y", revenue: 75 }],
    },
  ];

  const parentRowA: PivotDataRow = nestedData[0];
  const childRowXUnderA: PivotDataRow = nestedData[0].subRows![0];

  function setupAnchoringTest() {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    vi.mocked(getFiltersForRowHeader).mockImplementation(
      (_config, rowId, _data) => {
        if (rowId === "1") return makePivotFilter("outer", ["A"]);
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    vi.mocked(getFiltersForCell).mockImplementation(
      (_config, rowId, _colId) => {
        if (rowId === "1.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["A"] },
            { name: "inner", values: ["X"] },
          ]);
        }
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(
          nestedConfigTwoDims(),
        ) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(nestedData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    return { result, filterClass, selfFilteredDimensions };
  }

  it("clicking derived-blue header adds border without filter change", () => {
    const { result, filterClass } = setupAnchoringTest();

    // Click child cell X under A
    result.handleCellClickToFilter("1.0", "revenue", false, childRowXUnderA);
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    // Now click parent header A (derived-blue from child selection)
    result.handleCellClickToFilter("1", "outer", true, parentRowA);

    const sel = get(result.clickSelection);
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "" }, [
      "outer",
      "inner",
    ]);

    // Header should now be bordered
    expect(sel.isRowHeaderSelected(dkA)).toBe(true);

    // Child cell should still be selected
    const dkChild = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkChild, "revenue")).toBe(true);

    // No filter changes should have occurred
    expect(filterClass.addDimensionValueSelections).not.toHaveBeenCalled();
    expect(filterClass.toggleDimensionValueSelections).not.toHaveBeenCalled();

    result.destroy();
  });

  it("re-clicking bordered header with children removes border only", () => {
    const { result, filterClass } = setupAnchoringTest();

    // Click child cell, then border the parent
    result.handleCellClickToFilter("1.0", "revenue", false, childRowXUnderA);
    result.handleCellClickToFilter("1", "outer", true, parentRowA);
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    // Re-click the bordered parent header
    result.handleCellClickToFilter("1", "outer", true, parentRowA);

    const sel = get(result.clickSelection);
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "" }, [
      "outer",
      "inner",
    ]);

    // Border should be removed
    expect(sel.isRowHeaderSelected(dkA)).toBe(false);

    // Child cell should still be selected
    const dkChild = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkChild, "revenue")).toBe(true);

    // No filter changes
    expect(filterClass.addDimensionValueSelections).not.toHaveBeenCalled();
    expect(filterClass.toggleDimensionValueSelections).not.toHaveBeenCalled();

    result.destroy();
  });

  it("bordered header persists when child cell is deselected", () => {
    const { result, filterClass } = setupAnchoringTest();

    // Click child cell, then border the parent
    result.handleCellClickToFilter("1.0", "revenue", false, childRowXUnderA);
    result.handleCellClickToFilter("1", "outer", true, parentRowA);
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    // Deselect the child cell
    result.handleCellClickToFilter("1.0", "revenue", false, childRowXUnderA);

    const sel = get(result.clickSelection);

    // Child cell should be gone
    const dkChild = dimKeyFromDimValues({ outer: "A", inner: "X" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isCellSelected(dkChild, "revenue")).toBe(false);

    // Bordered parent should still be there
    const dkA = dimKeyFromDimValues({ outer: "A", inner: "" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isRowHeaderSelected(dkA)).toBe(true);

    // "inner" and "outer" from child should be removed (orphaned)
    // but "outer"="A" should be retained by the bordered header
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner",
      ["X"],
      false,
      false,
    );
    // "outer"="A" should NOT be removed since the bordered header retains it
    const outerToggleCalls =
      filterClass.toggleDimensionValueSelections.mock.calls.filter(
        (call: unknown[]) => call[0] === "outer",
      );
    expect(outerToggleCalls.length).toBe(0);

    result.destroy();
  });
});

describe("pivot-click-to-filter: parent×parent promotion", () => {
  // Nested rows: outer > inner, columns: region > category
  const nestedData: PivotDataRow[] = [
    {
      outer: "Zoom",
      revenue: 100,
      subRows: [{ outer: "US-East", inner: "US-East", revenue: 50 }],
    },
  ];

  const parentRowZoom: PivotDataRow = nestedData[0];
  const childRowUSEast: PivotDataRow = nestedData[0].subRows![0];

  it("bordered parent + parent column click promotes to parent×parent", () => {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    // Child cell clicks return outer+inner filters
    vi.mocked(getFiltersForCell).mockImplementation(
      (_config, rowId, colId) => {
        if (rowId === "1.0") {
          return makeMultiDimPivotFilter([
            { name: "outer", values: ["Zoom"] },
            { name: "inner", values: ["US-East"] },
            { name: "region", values: ["NA"] },
          ]);
        }
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    vi.mocked(getFiltersForColumnHeader).mockImplementation(
      (_config, path) => {
        if (path["quarter"]) {
          return makePivotFilter("quarter", [path["quarter"]]);
        }
        return {
          filters: undefined,
          timeRange: { start: undefined, end: undefined },
        };
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(
          nestedConfigTwoDims(),
        ) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(nestedData),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    // Click 1: child cell (Zoom > US-East × NA region)
    result.handleCellClickToFilter("1.0", "revenue", false, childRowUSEast);

    // Click 2: border the parent header (Zoom)
    result.handleCellClickToFilter("1", "outer", true, parentRowZoom);

    let sel = get(result.clickSelection);
    const dkZoom = dimKeyFromDimValues({ outer: "Zoom", inner: "" }, [
      "outer",
      "inner",
    ]);
    expect(sel.isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel.cellSelections.size).toBe(1);

    // Click 3: parent column header Q3
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    result.handleColumnHeaderClick({ quarter: "Q3" });

    sel = get(result.clickSelection);

    // Cell selections should be cleared (promoted to parent×parent)
    expect(sel.cellSelections.size).toBe(0);

    // Row header should still be there
    expect(sel.isRowHeaderSelected(dkZoom)).toBe(true);

    // Column header should be selected
    expect(sel.isColumnHeaderSelected({ quarter: "Q3" })).toBe(true);

    // Child filter values (inner, region) should have been removed
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner",
      ["US-East"],
      false,
      false,
    );

    // New column filter should have been added
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "quarter",
      ["Q3"],
    );

    result.destroy();
  });
});

describe("pivot-click-to-filter: deselect retains shared column filters", () => {
  it("should retain column dimension values still needed by remaining cells", () => {
    const selfFilteredDimensions = writable<Set<string>>(new Set());
    const { fm, filterClass } = stubFilterManagerWithClass("mv1");

    // Two rows: New York and Bronx; column: status=Closed, type=Intersection
    const data: PivotDataRow[] = [
      { borough: "New York", revenue: 100 },
      { borough: "Bronx", revenue: 200 },
    ];

    // Cell clicks return row + column filters
    vi.mocked(getFiltersFromRow).mockImplementation(
      (_config, rowData, _colId) => {
        const borough = rowData["borough"] as string;
        return makeMultiDimPivotFilter([
          { name: "borough", values: [borough] },
          { name: "status", values: ["Closed"] },
          { name: "type", values: ["Intersection"] },
        ]);
      },
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable({
          rowDimensionNames: ["borough"],
          colDimensionNames: ["status", "type"],
          measureNames: ["revenue"],
          isFlat: true,
        }) as Readable<PivotDataStoreConfig>,
        pivotDataStore: stubPivotDataStore(data),
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
        selfFilteredDimensions,
      }),
    );

    // Click cell 1: New York × Closed > Intersection
    result.handleCellClickToFilter("1", "revenue", false, data[0]);

    // Click cell 2: Bronx × Closed > Intersection
    result.handleCellClickToFilter("2", "revenue", false, data[1]);

    // Verify both cells are selected
    const dkNY = dimKeyFromRow(data[0], ["borough"]);
    const dkBronx = dimKeyFromRow(data[1], ["borough"]);
    let sel = get(result.clickSelection);
    expect(sel.isCellSelected(dkNY, "revenue")).toBe(true);
    expect(sel.isCellSelected(dkBronx, "revenue")).toBe(true);

    // Now deselect Bronx
    filterClass.addDimensionValueSelections.mockClear();
    filterClass.toggleDimensionValueSelections.mockClear();

    result.handleCellClickToFilter("2", "revenue", false, data[1]);

    sel = get(result.clickSelection);
    expect(sel.isCellSelected(dkBronx, "revenue")).toBe(false);
    expect(sel.isCellSelected(dkNY, "revenue")).toBe(true);

    // "borough"="Bronx" should be removed (orphaned)
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "borough",
      ["Bronx"],
      false,
      false,
    );

    // "status"="Closed" and "type"="Intersection" should NOT be removed
    // because Cell 1 (New York) still needs them
    const statusToggleCalls =
      filterClass.toggleDimensionValueSelections.mock.calls.filter(
        (call: unknown[]) => call[0] === "status",
      );
    const typeToggleCalls =
      filterClass.toggleDimensionValueSelections.mock.calls.filter(
        (call: unknown[]) => call[0] === "type",
      );
    expect(statusToggleCalls.length).toBe(0);
    expect(typeToggleCalls.length).toBe(0);

    result.destroy();
  });
});
