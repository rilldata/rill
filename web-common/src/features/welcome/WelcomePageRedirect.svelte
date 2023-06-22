<!--
  This file checks for the existence of a `rill.yaml` file and handles the corresponding scenarios:
  - If the file exists, the app continues as normal.
  - If the file does not exist, the user is redirected to the Welcome page.

  We perform the check in two different ways:
  - onMount: Ensures that on the initial page load, we only proceed after the check is complete to avoid any content flash.
  - continuously: If the user deletes the `rill.yaml` file, they are immediately redirected to the Welcome page.
-->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import {
    isProjectInitialized,
    useIsProjectInitialized,
  } from "./is-project-initialized";

  // Initial check onMount
  let ready = false;
  onMount(async () => {
    const initialized = await isProjectInitialized($runtime.instanceId);

    if (!initialized) {
      await goto("/welcome");
    }
    ready = true;
  });

  // Continuous check
  $: isProjectInitializedQuery = useIsProjectInitialized($runtime.instanceId);
  $: if ($isProjectInitializedQuery.data === false) {
    goto("/welcome");
  }
</script>

{#if ready}
  <slot />
{/if}
