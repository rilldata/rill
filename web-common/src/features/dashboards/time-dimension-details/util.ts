export function transposeArray(arr, rowCount, columnCount) {
  const columnarBody = [];

  for (let i = 0; i < columnCount; i++) {
    const column = [];
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
      base: "bg-blue-50",
      highlighted: "bg-blue-100",
      doubleHighlighted: "bg-blue-200",
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
