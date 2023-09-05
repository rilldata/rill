<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import DevJWTProvider from "@rilldata/web-common/features/dashboards/granular-access-policies/DevJWTProvider.svelte";
  import { retainFeaturesFlags } from "@rilldata/web-common/features/feature-flags";
  import RillDeveloperLayout from "@rilldata/web-common/layout/RillDeveloperLayout.svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { RuntimeUrl } from "@rilldata/web-local/lib/application-state-stores/initialize-node-store-contexts";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { createQueryClient } from "../../lib/svelte-query/globalQueryClient";

  const queryClient = createQueryClient();

  beforeNavigate(retainFeaturesFlags);

  // TODO: re-enable once controller changes are merged.
  // onMount(() => {
  //   const stopWatchFilesClient = startWatchFilesClient(queryClient);
  //   const stopWatchResourcesClient = startWatchResourcesClient(queryClient);
  //   return () => {
  //     stopWatchFilesClient();
  //     stopWatchResourcesClient();
  //   };
  // });
</script>

<QueryClientProvider client={queryClient}>
  <DevJWTProvider let:jwt>
    <RuntimeProvider host={RuntimeUrl} instanceId="default" {jwt}>
      <RillDeveloperLayout>
        <slot />
      </RillDeveloperLayout>
    </RuntimeProvider>
  </DevJWTProvider>
</QueryClientProvider>
