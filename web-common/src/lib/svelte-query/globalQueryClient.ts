import { type Query, QueryCache, QueryClient } from "@tanstack/svelte-query";
import { isNetworkError } from "../errors";

const MaxNetworkErrorRetries = 2;

type QueryErrorListener = (error: unknown, query: Query) => void;

const queryErrorListeners = new Set<QueryErrorListener>();

/**
 * Calls the listener with the error of every failed query.
 * Use this instead of assigning the query cache's `onError`, which would replace the other listeners.
 * Returns a function that removes the listener.
 */
export function onQueryError(listener: QueryErrorListener): () => void {
  queryErrorListeners.add(listener);
  return () => queryErrorListeners.delete(listener);
}

export function createQueryClient() {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (error, query) =>
        queryErrorListeners.forEach((l) => l(error, query as Query)),
    }),
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: (failureCount, error) =>
          isNetworkError(error) && failureCount < MaxNetworkErrorRetries,
        retryDelay: (failureCount) => Math.min(1000 * 2 ** failureCount, 4000),
        networkMode: "always",
      },
      mutations: {
        networkMode: "always",
      },
    },
  });
}

export const queryClient = createQueryClient();
