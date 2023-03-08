<script lang="ts">
  import { RillDeveloperLayout } from "@rilldata/web-common/layout";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { createQueryClient } from "../../lib/svelte-query/globalQueryClient";

  const queryClient = createQueryClient();
</script>

<QueryClientProvider client={queryClient}>
  <RuntimeProvider host="http://localhost:9009" instanceId="default">
    <RillDeveloperLayout>
      <slot />
    </RillDeveloperLayout>
  </RuntimeProvider>
</QueryClientProvider>
