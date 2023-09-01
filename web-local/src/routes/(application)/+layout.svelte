<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
  import { startWatchFilesClient } from "@rilldata/web-common/features/entity-management/watch-files-client";
  import { startWatchResourcesClient } from "@rilldata/web-common/features/entity-management/watch-resources-client";
  import RillDeveloperLayout from "@rilldata/web-common/layout/RillDeveloperLayout.svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { createQueryClient } from "../../lib/svelte-query/globalQueryClient";

  const queryClient = createQueryClient();

  beforeNavigate(retainFeaturesFlags);

  onMount(() => {
    const stopWatchFilesClient = startWatchFilesClient(queryClient);
    const stopWatchResourcesClient = startWatchResourcesClient(queryClient);
    return () => {
      stopWatchFilesClient();
      stopWatchResourcesClient();
    };
  });
</script>

<QueryClientProvider client={queryClient}>
  <RuntimeProvider host={RuntimeUrl} instanceId="default">
    <RillDeveloperLayout>
      <slot />
    </RillDeveloperLayout>
  </RuntimeProvider>
</QueryClientProvider>
