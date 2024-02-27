<script lang="ts">
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;

  $: runtime.set({
    host: host,
    instanceId: instanceId,
    jwt: jwt
      ? {
          token: jwt,
          receivedAt: Date.now(),
        }
      : undefined,
  });
</script>

{#if $runtime.host !== undefined && $runtime.instanceId}
  <slot />
{/if}
