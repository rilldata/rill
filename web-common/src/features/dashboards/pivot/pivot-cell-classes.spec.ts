import { describe, expect, it } from "vitest";
import {
  computeEffectiveDimIdx,
  flatCellState,
  flatRowState,
  nestedCellState,
  nestedHeaderState,
  nestedRowState,
} from "./pivot-cell-classes";
import { computePivotRowSelection } from "./pivot-row-selection";
import type { PivotDataRow, PivotDataStoreConfig } from "./types";

// ---- computeEffectiveDimIdx ----

describe("computeEffectiveDimIdx", () => {
  it("returns clickedDimIdx when cell is clicked and index is valid", () => {
    expect(computeEffectiveDimIdx(true, 1, 2, false, -1)).toBe(1);
  });

  it("falls back to lastDimIdx when clicked but dimIdx is -1", () => {
    expect(computeEffectiveDimIdx(true, -1, 2, false, -1)).toBe(2);
  });

  it("uses maxFilteredDimensionIndex when row is selected but no cell clicked", () => {
    expect(computeEffectiveDimIdx(false, -1, 2, true, 1)).toBe(1);
  });

  it("returns -1 when nothing is selected", () => {
    expect(computeEffectiveDimIdx(false, -1, 2, false, -1)).toBe(-1);
  });
});

// ---- flatRowState ----

describe("flatRowState", () => {
  it("returns selectedRow when row is selected without cell clicks", () => {
    const result = flatRowState({
      isSelected: true,
      hasSelection: true,
      hasClickedCell: false,
      effectiveDimIdx: -1,
    });
    expect(result.selectedRow).toBe(true);
  });

  it("does not return selectedRow when a cell is clicked", () => {
    const result = flatRowState({
      isSelected: true,
      hasSelection: true,
      hasClickedCell: true,
      effectiveDimIdx: 0,
    });
    expect(result.selectedRow).toBe(false);
  });

  it("returns dimmedRow for unselected rows during active selection", () => {
    const result = flatRowState({
      isSelected: false,
      hasSelection: true,
      hasClickedCell: false,
      effectiveDimIdx: -1,
    });
    expect(result.dimmedRow).toBe(true);
  });

  it("returns all false when no selection is active", () => {
    const result = flatRowState({
      isSelected: false,
      hasSelection: false,
      hasClickedCell: false,
      effectiveDimIdx: -1,
    });
    expect(result.selectedRow).toBe(false);
    expect(result.dimmedRow).toBe(false);
  });
});

// ---- flatCellState ----

describe("flatCellState", () => {
  const base = {
    isActive: false,
    isClicked: false,
    colDimIdx: -1,
    effectiveDimIdx: -1,
    lastDimIdx: 1,
    isTotalsRow: false,
    canShowDataViewer: false,
    enableClickToFilter: false,
    hasValue: true,
  };

  it("returns activeCell when cell is active", () => {
    expect(flatCellState({ ...base, isActive: true }).activeCell).toBe(true);
  });

  it("returns selectedCell when cell is clicked", () => {
    expect(flatCellState({ ...base, isClicked: true }).selectedCell).toBe(true);
  });

  it("returns selectedContextCell for dimension cells left of click", () => {
    const result = flatCellState({
      ...base,
      colDimIdx: 0,
      effectiveDimIdx: 1,
      lastDimIdx: 1,
    });
    expect(result.selectedContextCell).toBe(true);
  });

  it("returns selectedContextCell for measure cols when clicked at last dim", () => {
    const result = flatCellState({
      ...base,
      colDimIdx: -1,
      effectiveDimIdx: 1,
      lastDimIdx: 1,
    });
    expect(result.selectedContextCell).toBe(true);
  });

  it("returns mutedCell for dimension cells right of click", () => {
    const result = flatCellState({
      ...base,
      colDimIdx: 1,
      effectiveDimIdx: 0,
      lastDimIdx: 1,
    });
    expect(result.mutedCell).toBe(true);
  });

  it("returns mutedCell for measure cols when not clicked at last dim", () => {
    const result = flatCellState({
      ...base,
      colDimIdx: -1,
      effectiveDimIdx: 0,
      lastDimIdx: 1,
    });
    expect(result.mutedCell).toBe(true);
  });

  it("returns interactiveCell with enableClickToFilter and value", () => {
    const result = flatCellState({
      ...base,
      enableClickToFilter: true,
      hasValue: true,
    });
    expect(result.interactiveCell).toBe(true);
  });

  it("does not return interactiveCell for totals row without canShowDataViewer", () => {
    const result = flatCellState({
      ...base,
      isTotalsRow: true,
      enableClickToFilter: true,
      hasValue: true,
    });
    expect(result.interactiveCell).toBe(false);
  });

  it("returns all false for a plain unselected cell", () => {
    const result = flatCellState(base);
    expect(result.activeCell).toBe(false);
    expect(result.selectedCell).toBe(false);
    expect(result.selectedContextCell).toBe(false);
    expect(result.mutedCell).toBe(false);
    expect(result.interactiveCell).toBe(false);
  });
});

// ---- nestedRowState ----

describe("nestedRowState", () => {
  const base = {
    isSelected: false,
    hasSelection: false,
    isRowHeaderSelected: false,
    hasClickedCell: false,
    hasCrossSelection: false,
    isAncestorOfSelectedHeader: false,
    isShowMore: false,
  };

  it("returns showMoreRow for show more rows", () => {
    expect(nestedRowState({ ...base, isShowMore: true }).showMoreRow).toBe(
      true,
    );
  });

  it("returns selectedRow when row header is selected without cross-selection", () => {
    const result = nestedRowState({
      ...base,
      isSelected: true,
      isRowHeaderSelected: true,
    });
    expect(result.selectedRow).toBe(true);
  });

  it("does not return selectedRow during cross-selection", () => {
    const result = nestedRowState({
      ...base,
      isSelected: true,
      isRowHeaderSelected: true,
      hasCrossSelection: true,
    });
    expect(result.selectedRow).toBe(false);
  });

  it("returns dimmedRow for non-selected rows", () => {
    const result = nestedRowState({
      ...base,
      hasSelection: true,
    });
    expect(result.dimmedRow).toBe(true);
  });

  it("does not dim ancestor rows", () => {
    const result = nestedRowState({
      ...base,
      hasSelection: true,
      isAncestorOfSelectedHeader: true,
    });
    expect(result.dimmedRow).toBe(false);
    expect(result.ancestorOfSelectedRow).toBe(true);
  });
});

// ---- nestedCellState ----

describe("nestedCellState", () => {
  const base = {
    isActive: false,
    isClicked: false,
    cellIndex: 1,
    hasClickedCell: false,
    inHoveredCol: false,
    inSelectedCol: false,
    isRowHeaderSelected: false,
    hasCrossSelection: false,
    isAncestorOfSelectedHeader: false,
    isTotalsRow: false,
    canShowDataViewer: false,
    enableClickToFilter: false,
  };

  it("returns crossIntersection when row+col selected", () => {
    const result = nestedCellState({
      ...base,
      hasCrossSelection: true,
      isRowHeaderSelected: true,
      inSelectedCol: true,
    });
    expect(result.crossIntersection).toBe(true);
    expect(result.crossRowArm).toBe(false);
    expect(result.crossColArm).toBe(false);
  });

  it("returns crossRowArm for selected row, unselected column", () => {
    const result = nestedCellState({
      ...base,
      hasCrossSelection: true,
      isRowHeaderSelected: true,
      inSelectedCol: false,
    });
    expect(result.crossRowArm).toBe(true);
  });

  it("returns crossColArm for selected column, unselected row", () => {
    const result = nestedCellState({
      ...base,
      hasCrossSelection: true,
      inSelectedCol: true,
      isRowHeaderSelected: false,
    });
    expect(result.crossColArm).toBe(true);
  });

  it("does not return crossColArm for ancestor rows", () => {
    const result = nestedCellState({
      ...base,
      hasCrossSelection: true,
      inSelectedCol: true,
      isAncestorOfSelectedHeader: true,
    });
    expect(result.crossColArm).toBe(false);
  });

  it("returns crossSelectedRowHeader for row header in cross-selection", () => {
    const result = nestedCellState({
      ...base,
      cellIndex: 0,
      hasCrossSelection: true,
      isRowHeaderSelected: true,
    });
    expect(result.crossSelectedRowHeader).toBe(true);
    // Row header should NOT get crossIntersection
    expect(result.crossIntersection).toBe(false);
  });

  it("returns cellSelectedRowHeader for row header with clicked cell", () => {
    const result = nestedCellState({
      ...base,
      cellIndex: 0,
      hasClickedCell: true,
    });
    expect(result.cellSelectedRowHeader).toBe(true);
  });

  it("returns colDimHoverBody for cells in hovered column", () => {
    expect(
      nestedCellState({ ...base, inHoveredCol: true }).colDimHoverBody,
    ).toBe(true);
  });

  it("returns selectedColBody without cross-selection", () => {
    expect(
      nestedCellState({ ...base, inSelectedCol: true }).selectedColBody,
    ).toBe(true);
  });

  it("does not return selectedColBody during cross-selection", () => {
    const result = nestedCellState({
      ...base,
      inSelectedCol: true,
      hasCrossSelection: true,
    });
    expect(result.selectedColBody).toBe(false);
  });

  it("returns interactiveCell with enableClickToFilter", () => {
    expect(
      nestedCellState({ ...base, enableClickToFilter: true }).interactiveCell,
    ).toBe(true);
  });

  it("returns all false for a plain cell", () => {
    const result = nestedCellState(base);
    expect(result.activeCell).toBe(false);
    expect(result.selectedCell).toBe(false);
    expect(result.crossIntersection).toBe(false);
    expect(result.crossRowArm).toBe(false);
    expect(result.crossColArm).toBe(false);
    expect(result.interactiveCell).toBe(false);
  });
});

// ---- nestedHeaderState ----

describe("nestedHeaderState", () => {
  const base = {
    isTheHoveredHeader: false,
    inHoverRange: false,
    isSelfSelected: false,
    inSelectedRange: false,
    inCellSelectedCol: false,
    isAncestorOfSelected: false,
  };

  it("returns colDimHoverSelf for hovered header", () => {
    const result = nestedHeaderState({
      ...base,
      isTheHoveredHeader: true,
      inHoverRange: true,
    });
    expect(result.colDimHoverSelf).toBe(true);
    expect(result.colDimHoverChild).toBe(false);
  });

  it("returns colDimHoverChild for child in hover range", () => {
    expect(
      nestedHeaderState({ ...base, inHoverRange: true }).colDimHoverChild,
    ).toBe(true);
  });

  it("returns selectedColHeader when self selected", () => {
    expect(
      nestedHeaderState({ ...base, isSelfSelected: true }).selectedColHeader,
    ).toBe(true);
  });

  it("returns inSelectedColRange when in range", () => {
    expect(
      nestedHeaderState({ ...base, inSelectedRange: true }).inSelectedColRange,
    ).toBe(true);
  });

  it("returns cellSelectedColHeader when has cell selection", () => {
    expect(
      nestedHeaderState({ ...base, inCellSelectedCol: true })
        .cellSelectedColHeader,
    ).toBe(true);
  });

  it("returns ancestorSelectedColHeader when ancestor of selected", () => {
    expect(
      nestedHeaderState({ ...base, isAncestorOfSelected: true })
        .ancestorSelectedColHeader,
    ).toBe(true);
  });

  it("returns all false with no state", () => {
    const result = nestedHeaderState(base);
    expect(result.colDimHoverSelf).toBe(false);
    expect(result.colDimHoverChild).toBe(false);
    expect(result.selectedColHeader).toBe(false);
    expect(result.inSelectedColRange).toBe(false);
    expect(result.cellSelectedColHeader).toBe(false);
    expect(result.ancestorSelectedColHeader).toBe(false);
  });
});

describe("3-dimension nested table: row header click styling by depth", () => {
  const config: PivotDataStoreConfig = {
    rowDimensionNames: ["A", "B", "C"],
    measureNames: ["revenue"],
    colDimensionNames: [],
    isFlat: false,
    time: { timeDimension: "", timeStart: undefined, timeEnd: undefined },
  } as unknown as PivotDataStoreConfig;

  const aRow: PivotDataRow = { A: "a1", revenue: 100 };
  const bRow: PivotDataRow = { A: "b1", revenue: 50 };
  const cRow: PivotDataRow = { A: "c1", revenue: 25 };

  function rowSelectionFor(filters: Record<string, string>) {
    const dimensionFilters = new Map<string, Set<string>>();
    for (const [dim, val] of Object.entries(filters)) {
      dimensionFilters.set(dim, new Set([val]));
    }
    return computePivotRowSelection(config, [], dimensionFilters);
  }

  const aClickedFilters = { A: "a1" };
  const bClickedFilters = { A: "a1", B: "b1" };
  const cClickedFilters = { A: "a1", B: "b1", C: "c1" };

  // hasAnySelection is true whenever a row header is selected
  const hasAnySelection = true;
  const hasCrossSelection = false;

  function rowState(opts: {
    row: PivotDataRow;
    depth: number;
    parents: PivotDataRow[];
    isRowHeaderSelected: boolean;
    isAncestorOfSelectedHeader: boolean;
    filters: Record<string, string>;
  }) {
    const rs = rowSelectionFor(opts.filters);
    const filterSelected = rs.isRowSelected(opts.row, opts.depth, opts.parents);
    const hasClickedCell = false;
    const isSelected =
      opts.depth > 0 && hasAnySelection
        ? filterSelected && (opts.isRowHeaderSelected || hasClickedCell)
        : filterSelected;
    return nestedRowState({
      isSelected,
      hasSelection: rs.hasActiveSelection,
      isRowHeaderSelected: opts.isRowHeaderSelected,
      hasClickedCell,
      hasCrossSelection,
      isAncestorOfSelectedHeader: opts.isAncestorOfSelectedHeader,
      isShowMore: false,
    });
  }

  function measureCellState(isAncestorOfSelectedHeader: boolean) {
    return nestedCellState({
      isActive: false,
      isClicked: false,
      cellIndex: 1, // measure column
      hasClickedCell: false,
      inHoveredCol: false,
      inSelectedCol: false,
      isRowHeaderSelected: false,
      hasCrossSelection,
      isAncestorOfSelectedHeader,
      isTotalsRow: false,
      canShowDataViewer: false,
      enableClickToFilter: true,
    });
  }

  it("A clicked: A row gets selectedRow (blue), not partial-aggregate", () => {
    const rs = rowState({
      row: aRow,
      depth: 0,
      parents: [],
      isRowHeaderSelected: true,
      isAncestorOfSelectedHeader: false,
      filters: aClickedFilters,
    });
    expect(rs.selectedRow).toBe(true);
    expect(rs.dimmedRow).toBe(false);
    expect(measureCellState(false).partialAggregateCell).toBe(false);
  });

  it("B clicked: B row gets selectedRow (blue), not partial-aggregate", () => {
    const rs = rowState({
      row: bRow,
      depth: 1,
      parents: [aRow],
      isRowHeaderSelected: true,
      isAncestorOfSelectedHeader: false,
      filters: bClickedFilters,
    });
    expect(rs.selectedRow).toBe(true);
    expect(rs.dimmedRow).toBe(false);
    expect(measureCellState(false).partialAggregateCell).toBe(false);
  });

  it("B clicked: A row is ancestor — partial-aggregate (grey) on measure cells", () => {
    const rs = rowState({
      row: aRow,
      depth: 0,
      parents: [],
      isRowHeaderSelected: false,
      isAncestorOfSelectedHeader: true,
      filters: bClickedFilters,
    });
    expect(rs.ancestorOfSelectedRow).toBe(true);
    expect(rs.selectedRow).toBe(false);
    expect(measureCellState(true).partialAggregateCell).toBe(true);
  });

  it("C clicked: C row gets selectedRow (blue), not partial-aggregate", () => {
    const rs = rowState({
      row: cRow,
      depth: 2,
      parents: [aRow, bRow],
      isRowHeaderSelected: true,
      isAncestorOfSelectedHeader: false,
      filters: cClickedFilters,
    });
    expect(rs.selectedRow).toBe(true);
    expect(rs.dimmedRow).toBe(false);
    expect(measureCellState(false).partialAggregateCell).toBe(false);
  });

  it("C clicked: B row is ancestor — partial-aggregate (grey) on measure cells", () => {
    const rs = rowState({
      row: bRow,
      depth: 1,
      parents: [aRow],
      isRowHeaderSelected: false,
      isAncestorOfSelectedHeader: true,
      filters: cClickedFilters,
    });
    expect(rs.ancestorOfSelectedRow).toBe(true);
    expect(rs.selectedRow).toBe(false);
    expect(measureCellState(true).partialAggregateCell).toBe(true);
  });
});
