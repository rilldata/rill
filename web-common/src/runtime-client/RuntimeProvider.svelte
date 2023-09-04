<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { invalidateAllMetricsViews } from "./invalidation";
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;
  export let jwt: string = undefined;

  $: runtime.set({
    host: host,
    instanceId: instanceId,
    jwt: jwt,
  });

  const queryClient = useQueryClient();

  // Whenever the runtime's (dev) JWT changes (even to `null`), invalidate all metrics views.
  // Metrics views may have a security policy, for which a JWT grants permission.
  $: ($runtime?.jwt || $runtime?.jwt === null) &&
    invalidateAllMetricsViews(queryClient, instanceId);
</script>

{#if $runtime.host !== undefined && $runtime.instanceId}
  <slot />
{/if}
