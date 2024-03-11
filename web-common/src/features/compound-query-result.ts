import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { Readable } from "svelte/store";

/**
 * Temporary type for derived data based on multiple queries.
 * TODO: get rid of this once we move to tanstack v5
 */
export type CompoundQueryResult<T> = Readable<
  Pick<QueryObserverResult, "error" | "isFetching"> & {
    data?: T;
  }
>;
