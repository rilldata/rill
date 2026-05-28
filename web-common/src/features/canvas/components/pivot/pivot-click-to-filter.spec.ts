import type {
  PivotDataRow,
  PivotDataStore,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
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
// Shared test helpers
// ---------------------------------------------------------------------------

function dk(dims: Record<string, string | null>, order: string[]): string {
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

// Builds a minimal-but-real PivotDataStoreConfig. Only the fields the
// filter-extraction helpers read are populated; the rest are left undefined.
// The cast is needed because the production type has many fields that the
// query-building code reads but the filter helpers do not.
function makeConfig(
  overrides: Partial<PivotDataStoreConfig> & {
    rowDimensionNames: string[];
    measureNames: string[];
  },
): PivotDataStoreConfig {
  return {
    colDimensionNames: [],
    isFlat: false,
    time: { timeDimension: "", timeStart: undefined, timeEnd: undefined },
    whereFilter: undefined,
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
      makeConfig({
        rowDimensionNames: ["country"],
        measureNames: ["total"],
        isFlat: true,
      }),
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

  it("replaces existing cell in the same row", () => {
    const { result } = setup(config, data);

    result.handleCellClickToFilter("0", "country", false, data[0]);
    expect(sel(result).isCellSelected(dkRow0, "country")).toBe(true);

    result.handleCellClickToFilter("0", "city", false, data[0]);
    expect(sel(result).isCellSelected(dkRow0, "country")).toBe(false);
    expect(sel(result).isCellSelected(dkRow0, "city")).toBe(true);
    expect(sel(result).cellSelections.size).toBe(1);

    result.destroy();
  });

  it("deselects by re-clicking the same cell", () => {
    const { result } = setup(config, data);

    result.handleCellClickToFilter("0", "country", false, data[0]);
    result.handleCellClickToFilter("0", "country", false, data[0]);
    expect(sel(result).cellSelections.size).toBe(0);

    result.destroy();
  });

  it("allows selections across different rows", () => {
    const { result } = setup(config, data);

    result.handleCellClickToFilter("0", "country", false, data[0]);
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
    measureNames: ["revenue", "other_measure"],
  });
  const data: PivotDataRow[] = [
    {
      country: "US",
      revenue: 100,
      subRows: [{ country: "US-East", revenue: 50 }],
    },
  ];
  const dkRow0 = dimKeyFromRow(data[0], ["country"]);

  it("allows multiple cells in the same row", () => {
    const { result } = setup(config, data);

    result.handleCellClickToFilter("1", "revenue", false, data[0]);
    result.handleCellClickToFilter("1", "other_measure", false, data[0]);

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
  const innerRowXUnderA = data[0].subRows![0];
  const dims = ["outer", "inner"];

  it("produces distinct dimKeys for same inner value under different parents", () => {
    expect(dk({ outer: "A", inner: "X" }, dims)).toBe("A\0X");
    expect(dk({ outer: "B", inner: "X" }, dims)).toBe("B\0X");
    expect(dk({ outer: "A", inner: "X" }, dims)).not.toBe(
      dk({ outer: "B", inner: "X" }, dims),
    );
  });

  it("does NOT select X under B when clicking X under A", () => {
    const { result } = setup(config, data);

    result.handleCellClickToFilter("1.0", "revenue", false, innerRowXUnderA);

    expect(
      sel(result).isCellSelected(
        dk({ outer: "A", inner: "X" }, dims),
        "revenue",
      ),
    ).toBe(true);
    expect(
      sel(result).isCellSelected(
        dk({ outer: "B", inner: "X" }, dims),
        "revenue",
      ),
    ).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);

    result.destroy();
  });

  it("does NOT select row header X under B when clicking X under A", () => {
    const { result } = setup(config, data);

    result.handleCellClickToFilter("1.0", "outer", true, innerRowXUnderA);

    expect(
      sel(result).isRowHeaderSelected(dk({ outer: "A", inner: "X" }, dims)),
    ).toBe(true);
    expect(
      sel(result).isRowHeaderSelected(dk({ outer: "B", inner: "X" }, dims)),
    ).toBe(false);

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

  it("selects a cell with null dimension value", () => {
    const { result, filterClass } = setup(config, data);

    result.handleCellClickToFilter("0", "total", false, data[0]);
    expect(sel(result).isCellSelected(dkNull, "total")).toBe(true);
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "country",
      [null],
    );

    result.destroy();
  });

  it("deselects a cell with null dimension value", () => {
    const { result, filterClass } = setup(config, data);

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
      data: [
        { country: "UK", revenue: 200 },
        { country: "US", revenue: 100 },
      ],
      columnDef: [],
      assembled: true,
      totalColumns: 0,
      columnDimensionAxes: {},
    });

    expect(sel(result).isCellSelected(usDk, "total")).toBe(true);
    expect(
      sel(result).isCellSelected(
        dimKeyFromRow({ country: "UK" }, ["country"]),
        "total",
      ),
    ).toBe(false);

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

  it("allows multiple selections at the same level", () => {
    const { result } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isColumnHeaderSelected({ region: "EU" })).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(2);

    result.destroy();
  });

  it("replaces selections when clicking a different level", () => {
    const { result, filterClass } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(false);
    expect(
      sel(result).isColumnHeaderSelected({
        region: "NA",
        category: "Electronics",
      }),
    ).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(1);
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "category",
      ["Electronics"],
    );

    result.destroy();
  });

  it("removes orphaned values when switching levels", () => {
    const { result, filterClass } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });
    filterClass.toggleDimensionValueSelections.mockClear();

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "region",
      ["EU"],
      false,
      false,
    );

    result.destroy();
  });

  it("replaces multiple same-level selections when switching levels", () => {
    const { result } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(false);
    expect(sel(result).isColumnHeaderSelected({ region: "EU" })).toBe(false);
    expect(
      sel(result).isColumnHeaderSelected({
        region: "NA",
        category: "Electronics",
      }),
    ).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("deselects by re-clicking the same header", () => {
    const { result } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expect(sel(result).columnHeaderSelections.size).toBe(0);
    result.destroy();
  });

  it("allows fresh selection at any level after all deselected", () => {
    const { result } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA" }); // deselect

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });
    expect(
      sel(result).isColumnHeaderSelected({
        region: "NA",
        category: "Electronics",
      }),
    ).toBe(true);
    expect(sel(result).columnHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("does not remove shared dimension values when switching levels", () => {
    const { result, filterClass } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    expectNoToggle(filterClass, "region");
    expect(filterClass.addDimensionValueSelections).toHaveBeenCalledWith(
      "category",
      ["Electronics"],
    );

    result.destroy();
  });

  it("removes shared child-level dim values on child-to-parent level switch across multiple children", () => {
    const { result, filterClass } = setup(config, []);

    // Two leaf (level 2) col headers sharing category=Electronics
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });
    result.handleColumnHeaderClick({ region: "EU", category: "Electronics" });
    filterClass.toggleDimensionValueSelections.mockClear();

    // Click parent (level 1) — should replace both, dropping the shared
    // category value since the new selection doesn't mention category
    result.handleColumnHeaderClick({ region: "NA" });

    expect(sel(result).columnHeaderSelections.size).toBe(1);
    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);

    // Shared category=Electronics must be toggled off exactly once and not
    // re-added by a duplicate toggle
    const categoryCalls =
      filterClass.toggleDimensionValueSelections.mock.calls.filter(
        (c: unknown[]) => c[0] === "category",
      );
    expect(categoryCalls.length).toBe(1);
    expect(categoryCalls[0]).toEqual([
      "category",
      ["Electronics"],
      false,
      false,
    ]);

    result.destroy();
  });
});

describe("deselect retains shared column filters", () => {
  it("retains column dimension values still needed by remaining cells", () => {
    const config = makeConfig({
      rowDimensionNames: ["borough"],
      colDimensionNames: ["status", "type"],
      measureNames: ["revenue"],
    });
    const data: PivotDataRow[] = [
      { borough: "New York", revenue: 100 },
      { borough: "Bronx", revenue: 200 },
    ];
    const columnDimensionAxes = {
      status: ["Closed"],
      type: ["Intersection"],
    };
    const colId = "c0v0_c1v0m0";

    const { result, filterClass } = setup(config, data, columnDimensionAxes);

    result.handleCellClickToFilter("1", colId, false, data[0]);
    result.handleCellClickToFilter("2", colId, false, data[1]);

    const dkNY = dimKeyFromRow(data[0], ["borough"]);
    const dkBronx = dimKeyFromRow(data[1], ["borough"]);
    expect(sel(result).isCellSelected(dkNY, colId)).toBe(true);
    expect(sel(result).isCellSelected(dkBronx, colId)).toBe(true);

    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("2", colId, false, data[1]);

    expect(sel(result).isCellSelected(dkBronx, colId)).toBe(false);
    expect(sel(result).isCellSelected(dkNY, colId)).toBe(true);

    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "borough",
      ["Bronx"],
      false,
      false,
    );
    expectNoToggle(filterClass, "status");
    expectNoToggle(filterClass, "type");

    result.destroy();
  });
});

describe("header/cell mutual exclusivity", () => {
  const nestedConfig = makeConfig({
    rowDimensionNames: ["outer", "inner"],
    measureNames: ["revenue", "other_measure"],
  });
  const nestedData: PivotDataRow[] = [
    {
      outer: "Zoom",
      revenue: 100,
      subRows: [{ outer: "US-East", inner: "US-East", revenue: 50 }],
    },
    {
      outer: "Airtable",
      revenue: 200,
      subRows: [{ outer: "US-West", inner: "US-West", revenue: 75 }],
    },
  ];
  const dims = ["outer", "inner"];
  const parentZoom = nestedData[0];
  const childUSEast = nestedData[0].subRows![0];
  const childUSWest = nestedData[1].subRows![0];

  function setupNested() {
    return setup(nestedConfig, nestedData);
  }

  it("row header click evicts child cells under it", () => {
    const { result, filterClass } = setupNested();
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom" }, dims);

    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);

    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("1", "outer", true, parentZoom);

    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(false);
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner",
      ["US-East"],
      false,
      false,
    );

    result.destroy();
  });

  it("cell click evicts ancestor row header", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
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

    expect(sel(result).isRowHeaderSelected(dk({ outer: "Zoom" }, dims))).toBe(
      true,
    );
    expect(
      sel(result).isCellSelected(
        dk({ outer: "Airtable", inner: "US-West" }, dims),
        "revenue",
      ),
    ).toBe(true);

    result.destroy();
  });

  it("parent row header click evicts child row header under it", () => {
    const { result, filterClass } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);

    // Select child row header first
    result.handleCellClickToFilter("1.0", "inner", true, childUSEast);
    expect(sel(result).isRowHeaderSelected(dkChild)).toBe(true);

    // Click parent row header — child must be evicted
    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("1", "outer", true, parentZoom);

    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkChild)).toBe(false);
    expect(sel(result).rowHeaderSelections.size).toBe(1);
    // Orphaned inner value is removed from the global filter
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner",
      ["US-East"],
      false,
      false,
    );

    result.destroy();
  });

  it("child row header click evicts ancestor row header above it", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);

    // Select parent first
    result.handleCellClickToFilter("1", "outer", true, parentZoom);
    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);

    // Click child row header — parent must be evicted
    result.handleCellClickToFilter("1.0", "inner", true, childUSEast);

    expect(sel(result).isRowHeaderSelected(dkChild)).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(false);
    expect(sel(result).rowHeaderSelections.size).toBe(1);

    result.destroy();
  });

  it("parent row header click keeps sibling-lineage row headers intact", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
    const dkAirtableChild = dk({ outer: "Airtable", inner: "US-West" }, dims);

    // Select a child under a different parent first
    result.handleCellClickToFilter("2.0", "inner", true, childUSWest);
    expect(sel(result).isRowHeaderSelected(dkAirtableChild)).toBe(true);

    // Click Zoom parent — Airtable's child header is a different lineage, must coexist
    result.handleCellClickToFilter("1", "outer", true, parentZoom);

    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkAirtableChild)).toBe(true);
    expect(sel(result).rowHeaderSelections.size).toBe(2);

    result.destroy();
  });

  it("parent row cell click evicts child row headers under it", () => {
    const { result, filterClass } = setupNested();
    const dkChildHeader = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom" }, dims);

    // Select a child row header first
    result.handleCellClickToFilter("1.0", "inner", true, childUSEast);
    expect(sel(result).isRowHeaderSelected(dkChildHeader)).toBe(true);

    // Click parent row's measure cell — child row header must be evicted
    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);

    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkChildHeader)).toBe(false);
    expect(sel(result).rowHeaderSelections.size).toBe(0);
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner",
      ["US-East"],
      false,
      false,
    );

    result.destroy();
  });

  it("child row cell click evicts ancestor parent row header", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);

    // Select parent row header first
    result.handleCellClickToFilter("1", "outer", true, parentZoom);
    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);

    // Click child row's measure cell — parent row header must be evicted
    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);

    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(false);
    expect(sel(result).rowHeaderSelections.size).toBe(0);

    result.destroy();
  });

  it("parent row cell click evicts child row cells under it", () => {
    const { result, filterClass } = setupNested();
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom" }, dims);

    // Select a child measure cell first
    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);

    // Click the parent row's measure cell — child cell must be evicted
    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);

    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);
    // Orphaned inner value is removed from the global filter
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "inner",
      ["US-East"],
      false,
      false,
    );

    result.destroy();
  });

  it("parent row cell click evicts multiple child row cells under it", () => {
    const { result } = setupNested();
    const dkChildEast = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkChildExpanded = nestedData[0].subRows!;
    // Add a second child to the Zoom parent for this test
    const childUSWestUnderZoom = {
      outer: "US-West",
      inner: "US-West",
      revenue: 25,
    };
    nestedData[0].subRows = [...dkChildExpanded, childUSWestUnderZoom];

    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);
    result.handleCellClickToFilter(
      "1.1",
      "revenue",
      false,
      childUSWestUnderZoom,
    );
    expect(sel(result).cellSelections.size).toBe(2);

    // Click parent row cell — both children evicted
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);

    expect(
      sel(result).isCellSelected(dk({ outer: "Zoom" }, dims), "revenue"),
    ).toBe(true);
    expect(sel(result).isCellSelected(dkChildEast, "revenue")).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);

    // Restore data for subsequent tests
    nestedData[0].subRows = dkChildExpanded;
    result.destroy();
  });

  it("child row cell click evicts ancestor parent row cell", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);

    // Select parent row cell first
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);
    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);

    // Click child row cell — parent cell must be evicted
    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);

    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);

    result.destroy();
  });

  it("parent row cell click keeps sibling-lineage cells intact", () => {
    const { result } = setupNested();
    const dkAirtableChild = dk({ outer: "Airtable", inner: "US-West" }, dims);

    // Select a cell under Airtable parent
    result.handleCellClickToFilter("2.0", "revenue", false, childUSWest);
    expect(sel(result).isCellSelected(dkAirtableChild, "revenue")).toBe(true);

    // Click Zoom parent row cell — Airtable's child cell is a different
    // lineage and must coexist
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);

    expect(
      sel(result).isCellSelected(dk({ outer: "Zoom" }, dims), "revenue"),
    ).toBe(true);
    expect(sel(result).isCellSelected(dkAirtableChild, "revenue")).toBe(true);
    expect(sel(result).cellSelections.size).toBe(2);

    result.destroy();
  });

  it("parent cell + same-row sibling-column cells coexist (not lineage)", () => {
    const { result } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);

    // Two cells in the same parent row, different columns: same dimValues,
    // not in a strict subset/superset relationship — both must coexist.
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);
    result.handleCellClickToFilter("1", "other_measure", false, parentZoom);

    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkZoom, "other_measure")).toBe(true);
    expect(sel(result).cellSelections.size).toBe(2);

    result.destroy();
  });

  // Nested rows + nested columns: parent-row cell click should evict every
  // child-row cell in that lineage regardless of which column the parent or
  // child cells sit in. Cell-on-cell row lineage ignores column dims.
  describe("with column dimensions", () => {
    const nestedColConfig = makeConfig({
      rowDimensionNames: ["outer", "inner"],
      colDimensionNames: ["quarter", "env"],
      measureNames: ["revenue"],
    });
    const colData: PivotDataRow[] = [
      {
        outer: "Zoom",
        revenue: 100,
        subRows: [
          { outer: "US-East", inner: "US-East", revenue: 50 },
          { outer: "US-West", inner: "US-West", revenue: 30 },
        ],
      },
    ];
    // Column-dim accessor format: c<colDimIdx>v<axisIdx>_..._m<measureIdx>.
    // Axes below are positional: quarter axis[0]=Q1, axis[1]=Q2; env
    // axis[0]=Prod, axis[1]=Dev.
    const columnDimensionAxes = {
      quarter: ["Q1", "Q2"],
      env: ["Prod", "Dev"],
    };
    const childEastColId = "c0v0_c1v0m0"; // Q1, Prod
    const childWestColId = "c0v1_c1v1m0"; // Q2, Dev
    const totalsColId = "m0"; // no col dims
    const q1TotalColId = "c0v0m0"; // quarter=Q1 only
    const q1ProdColId = "c0v0_c1v0m0"; // Q1, Prod

    const colDims = ["outer", "inner"];
    const parentZoomCol = colData[0];
    const childEastCol = colData[0].subRows![0];
    const childWestCol = colData[0].subRows![1];

    function setupNestedWithCols() {
      return setup(nestedColConfig, colData, columnDimensionAxes);
    }

    it("parent cell at totals column evicts all child cells", () => {
      const { result } = setupNestedWithCols();
      const dkChildEast = dk({ outer: "Zoom", inner: "US-East" }, colDims);
      const dkChildWest = dk({ outer: "Zoom", inner: "US-West" }, colDims);
      const dkZoom = dk({ outer: "Zoom" }, colDims);

      result.handleCellClickToFilter(
        "1.0",
        childEastColId,
        false,
        childEastCol,
      );
      result.handleCellClickToFilter(
        "1.1",
        childWestColId,
        false,
        childWestCol,
      );
      expect(sel(result).cellSelections.size).toBe(2);

      result.handleCellClickToFilter("1", totalsColId, false, parentZoomCol);

      expect(sel(result).isCellSelected(dkZoom, totalsColId)).toBe(true);
      expect(sel(result).isCellSelected(dkChildEast, childEastColId)).toBe(
        false,
      );
      expect(sel(result).isCellSelected(dkChildWest, childWestColId)).toBe(
        false,
      );
      expect(sel(result).cellSelections.size).toBe(1);

      result.destroy();
    });

    it("parent cell at quarter-aggregate column evicts all child cells across columns", () => {
      const { result } = setupNestedWithCols();
      const dkChildEast = dk({ outer: "Zoom", inner: "US-East" }, colDims);
      const dkChildWest = dk({ outer: "Zoom", inner: "US-West" }, colDims);
      const dkZoom = dk({ outer: "Zoom" }, colDims);

      // Children sit at different quarters
      result.handleCellClickToFilter(
        "1.0",
        childEastColId,
        false,
        childEastCol,
      );
      result.handleCellClickToFilter(
        "1.1",
        childWestColId,
        false,
        childWestCol,
      );
      expect(sel(result).cellSelections.size).toBe(2);

      // Click parent's Q1-total cell — both children must be evicted even
      // though one is at Q2.
      result.handleCellClickToFilter("1", q1TotalColId, false, parentZoomCol);

      expect(sel(result).isCellSelected(dkZoom, q1TotalColId)).toBe(true);
      expect(sel(result).isCellSelected(dkChildEast, childEastColId)).toBe(
        false,
      );
      expect(sel(result).isCellSelected(dkChildWest, childWestColId)).toBe(
        false,
      );
      expect(sel(result).cellSelections.size).toBe(1);

      result.destroy();
    });

    it("parent cell at leaf column evicts all child cells across columns", () => {
      const { result } = setupNestedWithCols();
      const dkChildEast = dk({ outer: "Zoom", inner: "US-East" }, colDims);
      const dkChildWest = dk({ outer: "Zoom", inner: "US-West" }, colDims);
      const dkZoom = dk({ outer: "Zoom" }, colDims);

      result.handleCellClickToFilter(
        "1.0",
        childEastColId,
        false,
        childEastCol,
      );
      result.handleCellClickToFilter(
        "1.1",
        childWestColId,
        false,
        childWestCol,
      );
      expect(sel(result).cellSelections.size).toBe(2);

      // Click parent's Q1×Prod leaf cell — both children must be evicted
      // regardless of which column they sit in.
      result.handleCellClickToFilter("1", q1ProdColId, false, parentZoomCol);

      expect(sel(result).isCellSelected(dkZoom, q1ProdColId)).toBe(true);
      expect(sel(result).isCellSelected(dkChildEast, childEastColId)).toBe(
        false,
      );
      expect(sel(result).isCellSelected(dkChildWest, childWestColId)).toBe(
        false,
      );
      expect(sel(result).cellSelections.size).toBe(1);

      result.destroy();
    });
  });

  // Column header mutual exclusivity uses a nested config with one column dim
  // so the real getFiltersForCell carries the region value through into the
  // cell's stored dimValues, which is what makes the cell appear "under" the
  // column header for eviction purposes.
  const nestedWithColConfig = makeConfig({
    rowDimensionNames: ["country"],
    colDimensionNames: ["region"],
    measureNames: ["revenue"],
  });
  const nestedFlatData: PivotDataRow[] = [{ country: "US", revenue: 100 }];
  const colDimAxes = { region: ["NA", "EU"] };
  const naColId = "c0v0m0";
  const euColId = "c0v1m0";

  it("column header evicts cells under it", () => {
    const { result, filterClass } = setup(
      nestedWithColConfig,
      nestedFlatData,
      colDimAxes,
    );
    const dkUS = dimKeyFromRow(nestedFlatData[0], ["country"]);

    result.handleCellClickToFilter("1", naColId, false, nestedFlatData[0]);
    expect(sel(result).isCellSelected(dkUS, naColId)).toBe(true);

    filterClass.toggleDimensionValueSelections.mockClear();
    result.handleColumnHeaderClick({ region: "NA" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isCellSelected(dkUS, naColId)).toBe(false);
    expect(filterClass.toggleDimensionValueSelections).toHaveBeenCalledWith(
      "country",
      ["US"],
      false,
      false,
    );

    result.destroy();
  });

  it("cell click evicts ancestor column header", () => {
    const { result } = setup(nestedWithColConfig, nestedFlatData, colDimAxes);
    const dkUS = dimKeyFromRow(nestedFlatData[0], ["country"]);

    result.handleColumnHeaderClick({ region: "NA" });
    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);

    result.handleCellClickToFilter("1", naColId, false, nestedFlatData[0]);

    expect(sel(result).isCellSelected(dkUS, naColId)).toBe(true);
    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(false);

    result.destroy();
  });

  it("column header + cell in different column coexist", () => {
    const { result } = setup(nestedWithColConfig, nestedFlatData, colDimAxes);
    const dkUS = dimKeyFromRow(nestedFlatData[0], ["country"]);

    result.handleColumnHeaderClick({ region: "NA" }); // header is NA
    // cell is in EU column, so they do NOT overlap
    result.handleCellClickToFilter("1", euColId, false, nestedFlatData[0]);

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isCellSelected(dkUS, euColId)).toBe(true);

    result.destroy();
  });
});
