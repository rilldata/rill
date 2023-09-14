import { QueryKey, useQueryClient } from "@tanstack/svelte-query";
import type { PivotPos } from "./types";
import { mergeBlocks } from "./util";
import { derived, writable } from "svelte/store";
import type { CreateQueriesResult } from "@tanstack/svelte-query/build/lib/createQueries";

export function createColumnDataProvider(
  queries: CreateQueriesResult<any>,
  lookupFn: (pos: PivotPos) => QueryKey[]
) {
  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();

  const getData = (pos: PivotPos) => {
    // Splice any cached data into this array
    let data = new Array(pos.x1 - pos.x0).fill(null);
    const keys = lookupFn(pos);
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
    const mergedBlock = {
      block: [pos.x0, pos.x1],
      data,
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
    return $queries.subscribe(set);
  });

  const data = derived(latestQueries, ($queries) => {
    return mergeBlocks($queries.map((q) => q.data));
  });

  return {
    getData,
    data,
    setQueries,
  };
}

export type ColumnDataProvider = ReturnType<typeof createColumnDataProvider>;
