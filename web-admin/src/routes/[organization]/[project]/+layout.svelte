<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { startWatchResourcesClient } from "@rilldata/web-common/features/entity-management/watch-resources-client";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/RuntimeProvider.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { useProjectRuntime } from "../../../features/projects/selectors";
  import { viewAsUserStore } from "../../../features/view-as-user/viewAsUserStore";

  const queryClient = useQueryClient();

  $: projRuntime = useProjectRuntime(
    $page.params.organization,
    $page.params.project
  );

  $: isRuntimeHibernating = $projRuntime.isSuccess && !$projRuntime.data;

  $: if (isRuntimeHibernating) {
    // Redirect any nested routes (notably dashboards) to the project page
    goto(`/${$page.params.organization}/${$page.params.project}`);
  }

  // start the global resource watcher
  onMount(() => startWatchResourcesClient(queryClient));
</script>

<!-- Note: we don't provide the runtime here when the user is being spoofed via the "View As" functionality.
    In these cases, the "View as" actions manually set the runtime.  -->
{#if $projRuntime.data && !$viewAsUserStore}
  <RuntimeProvider
    host={$projRuntime.data.host}
    instanceId={$projRuntime.data.instanceId}
    jwt={$projRuntime.data?.jwt}
  >
    <slot />
  </RuntimeProvider>
{:else}
  <slot />
{/if}
