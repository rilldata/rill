import type { PivotPos } from "@rilldata/web-common/features/dashboards/pivot/types";
import { range, transpose2DArray } from "../util";
import type { QueryCache } from "@tanstack/svelte-query";
import {
  getBodyKeysFromPos,
  getColHeaderKeysFromPos,
  getRowHeaderKeysFromPos,
} from "./query-keys";

export function createRowHeaderDataGetter({
  getConfig,
  cache,
}: {
  getConfig: () => any;
  cache: QueryCache;
}) {
  return (pos: PivotPos) => {
    // Splice any cached data into this array
    let data = new Array(pos.y1 - pos.y0).fill(null);
    const keys = getRowHeaderKeysFromPos(pos, getConfig());
    keys.forEach((key) => {
      const cachedBlock = cache.find(key)?.state?.data;
      if (cachedBlock) {
        const b = cachedBlock.block;
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
      }
    });
    return data;
  };
}

export function createColumnHeaderDataGetter({
  getConfig,
  cache,
}: {
  getConfig: () => any;
  cache: QueryCache;
}) {
  return (pos: PivotPos) => {
    // Splice any cached data into this array
    let data = new Array(pos.x1 - pos.x0).fill(null);
    const keys = getColHeaderKeysFromPos(pos, getConfig());
    keys.forEach((key) => {
      const cachedBlock = cache.find(key)?.state?.data as
        | {
            block: [number, number];
            data: any[];
          }
        | undefined;
      if (cachedBlock) {
        const b = cachedBlock.block;
        const startingValue = Math.max(b[0], pos.x0);
        const startingValueLocationInBlock = startingValue - b[0];
        const endingValue = Math.min(b[1], pos.x1);
        const endingValueLocationInBlock = endingValue - b[0];
        const valuesToInclude = cachedBlock.data.slice(
          startingValueLocationInBlock,
          endingValueLocationInBlock
        );
        const targetStartPt = Math.max(b[0], pos.x0) - pos.x0;
        data.splice(targetStartPt, valuesToInclude.length, ...valuesToInclude);
      }
    });
    return data;
  };
}

const MOCK_BODY_DATA = range(0, 1000, (x) =>
  range(0, 1000, (y) => "$" + (Math.random() * 10).toFixed(2))
);

export function createBodyDataGetter({
  getConfig,
  cache,
}: {
  getConfig: () => any;
  cache: QueryCache;
}) {
  return (pos: PivotPos) => {
    // Splice any cached data into this array. regular-table expects data to be columnar
    const sampleColData = new Array(pos.x1 - pos.x0).fill(null);
    let data = new Array(pos.y1 - pos.y0)
      .fill(null)
      .map(() => sampleColData.slice());
    const keys = getBodyKeysFromPos(pos, getConfig());
    keys.forEach((key) => {
      const cachedBlock = cache.find(key)?.state?.data as
        | {
            block: {
              x: [number, number];
              y: [number, number];
            };
            data: any[];
          }
        | undefined;
      if (cachedBlock) {
        const rowBlock = cachedBlock.block.y;
        const colBlock = cachedBlock.block.x;
        const targetStartRowIndex = Math.max(rowBlock[0], pos.y0);
        const targetStartRowIndexInBlock = targetStartRowIndex - rowBlock[0];
        const targetEndRowIndex = Math.min(rowBlock[1], pos.y1);
        const targetEndRowIndexInBlock = targetEndRowIndex - rowBlock[0];
        for (
          var i = targetStartRowIndexInBlock;
          i < targetEndRowIndexInBlock;
          i++
        ) {
          const row = cachedBlock.data[i];

          // Determine target columns
          const targetStartColIndex = Math.max(colBlock[0], pos.x0);
          const targetStartColIndexInBlock = targetStartColIndex - colBlock[0];
          const targetEndColIndex = Math.min(colBlock[1], pos.x1);
          const targetEndColIndexInBlock = targetEndColIndex - colBlock[0];
          const colsToInclude = row.slice(
            targetStartColIndexInBlock,
            targetEndColIndexInBlock
          );
          // Splice them in
          const mergedDataRowIndex = i + rowBlock[0] - pos.y0;
          const targetColIndexInSource = targetStartColIndex - pos.x0;
          data[mergedDataRowIndex].splice(
            targetColIndexInSource,
            colsToInclude.length,
            ...colsToInclude
          );
        }
      }
    });
    // regular-table expects data to be columnar, so transpose it
    // TODO: This step can probably be combined with the previous merging step to save time
    // alternatively, the data can be transposed at the time its fetched
    return data.length > 0 ? transpose2DArray(data) : data;
  };
}
