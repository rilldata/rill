import { QueryKey, useQueryClient } from "@tanstack/svelte-query";
import type { PivotPos } from "./types";
import { merge2DBlocks, mergeBlocks } from "./util";
import { derived, writable } from "svelte/store";
import type { CreateQueriesResult } from "@tanstack/svelte-query/build/lib/createQueries";

function transpose2DArray(matrix) {
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

export function createBodyDataProvider(
  queries: CreateQueriesResult<any>,
  lookupFn: (pos: PivotPos) => QueryKey[]
) {
  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();

  const getData = (pos: PivotPos) => {
    // Splice any cached data into this array. regular-table expects data to be columnar
    const sampleColData = new Array(pos.x1 - pos.x0).fill(null);
    let data = new Array(pos.y1 - pos.y0)
      .fill(null)
      .map(() => sampleColData.slice());
    const keys = lookupFn(pos);
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
    const mergedBlock = {
      block: {
        x: [pos.x0, pos.x1],
        y: [pos.y0, pos.y1],
      },
      // regular-table expects data to be columnar, so transpose it
      // TODO: This step can probably be combined with the previous merging step to save time
      // alternatively, the data can be transposed at the time its fetched
      data: data.length > 0 ? transpose2DArray(data) : data,
    };

    return mergedBlock ?? null;
  };

  // Share latest queries set
  const queriesStore = writable(queries);
  const setQueries = (queries) => {
    queriesStore.set(queries);
  };

  const latestQueries: CreateQueriesResult<
    { block: [number, number]; data: any[] }[]
  > = derived(queriesStore, ($queries, set) => {
    if ($queries) $queries.subscribe(set);
    else return undefined;
  });

  const data = derived(latestQueries, ($queries) => {
    if ($queries) return merge2DBlocks($queries.map((q) => q.data));
    return undefined;
  });

  return {
    getData,
    data,
    setQueries,
  };
}

export type BodyDataProvider = ReturnType<typeof createBodyDataProvider>;
