<script lang="ts">
  import { invalidateRuntimeQueries } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;
  export let jwt: string = undefined;

  $: runtime.set({
    host: host,
    instanceId: instanceId,
    jwt: jwt,
  });

  // Re-run all runtime queries when `host` changes
  // By default, a new `instanceId` triggers a re-run because it's in the queryKeys
  const queryClient = useQueryClient();
  $: host && invalidateRuntimeQueries(queryClient);
</script>

{#if $runtime.host !== undefined && $runtime.instanceId}
  <slot />
{/if}
