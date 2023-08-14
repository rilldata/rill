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
