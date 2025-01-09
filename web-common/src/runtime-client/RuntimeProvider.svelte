<script lang="ts">
  import { useQueryClient } from "@tanstack/svelte-query";
  import { type AuthContext, runtime } from "./runtime-store";

  const queryClient = useQueryClient();

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;
  export let authContext: AuthContext;

  $: runtime
    .setRuntime(
      queryClient,
      "http://localhost:8081",
      "03000859377c4b33a8fe18062dfa1dca",
      jwt,
      authContext,
    )
    .catch(console.error);

  $: ({ instanceId, host } = $runtime);
</script>

{#if host && instanceId}
  <slot />
{/if}
