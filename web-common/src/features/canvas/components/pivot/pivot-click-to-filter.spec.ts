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

// ---------------------------------------------------------------------------
// Partial mocks: override only filter-extraction functions; keep real exports
// (extractDimensionFiltersFromExpression, getActiveDimensionNames, etc.)
// ---------------------------------------------------------------------------

vi.mock(
  "@rilldata/web-common/features/dashboards/pivot/pivot-utils",
  async () => ({
    ...(await vi.importActual(
      "@rilldata/web-common/features/dashboards/pivot/pivot-utils",
    )),
    getFiltersFromRow: vi.fn((): PivotFilter => EMPTY_FILTER),
    getFiltersForCell: vi.fn((): PivotFilter => EMPTY_FILTER),
  }),
);

vi.mock(
  "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection",
  async () => ({
    ...(await vi.importActual(
      "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection",
    )),
    getFiltersForRowData: vi.fn((): PivotFilter => EMPTY_FILTER),
    getFiltersForRowHeader: vi.fn((): PivotFilter => EMPTY_FILTER),
    getFiltersForColumnHeader: vi.fn((): PivotFilter => EMPTY_FILTER),
  }),
);

// ---------------------------------------------------------------------------
// Shared test helpers
// ---------------------------------------------------------------------------

const EMPTY_FILTER: PivotFilter = {
  filters: undefined,
  timeRange: { start: undefined, end: undefined },
};

function filter(
  ...dims: Array<{ name: string; values: (string | null)[] }>
): PivotFilter {
  return {
    filters: createAndExpression(
      dims.map(({ name, values }) => createInExpression(name, values)),
    ),
    timeRange: { start: undefined, end: undefined },
  };
}

/** Shorthand: single-dimension filter */
function filter1(name: string, values: (string | null)[]): PivotFilter {
  return filter({ name, values });
}

function dk(
  dims: Record<string, string | null>,
  order: string[],
): string {
  return dimKeyFromDimValues(dims, order);
}

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

function makeConfig(
  overrides: Partial<PivotDataStoreConfig> & {
    rowDimensionNames: string[];
    measureNames: string[];
  },
) {
  return {
    colDimensionNames: [],
    isFlat: false,
    time: { timeDimension: "", timeStart: undefined, timeEnd: undefined },
    ...overrides,
  } as unknown as PivotDataStoreConfig;
}

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

function createFactoryArgs(
  overrides: Partial<Parameters<typeof createPivotClickToFilter>[0]> = {},
): Parameters<typeof createPivotClickToFilter>[0] {
  return {
    pivotConfig: writable(
      makeConfig({ rowDimensionNames: ["country"], measureNames: ["total"], isFlat: true }),
    ) as Readable<PivotDataStoreConfig>,
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

/** Create factory with a working filterClass and active component */
function setup(
  config: PivotDataStoreConfig,
  data: PivotDataRow[],
  columnDimensionAxes: Record<string, string[]> = {},
) {
  const selfFilteredDimensions = writable<Set<string>>(new Set());
  const { fm, filterClass } = stubFilterManagerWithClass("mv1");

  const result = createPivotClickToFilter(
    createFactoryArgs({
      pivotConfig: writable(config) as Readable<PivotDataStoreConfig>,
      pivotDataStore: stubPivotDataStore(data, columnDimensionAxes),
      filterManager: fm,
      activeComponent: writable<string | null>("pivot-1"),
      selfFilteredDimensions,
    }),
  );

  return { result, filterClass, selfFilteredDimensions, fm };
}

/** Read current click selection */
function sel(result: ReturnType<typeof setup>["result"]) {
  return get(result.clickSelection);
}

/** Assert that toggle was NOT called for a given dimension */
function expectNoToggle(
  filterClass: ReturnType<typeof stubFilterManagerWithClass>["filterClass"],
  dimensionName: string,
) {
  const calls = filterClass.toggleDimensionValueSelections.mock.calls.filter(
    (call: unknown[]) => call[0] === dimensionName,
  );
  expect(calls.length).toBe(0);
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("clearActiveComponent", () => {
  it("clears selfFilteredDimensions when activeComponent changes", () => {
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

    activeComponent.set("pivot-1");
    selfFilteredDimensions.set(new Set(["country"]));
    onBecomeInactive.mockClear();

    // Another component becomes active
    activeComponent.set("pivot-2");
    expect(get(selfFilteredDimensions).size).toBe(0);
    expect(onBecomeInactive).toHaveBeenCalled();

    result.destroy();
  });

  it("does NOT clear when this component is set as active", () => {
    const activeComponent = writable<string | null>(null);
    const selfFilteredDimensions = writable<Set<string>>(new Set());

    const result = createPivotClickToFilter(
      createFactoryArgs({ activeComponent, selfFilteredDimensions }),
    );

    selfFilteredDimensions.set(new Set(["country"]));
    activeComponent.set("pivot-1");

    expect(get(selfFilteredDimensions).has("country")).toBe(true);
    result.destroy();
  });
});

describe("flat table: single-cell-per-row", () => {
  const config = makeConfig({
    rowDimensionNames: ["country", "city"],
    measureNames: ["revenue"],
    isFlat: true,
  });
  const data: PivotDataRow[] = [
    { country: "US", city: "NYC", revenue: 100 },
    { country: "UK", city: "London", revenue: 200 },
  ];
  const dkRow0 = dimKeyFromRow(data[0], ["country", "city"]);
  const dkRow1 = dimKeyFromRow(data[1], ["country", "city"]);

  function setupFlat() {
    vi.mocked(getFiltersFromRow).mockImplementation((_cfg, _row, colId) => {
      if (colId === "country") return filter1("country", ["US"]);
      if (colId === "city") return filter1("city", ["NYC"]);
      return filter1("country", ["US"]);
    });
    return setup(config, data);
  }

  it("replaces existing cell in the same row", () => {
    const { result } = setupFlat();

    result.handleCellClickToFilter("0", "country", false, data[0]);
    expect(sel(result).isCellSelected(dkRow0, "country")).toBe(true);

    result.handleCellClickToFilter("0", "city", false, data[0]);
    expect(sel(result).isCellSelected(dkRow0, "country")).toBe(false);
    expect(sel(result).isCellSelected(dkRow0, "city")).toBe(true);
    expect(sel(result).cellSelections.size).toBe(1);

    result.destroy();
  });

  it("deselects by re-clicking the same cell", () => {
    const { result } = setupFlat();

    result.handleCellClickToFilter("0", "country", false, data[0]);
    result.handleCellClickToFilter("0", "country", false, data[0]);
    expect(sel(result).cellSelections.size).toBe(0);

    result.destroy();
  });

  it("allows selections across different rows", () => {
    const { result } = setupFlat();

    result.handleCellClickToFilter("0", "country", false, data[0]);
    vi.mocked(getFiltersFromRow).mockImplementation(() =>
      filter1("country", ["UK"]),
    );
    result.handleCellClickToFilter("1", "country", false, data[1]);

    expect(sel(result).isCellSelected(dkRow0, "country")).toBe(true);
    expect(sel(result).isCellSelected(dkRow1, "country")).toBe(true);
    expect(sel(result).cellSelections.size).toBe(2);

    result.destroy();
  });
});

describe("nested table: multi-select", () => {
  const config = makeConfig({
    rowDimensionNames: ["country"],
    measureNames: ["revenue"],
  });
  const data: PivotDataRow[] = [
    { country: "US", revenue: 100, subRows: [{ country: "US-East", revenue: 50 }] },
  ];
  const dkRow0 = dimKeyFromRow(data[0], ["country"]);

  it("allows multiple cells in the same row", () => {
    vi.mocked(getFiltersForCell).mockImplementation(() =>
      filter1("country", ["US"]),
    );
    const { result } = setup(config, data);

    result.handleCellClickToFilter("0", "revenue", false, data[0]);
    result.handleCellClickToFilter("0", "other_measure", false, data[0]);

    expect(sel(result).isCellSelected(dkRow0, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkRow0, "other_measure")).toBe(true);
    expect(sel(result).cellSelections.size).toBe(2);

    result.destroy();
  });
});

describe("nested table: cross-parent selection isolation", () => {
  const config = makeConfig({
    rowDimensionNames: ["outer", "inner"],
    measureNames: ["revenue"],
  });
  const data: PivotDataRow[] = [
    { outer: "A", revenue: 100, subRows: [{ outer: "X", inner: "X", revenue: 50 }] },
    { outer: "B", revenue: 200, subRows: [{ outer: "X", inner: "X", revenue: 75 }] },
  ];
  const innerRowXUnderA = data[0].subRows![0];
  const dims = ["outer", "inner"];

  function setupCrossParent() {
    vi.mocked(getFiltersForCell).mockImplementation((_cfg, rowId) => {
      if (rowId === "1.0") return filter({ name: "outer", values: ["A"] }, { name: "inner", values: ["X"] });
      if (rowId === "2.0") return filter({ name: "outer", values: ["B"] }, { name: "inner", values: ["X"] });
      return EMPTY_FILTER;
    });
    vi.mocked(getFiltersForRowHeader).mockImplementation((_cfg, rowId) => {
      if (rowId === "1.0") return filter({ name: "outer", values: ["A"] }, { name: "inner", values: ["X"] });
      if (rowId === "2.0") return filter({ name: "outer", values: ["B"] }, { name: "inner", values: ["X"] });
      return EMPTY_FILTER;
    });
    return setup(config, data);
  }

  it("produces distinct dimKeys for same inner value under different parents", () => {
    expect(dk({ outer: "A", inner: "X" }, dims)).toBe("A\0X");
    expect(dk({ outer: "B", inner: "X" }, dims)).toBe("B\0X");
    expect(dk({ outer: "A", inner: "X" }, dims)).not.toBe(
      dk({ outer: "B", inner: "X" }, dims),
    );
  });

  it("does NOT select X under B when clicking X under A", () => {
    const { result } = setupCrossParent();

    result.handleCellClickToFilter("1.0", "revenue", false, innerRowXUnderA);

    expect(sel(result).isCellSelected(dk({ outer: "A", inner: "X" }, dims), "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dk({ outer: "B", inner: "X" }, dims), "revenue")).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);

    result.destroy();
  });

  it("does NOT select row header X under B when clicking X under A", () => {
    const { result } = setupCrossParent();

    result.handleCellClickToFilter("1.0", "outer", true, innerRowXUnderA);

    expect(sel(result).isRowHeaderSelected(dk({ outer: "A", inner: "X" }, dims))).toBe(true);
    expect(sel(result).isRowHeaderSelected(dk({ outer: "B", inner: "X" }, dims))).toBe(false);

    result.destroy();
  });
});

describe("null dimension values", () => {
  const config = makeConfig({
    rowDimensionNames: ["country"],
    measureNames: ["total"],
    isFlat: true,
  });
  const data: PivotDataRow[] = [
    { country: null, revenue: 100 },
    { country: "US", revenue: 200 },
  ];
  const dkNull = dimKeyFromRow(data[0], ["country"]);

  function setupNull() {
    vi.mocked(getFiltersFromRow).mockImplementation((_cfg, rowData) =>
      filter1("country", [rowData["country"] as string]),
    );
    return setup(config, data);
  }

  it("selects a cell with null dimension value", () => {
    const { result, filterClass } = setupNull();

    result.handleCellClickToFilter("0", "total", false, data[0]);
    expect(sel(result).isCellSelected(dkNull, "total")).toBe(true);
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith("country", [null]);

    result.destroy();
  });

  it("deselects a cell with null dimension value", () => {
    const { result, filterClass } = setupNull();

    result.handleCellClickToFilter("0", "total", false, data[0]);
    result.handleCellClickToFilter("0", "total", false, data[0]);

    expect(sel(result).cellSelections.size).toBe(0);
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalled();

    result.destroy();
  });
});

describe("selection survives sorting", () => {
  it("identifies same row after data order changes", () => {
    const { fm } = stubFilterManagerWithClass("mv1");
    const config = makeConfig({
      rowDimensionNames: ["country"],
      measureNames: ["total"],
      isFlat: true,
    });

    const dataBefore: PivotDataRow[] = [
      { country: "US", revenue: 100 },
      { country: "UK", revenue: 200 },
    ];
    const pivotDataStore = writable({
      isFetching: false,
      data: dataBefore,
      columnDef: [],
      assembled: true,
      totalColumns: 0,
      columnDimensionAxes: {},
    });

    vi.mocked(getFiltersFromRow).mockImplementation(() =>
      filter1("country", ["US"]),
    );

    const result = createPivotClickToFilter(
      createFactoryArgs({
        pivotConfig: writable(config) as Readable<PivotDataStoreConfig>,
        pivotDataStore: pivotDataStore as unknown as PivotDataStore,
        filterManager: fm,
        activeComponent: writable<string | null>("pivot-1"),
      }),
    );

    const usDk = dimKeyFromRow(dataBefore[0], ["country"]);
    result.handleCellClickToFilter("0", "total", false, dataBefore[0]);
    expect(sel(result).isCellSelected(usDk, "total")).toBe(true);

    // Simulate sort: UK now first
    pivotDataStore.set({
      isFetching: false,
      data: [{ country: "UK", revenue: 200 }, { country: "US", revenue: 100 }],
      columnDef: [],
      assembled: true,
      totalColumns: 0,
      columnDimensionAxes: {},
    });

    expect(sel(result).isCellSelected(usDk, "total")).toBe(true);
    expect(sel(result).isCellSelected(dimKeyFromRow({ country: "UK" }, ["country"]), "total")).toBe(false);

    result.destroy();
  });
});

describe("column header level selection constraint", () => {
  const config = makeConfig({
    rowDimensionNames: ["country"],
    colDimensionNames: ["region", "category", "product"],
    measureNames: ["revenue"],
    whereFilter: createAndExpression([]),
  });

  function setupColHeaders() {
    vi.mocked(getFiltersForColumnHeader).mockImplementation((_cfg, path) => {
      const dims = Object.entries(path).map(([name, value]) => ({
        name,
        values: [value],
      }));
      return filter(...dims);
    });
    return setup(config, []);
  }

  it("allows multiple selections at the same level", () => {
    const { result } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isColumnHeaderSelected({ region: "EU" })).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(2);

    result.destroy();
  });

  it("replaces selections when clicking a different level", () => {
    const { result, filterClass } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(false);
    expect(sel(result).isColumnHeaderSelected({ region: "NA", category: "Electronics" })).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(1);
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith("category", ["Electronics"]);

    result.destroy();
  });

  it("removes orphaned values when switching levels", () => {
    const { result, filterClass } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });
    filterClass.toggleDimensionValueSelections.mockClear();

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "region", ["EU"], false, false,
    );

    result.destroy();
  });

  it("replaces multiple same-level selections when switching levels", () => {
    const { result } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(false);
    expect(sel(result).isColumnHeaderSelected({ region: "EU" })).toBe(false);
    expect(sel(result).isColumnHeaderSelected({ region: "NA", category: "Electronics" })).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("deselects by re-clicking the same header", () => {
    const { result } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(sel(result).columnHeaderSelections.size).toBe(0);
    result.destroy();
  });

  it("allows fresh selection at any level after all deselected", () => {
    const { result } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA" }); // deselect

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });
    expect(sel(result).isColumnHeaderSelected({ region: "NA", category: "Electronics" })).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("does not remove shared dimension values when switching levels", () => {
    const { result, filterClass } = setupColHeaders();

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expectNoToggle(filterClass, "region");
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith("category", ["Electronics"]);

    result.destroy();
  });
});

describe("deselect retains shared column filters", () => {
  it("retains column dimension values still needed by remaining cells", () => {
    const config = makeConfig({
      rowDimensionNames: ["borough"],
      colDimensionNames: ["status", "type"],
      measureNames: ["revenue"],
      isFlat: true,
    });
    const data: PivotDataRow[] = [
      { borough: "New York", revenue: 100 },
      { borough: "Bronx", revenue: 200 },
    ];

    vi.mocked(getFiltersFromRow).mockImplementation((_cfg, rowData) => {
      return filter(
        { name: "borough", values: [rowData["borough"] as string] },
        { name: "status", values: ["Closed"] },
        { name: "type", values: ["Intersection"] },
      );
    });

    const { result, filterClass } = setup(config, data);

    result.handleCellClickToFilter("1", "revenue", false, data[0]);
    result.handleCellClickToFilter("2", "revenue", false, data[1]);

    const dkNY = dimKeyFromRow(data[0], ["borough"]);
    const dkBronx = dimKeyFromRow(data[1], ["borough"]);
    expect(sel(result).isCellSelected(dkNY, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkBronx, "revenue")).toBe(true);

    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("2", "revenue", false, data[1]);

    expect(sel(result).isCellSelected(dkBronx, "revenue")).toBe(false);
    expect(sel(result).isCellSelected(dkNY, "revenue")).toBe(true);

    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "borough", ["Bronx"], false, false,
    );
    expectNoToggle(filterClass, "status");
    expectNoToggle(filterClass, "type");

    result.destroy();
  });
});

describe("header/cell mutual exclusivity", () => {
  const nestedConfig = makeConfig({
    rowDimensionNames: ["outer", "inner"],
    measureNames: ["revenue"],
  });
  const nestedData: PivotDataRow[] = [
    { outer: "Zoom", revenue: 100, subRows: [{ outer: "US-East", inner: "US-East", revenue: 50 }] },
    { outer: "Airtable", revenue: 200, subRows: [{ outer: "US-West", inner: "US-West", revenue: 75 }] },
  ];
  const dims = ["outer", "inner"];
  const parentZoom = nestedData[0];
  const childUSEast = nestedData[0].subRows![0];
  const childUSWest = nestedData[1].subRows![0];

  function setupNested() {
    vi.mocked(getFiltersForRowHeader).mockImplementation((_cfg, rowId) => {
      if (rowId === "1") return filter1("outer", ["Zoom"]);
      if (rowId === "2") return filter1("outer", ["Airtable"]);
      return EMPTY_FILTER;
    });
    vi.mocked(getFiltersForCell).mockImplementation((_cfg, rowId) => {
      if (rowId === "1.0") return filter({ name: "outer", values: ["Zoom"] }, { name: "inner", values: ["US-East"] });
      if (rowId === "2.0") return filter({ name: "outer", values: ["Airtable"] }, { name: "inner", values: ["US-West"] });
      return EMPTY_FILTER;
    });
    return setup(nestedConfig, nestedData);
  }

  it("row header click evicts child cells under it", () => {
    const { result, filterClass } = setupNested();
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom", inner: "" }, dims);

    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);

    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("1", "outer", true, parentZoom);

    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(false);
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner", ["US-East"], false, false,
    );

    result.destroy();
  });

  it("cell click evicts ancestor row header", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom", inner: "" }, dims);
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);

    result.handleCellClickToFilter("1", "outer", true, parentZoom);
    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);

    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);

    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(false);

    result.destroy();
  });

  it("different lineage coexists: header + cell under different parent", () => {
    const { result } = setupNested();

    result.handleCellClickToFilter("1", "outer", true, parentZoom);
    result.handleCellClickToFilter("2.0", "revenue", false, childUSWest);

    expect(sel(result).isRowHeaderSelected(dk({ outer: "Zoom", inner: "" }, dims))).toBe(true);
    expect(sel(result).isCellSelected(dk({ outer: "Airtable", inner: "US-West" }, dims), "revenue")).toBe(true);

    result.destroy();
  });

  // Column header mutual exclusivity uses flat config with column dims
  const flatWithColConfig = makeConfig({
    rowDimensionNames: ["country"],
    colDimensionNames: ["region"],
    measureNames: ["revenue"],
    isFlat: true,
  });
  const flatData: PivotDataRow[] = [{ country: "US", revenue: 100 }];

  function setupFlatWithCol(cellRegion: string) {
    vi.mocked(getFiltersFromRow).mockImplementation(() =>
      filter({ name: "country", values: ["US"] }, { name: "region", values: [cellRegion] }),
    );
    vi.mocked(getFiltersForColumnHeader).mockImplementation((_cfg, path) =>
      filter1("region", [path["region"]]),
    );
    return setup(flatWithColConfig, flatData);
  }

  it("column header evicts cells under it", () => {
    const { result, filterClass } = setupFlatWithCol("NA");
    const dkUS = dimKeyFromRow(flatData[0], ["country"]);

    result.handleCellClickToFilter("1", "revenue", false, flatData[0]);
    expect(sel(result).isCellSelected(dkUS, "revenue")).toBe(true);

    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleColumnHeaderClick({ region: "NA" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isCellSelected(dkUS, "revenue")).toBe(false);
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "country", ["US"], false, false,
    );

    result.destroy();
  });

  it("cell click evicts ancestor column header", () => {
    const { result } = setupFlatWithCol("NA");
    const dkUS = dimKeyFromRow(flatData[0], ["country"]);

    result.handleColumnHeaderClick({ region: "NA" });
    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);

    result.handleCellClickToFilter("1", "revenue", false, flatData[0]);

    expect(sel(result).isCellSelected(dkUS, "revenue")).toBe(true);
    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(false);

    result.destroy();
  });

  it("column header + cell in different column coexist", () => {
    const { result } = setupFlatWithCol("EU"); // cell is in EU column
    const dkUS = dimKeyFromRow(flatData[0], ["country"]);

    result.handleColumnHeaderClick({ region: "NA" }); // header is NA
    result.handleCellClickToFilter("1", "revenue", false, flatData[0]);

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isCellSelected(dkUS, "revenue")).toBe(true);

    result.destroy();
  });
});
