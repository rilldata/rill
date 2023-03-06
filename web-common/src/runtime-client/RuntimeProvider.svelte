<script lang="ts">
  import { runtime } from "./runtime-store";

  export let host: string;
  export let instanceId: string;
  export let jwt: string = undefined;

  $: runtime.set({
    host: host,
    instanceId: instanceId,
    jwt: jwt,
  });
</script>

<slot />
