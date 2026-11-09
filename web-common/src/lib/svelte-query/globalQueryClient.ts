import { QueryCache, QueryClient } from "@tanstack/svelte-query";
import { isNetworkError } from "../errors";

const MaxNetworkErrorRetries = 2;

const queryErrorListeners = new Set<(error: unknown) => void>();

/**
 * Calls the listener with the error of every failed query.
 * Returns a function that removes the listener.
 */
export function onQueryError(listener: (error: unknown) => void): () => void {
  queryErrorListeners.add(listener);
  return () => queryErrorListeners.delete(listener);
}

export function createQueryClient() {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (error) => queryErrorListeners.forEach((l) => l(error)),
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
