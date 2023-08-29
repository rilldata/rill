<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
  import { filesRequests } from "@rilldata/web-common/features/reconcile/files-requests";
  import { resourcesRequests } from "@rilldata/web-common/features/reconcile/resources-requests";
  import RillDeveloperLayout from "@rilldata/web-common/layout/RillDeveloperLayout.svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import type { Unsubscriber } from "svelte/store";
  import { createQueryClient } from "../../lib/svelte-query/globalQueryClient";

  const queryClient = createQueryClient();

  beforeNavigate(retainFeaturesFlags);

  let filesUnsubscribe: Unsubscriber;
  $: {
    if (filesUnsubscribe) filesUnsubscribe();
    filesUnsubscribe = filesRequests(queryClient);
  }

  let resourcesUnsubscribe: Unsubscriber;
  $: {
    if (resourcesUnsubscribe) resourcesUnsubscribe();
    resourcesUnsubscribe = resourcesRequests(queryClient);
  }
</script>

<QueryClientProvider client={queryClient}>
  <RuntimeProvider host={RuntimeUrl} instanceId="default">
    <RillDeveloperLayout>
      <slot />
    </RillDeveloperLayout>
  </RuntimeProvider>
</QueryClientProvider>
