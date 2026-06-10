import { QueryClient } from "@tanstack/svelte-query";
import { isNetworkError } from "../errors";

const MaxNetworkErrorRetries = 2;

export function createQueryClient() {
  return new QueryClient({
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
