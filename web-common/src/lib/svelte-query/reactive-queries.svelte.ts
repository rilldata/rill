import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { ConnectError } from "@connectrpc/connect";
import {
  createQueries,
  type CreateQueryOptions,
  type QueryObserverResult,
} from "@tanstack/svelte-query";
import { fromStore, writable } from "svelte/store";

export type ReactiveQueriesResult<TData> = {
  data: TData;
  error: ConnectError | null;
  isFetching: boolean;
  isLoading: boolean;
};

/**
 * Runs one query per entry in `getQueryOptions` and combines their results into a single reactive
 * result.
 *
 * The option list is rebuilt whenever the state it reads changes, so queries are added, removed and
 * refetched without recreating the observer. `getData` receives the data of each query in the order
 * of the option list.
 *
 * Must be called during component init, since it owns an effect.
 */
export function createReactiveQueries<TQueryFnData, TQueryData, TData>(
  getQueryOptions: () => CreateQueryOptions<
    TQueryFnData,
    ConnectError,
    TQueryData
  >[],
  getData: (data: (TQueryData | undefined)[]) => TData,
): { readonly current: ReactiveQueriesResult<TData> } {
  // svelte-query takes the options as a store rather than as a rune. Pushing into the store lets a
  // single observer follow the option list, instead of recreating the queries on every change.
  const queriesStore = writable(getQueryOptions());
  $effect(() => {
    queriesStore.set(getQueryOptions());
  });

  return fromStore(
    createQueries(
      {
        queries: queriesStore,
        combine: (anyRes: any) => {
          // TODO: fix types to not need this
          const results = anyRes as QueryObserverResult<
            TQueryData,
            ConnectError
          >[];
          return {
            data: getData(results.map((result) => result.data)),
            error: results.find((result) => !!result.error)?.error ?? null,
            isFetching: results.some((result) => result.isFetching),
            isLoading: results.some((result) => result.isLoading),
          };
        },
      },
      queryClient,
    ),
  );
}
