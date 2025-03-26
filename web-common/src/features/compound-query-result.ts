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

type CreateQueryResponses<Q> = {
  [K in keyof Q]: Q[K] extends CreateQueryResult<infer U>
    ? U | undefined
    : never;
};

type QueryResults =
  | [CreateQueryResult<any>, ...Array<CreateQueryResult<any>>]
  | Array<CreateQueryResult<any>>;
export function getCompoundQuery<Queries extends QueryResults, T>(
  queries: Queries,
  getter: (data: CreateQueryResponses<Queries>) => T,
) {
  return derived(queries, ($queries) => {
    const someQueryFetching = $queries.some((q) => q.isFetching);
    const someQueryLoading = $queries.some((q) => q.isLoading);
    const errors = $queries.filter((q) => q.isError).map((q) => q.error);
    const data = getter(
      $queries.map((query) => query.data) as CreateQueryResponses<Queries>,
    );

    return {
      data,
      // TODO: merge multiple errors
      error: errors[0],
      isFetching: someQueryFetching,
      isLoading: someQueryLoading,
    };
  });
}
