import type { TDDCellData } from "./types";

export function transposeArray(
  arr: TDDCellData[][],
  rowCount: number,
  columnCount: number,
) {
  const columnarBody: TDDCellData[][] = [];

  if (rowCount === 0 || columnCount === 0) return [];

  // Check if transposition is possible
  if (arr.length !== rowCount) {
    console.warn(
      `The actual row count (${arr.length}) does not match the expected row count (${rowCount}).`,
    );
    return [];
  }

  for (let i = 0; i < rowCount; i++) {
    if (arr[i].length !== columnCount) {
      console.warn(
        `Row ${i} length (${arr[i].length}) does not match the expected column count (${columnCount}).`,
      );
      return [];
    }
  }

  for (let i = 0; i < columnCount; i++) {
    const column: TDDCellData[] = [];
    for (let j = 0; j < rowCount; j++) {
      column.push(arr[j][i]);
    }
    columnarBody.push(column);
  }

  return columnarBody;
}

const BG_BASE = "bg-surface-base";
const BG_HIGHLIGHTED = "bg-surface-hover/50";
const BG_DOUBLE_HIGHLIGHTED = "bg-surface-hover";

export function getClassForCell(
  _palette: "fixed" | "default",
  highlightedRow: number | undefined,
  highlightedColStart: number | undefined,
  highlightedColEnd: number | undefined,
  rowIdx: number,
  colIdx: number,
) {
  const isRowHighlighted = highlightedRow === rowIdx;
  const isColHighlighted =
    highlightedColStart !== undefined &&
    highlightedColEnd !== undefined &&
    colIdx >= highlightedColStart &&
    colIdx <= highlightedColEnd;

  if (isRowHighlighted && isColHighlighted) return BG_DOUBLE_HIGHLIGHTED;
  if (isRowHighlighted || isColHighlighted) return BG_HIGHLIGHTED;
  return BG_BASE;
}
