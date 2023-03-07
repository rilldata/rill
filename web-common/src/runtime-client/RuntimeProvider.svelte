<script lang="ts">
  import { invalidateRuntimeQueries } from "@rilldata/web-local/lib/svelte-query/invalidation";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;
  export let jwt: string = undefined;

  $: runtime.set({
    host: host,
    instanceId: instanceId,
    jwt: jwt,
  });

  // Re-run all runtime queries when `host` or `jwt` change
  // By default, a new `instanceId` triggers a re-run because it's in the queryKeys
  const queryClient = useQueryClient();
  $: (host || jwt) && invalidateRuntimeQueries(queryClient);
</script>

{#if $runtime.host && $runtime.instanceId}
  <slot />
{/if}
