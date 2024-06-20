<script lang="ts">
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;
  export let jwt: string | undefined = undefined;

  $: runtime.setRuntime(host, instanceId, jwt);
</script>

{#if $runtime.host && $runtime.instanceId}
  <slot />
{/if}
