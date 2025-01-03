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

  $: ({ instanceId: _instanceId, host: _host } = $runtime);
</script>

{#if _instanceId && _host}
  <slot />
{/if}
