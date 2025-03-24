import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

/**
 * Temporary type for derived data based on multiple queries.
 * TODO: get rid of this once we move to tanstack v5
 */
export type CompoundQueryResult<T> = Readable<
  Pick<QueryObserverResult, "error" | "isFetching" | "isLoading"> & {
    data?: T;
  }
>;

export function getCompoundQuery<R, T>(
  queries: CreateQueryResult<R, HTTPError>[],
  getter: (data: (R | undefined)[]) => T,
): CompoundQueryResult<T> {
  return derived(queries, ($queries) => {
    const someQueryFetching = $queries.some((q) => q.isFetching);
    const someQueryLoading = $queries.some((q) => q.isLoading);
    const errors = $queries.filter((q) => q.isError).map((q) => q.error);
    const data = getter($queries.map((query) => query.data));

    return {
      data,
      // TODO: merge multiple errors
      error: errors[0]?.response?.data.message,
      isFetching: someQueryFetching,
      isLoading: someQueryLoading,
    };
  });
}
