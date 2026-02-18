import { QueryClient } from "@tanstack/svelte-query";

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
        networkMode: "always",
      },
      mutations: {
        networkMode: "always",
      },
    },
  });
}

export const queryClient = createQueryClient();

/**
 * Getter function for the query client.
 * Required by @tanstack/svelte-query v6 when passing queryClient as a parameter.
 */
export const getQueryClient = () => queryClient;
