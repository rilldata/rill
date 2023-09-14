import { getBlock } from "../time-dimension-details/util";
import type { PivotPos } from "./types";

export function range(x0: number, x1: number, f: (x: number) => any) {
  return Array.from(Array(x1 - x0).keys()).map((x) => f(x + x0));
}

export const isEmptyPos = (pos: PivotPos) =>
  pos.x0 === 0 && pos.x1 === 0 && pos.y0 === 0 && pos.y1 === 0;

const blockSize = 50;

export function createRowDataGetter(cache) {
  const getData = (pos: PivotPos) => {
    // get multiple blocks and merge them
    let block = getBlock(blockSize, pos.y0, pos.y0);
    const blocks = [block];
    while (block[1] < pos.y1) {
      // HACK: fetch intermediate pages too
      block = getBlock(
        blockSize,
        block[0] + blockSize / 2,
        block[1] + blockSize / 2
      );
      blocks.push(block);
    }

    let data = new Array(pos.y1 - pos.y0).fill(null);

    blocks.forEach((b) => {
      const cachedBlock = cache.find(["example-pivot-row-header", b[0], b[1]])
        ?.state?.data;
      if (cachedBlock) {
        const startingValue = Math.max(b[0], pos.y0);
        const startingValueLocationInBlock = startingValue - b[0];
        const endingValue = Math.min(b[1], pos.y1);
        const endingValueLocationInBlock = endingValue - b[0];
        const valuesToInclude = cachedBlock.data.slice(
          startingValueLocationInBlock,
          endingValueLocationInBlock
        );
        const targetStartPt = Math.max(b[0], pos.y0) - pos.y0;
        data.splice(targetStartPt, valuesToInclude.length, ...valuesToInclude);
      } else {
        // how to know what options?? also won't work outside component :/
        // queryClient.ensureQueryData({
        //   query
        // })
      }
    });

    const mergedBlock = {
      block: [pos.y0, pos.y1],
      data,
    };
    return mergedBlock ?? null;
  };
  return getData;
}

export function mergeBlocks(blocks) {
  let mergedData = null;
  const results = blocks.slice(0);
  let currentData = results.shift();
  while (currentData) {
    mergedData = mergedData ?? { block: [Infinity, -Infinity], data: [] };
    mergedData.block[0] = Math.min(mergedData.block[0], currentData.block[0]);
    mergedData.block[1] = Math.max(mergedData.block[1], currentData.block[1]);
    mergedData.data = mergedData.data.concat(currentData.data);
    currentData = results.shift();
  }
  return mergedData;
}

export function merge2DBlocks(blocks) {
  let mergedData = null;
  if (blocks.every((b) => !b)) return mergedData;

  // const fullArea = {
  //   x0: Math.min(...blocks.map((b) => b.block.x[0])),
  //   x1: Math.max(...blocks.map((b) => b.block.x[1])),
  //   y0: Math.min(...blocks.map((b) => b.block.y[0])),
  //   y1: Math.max(...blocks.map((b) => b.block.y[1])),
  // };
  // console.log("blocks", blocks, fullArea);
  // do we even need to merge this stuff, we never use it...
  // const sampleColData = new Array(pos.x1 - pos.x0).fill(null);
  //   let data = new Array(pos.y1 - pos.y0)
  //     .fill(null)
  //     .map(() => sampleColData.slice());
  const results = blocks.slice(0);
  let currentData = results.shift();
  while (currentData) {
    mergedData = mergedData ?? { block: [Infinity, -Infinity], data: [] };
    mergedData.block[0] = Math.min(mergedData.block[0], currentData.block[0]);
    mergedData.block[1] = Math.max(mergedData.block[1], currentData.block[1]);
    mergedData.data = mergedData.data.concat(currentData.data);
    currentData = results.shift();
  }
  return mergedData;
}
