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

export function getClassForCell(
  palette: "fixed" | "scrubbed" | "default",
  highlightedRow,
  highlightedCol,
  rowIdx,
  colIdx,
) {
  const bgColors = {
    fixed: {
      base: "bg-slate-50",
      highlighted: "bg-slate-100",
      doubleHighlighted: "bg-slate-200",
    },
    scrubbed: {
      base: "bg-primary-50",
      highlighted: "bg-primary-100",
      doubleHighlighted: "bg-primary-200",
    },
    default: {
      base: "bg-white",
      highlighted: "bg-gray-100",
      doubleHighlighted: "bg-gray-200",
    },
  };

  // Determine background color based on store
  const isRowHighlighted = highlightedRow === rowIdx;
  const isColHighlighted = highlightedCol === colIdx;
  const isHighlighted = isRowHighlighted || isColHighlighted;
  const isDoubleHighlighted = isRowHighlighted && isColHighlighted;

  let colorName = bgColors[palette].base;
  if (isDoubleHighlighted) colorName = bgColors[palette].doubleHighlighted;
  else if (isHighlighted) colorName = bgColors[palette].highlighted;

  return colorName;
}
