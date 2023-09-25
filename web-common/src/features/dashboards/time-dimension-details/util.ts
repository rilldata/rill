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
