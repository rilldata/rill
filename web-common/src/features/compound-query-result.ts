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
  Pick<QueryObserverResult, "error" | "isFetching"> & {
    data?: T;
  }
>;

export function getCompoundAggregationQuery<R, T>(
  queries: CreateQueryResult<R, HTTPError>[],
  getter: (data: (R | undefined)[]) => T,
): CompoundQueryResult<T> {
  return derived(queries, ($queries) => {
    const someQueryFetching = $queries.some((q) => q.isFetching);
    if (someQueryFetching) {
      return {
        data: undefined,
        error: undefined,
        isFetching: true,
      };
    }
    const errors = $queries.filter((q) => q.isError).map((q) => q.error);
    if (errors.length > 0) {
      return {
        data: undefined,
        // TODO: merge multiple errors
        error: errors[0]?.response?.data.message,
        isFetching: false,
      };
    }

    const rawData = $queries.map((query) => query.data);
    return {
      data: getter(rawData),
      error: undefined,
      isFetching: false,
    };
  });
}
