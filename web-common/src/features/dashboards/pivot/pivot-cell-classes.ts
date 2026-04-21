/**
 * Pure functions that compute boolean flags for pivot table cell, row,
 * and header styling.
 *
 * Returns plain objects of booleans so callers can use Svelte's
 * scoped `class:` directives directly
 */

// ---- Flat table ----

export interface FlatRowContext {
  isSelected: boolean;
  hasSelection: boolean;
  hasClickedCell: boolean;
  effectiveDimIdx: number;
}

export interface FlatRowState {
  selectedRow: boolean;
  dimmedRow: boolean;
}

export interface FlatCellContext {
  isActive: boolean;
  isClicked: boolean;
  /** Index into rowDimensionNames; -1 for measure columns */
  colDimIdx: number;
  effectiveDimIdx: number;
  lastDimIdx: number;
  isTotalsRow: boolean;
  canShowDataViewer: boolean;
  enableClickToFilter: boolean;
  hasValue: boolean;
}

export interface FlatCellState {
  activeCell: boolean;
  selectedCell: boolean;
  selectedContextCell: boolean;
  mutedCell: boolean;
  interactiveCell: boolean;
}

/**
 * Compute the effective dimension index for a flat table row.
 * This determines the "depth" at which the user clicked a dimension cell,
 * which in turn drives muted/context styling for other dimension cells.
 */
export function computeEffectiveDimIdx(
  hasClickedCell: boolean,
  clickedDimIdx: number,
  lastDimIdx: number,
  isSelected: boolean,
  maxFilteredDimensionIndex: number,
): number {
  if (hasClickedCell) {
    return clickedDimIdx >= 0 ? clickedDimIdx : lastDimIdx;
  }
  if (isSelected) {
    return maxFilteredDimensionIndex;
  }
  return -1;
}

/** Boolean flags for a flat table <tr> */
export function flatRowState(ctx: FlatRowContext): FlatRowState {
  return {
    selectedRow:
      ctx.isSelected && !ctx.hasClickedCell && ctx.effectiveDimIdx < 0,
    dimmedRow: ctx.hasSelection && !ctx.isSelected && !ctx.hasClickedCell,
  };
}

/** Boolean flags for a flat table <td> (selection-related only) */
export function flatCellState(ctx: FlatCellContext): FlatCellState {
  return {
    activeCell: ctx.isActive,
    selectedCell: ctx.isClicked,
    selectedContextCell:
      ctx.effectiveDimIdx >= 0 &&
      (ctx.colDimIdx >= 0
        ? ctx.colDimIdx <= ctx.effectiveDimIdx
        : ctx.effectiveDimIdx === ctx.lastDimIdx),
    mutedCell:
      (ctx.effectiveDimIdx >= 0 && ctx.colDimIdx > ctx.effectiveDimIdx) ||
      (ctx.colDimIdx === -1 &&
        ctx.effectiveDimIdx >= 0 &&
        ctx.effectiveDimIdx < ctx.lastDimIdx),
    interactiveCell:
      (ctx.isTotalsRow
        ? ctx.canShowDataViewer
        : ctx.canShowDataViewer || ctx.enableClickToFilter) && ctx.hasValue,
  };
}

// ---- Nested table ----

export interface NestedRowContext {
  isSelected: boolean;
  hasSelection: boolean;
  isRowHeaderSelected: boolean;
  hasClickedCell: boolean;
  hasCrossSelection: boolean;
  isAncestorOfSelectedHeader: boolean;
  isShowMore: boolean;
}

export interface NestedRowState {
  showMoreRow: boolean;
  selectedRow: boolean;
  dimmedRow: boolean;
  ancestorOfSelectedRow: boolean;
}

export interface NestedCellContext {
  isActive: boolean;
  isClicked: boolean;
  /** 0 = row header column; > 0 = data/measure columns */
  cellIndex: number;
  hasClickedCell: boolean;
  inHoveredCol: boolean;
  inSelectedCol: boolean;
  isRowHeaderSelected: boolean;
  hasCrossSelection: boolean;
  isAncestorOfSelectedHeader: boolean;
  isTotalsRow: boolean;
  canShowDataViewer: boolean;
  enableClickToFilter: boolean;
}

export interface NestedCellState {
  activeCell: boolean;
  selectedCell: boolean;
  colDimHoverBody: boolean;
  selectedColBody: boolean;
  cellSelectedRowHeader: boolean;
  /** Grey background for data cells on parent rows that partially contain filtered data */
  partialAggregateCell: boolean;
  crossIntersection: boolean;
  crossRowArm: boolean;
  crossColArm: boolean;
  crossSelectedRowHeader: boolean;
  interactiveCell: boolean;
}

export interface NestedHeaderContext {
  isTheHoveredHeader: boolean;
  inHoverRange: boolean;
  isSelfSelected: boolean;
  inSelectedRange: boolean;
  inCellSelectedCol: boolean;
  isAncestorOfSelected: boolean;
}

export interface NestedHeaderState {
  colDimHoverSelf: boolean;
  colDimHoverChild: boolean;
  selectedColHeader: boolean;
  inSelectedColRange: boolean;
  cellSelectedColHeader: boolean;
  ancestorSelectedColHeader: boolean;
}

/** Boolean flags for a nested table <tr> */
export function nestedRowState(ctx: NestedRowContext): NestedRowState {
  return {
    showMoreRow: ctx.isShowMore,
    selectedRow:
      ctx.isSelected && ctx.isRowHeaderSelected && !ctx.hasCrossSelection,
    dimmedRow:
      ctx.hasSelection &&
      !ctx.isSelected &&
      !ctx.hasClickedCell &&
      !ctx.isAncestorOfSelectedHeader,
    ancestorOfSelectedRow: ctx.isAncestorOfSelectedHeader,
  };
}

/** Boolean flags for a nested table <td> (selection-related only) */
export function nestedCellState(ctx: NestedCellContext): NestedCellState {
  return {
    activeCell: ctx.isActive,
    selectedCell: ctx.isClicked,
    colDimHoverBody: ctx.inHoveredCol,
    selectedColBody: ctx.inSelectedCol && !ctx.hasCrossSelection,
    cellSelectedRowHeader: ctx.cellIndex === 0 && ctx.hasClickedCell,
    partialAggregateCell:
      ctx.isAncestorOfSelectedHeader && ctx.cellIndex > 0 && !ctx.hasCrossSelection,
    crossIntersection:
      ctx.hasCrossSelection &&
      ctx.isRowHeaderSelected &&
      ctx.inSelectedCol &&
      ctx.cellIndex > 0,
    crossRowArm:
      ctx.hasCrossSelection &&
      ctx.isRowHeaderSelected &&
      !ctx.inSelectedCol &&
      ctx.cellIndex > 0,
    crossColArm:
      ctx.hasCrossSelection &&
      ctx.inSelectedCol &&
      !ctx.isRowHeaderSelected &&
      !ctx.isAncestorOfSelectedHeader,
    crossSelectedRowHeader:
      ctx.hasCrossSelection && ctx.isRowHeaderSelected && ctx.cellIndex === 0,
    interactiveCell: ctx.isTotalsRow
      ? ctx.canShowDataViewer
      : ctx.canShowDataViewer || ctx.enableClickToFilter,
  };
}

/** Boolean flags for a nested table <th> (column header) */
export function nestedHeaderState(ctx: NestedHeaderContext): NestedHeaderState {
  return {
    colDimHoverSelf: ctx.isTheHoveredHeader,
    colDimHoverChild: ctx.inHoverRange && !ctx.isTheHoveredHeader,
    selectedColHeader: ctx.isSelfSelected,
    inSelectedColRange: ctx.inSelectedRange,
    cellSelectedColHeader: ctx.inCellSelectedCol,
    ancestorSelectedColHeader: ctx.isAncestorOfSelected,
  };
}
