<script lang="ts">
  /*global  RILL_RUNTIME_URL */
  import { RillDeveloperLayout } from "@rilldata/web-common/layout";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { createQueryClient } from "../../lib/svelte-query/globalQueryClient";

  const queryClient = createQueryClient();

  // Note: this right hand side is set by Vite at build time.
  const rillRuntimeURL = RILL_RUNTIME_URL;
</script>

/*global RILL_RUNTIME_URL */

<QueryClientProvider client={queryClient}>
  <RuntimeProvider host={rillRuntimeURL} instanceId="default">
    <RillDeveloperLayout>
      <slot />
    </RillDeveloperLayout>
  </RuntimeProvider>
</QueryClientProvider>
