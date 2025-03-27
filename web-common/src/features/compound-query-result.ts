import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function getCompoundQuery<R, T>(
  queries: CreateQueryResult<R, HTTPError>[],
  getter: (data: (R | undefined)[]) => T,
) {
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
