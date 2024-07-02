<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "./runtime-store";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;

  $: runtime
    .setRuntime(queryClient, host, instanceId, jwt)
    .catch(console.error);
</script>

{#if $runtime.host && $runtime.instanceId}
  <slot />
{/if}
