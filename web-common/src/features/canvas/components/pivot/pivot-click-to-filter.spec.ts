import type {
  PivotDataRow,
  PivotDataStore,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  getValuesInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import { get, writable, type Readable } from "svelte/store";
import {
  afterAll,
  afterEach,
  beforeAll,
  describe,
  expect,
  it,
  vi,
} from "vitest";
import {
  dimKeyFromDimValues,
  dimKeyFromRow,
} from "../../../dashboards/pivot/pivot-click-selection";
import { createPivotClickToFilter } from "./pivot-click-to-filter";
import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import {
  createInEffectRoot,
  createTestMetricsViewsProvider,
  useMetricsViewMocks,
} from "@rilldata/web-common/features/metrics-views/providers/test/metrics-views-test-utils.svelte.ts";
import { PIVOT_METRICS_INIT, PIVOT_TEST_METRICS_NAME } from "./pivot-test-data";

// ---------------------------------------------------------------------------
// Shared test helpers
// ---------------------------------------------------------------------------

useMetricsViewMocks({ [PIVOT_TEST_METRICS_NAME]: PIVOT_METRICS_INIT });

// The specs are read-only, so a single provider serves every test in this file.
// Each test still gets its own ExpressionFilterManager.
let metricsViewsProvider: MetricsViewsProvider;
let destroyProvider: () => void;
const filterManagerCleanups: (() => void)[] = [];

beforeAll(async () => {
  const provider = await createTestMetricsViewsProvider([
    PIVOT_TEST_METRICS_NAME,
  ]);
  metricsViewsProvider = provider.value;
  destroyProvider = provider.destroy;
});

afterEach(() => {
  filterManagerCleanups.splice(0).forEach((destroy) => destroy());
});

afterAll(() => destroyProvider());

function dk(dims: Record<string, string | null>, order: string[]): string {
  return dimKeyFromDimValues(dims, order);
}

/** A real filter manager over {@link PIVOT_METRICS_INIT}, torn down after the test. */
function createFilterManager() {
  const { value, destroy } = createInEffectRoot(
    () =>
      new ExpressionFilterManager(
        metricsViewsProvider,
        new YAMLConfigProvider(),
      ),
  );
  filterManagerCleanups.push(destroy);
  return value;
}

/** Values currently filtered on `dimensionName`, in selection order. */
function selectedValues(
  fm: ExpressionFilterManager,
  dimensionName: string,
): (string | null)[] {
  const expr = fm.exprByMetricsView[PIVOT_TEST_METRICS_NAME];
  const dimensionExpr = expr?.cond?.exprs?.find(
    (e) => e.cond?.exprs?.[0]?.ident === dimensionName,
  );
  return dimensionExpr ? getValuesInExpression(dimensionExpr) : [];
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
  // Destructured so that the default filter manager is only built when the
  // caller does not bring its own.
  const { filterManager = createFilterManager(), ...rest } = overrides;

  return {
    pivotConfig: writable(
      makeConfig({
        rowDimensionNames: ["country"],
        measureNames: ["total"],
        isFlat: true,
      }),
    ) as Readable<PivotDataStoreConfig>,
    pivotDataStore: stubPivotDataStore([]),
    filterManager,
    componentId: "pivot-1",
    activeComponent: writable<string | null>(null),
    selfFilteredDimensions: writable<Set<string>>(new Set()),
    whereFilterStore: writable<V1Expression | undefined>(undefined),
    ...rest,
  };
}

/** Create factory with a real filter manager and this component set as active */
function setup(
  config: PivotDataStoreConfig,
  data: PivotDataRow[],
  columnDimensionAxes: Record<string, string[]> = {},
) {
  const selfFilteredDimensions = writable<Set<string>>(new Set());
  const fm = createFilterManager();

  const result = createPivotClickToFilter(
    createFactoryArgs({
      pivotConfig: writable(config) as Readable<PivotDataStoreConfig>,
      pivotDataStore: stubPivotDataStore(data, columnDimensionAxes),
      filterManager: fm,
      activeComponent: writable<string | null>("pivot-1"),
      selfFilteredDimensions,
    }),
  );

  return { result, selfFilteredDimensions, fm };
}

/** Read current click selection */
function sel(result: ReturnType<typeof setup>["result"]) {
  return get(result.clickSelection);
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
    const { result, fm } = setup(config, data);

    result.handleCellClickToFilter("0", "total", false, data[0]);
    expect(sel(result).isCellSelected(dkNull, "total")).toBe(true);
    expect(selectedValues(fm, "country")).toEqual([null]);

    result.destroy();
  });

  it("deselects a cell with null dimension value", () => {
    const { result, fm } = setup(config, data);

    result.handleCellClickToFilter("0", "total", false, data[0]);
    result.handleCellClickToFilter("0", "total", false, data[0]);

    expect(sel(result).cellSelections.size).toBe(0);
    expect(selectedValues(fm, "country")).toEqual([]);

    result.destroy();
  });
});

describe("selection survives sorting", () => {
  it("identifies same row after data order changes", () => {
    const fm = createFilterManager();
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
    const { result, fm } = setup(config, []);

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
    expect(selectedValues(fm, "category")).toEqual(["Electronics"]);

    result.destroy();
  });

  it("removes orphaned values when switching levels", () => {
    const { result, fm } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "EU" });
    expect(selectedValues(fm, "region")).toEqual(["NA", "EU"]);

    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    // EU is no longer part of any selection, so it is dropped from the filter
    expect(selectedValues(fm, "region")).toEqual(["NA"]);

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
    const { result, fm } = setup(config, []);

    result.handleColumnHeaderClick({ region: "NA" });
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });

    // region=NA is part of the new selection too, so it survives the switch
    expect(selectedValues(fm, "region")).toEqual(["NA"]);
    expect(selectedValues(fm, "category")).toEqual(["Electronics"]);

    result.destroy();
  });

  it("removes shared child-level dim values on child-to-parent level switch across multiple children", () => {
    const { result, fm } = setup(config, []);

    // Two leaf (level 2) col headers sharing category=Electronics
    result.handleColumnHeaderClick({ region: "NA", category: "Electronics" });
    result.handleColumnHeaderClick({ region: "EU", category: "Electronics" });
    expect(selectedValues(fm, "category")).toEqual(["Electronics"]);

    // Click parent (level 1) — should replace both, dropping the shared
    // category value since the new selection doesn't mention category
    result.handleColumnHeaderClick({ region: "NA" });

    expect(sel(result).columnHeaderSelections.size).toBe(1);
    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);

    // The category value both old headers shared is gone, and the surviving
    // region value is not duplicated
    expect(selectedValues(fm, "category")).toEqual([]);
    expect(selectedValues(fm, "region")).toEqual(["NA"]);

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

    const { result, fm } = setup(config, data, columnDimensionAxes);

    result.handleCellClickToFilter("1", colId, false, data[0]);
    result.handleCellClickToFilter("2", colId, false, data[1]);

    const dkNY = dimKeyFromRow(data[0], ["borough"]);
    const dkBronx = dimKeyFromRow(data[1], ["borough"]);
    expect(sel(result).isCellSelected(dkNY, colId)).toBe(true);
    expect(sel(result).isCellSelected(dkBronx, colId)).toBe(true);

    result.handleCellClickToFilter("2", colId, false, data[1]);

    expect(sel(result).isCellSelected(dkBronx, colId)).toBe(false);
    expect(sel(result).isCellSelected(dkNY, colId)).toBe(true);

    expect(selectedValues(fm, "borough")).toEqual(["New York"]);
    // The remaining cell still sits in this column, so its values stay
    expect(selectedValues(fm, "status")).toEqual(["Closed"]);
    expect(selectedValues(fm, "type")).toEqual(["Intersection"]);

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
    const { result, fm } = setupNested();
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom" }, dims);

    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);
    expect(selectedValues(fm, "inner")).toEqual(["US-East"]);

    result.handleCellClickToFilter("1", "outer", true, parentZoom);

    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(false);
    expect(selectedValues(fm, "inner")).toEqual([]);

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
    const { result, fm } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);

    // Select child row header first
    result.handleCellClickToFilter("1.0", "inner", true, childUSEast);
    expect(sel(result).isRowHeaderSelected(dkChild)).toBe(true);
    expect(selectedValues(fm, "inner")).toEqual(["US-East"]);

    // Click parent row header — child must be evicted
    result.handleCellClickToFilter("1", "outer", true, parentZoom);

    expect(sel(result).isRowHeaderSelected(dkZoom)).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkChild)).toBe(false);
    expect(sel(result).rowHeaderSelections.size).toBe(1);
    // Orphaned inner value is removed from the global filter
    expect(selectedValues(fm, "inner")).toEqual([]);

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
    const { result, fm } = setupNested();
    const dkChildHeader = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom" }, dims);

    // Select a child row header first
    result.handleCellClickToFilter("1.0", "inner", true, childUSEast);
    expect(sel(result).isRowHeaderSelected(dkChildHeader)).toBe(true);
    expect(selectedValues(fm, "inner")).toEqual(["US-East"]);

    // Click parent row's measure cell — child row header must be evicted
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);

    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);
    expect(sel(result).isRowHeaderSelected(dkChildHeader)).toBe(false);
    expect(sel(result).rowHeaderSelections.size).toBe(0);
    expect(selectedValues(fm, "inner")).toEqual([]);

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
    const { result, fm } = setupNested();
    const dkChild = dk({ outer: "Zoom", inner: "US-East" }, dims);
    const dkZoom = dk({ outer: "Zoom" }, dims);

    // Select a child measure cell first
    result.handleCellClickToFilter("1.0", "revenue", false, childUSEast);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(true);
    expect(selectedValues(fm, "inner")).toEqual(["US-East"]);

    // Click the parent row's measure cell — child cell must be evicted
    result.handleCellClickToFilter("1", "revenue", false, parentZoom);

    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkChild, "revenue")).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);
    // Orphaned inner value is removed from the global filter
    expect(selectedValues(fm, "inner")).toEqual([]);

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

  it("deselect that orphans nothing leaves the filter untouched", () => {
    const { result, fm } = setupNested();
    const dkZoom = dk({ outer: "Zoom" }, dims);

    result.handleCellClickToFilter("1", "revenue", false, parentZoom);
    result.handleCellClickToFilter("1", "other_measure", false, parentZoom);
    expect(selectedValues(fm, "outer")).toEqual(["Zoom"]);

    // Deselecting one of the two cells removes no dimension value, since the
    // remaining cell in the same row still needs outer=Zoom.
    result.handleCellClickToFilter("1", "other_measure", false, parentZoom);

    expect(sel(result).isCellSelected(dkZoom, "revenue")).toBe(true);
    expect(sel(result).isCellSelected(dkZoom, "other_measure")).toBe(false);
    expect(sel(result).cellSelections.size).toBe(1);
    expect(selectedValues(fm, "outer")).toEqual(["Zoom"]);

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
    const { result, fm } = setup(
      nestedWithColConfig,
      nestedFlatData,
      colDimAxes,
    );
    const dkUS = dimKeyFromRow(nestedFlatData[0], ["country"]);

    result.handleCellClickToFilter("1", naColId, false, nestedFlatData[0]);
    expect(sel(result).isCellSelected(dkUS, naColId)).toBe(true);
    expect(selectedValues(fm, "country")).toEqual(["US"]);

    result.handleColumnHeaderClick({ region: "NA" });

    expect(sel(result).isColumnHeaderSelected({ region: "NA" })).toBe(true);
    expect(sel(result).isCellSelected(dkUS, naColId)).toBe(false);
    expect(selectedValues(fm, "country")).toEqual([]);

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
