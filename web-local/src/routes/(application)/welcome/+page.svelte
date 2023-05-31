<script lang="ts">
  import { goto } from "$app/navigation";
  import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized";
  import Welcome from "@rilldata/web-common/features/welcome/Welcome.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  onMount(async () => {
    const initialized = await isProjectInitialized($runtime.instanceId);

    if (initialized) {
      await goto("/");
    }
  });
</script>

<Welcome />
