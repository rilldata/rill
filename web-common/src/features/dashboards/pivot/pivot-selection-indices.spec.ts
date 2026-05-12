import type { HeaderGroup, Row } from "tanstack-table-8-svelte-5";
import { describe, expect, it } from "vitest";
import {
  buildClickSelection,
  columnHeaderKey,
  nestedDimKeyFromRow,
} from "./pivot-click-selection";
import {
  computeAncestorRowIds,
  computeCellSelectedColDimGroupIndices,
  computeSelectedColIndices,
  isHeaderInHoveredRange,
  isHoveredHeader,
  isInCellSelectedColRange,
  isInSelectedColRange,
} from "./pivot-selection-indices";
import type { PivotDataRow } from "./types";

function header(
  colSpan: number,
  dimensionPath?: Record<string, string>,
): HeaderGroup<PivotDataRow>["headers"][number] {
  return {
    colSpan,
    column: {
      columnDef: {
        meta: dimensionPath ? { dimensionPath } : undefined,
      },
    },
  } as unknown as HeaderGroup<PivotDataRow>["headers"][number];
}

// ---- isHeaderInHoveredRange ----

describe("isHeaderInHoveredRange", () => {
  it("returns true when header is within hovered range", () => {
    expect(isHeaderInHoveredRange(2, 1, { start: 0, size: 4 })).toBe(true);
  });

  it("returns true when header exactly matches hovered range", () => {
    expect(isHeaderInHoveredRange(0, 4, { start: 0, size: 4 })).toBe(true);
  });

  it("returns false when header extends beyond hovered range", () => {
    expect(isHeaderInHoveredRange(2, 3, { start: 0, size: 4 })).toBe(false);
  });

  it("returns false when header is before hovered range", () => {
    expect(isHeaderInHoveredRange(0, 1, { start: 2, size: 2 })).toBe(false);
  });

  it("returns false when no hover", () => {
    expect(isHeaderInHoveredRange(0, 1, null)).toBe(false);
  });
});

// ---- isHoveredHeader ----

describe("isHoveredHeader", () => {
  it("returns true when header is the exact hovered header", () => {
    expect(isHoveredHeader(2, 3, { start: 2, size: 3 })).toBe(true);
  });

  it("returns false when start differs", () => {
    expect(isHoveredHeader(0, 3, { start: 2, size: 3 })).toBe(false);
  });

  it("returns false when size differs", () => {
    expect(isHoveredHeader(2, 1, { start: 2, size: 3 })).toBe(false);
  });

  it("returns false when no hover", () => {
    expect(isHoveredHeader(0, 1, null)).toBe(false);
  });
});

// ---- isInSelectedColRange ----

describe("isInSelectedColRange", () => {
  it("returns true when all columns in range are selected", () => {
    const indices = new Set([0, 1, 2, 3]);
    expect(isInSelectedColRange(1, 2, false, indices)).toBe(true);
  });

  it("returns false when not all columns in range are selected", () => {
    const indices = new Set([0, 2]);
    expect(isInSelectedColRange(0, 3, false, indices)).toBe(false);
  });

  it("returns false when self-selected (avoid double-highlighting)", () => {
    const indices = new Set([0, 1]);
    expect(isInSelectedColRange(0, 2, true, indices)).toBe(false);
  });

  it("returns false when no selections", () => {
    expect(isInSelectedColRange(0, 2, false, new Set())).toBe(false);
  });

  it("returns false when colSpan is 0", () => {
    const indices = new Set([0]);
    expect(isInSelectedColRange(0, 0, false, indices)).toBe(false);
  });
});

// ---- isInCellSelectedColRange ----

describe("isInCellSelectedColRange", () => {
  it("returns true when any column in range has a cell selected", () => {
    const indices = new Set([3]);
    expect(isInCellSelectedColRange(2, 3, indices)).toBe(true);
  });

  it("returns false when no columns in range have a cell selected", () => {
    const indices = new Set([5]);
    expect(isInCellSelectedColRange(0, 3, indices)).toBe(false);
  });

  it("returns false with empty set", () => {
    expect(isInCellSelectedColRange(0, 3, new Set())).toBe(false);
  });
});

describe("computeSelectedColIndices", () => {
  const headerGroups = [
    {
      headers: [
        header(1),
        header(3, { component: "cold" }),
        header(2, { component: "batch" }),
      ],
    },
    {
      headers: [
        header(1),
        header(1, { component: "cold", plan: "Standard Plan" }),
        header(1, { component: "cold", plan: "Legacy" }),
        header(1, { component: "cold", plan: "POC" }),
        header(1, { component: "batch", plan: "Standard Plan" }),
        header(1, { component: "batch", plan: "Legacy" }),
      ],
    },
  ] as unknown as HeaderGroup<PivotDataRow>[];

  it("includes cross-product child columns from selected column-header filters", () => {
    const selection = buildClickSelection(
      new Map(),
      new Map(),
      new Set([
        columnHeaderKey({ component: "cold", plan: "Standard Plan" }),
        columnHeaderKey({ component: "batch", plan: "Legacy" }),
      ]),
    );

    expect([...computeSelectedColIndices(selection, headerGroups)]).toEqual([
      1, 2, 4, 5,
    ]);
  });

  it("keeps parent header selections scoped to the parent span", () => {
    const selection = buildClickSelection(
      new Map(),
      new Map(),
      new Set([columnHeaderKey({ component: "cold" })]),
    );

    expect([...computeSelectedColIndices(selection, headerGroups)]).toEqual([
      1, 2, 3,
    ]);
  });
});

describe("computeCellSelectedColDimGroupIndices", () => {
  const rowDimensionNames = ["city"];

  function selectionFor(
    colDimValues: Record<string, string | null>,
    columnId = "selected_col",
  ) {
    const dimKey = "Bronx";
    return buildClickSelection(
      new Map(),
      new Map([
        [
          `${dimKey}\t${columnId}`,
          {
            dimKey,
            dimValues: { city: "Bronx", ...colDimValues },
            columnId,
          },
        ],
      ]),
      new Set(),
    );
  }

  it("requires the full column dimension path for a single-measure cell group", () => {
    const headerGroups = [
      {
        headers: [
          header(1),
          header(3, { status: "closed" }),
          header(1, { status: "open" }),
        ],
      },
      {
        headers: [
          header(1),
          header(1, { status: "closed", facility_type: "null" }),
          header(1, { status: "closed", facility_type: "DS" }),
          header(1, { status: "closed", facility_type: "n/a" }),
          header(1, { status: "open", facility_type: "n/a" }),
        ],
      },
      {
        headers: [header(1), header(1), header(1), header(1), header(1)],
      },
    ] as unknown as HeaderGroup<PivotDataRow>[];
    const selection = selectionFor({
      status: "closed",
      facility_type: "n/a",
    });

    expect([
      ...computeCellSelectedColDimGroupIndices(
        selection,
        headerGroups,
        rowDimensionNames,
      ),
    ]).toEqual([3]);
  });

  it("keeps sibling measures highlighted within the same full column dimension path", () => {
    const headerGroups = [
      {
        headers: [
          header(1),
          header(6, { status: "closed" }),
          header(2, { status: "open" }),
        ],
      },
      {
        headers: [
          header(1),
          header(2, { status: "closed", facility_type: "null" }),
          header(2, { status: "closed", facility_type: "DS" }),
          header(2, { status: "closed", facility_type: "n/a" }),
          header(2, { status: "open", facility_type: "n/a" }),
        ],
      },
      {
        headers: [
          header(1),
          header(1),
          header(1),
          header(1),
          header(1),
          header(1),
          header(1),
          header(1),
          header(1),
        ],
      },
    ] as unknown as HeaderGroup<PivotDataRow>[];
    const selection = selectionFor(
      { status: "closed", facility_type: "n/a" },
      "closed_na_revenue",
    );

    expect([
      ...computeCellSelectedColDimGroupIndices(
        selection,
        headerGroups,
        rowDimensionNames,
      ),
    ]).toEqual([5, 6]);
  });
});

describe("computeAncestorRowIds", () => {
  // Build a minimal fake TanStack Row tree for a 3-dim table.
  // In nested mode, each row stores its own value under rowDimensions[0].
  function makeRow(
    id: string,
    depth: number,
    originalFirstDimValue: string | null,
    parents: Row<PivotDataRow>[],
  ): Row<PivotDataRow> {
    return {
      id,
      depth,
      original: { A: originalFirstDimValue } as PivotDataRow,
      getParentRows: () => parents,
    } as unknown as Row<PivotDataRow>;
  }

  const rowDimensionNames = ["A", "B", "C"];

  // Tree: A expanded, B expanded. Visible rows: aRow, bRow, c1Row, c2Row.
  const aRow = makeRow("1", 0, "a_val", []);
  const bRow = makeRow("1.0", 1, "b_val", [aRow]);
  const c1Row = makeRow("1.0.0", 2, "c1_val", [aRow, bRow]);
  const c2Row = makeRow("1.0.1", 2, "c2_val", [aRow, bRow]);
  const allRows = [aRow, bRow, c1Row, c2Row];

  it("B header clicked with C rows visible: B's own id is NOT in ancestor set", () => {
    // Simulate a B row-header click. dk built via the same path
    // handleCellClickToFilter uses for nested child rows: only includes
    // dimensions up to the clicked row's depth.
    const dkB = ["a_val", "b_val"].join("\0");
    const rowHeaders = new Map([
      [
        dkB,
        {
          dimKey: dkB,
          dimValues: { A: "a_val", B: "b_val" },
          columnId: "B",
        },
      ],
    ]);
    const selection = buildClickSelection(rowHeaders, new Map(), new Set());

    const ids = computeAncestorRowIds(selection, allRows, rowDimensionNames);

    // A's rowId "1" should be in the set (A is an ancestor of B).
    expect(ids.has("1")).toBe(true);
    // B's own rowId "1.0" must NOT be in the set — B is the clicked row.
    expect(ids.has("1.0")).toBe(false);
  });

  it("B header clicked and a C row has a null leaf value: keys do not collide", () => {
    // Real-world shape: some C rows have a null firstDim value (e.g., a
    // landmark is null for a given city+agency pair). The dimKey for
    // such a child row must remain distinct from its B parent's dk so
    // B's own selection does not bleed into a null-valued descendant.
    const c1RowNullLeaf = makeRow("1.0.0", 2, null, [aRow, bRow]);
    const rows = [aRow, bRow, c1RowNullLeaf, c2Row];

    const dkB = ["a_val", "b_val"].join("\0");
    const rowHeaders = new Map([
      [
        dkB,
        {
          dimKey: dkB,
          dimValues: { A: "a_val", B: "b_val" },
          columnId: "B",
        },
      ],
    ]);
    const selection = buildClickSelection(rowHeaders, new Map(), new Set());

    // Keys at different depths must never collide, even when a deeper
    // row's leaf value is null.
    expect(nestedDimKeyFromRow(c1RowNullLeaf, rowDimensionNames)).not.toBe(dkB);

    const ids = computeAncestorRowIds(selection, rows, rowDimensionNames);

    expect(ids.has("1")).toBe(true);
    expect(ids.has("1.0")).toBe(false);
  });

  it("A (depth-0) header clicked: a depth-1 child with null value is not selected", () => {
    const bRowNull = makeRow("1.0", 1, null, [aRow]);

    const dkA = nestedDimKeyFromRow(aRow, rowDimensionNames); // depth 0
    const dkBNull = nestedDimKeyFromRow(bRowNull, rowDimensionNames); // depth 1, null
    expect(dkA).not.toBe(dkBNull);

    const rowHeaders = new Map([
      [
        dkA,
        {
          dimKey: dkA,
          dimValues: { A: "a_val" },
          columnId: "A",
        },
      ],
    ]);
    const selection = buildClickSelection(rowHeaders, new Map(), new Set());

    expect(selection.isRowHeaderSelected(dkA)).toBe(true);
    expect(selection.isRowHeaderSelected(dkBNull)).toBe(false);
  });
});
