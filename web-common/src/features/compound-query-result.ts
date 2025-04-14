import { derived, type Readable } from "svelte/store";
import type {
  CreateQueryResult,
  QueryObserverResult,
} from "@tanstack/svelte-query";

export type SupportedCompoundQueryResult<R = any, E = any> =
  | CreateQueryResult<R, E>
  | CompoundQueryResult<R>;

export type CompoundQueryResult<T> = Readable<
  Pick<QueryObserverResult, "error" | "isFetching" | "isLoading"> & {
    data?: T;
  }
>;

type ExtractQueryData<T> =
  T extends CreateQueryResult<infer R, any>
    ? R | undefined
    : T extends CompoundQueryResult<infer R>
      ? R | undefined
      : never;

export function getCompoundQuery<
  T extends readonly SupportedCompoundQueryResult[],
  R,
>(
  queries: [...T],
  getter: (data: { [K in keyof T]: ExtractQueryData<T[K]> }) => R,
): CompoundQueryResult<R> {
  return derived(queries, ($queries) => {
    const someQueryFetching = $queries.some((q) => q.isFetching);
    const someQueryLoading = $queries.some((q) => q.isLoading);
    const errors = $queries.filter((q) => !!q.error).map((q) => q.error);

    const data = getter(
      $queries.map((q) => q.data) as {
        [K in keyof T]: ExtractQueryData<T[K]>;
      },
    );

    return {
      data,
      error: errors[0],
      isFetching: someQueryFetching,
      isLoading: someQueryLoading,
    };
  });
}
