import type { PivotPos } from "./types";

export function range(x0: number, x1: number, f: (x: number) => any) {
  return Array.from(Array(x1 - x0).keys()).map((x) => f(x + x0));
}

export const isEmptyPos = (pos: PivotPos) =>
  pos.x0 === 0 && pos.x1 === 0 && pos.y0 === 0 && pos.y1 === 0;

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

export function transpose2DArray(matrix: any[][]) {
  const numRows = matrix[0].length;
  const numCols = matrix.length;
  const transposed = new Array(numRows)
    .fill(null)
    .map(() => new Array(numCols));

  for (let i = 0; i < matrix.length; i++) {
    for (let j = 0; j < matrix[i].length; j++) {
      transposed[j][i] = matrix[i][j];
    }
  }

  return transposed;
}

export function createLoadingCell() {
  return {
    isLoader: true,
    value: `<div class="h-4 bg-gray-100 rounded loading-cell" style="width: 100%; min-width: 32px;"/>`,
  };
}
