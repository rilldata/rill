<script lang="ts">
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;

  $: runtime.set({
    host: host,
    instanceId: instanceId,
  });
</script>

<slot />
