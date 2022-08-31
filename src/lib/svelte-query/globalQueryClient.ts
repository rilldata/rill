/**
 * This is temporary until everything is moved to using svelte-query
 */
import { QueryClient } from "@sveltestack/svelte-query";

export let queryClient: QueryClient;
export function createQueryClient() {
  queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: true,
        refetchOnWindowFocus: true, // though `false` is handy for development
        placeholderData: {},
      },
    },
  });
}
