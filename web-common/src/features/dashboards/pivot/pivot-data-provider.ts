import {
  CreateQueryResult,
  QueryKey,
  useQueryClient,
} from "@tanstack/svelte-query";
import type { PivotPos } from "./types";
import { getBlock } from "../time-dimension-details/util";

export function createPivotDataProvider(
  query: CreateQueryResult,
  lookupFn: (pos: PivotPos) => QueryKey
) {
  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();

  const blockSize = 50;
  const getData = (pos: PivotPos) => {
    const key = lookupFn(pos);
    const cachedBlock = cache.find(key)?.state?.data;
    return cachedBlock ?? null;
  };

  return {
    getData,
    query,
  };
}

export type PivotDataProvider = ReturnType<typeof createPivotDataProvider>;
