export function transposeArray(arr, rowCount, columnCount) {
  const columnarBody = [];

  for (let i = 0; i < columnCount; i++) {
    const column = [];
    for (let j = 0; j < rowCount; j++) {
      try {
        column.push(arr[j][i]);
      } catch (e) {
        column.push(null);
        console.error(
          `failed to access arr[${j}][${i}] during transpose of array ${arr}; see issue https://github.com/rilldata/rill/issues/3989`,
          e,
        );
      }
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
