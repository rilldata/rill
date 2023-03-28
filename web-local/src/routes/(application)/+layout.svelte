<script lang="ts">
  import { RillDeveloperLayout } from "@rilldata/web-common/layout";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { QueryClientProvider } from "@sveltestack/svelte-query";
  import { createQueryClient } from "../../lib/svelte-query/globalQueryClient";

  const queryClient = createQueryClient();
</script>

<QueryClientProvider client={queryClient}>
  <RuntimeProvider host={RuntimeUrl} instanceId="default">
    <RillDeveloperLayout>
      <slot />
    </RillDeveloperLayout>
  </RuntimeProvider>
</QueryClientProvider>
