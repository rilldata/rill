<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { type AuthContext, runtime } from "./runtime-store";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext;

  $: runtime
    .setRuntime(queryClient, host, instanceId, jwt, authContext)
    .catch(console.error);

  $: ({ host: _host, instanceId: _instanceId } = $runtime);
</script>

{#if _host && _instanceId}
  <slot />
{/if}
