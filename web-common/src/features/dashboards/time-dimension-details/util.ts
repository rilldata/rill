// given area, return block to fetch
export const getBlock = (blockSize: number, start: number, end: number) => {
  // If distance is bigger than possible block, throw an error
  if (end - start > blockSize) {
    throw new Error("Range is too big");
  }
  // Calculate the nearest block to the start
  let startBlock = Math.floor(start / blockSize) * blockSize;
  // Calculate the end of that block
  let endBlock = startBlock + blockSize;

  // If end is not in this block, increment the block by 1/2
  if (end > endBlock) {
    startBlock += blockSize * 0.5;
    endBlock = startBlock + blockSize;
  }
  return [startBlock, endBlock];
};

export const getBlocks = (blockSize: number, start: number, end: number) => {
  const blocks: number[][] = [];
  let block = getBlock(blockSize, start, start);
  blocks.push(block);
  while (block[1] < end) {
    block = getBlock(blockSize, block[0] + blockSize, block[1] + blockSize);
    blocks.push(block);
  }
  return blocks;
};

export const get2DBlocks = ({
  blockSizeX,
  blockSizeY,
  x0,
  x1,
  y0,
  y1,
}: {
  blockSizeX: number;
  blockSizeY: number;
  x0: number;
  x1: number;
  y0: number;
  y1: number;
}) => {
  const rowBlocks = getBlocks(blockSizeY, y0, y1);
  return rowBlocks.flatMap((rowBlock) =>
    getBlocks(blockSizeX, x0, x1).map((colBlock) => ({
      x: colBlock,
      y: rowBlock,
    }))
  );
};

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
  colIdx
) {
  const bgColors = {
    fixed: {
      base: "bg-slate-50",
      highlighted: "bg-slate-100",
      doubleHighlighted: "bg-slate-200",
    },
    scrubbed: {
      base: "!bg-blue-50",
      highlighted: "!bg-blue-100",
      doubleHighlighted: "!bg-blue-200",
    },
    default: {
      base: "!bg-white",
      highlighted: "!bg-gray-100",
      doubleHighlighted: "!bg-gray-200",
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
