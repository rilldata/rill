import { onMount } from "svelte";
import { queryClient } from "./globalQueryClient";

/**
 * Cancels all in-flight queries when the browser tab becomes hidden.
 * This prevents expensive server-side queries from continuing when users switch tabs.
 *
 * Call this function from the root layout component.
 */
export function cancelQueriesOnHidden() {
  onMount(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        queryClient.cancelQueries();
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () =>
      document.removeEventListener("visibilitychange", handleVisibilityChange);
  });
}
