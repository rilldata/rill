import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function createQueryClient() {
  const isLocal = !get(featureFlags)?.readOnly;
  return new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
        networkMode: isLocal ? "always" : "online",
      },
    },
  });
}
